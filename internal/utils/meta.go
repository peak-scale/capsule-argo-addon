package utils

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Tracking Labels for resources provisioned by this controller
func TranslatorTrackingLabels(tenant *capsulev1beta2.Tenant) map[string]string {
	labels := TrackingLabels()
	labels["argocd.addons.projectcapsule.dev/tenant"] = tenant.Name

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
