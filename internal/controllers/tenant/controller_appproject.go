package tenant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"dario.cat/mergo"
	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	translatorctl "github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
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
func (i *TenancyController) reconcileArgoProject(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant, translators []*v1alpha1.ArgoTranslator) (err error) {
	// Collect Service-Account
	token, err := i.reconcileArgoServiceAccount(ctx, log, tenant)
	if err != nil {
		return err
	}

	// Reconcile Argo Cluster
	cluster, err := i.reconcileArgoCluster(ctx, log, tenant, token)
	if err != nil {
		return err
	}

	// Initialize AppProject
	appProject := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      meta.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		},
	}

	// Fetch the current state of the AppProject
	gerr := i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().Argo.Namespace}, appProject)
	if gerr != nil && !k8serrors.IsNotFound(gerr) {
		return gerr
	}

	// Don't Force, When project already exists
	if !translator.ContainsTranslatorFinalizer(appProject) {
		if !i.Settings.Get().Force && !k8serrors.IsNotFound(gerr) {
			log.V(1).Info("appproject already present, not overriding", "appproject", appProject.Name)

			return nil
		}

	}

	// Lifecycle Approject (If marked for deletion remove finalizers)
	if !appProject.ObjectMeta.DeletionTimestamp.IsZero() {
		log.V(5).Info("removing finalizers", "appproject", appProject.Name)
		_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
			for _, translator := range translators {
				if controllerutil.ContainsFinalizer(appProject, translatorctl.TranslatorFinalizer(translator)) {
					controllerutil.RemoveFinalizer(appProject, translatorctl.TranslatorFinalizer(translator))
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		return nil
	}

	// Handle when the tenant is being deleted but the AppProject is decoupled
	// In this case we remove the owner reference and the tenant tracking label so the Appproject can still exist
	if tenant.ObjectMeta.DeletionTimestamp != nil && meta.TenantDecoupleProject(tenant) {
		log.V(5).Info("decoupling appproject", "appproject", appProject.Name)
		_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
			// Remove any Translator References
			for _, translator := range translators {
				if controllerutil.ContainsFinalizer(appProject, translatorctl.TranslatorFinalizer(translator)) {
					controllerutil.RemoveFinalizer(appProject, translatorctl.TranslatorFinalizer(translator))
				}
			}

			// Remove References to origin Tenant
			if err := i.DynamicRemoveOwnerReference(ctx, appProject, tenant); err != nil {
				return err
			}

			// Remove tenant tracking label
			appProject.Labels = meta.TranslatorRemoveTenantLabels(appProject.GetLabels())

			return nil
		})

		return nil
	}

	// Lifecycle Approject (If no translators are present, remove the Approject)
	if len(translators) == 0 {
		// Approject is already absent
		if k8serrors.IsNotFound(gerr) {
			return nil
		}

		// Delete the AppProject when it's not decoupled
		if !meta.TenantDecoupleProject(tenant) {
			return i.Client.Delete(ctx, appProject)
		} else {
			// Remove References to origin Tenant
			if err := i.DynamicRemoveOwnerReference(ctx, appProject, tenant); err != nil {
				return err
			}
		}
	}

	log.Info("reconcile appproject", "appproject", appProject.Name)

	_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
		// Prepare metadata
		appProject.ObjectMeta.Labels = meta.TranslatorTrackingLabels(tenant)
		if appProject.ObjectMeta.Annotations == nil {
			appProject.ObjectMeta.Annotations = make(map[string]string)
		}

		translatedSpec := &argocdv1alpha1.AppProjectSpec{}
		for _, translator := range translators {
			// Get Approject Config with templating
			translatorCfg, err := translator.Spec.ProjectSettings.GetConfig(
				tpl.ConfigContext(cluster, translator, i.Settings.Get(), tenant), tpl.ExtraFuncMap())
			if err != nil {
				return err
			}

			cfg1, cfg2, err := translator.Spec.ProjectSettings.GetConfigs(
				tpl.ConfigContext(cluster, translator, i.Settings.Get(), tenant), tpl.ExtraFuncMap())
			if err != nil {
				return err
			}
			log.V(7).Info("translator-config", "appproject", appProject.Name, "structured", cfg1, "templated", cfg2)

			log.V(7).Info("translator-config", "appproject", appProject.Name, "config", translatorCfg.ProjectSpec)

			// Use mergo to merge non-empty fields from translatorCfg.ProjectSpec into appProject.Spec
			err = mergo.Merge(translatedSpec, translatorCfg.ProjectSpec)
			if err != nil {
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
			finalizers := append(translatorCfg.ProjectMeta.Finalizers, translatorctl.TranslatorFinalizer(translator))
			for _, finalizer := range finalizers {
				if !controllerutil.ContainsFinalizer(appProject, finalizer) {
					controllerutil.AddFinalizer(appProject, finalizer)
				}
			}
		}

		log.V(7).Info("combined translators config", "appproject", appProject.Name, "config", translatedSpec)

		//// Merge the translatedSpec into the appProject.Spec
		if meta.TenantReadOnly(tenant) {
			log.V(5).Info("overwriting spec", "appproject", appProject.Name)
			// Overwrite translatedSpec into the appProject.Spec
			appProject.Spec = *translatedSpec
		} else {
			log.V(5).Info("merging spec")
			// Merge with current Spec
			err = mergo.Merge(&appProject.Spec, translatedSpec, mergo.WithOverride)
			if err != nil {
				return fmt.Errorf("failed to merge project spec: %w", err)
			}
		}

		// Register the Tenant as a Destination
		proxyDestination := argocdv1alpha1.ApplicationDestination{
			Name:      tenant.Name,
			Server:    cluster,
			Namespace: "*",
		}

		switch {
		// Add the proxy destination when the proxy is enabled and there are translators
		case i.Settings.Get().Proxy.Enabled && len(translators) > 0:
			if !argo.ProjectHasDestination(appProject, proxyDestination) {
				log.V(5).Info("adding proxy destination", "appproject", appProject.Name)
				appProject.Spec.Destinations = append(appProject.Spec.Destinations, proxyDestination)
			}
		// Remove the proxy destination
		default:
			if argo.ProjectHasDestination(appProject, proxyDestination) {
				log.V(5).Info("removing proxy destination", "appproject", appProject.Name)
				argo.RemoveProjectDestination(appProject, proxyDestination)
			}
		}

		// Couple oder Decouple the AppProject

		// Check if tenant is being deleted (Remove owner reference)
		log.V(5).Info("ensuring ownerreference", "appproject", appProject.Name)
		if err := i.DynamicOwnerReference(ctx, appProject, tenant); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Reflect Argo RBAC
	err = i.reflectArgoRBAC(ctx, tenant, translators)
	if err != nil {
		return err
	}

	log.V(5).Info("reflected argo permissions", "appproject", appProject.Name, "configmap", i.Settings.Get().Argo.RBACConfigMap, "namespace", i.Settings.Get().Argo.Namespace, "key", argo.ArgoPolicyName(tenant))
	return nil
}

