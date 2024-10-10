package argo

import (
	"context"

	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Patch the tenant from the argocd configmap
func (i *TenancyController) lifecycle(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	if controllerutil.ContainsFinalizer(tenant, utils.ControllerFinalizer) {
		return nil
	}

	// Update existing configmap with new csv
	configmap := &corev1.ConfigMap{}
	err := i.Client.Get(ctx, client.ObjectKey{
		Name:      i.Settings.Get().ArgoCD.RBACConfigMap,
		Namespace: i.Settings.Get().ArgoCD.Namespace},
		configmap,
	)
	if err != nil {
		return err
	}

	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
		delete(configmap.Data, utils.ArgoPolicyName(tenant))

		return nil
	})
	if err != nil {
		return err
	}

	// Remove Finalizers after tenant
	controllerutil.RemoveFinalizer(tenant, utils.ControllerFinalizer)
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		if err := i.Client.Update(ctx, tenant); err != nil {
			return err
		}

		return
	})

	return nil
}
