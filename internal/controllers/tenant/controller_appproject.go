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
	translatorctl "github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
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
			Name:      utils.TenantProjectName(tenant),
			Namespace: i.Settings.Get().Argo.Namespace,
		},
	}

	// Fetch the current state of the AppProject
	gerr := i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().Argo.Namespace}, appProject)
	if gerr != nil && !k8serrors.IsNotFound(gerr) {
		return gerr
	}

	// Lifecycle Approject
	if len(translators) == 0 {
		// Approject is already absent
		if k8serrors.IsNotFound(gerr) {
			return nil
		}

		// Delete the AppProject when it's not decoupled
		if !utils.TenantDecoupleProject(tenant) {
			return i.Client.Delete(ctx, appProject)
		}
	}
	// Fetch the current state of the AppProject
	err = i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().Argo.Namespace}, appProject)
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}

	log.Info("reconciling argo project", "project", appProject)

	_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
		// Prepare metadata
		appProject.ObjectMeta.Labels = utils.TranslatorTrackingLabels(tenant)
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
			log.V(7).Info("translator-config", "structured", cfg1, "templated", cfg2)

			log.V(7).Info("translator-config", "config", translatorCfg.ProjectSpec)

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

		log.V(7).Info("combined translators config", "config", translatedSpec)

		//// Merge the translatedSpec into the appProject.Spec
		if utils.TenantReadOnly(tenant) {
			log.V(5).Info("overwriting appproject")
			// Overwrite translatedSpec into the appProject.Spec
			appProject.Spec = *translatedSpec
		} else {
			log.V(5).Info("combining appproject")
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
				log.V(5).Info("adding proxy destination")
				appProject.Spec.Destinations = append(appProject.Spec.Destinations, proxyDestination)
			}
		// Remove the proxy destination
		default:
			if argo.ProjectHasDestination(appProject, proxyDestination) {
				log.V(5).Info("removing proxy destination")
				argo.RemoveProjectDestination(appProject, proxyDestination)
			}
		}

		// Couple oder Decouple the AppProject

		// Check if tenant is being deleted (Remove owner reference)
		log.V(5).Info("ensuring ownerreference", appProject)
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

	log.V(5).Info("reflected argo permissions", "configmap", i.Settings.Get().Argo.RBACConfigMap, "namespace", i.Settings.Get().Argo.Namespace, "key", argo.ArgoPolicyName(tenant))
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
	sb.WriteString(argo.DefaultPolicies(tenant, i.provisionProxyService(ctx, tenant)))

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

//func (i *TenancyController) argoPolicyTranslator(ctx context.Context, tenant *capsulev1beta2.Tenant, translators []v1alpha1.TenantTranslator) (roles []argocdv1alpha1.ProjectRole, err error) {
//	// Iterate over the translators
//	for _, translator := range translators {
//		for i, translatorMap := range *translator.ProjectRoles {
//			role := &argocdv1alpha1.ProjectRole{
//				Name: tenant.Name + "-" + translatorMap.Name + "-" + strconv.Itoa(i),
//			}
//
//			// Translate the policies
//			var policies []string
//			for _, pol := range translatorMap.Policies {
//				policies = append(policies, argoPolicyString(tenant, translator, pol))
//			}
//			role.Policies = policies
//		}
//	}
//}
