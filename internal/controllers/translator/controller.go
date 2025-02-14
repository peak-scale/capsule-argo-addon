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
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
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

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

var _ reconcile.Reconciler = &Controller{}

type Controller struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Settings *stores.ConfigStore
	requeue  chan event.GenericEvent
}

func (i *Controller) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize Channel
	i.requeue = make(chan event.GenericEvent)

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.ArgoTranslator{}).
		// Reconcile when an appproject is directly deleted and may include non translated
		// attributes, which should not be wiped.
		Watches(&argocdapi.AppProject{},
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, a client.Object) []reconcile.Request {
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

func (i *Controller) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("translator", request.Name)

	origin := &configv1alpha1.ArgoTranslator{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Finalize Dependencies
	//nolint:nestif
	if !origin.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
			log.V(5).Info("finalizing translator")

			if err := i.finalize(ctx, origin); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(origin, meta.ControllerFinalizer)

			if err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
				if err := i.Client.Update(ctx, origin); err != nil {
					return err
				}

				return
			}); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{
			Requeue: false,
		}, nil
	}

	// Collect Tenants from Status and verify if they are still active
	if err := i.garbageCollectTenants(ctx, origin); err != nil {
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(origin, meta.ControllerFinalizer) {
		controllerutil.AddFinalizer(origin, meta.ControllerFinalizer)

		if err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			if err := i.Client.Update(ctx, origin); err != nil {
				return err
			}

			return
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// Remove Tenants which are no longer existent from status.
func (i *Controller) garbageCollectTenants(
	ctx context.Context,
	translator *configv1alpha1.ArgoTranslator,
) error {
	tnts := translator.GetTenantNames()
	for _, tnt := range tnts {
		tenant := &capsulev1beta2.Tenant{}
		err := i.Client.Get(ctx, client.ObjectKey{
			Name: tnt,
		}, tenant)
		// Remove unexisting tenants
		if k8serrors.IsNotFound(err) {
			translator.RemoveTenantCondition(tnt)

			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Takes Old specification and removes it.
// func (i *Controller) garbageCollectTranslator(
//	ctx context.Context,
//	translator *configv1alpha1.ArgoTranslator,
// ) error {
//	log := i.Log.WithValues("translator", translator.Name)
//	log.V(5).Info("garbage collection", "translator", translator.Name)
//
//	// Remove old Translator-Layout for any tracked Tenants.
//	tenantNames := translator.GetTenantNames()
//	for _, tnt := range tenantNames {
//		// Get the Tenant.
//		tenant := &capsulev1beta2.Tenant{}
//		if err := i.Client.Get(ctx, client.ObjectKey{Name: tnt}, tenant); err != nil {
//			if k8serrors.IsNotFound(err) {
//				// Tenant no longer exists, skip it.
//				return nil
//			}
//
//			return err
//		}
//
//		// Fetch the current state of the AppProject for this tenant.
//		appProject := &argocdapi.AppProject{}
//		appProjectKey := client.ObjectKey{
//			Name:      tenant.Name,
//			Namespace: i.Settings.Get().Argo.Namespace,
//		}
//
//		err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
//			if err := i.Client.Get(ctx, appProjectKey, appProject); err != nil {
//				if k8serrors.IsNotFound(err) {
//					// AppProject no longer exists, skip it.
//					return nil
//				}
//
//				return err
//			}
//
//			// Retrieve the configuration for the translator's project settings.
//			cfg, err := translator.GetProjectConfig(
//				tpl.ConfigContext(i.Settings.Get(), tenant),
//				tpl.ExtraFuncMap(),
//			)
//			if err != nil {
//				return err
//			}
//
//			// Make a deep copy of the current AppProject spec.
//			currentSpec := appProject.Spec.DeepCopy()
//
//			// Use CreateOrUpdate to subtract the translator's AppProject layout.
//			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, appProject, func() error {
//				// Subtract the translator-specific settings.
//				reflection.Subtract(currentSpec, &cfg.ProjectSpec)
//				appProject.Spec = *currentSpec
//
//				return nil
//			})
//
//			return err
//		})
//		if err != nil {
//			i.Log.Error(err, "failed removing old specification")
//
//			continue
//		}
//	}
//
//	return nil
//}

// Remove Translator for tenant.
func RemoveTranslatorForTenant(
	translator *configv1alpha1.ArgoTranslator,
	tenant *capsulev1beta2.Tenant,
	appproject *argocdapi.AppProject,
	settings *stores.ConfigStore,
) (err error) {
	finalizer := meta.TranslatorFinalizer(translator.Name)
	if controllerutil.ContainsFinalizer(appproject, finalizer) {
		controllerutil.RemoveFinalizer(appproject, finalizer)
	}

	return SubstractTranslatorSpec(translator, tenant, appproject, settings)
}

// Remove Translator for tenant.
func SubstractTranslatorSpec(
	translator *configv1alpha1.ArgoTranslator,
	tenant *capsulev1beta2.Tenant,
	appproject *argocdapi.AppProject,
	settings *stores.ConfigStore,
) (err error) {
	// Verify if currently Something is serving
	stat := translator.GetTenantStatus(tenant)
	if stat == nil {
		return nil
	}

	cfg, err := stat.Serving.GetConfig(
		tpl.ConfigContext(settings.Get(), tenant), tpl.ExtraFuncMap())
	if err != nil {
		return err
	}

	if cfg == nil {
		return nil
	}

	// Specification
	reflection.Subtract(&appproject.Spec, cfg.ProjectSpec)

	// Metadata
	if cfg.ProjectMeta == nil {
		return nil
	}
	// Remove transformer labels from the approject
	for key, value := range cfg.ProjectMeta.Labels {
		if currentValue, ok := appproject.Labels[key]; ok {
			if currentValue == value {
				delete(appproject.Labels, key)
			}
		}
	}
	// Remove transformer annotations from the approject
	for key, value := range cfg.ProjectMeta.Annotations {
		if currentValue, ok := appproject.Annotations[key]; ok {
			if currentValue == value {
				delete(appproject.Annotations, key)
			}
		}
	}

	// Remove Finalizers from the approject

	for _, finalizer := range cfg.ProjectMeta.Finalizers {
		if controllerutil.ContainsFinalizer(appproject, finalizer) {
			controllerutil.RemoveFinalizer(appproject, finalizer)
		}
	}

	return nil
}

func (i *Controller) finalize(
	ctx context.Context,
	translator *configv1alpha1.ArgoTranslator,
) error {
	// Finalize all tenants (approjects)
	tnts := translator.GetTenantNames()

	for _, tnt := range tnts {
		tenant := &capsulev1beta2.Tenant{}
		if err := i.Client.Get(ctx, client.ObjectKey{
			Name: tnt,
		}, tenant); err != nil && !k8serrors.IsNotFound(err) {
			return err
		}

		// Remove the approject from the tenant
		approject := &argocdapi.AppProject{}
		err := i.Client.Get(ctx, client.ObjectKey{
			Name:      meta.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		}, approject)

		if k8serrors.IsNotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		if err := RemoveTranslatorForTenant(translator, tenant, approject, i.Settings); err != nil {
			return err
		}

		if err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
			return i.Client.Update(ctx, approject)
		}); err != nil {
			return err
		}
	}

	return nil
}
