package e2e

import (
	"time"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	defaultTimeoutInterval = 20 * time.Second
	defaultPollInterval    = time.Second
)

var e2eLabels = map[string]string{
	"e2e.argoproj.io/test": "true",
}

func e2eSelector() labels.Selector {
	reqs, _ := labels.NewRequirement("e2e.argoproj.io/test", selection.Equals, []string{"true"})
	return labels.NewSelector().Add(*reqs)
}

// Base Structs for Testing
var tntSolar = &capsulev1beta2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "solar-e2e",
		Labels:      e2eLabels,
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
		Labels:      e2eLabels,
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

var tntWind = &capsulev1beta2.Tenant{
	ObjectMeta: metav1.ObjectMeta{
		Name:        "wind-e2e",
		Labels:      e2eLabels,
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
