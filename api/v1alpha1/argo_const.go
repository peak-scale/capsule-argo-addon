// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// Create enums for ArgoCD Permissions Actions.
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

// Create enums for ArgoCD Permissions Resources.
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

// Create enums for ArgoCD Permissions Resources.
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
