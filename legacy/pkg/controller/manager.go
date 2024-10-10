package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// IngressController should implement the Reconciler interface
var _ reconcile.Reconciler = &TenancyController{}

type TenancyController struct {
	Client   client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Options  TenancyControllerOptions
}

type TenancyControllerOptions struct {
	CapsuleProxyServiceName      string
	CapsuleProxyServiceNamespace string
	CapsuleProxyServicePort      int32
	SystemTenantNamespace        string
	UserTenantNamespace          string
	ArgoCDNamespace              string
}

func (i *TenancyController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&capsulev1beta2.Tenant{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Complete(i)
}

func (i *TenancyController) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("tenant", request.NamespacedName)
	//reconcileStart := time.Now()
	//reconciliationLoopID := uuid.New().String()
	//log := ctrl.LoggerFrom(ctx, "reconciliation-loop-id", reconciliationLoopID, "start-time", reconcileStart)
	i.Log.V(3).Info("Reconciling",
		"tenant", request.Name,
	)

	log.V(3).Info("Fetch Tenant Resource")
	origin := &capsulev1beta2.Tenant{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		log.V(1).Error(err, "Unable to fetch tenant")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Finalize Dependencies
	if !origin.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(origin, ControllerFinalizer) {
			err := i.finalize(origin, ctx)
			if err != nil {
				return ctrl.Result{}, fmt.Errorf("Finalize tenant %s", err)
			}
			controllerutil.RemoveFinalizer(origin, ControllerFinalizer)
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

	i.Log.V(3).Info("Addons reconcile", "triggered-by", request.NamespacedName)

	err := i.reconcileAddons(origin, ctx)
	if err != nil {
		log.V(1).Error(err, "addons error")
		return ctrl.Result{}, nil
	}

	i.Log.V(3).Info("Addons reconciled", "triggered-by", request.NamespacedName)

	if !controllerutil.ContainsFinalizer(origin, ControllerFinalizer) {
		controllerutil.AddFinalizer(origin, ControllerFinalizer)
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

	i.Log.V(3).Info("Reconcile completed")
	return ctrl.Result{}, nil

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
