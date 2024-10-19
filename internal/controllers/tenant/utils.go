package tenant

import (
	"context"

	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Add Ownerreference, which does not cascade a deletion of the tenant
// Also considers Decoupling from the tenant
func (i *TenancyController) DynamicOwnerReference(ctx context.Context, obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	if utils.TenantDecoupleProject(tenant) {
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
	} else {
		return controllerutil.SetControllerReference(tenant, obj, i.Client.Scheme())
	}

	return nil
}

// Determines if the proxy service should be registered
func (i *TenancyController) provisionProxyService(ctx context.Context, tenant *capsulev1beta2.Tenant) (provision bool) {
	provision = false

	// Check if the tenant is registered for the proxy
	if i.Settings.Get().Proxy.Enabled {
		if utils.TenantProxyRegister(tenant) {
			provision = true
		}
	}

	return
}