// Applies RBAC to the ArgoCD RBAC configmap in
func (i *TenancyController) reflectArgoRBAC(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
) (err error) {
	// Initialize target configmap
	configmap := &corev1.ConfigMap{}
	err = i.Client.Get(ctx, client.ObjectKey{
		Name:      i.Settings.Get().Argo.RBACConfigMap,
		Namespace: i.Settings.Get().Argo.Namespace}, configmap)
	if err != nil {
		return err
	}

	// Empty Translators, attempt to remove the tenant from the configmap
	if len(translators) == 0 {
		if _, ok := configmap.Data[argo.ArgoPolicyName(tenant)]; ok {
			_, err = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
				delete(configmap.Data, argo.ArgoPolicyName(tenant))

				return nil
			})
			return err
		}
	}

	// Generate Argo RBAC permissions
	rbacCSV, err := i.reflectArgoCSV(ctx, tenant, translators)
	if err != nil {
		return err
	}

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
	}

	return nil
}

// Creates CSV file to be applied to the argo configmap
func (i *TenancyController) reflectArgoCSV(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
) (rbac string, err error) {
	var sb strings.Builder

	// Get Permissions for Tenant
	roles := utils.GetClusterRolePermissions(tenant)

	// Add Default Policies for App-Project
	for _, dlts := range argo.DefaultPolicies(tenant, i.provisionProxyService(tenant)) {
		sb.WriteString(dlts)
	}

	// Iterate over the translators custom CSV and append them
	for _, translator := range translators {
		// Default Policies

		// Translate Policies
		for _, argopolicy := range translator.Spec.ProjectRoles {
			// Role-Name
			roleName := fmt.Sprintf("role:%s:%s", tenant.Name, argopolicy.Name)

			// Create Argo Policy
			for _, pol := range argopolicy.Policies {
				sb.WriteString(argo.PolicyString(roleName, tenant.Name, pol))
			}

			// Assign Users/Groups
			sb.WriteString("\n")
			for _, clusterRole := range argopolicy.ClusterRoles {
				if val, ok := roles[clusterRole]; ok {
					for _, subject := range val {
						sb.WriteString(argo.BindingString(subject, roleName))

						// Assign Access to the tenant
						sb.WriteString(argo.BindingString(subject, argo.DefaultPolicyReadOnly(tenant)))
						if argopolicy.Owner {
							sb.WriteString(argo.BindingString(subject, argo.DefaultPolicyOwner(tenant)))
						}
					}
				}
			}
		}

		// Update Custom-Policies
		if translator.Spec.CustomPolicy != "" {
			sb.WriteString("\n")
			sb.WriteString(translator.Spec.CustomPolicy)
			sb.WriteString("\n")
		}
	}

	// Template CSV
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
