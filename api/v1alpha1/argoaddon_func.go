// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fmt"

	"github.com/peak-scale/capsule-argo-addon/internal/meta"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Get the Cluster-URL within argo-cd.
func (in *ArgoAddonSpec) GetClusterDestination(_ *capsulev1beta2.Tenant) (dest string) {
	return in.Argo.Destination
}

// Namespace where the serviceaccount will be placed.
func (in *ArgoAddonSpec) ServiceAccountNamespace(tenant *capsulev1beta2.Tenant) (namespace string) {
	namespace = in.Argo.ServiceAccountNamespace

	// Verify if ServiceAccount-Namespace is declared on tenant-basis
	if ns := meta.TenantServiceAccountNamespace(tenant); ns != "" {
		namespace = ns
	}

	return
}

// Prints Argo Destination annotation.
func (in *ArgoAddonSpec) DestinationServiceAccount(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("%s:%s", in.ServiceAccountNamespace(tenant), tenant.Name)
}

// Determines if resources should be decoupled.
func (in *ArgoAddonSpec) DecoupleTenant(tenant *capsulev1beta2.Tenant) bool {
	return meta.ProccessBoolean(tenant.GetAnnotations()[meta.AnnotationProjectDecouple], in.Decouple)
}

// Determines if existing resources should be overwritten (forced).
func (in *ArgoAddonSpec) ForceTenant(tenant *capsulev1beta2.Tenant) bool {
	return meta.ProccessBoolean(tenant.GetAnnotations()[meta.AnnotationForce], in.Force)
}

// Determines read-only is applied on tenant-basis.
func (in *ArgoAddonSpec) ReadOnlyTenant(tenant *capsulev1beta2.Tenant) bool {
	return meta.ProccessBoolean(tenant.GetAnnotations()[meta.AnnotationProjectReadOnly], in.ReadOnly)
}

// Determines if an argo cluster destination should be registered on a tenant-basis.
func (in *ArgoAddonSpec) RegisterCluster(tenant *capsulev1beta2.Tenant) (provision bool) {
	provision = false

	if val, ok := tenant.Annotations[meta.AnnotationDestinationRegister]; ok {
		return meta.ProccessBoolean(val, false)
	}

	return
}
