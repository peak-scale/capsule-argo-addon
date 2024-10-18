package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Creates or updates the ArgoCD Cluster for the tenant (Tenant ServiceAccount, Cluster Secret)
func (i *TenancyController) reconcileArgoCluster(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant, token string) (string, error) {

	// Initialize Secret
	serverSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Settings.Get().ArgoCD.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
	}

	// Handle the Proxy-Service for the tenant
	cluster, _ := i.proxyService(ctx, tenant)

	// Remove Cluster-Secret if not enabled. Token is deleted cascading via OwnerReference
	if !i.Settings.Get().Proxy.Enabled || !utils.TenantProxyRegister(tenant) {
		err := i.Client.Delete(ctx, serverSecret)
		if err != nil && !k8serrors.IsNotFound(err) {
			return "", fmt.Errorf("failed to lifecycle serviceaccount: %w", err)
		}
		return "", nil
	}

	if token == "" {
		return "", fmt.Errorf("no token provided")
	}

	// Create Cluster-Secret
	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, serverSecret, func() error {
		// Update secret metadata
		serverSecret.Labels = utils.TranslatorTrackingLabels(tenant)
		serverSecret.Labels = map[string]string{
			"argocd.argoproj.io/secret-type": "cluster",
		}

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
			"server":  cluster,
			"config":  string(jsonData),
		}
		return controllerutil.SetOwnerReference(tenant, serverSecret, i.Client.Scheme())
	})
	if err != nil {
		return "", err
	}
	log.Info("Argo Server created", "name", tenant.Name)
	return cluster, nil
}

// Proxy Service for the tenant
func (i *TenancyController) proxyService(ctx context.Context, tenant *capsulev1beta2.Tenant) (url string, err error) {
	// Create a dedicated service for the tenant
	replicatedName := tenant.Name
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Settings.Get().Proxy.CapsuleProxyServiceNamespace,
		},
	}

	// Validate if Proxy is enabled, lifeycle the service if not
	if !i.Settings.Get().Proxy.Enabled || !utils.TenantProxyRegister(tenant) {
		err := i.Client.Delete(ctx, service)
		if err != nil && !k8serrors.IsNotFound(err) {
			return "", fmt.Errorf("failed to lifecycle serviceaccount: %w", err)
		}

		// Return proxy service url
		if !i.Settings.Get().Proxy.Enabled {
			return i.proxyServiceName(tenant)
		}

		return "", nil

	}

	proxySvc := &corev1.Service{}
	err = i.Client.Get(ctx, types.NamespacedName{
		Namespace: i.Settings.Get().Proxy.CapsuleProxyServiceNamespace,
		Name:      i.Settings.Get().Proxy.CapsuleProxyServiceName,
	}, proxySvc)
	if err != nil {
		return "", fmt.Errorf("failed to resolve proxy service: %w", err)
	}

	// Replicate a proxy service for the tenant
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, service, func() error {
		service.Labels = utils.TranslatorTrackingLabels(tenant)

		// Replicate the proxy service ports
		service.Spec.Ports = proxySvc.Spec.Ports

		// Replicate the proxy service selector
		service.Spec.Selector = proxySvc.Spec.Selector

		return controllerutil.SetOwnerReference(tenant, service, i.Client.Scheme())
	})
	if err != nil {
		return "", err
	}

	i.Log.V(5).Info("Proxy Service created", "name", tenant.Name)

	// Returns the proxy service url
	return "https://" + replicatedName + "." +
		i.Settings.Get().Proxy.CapsuleProxyServiceNamespace + ".svc:" +
		strconv.Itoa(int(i.Settings.Get().Proxy.CapsuleProxyServicePort)), nil
}

// Resolve Proxy-Service for the tenant
func (i *TenancyController) proxyServiceName(tenant *capsulev1beta2.Tenant) (url string, err error) {
	return "https://" + i.Settings.Get().Proxy.CapsuleProxyServiceName + "." +
		i.Settings.Get().Proxy.CapsuleProxyServiceNamespace + ".svc:" +
		strconv.Itoa(int(i.Settings.Get().Proxy.CapsuleProxyServicePort)), nil
}
