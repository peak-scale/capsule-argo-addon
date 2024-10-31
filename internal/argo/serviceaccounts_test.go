package argo

import (
	"testing"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestProjectHasServiceAccount(t *testing.T) {
	// Set up test data
	appProject := &argocdv1alpha1.AppProject{
		Spec: argocdv1alpha1.AppProjectSpec{
			DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
				{DefaultServiceAccount: "serviceaccount1", Namespace: "namespace1", Server: "https://example.com"},
				{DefaultServiceAccount: "serviceaccount2", Namespace: "namespace2", Server: "https://example.org"},
			},
		},
	}

	tests := []struct {
		name           string
		serviceAccount argocdv1alpha1.ApplicationDestinationServiceAccount
		expected       bool
	}{
		{
			name:           "Service account exists",
			serviceAccount: argocdv1alpha1.ApplicationDestinationServiceAccount{DefaultServiceAccount: "serviceaccount1", Namespace: "namespace1", Server: "https://example.com"},
			expected:       true,
		},
		{
			name:           "Service account does not exist",
			serviceAccount: argocdv1alpha1.ApplicationDestinationServiceAccount{DefaultServiceAccount: "serviceaccount3", Namespace: "namespace3", Server: "https://nonexistent.com"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProjectHasServiceAccount(appProject, tt.serviceAccount)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveProjectServiceaccount(t *testing.T) {
	// Set up test data
	appProject := &argocdv1alpha1.AppProject{
		Spec: argocdv1alpha1.AppProjectSpec{
			DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
				{DefaultServiceAccount: "serviceaccount1", Namespace: "namespace1", Server: "https://example.com"},
				{DefaultServiceAccount: "serviceaccount2", Namespace: "namespace2", Server: "https://example.org"},
				{DefaultServiceAccount: "serviceaccount3", Namespace: "namespace3", Server: "https://example.net"},
			},
		},
	}

	// Define the service account to remove
	removeServiceAccount := argocdv1alpha1.ApplicationDestinationServiceAccount{
		DefaultServiceAccount: "serviceaccount2", Namespace: "namespace2", Server: "https://example.org",
	}

	// Call RemoveProjectServiceaccount
	RemoveProjectServiceaccount(appProject, removeServiceAccount)

	// Expected result after removal
	expectedServiceAccounts := []argocdv1alpha1.ApplicationDestinationServiceAccount{
		{DefaultServiceAccount: "serviceaccount1", Namespace: "namespace1", Server: "https://example.com"},
		{DefaultServiceAccount: "serviceaccount3", Namespace: "namespace3", Server: "https://example.net"},
	}

	assert.Equal(t, expectedServiceAccounts, appProject.Spec.DestinationServiceAccounts)
}
