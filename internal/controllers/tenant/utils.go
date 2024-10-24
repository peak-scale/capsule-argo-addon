package tenant

import (
	"context"

	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Add Ownerreference, which does not cascade a deletion of the tenant
// Also considers Decoupling from the tenant
func (i *TenancyController) DynamicOwnerReference(ctx context.Context, obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	err = controllerutil.SetControllerReference(tenant, obj, i.Client.Scheme())
	if err != nil {
		return
	}

	if meta.TenantDecoupleProject(tenant) {
		ownerRefs := obj.GetOwnerReferences()
		// Remove blockOwnerDeletion and controller only if they are currently set
		needsUpdate := false
		for i, ownerRef := range ownerRefs {
			if ownerRef.UID == tenant.UID {
				if ownerRef.BlockOwnerDeletion != nil || ownerRef.Controller != nil {
					ownerRefs[i].BlockOwnerDeletion = nil
					ownerRefs[i].Controller = nil
					needsUpdate = true
				}
				break
			}
		}
		if needsUpdate {
			obj.SetOwnerReferences(ownerRefs)
		}
	}

	return nil
}

// Remove an OwnerReference from an object from a tenant
func (i *TenancyController) DynamicRemoveOwnerReference(ctx context.Context, obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	ownerRefs := obj.GetOwnerReferences()
	// Remove blockOwnerDeletion and controller only if they are currently set
	needsUpdate := false
	for i, ownerRef := range ownerRefs {
		if ownerRef.UID == tenant.UID {
			ownerRefs = append(ownerRefs[:i], ownerRefs[i+1:]...)
			needsUpdate = true
			break
		}
	}
	if needsUpdate {
		obj.SetOwnerReferences(ownerRefs)
	}

	return nil

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
