package template

import (
	"bytes"
	"fmt"
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
	}, v1alpha1.ArgoAddonSpec{
		Proxy: v1alpha1.ControllerCapsuleProxyConfig{
			Enabled: true,
		},
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
			},

			"argocd": map[string]interface{}{
				"namespace": "argocd-namespace",
			},
		},
	}

	fmt.Printf("Expected: %#v\n", expectedMap)
	fmt.Printf("Got: %#v\n", renderedMap)

	// Deep check the result against the expected map
	if !reflect.DeepEqual(renderedMap, expectedMap) {
		t.Errorf("expected %+v, got %+v", expectedMap, renderedMap)
	}
}
