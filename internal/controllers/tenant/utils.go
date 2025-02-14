// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"fmt"

	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/reflection"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Gets the API Server given via Rest-Config.
func (i *TenancyController) RetrieveAPIServerURL() string {
	return i.Rest.Host
}

// Decouple a Tenant from an Object.
func (i *TenancyController) DecoupleTenant(obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
	if err = meta.RemoveDynamicTenantOwnerReference(obj, tenant); err != nil {
		return
	}

	// Remove Tracking Labels
	obj.SetLabels(meta.TranslatorRemoveTenantLabels(obj.GetLabels()))

	return
}

// Decouple a Tenant from an Object.
func GetMergedConfig(
	tenant *capsulev1beta2.Tenant,
	translator *configv1alpha1.ArgoTranslator,
	settings *stores.ConfigStore,
) (cfg *configv1alpha1.ArgocdProjectStructuredProperties, err error) {
	structured, templated, err := translator.Spec.ProjectSettings.GetConfigs(
		tpl.ConfigContext(settings.Get(), tenant), tpl.ExtraFuncMap())
	if err != nil {
		return
	}

	if err = reflection.Merge(templated, structured); err != nil {
		return nil, fmt.Errorf("failed to merge translator spec: %w", err)
	}

	return templated, nil
}
