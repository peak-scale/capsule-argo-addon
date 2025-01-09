//nolint:all
package e2e_test

import (
	"context"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/argo"
)

var _ = Describe("Argo RBAC Reflection", func() {
	// Resources
	selector := e2eLabels("e2e_argo_rbac")

	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}

	// Create a Translator for all the tests
	translator1 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-rbac-1",
			Labels: selector,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "app.kubernetes.io/env",
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{"dev", "prod"},
					},
				},
			},
			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
				{
					Name:         "viewer",
					ClusterRoles: []string{"tenant-viewer", "tenant-operators"},
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
			Labels: selector,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/env": "dev",
				},
			},

			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
				{
					Name: "operators",
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "*",
							Action:   []string{"get"},
							Verb:     "allow",
						},
					},
				},
				{
					Name: "owner",
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "applications",
							Action:   []string{"*"},
							Verb:     "allow",
						},
					},
				},
			},
		},
	}

	// Create Tenants
	solar := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "solar-rbac-e2e",
			Labels: map[string]string{
				e2eLabel:                "true",
				suiteLabel:              "e2e_argo_rbac",
				"app.kubernetes.io/env": "dev",
			},
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

	oil := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "oil-rbac-e2e",
			Labels: map[string]string{
				e2eLabel:                "true",
				suiteLabel:              "e2e_argo_rbac",
				"app.kubernetes.io/env": "prod",
			},
			Annotations: map[string]string{},
		},
		Spec: capsulev1beta2.TenantSpec{
			AdditionalRoleBindings: []capsuleapi.AdditionalRoleBindingsSpec{
				{
					ClusterRoleName: "tenant-viewer",
					Subjects: []rbacv1.Subject{
						{
							Kind: "User",
							Name: "alice",
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
					Name: "bob",
					Kind: capsulev1beta2.GroupOwner,
				},
			},
		},
	}

	JustBeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()

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
		Expect(CleanTenants(e2eSelector("e2e_argo_rbac"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_argo_rbac"))).ToNot(HaveOccurred())

		// Restore Configuration
		Eventually(func() error {
			argoaddon = originalArgoAddon.DeepCopy()
			return k8sClient.Update(context.Background(), argoaddon)
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	})

	// Test case for ensuring the tenant is created successfully
	It("Correctly Reflect Argo RBAC", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, argoaddon)
			argoaddon.Spec.Argo.Namespace = "argocd"
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("verify argo default rbac permissions csv (solar)", func() {

			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{
				Name:      argoaddon.Spec.Argo.RBACConfigMap,
				Namespace: argoaddon.Spec.Argo.Namespace,
			}, configmap)).To(Succeed())

			rbacSolar, ok := configmap.Data[argo.ArgoPolicyName(solar)]
			Expect(ok).To(BeTrue(), "RBAC CSV entry for solar is missing in ConfigMap")

			// Define Which Lines we are expecting in the CSV
			expectedLines := append(argo.DefaultPolicies(solar, argoaddon.Spec.Argo.Destination), []string{
				argo.PolicyString(argo.TenantPolicy(solar, "viewer"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"get"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(solar, "viewer"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"update"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(solar, "viewer"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"delete"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(solar, "operators"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "*",
						Action:   []string{"get"},
						Verb:     "allow",
						Path:     "*",
					}),

				argo.PolicyString(argo.TenantPolicy(solar, "owner"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "repositories",
						Action:   []string{"*"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(solar, "owner"),
					solar.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"*"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.BindingString(rbacv1.Subject{
					Name: "alice",
					Kind: "User",
				},
					argo.DefaultPolicyReadOnly(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "alice",
					Kind: "User",
				},
					argo.DefaultPolicyOwner(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "solar-users",
					Kind: "Group",
				},
					argo.DefaultPolicyReadOnly(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "solar-users",
					Kind: "Group",
				},
					argo.DefaultPolicyOwner(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "solar-users",
					Kind: "Group",
				},
					argo.TenantPolicy(solar, "viewer"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "bob",
					Kind: "User",
				},
					argo.DefaultPolicyReadOnly(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "bob",
					Kind: "User",
				},
					argo.TenantPolicy(solar, "viewer"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "operators",
					Kind: "Group",
				},
					argo.DefaultPolicyReadOnly(solar),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "operators",
					Kind: "Group",
				},
					argo.TenantPolicy(solar, "viewer"),
				),
			}...)

			var extractedLines []string
			for _, line := range strings.Split(rbacSolar, "\n") {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine != "" {
					extractedLines = append(extractedLines, trimmedLine)
				}
			}

			By("verifying each expected line exists in the extracted CSV")
			var missingLines []string
			for _, expectedLine := range expectedLines {
				success, err := ContainElement(strings.TrimSpace(expectedLine)).Match(extractedLines)
				if err != nil {
					Fail(fmt.Sprintf("Error checking line presence: %v", err))
				}
				if !success {
					missingLines = append(missingLines, expectedLine)
				}
			}

			// Fail the test with details on missing lines, if any
			Expect(missingLines).To(BeEmpty(), "missing expected CSV lines: %v", missingLines)
		})

		By("verify argo default rbac permissions csv (oil)", func() {

			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{
				Name:      argoaddon.Spec.Argo.RBACConfigMap,
				Namespace: argoaddon.Spec.Argo.Namespace,
			}, configmap)).To(Succeed())

			rbacOil, ok := configmap.Data[argo.ArgoPolicyName(oil)]
			Expect(ok).To(BeTrue(), "RBAC CSV entry for oil is missing in ConfigMap")

			// Define Which Lines we are expecting in the CSV
			expectedLines := append(argo.DefaultPolicies(oil, argoaddon.Spec.Argo.Destination), []string{
				argo.PolicyString(argo.TenantPolicy(oil, "viewer"),
					oil.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"get"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(oil, "viewer"),
					oil.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"update"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(oil, "viewer"),
					oil.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "applications",
						Action:   []string{"delete"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.PolicyString(argo.TenantPolicy(oil, "owner"),
					oil.Name,
					addonsv1alpha1.ArgocdPolicyDefinition{
						Resource: "repositories",
						Action:   []string{"*"},
						Verb:     "allow",
						Path:     "*",
					}),
				argo.BindingString(rbacv1.Subject{
					Name: "alice",
					Kind: "User",
				},
					argo.DefaultPolicyReadOnly(oil),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "alice",
					Kind: "User",
				},
					argo.TenantPolicy(oil, "viewer"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "oil-users",
					Kind: "Group",
				},
					argo.DefaultPolicyReadOnly(oil),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "oil-users",
					Kind: "Group",
				},
					argo.TenantPolicy(oil, "viewer"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "operators",
					Kind: "Group",
				},
					argo.DefaultPolicyReadOnly(oil),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "operators",
					Kind: "Group",
				},
					argo.TenantPolicy(oil, "viewer"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "solar-users",
					Kind: "Group",
				},
					argo.DefaultPolicyReadOnly(oil),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "solar-users",
					Kind: "Group",
				},
					argo.DefaultPolicyOwner(oil),
				),

				argo.BindingString(rbacv1.Subject{
					Name: "bob",
					Kind: "User",
				},
					argo.TenantPolicy(oil, "owner"),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "bob",
					Kind: "User",
				},
					argo.DefaultPolicyReadOnly(oil),
				),
				argo.BindingString(rbacv1.Subject{
					Name: "bob",
					Kind: "User",
				},
					argo.DefaultPolicyOwner(oil),
				),
			}...)

			var extractedLines []string
			for _, line := range strings.Split(rbacOil, "\n") {
				trimmedLine := strings.TrimSpace(line)
				if trimmedLine != "" {
					extractedLines = append(extractedLines, trimmedLine)
				}
			}

			By("verifying each expected line exists in the extracted CSV")
			var missingLines []string
			for _, expectedLine := range expectedLines {
				success, err := ContainElement(strings.TrimSpace(expectedLine)).Match(extractedLines)
				if err != nil {
					Fail(fmt.Sprintf("Error checking line presence: %v", err))
				}
				if !success {
					missingLines = append(missingLines, expectedLine)
				}
			}

			// Fail the test with details on missing lines, if any
			Expect(missingLines).To(BeEmpty(), "missing expected CSV lines: %v", missingLines)
		})

		By("remove solar tenant", func() {
			Expect(k8sClient.Delete(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("verify argo rbac was finalized", func() {
			configmap := &corev1.ConfigMap{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{
				Name:      argoaddon.Spec.Argo.RBACConfigMap,
				Namespace: argoaddon.Spec.Argo.Namespace,
			}, configmap)).To(Succeed())

			_, ok := configmap.Data[argo.ArgoPolicyName(solar)]
			Expect(ok).To(BeFalse(), "RBAC CSV entry for solar should be missing in ConfigMap")
		})

	})
})
