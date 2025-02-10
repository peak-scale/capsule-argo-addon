// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"testing"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestAddDynamicTenantOwnerReference(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = capsulev1beta2.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
			UID:  types.UID("1234"),
		},
	}

	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	err := AddDynamicTenantOwnerReference(scheme, obj, tenant, false)
	assert.NoError(t, err)

	// Verify the owner reference was added
	assert.True(t, HasTenantOwnerReference(obj, tenant), "Expected tenant owner reference to be added")
}

func TestRemoveDynamicTenantOwnerReference(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
			UID:  types.UID("1234"),
		},
	}

	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: capsulev1beta2.GroupVersion.String(),
					Kind:       "Tenant",
					Name:       tenant.Name,
					UID:        tenant.UID,
				},
			},
		},
	}

	// Ensure the owner reference is present initially
	assert.True(t, HasTenantOwnerReference(obj, tenant), "Expected tenant owner reference to exist")

	// Remove the owner reference and verify it was removed
	err := RemoveDynamicTenantOwnerReference(obj, tenant)
	assert.NoError(t, err)
	assert.False(t, HasTenantOwnerReference(obj, tenant), "Expected tenant owner reference to be removed")
}

func TestHasTenantOwnerReference(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-tenant",
			UID:  types.UID("1234"),
		},
	}

	// Object without owner references
	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	assert.False(t, HasTenantOwnerReference(obj, tenant), "Expected tenant owner reference to be absent")

	// Add the owner reference manually
	obj.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: capsulev1beta2.GroupVersion.String(),
			Kind:       "Tenant",
			Name:       tenant.Name,
			UID:        tenant.UID,
		},
	}

	assert.True(t, HasTenantOwnerReference(obj, tenant), "Expected tenant owner reference to be present")
}
