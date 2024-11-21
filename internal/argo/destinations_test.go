// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package argo

import (
	"testing"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestProjectHasDestination(t *testing.T) {
	appProject := &argocdv1alpha1.AppProject{
		Spec: argocdv1alpha1.AppProjectSpec{
			Destinations: []argocdv1alpha1.ApplicationDestination{
				{
					Name:      "cluster1",
					Namespace: "namespace1",
					Server:    "https://example.com",
				},
				{
					Name:      "cluster2",
					Namespace: "namespace2",
					Server:    "https://example2.com",
				},
			},
		},
	}

	tests := []struct {
		dest     argocdv1alpha1.ApplicationDestination
		expected bool
		testName string
	}{
		{
			dest: argocdv1alpha1.ApplicationDestination{
				Name:      "cluster1",
				Namespace: "namespace1",
				Server:    "https://example.com",
			},
			expected: true,
			testName: "Existing destination should return true",
		},
		{
			dest: argocdv1alpha1.ApplicationDestination{
				Name:      "cluster3",
				Namespace: "namespace3",
				Server:    "https://example3.com",
			},
			expected: false,
			testName: "Non-existing destination should return false",
		},
		{
			dest: argocdv1alpha1.ApplicationDestination{
				Name:      "cluster2",
				Namespace: "namespace2",
				Server:    "https://example2.com",
			},
			expected: true,
			testName: "Another existing destination should return true",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			result := ProjectHasDestination(appProject, test.dest)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestRemoveProjectDestination(t *testing.T) {
	appProject := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-project",
		},
		Spec: argocdv1alpha1.AppProjectSpec{
			Destinations: []argocdv1alpha1.ApplicationDestination{
				{
					Name:      "cluster1",
					Namespace: "namespace1",
					Server:    "https://example.com",
				},
				{
					Name:      "cluster2",
					Namespace: "namespace2",
					Server:    "https://example2.com",
				},
				{
					Name:      "cluster3",
					Namespace: "namespace3",
					Server:    "https://example3.com",
				},
			},
		},
	}

	tests := []struct {
		destToRemove argocdv1alpha1.ApplicationDestination
		expectedDest []argocdv1alpha1.ApplicationDestination
		testName     string
	}{
		{
			destToRemove: argocdv1alpha1.ApplicationDestination{
				Name:      "cluster2",
				Namespace: "namespace2",
				Server:    "https://example2.com",
			},
			expectedDest: []argocdv1alpha1.ApplicationDestination{
				{
					Name:      "cluster1",
					Namespace: "namespace1",
					Server:    "https://example.com",
				},
				{
					Name:      "cluster3",
					Namespace: "namespace3",
					Server:    "https://example3.com",
				},
			},
			testName: "Remove existing destination",
		},
		{
			destToRemove: argocdv1alpha1.ApplicationDestination{
				Name:      "non-existing-cluster",
				Namespace: "non-existing-namespace",
				Server:    "https://non-existing-server.com",
			},
			expectedDest: []argocdv1alpha1.ApplicationDestination{
				{
					Name:      "cluster1",
					Namespace: "namespace1",
					Server:    "https://example.com",
				},
				{
					Name:      "cluster2",
					Namespace: "namespace2",
					Server:    "https://example2.com",
				},
				{
					Name:      "cluster3",
					Namespace: "namespace3",
					Server:    "https://example3.com",
				},
			},
			testName: "No changes if destination doesn't exist",
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			// Create a copy of the appProject to not mutate the original one
			clone := appProject.DeepCopy()

			// Remove the destination
			RemoveProjectDestination(clone, test.destToRemove)

			// Check if the destination was removed as expected
			assert.Equal(t, test.expectedDest, clone.Spec.Destinations)
		})
	}
}
