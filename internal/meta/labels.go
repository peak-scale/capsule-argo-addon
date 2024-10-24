package meta

import capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"

const (
	// ManagerLabel
	ManagedTenantLabel = "argo.addons.projectcapsule.dev/tenant"

	// ManagedByLabel
	ManagedByLabel      = "app.kubernetes.io/managed-by"
	ManagedByLabelValue = "capsule-argocd-addon"
)

// Tracking Labels for resources provisioned by this controller
func TranslatorTrackingLabels(tenant *capsulev1beta2.Tenant) map[string]string {
	labels := TrackingLabels()
	labels[ManagedTenantLabel] = tenant.Name

	return labels
}

func TranslatorRemoveTenantLabels(labels map[string]string) map[string]string {
	delete(labels, ManagedTenantLabel)

	return labels
}

// Common Labels for tracking resources provisioned by this controller
func TrackingLabels() map[string]string {
	return map[string]string{
		ManagedByLabel: ManagedByLabelValue,
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
