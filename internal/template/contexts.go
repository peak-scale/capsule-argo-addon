package template

import (
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func ConfigContext(cluster string, translator *v1alpha1.ArgoTranslator, config *v1alpha1.ArgoAddonSpec, tenant *capsulev1beta2.Tenant) interface{} {

	ctx := map[string]interface{}{
		"Tenant": map[string]interface{}{
			"Name":       tenant.Name,
			"Namespaces": tenant.Status.Namespaces,
			"Object":     utils.Mapify(tenant),
		},
		"Config":   utils.Mapify(config),
		"Endpoint": cluster,
	}

	return ctx
}
