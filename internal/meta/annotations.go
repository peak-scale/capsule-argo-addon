// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"strings"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

const (
	// Finalizer for the ArgoCD addon
	ControllerFinalizer = "argo.addons.projectcapsule.dev/finalize"

	// Annotation on Tenant
	// Change the Appproject Name for the tenant
	AnnotationProjectName = "argo.addons.projectcapsule.dev/name"

	// Annotation on Tenant
	// Change the ServiceAccount Namespace for the tenant
	AnnotationServiceAccountNamespace = "argo.addons.projectcapsule.dev/service-account-namespace"

	// Annotation on Tenant
	// Apply force for this tenant
	AnnotationForce = "argo.addons.projectcapsule.dev/force"

	// Annotation on Tenant
	// Annotation to control if a ServiceAccount should be registered as destination
	AnnotationDestinationRegister = "argo.addons.projectcapsule.dev/register-dest"

	// Annotation on Tenant
	// Annotation to control the proxy registration
	AnnotationUseProxy = "argo.addons.projectcapsule.dev/use-proxy"

	// Annotation on Tenant
	// Decouple Ownerreference from the origin tenant, to avoid deletion of the appproject
	AnnotationProjectDecouple = "argo.addons.projectcapsule.dev/decouple"

	// Annotation on Tenant
	// Read-Only mode for the approject (every change from approject ownership is ignored)
	AnnotationProjectReadOnly = "argo.addons.projectcapsule.dev/read-only"
)

// Tenant Approject-Name
func TenantProjectName(tenant *capsulev1beta2.Tenant) (name string) {
	name = tenant.Annotations[AnnotationProjectName]
	if name == "" {
		name = tenant.Name
	}

	return
}

// Tenant ServiceAccount Namespace
func TenantServiceAccountNamespace(tenant *capsulev1beta2.Tenant) string {
	return tenant.Annotations[AnnotationServiceAccountNamespace]
}

func TenantRegisterDestination(tenant *capsulev1beta2.Tenant) bool {
	return ProccessBoolean(tenant.Annotations[AnnotationDestinationRegister], false)
}

func TenantDecoupleProject(tenant *capsulev1beta2.Tenant) bool {
	return ProccessBoolean(tenant.GetAnnotations()[AnnotationProjectDecouple], false)
}

func TenantReadOnly(tenant *capsulev1beta2.Tenant) bool {
	return ProccessBoolean(tenant.GetAnnotations()[AnnotationProjectReadOnly], false)
}

func ProccessBoolean(val string, def bool) bool {
	switch strings.ToLower(val) {
	case "true", "enable":
		return true
	case "false", "disable":
		return false
	default:
		return def
	}
}
