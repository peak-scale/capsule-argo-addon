package meta

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// ManagerLabel
	ManagedTenantLabel = "argo.addons.projectcapsule.dev/tenant"

	// ManagedByLabel
	ManagedByLabel      = "app.kubernetes.io/managed-by"
	ManagedByLabelValue = "capsule-argocd-addon"
	ProvisionedByLabel  = "app.kubernetes.io/provisioned-by"
)

// Tracking Labels for resources provisioned by this controller
func TranslatorTrackingLabels(tenant *capsulev1beta2.Tenant) map[string]string {
	labels := TrackingLabels()
	labels[ProvisionedByLabel] = ManagedByLabelValue
	labels[ManagedTenantLabel] = tenant.Name

	return labels
}

// Respects the labels from the objects and just overwrites the tracking labels
func WithTranslatorTrackingLabels(obj client.Object, tenant *capsulev1beta2.Tenant) (labels map[string]string) {
	labels = obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	for key, value := range TranslatorTrackingLabels(tenant) {
		labels[key] = value
	}

	return
}

func TranslatorRemoveTenantLabels(labels map[string]string) map[string]string {
	delete(labels, ManagedTenantLabel)
	delete(labels, ManagedByLabel)

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
