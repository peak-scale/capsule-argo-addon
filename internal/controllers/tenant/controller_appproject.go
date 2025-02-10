// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	translatorctl "github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	ccaerrrors "github.com/peak-scale/capsule-argo-addon/internal/errors"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/reflection"
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Creates or updates the ArgoCD Application Project for the tenant
//
//nolint:gocyclo,gocognit,cyclop,maintidx
func (i *TenancyController) reconcileArgoProject(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
	unmatchedTranslators map[string]*configv1alpha1.ArgoTranslator,
) (err error) {
	// Initialize AppProject
	appProject := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		},
	}

	log.V(7).Info("reconciling appproject", "appproject", appProject.Name)

	// Fetch the current state of the AppProject
	gerr := i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().Argo.Namespace}, appProject)
	if gerr != nil && !k8serrors.IsNotFound(gerr) {
		return gerr
	}

	// Don't Force, When project already exists
	// Check this before bootstraping any dependencies
	if !meta.HasTenantOwnerReference(appProject, tenant) || len(meta.GetTranslatingFinalizers(appProject)) == 0 {
		if !i.Settings.Get().ForceTenant(tenant) && !k8serrors.IsNotFound(gerr) {
			log.V(1).Info("appproject already present, not overriding", "appproject", appProject.Name)

			return ccaerrrors.NewObjectAlreadyExistsError(appProject)
		}
	}

	// Collect Service-Account
	token, err := i.reconcileArgoServiceAccount(ctx, log, tenant, translators)
	if err != nil {
		return err
	}

	// Reconcile Argo Cluster
	err = i.reconcileArgoCluster(ctx, log, tenant, token, translators)
	if err != nil {
		return err
	}

	// Get Destination
	destination := i.Settings.Get().GetClusterDestination(tenant)

	// Lifecycle Approject (If marked for deletion remove finalizers)
	//nolint:nestif
	if !appProject.ObjectMeta.DeletionTimestamp.IsZero() || !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(5).Info("removing finalizers for approject", "appproject", appProject.Name)

		_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
			// Remove unmatched Translators based on finalizers
			presentTranslators := meta.GetTranslatingFinalizers(appProject)
			for _, translatorName := range presentTranslators {
				if translator, found := unmatchedTranslators[translatorName]; found {
					log.V(7).Info("removing translator config", "appproject", appProject.Name, "translator", translatorName)

					// Call RemoveTranslatorForTenant with the actual translator object
					err := translatorctl.RemoveTranslatorForTenant(log, translator, tenant, appProject, i.Settings)
					if err != nil {
						log.Error(err, "failed to remove translator", "translator", translatorName)

						return err
					}
				} else {
					log.V(3).Info(
						"removing no longer present translator finalizer",
						"appproject", appProject.Name,
						"translator", translatorName)
					controllerutil.RemoveFinalizer(appProject, meta.TranslatorFinalizer(translatorName))
				}
			}

			// Handle when the tenant is being deleted but the AppProject is decoupled
			// In this case we remove the owner reference and the tenant tracking label so the Appproject can still exist
			if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
				if i.Settings.Get().DecoupleTenant(tenant) {
					log.V(5).Info("decoupling appproject", "appproject", appProject.Name)

					if err := i.DecoupleTenant(appProject, tenant); err != nil {
						return err
					}
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	}

	// Lifecycle Approject (If no translators are present, remove the Appproject)
	if len(translators) == 0 {
		// Appproject is already absent
		if k8serrors.IsNotFound(gerr) {
			return nil
		}

		// Delete the AppProject when it's not decoupled
		if !i.Settings.Get().DecoupleTenant(tenant) {
			return i.Client.Delete(ctx, appProject)
		}

		log.V(5).Info("decoupling appproject", "appproject", appProject.Name)

		if err := i.DecoupleTenant(appProject, tenant); err != nil {
			return err
		}
	}

	log.Info("reconcile appproject", "appproject", appProject.Name)

	_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
		// Prepare metadata
		appProject.ObjectMeta.Labels = meta.TranslatorTrackingLabels(tenant)
		if appProject.ObjectMeta.Annotations == nil {
			appProject.ObjectMeta.Annotations = make(map[string]string)
		}

		appliedTranslatorsSet := make(map[string]struct{})
		translatedSpec := &argocdv1alpha1.AppProjectSpec{}

		for _, translator := range translators {
			// Get Approject Config with templating
			translatorCfg, err := translator.Spec.ProjectSettings.GetConfig(
				tpl.ConfigContext(translator, i.Settings.Get(), tenant), tpl.ExtraFuncMap())
			if err != nil {
				return err
			}

			cfg1, cfg2, err := translator.Spec.ProjectSettings.GetConfigs(
				tpl.ConfigContext(translator, i.Settings.Get(), tenant), tpl.ExtraFuncMap())
			if err != nil {
				return err
			}

			log.V(10).Info(
				"translator-config",
				"translator", translator.Name,
				"appproject", appProject.Name,
				"structured", cfg1,
				"templated", cfg2)

			log.V(7).Info(
				"translator-config",
				"translator", translator.Name,
				"appproject", appProject.Name,
				"config", translatorCfg.ProjectSpec)

			// Use mergo to merge non-empty fields from translatorCfg.ProjectSpec into appProject.Spec
			if err := reflection.Merge(translatedSpec, &translatorCfg.ProjectSpec); err != nil {
				return fmt.Errorf("failed to merge translator spec: %w", err)
			}

			// Use Metadata
			for key, value := range translatorCfg.ProjectMeta.Labels {
				appProject.Labels[key] = value
			}

			for key, value := range translatorCfg.ProjectMeta.Annotations {
				appProject.Annotations[key] = value
			}

			// Handle Finalizers
			//nolint:gocritic
			finalizers := append(translatorCfg.ProjectMeta.Finalizers, meta.TranslatorFinalizer(translator.Name))
			for _, finalizer := range finalizers {
				if !controllerutil.ContainsFinalizer(appProject, finalizer) {
					controllerutil.AddFinalizer(appProject, finalizer)
				}
			}

			appliedTranslatorsSet[translator.Name] = struct{}{}

			log.V(7).Info("reconciled", "translator", translator.Name, "appproject", appProject.Name)
		}

		// Remove unmatched Translators based on finalizers
		allTranslators := meta.GetTranslatingFinalizers(appProject)
		for _, translatorName := range allTranslators {
			if _, exists := appliedTranslatorsSet[translatorName]; !exists {
				if translator, found := unmatchedTranslators[translatorName]; found {
					log.V(7).Info("removing translator config", "appproject", appProject.Name, "translator", translatorName)

					// Call RemoveTranslatorForTenant with the actual translator object
					err := translatorctl.RemoveTranslatorForTenant(log, translator, tenant, appProject, i.Settings)
					if err != nil {
						log.Error(err, "failed to remove translator", "translator", translatorName)

						return err
					}
				}

				log.V(7).Info(
					"translator not present",
					"appproject", appProject.Name,
					"translator", translatorName)
			}
		}

		log.V(7).Info("combined translators config", "appproject", appProject.Name, "config", translatedSpec)

		//// Merge the translatedSpec into the appProject.Spec
		if i.Settings.Get().ReadOnlyTenant(tenant) {
			log.V(5).Info("overwriting spec", "appproject", appProject.Name)
			// Overwrite translatedSpec into the appProject.Spec
			appProject.Spec = *translatedSpec
		} else {
			log.V(5).Info("merging spec")
			// Merge with current Spec
			err = reflection.Merge(&appProject.Spec, translatedSpec)
			if err != nil {
				return fmt.Errorf("failed to merge project spec: %w", err)
			}
		}

		// Process ServiceAccount (Impersonation)
		impersonation := argocdv1alpha1.ApplicationDestinationServiceAccount{
			Server:                destination,
			Namespace:             "*",
			DefaultServiceAccount: i.Settings.Get().DestinationServiceAccount(tenant),
		}

		switch {
		// Add the proxy destination when the proxy is enabled and there are translators
		case i.Settings.Get().Argo.DestinationServiceAccounts && len(translators) > 0:
			if !argo.ProjectHasServiceAccount(appProject, impersonation) {
				log.V(5).Info("adding serviceaccount", "appproject", appProject.Name, "account", impersonation)
				appProject.Spec.DestinationServiceAccounts = append(appProject.Spec.DestinationServiceAccounts, impersonation)
			}
		// Remove the proxy destination
		default:
			log.V(5).Info("removing serviceaccount", "appproject", appProject.Name, "account", impersonation)
			argo.RemoveProjectServiceaccount(appProject, impersonation)
		}

		// Check if tenant is being deleted (Remove owner reference)
		log.V(5).Info("ensuring ownerreference", "appproject", appProject.Name)

		return meta.AddDynamicTenantOwnerReference(i.Client.Scheme(), appProject, tenant, i.Settings.Get().DecoupleTenant(tenant))
	})
	if err != nil {
		return err
	}

	// Reflect Argo RBAC
	err = i.reflectArgoRBAC(ctx, log, tenant, translators)
	if err != nil {
		return err
	}

	log.V(5).Info(
		"reflected argo permissions",
		"appproject", appProject.Name,
		"configmap", i.Settings.Get().Argo.RBACConfigMap,
		"namespace", i.Settings.Get().Argo.Namespace,
		"key", argo.ArgoPolicyName(tenant),
	)

	return nil
}

