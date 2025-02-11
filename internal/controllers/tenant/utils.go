// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Gets the API Server given via Rest-Config.
func (i *TenancyController) RetrieveAPIServerURL() string {
	return i.Rest.Host
}

// Decouple a Tenant from an Object.
func (i *TenancyController) DecoupleTenant(obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	if err = meta.RemoveDynamicTenantOwnerReference(obj, tenant); err != nil {
		return
	}

	// Remove Tracking Labels
	obj.SetLabels(meta.TranslatorRemoveTenantLabels(obj.GetLabels()))

	return
}
