package e2e_test

import (
	"context"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/controllers/translator"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
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
	translator1 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "e2e-translation-primary",
			Labels: e2eLabels(),
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: e2eLabels(),
			},
			ProjectSettings: v1alpha1.ArgocdProjectProperties{
				Structured: v1alpha1.ArgocdProjectStructuredProperties{
					ProjectMeta: v1alpha1.ArgocdProjectPropertieMeta{
						Labels: map[string]string{
							"translator1": "label",
							"structured":  "exclusive",
							"override":    "structured",
						},
						Annotations: map[string]string{
							"translator1": "annotation",
							"structured":  "exclusive",
							"override":    "structured",
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
			},
		},
	}

	translator2 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "e2e-translation-secondary",
			Labels: e2eLabels(),
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: e2eLabels(),
			},
			ProjectSettings: v1alpha1.ArgocdProjectProperties{
				Structured: v1alpha1.ArgocdProjectStructuredProperties{
					ProjectMeta: v1alpha1.ArgocdProjectPropertieMeta{
						Labels: map[string]string{
							"translator2": "label",
							"structured":  "exclusive",
							"override":    "structured",
						},
						Annotations: map[string]string{
							"translator2": "annotation",
							"structured":  "exclusive",
							"override":    "structured",
						},
						Finalizers: []string{
							"resources-finalizer.argocd.argoproj.io",
						},
					},
					ProjectSpec: argocdv1alpha1.AppProjectSpec{
						PermitOnlyProjectScopedClusters: false,
						Destinations: []argocdv1alpha1.ApplicationDestination{
							{
								Name:      "some-other-server",
								Namespace: "tenant-*",
							},
						},
						SourceNamespaces: []string{
							"a-second-place",
						},
						ClusterResourceWhitelist: []metav1.GroupKind{
							{
								Group: "vcluster.alhpa.com",
								Kind:  "Cluster",
							},
						},
						NamespaceResourceBlacklist: []metav1.GroupKind{
							{
								Group: "*",
								Kind:  "Pods",
							},
						},
					},
				},
			},
		},
	}

	BeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()
	})

	//AfterEach(func() {
	//	Expect(CleanTenants(e2eSelector())).ToNot(HaveOccurred())
	//
	//	// Define Resources which are lifecycled after each test
	//	resourcesToClean := []client.Object{
	//		&v1alpha1.ArgoTranslator{},
	//		&capsulev1beta2.Tenant{},
	//	}
	//
	//	Eventually(func() error {
	//		return cleanResources(resourcesToClean, e2eSelector())
	//	}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	//
	//	// Restore Configuration
	//	Eventually(func() error {
	//		argoaddon = originalArgoAddon.DeepCopy()
	//		return k8sClient.Update(context.Background(), argoaddon)
	//	}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	//
	//})

	It("Does correctly translate", func() {
		By("create single translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator1)).ToNot(HaveOccurred())
		})

		By("create matching tenant", func() {
			solar.ResourceVersion = ""
			Expect(k8sClient.Create(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("verify translated spec", func() {
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
					{
						Server:    argoaddon.Spec.ProxyServiceString(solar),
						Name:      solar.Name,
						Namespace: "*",
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
			Expect(translator.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator1":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
			}

			expectedAnnotations := map[string]string{
				"translator1": "annotation",
				"structured":  "exclusive",
				"override":    "structured",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedLabels), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedAnnotations), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers := []string{
				"resources-finalizer.argocd.argoproj.io",
				translator.TranslatorFinalizer(translator1),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

		//By("create second translation", func() {
		//	Expect(k8sClient.Create(context.TODO(), translator2)).ToNot(HaveOccurred())
		//})

		By("verify translated spec", func() {
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
					{
						Server:    argoaddon.Spec.ProxyServiceString(solar),
						Name:      solar.Name,
						Namespace: "*",
					},
					{
						Name:      "some-other-server",
						Namespace: "tenant-*",
					},
				},
				SourceNamespaces: []string{
					"somewhere",
					"a-second-place",
				},
				ClusterResourceWhitelist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "ConfigMap",
					},
					{
						Group: "vcluster.alhpa.com",
						Kind:  "Cluster",
					},
				},
				NamespaceResourceBlacklist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "ConfigMap",
					},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(translator.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator1":           "label",
				"translator2":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
			}

			expectedAnnotations := map[string]string{
				"translator1": "annotation",
				"translator2": "annotation",
				"structured":  "exclusive",
				"override":    "structured",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedLabels), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedAnnotations), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers := []string{
				"resources-finalizer.argocd.argoproj.io",
				translator.TranslatorFinalizer(translator1),
				translator.TranslatorFinalizer(translator2),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

	})
})
