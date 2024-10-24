package tenant

import (
	"context"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
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
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &TenancyController{}

type TenancyController struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Settings *stores.ConfigStore
}

func (i *TenancyController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capsulev1beta2.Tenant{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		//Owns(&argocdapi.AppProject{}).
		Watches(&argocdapi.AppProject{}, handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), &capsulev1beta2.Tenant{})).
		// Whenever a translator is updated, we need to reconcile all tenants
		Watches(&configv1alpha1.ArgoTranslator{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {

				tenants := &capsulev1beta2.TenantList{}
				err := i.Client.List(context.TODO(), tenants)
				if err != nil {
					return nil
				}

				var requests []reconcile.Request
				for _, tenant := range tenants.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name: tenant.Name,
						},
					})
				}

				return requests
			}),
		).
		Complete(i)
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
		if len(translator.GetTranslatingFinalizers(origin)) == 0 {
			if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
				log.V(5).Info("finalizing tenant")
				err := i.finalize(origin, ctx)
				if err != nil {
					return ctrl.Result{}, err
				}
			}

			return ctrl.Result{
				Requeue: false,
			}, nil
		}

		log.V(5).Info("")

		if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
			log.V(5).Info("finalizing tenant")
			err := i.finalize(origin, ctx)
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

	// Reconcile the Argo Assets
	err = i.reconcileArgoProject(ctx, log, tenant, translators)
	if err != nil {
		return translators, err
	}

	// Update the tenant status
	for _, selected := range translators {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tenant.DeepCopy(), func() error {
				selected.AssignTenant(tenant)

				return i.Client.Status().Update(ctx, selected, &client.SubResourceUpdateOptions{})
			})

			return
		})
		if err != nil {
			return translators, err
		}
	}

	log.V(7).Info("unmatched translators", "count", len(unmatchedTranslators))

	// Lifecycle from unmatched tenants
	for _, unmatched := range unmatchedTranslators {
		err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tenant.DeepCopy(), func() error {
				unmatched.UnassignTenant(tenant.Name)

				return i.Client.Status().Update(ctx, unmatched, &client.SubResourceUpdateOptions{})
			})

			return
		})
		if err != nil {
			return translators, err
		}
	}
	return translators, nil
}

// Selects all the translators from the configuration, which match the tenant's labels
// Returns all translators to run garbage collection on them
func (i *TenancyController) aggregateConfigTranslators(allTranslators *v1alpha1.ArgoTranslatorList, tenant *capsulev1beta2.Tenant) (
	matchedTranslators []*v1alpha1.ArgoTranslator,
	unmatchedTranslators []*v1alpha1.ArgoTranslator,
	err error,
) {
	matchedTranslators = make([]*v1alpha1.ArgoTranslator, 0)
	unmatchedTranslators = make([]*v1alpha1.ArgoTranslator, 0)
	tenantLabels := labels.Set(tenant.Labels)

	for _, translator := range allTranslators.Items {
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

func (i *TenancyController) finalize(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	return i.lifecycle(ctx, tenant)
}

// Patch the tenant from the argocd configmap
func (i *TenancyController) lifecycle(ctx context.Context, tenant *capsulev1beta2.Tenant) (err error) {
	if !controllerutil.ContainsFinalizer(tenant, meta.ControllerFinalizer) {
		return nil
	}

	// Update existing configmap with new csv
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
