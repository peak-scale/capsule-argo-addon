package utils

import (
	"strings"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

const (
	// ManagerLabel
	ManagedTenantLabel = "argo.addons.projectcapsule.dev/tenant"

	// Finalizer for the ArgoCD addon
	ControllerFinalizer = "argo.addons.projectcapsule.dev/finalize"

	// Annotation on Tenant
	// Change the Appproject Name for the tenant
	AnnotationProjectName = "argo.addons.projectcapsule.dev/name"

	// Annotation on Tenant
	// Change the ServiceAccount Namespace for the tenant
	AnnotationServiceAccountNamespace = "argo.addons.projectcapsule.dev/service-account-namespace"

	// Annotation on Tenant
	// Annotation to control the proxy registration
	AnnotationProxyRegister = "argo.addons.projectcapsule.dev/register-proxy"

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

func TenantProxyRegister(tenant *capsulev1beta2.Tenant) bool {
	return proccessBoolean(tenant.Annotations[AnnotationProxyRegister], true)
}

func TenantDecoupleProject(tenant *capsulev1beta2.Tenant) bool {
	return proccessBoolean(tenant.GetAnnotations()[AnnotationProjectDecouple], false)
}

func TenantReadOnly(tenant *capsulev1beta2.Tenant) bool {
	return proccessBoolean(tenant.GetAnnotations()[AnnotationProjectReadOnly], false)
}

func proccessBoolean(val string, def bool) bool {
	switch strings.ToLower(val) {
	case "true", "enable":
		return true
	case "false", "disable":
		return false
	default:
		return def
	}
}

// Tracking Labels for resources provisioned by this controller
func TranslatorTrackingLabels(tenant *capsulev1beta2.Tenant) map[string]string {
	labels := TrackingLabels()
	labels[ManagedTenantLabel] = tenant.Name

	return labels
}

// Common Labels for tracking resources provisioned by this controller
func TrackingLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "capsule-argocd-addon",
	}
}

// Verify if a string slice contains a specific element
func StringSliceContains(slice []string, element string) bool {
	for _, sliceElement := range slice {
		if sliceElement == element {
			return true
		}
	}
	return false
}
