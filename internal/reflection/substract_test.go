// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package reflection

import (
	"testing"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestStruct struct {
	Whitelist []string
	Blacklist []string
}

func TestSubtractComplexStruct(t *testing.T) {
	// Define target and source complex structs
	target := &configv1alpha1.ArgocdProjectStructuredProperties{
		ProjectMeta: &configv1alpha1.ArgocdProjectPropertieMeta{
			Labels: map[string]string{
				"team": "dev",
				"env":  "prod",
			},
			Annotations: map[string]string{
				"annotation1": "value1",
			},
		},
		ProjectSpec: argocdv1alpha1.AppProjectSpec{
			SourceRepos:                     []string{"https://github.com/target/repo1", "https://github.com/target/repo2"},
			Destinations:                    []argocdv1alpha1.ApplicationDestination{{Name: "cluster1"}, {Name: "cluster2"}},
			Description:                     "This is a target project",
			ClusterResourceWhitelist:        []metav1.GroupKind{{Group: "*", Kind: "*"}},
			NamespaceResourceBlacklist:      []metav1.GroupKind{{Group: "core", Kind: "Pod"}},
			PermitOnlyProjectScopedClusters: true,
		},
	}

	source := &configv1alpha1.ArgocdProjectStructuredProperties{
		ProjectMeta: &configv1alpha1.ArgocdProjectPropertieMeta{
			Labels: map[string]string{
				"team": "dev",
			},
			Annotations: map[string]string{
				"annotation1": "value1",
			},
		},
		ProjectSpec: argocdv1alpha1.AppProjectSpec{
			SourceRepos:                     []string{"https://github.com/target/repo1"},
			Destinations:                    []argocdv1alpha1.ApplicationDestination{{Name: "cluster1"}},
			ClusterResourceWhitelist:        []metav1.GroupKind{{Group: "*", Kind: "*"}},
			PermitOnlyProjectScopedClusters: true,
		},
	}

	// Perform subtraction
	Subtract(target, source)

	// Check that only the non-matching fields remain
	//assert.Equal(t, map[string]string{"env": "prod"}, target.ProjectMeta.Labels)
	assert.Equal(t, []string{"https://github.com/target/repo2"}, target.ProjectSpec.SourceRepos)
	assert.Equal(t, []argocdv1alpha1.ApplicationDestination{{Name: "cluster2"}}, target.ProjectSpec.Destinations)
	assert.Equal(t, []metav1.GroupKind{}, target.ProjectSpec.ClusterResourceWhitelist)
	assert.Equal(t, []metav1.GroupKind{{Group: "core", Kind: "Pod"}}, target.ProjectSpec.NamespaceResourceBlacklist)
}

func TestSubstractStruct(t *testing.T) {
	// Define target and source structs
	target := &TestStruct{
		Whitelist: []string{"item1", "item2", "item3"},
		Blacklist: []string{"black1", "black2"},
	}

	source := &TestStruct{
		Whitelist: []string{"item1", "item3"},
		Blacklist: []string{"black3"},
	}

	// Perform subtraction
	Subtract(target, source)

	// Check that the whitelist has been subtracted correctly
	assert.Equal(t, []string{"item2"}, target.Whitelist)

	// Check that the blacklist has not been modified
	assert.Equal(t, []string{"black1", "black2"}, target.Blacklist)
}
