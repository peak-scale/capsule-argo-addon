package tenant

import (
	"context"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ccaerrrors "github.com/peak-scale/capsule-argo-addon/internal/errors"
)

var _ reconcile.Reconciler = &TenancyController{}

type TenancyController struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Settings *stores.ConfigStore
	requeue  chan event.GenericEvent
}

func (i *TenancyController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	i.requeue = make(chan event.GenericEvent)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // Exit the goroutine if the context is canceled
			case <-i.Settings.NotifyChannel():
				// Send a requeue event to trigger reconciliation
				i.requeue <- event.GenericEvent{
					Object: &capsulev1beta2.Tenant{},
				}
			}
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&capsulev1beta2.Tenant{}).
		Watches(
			&corev1.ServiceAccount{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&capsulev1beta2.Tenant{},
			)).
		Watches(
			&corev1.Service{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&capsulev1beta2.Tenant{},
			)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&capsulev1beta2.Tenant{},
			)).
		//Owns(&argocdapi.AppProject{}).
		Watches(
			&argocdapi.AppProject{},
			handler.EnqueueRequestForOwner(
				mgr.GetScheme(),
				mgr.GetRESTMapper(),
				&capsulev1beta2.Tenant{},
			)).
		// Whenever a translator is updated, we need to reconcile all tenants
		Watches(&configv1alpha1.ArgoTranslator{}, i.TenantRequeueHandler()).
		// Reconcile When Configuration Changes
		WatchesRawSource(&source.Channel{Source: i.requeue}, i.TenantRequeueHandler()).
		Complete(i)
}

// Handler to reconcile all Tenants
func (i *TenancyController) TenantRequeueHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
		// List all tenants
		tenants := &capsulev1beta2.TenantList{}
		err := i.Client.List(ctx, tenants)
		if err != nil {
			i.Log.Error(err, "Failed to list tenants for reconciliation")
			return nil
		}

		// Enqueue each tenant for reconciliation
		var requests []reconcile.Request
		for _, tenant := range tenants.Items {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: tenant.Name,
				},
			})
		}

		return requests
	})
}

func (i *TenancyController) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("tenant", request.Name)

	log.V(7).Info("controller configuration", "config", i.Settings.Get())

	origin := &capsulev1beta2.Tenant{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	log.V(5).Info("reconciling addons")
	translators, err := i.reconcile(ctx, log, origin)
	if err != nil {
		log.Error(err, "reconcile error")
		return ctrl.Result{}, nil
	}

	if !origin.ObjectMeta.DeletionTimestamp.IsZero() || len(translators) == 0 {
		// Wait until all translators have finished
		if len(meta.GetTranslatingFinalizers(origin)) == 0 {
			if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
				log.V(5).Info("finalizing tenant")
				err := i.lifecycle(ctx, log, origin)
				if err != nil {
					return ctrl.Result{}, err
				}
			}

			return ctrl.Result{
				Requeue: false,
			}, nil
		}

		if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
			log.V(5).Info("finalizing tenant")
			err := i.lifecycle(ctx, log, origin)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{
			Requeue: false,
		}, nil
	}

	if !controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
		controllerutil.AddFinalizer(origin, meta.ControllerFinalizer)
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			if err := i.Client.Update(ctx, origin); err != nil {
				return err
			}

			return
		})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil

}

// Reconcile all the assets
func (i *TenancyController) reconcile(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
) (translators []*configv1alpha1.ArgoTranslator, err error) {
	allTranslators := &v1alpha1.ArgoTranslatorList{}
	if err := i.Client.List(context.Background(), allTranslators); err != nil {
		return nil, err
	}

	log.V(3).Info("available translators", "count", len(allTranslators.Items))

	// Fetch Translators Applying to the Tenant
	var unmatchedTranslators []*configv1alpha1.ArgoTranslator
	translators, unmatchedTranslators, err = i.aggregateConfigTranslators(allTranslators, tenant)
	log.V(3).Info("matched translators", "count", len(translators))
	if err != nil {
		return translators, err
	}
	unmatchedTranslatorMap := make(map[string]*configv1alpha1.ArgoTranslator)
	for _, translator := range unmatchedTranslators {
		unmatchedTranslatorMap[translator.Name] = translator
	}

	// Reconcile the Argo Assets
	reconcileErr := i.reconcileArgoProject(ctx, log, tenant, translators, unmatchedTranslatorMap)

	// Status handling always runs even when reconciliation failed
	// Evaluate Condition
	condition := i.handleCondition(tenant, reconcileErr)

	// Update the tenant status.
	for _, selected := range translators {
		log.V(5).Info("updating translator conditions", "translator", selected.Name)
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tenant.DeepCopy(), func() error {

				if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
					selected.RemoveTenantCondition(tenant.Name)
				} else {
					selected.UpdateTenantCondition(configv1alpha1.TenantStatus{
						Name:      tenant.Name,
						UID:       tenant.UID,
						Condition: condition,
					})
				}

				log.V(10).Info("new translator status", "translator", selected.Name, "status", selected.Status)

				return i.Client.Status().Update(ctx, selected, &client.SubResourceUpdateOptions{})
			})

			return
		})
		if err != nil {
			log.Info("failed to update translator statius")
			return translators, err
		}
	}

	log.V(7).Info("unmatched translators", "count", len(unmatchedTranslators))

	// Lifecycle from unmatched tenants
	for _, unmatched := range unmatchedTranslators {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tenant.DeepCopy(), func() error {
				unmatched.RemoveTenantCondition(tenant.Name)

				return i.Client.Status().Update(ctx, unmatched, &client.SubResourceUpdateOptions{})
			})

			return
		})
		if err != nil {
			return translators, err
		}
	}

	// Finally return if reconciliation had an error.
	if reconcileErr != nil {
		return translators, err
	}

	// Return on success
	return translators, nil
}

