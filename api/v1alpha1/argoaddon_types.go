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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArgoAddonSpec defines the desired state of ArgoAddon
type ArgoAddonSpec struct {
	// When force is enabled, approjects which already exist with the same name as a tenant will be adopted
	// and overwritten. When disabled the approjects will not be changed or adopted.
	// This is true for any other resource as well
	//+kubebuilder:default=false
	Force bool `json:"force"`

	// Capsule-Proxy configuration for the controller
	//+kubebuilder:default={}
	Proxy ControllerCapsuleProxyConfig `json:"proxy,omitempty"`

	// ArgoCD configuration
	// +kubebuilder:default={namespace: argocd, rbacConfigMap: argocd-rbac-cm}
	Argo ControllerArgoCDConfig `json:"argo,omitempty"`

	// Translator selector. Only translators matching this selector will be used for this controller, if empty all translators will be used.
	// +optional
	//TranslatorSelector *metav1.LabelSelector `json:"translatorSelector,omitempty"`
}

// Controller Configuration for ArgoCD
type ControllerCapsuleProxyConfig struct {
	// Enable the capsule-proxy integration. This automatically creates ServiceAccounts for tenants and registers them as destination
	// on the argo appproject.
	// +kubebuilder:default=true
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

	// Default Namespace to create ServiceAccounts in for proxy access.
	// Can be overwritten on tenant-basis
	ServiceAccountNamespace string `json:"serviceAccountNamespace,omitempty"`
}

// Controller Configuration for ArgoCD
type ControllerArgoCDConfig struct {
	// Namespace where the ArgoCD instance is running
	Namespace string `json:"namespace,omitempty"`

	// Name of the ArgoCD rbac configmap (required for the controller)
	RBACConfigMap string `json:"rbacConfigMap,omitempty"`
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
