package controller

import (
	"context"
	"encoding/json"
	"reflect"

	"git.bedag.cloud/gelan/gelan-infra/controllers/tenancy-controller/internal/roles"
	"git.bedag.cloud/gelan/gelan-infra/controllers/tenancy-controller/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (i *TenancyController) reconcileAddons(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	return i.tenantArgoProject(tenant, ctx)
}

// Creates Teanant Service Account with the given name and namespace
func (i *TenancyController) tenantServiceAccount(tenant *capsulev1beta2.Tenant, ctx context.Context) (token string, err error) {
	targetNamespace := i.Options.UserTenantNamespace
	if utils.IsSystemTenant(tenant) {
		targetNamespace = i.Options.SystemTenantNamespace
	}

	accountResource := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: targetNamespace,
			Labels:    utils.CommonLabels(),
			OwnerReferences: []metav1.OwnerReference{
				utils.GetOwnerReference(tenant),
			},
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, accountResource, func() (err error) {
		return controllerutil.SetControllerReference(tenant, accountResource, i.Client.Scheme())
	})

	tokenResource := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: targetNamespace,
			Labels:    utils.CommonLabels(),
			Annotations: map[string]string{
				"kubernetes.io/service-account.name": tenant.Name,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	// Create Account Token
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tokenResource, func() (err error) {
		return controllerutil.SetControllerReference(tenant, tokenResource, i.Client.Scheme())
	})

	var secret corev1.Secret
	if err = i.Client.Get(ctx, client.ObjectKey{
		Name:      tokenResource.Name,
		Namespace: targetNamespace,
	}, &secret); err != nil {
		return "", err
	}

	// Assuming the token is stored under a specific key, e.g., "token"
	t, exists := secret.Data["token"]
	if !exists {
		return "", err
	}

	token = string(t)

	err = i.addServiceAccountOwner(accountResource.Namespace, tenant.Name, tenant, ctx)
	if err != nil {
		return "", err
	}

	i.Log.V(5).Info("SeriviceAccount created", "name", tenant.Name)

	return
}

// Creates a new services for the tenant
func (i *TenancyController) tenantProxyService(name string, namespace string, tenant *capsulev1beta2.Tenant, ctx context.Context) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    utils.CommonLabels(),
			OwnerReferences: []metav1.OwnerReference{
				utils.GetOwnerReference(tenant),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					Port:       9001,
					TargetPort: intstr.FromInt32(i.Options.CapsuleProxyServicePort),
				},
			},
			Selector: map[string]string{
				"app.kubernetes.io/instance": "capsule-proxy",
				"app.kubernetes.io/name":     "capsule-proxy",
			},
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, i.Client, service, func() error {

		return controllerutil.SetControllerReference(tenant, service, i.Client.Scheme())
	})
	if err != nil {
		return err
	}

	i.Log.V(5).Info("Proxy Service created", "name", tenant.Name)

	return nil
}

func (i *TenancyController) tenantArgoServer(tenant *capsulev1beta2.Tenant, ctx context.Context) error {

	svc, url := i.getProxyServiceName(tenant)

	token, err := i.tenantServiceAccount(tenant, ctx)
	if err != nil {
		return err
	}

	err = i.tenantProxyService(svc, i.Options.CapsuleProxyServiceNamespace, tenant, ctx)
	if err != nil {
		return err
	}

	serverSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tenant.Name,
			Namespace: i.Options.ArgoCDNamespace,
			Labels: map[string]string{
				"argocd.argoproj.io/secret-type": "cluster",
			},
		},
		Type: corev1.SecretTypeOpaque,
	}

	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, serverSecret, func() error {

		extraData := map[string]interface{}{
			"bearerToken": token,
			"tlsClientConfig": map[string]interface{}{
				"insecure": true,
			},
		}

		jsonData, err := json.Marshal(extraData)

		serverSecret.StringData = map[string]string{
			"name":   tenant.Name,
			"server": url,
			"config": string(jsonData),
		}

		return err
	})
	if err != nil {
		return err
	}
	i.Log.V(5).Info("Argo Server created", "name", tenant.Name)

	return controllerutil.SetControllerReference(tenant, serverSecret, i.Client.Scheme())

}

