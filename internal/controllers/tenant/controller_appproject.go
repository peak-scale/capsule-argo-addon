// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package tenant

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"text/template"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	ccaerrrors "github.com/peak-scale/capsule-argo-addon/internal/errors"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/peak-scale/capsule-argo-addon/internal/reflection"
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

// Reconciler.
func (i *Reconciler) reconcileProject(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*configv1alpha1.ArgoTranslator,
) (finalize bool, err error) {
	defer func() {
		var errs error

		for _, translator := range translators {
			if err != nil {
				var condition metav1.Condition

				// Check the type of error with a type switch
				eo := &ccaerrrors.ObjectAlreadyExistsError{}
				if errors.As(err, &eo) {
					condition = meta.NewAlreadyExistsCondition(tenant, err.Error())
				} else {
					// Default NotReady condition for other errors
					condition = meta.NewNotReadyCondition(tenant, err.Error())
				}

				translator.UpdateTenantCondition(configv1alpha1.TenantStatus{
					Name:      tenant.Name,
					UID:       tenant.UID,
					Condition: condition,
					Serving:   translator.Spec.ProjectSettings,
				})
			}

			// Update Translator
			errs = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
				return i.Client.Status().Update(ctx, translator)
			})

			log.V(7).Info("updated", "translation", translator.Name, "err", err)
		}

		err = errs
	}()

	finalize = false

	// Initialize AppProject
	origin := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		},
	}

	log.V(7).Info("reconciling appproject", "appproject", origin.Name)

	// Fetch the current state of the AppProject
	gerr := i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().Argo.Namespace}, origin)
	if gerr != nil && !k8serrors.IsNotFound(gerr) {
		return finalize, gerr
	}

	appProject := origin.DeepCopy()

	if gerr != nil && k8serrors.IsNotFound(gerr) {
		appProject.ResourceVersion = ""
	}

	// Currently Applied Translators
	presentTranslators := meta.GetTranslatingFinalizers(appProject)

	// Track state of matching translators
	appliedTranslators := []*configv1alpha1.ArgoTranslator{}

	//// Merge the translatedSpec into the appProject.Spec
	if i.Settings.Get().ReadOnlyTenant(tenant) {
		// Overwrite relevant Meta, we don't want to overwrite ObjectMeta
		// since this would loose the resourceVersion and UID
		appProject.Labels = make(map[string]string)
		appProject.Annotations = make(map[string]string)
		appProject.Finalizers = []string{}
		appProject.Spec = argocdv1alpha1.AppProjectSpec{}
	}

	// Execute Translators
	for _, translator := range translators {
		tlog := log.WithValues("translator", translator.Name)

		// Always remove Finalizer
		controllerutil.RemoveFinalizer(appProject, meta.TranslatorFinalizer(translator.Name))

		// Remove Finalizer when it's being deleted
		if !appProject.ObjectMeta.DeletionTimestamp.IsZero() {
			continue
		}

		// Reconcile Translator
		applied, err := i.reconcileTranslator(
			tlog,
			tenant,
			appProject,
			translator,
			presentTranslators,
		)
		// An error is considered not applied
		if err != nil {
			continue
		}

		// Update Status for Translator
		if applied {
			finalize = true

			controllerutil.AddFinalizer(appProject, meta.TranslatorFinalizer(translator.Name))
			appliedTranslators = append(appliedTranslators, translator)
		}
	}

	// Lifecycle Approject (If no translators are present, remove the Appproject)
	if len(appliedTranslators) == 0 {
		return finalize, nil
	}

	// Provision Other resources
	token, err := i.reconcileArgoServiceAccount(ctx, log, tenant)
	if err != nil {
		return finalize, err
	}

	err = i.reconcileArgoCluster(ctx, log, tenant, token)
	if err != nil {
		return finalize, err
	}

	if len(appliedTranslators) != 0 {
		if !meta.HasTenantOwnerReference(appProject, tenant) || len(meta.GetTranslatingFinalizers(appProject)) == 0 {
			if !i.Settings.Get().ForceTenant(tenant) && !k8serrors.IsNotFound(gerr) {
				log.V(1).Info("appproject already present, not overriding", "appproject", appProject.Name)

				return finalize, ccaerrrors.NewObjectAlreadyExistsError(appProject)
			}
		}
	}

	// Update Project
	_, err = controllerutil.CreateOrPatch(ctx, i.Client, origin, func() error {
		appProject.Labels = meta.WithTranslatorTrackingLabels(appProject, tenant)

		// Redirect Specification
		origin.ObjectMeta = appProject.ObjectMeta
		origin.Spec = appProject.Spec

		// Further Project Properties
		log.V(7).Info("combined translators config", "appproject", appProject.Name, "config", appProject.Spec)

		// Process ServiceAccount (Impersonation)
		impersonation := argocdv1alpha1.ApplicationDestinationServiceAccount{
			Server:                i.Settings.Get().GetClusterDestination(tenant),
			Namespace:             "*",
			DefaultServiceAccount: i.Settings.Get().DestinationServiceAccount(tenant),
		}

		switch {
		// Add the proxy destination when the proxy is enabled and there are translators
		case i.Settings.Get().Argo.DestinationServiceAccounts && len(translators) > 0:
			if !argo.ProjectHasServiceAccount(origin, impersonation) {
				log.V(5).Info("adding serviceaccount", "appproject", origin.Name, "account", impersonation)
				origin.Spec.DestinationServiceAccounts = append(origin.Spec.DestinationServiceAccounts, impersonation)
			}
		// Remove the proxy destination
		default:
			log.V(5).Info("removing serviceaccount", "appproject", origin.Name, "account", impersonation)
			argo.RemoveProjectServiceaccount(origin, impersonation)
		}

		// Check if tenant is being deleted (Remove owner reference)
		log.V(5).Info("ensuring ownerreference", "appproject", origin.Name)

		return meta.AddDynamicTenantOwnerReference(i.Client.Scheme(), origin, tenant, i.Settings.Get().DecoupleTenant(tenant))
	})
	if err != nil {
		return finalize, err
	}

	// Reflect Argo RBAC
	err = i.reflectArgoRBAC(ctx, log, tenant, translators)
	if err != nil {
		return finalize, err
	}

	log.V(5).Info(
		"reflected argo permissions",
		"appproject", appProject.Name,
		"configmap", i.Settings.Get().Argo.RBACConfigMap,
		"namespace", i.Settings.Get().Argo.Namespace,
		"key", argo.ArgoPolicyName(tenant),
	)

	return finalize, nil
}

