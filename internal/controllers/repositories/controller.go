// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package repositories

import (
	"context"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

const (
	ArgoRepositoryLabel      = "argocd.argoproj.io/secret-type"
	ArgoRepositoryLabelValue = "repository"
)

type Reconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Log        logr.Logger
	Settings   *stores.ConfigStore
	ConfigName string
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(secretPredicate())).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(func(_ context.Context, obj client.Object) []reconcile.Request {
				ns := obj.GetLabels()[meta.RepositorySourceNamespaceLabel]
				name := obj.GetLabels()[meta.RepositorySourceNameLabel]

				if ns == "" || name == "" {
					return nil
				}

				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Namespace: ns,
						Name:      name,
					}},
				}
			}),
		).
		Watches(
			&configv1alpha1.ArgoAddon{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, _ client.Object) []reconcile.Request {
				var list corev1.SecretList
				if err := r.Client.List(ctx, &list,
					client.MatchingLabels{ArgoRepositoryLabel: ArgoRepositoryLabelValue},
				); err != nil {
					return nil
				}

				reqs := make([]reconcile.Request, 0, len(list.Items))
				for _, item := range list.Items {
					reqs = append(reqs, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Namespace: item.Namespace,
							Name:      item.Name,
						},
					})
				}

				return reqs
			}),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return e.Object.GetName() == r.ConfigName
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return e.Object.GetName() == r.ConfigName
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					if e.ObjectNew.GetName() != r.ConfigName {
						return false
					}

					oldObj, ok1 := e.ObjectOld.(*configv1alpha1.ArgoAddon)
					newObj, ok2 := e.ObjectNew.(*configv1alpha1.ArgoAddon)
					if !ok1 || !ok2 {
						return false
					}

					return oldObj.Spec.AllowRepositoryCreation != newObj.Spec.AllowRepositoryCreation
				},
			}),
		).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secret", req.Name, "namespace", req.Namespace)

	var src corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &src); err != nil {
		if k8serrors.IsNotFound(err) {
			return r.deleteReplica(ctx, req.Name, req.Namespace)
		}

		return ctrl.Result{}, err
	}

	hasLabel := src.Labels[ArgoRepositoryLabel] == ArgoRepositoryLabelValue
	if !r.Settings.Get().AllowRepositoryCreation || !hasLabel {
		return r.deleteReplica(ctx, src.Name, req.Namespace)
	}

	tntList := &capsulev1beta2.TenantList{}
	if err := r.Client.List(ctx, tntList, client.MatchingFieldsSelector{
		Selector: fields.OneTermEqualSelector(".status.namespaces", src.Namespace),
	}); err != nil {
		return ctrl.Result{}, err
	}

	// Not considered as part of tenants
	if len(tntList.Items) == 0 {
		log.V(5).Info("not part of a tenant namespace", "secret", src.Name, "namespace", src.Namespace)

		return ctrl.Result{}, nil
	}

	tnt := tntList.Items[0]

	// replicate into target namespace
	replica := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      src.Namespace + "-" + src.Name,
			Namespace: r.Settings.Get().Argo.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, replica, func() error {
		if replica.Annotations == nil {
			replica.Annotations = map[string]string{}
		}

		for l, v := range src.Annotations {
			replica.Annotations[l] = v
		}

		if replica.Labels == nil {
			replica.Labels = map[string]string{}
		}

		for l, v := range src.Labels {
			replica.Labels[l] = v
		}

		replica.Labels[ArgoRepositoryLabel] = ArgoRepositoryLabelValue
		replica.Labels[meta.ManagedByLabel] = meta.ManagedByLabelValue
		replica.Labels[meta.ProvisionedByLabel] = meta.ManagedByLabelValue
		replica.Labels[meta.RepositorySourceNameLabel] = src.Name
		replica.Labels[meta.RepositorySourceNamespaceLabel] = src.Namespace

		replica.Data = make(map[string][]byte, len(src.Data))
		for k, v := range src.Data {
			replica.Data[k] = v
		}

		replica.Data["project"] = []byte(tnt.GetName())
		replica.Type = src.Type

		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Info("Replicated secret", "source", req.NamespacedName, "target", r.Settings.Get().Argo.Namespace+"/"+src.Name)

	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteReplica(ctx context.Context, secretName string, secretNamespace string) (ctrl.Result, error) {
	replica := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretNamespace + "-" + secretName,
			Namespace: r.Settings.Get().Argo.Namespace,
		},
	}

	if err := r.Delete(ctx, replica); err != nil && !k8serrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func secretPredicate() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetLabels()[ArgoRepositoryLabel] == ArgoRepositoryLabelValue
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldHas := e.ObjectOld.GetLabels()[ArgoRepositoryLabel] == ArgoRepositoryLabelValue
			newHas := e.ObjectNew.GetLabels()[ArgoRepositoryLabel] == ArgoRepositoryLabelValue

			return oldHas || newHas
		},
	}
}
