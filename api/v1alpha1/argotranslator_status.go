/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	k8stypes "k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoTranslatorStatus defines the observed state of ArgoTranslator
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
}
