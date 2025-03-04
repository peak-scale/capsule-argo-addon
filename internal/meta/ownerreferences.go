// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Also considers Decoupling from the tenant.
func AddDynamicTenantOwnerReference(
	schema *runtime.Scheme,
	obj client.Object,
	tenant *capsulev1beta2.Tenant,
	decouple bool,
) (err error) {
	err = controllerutil.SetOwnerReference(tenant, obj, schema)
	if err != nil {
		return err
	}

	if decouple {
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

// Remove an OwnerReference from an object from a tenant.
func RemoveDynamicTenantOwnerReference(obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
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

// If not returns false.
func HasTenantOwnerReference(obj client.Object, tenant *capsulev1beta2.Tenant) bool {
	for _, ownerRef := range obj.GetOwnerReferences() {
		if ownerRef.UID == tenant.UID {
			return true
		}
	}

	return false
}
