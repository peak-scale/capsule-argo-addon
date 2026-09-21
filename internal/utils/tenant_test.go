// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func TestTenantPermissions(t *testing.T) {
	tests := []struct {
		name      string
		tenant    string
		byRole    map[string][]rbacv1.Subject
		bySubject map[string]map[string]TenantPermission
	}{
		{
			name:      "empty tenant",
			tenant:    `{}`,
			byRole:    map[string][]rbacv1.Subject{},
			bySubject: map[string]map[string]TenantPermission{},
		},
		{
			name: "spec owners are not used before Capsule resolves status",
			tenant: `{"spec":{"owners":[
				{"kind":"User","name":"alice","clusterRoles":["admin"]},
				{"kind":"Group","name":"team","clusterRoles":["admin"]}
			]}}`,
			byRole:    map[string][]rbacv1.Subject{},
			bySubject: map[string]map[string]TenantPermission{},
		},
		{
			name: "status owners are authoritative and include external owners",
			tenant: `{
				"spec":{"owners":[
					{"kind":"User","name":"alice","clusterRoles":["admin"]},
					{"kind":"Group","name":"unresolved","clusterRoles":["admin"]}
				]},
				"status":{"owners":[
					{"kind":"User","name":"alice","clusterRoles":["view","edit"]},
					{"kind":"Group","name":"external","clusterRoles":["view"]},
					{"kind":"ServiceAccount","name":"system:serviceaccount:default:robot","clusterRoles":["admin"]}
				]}
			}`,
			byRole: map[string][]rbacv1.Subject{
				"view": {{Kind: "User", Name: "alice"}, {Kind: "Group", Name: "external"}},
				"edit": {{Kind: "User", Name: "alice"}},
			},
			bySubject: map[string]map[string]TenantPermission{
				"User":  {"alice": {ClusterRoles: []string{"view", "edit"}}},
				"Group": {"external": {ClusterRoles: []string{"view"}}},
			},
		},
		{
			name: "legacy additional role bindings without owners",
			tenant: `{"spec":{"additionalRoleBindings":[
				{"clusterRoleName":"edit","subjects":[
					{"kind":"User","name":"alice"},
					{"kind":"Group","name":"team"},
					{"kind":"ServiceAccount","name":"robot","namespace":"default"}
				]}
			]}}`,
			byRole: map[string][]rbacv1.Subject{
				"edit": {{Kind: "User", Name: "alice"}, {Kind: "Group", Name: "team"}},
			},
			bySubject: map[string]map[string]TenantPermission{
				"User":  {"alice": {ClusterRoles: []string{"edit"}}},
				"Group": {"team": {ClusterRoles: []string{"edit"}}},
			},
		},
		{
			name: "bindings from every rule including rules with namespace selectors",
			tenant: `{"spec":{"rules":[
				null, {}, {"permissions":{}},
				{"permissions":{"bindings":[
					{"clusterRoleName":"view","subjects":[
						{"kind":"User","name":"alice","apiGroup":"rbac.authorization.k8s.io"},
						{"kind":"ServiceAccount","name":"robot","namespace":"default"}
					]}
				]}},
				{"namespaceSelector":{"matchLabels":{"environment":"dev"}},"permissions":{"bindings":[
					{"clusterRoleName":"edit","subjects":[{"kind":"Group","name":"team"}]},
					{"clusterRoleName":"unused","subjects":[]}
				]}}
			]}}`,
			byRole: map[string][]rbacv1.Subject{
				"view": {{Kind: "User", Name: "alice"}},
				"edit": {{Kind: "Group", Name: "team"}},
			},
			bySubject: map[string]map[string]TenantPermission{
				"User":  {"alice": {ClusterRoles: []string{"view"}}},
				"Group": {"team": {ClusterRoles: []string{"edit"}}},
			},
		},
		{
			name: "merge and deduplicate all sources while keeping subject kinds distinct",
			tenant: `{
				"status":{"owners":[
					{"kind":"User","name":"team","clusterRoles":["view","view"]},
					{"kind":"Group","name":"team","clusterRoles":["view"]}
				]},
				"spec":{
					"additionalRoleBindings":[
						{"clusterRoleName":"view","subjects":[{"kind":"User","name":"team"}]},
						{"clusterRoleName":"edit","subjects":[{"kind":"Group","name":"team"}]}
					],
					"rules":[
						{"permissions":{"bindings":[
							{"clusterRoleName":"view","subjects":[{"kind":"User","name":"team"},{"kind":"Group","name":"team"}]},
							{"clusterRoleName":"edit","subjects":[{"kind":"Group","name":"team"}]},
							{"clusterRoleName":"deploy","subjects":[{"kind":"User","name":"team"},{"kind":"Group","name":"team"}]}
						]}},
						{"permissions":{"bindings":[
							{"clusterRoleName":"deploy","subjects":[{"kind":"User","name":"team"}]}
						]}}
					]
				}
			}`,
			byRole: map[string][]rbacv1.Subject{
				"view":   {{Kind: "User", Name: "team"}, {Kind: "Group", Name: "team"}},
				"edit":   {{Kind: "Group", Name: "team"}},
				"deploy": {{Kind: "User", Name: "team"}, {Kind: "Group", Name: "team"}},
			},
			bySubject: map[string]map[string]TenantPermission{
				"User":  {"team": {ClusterRoles: []string{"view", "deploy"}}},
				"Group": {"team": {ClusterRoles: []string{"view", "edit", "deploy"}}},
			},
		},
		{
			name:   "owners without roles retain their group membership",
			tenant: `{"status":{"owners":[{"kind":"Group","name":"team"}]}}`,
			byRole: map[string][]rbacv1.Subject{},
			bySubject: map[string]map[string]TenantPermission{
				"Group": {"team": {ClusterRoles: []string{}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &capsulev1beta2.Tenant{}
			require.NoError(t, json.Unmarshal([]byte(tt.tenant), tenant))
			original := tenant.DeepCopy()

			require.Equal(t, tt.byRole, GetClusterRolePermissions(tenant))
			require.Equal(t, tt.bySubject, GetTenantPermissions(tenant))
			require.Equal(t, tt.bySubject["Group"], GetTenantGroups(tenant))
			require.Equal(t, original, tenant, "permission lookup must not mutate the tenant")
		})
	}
}
