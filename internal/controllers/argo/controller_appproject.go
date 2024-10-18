package argo

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
	"github.com/argoproj/argo-cd/v2/util/rbac"
	"github.com/go-logr/logr"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	tpl "github.com/peak-scale/capsule-argo-addon/internal/template"
	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Creates or updates the ArgoCD Application Project for the tenant
func (i *TenancyController) reconcileArgoProject(ctx context.Context, log logr.Logger, tenant *capsulev1beta2.Tenant, translators []*v1alpha1.ArgoTranslator) (err error) {
	// Collect Service-Account
	token, err := i.reconcileArgoServiceAccount(ctx, tenant)
	if err != nil {
		return err
	}

	// Reconcile Argo Cluster
	cluster, err := i.reconcileArgoCluster(ctx, log, tenant, token)
	if err != nil {
		return err
	}

	// Collect Service-Account

	appProject := &argocdv1alpha1.AppProject{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.TenantProjectName(tenant),
			Namespace: i.Settings.Get().ArgoCD.Namespace,
		},
	}

	// Fetch the current state of the AppProject
	err = i.Client.Get(ctx, client.ObjectKey{Name: tenant.Name, Namespace: i.Settings.Get().ArgoCD.Namespace}, appProject)
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
			finalizers := append(translatorCfg.ProjectMeta.Finalizers, TranslatorFinalizer(translator))
			for _, finalizer := range finalizers {
				if !controllerutil.ContainsFinalizer(appProject, finalizer) {
					controllerutil.AddFinalizer(appProject, finalizer)
				}
			}
		}

		log.V(6).Info("combined translators config", "config", translatedSpec)

		//// Merge the translatedSpec into the appProject.Spec
		if utils.TenantReadOnly(tenant) {
			log.V(6).Info("overwriting appproject")
			// Overwrite translatedSpec into the appProject.Spec
			appProject.Spec = *translatedSpec
		} else {
			// Merge with current Spec
			err = mergo.Merge(&appProject.Spec, translatedSpec, mergo.WithOverride)
			if err != nil {
				return fmt.Errorf("failed to merge project spec: %w", err)
			}
		}

		// Register the Tenant as a Destination
		if i.Settings.Get().Proxy.Enabled {
			proxyDestination := argocdv1alpha1.ApplicationDestination{
				Name:      tenant.Name,
				Server:    cluster,
				Namespace: "*",
			}

			if !projectHasDestination(appProject, proxyDestination) {
				appProject.Spec.Destinations = append(appProject.Spec.Destinations, proxyDestination)
			}
		}

		// Couple oder Decouple the AppProject
		if utils.TenantDecoupleProject(tenant) {
			// Unset blockerOwnerDeletion and controller
			if controllerutil.HasControllerReference(appProject) {
				log.V(5).Info("removing controller reference", "project", appProject)
				return controllerutil.RemoveControllerReference(tenant, appProject, i.Client.Scheme())
			}
		} else {
			log.V(5).Info("adding controller reference", "project", appProject)
			return controllerutil.SetControllerReference(tenant, appProject, i.Client.Scheme())
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

	log.V(5).Info("reflected argo permissions", "configmap", i.Settings.Get().ArgoCD.RBACConfigMap, "namespace", i.Settings.Get().ArgoCD.Namespace, "key", utils.ArgoPolicyName(tenant))
	return nil
}

