package tenant

import (
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Get Destination String
func (i *TenancyController) GetClusterDestination(tenant *capsulev1beta2.Tenant) (dest string) {
	dest = i.Settings.Get().Argo.Destination

	if i.provisionProxyService() {
		i.Settings.Get().ProxyServiceString(tenant)
	}

	return
}

// Gets the API Server given via Rest-Config
func (i *TenancyController) RetrieveAPIServerURL() string {
	return i.Rest.Host
}

// Decouple a Tenant from an Object
func (i *TenancyController) DecoupleTenant(obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	if err = meta.RemoveDynamicTenantOwnerReference(obj, tenant); err != nil {
		return
	}

	// Remove Tracking Labels
	obj.SetLabels(meta.TranslatorRemoveTenantLabels(obj.GetLabels()))

	return
}

// Determines if the proxy service should be registered
func (i *TenancyController) ForceTenant(tenant *capsulev1beta2.Tenant) bool {
	return meta.ProccessBoolean(tenant.GetAnnotations()[meta.AnnotationForce], i.Settings.Get().Force)
}

// Determines if the proxy service should be registered
func (i *TenancyController) provisionProxyService() (provision bool) {
	provision = false

	// Check if the tenant is registered for the proxy
	if i.Settings.Get().Proxy.Enabled && !i.Settings.Get().Argo.DestinationServiceAccounts {
		provision = true
	}

	return
}

// Determines if an argo cluster destination should be registered
//
//nolint:gosimple
func (i *TenancyController) registerCluster(tenant *capsulev1beta2.Tenant) (provision bool) {
	provision = false

	if val, ok := tenant.Annotations[meta.AnnotationDestinationRegister]; ok {
		return meta.ProccessBoolean(val, false)
	}

	// If you use serviceaccounts
	if i.Settings.Get().Proxy.Enabled && !i.Settings.Get().Argo.DestinationServiceAccounts {
		provision = true
	}

	return
}
