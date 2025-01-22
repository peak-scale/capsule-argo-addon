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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoAddonSpec defines the desired state of ArgoAddon
type ArgoAddonSpec struct {
	// When force is enabled, appprojects which already exist with the same name as a tenant will be adopted
	// and overwritten. When disabled the appprojects will not be changed or adopted.
	// This is true for any other resource as well. This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=false
	Force bool `json:"force"`
	// When decouple is enabled, appprojects are preserved even in the case when the origin tenant is deleted.
	// This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=false
	Decouple bool `json:"decouple"`
	// All appprojects, which are collected by this controller, are set into ready-only mode
	// That means only properties from matching translators are respected. Any changes from users are
	// overwritten. This can also be set on a per-tenant basis via annotations.
	//+kubebuilder:default=false
	ReadOnly bool `json:"readonly"`

	// Capsule-Proxy configuration for the controller
	Proxy ControllerCapsuleProxyConfig `json:"proxy"`

	// Argo configuration
	Argo ControllerArgoCDConfig `json:"argo"`

	// Translator selector. Only translators matching this selector will be used for this controller, if empty all translators will be used.
	// +optional
	//TranslatorSelector *metav1.LabelSelector `json:"translatorSelector,omitempty"`
}

// Controller Configuration for ArgoCD
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

	// +optional
	//DefaultServerNamespace string `json:"defaultNamespace,omitempty"`

	// This is a feature which will be released with argocd +v2.13.0
	// If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha
	// +kubebuilder:default=false
	DestinationServiceAccounts bool `json:"destinationServiceAccounts,omitempty"`

	// Default Namespace to create ServiceAccounts used by arog-cd
	// The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
	// It's best to have a dedicated namespace for these serviceaccounts
	ServiceAccountNamespace string `json:"serviceAccountNamespace"`
}

// Controller Configuration for ArgoCD
type ControllerCapsuleProxyConfig struct {
	// Enable the capsule-proxy integration.
	// This automatically creates services for tenants and registers them as destination
	// on the argo appproject.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// Name of the capsule-proxy service
	// +kubebuilder:default=capsule-proxy
	CapsuleProxyServiceName string `json:"serviceName,omitempty"`

	//  Namespace where the capsule-proxy service is running
	// +kubebuilder:default=capsule-system
	CapsuleProxyServiceNamespace string `json:"serviceNamespace,omitempty"`

	// Port of the capsule-proxy service
	// +kubebuilder:default=9001
	CapsuleProxyServicePort int32 `json:"servicePort,omitempty"`

	// Port of the capsule-proxy service
	// +kubebuilder:default=true
	CapsuleProxyTLS bool `json:"tls,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ArgoAddon is the Schema for the ArgoAddons API
type ArgoAddon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoAddonSpec   `json:"spec,omitempty"`
	Status ArgoAddonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArgoAddonList contains a list of ArgoAddon
type ArgoAddonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoAddon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArgoAddon{}, &ArgoAddonList{})
}
