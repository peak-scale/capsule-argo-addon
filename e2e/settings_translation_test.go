package e2e_test

import (
	"context"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Translation Test", func() {
	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}
	solar := tntSolar.DeepCopy()
	solar.Name = "solar-e2e-translation"

	oil := tntSolar.DeepCopy()
	oil.Name = "oil-e2e-translation"

	// Translator
	baseTranslator.Spec.ProjectSettings = v1alpha1.ArgocdProjectProperties{
		Structured: v1alpha1.ArgocdProjectStructuredProperties{
			ProjectMeta: v1alpha1.ArgocdProjectPropertieMeta{
				Labels: map[string]string{
					"structured": "exclusive",
					"override":   "structured",
				},
				Annotations: map[string]string{
					"structured": "exclusive",
					"override":   "structured",
				},
				Finalizers: []string{
					"resources-finalizer.argocd.argoproj.io",
				},
			},
			ProjectSpec: argocdv1alpha1.AppProjectSpec{
				PermitOnlyProjectScopedClusters: true,
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Name:      "custom-server",
						Namespace: "selected,namespaces",
					},
				},
				SourceNamespaces: []string{
					"somewhere",
				},
				ClusterResourceWhitelist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "ConfigMap",
					},
				},
			},
		},
		Template: `
meta:
  labels:
	templated: "exclusive"
	override:   "templated"
  annotations:
	templated: "exclusive"
	override:   "templated"
  finalizers:
    - "my-custom.argocd.finalizer.io"
spec:
  description: "Projetc {{ $.Tenant.Name }}
  destinations:
    - name: "template-cluster"
      namespace: "{{ $.Proxy }}"
  clusterResourceWhitelist:
    - group: "*"
      kind: "*"
  namespaceResourceWhitelist:
    - group: "*"
      kind: "Pod"
  sourceNamespaces:
    - {{ $.Config.Argo.Namespace | quote }}
  {{- range $_, $value := $.Tenant.Namespaces }}
    - {{ $value | quote }}
  {{- end }}`,
	}

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
		// Restore Configuration
		Eventually(func() error {
			argoaddon = originalArgoAddon.DeepCopy()
			return k8sClient.Update(context.Background(), argoaddon)
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())

		// Define Resources which are lifecycled after each test
		resourcesToClean := []client.Object{
			&v1alpha1.ArgoTranslator{},
			&argocdv1alpha1.AppProject{},
			&capsulev1beta2.Tenant{},
		}

		Eventually(func() error {
			return cleanResources(resourcesToClean, e2eSelector())
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	})

	It("Does correctly translate", func() {

		By("Verify approject was not adopted", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Expected Translation
			expected := &argocdv1alpha1.AppProjectSpec{
				PermitOnlyProjectScopedClusters: true,
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Name:      "custom-server",
						Namespace: "selected,namespaces",
					},
				},
				SourceNamespaces: []string{
					"somewhere",
				},
				ClusterResourceWhitelist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "ConfigMap",
					},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(translator.ContainsTranslatorFinalizer(approject)).To(BeFalse(), "AppProject should not contain translator finalizer")

			// Expected labels
			expectedMeta := map[string]string{
				"structured": "exclusive",
				"templated":  "exclusive",
				"override":   "templated",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedMeta), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedMeta), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers := []string{
				"resources-finalizer.argocd.argoproj.io",
				"my-custom.argocd.finalizer.io",
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})
	})
})
