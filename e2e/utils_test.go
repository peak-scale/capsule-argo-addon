package e2e_test

import (
	"context"
	"fmt"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultTimeoutInterval = 20 * time.Second
	defaultPollInterval    = time.Second
)

func e2eConfigName() string {
	return "default"
}

// Returns labels to identify e2e resources.
func e2eLabels() map[string]string {
	return map[string]string{
		"argo.addons.projectcapsule.dev/e2e": "true",
	}
}

// Returns a label selector to filter e2e resources.
func e2eSelector() labels.Selector {
	return labels.SelectorFromSet(e2eLabels())
}

// Pass objects which require cleanup and a label selector to filter them.
func cleanResources(res []client.Object, selector labels.Selector) (err error) {
	for _, resource := range res {
		err = k8sClient.DeleteAllOf(context.TODO(), resource, &client.MatchingLabels{"argo.addons.projectcapsule.dev/e2e": "true"})

		if err != nil {
			return err
		}
	}

	return nil
}

func CleanAppProjects(selector labels.Selector, namespace string) error {
	res := &argocdv1alpha1.AppProjectList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// If a namespace is provided, set it in the list options
	if namespace != "" {
		listOptions.Namespace = namespace
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete resource %s: %w", app.GetName(), err)
		}
	}

	return nil
}

// Base Structs for Testing
var tntSolar = &capsulev1beta2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "solar-e2e",
		Labels:      e2eLabels(),
		Annotations: map[string]string{},
	},
	Spec: capsulev1beta2.TenantSpec{
		AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
			{
				ClusterRoleName: "tenant-viewer",
				Subjects: []rbacv1.Subject{
					{
						Kind: "User",
						Name: "bob",
					},
					{
						Name: "solar-users",
						Kind: "Group",
					},
				},
			},

			{
				ClusterRoleName: "tenant-operators",
				Subjects: []rbacv1.Subject{
					{
						Kind: "User",
						Name: "gatsby",
					},
					{
						Name: "operators",
						Kind: "Group",
					},
				},
			},
		},
		Owners: []capsulev1beta2.OwnerSpec{
			{
				Name: "solar-users",
				Kind: capsulev1beta2.GroupOwner,
			},
			{
				Name: "alice",
				Kind: capsulev1beta2.GroupOwner,
			},
		},
	},
}

var tntOil = &capsulev1beta2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "oil-e2e",
		Labels:      e2eLabels(),
		Annotations: map[string]string{},
	},
	Spec: capsulev1beta2.TenantSpec{
		AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
			{
				ClusterRoleName: "tenant-viewer",
				Subjects: []rbacv1.Subject{
					{
						Kind: "User",
						Name: "bob",
					},
					{
						Name: "oil-users",
						Kind: "Group",
					},
					{
						Name: "operators",
						Kind: "Group",
					},
				},
			},
		},
		Owners: []capsulev1beta2.OwnerSpec{
			{
				Name: "solar-users",
				Kind: capsulev1beta2.GroupOwner,
			},
			{
				Name: "alice",
				Kind: capsulev1beta2.GroupOwner,
			},
		},
	},
}

var tntWind = &capsulev1beta2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "wind-e2e",
		Labels:      e2eLabels(),
		Annotations: map[string]string{},
	},
	Spec: capsulev1beta2.TenantSpec{
		AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
			{
				ClusterRoleName: "tenant-viewer",
				Subjects: []rbacv1.Subject{
					{
						Kind: "User",
						Name: "bob",
					},
					{
						Name: "solar-users",
						Kind: "Group",
					},
				},
			},
		},
		Owners: []capsulev1beta2.OwnerSpec{
			{
				Name: "solar-users",
				Kind: capsulev1beta2.GroupOwner,
			},
			{
				Name: "alice",
				Kind: capsulev1beta2.GroupOwner,
			},
		},
	},
}

var baseTranslator = &v1alpha1.ArgoTranslator{
	ObjectMeta: metav1.ObjectMeta{
		Name:   "e2e-base-translator",
		Labels: e2eLabels(),
	},
	Spec: v1alpha1.ArgoTranslatorSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: e2eLabels(),
		},
	},
}
