package e2e

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
)

var _ = Describe("Tenant Labels, Annotations & Finalizers", func() {
	argoaddon := &v1alpha1.ArgoAddon{}

	// Create a Translator for all the tests
	translator1 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-annotations-1",
			Labels: e2eLabels,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/type": "prod",
				},
			},
		},
	}
	translator2 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-annotations-2",
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
		},
	}

	// Create Tenants
	solar := tntSolar
	solar.Labels["app.kubernetes.io/type"] = "dev"
	solar.Annotations = map[string]string{}

	oil := tntOil
	oil.Labels["app.kubernetes.io/type"] = "prod"
	oil.Annotations = map[string]string{}

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
	It("should add finalizers for matched translators", func() {
		By("configuration alignment", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, argoaddon)
			argoaddon.Spec.TranslatorSelector = &metav1.LabelSelector{
				MatchLabels: e2eLabels,
			}
			argoaddon.Spec.ArgoCD.Namespace = "argocd"
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("appproject (solar)", func() {
			approject := &argocdv1alpha1.AppProject{}
			err := k8sClient.Get(context.Background(), client.ObjectKey{
				Name:      solar.GetName(),
				Namespace: argoaddon.Spec.ArgoCD.Namespace,
			}, approject)
			Expect(err).ToNot(HaveOccurred())

			expectedProjectFinalizers := []string{
				translator.TranslatorFinalizer(translator1),
				translator.TranslatorFinalizer(translator2),
			}

			for _, finalizer := range expectedProjectFinalizers {
				Expect(approject.GetFinalizers()).To(ContainElement(finalizer), "Missing expected finalizer: %s", finalizer)
			}
		})

	})
})
