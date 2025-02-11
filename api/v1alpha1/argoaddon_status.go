// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// ArgoAddonStatus defines the observed state of ArgoAddon.
type ArgoAddonStatus struct {
	// Last applied valid configuration
	Config ArgoAddonSpec `json:"loaded,omitempty"`
}
