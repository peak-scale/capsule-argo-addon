package v1alpha1

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

// Create enums for ArgoCD Permissions Actions
type ArgoAction string

const (
	ActionGet      ArgoAction = "get"
	ActionCreate   ArgoAction = "create"
	ActionUpdate   ArgoAction = "update"
	ActionDelete   ArgoAction = "delete"
	ActionSync     ArgoAction = "sync"
	ActionAction   ArgoAction = "action"
	ActionOverride ArgoAction = "override"
	ActionInvoke   ArgoAction = "invoke"
	ActionWildcard ArgoAction = "*"
)

// Create enums for ArgoCD Permissions Resources
type ArgoResource string

const (
	ResourceApplications    ArgoResource = "applications"
	ResourceApplicationSets ArgoResource = "applicationsets"
	ResourceClusters        ArgoResource = "clusters"
	ResourceProjects        ArgoResource = "projects"
	ResourceRepositories    ArgoResource = "repositories"
	ResourceAccounts        ArgoResource = "accounts"
	ResourceCertificates    ArgoResource = "certificates"
	ResourceGpgKeys         ArgoResource = "gpgkeys"
	ResourceLogs            ArgoResource = "logs"
	ResourceExec            ArgoResource = "exec"
	ResourceExtensions      ArgoResource = "extensions"
	ResourceWildcard        ArgoResource = "*"
)

// Create enums for ArgoCD Permissions Resources
type ArgoVerb string

const (
	VerbAllow ArgoVerb = "allow"
	VerbDeny  ArgoVerb = "deny"
)

type ArgocdPolicyDefinition struct {
	// Name for permission mapping
	Resource ArgoResource `json:"resource,omitempty"`

	// Allowed actions for this permission. You may specify multiple actions. To allow all actions use "*"
	// +kubebuilder:default={get}
	Action []string `json:"action,omitempty"`

	// Verb for this permission (can be allow, deny)
	// +kubebuilder:default=allow
	Verb ArgoVerb `json:"verb,omitempty"`

	// You may specify a custom path for the resource. The available path for argo is <app-project>/<app-ns>/<app-name>
	// however <app-project> is already set to the argocd project name. Therefor you can only add <app-ns>/<app-name>
	// +kubebuilder:default="*"
	Path string `json:"path,omitempty"`
}