// Applies RBAC to the ArgoCD RBAC configmap in.
func (i *TenancyController) reflectArgoRBAC(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
) (err error) {
	// Initialize target configmap
	configmap := &corev1.ConfigMap{}
	if err := i.Client.Get(ctx, client.ObjectKey{
		Name:      i.Settings.Get().Argo.RBACConfigMap,
		Namespace: i.Settings.Get().Argo.Namespace}, configmap); err != nil {
		return err
	}

	// Empty Translators, attempt to remove the tenant from the configmap
	if len(translators) == 0 {
		log.V(7).Info("removing argo rbac", "tenant", tenant.Name)

		if _, ok := configmap.Data[argo.ArgoPolicyName(tenant)]; ok {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
				delete(configmap.Data, argo.ArgoPolicyName(tenant))

				return nil
			})

			return err
		}
	}

	// Generate Argo RBAC permissions
	rbacCSV, err := i.reflectArgoCSV(log, tenant, translators)
	if err != nil {
		return err
	}

	log.V(7).Info("resulting argo CSV", "tenant", tenant.Name, "csv", rbacCSV)

	// Apply the CSV to the configmap
	if !reflect.DeepEqual(configmap.Data[argo.ArgoPolicyName(tenant)], rbacCSV) {
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (conflictErr error) {
			_, conflictErr = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
				configmap.Data[argo.ArgoPolicyName(tenant)] = rbacCSV

				return nil
			})

			return
		})
		if err != nil {
			return err
		}
	} else {
		log.V(7).Info("csv already updated", "tenant", tenant.Name)
	}

	return nil
}

