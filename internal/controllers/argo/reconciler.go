package argo

import (
	"context"
	"fmt"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
		Owns(&argocdapi.AppProject{}).
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

// Predicate to Trigger on Translators
func ConfigMapEventPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// You can add filtering logic here if needed
			return true // Reconcile on any update
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true // Reconcile on any creation
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true // Reconcile on any deletion
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return true // Reconcile on any generic event
		},
	}
}

func (i *TenancyController) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("tenant", request.Name)

	origin := &capsulev1beta2.Tenant{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		log.V(1).Error(err, "Unable to fetch tenant")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Finalize Dependencies
	if !origin.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(origin, utils.ControllerFinalizer) {
			err := i.finalize(origin, ctx)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Finalize tenant %s", err)
			}
			controllerutil.RemoveFinalizer(origin, utils.ControllerFinalizer)
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
				if err := i.Client.Update(ctx, origin); err != nil {
					return err
				}

				return
			})
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	err := i.reconcile(ctx, log, origin)
	if err != nil {
		log.Error(err, "reconcile error")
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(origin, utils.ControllerFinalizer) {
		controllerutil.AddFinalizer(origin, utils.ControllerFinalizer)
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
func (i *TenancyController) reconcile(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant) error {
	allTranslators := &v1alpha1.ArgoTranslatorList{}
	if err := i.Client.List(context.Background(), allTranslators); err != nil {
		return err
	}

	log.V(1).Info("Available Translators", "count", len(allTranslators.Items))

	// Fetch Translators Applying to the Tenant
	translators, err := i.aggregateConfigTranslators(allTranslators, tenant)
	log.V(1).Info("Matched Translators", "count", len(translators))
	if err != nil {
		return err
	}

	// Remove the lifecycle if there are no translators
	if len(translators) == 0 {
		if controllerutil.ContainsFinalizer(tenant, utils.ControllerFinalizer) {
			err = i.lifecycle(tenant, ctx)
			if err != nil {
				return err
			}
		}

	} else {
		err = i.reconcileArgoProject(ctx, log, tenant, translators)
		if err != nil {
			return err
		}
	}

	return nil
}

// Selects all the translators from the configuration, which match the tenant's labels
// Returns all translators to run garbage collection on them
func (i *TenancyController) aggregateConfigTranslators(allTranslators *v1alpha1.ArgoTranslatorList, tenant *capsulev1beta2.Tenant) (
	matchedTranslators []*v1alpha1.ArgoTranslator,
	err error,
) {
	tenantLabels := labels.Set(tenant.Labels)

	for _, translator := range allTranslators.Items {
		if translator.Spec.Selector == nil {
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
		}
	}

	return
}

func (r *TenancyController) updateTenantStatus(ctx context.Context, tnt *capsulev1beta2.Tenant) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if tnt.Spec.Cordoned {
			tnt.Status.State = capsulev1beta2.TenantStateCordoned
		} else {
			tnt.Status.State = capsulev1beta2.TenantStateActive
		}

		return r.Client.Status().Update(ctx, tnt)
	})
}

func (i *TenancyController) finalize(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	return i.lifecycle(tenant, ctx)
}
