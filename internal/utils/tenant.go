package utils

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	rbacv1 "k8s.io/api/rbac/v1"
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

	// Helper to add permissions for a given clusterRole
	addPermission := func(clusterRole string, permission rbacv1.Subject) {
		if _, exists := rolePerms[clusterRole]; !exists {
			rolePerms[clusterRole] = []rbacv1.Subject{}
		}
		rolePerms[clusterRole] = append(rolePerms[clusterRole], permission)
	}

	// Process owners
	for _, owner := range tenant.Spec.Owners {
		if owner.Kind != capsulev1beta2.ServiceAccountOwner {
			for _, clusterRole := range owner.ClusterRoles {
				perm := rbacv1.Subject{
					Name: owner.Name,
					Kind: owner.Kind.String(),
				}
				addPermission(clusterRole, perm)
			}
		}
	}

	// Process additional role bindings
	for _, role := range tenant.Spec.AdditionalRoleBindings {
		for _, subject := range role.Subjects {
			if subject.Kind != "ServiceAccount" {
				perm := rbacv1.Subject{
					Name: subject.Name,
					Kind: subject.Kind,
				}
				addPermission(role.ClusterRoleName, perm)
			}
		}
	}

	return rolePerms
}

// Get the permissions for a tenant ordered by groups and users
func GetTenantPermissions(tenant *capsulev1beta2.Tenant) map[string]map[string]TenantPermission {
	permissions := make(map[string]map[string]TenantPermission)

	// Initialize a nested map for kind ("User", "Group") and name
	initNestedMap := func(kind string) {
		if _, exists := permissions[kind]; !exists {
			permissions[kind] = make(map[string]TenantPermission)
		}
	}

	// Process owners
	for _, owner := range tenant.Spec.Owners {
		if owner.Kind == "User" || owner.Kind == "Group" {
			initNestedMap(owner.Kind.String())
			if perm, exists := permissions[owner.Kind.String()][owner.Name]; exists {
				// If the permission entry already exists, append cluster roles
				perm.ClusterRoles = append(perm.ClusterRoles, owner.ClusterRoles...)
				permissions[owner.Kind.String()][owner.Name] = perm
			} else {
				// Create a new permission entry
				permissions[owner.Kind.String()][owner.Name] = TenantPermission{
					ClusterRoles: owner.ClusterRoles,
				}
			}
		}
	}

	// Process additional role bindings
	for _, role := range tenant.Spec.AdditionalRoleBindings {
		for _, subject := range role.Subjects {
			if subject.Kind == "User" || subject.Kind == "Group" {
				initNestedMap(subject.Kind)
				if perm, exists := permissions[subject.Kind][subject.Name]; exists {
					// If the permission entry already exists, append cluster roles
					perm.ClusterRoles = append(perm.ClusterRoles, role.ClusterRoleName)
					permissions[subject.Kind][subject.Name] = perm
				} else {
					// Create a new permission entry
					permissions[subject.Kind][subject.Name] = TenantPermission{
						ClusterRoles: []string{role.ClusterRoleName},
					}
				}
			}
		}
	}

	// Remove duplicates from cluster roles in both maps
	for kind, nameMap := range permissions {
		for name, perm := range nameMap {
			perm.ClusterRoles = uniqueStrings(perm.ClusterRoles)
			permissions[kind][name] = perm
		}
	}

	return permissions
}

// Helper function to remove duplicates from a slice of strings
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
