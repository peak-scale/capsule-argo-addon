package utils

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const TenantType = "kubernetes.gelan.cloud/type"

func ArgoPolicyName(tenant *capsulev1beta2.Tenant) string {
	return "policy." + tenant.Name + ".csv"
}

func CommonLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": "tenancy-controller",
		"app.kubernetes.io/name":     "tenancy-controller",
	}
}

func GetOwnerReference(tenant *capsulev1beta2.Tenant) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         tenant.APIVersion, // Ensure this is the correct APIVersion for the tenant
		Kind:               tenant.Kind,       // Ensure this is the correct Kind for the tenant
		Name:               tenant.Name,
		UID:                tenant.UID,
		Controller:         ptr.To(true),
		BlockOwnerDeletion: ptr.To(false),
	}
}

func IsSystemTenant(tenant *capsulev1beta2.Tenant) bool {
	return tenant.Labels[TenantType] == "system"
}

func StringSliceContains(slice []string, element string) bool {
	for _, sliceElement := range slice {
		if sliceElement == element {
			return true
		}
	}
	return false
}
