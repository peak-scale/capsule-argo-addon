package v1alpha1

import (
	"strconv"

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
