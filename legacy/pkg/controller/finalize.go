package controller

import (
	"context"

	"git.bedag.cloud/gelan/gelan-infra/controllers/tenancy-controller/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const ControllerFinalizer = "kubernetes.gelan.cloud/tenancy-controller"

func (i *TenancyController) finalize(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	return i.finalizeArgo(tenant, ctx)
}

func (i *TenancyController) finalizeArgo(tenant *capsulev1beta2.Tenant, ctx context.Context) error {

	// Update existing configmap with new csv
	configmap := &corev1.ConfigMap{}
	err := i.Client.Get(ctx, client.ObjectKey{Name: "argocd-rbac-cm", Namespace: i.Options.ArgoCDNamespace}, configmap)
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

	return nil
}
