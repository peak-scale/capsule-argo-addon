// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
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
func (i *TenancyController) reconcileArgoCluster(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	token string,
	translators []*v1alpha1.ArgoTranslator,
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

	// Decouple Object
	//nolint:nestif
	if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		if i.Settings.Get().DecoupleTenant(tenant) && !k8serrors.IsNotFound(err) {
			_, err := controllerutil.CreateOrPatch(
				ctx,
				i.Client,
				serverSecret,
				func() error {
					log.V(5).Info("decoupling server secret", "secret", serverSecret.Name)

					if err := i.DecoupleTenant(serverSecret, tenant); err != nil {
						return err
					}

					return i.DecoupleTenant(serverSecret, tenant)
				})
			if err != nil {
				return err
			}

			return nil
		}
	}

	// Remove when umatched
	if len(translators) == 0 {
		// Approject is already absent
		if k8serrors.IsNotFound(err) {
			return nil
		}

		log.V(7).Info("reconciling cluster", "secret", tenant.Name, "namespace", i.Settings.Get().Argo.Namespace)

		// Delete the AppProject when it's not decoupled
		if !i.Settings.Get().DecoupleTenant(tenant) {
			return i.Client.Delete(ctx, serverSecret)
		}

		log.V(5).Info(
			"decoupling serviceaccount",
			"secret", tenant.Name,
			"namespace", i.Settings.Get().Argo.Namespace,
		)

		if err := i.DecoupleTenant(serverSecret, tenant); err != nil {
			return err
		}
	}

	// Handle Force, if an object already exists with the same name
	if !meta.HasTenantOwnerReference(serverSecret, tenant) {
		if !i.Settings.Get().ForceTenant(tenant) && !k8serrors.IsNotFound(err) {
			log.V(5).Info(
				"cluster secret already present, not overriding",
				"secret", tenant.Name,
				"namespace", i.Settings.Get().Argo.Namespace)

			return ccaerrrors.NewObjectAlreadyExistsError(serverSecret)
		}
	}

	// Remove Cluster-Secret if not enabled. Token is deleted cascading via OwnerReference
	if !i.Settings.Get().RegisterCluster(tenant) || token == "" {
		err := i.Client.Delete(ctx, serverSecret)
		if err != nil && !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to lifecycle destination: %w", err)
		}

		return nil
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