func (i *Reconciler) reconcileTranslator(
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	appProject *argocdv1alpha1.AppProject,
	translator *configv1alpha1.ArgoTranslator,
	appliedTranslators []string,
) (match bool, err error) {
	// Always Handle the Error for the Translator
	defer func() {
		log.V(7).Info("status", "match", match, "error", err)

		switch {
		// Add the proxy destination when the proxy is enabled and there are translators
		case !match && err == nil:
			translator.RemoveTenantCondition(tenant.Name)

			log.V(5).Info("removed translation")
		case err != nil:
			translator.UpdateTenantCondition(configv1alpha1.TenantStatus{
				Name:      tenant.Name,
				UID:       tenant.UID,
				Condition: meta.NewNotReadyCondition(tenant, err.Error()),
				Serving:   translator.Spec.ProjectSettings,
			})

			log.V(5).Info("executed translation", "status", translator.GetTenantCondition(tenant))
		default:
			translator.UpdateTenantCondition(configv1alpha1.TenantStatus{
				Name:      tenant.Name,
				UID:       tenant.UID,
				Condition: meta.NewReadyCondition(tenant),
				Serving:   translator.Spec.ProjectSettings,
			})

			log.V(5).Info("executed translation", "status", translator.GetTenantCondition(tenant))
		}
	}()

	// Evaluate Matching
	match = translator.MatchesObject(tenant)

	log.V(5).Info("matches", "state", match)

	// When a tenant is deleted it's considered not a match
	if !tenant.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(5).Info("tenant is being deleted", "state", match)

		return false, nil
	}

	// When a tenant is deleted it's considered not a match
	if !translator.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(5).Info("translator is being deleted", "state", match)

		return false, nil
	}

	// Lifecycle Translator (If not matching)
	if !match {
		for _, appliedTranslator := range appliedTranslators {
			if translator.Name == appliedTranslator {
				// Call RemoveTranslatorForTenant with the actual translator object
				if err = RemoveTranslatorForTenant(translator, tenant, appProject, i.Settings); err != nil {
					return match, err
				}

				break
			}
		}

		return match, err
	}

	log.V(7).Info("lifecycling previous specification")

	// We might need to Lifecycle Old Translation with this
	if err = SubstractTranslatorSpec(
		translator,
		tenant,
		appProject,
		i.Settings,
	); err != nil {
		return match, err
	}

	log.V(7).Info("merging configuration")

	translatorCfg, err := GetMergedConfig(
		tenant,
		translator.Spec.ProjectSettings,
		i.Settings,
	)
	if err != nil {
		return match, err
	}

	// We can skip when the config is empty
	if translatorCfg == nil {
		return match, err
	}

	log.V(7).Info("adding finalizers")

	// Add Translator Finalizer.
	// We can now assume there are going to be changes from this translator.
	finalizer := meta.TranslatorFinalizer(translator.Name)
	if !controllerutil.ContainsFinalizer(appProject, finalizer) {
		controllerutil.AddFinalizer(appProject, finalizer)
	}

	log.V(7).Info(
		"translator-config",
		"appproject", appProject.Name,
		"config", translatorCfg.ProjectSpec)

	// Use mergo to merge non-empty fields from translatorCfg.ProjectSpec into appProject.Spec
	if err = reflection.Merge(&appProject.Spec, &translatorCfg.ProjectSpec); err != nil {
		return match, err
	}

	if translatorCfg.ProjectMeta != nil {
		if appProject.ObjectMeta.Labels == nil {
			appProject.ObjectMeta.Labels = make(map[string]string)
		}

		// Use Metadata
		for key, value := range translatorCfg.ProjectMeta.Labels {
			appProject.Labels[key] = value
		}

		if appProject.ObjectMeta.Annotations == nil {
			appProject.ObjectMeta.Annotations = make(map[string]string)
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
	}

	log.V(7).Info("reconciled", "translator", translator.Name, "appproject", appProject.Name)

	return match, err
}

