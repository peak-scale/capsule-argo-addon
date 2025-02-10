// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

func ConfigContext(translator *v1alpha1.ArgoTranslator, config *v1alpha1.ArgoAddonSpec, tenant *capsulev1beta2.Tenant) interface{} {

	ctx := map[string]interface{}{
		"Tenant": map[string]interface{}{
			"Name":       tenant.Name,
			"Namespaces": tenant.Status.Namespaces,
			"Object":     utils.Mapify(tenant),
		},
		"Config":   utils.Mapify(config),
		"Endpoint": config.GetClusterDestination(tenant),
	}

	return ctx
}
