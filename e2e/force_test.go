package e2e_test

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Force Appproject", func() {
	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}
	solar := tntSolar.DeepCopy()
	solar.Name = "solar-e2e-force"
	translators := []*v1alpha1.ArgoTranslator{baseTranslator}

	BeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()

		for _, tran := range translators {
			Eventually(func() error {
				tran.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), tran)
			}).Should(Succeed())
		}

	})

	AfterEach(func() {
		Expect(k8sClient.Delete(context.TODO(), solar)).Should(Succeed())
		Expect(k8sClient.Delete(context.TODO(), baseTranslator)).Should(Succeed())

		// Restore Configuration
		Eventually(func() error {
			err := retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
				k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
				argoaddon.Spec = originalArgoAddon.DeepCopy().Spec

				err = k8sClient.Update(context.Background(), argoaddon)
				return
			})

			return err
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	})

	It("Does overwrite appproject", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.Force = true
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create solar appproject", func() {
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels(),
					Annotations: map[string]string{
						"force": "true",
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), appproject)).To(Succeed())
		})

		By("create tenant solar", func() {
			Eventually(func() error {
				solar.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("Verify approject was adopted", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(translator.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")
		})
	})

	It("Does not overwrite appproject", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.Force = false
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create solar appproject", func() {
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels(),
					Annotations: map[string]string{
						"force": "false",
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), appproject)).To(Succeed())
		})

		By("create tenant solar", func() {
			Eventually(func() error {
				solar.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("Verify approject was not adopted", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(translator.ContainsTranslatorFinalizer(approject)).To(BeFalse(), "AppProject should not contain translator finalizer")
		})

		By("Attempting to delete the AppProject", func() {
			appproject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, appproject)).To(Succeed())

			// Try to delete the resource and log any errors
			err := k8sClient.Delete(context.TODO(), appproject)
			fmt.Printf("Delete Error: %v\n", err)
			Expect(err).ShouldNot(HaveOccurred(), "Failed to delete AppProject")
		})

	})
})