// Creates CSV file to be applied to the argo configmap.
func (i *TenancyController) reflectArgoCSV(
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
) (rbac string, err error) {
	var sb strings.Builder

	// Get Permissions for Tenant
	roles := utils.GetClusterRolePermissions(tenant)
	log.V(10).Info("extracted roles for tenant", "tenant", tenant.Name, "roles", roles)

	// Add Default Policies for App-Project
	for _, dlts := range argo.DefaultPolicies(tenant, i.Settings.Get().GetClusterDestination(tenant)) {
		sb.WriteString(dlts)
	}

	// Iterate over the translators custom CSV and append them
	for _, translator := range translators {
		// Default Policies
		log.V(7).Info("generating policies from translator", "translator", translator.Name)

		// Translate Policies
		for _, argopolicy := range translator.Spec.ProjectRoles {
			// Role-Name
			roleName := argo.TenantPolicy(tenant, argopolicy.Name)

			// Create Argo Policy
			for _, pol := range argopolicy.Policies {
				policy := argo.PolicyString(roleName, tenant.Name, pol)
				sb.WriteString(policy)
				log.V(10).Info("generated policy", "translator", translator.Name, "policy", policy)
			}

			log.V(7).Info("generating bindings")

			// Assign Users/Groups
			sb.WriteString("\n")

			for _, clusterRole := range argopolicy.ClusterRoles {
				log.V(7).Info("generating for subjects matching clusterrrole", "translator", translator.Name, "clusterrole", clusterRole)

				if val, ok := roles[clusterRole]; ok {
					log.V(10).Info("found subjects for clusterRole", "translator", translator.Name, "clusterrole", clusterRole, "subjects", val)

					for _, subject := range val {
						sb.WriteString(argo.BindingString(subject, roleName))

						// Assign Access to the tenant
						sb.WriteString(argo.BindingString(subject, argo.DefaultPolicyReadOnly(tenant)))

						if argopolicy.Owner {
							sb.WriteString(argo.BindingString(subject, argo.DefaultPolicyOwner(tenant)))
						}
					}
				} else {
					log.V(7).Info("no subjects found for clusterRole", "clusterrole", clusterRole)
				}
			}
		}

		// Update Custom-Policies
		if translator.Spec.CustomPolicy != "" {
			log.V(7).Info("appending custom policy from translator", "translator", translator.Name)

			sb.WriteString("\n")
			sb.WriteString(translator.Spec.CustomPolicy)
			sb.WriteString("\n")
		}
	}

	// Template CSV
	log.V(7).Info("templating argo csv")

	ArgoCSVTemplate := sb.String()

	tmpl, err := template.New("rbac").Funcs(tpl.ExtraFuncMap()).Parse(ArgoCSVTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", err
	}

	finalCSV := buf.String()

	if err := argo.ValidateCSV(finalCSV); err != nil {
		return "", errors.New("invalid argo csv: " + err.Error())
	}

	return finalCSV, nil
}
