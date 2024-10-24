package v1alpha1

import (
	"strconv"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Assign Tenants to the ArgoTranslator
func (in *ArgoAddonSpec) ProxyServiceString(tenant *capsulev1beta2.Tenant) string {
	return "https://" + in.Proxy.CapsuleProxyServiceName + "." +
		in.Proxy.CapsuleProxyServiceNamespace + ".svc:" +
		strconv.Itoa(int(in.Proxy.CapsuleProxyServicePort))
}
