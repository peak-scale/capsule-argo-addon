// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"fmt"

	argocdapi "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/reflection"
	"github.com/peak-scale/capsule-argo-addon/internal/stores"
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Gets the API Server given via Rest-Config.
func (i *Reconciler) RetrieveAPIServerURL() string {
	return i.Rest.Host
}

// Decouple a Tenant from an Object.
func (i *Reconciler) DecoupleTenant(obj client.Object, tenant *capsulev1beta2.Tenant) (err error) {
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
	translator *configv1alpha1.ArgocdProjectProperties,
	settings *stores.ConfigStore,
) (cfg *configv1alpha1.ArgocdProjectStructuredProperties, err error) {
	if translator == nil {
		return nil, nil
	}

	structured, templated, err := translator.GetConfigs(
		tpl.ConfigContext(settings.Get(), tenant), tpl.ExtraFuncMap())
	if err != nil {
		return
	}

	switch {
	case structured == nil && templated == nil:
		return nil, nil
	case structured == nil:
		structured = &configv1alpha1.ArgocdProjectStructuredProperties{}
	case templated == nil:
		templated = &configv1alpha1.ArgocdProjectStructuredProperties{}
	}

	if err = reflection.Merge(templated, structured); err != nil {
		return nil, fmt.Errorf("failed to merge translator spec: %w", err)
	}

	return templated, nil
}

// Remove Translator for tenant.
func RemoveTranslatorForTenant(
	translator *configv1alpha1.ArgoTranslator,
	tenant *capsulev1beta2.Tenant,
	appproject *argocdapi.AppProject,
	settings *stores.ConfigStore,
) (err error) {
	finalizer := meta.TranslatorFinalizer(translator.Name)
	if controllerutil.ContainsFinalizer(appproject, finalizer) {
		controllerutil.RemoveFinalizer(appproject, finalizer)
	}

	return SubstractTranslatorSpec(translator, tenant, appproject, settings)
}

// Remove Translator for tenant.
func SubstractTranslatorSpec(
	translator *configv1alpha1.ArgoTranslator,
	tenant *capsulev1beta2.Tenant,
	appproject *argocdapi.AppProject,
	settings *stores.ConfigStore,
) (err error) {
	// Verify if currently Something is serving
	stat := translator.GetTenantStatus(tenant)
	if stat == nil {
		return nil
	}

	if stat.Serving == nil {
		return nil
	}

	if translator.Spec.ProjectSettings == nil {
		translator.Spec.ProjectSettings = &configv1alpha1.ArgocdProjectProperties{}
	}

	if translator.Spec.ProjectSettings == stat.Serving {
		return nil
	}

	cfg, err := GetMergedConfig(
		tenant,
		stat.Serving,
		settings,
	)
	if err != nil {
		return err
	}

	if cfg == nil {
		return nil
	}

	// Specification
	reflection.Subtract(&appproject.Spec, &cfg.ProjectSpec)

	// Metadata
	if cfg.ProjectMeta == nil {
		return nil
	}
	// Remove transformer labels from the approject
	for key, value := range cfg.ProjectMeta.Labels {
		if currentValue, ok := appproject.Labels[key]; ok {
			if currentValue == value {
				delete(appproject.Labels, key)
			}
		}
	}
	// Remove transformer annotations from the approject
	for key, value := range cfg.ProjectMeta.Annotations {
		if currentValue, ok := appproject.Annotations[key]; ok {
			if currentValue == value {
				delete(appproject.Annotations, key)
			}
		}
	}

	// Remove Finalizers from the approject

	for _, finalizer := range cfg.ProjectMeta.Finalizers {
		if controllerutil.ContainsFinalizer(appproject, finalizer) {
			controllerutil.RemoveFinalizer(appproject, finalizer)
		}
	}

	return nil
}
