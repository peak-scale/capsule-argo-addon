// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// ArgoTranslatorStatus defines the observed state of ArgoTranslator.
type ArgoTranslatorStatus struct {
	// List of tenants selected by this translator
	Tenants []TenantStatus `json:"tenants,omitempty"`
	// Amount of tenants selected by this translator
	Size uint `json:"size,omitempty"`
	// Ready field indicating overall readiness of the translator
	Ready string `json:"ready,omitempty"`
}

type TenantStatus struct {
	// List of tenants selected by this translator
	Name string `json:"name,omitempty"`
	// UID of the tracked Tenant to pin point tracking
	k8stypes.UID `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid"`
	// Conditions represent the latest available observations of an object's state
	Condition metav1.Condition `json:"condition,omitempty"`
	// Serving  Settings for this Tenant
	Serving *ArgocdProjectProperties `json:"serving,omitempty"`
}
