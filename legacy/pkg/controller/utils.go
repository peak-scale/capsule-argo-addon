package controller

import (
	"context"
	"fmt"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
)

func (i *TenancyController) getProxyServiceName(tenant *capsulev1beta2.Tenant) (service string, url string) {
	serviceName := tenant.Name + "-proxy"

	return serviceName, fmt.Sprintf("https://%s.%s.svc:9001", serviceName, i.Options.CapsuleProxyServiceNamespace)
}

func (i *TenancyController) addServiceAccountOwner(namespace string, name string, tenant *capsulev1beta2.Tenant, ctx context.Context) (err error) {
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
