package argo

import (
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

const (
	// FinalizerName is the finalizer name for the translator
	FinalizerPrefix = "translator.addons.projectcapsule.dev/"
)

func TranslatorFinalizer(translator *v1alpha1.ArgoTranslator) string {
	return FinalizerPrefix + translator.Name
}

// Gets the template context
func (i *TenancyController) TranslatorContext(cluster string, translator *v1alpha1.ArgoTranslator, config *v1alpha1.ArgoAddonSpec, tenant *capsulev1beta2.Tenant) interface{} {

	ctx := map[string]interface{}{
		"Tenant": map[string]interface{}{
			"Name":       tenant.Name,
			"Namespaces": tenant.Status.Namespaces,
			"Object":     utils.ConvertStructToMap(tenant),
		},
		"Config":   utils.ConvertStructToMap(i.Settings.Get()),
		"Endpoint": cluster,
	}

	return ctx
}
