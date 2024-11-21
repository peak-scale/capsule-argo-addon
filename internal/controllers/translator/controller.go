// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package translator

import (
	"context"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/reflection"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
)

var _ reconcile.Reconciler = &TranslatorController{}

type TranslatorController struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Settings *stores.ConfigStore
	requeue  chan event.GenericEvent
}

func (i *TranslatorController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	// Initialize Channel
	i.requeue = make(chan event.GenericEvent)

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.ArgoTranslator{}).
		// Reconcile when an appproject is directly deleted and may include non translated
		// attributes, which should not be wiped.
		Watches(&argocdapi.AppProject{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, a client.Object) []reconcile.Request {
				// Based on finalizers get other finalizing translators
				translators := meta.GetTranslatingFinalizers(a)

				var requests []reconcile.Request
				for _, translator := range translators {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name: translator,
						},
					})
				}

				return requests
			}),
		).
		Complete(i)
}

func (i *TranslatorController) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("translator", request.Name)

	origin := &configv1alpha1.ArgoTranslator{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Finalize Dependencies
	if !origin.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
			log.V(5).Info("finalizing translator")
			err := i.finalize(ctx, log, origin)
			if err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(origin, meta.ControllerFinalizer)
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
		return ctrl.Result{
			Requeue: false,
		}, nil
	}

	// Collect Tenants from Status and verify if they are still active
	tnts := origin.GetTenantNames()
	for _, tnt := range tnts {
		tenant := &capsulev1beta2.Tenant{}
		err := i.Client.Get(ctx, client.ObjectKey{
			Name: tnt,
		}, tenant)
		// Remove unexisting tenants
		if k8serrors.IsNotFound(err) {
			log.V(5).Info("garbage collection", "tenant", tenant.Name)
			origin.RemoveTenantCondition(tnt)
			continue
		}
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update Status if necessary
	//err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
	//	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, origin, func() error {
	//		return i.Client.Status().Update(ctx, origin, &client.SubResourceUpdateOptions{})
	//	})
	//
	//	return
	//})
	//if err != nil {
	//	return ctrl.Result{}, err
	//}

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

func (i *TranslatorController) finalize(ctx context.Context, log logr.Logger, translator *configv1alpha1.ArgoTranslator) error {
	// Finalize all tenants (approjects)
	tnts := translator.GetTenantNames()

	for _, tnt := range tnts {
		tenant := &capsulev1beta2.Tenant{}
		err := i.Client.Get(ctx, client.ObjectKey{
			Name: tnt,
		}, tenant)
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		// Remove the approject from the tenant
		approject := &argocdapi.AppProject{}
		err = i.Client.Get(ctx, client.ObjectKey{
			Name:      meta.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		}, approject)
		if k8serrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return err
		}

		// if tenant is no longer managing an approject
		//if !controllerutil.ContainsFinalizer(tenant, meta.ControllerFinalizer) {
		//	continue
		//}

		if err := RemoveTranslatorForTenant(ctx, i.Client, log, translator, tenant, approject, i.Settings); err != nil {
			return err
		}

		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			err = i.Client.Update(ctx, approject)
			return
		})
		if err != nil {
			return err
		}

	}

	return nil
}

// Remove Translator for tenant
func RemoveTranslatorForTenant(
	ctx context.Context,
	c client.Client,
	log logr.Logger,
	translator *configv1alpha1.ArgoTranslator,
	tenant *capsulev1beta2.Tenant,
	approject *argocdapi.AppProject,
	settings *stores.ConfigStore,
) error {

	// Remove the approject from the tenant
	cfg, err := translator.Spec.ProjectSettings.GetConfig(
		tpl.ConfigContext("", translator, settings.Get(), tenant), tpl.ExtraFuncMap())
	if err != nil {
		return err
	}

	currentSpec := approject.Spec.DeepCopy()

	reflection.Subtract(currentSpec, &cfg.ProjectSpec)

	log.V(7).Info("finalized spec", "spec", currentSpec)
	approject.Spec = *currentSpec

	// Remove transformer labels from the approject
	for key, value := range cfg.ProjectMeta.Labels {
		if currentValue, ok := approject.Labels[key]; ok {
			if currentValue == value {
				delete(approject.Labels, key)
			}
		}
	}
	// Remove transformer annotations from the approject
	for key, value := range cfg.ProjectMeta.Annotations {
		if currentValue, ok := approject.Annotations[key]; ok {
			if currentValue == value {
				delete(approject.Annotations, key)
			}
		}
	}

	// Remove Finalizers from the approject
	finalizers := append(cfg.ProjectMeta.Finalizers, meta.TranslatorFinalizer(translator.Name))
	for _, finalizer := range finalizers {
		if controllerutil.ContainsFinalizer(approject, finalizer) {
			controllerutil.RemoveFinalizer(approject, finalizer)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
