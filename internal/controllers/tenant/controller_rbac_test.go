// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func TestReflectArgoCSVResolvedPermissions(t *testing.T) {
	tenant := &capsulev1beta2.Tenant{}
	require.NoError(t, json.Unmarshal([]byte(`{
		"metadata":{"name":"solar"},
		"spec":{
			"owners":[{"kind":"User","name":"unresolved","clusterRoles":["edit"]}],
			"additionalRoleBindings":[{"clusterRoleName":"edit","subjects":[{"kind":"User","name":"legacy"}]}],
			"rules":[{"permissions":{"bindings":[{"clusterRoleName":"edit","subjects":[
				{"kind":"Group","name":"developers"},
				{"kind":"User","name":"resolved"},
				{"kind":"ServiceAccount","name":"robot","namespace":"default"}
			]}]}}]
		},
		"status":{"owners":[{"kind":"User","name":"resolved","clusterRoles":["edit"]}]}
	}`), tenant))

	reconciler := &Reconciler{Settings: stores.NewConfigStore()}
	translators := []*configv1alpha1.ArgoTranslator{{
		Spec: configv1alpha1.ArgoTranslatorSpec{
			ProjectRoles: []configv1alpha1.ArgocdProjectRolesTranslator{{
				Name: "editor", ClusterRoles: []string{"edit"}, Owner: true,
				Policies: []configv1alpha1.ArgocdPolicyDefinition{{
					Resource: "applications", Action: []string{"get"}, Verb: "allow",
				}},
			}},
		},
	}}
	csv, err := reconciler.reflectArgoCSV(logr.Discard(), tenant, translators)
	require.NoError(t, err)

	for _, subject := range []rbacv1.Subject{
		{Kind: "User", Name: "resolved"},
		{Kind: "User", Name: "legacy"},
		{Kind: "Group", Name: "developers"},
	} {
		for _, role := range []string{
			argo.TenantPolicy(tenant, "editor"), argo.DefaultPolicyReadOnly(tenant), argo.DefaultPolicyOwner(tenant),
		} {
			require.Equal(t, 1, strings.Count(csv, argo.BindingString(subject, role)), "binding for %s to %s", subject.Name, role)
		}
	}
	require.NotContains(t, csv, "unresolved")
	require.NotContains(t, csv, "robot")

	// Losing resolved ownership or a rule binding removes the corresponding Argo access.
	tenant.Status.Owners = nil
	tenant.Spec.Rules[0].Permissions.Bindings = nil
	csv, err = reconciler.reflectArgoCSV(logr.Discard(), tenant, translators)
	require.NoError(t, err)
	require.NotContains(t, csv, "resolved")
	require.NotContains(t, csv, "developers")
	require.Contains(t, csv, argo.BindingString(rbacv1.Subject{Kind: "User", Name: "legacy"}, argo.TenantPolicy(tenant, "editor")))
}