// Applies RBAC to the ArgoCD RBAC configmap in
func (i *TenancyController) reflectArgoRBAC(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
	translators []*v1alpha1.ArgoTranslator,
) (err error) {

	// Generate RBAC CSV
	rbacCSV, err := i.reflectArgoCSV(ctx, tenant, translators)
	if err != nil {
		return err
	}

	// Update existing configmap with new csv
	configmap := &corev1.ConfigMap{}
	err = i.Client.Get(ctx, client.ObjectKey{
		Name:      i.Settings.Get().ArgoCD.RBACConfigMap,
		Namespace: i.Settings.Get().ArgoCD.Namespace}, configmap)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(configmap.Data[utils.ArgoPolicyName(tenant)], rbacCSV) {
		err = retry.RetryOnConflict(retry.DefaultBackoff, func() (conflictErr error) {
			_, conflictErr = controllerutil.CreateOrUpdate(ctx, i.Client, configmap, func() error {
				configmap.Data[utils.ArgoPolicyName(tenant)] = rbacCSV

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
	sb.WriteString(argoDefaultPolicies(tenant))

	// Iterate over the translators custom CSV and append them
	for _, translator := range translators {
		// Default Policies

		// Translate Policies
		for _, argopolicy := range translator.Spec.ProjectRoles {
			// Role-Name
			roleName := fmt.Sprintf("role:%s:%s", tenant.Name, argopolicy.Name)

			// Create Argo Policy
			for _, pol := range argopolicy.Policies {
				sb.WriteString(argoPolicyString(roleName, tenant, pol))
			}

			// Assign Users/Groups
			sb.WriteString("\n")
			for _, clusterRole := range argopolicy.ClusterRoles {
				if val, ok := roles[clusterRole]; ok {
					for _, subject := range val {
						sb.WriteString(argoAssignString(subject, roleName))

						// Assign Access to the tenant
						if argopolicy.Owner {
							sb.WriteString(argoAssignString(subject, argoDefaultPolicyOwner(tenant)))
						} else {
							sb.WriteString(argoAssignString(subject, argoDefaultPolicyReadOnly(tenant)))
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

	if err := i.isValidArgoCSV(finalCSV); err != nil {
		return "", errors.New("invalid argo csv: " + err.Error())
	}

	return finalCSV, nil
}

// Converts the ArgoCD Project Policy Definition to a string (common argo)
func argoPolicyString(role string, tenant *capsulev1beta2.Tenant, argopolicy addonsv1alpha1.ArgocdPolicyDefinition) string {
	var result string

	for _, action := range argopolicy.Action {
		// Accumulate each formatted string into the result
		result += fmt.Sprintf(
			"p, %s,%s,%s,%s/%s,%s\n",
			role,                // Project name
			argopolicy.Resource, // Resource (enum)
			action,              // Action (enum)
			tenant.Name,         // Tenant name
			argopolicy.Path,     // Path (enum)
			argopolicy.Verb,     // Verb (enum)
		)
	}

	return result
}

// Adds Default Policies (So Users can have basic interractions with the project)
func argoDefaultPolicies(tenant *capsulev1beta2.Tenant) string {
	var result string

	// Read-Only Policy
	result += fmt.Sprintf(
		"p, %s,projects,get,%s,allow\np, %s,clusters,get,%s/*,allow\n",
		argoDefaultPolicyReadOnly(tenant), // Project name
		tenant.Name,                       // Project name
		argoDefaultPolicyReadOnly(tenant), // Project name
		tenant.Name,                       // Project name
	)
	// Owner Policy
	result += fmt.Sprintf(
		"p, %s,projects,get,%s,allow\n",
		argoDefaultPolicyOwner(tenant), // Project name
		tenant.Name,                    // Project name
	)
	result += fmt.Sprintf(
		"p, %s,projects,update,%s,allow\n",
		argoDefaultPolicyOwner(tenant), // Project name
		tenant.Name,                    // Project name
	)

	return result
}

func argoDefaultPolicyAny(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("role:%s:owner", tenant.Name)
}

func argoDefaultPolicyOwner(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("role:%s:owner", tenant.Name)
}

func argoDefaultPolicyReadOnly(tenant *capsulev1beta2.Tenant) string {
	return fmt.Sprintf("role:%s:read-only", tenant.Name)
}

func argoAssignString(subject v1.Subject, role string) string {
	return fmt.Sprintf(
		"g, %s, %s\n",
		subject.Name,
		role,
	)
}

// Validates the ArgoCD RBAC CSV
func (i *TenancyController) isValidArgoCSV(csv string) error {
	return rbac.ValidatePolicy(csv)
}

func projectHasDestination(appProject *argocdv1alpha1.AppProject, newDestination argocdv1alpha1.ApplicationDestination) bool {
	// Check if the destination already exists
	exists := false
	for _, dest := range appProject.Spec.Destinations {
		if dest.Name == newDestination.Name && dest.Namespace == newDestination.Namespace && dest.Server == newDestination.Server {
			exists = true
			break
		}
	}

	return exists
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
