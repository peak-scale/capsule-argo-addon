package tenant

import (
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
func (i *TenancyController) provisionProxyService(tenant *capsulev1beta2.Tenant) (provision bool) {
	provision = false

	// Check if the tenant is registered for the proxy
	if i.Settings.Get().Proxy.Enabled && meta.TenantProxyRegister(tenant) {
		provision = true
	}

	return
}
