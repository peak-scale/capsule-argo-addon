package v1alpha1

import (
	"fmt"
	"strconv"

	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Assign Tenants to the ArgoTranslator
func (in *ArgoAddonSpec) ProxyServiceString(tenant *capsulev1beta2.Tenant) string {
	protocol := "https"
	if !in.Proxy.CapsuleProxyTLS {
		protocol = "http"
	}

	// Return Connection String
	return protocol + "://" + tenant.Name + "." +
		in.Proxy.CapsuleProxyServiceNamespace + ".svc:" +
		strconv.Itoa(int(in.Proxy.CapsuleProxyServicePort))
}

// Namespace where the serviceaccount will be placed
func (in *ArgoAddonSpec) ServiceAccountNamespace(tenant *capsulev1beta2.Tenant) (namespace string) {
	namespace = in.Argo.ServiceAccountNamespace

	// Verify if ServiceAccount-Namespace is declared on tenant-basis
	if ns := meta.TenantServiceAccountNamespace(tenant); ns != "" {
		namespace = ns
	}

	return
}

// Prints Argo Destination annotation
func (in *ArgoAddonSpec) DestinationServiceAccount(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("%s:%s", in.ServiceAccountNamespace(tenant), tenant.Name)
}
