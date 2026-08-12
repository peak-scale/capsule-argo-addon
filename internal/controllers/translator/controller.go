// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package translator

import (
	"context"
	"reflect"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/metrics"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
}

func (i *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize Channel
	i.requeue = make(chan event.GenericEvent)

	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.ArgoTranslator{}).
		Complete(i)
}

func (i *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log := i.Log.WithValues("translator", request.Name)

	origin := &configv1alpha1.ArgoTranslator{}
	if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Request object not found, could have been deleted after reconcile request")

			// Cleanup Metricss
			i.Metrics.DeleteTranslatorCondition(request.Name)

			return reconcile.Result{}, nil
		}

		return reconcile.Result{}, err
	}

	pruned, err := i.pruneMissingTenantStatuses(ctx, request.NamespacedName)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pruned {
		if err := i.Client.Get(ctx, request.NamespacedName, origin); err != nil {
			return reconcile.Result{}, err
		}
	}

	// Emit Metrics
	i.Metrics.RecordTranslatorCondition(origin)

	// Synchronize Finalizer status
	if err := i.syncFinalizerStatus(ctx, request.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (i *Reconciler) pruneMissingTenantStatuses(ctx context.Context, key client.ObjectKey) (bool, error) {
	tenants := &capsulev1beta2.TenantList{}
	if err := i.Client.List(ctx, tenants); err != nil {
		return false, err
	}

	pruned := false

	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current := &configv1alpha1.ArgoTranslator{}
		if err := i.Client.Get(ctx, key, current); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		previous := current.Status.DeepCopy()

		current.PruneMissingTenantStatuses(tenants.Items)

		if reflect.DeepEqual(previous, current.Status.DeepCopy()) {
			return nil
		}

		pruned = true

		return i.Client.Status().Update(ctx, current)
	})
	if err != nil {
		return false, err
	}

	return pruned, nil
}

func (i *Reconciler) syncFinalizerStatus(ctx context.Context, key client.ObjectKey) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		current := &configv1alpha1.ArgoTranslator{}
		if err := i.Client.Get(ctx, key, current); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}

			return err
		}

		previous := current.DeepCopy()
		current.SyncFinalizerStatus()

		if reflect.DeepEqual(previous.Finalizers, current.Finalizers) {
			return nil
		}

		return i.Client.Patch(ctx, current, client.MergeFrom(previous))
	})
}
