// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package argo

import argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"

// Verify if the project already has the destination (Don't respect namespace).
func ProjectHasServiceAccount(appProject *argocdv1alpha1.AppProject, sa argocdv1alpha1.ApplicationDestinationServiceAccount) bool {
	// Check if the destination already exists
	exists := false

	for _, e := range appProject.Spec.DestinationServiceAccounts {
		if e.DefaultServiceAccount == sa.DefaultServiceAccount && e.Namespace == sa.Namespace && e.Server == sa.Server {
			exists = true

			break
		}
	}

	return exists
}

// Remove a destination from the project (Don't respect namespace).
func RemoveProjectServiceaccount(appProject *argocdv1alpha1.AppProject, sa argocdv1alpha1.ApplicationDestinationServiceAccount) {
	newDestinationServiceAccounts := []argocdv1alpha1.ApplicationDestinationServiceAccount{}

	for _, e := range appProject.Spec.DestinationServiceAccounts {
		if !(e.DefaultServiceAccount == sa.DefaultServiceAccount) {
			newDestinationServiceAccounts = append(newDestinationServiceAccounts, e)
		}
	}

	appProject.Spec.DestinationServiceAccounts = newDestinationServiceAccounts
}