// Handle Condition assignment based on err provided
func (i *TenancyController) handleCondition(
	tenant *capsulev1beta2.Tenant,
	reconcileError error,
) (condition metav1.Condition) {
	if reconcileError == nil {
		// No error; set to Ready condition
		return meta.NewReadyCondition(tenant)
	}

	// Check the type of error with a type switch
	switch err := reconcileError.(type) {
	case *ccaerrrors.ObjectAlreadyExists:
		// Custom condition for ObjectAlreadyExistsError
		condition = meta.NewAlreadyExistsCondition(tenant, err.Error())
	default:
		// Default NotReady condition for other errors
		condition = meta.NewNotReadyCondition(tenant, reconcileError.Error())
	}

	return
}

// Selects all the translators from the configuration, which match the tenant's labels
// Returns all translators to run garbage collection on them
//
//nolint:nakedret
func (i *TenancyController) aggregateConfigTranslators(
	allTranslators *v1alpha1.ArgoTranslatorList,
	tenant *capsulev1beta2.Tenant,
) (
	matchedTranslators []*v1alpha1.ArgoTranslator,
	unmatchedTranslators []*v1alpha1.ArgoTranslator,
	err error,
) {
	matchedTranslators = make([]*v1alpha1.ArgoTranslator, 0)
	unmatchedTranslators = make([]*v1alpha1.ArgoTranslator, 0)
	tenantLabels := labels.Set(tenant.Labels)

	for _, trans := range allTranslators.Items {
		translator := trans

		// Skip translators that are being deleted
		if !translator.ObjectMeta.DeletionTimestamp.IsZero() {
			continue
		}

		if translator.Spec.Selector == nil {
			unmatchedTranslators = append(unmatchedTranslators, &translator)
			continue
		}

		// Convert LabelSelector to a labels.Selector
		var selector labels.Selector
		selector, err = metav1.LabelSelectorAsSelector(translator.Spec.Selector)
		if err != nil {

			return
		}

		// Check if tenant's labels match the translator's selector
		if selector.Matches(tenantLabels) {
			matchedTranslators = append(matchedTranslators, &translator)
		} else {
			unmatchedTranslators = append(unmatchedTranslators, &translator)
		}
	}

	return
}

// Patch the tenant from the argocd configmap
func (i *TenancyController) lifecycle(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant) (err error) {
	if !controllerutil.ContainsFinalizer(tenant, meta.ControllerFinalizer) {
		return nil
	}

	// Update existing configmap with new csv
	log.V(7).Info("lifecycling argo components")
	err = i.lifecycleArgo(ctx, tenant)

	// Remove Finalizers after tenant
	controllerutil.RemoveFinalizer(tenant, meta.ControllerFinalizer)
	if err := i.Client.Update(ctx, tenant); err != nil {
		return err
	}
	if err != nil {
		return err
	}

	return nil
}

func (i *TenancyController) lifecycleArgo(ctx context.Context, tenant *capsulev1beta2.Tenant) (err error) {
	// Update existing configmap with new csv
	if !meta.TenantDecoupleProject(tenant) {

		configmap := &corev1.ConfigMap{}
		err := i.Client.Get(ctx, client.ObjectKey{
			Name:      i.Settings.Get().Argo.RBACConfigMap,
			Namespace: i.Settings.Get().Argo.Namespace},
			configmap,
		)
		if err != nil {
			return err
		}

		_, err = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
			delete(configmap.Data, argo.ArgoPolicyName(tenant))

			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
