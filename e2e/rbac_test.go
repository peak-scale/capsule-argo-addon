package e2e

import (
	"context"
	"fmt"

	"github.com/bsm/gomega"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/peak-scale/capsule-argo-addon/internal/argo"
)

var _ = Describe("Argo RBAC Reflection", func() {
	argoaddon := &v1alpha1.ArgoAddon{}

	// Create a Translator for all the tests
	translator1 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-rbac-1",
			Labels: e2eLabels,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/type": "prod",
				},
			},
			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
				{
					Name:         "viewer",
					ClusterRoles: []string{"tenant-viewer"},
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "applications",
							Action:   []string{"get", "update", "delete"},
							Verb:     "allow",
						},
					},
				},
				{
					Name:         "owner",
					ClusterRoles: []string{"admin"},
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "repositories",
							Action:   []string{"*"},
							Verb:     "allow",
						},
					},
					Owner: true,
				},
			},
		},
	}
	translator2 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-rbac-2",
			Labels: e2eLabels,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app.kubernetes.io/type",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"dev", "prod"},
					},
				},
			},

			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
				{
					Name: "operators",
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "applications",
							Action:   []string{"get", "update", "delete"},
							Verb:     "allow",
						},
					},
				},
				{
					Name: "owner",
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "repositories",
							Action:   []string{"*"},
							Verb:     "allow",
						},
					},
				},
			},
		},
	}

	// Create Tenants
	solar := tntSolar
	solar.Name = "solar-rbac-e2e"
	solar.Labels["app.kubernetes.io/type"] = "dev"

	oil := tntOil
	oil.Name = "oil-rbac-e2e"
	oil.Labels["app.kubernetes.io/type"] = "dev"

	JustBeforeEach(func() {
		for _, tran := range []*v1alpha1.ArgoTranslator{translator1, translator2} {
			Eventually(func() error {
				tran.ResourceVersion = ""
				return k8sClient.Create(context.TODO(), tran)
			}).Should(Succeed())
		}

		for _, tnt := range []*capsulev1beta2.Tenant{solar, oil} {
			Eventually(func() error {
				tnt.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), tnt)
			}).Should(Succeed())
		}
	})
	JustAfterEach(func() {
		translators := &v1alpha1.ArgoTranslatorList{}
		tenants := &capsulev1beta2.TenantList{}

		Eventually(func() (err error) {
			return k8sClient.DeleteAllOf(context.TODO(), &v1alpha1.ArgoTranslator{}, &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: e2eSelector(),
				},
			})
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		fmt.Printf("Found %d ArgoTranslators to delete\n", len(translators.Items))

		Eventually(func() (err error) {
			return k8sClient.DeleteAllOf(context.TODO(), &capsulev1beta2.Tenant{}, &client.DeleteAllOfOptions{
				ListOptions: client.ListOptions{
					LabelSelector: e2eSelector(),
				},
			})
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		fmt.Printf("Found %d Tenants to delete\n", len(tenants.Items))
	})

	// Test case for ensuring the tenant is created successfully
	It("Correctly Reflect Argo RBAC", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, argoaddon)
			argoaddon.Spec.TranslatorSelector = &metav1.LabelSelector{
				MatchLabels: e2eLabels,
			}
			argoaddon.Spec.Argo.Namespace = "argocd"
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("verify argo rbac permissions csv (solar)", func() {

			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{
				Name:      argoaddon.Spec.Argo.RBACConfigMap,
				Namespace: argoaddon.Spec.Argo.Namespace,
			}, configmap)).To(Succeed())

			rbacSolar, ok := configmap.Data[argo.ArgoPolicyName(solar)]
			Expect(ok).To(.BeTrue(), "RBAC CSV entry for solar is missing in ConfigMap")

			// Extract CSV
			if rbacSolar, ok := configmap.Data[argo.ArgoPolicyName(solar)]; ok {

			}
		})

	})
})
