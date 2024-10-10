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

	"dario.cat/mergo"
	"gopkg.in/yaml.v3"
)

// Get Combined Configuration from structured and Template
func (t *ArgocdProjectProperties) GetConfig(
	data interface{},
	funcmap template.FuncMap,
) (props ArgocdProjectStructuredProperties, err error) {
	properties, err := t.RenderTemplate(data, funcmap)
	if err != nil {
		fmt.Println("Error rendering template:", err)
		return
	}

	// Merge the structured properties with the template properties
	props, err = MergeStructuredProperties(t.Structured, *properties)

	return
}

// Field templating for the ArgoCD project properties. Needs to unmarshal in json, because of the json tags from argocd
func (t *ArgocdProjectProperties) RenderTemplate(
	data interface{},
	funcmap template.FuncMap,
) (*ArgocdProjectStructuredProperties, error) {
	// Parse and execute the template using sprig functions
	tmpl, err := template.New("argoTemplate").Funcs(funcmap).Parse(t.Template)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	var rendered bytes.Buffer
	err = tmpl.Execute(&rendered, data)
	if err != nil {
		return nil, fmt.Errorf("error executing template: %w", err)
	}

	yamlBytes := rendered.Bytes()
	jsonBytes, err := yamlToJSON(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("error converting yaml to json: %w", err)
	}

	var structuredProperties ArgocdProjectStructuredProperties
	err = json.Unmarshal(jsonBytes, &structuredProperties)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling json: %w", err)
	}

	return &structuredProperties, nil
}

// Merge two properties
func MergeStructuredProperties(prop1, prop2 ArgocdProjectStructuredProperties) (ArgocdProjectStructuredProperties, error) {
	merged := prop1

	// Use mergo.Merge to merge prop2 into merged (prop1), with overwrite enabled
	if err := mergo.Merge(&merged, prop2, mergo.WithOverride); err != nil {
		return ArgocdProjectStructuredProperties{}, err
	}

	return merged, nil
}

func yamlToJSON(yamlBytes []byte) ([]byte, error) {
	var yamlObj interface{}
	err := yaml.Unmarshal(yamlBytes, &yamlObj)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling yaml: %w", err)
	}

	// Convert the YAML object to JSON
	jsonBytes, err := json.Marshal(yamlObj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling to json: %w", err)
	}

	return jsonBytes, nil
}
