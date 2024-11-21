// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"reflect"
	"testing"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"
	rbacv1 "k8s.io/api/rbac/v1"
)

// TestGetClusterRolePermissions tests the GetClusterRolePermissions function
func TestGetClusterRolePermissions(t *testing.T) {
	// Sample tenant object
	tenant := &capsulev1beta2.Tenant{
		Spec: capsulev1beta2.TenantSpec{
			Owners: []capsulev1beta2.OwnerSpec{
				{
					Kind:         "User",
					Name:         "user1",
					ClusterRoles: []string{"cluster-admin", "read-only"},
				},
				{
					Kind:         "Group",
					Name:         "group1",
					ClusterRoles: []string{"edit"},
				},
				{
					Kind:         capsulev1beta2.ServiceAccountOwner,
					Name:         "service",
					ClusterRoles: []string{"read-only"},
				},
			},
			AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
				{
					ClusterRoleName: "developer",
					Subjects: []rbacv1.Subject{
						{Kind: "User", Name: "user2"},
						{Kind: "Group", Name: "group1"},
					},
				},
				{
					ClusterRoleName: "mega-admin",
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "user1",
						},
						{
							Kind: "Group",
							Name: "group1",
						},
					},
				},
			},
		},
	}

	expected := map[string][]rbacv1.Subject{
		"mega-admin": {
			{Kind: "User", Name: "user1"},
			{Kind: "Group", Name: "group1"},
		},

		"cluster-admin": {
			{Kind: "User", Name: "user1"},
		},
		"read-only": {
			{Kind: "User", Name: "user1"},
		},
		"edit": {
			{Kind: "Group", Name: "group1"},
		},
		"developer": {
			{Kind: "User", Name: "user2"},
			{Kind: "Group", Name: "group1"},
		},
	}

	// Call the function to test
	permissions := GetClusterRolePermissions(tenant)

	if !reflect.DeepEqual(permissions, expected) {
		t.Errorf("Expected %v, but got %v", expected, permissions)
	}
}

func TestGetTenantPermissions(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{
		Spec: capsulev1beta2.TenantSpec{
			Owners: []capsulev1beta2.OwnerSpec{
				{
					Kind:         capsulev1beta2.UserOwner,
					Name:         "user1",
					ClusterRoles: []string{"cluster-admin"},
				},
				{
					Kind:         capsulev1beta2.UserOwner,
					Name:         "user2",
					ClusterRoles: []string{"cluster-admin"},
				},
				{
					Kind:         capsulev1beta2.GroupOwner,
					Name:         "group1",
					ClusterRoles: []string{"read-only"},
				},
				{
					Kind:         capsulev1beta2.GroupOwner,
					Name:         "group2",
					ClusterRoles: []string{"read-only"},
				},
				{
					Kind:         capsulev1beta2.ServiceAccountOwner,
					Name:         "service",
					ClusterRoles: []string{"read-only"},
				},
			},
			AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
				{
					ClusterRoleName: "edit",
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "user1",
						},
						{
							Kind: "Group",
							Name: "group1",
						},
					},
				},
				{
					ClusterRoleName: "mega-admin",
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "user1",
						},
						{
							Kind: "Group",
							Name: "group1",
						},
					},
				},
				{
					ClusterRoleName: "mega-admin",
					Subjects: []rbacv1.Subject{
						{
							Kind: "ServiceAccount",
							Name: "system:serviceaccount:default:service",
						},
					},
				},
			},
		},
	}

	expected := map[string]map[string]TenantPermission{
		"User": {
			"user1": {
				ClusterRoles: []string{"cluster-admin", "edit", "mega-admin"},
			},
			"user2": {
				ClusterRoles: []string{"cluster-admin"},
			},
		},
		"Group": {
			"group1": {
				ClusterRoles: []string{"read-only", "edit", "mega-admin"},
			},
			"group2": {
				ClusterRoles: []string{"read-only"},
			},
		},
	}

	permissions := GetTenantPermissions(tenant)

	if !reflect.DeepEqual(permissions, expected) {
		t.Errorf("Expected %v, but got %v", expected, permissions)
	}
}

// Helper function to run tests
func TestMain(t *testing.M) {
	t.Run()
}
