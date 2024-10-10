package roles

import (
	"bytes"
	"text/template"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

type ArgoProjectRole struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Groups      []string `json:"groups"`
	Policies    []string `json:"policies"`
}

func ArgoOwnerPolicies(tenantName string) []string {
	return []string{
		"p, proj:" + tenantName + ":owners, applicationsets, *, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":owners, applications, *, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":owners, logs, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":owners, exec, create, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":owners, repositories, *, " + tenantName + "/*, allow",
	}
}

func ArgoMaintainerPolicies(tenantName string) []string {
	return []string{
		"p, proj:" + tenantName + ":maintainers, applicationsets, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applicationsets, create, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applicationsets, update, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applicationsets, sync, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applicationsets, override, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applications, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applications, create, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applications, update, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applications, sync, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, applications, override, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, repositories, *, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, logs, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":maintainers, exec, create, " + tenantName + "/*, allow",
	}
}

func ArgoOperatorPolicies(tenantName string) []string {
	return []string{
		"p, proj:" + tenantName + ":operators, applicationsets, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":operators, applicationsets, sync, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":operators, applications, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":operators, applications, sync, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":operators, logs, get, " + tenantName + "/*, allow",
	}
}

func ArgoViewerPolicies(tenantName string) []string {
	return []string{
		"p, proj:" + tenantName + ":viewers, applications, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":viewers, logs, get, " + tenantName + "/*, allow",
		"p, proj:" + tenantName + ":viewers, repositories, get, " + tenantName + "/*, allow",
	}
}

func ArgoTenantCSV(cluster string, tenant *capsulev1beta2.Tenant) (string, error) {

	data := map[string]interface{}{
		"Tenant":   &tenant,
		"Endpoint": cluster,
	}

	// Create a new template and parse the text
	tmpl, err := template.New("rbac").Parse(ArgoCSVTemplate)
	if err != nil {
		return "", err
	}

	// Execute the template with the data
	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
