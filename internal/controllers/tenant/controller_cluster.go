package tenant

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	ccaerrrors "github.com/peak-scale/capsule-argo-addon/internal/errors"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Creates or updates the ArgoCD Cluster for the tenant (Tenant ServiceAccount, Cluster Secret)
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

	// Handle the Proxy-Service for the tenant
	if err := i.proxyService(ctx, log, tenant); err != nil {
		return fmt.Errorf("failed to reconcile destination service: %w", err)

	}

	// Decouple Object
	if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		if meta.TenantDecoupleProject(tenant) && !k8serrors.IsNotFound(err) {
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
		if !meta.TenantDecoupleProject(tenant) {
			return i.Client.Delete(ctx, serverSecret)
		} else {
			log.V(5).Info(
				"decoupling serviceaccount",
				"secret", tenant.Name,
				"namespace", i.Settings.Get().Argo.Namespace,
			)
			if err := i.DecoupleTenant(serverSecret, tenant); err != nil {
				return err
			}
		}
	}

	// Handle Force, if an object already exists with the same name
	if !meta.HasTenantOwnerReference(serverSecret, tenant) {
		if !i.ForceTenant(tenant) && !k8serrors.IsNotFound(err) {
			log.V(5).Info(
				"cluster secret already present, not overriding",
				"secret", tenant.Name,
				"namespace", i.Settings.Get().Argo.Namespace)

			return ccaerrrors.NewObjectAlreadyExistsError(serverSecret)
		}
	}

	// Remove Cluster-Secret if not enabled. Token is deleted cascading via OwnerReference
	if !i.registerCluster(tenant) || token == "" {
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
			"server":  i.GetClusterDestination(tenant),
			"config":  string(jsonData),
		}

		return meta.AddDynamicTenantOwnerReference(ctx, i.Client.Scheme(), serverSecret, tenant)
	})
	if err != nil {
		return err
	}
	log.Info("Argo Server created", "name", tenant.Name)
	return nil
}

// Proxy Service for the tenant
func (i *TenancyController) proxyService(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
) (err error) {
	// Create a dedicated service for the tenant
	replicatedName := tenant.Name
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Settings.Get().Proxy.CapsuleProxyServiceNamespace,
		},
	}

	log.V(7).Info(
		"reconciling service",
		"service", replicatedName,
		"namespace", i.Settings.Get().Proxy.CapsuleProxyServiceNamespace)

	// Get Cluster-Secret
	err = i.Client.Get(ctx, client.ObjectKey{Name: service.Name, Namespace: service.Namespace}, service)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	// Decouple Object
	if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		if meta.TenantDecoupleProject(tenant) && !k8serrors.IsNotFound(err) {
			_, err := controllerutil.CreateOrPatch(
				ctx,
				i.Client,
				service,
				func() error {
					log.V(5).Info("decoupling server secret", "secret", service.Name)
					if err := i.DecoupleTenant(service, tenant); err != nil {
						return err
					}

					return i.DecoupleTenant(service, tenant)
				})
			if err != nil {
				return err
			}

			return nil
		}
	}

	if !meta.HasTenantOwnerReference(service, tenant) {
		if !i.ForceTenant(tenant) && !k8serrors.IsNotFound(err) {
			log.V(5).Info("proxy already present, not overriding", "service", service.Name, "namespace", service.Namespace)

			return ccaerrrors.NewObjectAlreadyExistsError(service)
		}
	}

	// Validate if Proxy is enabled, lifeycle the service if not
	if !i.provisionProxyService() {
		log.V(7).Info("lifecycling proxy service")
		err := i.Client.Delete(ctx, service)
		if err != nil && !k8serrors.IsNotFound(err) {
			return fmt.Errorf("failed to lifecycle service: %w", err)
		}

		return nil
	}

	// Get Referenced Error
	proxySvc := &corev1.Service{}
	err = i.Client.Get(ctx, types.NamespacedName{
		Namespace: i.Settings.Get().Proxy.CapsuleProxyServiceNamespace,
		Name:      i.Settings.Get().Proxy.CapsuleProxyServiceName,
	}, proxySvc)
	if err != nil {
		return fmt.Errorf("failed to resolve proxy service %s/%s: %w",
			i.Settings.Get().Proxy.CapsuleProxyServiceNamespace,
			i.Settings.Get().Proxy.CapsuleProxyServiceName,
			err)
	}

	// Replicate a proxy service for the tenant
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, service, func() error {
		service.Labels = meta.TranslatorTrackingLabels(tenant)
		service.Spec.Ports = proxySvc.Spec.Ports
		service.Spec.Selector = proxySvc.Spec.Selector

		return meta.AddDynamicTenantOwnerReference(ctx, i.Client.Scheme(), service, tenant)
	})
	if err != nil {
		return err
	}

	i.Log.V(5).Info("Proxy Service created", "name", tenant.Name)

	// Returns the proxy service url
	return nil
}
