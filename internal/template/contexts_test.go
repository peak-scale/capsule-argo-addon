// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"text/template"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This test should prove that subkeys can be addressed in the template
func TestRenderTemplateAndUnmarshal(t *testing.T) {

	// Define your YAML template
	yamlTemplate := `
tenant:
  name: "{{ .Tenant.Name }}"
  namespaces:
  {{- range .Tenant.Namespaces }}
    - {{ . }}
  {{- end }}
    - {{ .Config.Argo.Namespace }}
config:
  proxy: {{- toYaml .Config.Proxy | nindent 4 }}
  argocd:
    namespace: "{{ .Config.Argo.Namespace }}"
`

	// Source the context (Mocking the required structs)
	tplCtx := ConfigContext("example-cluster", &v1alpha1.ArgoTranslator{
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/type": "prod",
				},
			},
		},
	}, &v1alpha1.ArgoAddonSpec{
		Argo: v1alpha1.ControllerArgoCDConfig{
			Namespace: "argocd-namespace",
		},
	}, &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-tenant",
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: []capsulev1beta2.OwnerSpec{
				{
					Kind: "User",
					Name: "example-user",
				},
				{
					Kind: "Group",
					Name: "example-group",
				},
			},
		},
		Status: capsulev1beta2.TenantStatus{
			Namespaces: []string{"namespace1", "namespace2"},
		},
	})

	// Run the template
	tmpl, err := template.New("test").Funcs(ExtraFuncMap()).Parse(yamlTemplate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tplCtx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unmarshal the result into a map
	var renderedMap map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &renderedMap)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling YAML: %v", err)
	}

	// Define the expected result as a map
	expectedMap := map[string]interface{}{
		"tenant": map[string]interface{}{
			"name": "example-tenant",
			"namespaces": []interface{}{
				"namespace1",
				"namespace2",
				"argocd-namespace",
			},
		},
		"config": map[string]interface{}{
			"proxy": map[string]interface{}{
				"Enabled":                      true,
				"CapsuleProxyServiceName":      "",
				"CapsuleProxyServicePort":      0,
				"CapsuleProxyServiceNamespace": "",
				"ServiceAccountNamespace":      "",
				"CapsuleProxyTLS":              false,
			},

			"argocd": map[string]interface{}{
				"namespace": "argocd-namespace",
			},
		},
	}

	// Deep check the result against the expected map
	if !reflect.DeepEqual(renderedMap, expectedMap) {
		t.Errorf("expected %+v, got %+v", expectedMap, renderedMap)
	}
}

// TestRenderContextToMarkdown renders the template context into a Markdown file
func TestRenderContextToMarkdown(t *testing.T) {
	// Load markdown template from file
	markdownTemplate, err := os.ReadFile("../../docs/templating.tpl")
	if err != nil {
		log.Fatalf("failed to read template file: %v", err)
	}

	// Load template context
	tplCtx := ConfigContext("example-cluster", &v1alpha1.ArgoTranslator{
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/type": "prod",
				},
			},
		},
	}, &v1alpha1.ArgoAddonSpec{
		Argo: v1alpha1.ControllerArgoCDConfig{
			RBACConfigMap: "argocd-rbac-cm",
			Namespace:     "argocd",
		},
	}, &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-tenant",
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: []capsulev1beta2.OwnerSpec{
				{
					Kind: "User",
					Name: "example-user",
				},
				{
					Kind: "Group",
					Name: "example-group",
				},
			},
		},
		Status: capsulev1beta2.TenantStatus{
			Namespaces: []string{"namespace1", "namespace2"},
		},
	})

	// Parse the markdown template
	tmpl, err := template.New("markdown").Funcs(ExtraFuncMap()).Parse(string(markdownTemplate))
	if err != nil {
		t.Fatalf("failed to parse markdown template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tplCtx); err != nil {
		t.Fatalf("failed to execute markdown template: %v", err)
	}

	// Write the rendered Markdown to a file
	fileName := "../../docs/templating.md"
	err = os.WriteFile(fileName, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("failed to write Markdown file: %v", err)
	}

	fmt.Printf("Markdown file %s created successfully.\n", fileName)
}
