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
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArgoTranslatorSpec defines the desired state of ArgoTranslator
type ArgoTranslatorSpec struct {
	// Selector to match tenants which are used for the translator
	Selector *metav1.LabelSelector `json:"selector,omitempty"`

	// Application-Project Roles for the tenant
	//+kubebuilder:optional
	ProjectRoles []ArgocdProjectRolesTranslator `json:"roles,omitempty"`

	// Additional settings for the argocd project
	//+kubebuilder:optional
	ProjectSettings ArgocdProjectProperties `json:"settings,omitempty"`

	// In this field you can define custom policies. It must result in a valid argocd policy format (CSV)
	// You can use Sprig Templating with this field
	//+kubebuilder:optional
	CustomPolicy string `json:"customPolicy,omitempty"`
}

// Define Permission mappings for an ArogCD Project
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
	//+kubebuilder:optional
	Structured ArgocdProjectStructuredProperties `json:"structured,omitempty"`

	// Use a template to generate to argo project settings
	//+kubebuilder:optional
	Template string `json:"template,omitempty"`
}

type ArgocdProjectStructuredProperties struct {
	// Project Metadata
	//+kubebuilder:optional
	ProjectMeta ArgocdProjectPropertieMeta `json:"meta,omitempty"`

	// Application Project Spec (Upstream ArgoCD)
	//+kubebuilder:optional
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

// ArgoTranslator is the Schema for the argotranslators API
type ArgoTranslator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoTranslatorSpec   `json:"spec,omitempty"`
	Status ArgoTranslatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ArgoTranslatorList contains a list of ArgoTranslator
type ArgoTranslatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoTranslator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArgoTranslator{}, &ArgoTranslatorList{})
}
