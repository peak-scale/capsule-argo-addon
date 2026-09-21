// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"slices"

	rbacv1 "k8s.io/api/rbac/v1"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsulerbac "github.com/projectcapsule/capsule/pkg/api/rbac"
)

const (
	ownerUser  = "User"
	ownerGroup = "Group"
)

type TenantPermission struct {
	Kind         string
	Name         string
	ClusterRoles []string
}

func GetTenantGroups(tenant *capsulev1beta2.Tenant) (groups map[string]TenantPermission) {
	permissions := GetTenantPermissions(tenant)

	if _, exists := permissions["Group"]; exists {
		groups = permissions["Group"]
	}

	return
}

// GetClusterRolePermissions returns a map where the clusterRole is the key
// and the value is a list of permission subjects (kind and name) that reference that role.
func GetClusterRolePermissions(tenant *capsulev1beta2.Tenant) (rolePerms map[string][]rbacv1.Subject) {
	rolePerms = make(map[string][]rbacv1.Subject)

	forEachTenantPermission(tenant, func(kind, name string, clusterRoles []string) {
		subject := rbacv1.Subject{Kind: kind, Name: name}
		for _, clusterRole := range clusterRoles {
			if !slices.Contains(rolePerms[clusterRole], subject) {
				rolePerms[clusterRole] = append(rolePerms[clusterRole], subject)
			}
		}
	})

	return rolePerms
}

// Get the permissions for a tenant ordered by groups and users.
func GetTenantPermissions(tenant *capsulev1beta2.Tenant) map[string]map[string]TenantPermission {
	permissions := make(map[string]map[string]TenantPermission)

	forEachTenantPermission(tenant, func(kind, name string, clusterRoles []string) {
		if _, exists := permissions[kind]; !exists {
			permissions[kind] = make(map[string]TenantPermission)
		}

		perm := permissions[kind][name]
		perm.ClusterRoles = append(perm.ClusterRoles, clusterRoles...)
		permissions[kind][name] = perm
	})

	// Remove duplicates from cluster roles in both maps
	for kind, nameMap := range permissions {
		for name, perm := range nameMap {
			perm.ClusterRoles = uniqueStrings(perm.ClusterRoles)
			permissions[kind][name] = perm
		}
	}

	return permissions
}

// forEachTenantPermission visits users and groups from Capsule's resolved owners
// and both forms of additional role bindings. Spec owners are resolved by Capsule
// into status before they are used for Argo permissions.
func forEachTenantPermission(tenant *capsulev1beta2.Tenant, visit func(kind, name string, clusterRoles []string)) {
	for _, owner := range tenant.Status.Owners {
		if owner.Kind == capsulerbac.UserOwner || owner.Kind == capsulerbac.GroupOwner {
			visit(owner.Kind.String(), owner.Name, owner.ClusterRoles)
		}
	}

	visitBindings := func(bindings []capsulerbac.AdditionalRoleBindingsSpec) {
		for _, binding := range bindings {
			for _, subject := range binding.Subjects {
				if subject.Kind == ownerUser || subject.Kind == ownerGroup {
					visit(subject.Kind, subject.Name, []string{binding.ClusterRoleName})
				}
			}
		}
	}

	visitBindings(tenant.Spec.AdditionalRoleBindings)

	for _, rule := range tenant.Spec.Rules {
		if rule != nil {
			visitBindings(rule.Permissions.Bindings)
		}
	}
}

// Helper function to remove duplicates from a slice of strings.
func uniqueStrings(input []string) []string {
	seen := make(map[string]struct{})
	result := []string{}

	for _, str := range input {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}

			result = append(result, str)
		}
	}

	return result
}