func (i *TenancyController) tenantArgoProject(tenant *capsulev1beta2.Tenant, ctx context.Context) error {

	err := i.tenantArgoServer(tenant, ctx)
	if err != nil {
		return err
	}

	_, url := i.getProxyServiceName(tenant)

	// Provision Argo Project
	appProject := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "AppProject",
			"metadata": map[string]interface{}{
				"name": tenant.Name,
				"labels": map[string]string{
					"app.kubernetes.io/instance": "tenancy-controller",
					"app.kubernetes.io/name":     "tenancy-controller",
				},
				"namespace": "argocd", // Specify the namespace if needed
			},
			// Add Spec and other necessary fields following the AppProject CRD structure
			"spec": map[string]interface{}{
				"description": "Application Project " + tenant.Name,
				"roles":       []roles.ArgoProjectRole{},
			},
		},
	}

	_, err = controllerutil.CreateOrPatch(ctx, i.Client, appProject, func() error {
		pol := []roles.ArgoProjectRole{}

		// Owners
		owners := roles.ArgoProjectRole{
			Name:        "owners",
			Description: "Project Owners",
			Policies:    roles.ArgoOwnerPolicies(tenant.Name),
			Groups:      []string{},
		}
		for _, owner := range tenant.Spec.Owners {
			if owner.Kind == "User" || owner.Kind == "Group" {
				owners.Groups = append(owners.Groups, owner.Name)
			}
		}
		pol = append(pol, owners)

		// Maintainers
		maintainers := roles.ArgoProjectRole{
			Name:        "maintainers",
			Description: "Project Maintainers",
			Policies:    roles.ArgoMaintainerPolicies(tenant.Name),
			Groups:      []string{},
		}
		for _, binding := range tenant.Spec.AdditionalRoleBindings {
			if binding.ClusterRoleName == "tenant:maintainer" {
				for _, subject := range binding.Subjects {
					if subject.Kind == "User" || subject.Kind == "Group" {
						maintainers.Groups = append(maintainers.Groups, subject.Name)
					}
				}
			}
		}
		pol = append(pol, maintainers)

		// Operators
		operators := roles.ArgoProjectRole{
			Name:        "operators",
			Description: "Project Operators",
			Policies:    roles.ArgoOperatorPolicies(tenant.Name),
			Groups:      []string{},
		}
		for _, binding := range tenant.Spec.AdditionalRoleBindings {
			if binding.ClusterRoleName == "tenant:operator" {
				for _, subject := range binding.Subjects {
					if subject.Kind == "User" || subject.Kind == "Group" {
						maintainers.Groups = append(maintainers.Groups, subject.Name)
					}
				}
			}
		}
		pol = append(pol, operators)

		// Viewers
		viewers := roles.ArgoProjectRole{
			Name:        "viewers",
			Description: "Project Viewers",
			Policies:    roles.ArgoViewerPolicies(tenant.Name),
			Groups:      []string{},
		}
		for _, binding := range tenant.Spec.AdditionalRoleBindings {
			if binding.ClusterRoleName == "tenant:viewer" {
				for _, subject := range binding.Subjects {
					if subject.Kind == "User" || subject.Kind == "Group" {
						maintainers.Groups = append(maintainers.Groups, subject.Name)
					}
				}
			}
		}
		pol = append(pol, viewers)

		// Add All Roles
		appProject.Object["spec"].(map[string]interface{})["roles"] = pol

		// Assign other Properties (which should not be overwriten)
		appProject.Object["spec"].(map[string]interface{})["clusterResourceWhitelist"] = []map[string]interface{}{
			{
				"group": "*",
				"kind":  "*",
			},
		}
		appProject.Object["spec"].(map[string]interface{})["namespaceResourceWhitelist"] = []map[string]interface{}{
			{
				"group": "*",
				"kind":  "*",
			},
		}
		appProject.Object["spec"].(map[string]interface{})["sourceNamespaces"] = []string{tenant.Name + "-*"}
		appProject.Object["spec"].(map[string]interface{})["sourceRepos"] = []string{"*"}
		appProject.Object["spec"].(map[string]interface{})["destinations"] = []map[string]interface{}{
			{
				"name":      tenant.Name,
				"namespace": tenant.Name + "-*",
				"server":    url,
			},
		}

		return controllerutil.SetControllerReference(tenant, appProject, i.Client.Scheme())
	})
	if err != nil {
		return err
	}

	rbacCSV, err := roles.ArgoTenantCSV(url, tenant)
	if err != nil {
		return err
	}

	// Update existing configmap with new csv
	configmap := &corev1.ConfigMap{}
	err = i.Client.Get(ctx, client.ObjectKey{Name: "argocd-rbac-cm", Namespace: i.Options.ArgoCDNamespace}, configmap)
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

	i.Log.V(5).Info("Argo Project created", "name", tenant.Name)

	return nil
}

//func (i *TenancyController) tenantArgoCSV(tenant *capsulev1beta2.Tenant, ctx context.Context) error {
//	configmap := &corev1.ConfigMap{}
//	i.Client.Get(ctx, types.NamespacedName{Name: "argo", Namespace: i.Options.ArgoCDNamespace}, configmap)
//}
