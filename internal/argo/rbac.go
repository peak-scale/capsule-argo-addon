// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package argo

import (
	"fmt"

	"github.com/argoproj/argo-cd/v2/util/rbac"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	v1 "k8s.io/api/rbac/v1"

	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
)

// Validates entire CSV policy and returns an error if it is invalid
func ValidateCSV(csv string) error {
	return rbac.ValidatePolicy(csv)
}

// Converts the ArgoCD Project Policy Definition to a string (common argo)
func PolicyString(policy string, tenant string, argopolicy addonsv1alpha1.ArgocdPolicyDefinition) (result string) {

	for _, action := range argopolicy.Action {
		path := argopolicy.Path
		if tenant != "" {
			if path != "" {
				path = fmt.Sprintf("%s/%s", tenant, path)
			} else {
				path = tenant
			}
		}

		// Accumulate each formatted string into the result
		result += fmt.Sprintf(
			"p, %s,%s,%s,%s,%s\n",
			policy,              // Project name
			argopolicy.Resource, // Resource (enum)
			action,              // Action (enum)
			path,                // Path (string)
			argopolicy.Verb,     // Verb (enum)
		)
	}

	return
}

// Converts the ArgoCD Project Policy Definition to a string
func BindingString(subject v1.Subject, role string) string {
	return fmt.Sprintf(
		"g, %s, %s\n",
		subject.Name,
		role,
	)
}

// Adds Default Policies (So Users can have basic interractions with the project)
func DefaultPolicies(tenant *capsulev1beta2.Tenant, destination string) (result []string) {
	// Read-Only Policy
	result = append(result, PolicyString(DefaultPolicyReadOnly(tenant),
		tenant.Name,
		addonsv1alpha1.ArgocdPolicyDefinition{
			Resource: "projects",
			Action:   []string{"get"},
			Verb:     "allow",
		}))
	result = append(result, PolicyString(DefaultPolicyReadOnly(tenant),
		tenant.Name,
		addonsv1alpha1.ArgocdPolicyDefinition{
			Resource: "projects",
			Action:   []string{"list"},
			Verb:     "allow",
		}))

	result = append(result, PolicyString(DefaultPolicyOwner(tenant),
		tenant.Name,
		addonsv1alpha1.ArgocdPolicyDefinition{
			Resource: "projects",
			Action:   []string{"update"},
			Verb:     "allow",
		}))

	result = append(result, PolicyString(DefaultPolicyReadOnly(tenant),
		tenant.Name,
		addonsv1alpha1.ArgocdPolicyDefinition{
			Resource: "clusters",
			Action:   []string{"get"},
			Verb:     "allow",
			Path:     "*",
		}))
	result = append(result, PolicyString(DefaultPolicyReadOnly(tenant),
		tenant.Name,
		addonsv1alpha1.ArgocdPolicyDefinition{
			Resource: "clusters",
			Action:   []string{"list"},
			Verb:     "allow",
			Path:     "*",
		}))

	// Allow getting specifc destination
	if destination != "" {
		result = append(result, PolicyString(DefaultPolicyReadOnly(tenant),
			destination,
			addonsv1alpha1.ArgocdPolicyDefinition{
				Resource: "clusters",
				Action:   []string{"get"},
				Verb:     "allow",
			}))
	}

	return result
}

// Default Policy for Tenant Owners
func DefaultPolicyOwner(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("caa:role:%s:owner", tenant.Name)
}

// Default Policy for Tenant Read-Only
func DefaultPolicyReadOnly(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("caa:role:%s:read-only", tenant.Name)
}

// Default Policy for Tenant Read-Only
func TenantPolicy(tenant *capsulev1beta2.Tenant, policyName string) string {
	return fmt.Sprintf("role:%s:%s", tenant.Name, policyName)
}
