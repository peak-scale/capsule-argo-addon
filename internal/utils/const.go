package utils

import (
	"strings"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

const (
	// Finalizer for the ArgoCD addon
	ControllerFinalizer = "argocd.addons.projectcapsule.dev/finalize"
	// Change the ServiceAccount Namespace for the tenant
	AnnotationServiceAccountNamespace = "argocd.addons.projectcapsule.dev/service-account-namespace"
	// Annotation to control the proxy registration
	AnnotationProxyRegister = "argocd.addons.projectcapsule.dev/register-proxy"
)

// Tenant ServiceAccount Namespace
func TenantServiceAccountNamespace(tenant *capsulev1beta2.Tenant) string {
	return tenant.Annotations[AnnotationServiceAccountNamespace]
}

func TenantProxyRegister(tenant *capsulev1beta2.Tenant) bool {
	return metaIsValueTrue(tenant.Annotations[AnnotationProxyRegister])
}

func metaIsValueTrue(val string) bool {
	switch strings.ToLower(val) {
	case "true", "enable", "":
		return true
	default:
		return false
	}
}

func metaIsValueFalse(val string) bool {
	switch strings.ToLower(val) {
	case "true", "enable":
		return true
	default:
		return false
	}
}
