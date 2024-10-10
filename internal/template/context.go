package template

import (
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Gets the template context
func TranslatorContext(cluster string, tenant *capsulev1beta2.Tenant) (data map[string]interface{}) {
	data = map[string]interface{}{
		"Name":     tenant.Name,
		"Tenant":   tenant,
		"Endpoint": cluster,
	}
	return data
}
