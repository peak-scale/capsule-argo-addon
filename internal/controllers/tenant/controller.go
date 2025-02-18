// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"fmt"

	argocdapi "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/metrics"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	client.Client
	Metrics  *metrics.Recorder
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
	Settings *stores.ConfigStore
	requeue  chan event.GenericEvent
	Rest     *rest.Config
}

func (i *Reconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
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
		WatchesRawSource(source.Channel(i.requeue, i.TenantRequeueHandler())).
		Complete(i)
}

// Handler to reconcile all Tenants.
func (i *Reconciler) TenantRequeueHandler() handler.EventHandler {
	return handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, _ client.Object) []reconcile.Request {
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

func (i *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("tenant", request.Name)

	origin := &capsulev1beta2.Tenant{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		if k8serrors.IsNotFound(err) {
			log.V(5).Info("Request object not found, could have been deleted after reconcile request")

			origin = &capsulev1beta2.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					Name:      request.NamespacedName.Name,
					Namespace: request.NamespacedName.Namespace,
				},
			}

			// Cleanup ArgoCD
			ferr := i.finalize(ctx, log, origin)

			return reconcile.Result{}, ferr
		}

		return reconcile.Result{}, err
	}

	allTranslators := &configv1alpha1.ArgoTranslatorList{}
	if err := i.Client.List(ctx, allTranslators); err != nil {
		return reconcile.Result{}, err
	}

	translatorsPtr := make([]*configv1alpha1.ArgoTranslator, len(allTranslators.Items))
	for i := range allTranslators.Items {
		translatorsPtr[i] = &allTranslators.Items[i]
	}

	log.V(3).Info("available translators", "count", len(allTranslators.Items))

	finalize, err := i.reconcileProject(ctx, log, origin, translatorsPtr)
	if err != nil {
		i.Metrics.RecordTenantCondition(origin, meta.NotReadyCondition)

		log.Error(err, "reconcile error")

		return ctrl.Result{}, err
	}

	log.V(5).Info("reconciled translators", "count", len(allTranslators.Items))

	// Handle lifecycle
	if _, err = controllerutil.CreateOrUpdate(ctx, i.Client, origin, func() error {
		if finalize {
			log.V(5).Info("adding finalizer")

			i.Metrics.RecordTenantCondition(origin, meta.ReadyCondition)

			controllerutil.AddFinalizer(origin, meta.ControllerFinalizer)
		} else {
			log.V(5).Info("lifecycling")

			return i.finalize(ctx, log, origin)
		}

		return nil
	}); err != nil {
		return ctrl.Result{}, fmt.Errorf("error finalizing tenant; %w", err)
	}

	return ctrl.Result{}, nil
}

// Patch the tenant from the argocd configmap.
func (i *Reconciler) finalize(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant) (err error) {
	// Skip if finalizer no longer present
	if !controllerutil.ContainsFinalizer(tenant, meta.ControllerFinalizer) {
		return
	}

	log.V(7).Info("lifecycling", "decoupling", i.Settings.Get().DecoupleTenant(tenant))

	// Make sure Metrics are absent
	i.Metrics.DeleteTenantCondition(tenant.Name)

	log.V(7).Info("lifecycling argo project")

	// Remove the approject from the tenant
	if err = i.lifecycleArgoProject(ctx, tenant); err != nil {
		return
	}

	log.V(7).Info("lifecycling argo cluster")

	if err = i.lifecycleArgoCluster(ctx, tenant); err != nil {
		return
	}

	log.V(7).Info("lifecycling argo serviceaccount")

	if err = i.lifecycleArgoServiceAccount(ctx, tenant); err != nil {
		return
	}

	// Update existing configmap with new csv
	log.V(7).Info("lifecycling argo components")

	if err = i.lifecycleArgoRbac(ctx, tenant); err != nil {
		return
	}

	// Remove Finalizers after tenant
	controllerutil.RemoveFinalizer(tenant, meta.ControllerFinalizer)

	return
}