// Applies RBAC to the ArgoCD RBAC configmap in.
func (i *Reconciler) reflectArgoRBAC(
	ctx context.Context,
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*configv1alpha1.ArgoTranslator,
) (err error) {
	// Initialize target configmap
	configmap := &corev1.ConfigMap{}
	if err := i.Client.Get(ctx, client.ObjectKey{
		Name:      i.Settings.Get().Argo.RBACConfigMap,
		Namespace: i.Settings.Get().Argo.Namespace,
	}, configmap); err != nil {
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
func (i *Reconciler) reflectArgoCSV(
	log logr.Logger,
	tenant *capsulev1beta2.Tenant,
	translators []*configv1alpha1.ArgoTranslator,
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

func (i *Reconciler) lifecycleArgoProject(ctx context.Context, tenant *capsulev1beta2.Tenant) (err error) {
	// Remove the approject from the tenant
	appProject := &argocdv1alpha1.AppProject{}

	err = i.Client.Get(ctx, client.ObjectKey{
		Name:      meta.TenantProjectName(tenant),
		Namespace: i.Settings.Get().Argo.Namespace,
	}, appProject)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return
	}

	if !meta.HasTenantOwnerReference(appProject, tenant) {
		return nil
	}

	_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
		if len(meta.GetTranslatingFinalizers(appProject)) != 0 {
			for _, finalizer := range meta.GetTranslatingFinalizers(appProject) {
				controllerutil.RemoveFinalizer(appProject, meta.TranslatorFinalizer(finalizer))
			}
		}

		if !i.Settings.Get().DecoupleTenant(tenant) {
			return i.Client.Delete(ctx, appProject)
		}

		return i.DecoupleTenant(appProject, tenant)
	})

	return
}

func (i *Reconciler) lifecycleArgoRbac(ctx context.Context, tenant *capsulev1beta2.Tenant) (err error) {
	// Update existing configmap with new csv
	if !i.Settings.Get().DecoupleTenant(tenant) {
		configmap := &corev1.ConfigMap{}
		if err = i.Client.Get(ctx, client.ObjectKey{
			Name:      i.Settings.Get().Argo.RBACConfigMap,
			Namespace: i.Settings.Get().Argo.Namespace,
		},
			configmap,
		); err != nil {
			return
		}

		_, err = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
			delete(configmap.Data, argo.ArgoPolicyName(tenant))

			return nil
		})
		if err != nil {
			return err
		}
	}

	return
}
