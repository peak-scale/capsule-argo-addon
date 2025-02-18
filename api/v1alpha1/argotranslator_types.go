// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoTranslatorSpec defines the desired state of ArgoTranslator.
type ArgoTranslatorSpec struct {
	// Selector to match tenants which are used for the translator
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Application-Project Roles for the tenant
	ProjectRoles []ArgocdProjectRolesTranslator `json:"roles,omitempty"`

	// Additional settings for the argocd project
	ProjectSettings *ArgocdProjectProperties `json:"settings,omitempty"`

	// In this field you can define custom policies. It must result in a valid argocd policy format (CSV)
	// You can use Sprig Templating with this field
	CustomPolicy string `json:"customPolicy,omitempty"`
}

// Define Permission mappings for an ArogCD Project.
type ArgocdProjectRolesTranslator struct {
	// Name for permission mapping
	Name string `json:"name,omitempty"`

	// TenantRoles selects tenant users based on their cluster roles to this Permission
	ClusterRoles []string `json:"clusterRoles,omitempty"`

	// Roles are reflected in the argocd rbac configmap
	Policies []ArgocdPolicyDefinition `json:"policies,omitempty"`

	// Define if the selected users are owners of the appproject. Being owner allows the users
	// to update the project and effectively manage everything. By default the selected users get
	// read-only access to the project.
	// +kubebuilder:default=false
	Owner bool `json:"owner,omitempty"`
}

type ArgocdProjectProperties struct {
	// Structured Properties for the argocd project
	Structured *ArgocdProjectStructuredProperties `json:"structured,omitempty"`
	// Use a template to generate to argo project settings
	Template string `json:"template,omitempty"`
}

type ArgocdProjectStructuredProperties struct {
	// Project Metadata
	ProjectMeta *ArgocdProjectPropertieMeta `json:"meta,omitempty"`
	// Application Project Spec (Upstream ArgoCD)
	ProjectSpec argocdv1alpha1.AppProjectSpec `json:"spec,omitempty"`
}

type ArgocdProjectPropertieMeta struct {
	// Labels for the project
	//+kubebuilder:optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for the project
	//+kubebuilder:optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Finalizers for the project
	//+kubebuilder:optional
	Finalizers []string `json:"finalizers,omitempty"`
}

//nolint:lll
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Tenants",type="integer",JSONPath=".status.size",description="The amount of tenants being translated"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.ready",description="Indicates if all tenants were successfully translated"

// ArgoTranslator is the Schema for the argotranslators API.
type ArgoTranslator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoTranslatorSpec   `json:"spec,omitempty"`
	Status ArgoTranslatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ArgoTranslatorList contains a list of ArgoTranslator.
type ArgoTranslatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoTranslator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArgoTranslator{}, &ArgoTranslatorList{})
}
