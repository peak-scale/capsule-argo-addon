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
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"dario.cat/mergo"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Get Combined Configuration from structured and Template
//func (t *ArgocdProjectProperties) GetAppProjectSpec(
//	data interface{},
//	funcmap template.FuncMap,
//) (props argocdapi.AppProjectSpec, err error) {
//	props, err = t.GetConfig(data, funcmap)
//	if err != nil {
//		return
//	}
//
//	props = &argocdapi.AppProjectSpec{
//
//
//}

// Get Combined Configuration from structured and Template
func (t *ArgocdProjectProperties) GetConfig(
	data interface{},
	funcmap template.FuncMap,
) (props ArgocdProjectStructuredProperties, err error) {
	props = t.Structured

	// Get Templated config
	templated, err := t.RenderTemplate(data, funcmap)
	if err != nil {
		return props, fmt.Errorf("error executing template: %w", err)
	}

	// Use mergo.Merge to merge prop2 into merged (prop1), with overwrite enabled
	err = mergo.Merge(&props, templated, mergo.WithAppendSlice)

	return
}

// Get Combined Configuration from structured and Template
func (t *ArgocdProjectProperties) GetConfigs(
	data interface{},
	funcmap template.FuncMap,
) (structured ArgocdProjectStructuredProperties, templated ArgocdProjectStructuredProperties, err error) {
	structured = t.Structured

	// Get Templated config
	templated, err = t.RenderTemplate(data, funcmap)
	if err != nil {
		fmt.Println("Error rendering template:", err)
		return
	}

	return
}

// Field templating for the ArgoCD project properties. Needs to unmarshal in json, because of the json tags from argocd
func (t *ArgocdProjectProperties) RenderTemplate(
	data interface{},
	funcmap template.FuncMap,
) (ArgocdProjectStructuredProperties, error) {
	var structuredProperties ArgocdProjectStructuredProperties
	// Parse and execute the template using sprig functions
	tmpl, err := template.New("argoTemplate").Funcs(funcmap).Parse(t.Template)
	if err != nil {
		return structuredProperties, fmt.Errorf("error parsing template: %w", err)
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return structuredProperties, fmt.Errorf("error executing template: %w", err)
	}

	yamlBytes := rendered.Bytes()
	jsonBytes, err := utils.YamlToJSON(yamlBytes)
	if err != nil {
		return structuredProperties, fmt.Errorf("error converting yaml to json: %w", err)
	}

	err = json.Unmarshal(jsonBytes, &structuredProperties)
	if err != nil {
		return structuredProperties, fmt.Errorf("error unmarshaling json: %w", err)
	}

	return structuredProperties, nil
}

// Assign Tenants to the ArgoTranslator
func (in *ArgoTranslator) GetTenants() []TenantStatus {
	return in.Status.Tenants
}

// Just Extract the tenants
func (in *ArgoTranslator) GetTenantNames() (tnts []string) {
	for _, tnt := range in.Status.Tenants {
		tnts = append(tnts, tnt.Name)
	}

	return
}

// Assign Tenants to the ArgoTranslator
func (in *ArgoTranslator) CollectStatus() {
	in.updateTenantSize()
	in.updateReadyStatus()
}

// Assign Tenants to the ArgoTranslator
func (in *ArgoTranslator) updateTenantSize() {
	in.Status.Size = uint(len(in.Status.Tenants))
}

func (in *ArgoTranslator) updateReadyStatus() {
	in.Status.Ready = meta.ReadyCondition // Assume ready until proven otherwise
	for _, tenant := range in.Status.Tenants {
		if tenant.Condition.Status != metav1.ConditionTrue || tenant.Condition.Type != "Ready" {
			in.Status.Ready = meta.NotReadyCondition
			return // Exit early if any tenant is not ready
		}
	}
}

// Update the condition for a single Tenant
func (in *ArgoTranslator) UpdateTenantCondition(tnt TenantStatus) {
	// Check if the tenant is already present in the status
	for i, existingTenant := range in.Status.Tenants {
		if existingTenant.Name == tnt.Name {
			in.Status.Tenants[i].Condition = tnt.Condition
			in.CollectStatus()
			return
		}
	}

	// If tenant not found, append it to the list
	in.Status.Tenants = append(in.Status.Tenants, tnt)
	in.CollectStatus()
}

// Get Condition for a tenant, if no condition is present (tenant absent) returns nil
func (in *ArgoTranslator) GetTenantCondition(tnt *capsulev1beta2.Tenant) *metav1.Condition {
	for _, tenant := range in.Status.Tenants {
		if tenant.Name == tnt.Name && tenant.UID == tnt.UID {
			return &tenant.Condition
		}
	}

	return nil
}

// Removes a tenant from the ArgoTranslator Status
func (in *ArgoTranslator) RemoveTenantCondition(tnt string) {
	// Filter out the tenant with the specified name
	filteredTenants := []TenantStatus{}
	for _, tenant := range in.Status.Tenants {
		if tenant.Name != tnt {
			filteredTenants = append(filteredTenants, tenant)
		}
	}

	// Update the tenants and adjust the size
	in.Status.Tenants = filteredTenants
	in.CollectStatus()
}
