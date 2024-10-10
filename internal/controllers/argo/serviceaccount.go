package argo

import (
	"context"
	"fmt"

	"github.com/peak-scale/capsule-argo-addon/internal/utils"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Creates Teanant Service Account with the given name and namespace
func (i *TenancyController) reconcileArgoServiceAccount(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
) (token string, err error) {

	// Get Required default values
	serviceAccount := tenant.Name
	namespace := i.Settings.Get().Proxy.ServiceAccountNamespace

	// Verify if ServiceAccount-Namespace is declared on tenant-basis
	if ns := utils.TenantServiceAccountNamespace(tenant); ns != "" {
		namespace = ns
	}

	accountResource := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccount,
			Namespace: namespace,
		},
	}

	if !i.Settings.Get().Proxy.Enabled || !utils.TenantProxyRegister(tenant) {
		err := i.Client.Delete(ctx, accountResource)
		if err != nil && !k8serrors.IsNotFound(err) {
			return "", fmt.Errorf("failed to lifecycle serviceaccount: %w", err)
		}
		return "", nil
	}

	// Create ServiceAccount
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, accountResource, func() (err error) {

		if accountResource.ObjectMeta.Labels == nil {
			accountResource.ObjectMeta.Labels = make(map[string]string)
		}
		accountResource.ObjectMeta.Labels = utils.TranslatorTrackingLabels(tenant)

		return controllerutil.SetOwnerReference(tenant, accountResource, i.Client.Scheme())
	})
	if err != nil {
		return "", err
	}

	// Lifecycle Old Serviceaccount
	//if r == controllerutil.OperationResultCreated {
	//
	//}

	tokenResource := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceAccount,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	// Create Account Token
	_, err = controllerutil.CreateOrUpdate(ctx, i.Client, tokenResource, func() (err error) {
		tokenResource.ObjectMeta.Labels = utils.TranslatorTrackingLabels(tenant)

		if tokenResource.ObjectMeta.Annotations == nil {
			tokenResource.ObjectMeta.Annotations = make(map[string]string)
		}
		tokenResource.ObjectMeta.Annotations["kubernetes.io/service-account.name"] = serviceAccount

		return controllerutil.SetOwnerReference(accountResource, tokenResource, i.Client.Scheme())
	})
	if err != nil {
		return "", err
	}

	var secret corev1.Secret
	if err = i.Client.Get(ctx, client.ObjectKey{
		Name:      tokenResource.Name,
		Namespace: namespace,
	}, &secret); err != nil {
		return "", err
	}

	t, exists := secret.Data["token"]
	if !exists {
		return "", err
	}

	token = string(t)

	err = i.addServiceAccountOwner(ctx, tenant, namespace, serviceAccount)
	if err != nil {
		return "", err
	}

	i.Log.V(5).Info("SeriviceAccount created", "name", tenant.Name)

	return
}

// Adds the given service account as an owner to the tenant
func (i *TenancyController) addServiceAccountOwner(
	ctx context.Context,
	tenant *capsulev1beta2.Tenant,
	namespace string,
	name string,
) (err error) {
	owner := capsulev1beta2.OwnerSpec{
		Kind: "ServiceAccount",
		Name: "system:serviceaccount:" + namespace + ":" + name,
	}

	// Check if the owner is already present
	for _, o := range tenant.Spec.Owners {
		if o.Kind == owner.Kind && o.Name == owner.Name {
			return nil
		}
	}

	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (conflict error) {
		_ = i.Client.Get(ctx, types.NamespacedName{Name: tenant.Name}, tenant)

		tenant.Spec.Owners = append(tenant.Spec.Owners, owner)
		if conflict = i.Client.Update(ctx, tenant); err != nil {
			return err
		}
		return
	})

	return
}
