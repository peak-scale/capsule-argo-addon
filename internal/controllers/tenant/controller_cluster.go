// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	ccaerrrors "github.com/peak-scale/capsule-argo-addon/internal/errors"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Creates or updates the ArgoCD Cluster for the tenant (Tenant ServiceAccount, Cluster Secret).
func (i *Reconciler) reconcileArgoCluster(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	token string,
) (
	err error,
) {
	// Initialize Secret
	serverSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Settings.Get().Argo.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Get Cluster-Secret
	err = i.Client.Get(ctx, client.ObjectKey{Name: serverSecret.Name, Namespace: serverSecret.Namespace}, serverSecret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	log.V(7).Info("reconciling cluster", "secret", tenant.Name, "namespace", i.Settings.Get().Argo.Namespace)

	// Remove Cluster-Secret if not enabled. Token is deleted cascading via OwnerReference
	if !i.Settings.Get().RegisterCluster(tenant) || token == "" {
		return i.lifecycleArgoCluster(ctx, tenant)
	}

	if !meta.HasTenantOwnerReference(serverSecret, tenant) {
		if !i.Settings.Get().ForceTenant(tenant) && !k8serrors.IsNotFound(err) {
			log.V(5).Info(
				"proxy already present, not overriding",
				"serviceaccount", serverSecret.Name,
				"namespace", serverSecret.Namespace)

			return ccaerrrors.NewObjectAlreadyExistsError(serverSecret)
		}
	}

	// Dynamic
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, serverSecret, func() error {
		// Update secret metadata
		labels := meta.WithTranslatorTrackingLabels(serverSecret, tenant)
		labels["argocd.argoproj.io/secret-type"] = "cluster"
		serverSecret.SetLabels(labels)

		extraData := map[string]interface{}{
			"bearerToken": token,
			"tlsClientConfig": map[string]interface{}{
				"insecure": true,
			},
		}

		jsonData, err := json.Marshal(extraData)
		if err != nil {
			return fmt.Errorf("failed to marshal secret data: %w", err)
		}

		serverSecret.StringData = map[string]string{
			"name":    tenant.Name,
			"project": tenant.Name,
			"server":  i.Settings.Get().GetClusterDestination(tenant),
			"config":  string(jsonData),
		}

		return meta.AddDynamicTenantOwnerReference(i.Client.Scheme(), serverSecret, tenant, i.Settings.Get().DecoupleTenant(tenant))
	})
	if err != nil {
		return err
	}

	log.Info("Argo Server created", "name", tenant.Name)

	return nil
}

// Remove/Decouple Cluster Secret...
func (i *Reconciler) lifecycleArgoCluster(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
) (
	err error,
) {
	// Initialize Secret
	serverSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Settings.Get().Argo.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	err = i.Client.Get(ctx, client.ObjectKey{Name: serverSecret.Name, Namespace: serverSecret.Namespace}, serverSecret)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	if k8serrors.IsNotFound(err) {
		return nil
	}

	// Delete the AppProject when it's not decoupled
	if !i.Settings.Get().DecoupleTenant(tenant) {
		return i.Client.Delete(ctx, serverSecret)
	}

	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, serverSecret, func() (err error) {
		return i.DecoupleTenant(serverSecret, tenant)
	})

	return
}
