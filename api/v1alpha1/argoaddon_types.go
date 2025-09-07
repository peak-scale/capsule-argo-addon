// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoAddonSpec defines the desired state of ArgoAddon.
type ArgoAddonSpec struct {
	// When force is enabled, appprojects which already exist with the same name as a tenant will be adopted
	// and overwritten. When disabled the appprojects will not be changed or adopted.
	// This is true for any other resource as well. This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=false
	Force bool `json:"force"`
	// When decouple is enabled, appprojects are preserved even in the case when the origin tenant is deleted.
	// This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=true
	Decouple bool `json:"decouple"`
	// All appprojects, which are collected by this controller, are set into ready-only mode
	// That means only properties from matching translators are respected. Any changes from users are
	// overwritten. This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=false
	ReadOnly bool `json:"readonly"`

	// Argo configuration
	Argo ControllerArgoCDConfig `json:"argo"`

	// Allows the creation of argo repository secrets which are then replicated to the argocd namespace.
	// This makes sense when users create repository via gitops and don't have access to the GUI (or where you prevent them from doing that on the GUI)
	//+kubebuilder:default=false
	AllowRepositoryCreation bool `json:"allowRepositoryCreation"`

	// Translator selector. Only translators matching this selector will be used for this controller, if empty all translators will be used.
	// +optional
	// TranslatorSelector *metav1.LabelSelector `json:"translatorSelector,omitempty"`
}

// Controller Configuration for ArgoCD.
type ControllerArgoCDConfig struct {
	// Namespace where the ArgoCD instance is running
	// +kubebuilder:default=argocd
	Namespace string `json:"namespace,omitempty"`

	// Name of the ArgoCD rbac configmap (required for the controller)
	// +kubebuilder:default=argocd-rbac-cm
	RBACConfigMap string `json:"rbacConfigMap,omitempty"`

	// If you are not using the capsule-proxy integration this destination is registered
	// for each appproject.
	// +kubebuilder:default="https://kubernetes.default.svc"
	Destination string `json:"destination,omitempty"`

	// This is a feature which will be released with argocd +v2.13.0
	// If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha
	// +kubebuilder:default=true
	DestinationServiceAccounts bool `json:"destinationServiceAccounts,omitempty"`

	// Default Namespace to create ServiceAccounts used by arog-cd
	// The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
	// It's best to have a dedicated namespace for these serviceaccounts
	ServiceAccountNamespace string `json:"serviceAccountNamespace"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ArgoAddon is the Schema for the ArgoAddons API.
type ArgoAddon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoAddonSpec   `json:"spec,omitempty"`
	Status ArgoAddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArgoAddonList contains a list of ArgoAddon.
type ArgoAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArgoAddon{}, &ArgoAddonList{})
}
