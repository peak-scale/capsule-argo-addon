//nolint:all
package e2e_test

import (
	"context"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
)

var _ = Describe("Translation Test", func() {
	suiteSelector := e2eLabels("e2e_settings_translation")

	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}

	solar := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "solar-e2e-translation",
			Labels:      e2eLabels("e2e_settings_translation"),
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

	// Translator
	translator1 := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "e2e-translation-primary",
			Labels: suiteSelector,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: suiteSelector,
			},
			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
				{
					Name:         "viewer",
					ClusterRoles: []string{"admin"},
					Owner:        true,
					Policies: []v1alpha1.ArgocdPolicyDefinition{
						{
							Resource: "applications",
							Action:   []string{"get", "update", "delete"},
							Verb:     "allow",
						},
					},
				},
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
			Labels: suiteSelector,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: suiteSelector,
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
						DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
							{DefaultServiceAccount: "custom-serviceaccount", Server: "custom-server"},
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

	AfterEach(func() {
		Expect(CleanTenants(e2eSelector("e2e_settings_translation"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_settings_translation"))).ToNot(HaveOccurred())

		// Restore Configuration
		Eventually(func() error {
			if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: originalArgoAddon.Name}, argoaddon); err != nil {
				return err
			}

			// Apply the initial configuration from originalArgoAddon to argoaddon
			argoaddon.Spec = originalArgoAddon.Spec
			return k8sClient.Update(context.Background(), argoaddon)
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	})

	It("Does correctly translate", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.Proxy.Enabled = false
			argoaddon.Spec.Argo.DestinationServiceAccounts = false
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create single translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator1)).ToNot(HaveOccurred())
		})

		By("create matching tenant", func() {
			solar.ResourceVersion = ""
			Expect(k8sClient.Create(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("verify translated spec (primary translator)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			expected := &argocdv1alpha1.AppProjectSpec{
				PermitOnlyProjectScopedClusters: true,
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Name:      "custom-server",
						Namespace: "selected,namespaces",
					},
				},
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
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
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator1":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
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
				meta.TranslatorFinalizer(translator1.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

		By("verify primary translator status", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator1.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("create second translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator2)).ToNot(HaveOccurred())
		})

		By("verify translated spec (primary and secondary translator)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Expected Translation
			expected2 := &argocdv1alpha1.AppProjectSpec{
				PermitOnlyProjectScopedClusters: true,
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Name:      "custom-server",
						Namespace: "selected,namespaces",
					},
					{
						Name:      "some-other-server",
						Namespace: "tenant-*",
					},
				},
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
					{DefaultServiceAccount: "custom-serviceaccount", Server: "custom-server"},
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
						Kind:  "Pods",
					},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected2), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels2 := map[string]string{
				"translator1":           "label",
				"translator2":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
			}

			expectedAnnotations2 := map[string]string{
				"translator1": "annotation",
				"translator2": "annotation",
				"structured":  "exclusive",
				"override":    "structured",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedLabels2), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedAnnotations2), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers2 := []string{
				"resources-finalizer.argocd.argoproj.io",
				meta.TranslatorFinalizer(translator1.Name),
				meta.TranslatorFinalizer(translator2.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers2), "AppProject should have the expected finalizers")
		})

		By("verify secondary translator status", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator2.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("primary translator should no longer select tenant", func() {
			translator := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.TODO(), client.ObjectKey{Name: translator1.Name}, translator)).ToNot(HaveOccurred())

			translator.Spec.Selector.MatchLabels = map[string]string{
				"points.to.noewhere": "e2e",
			}
			Expect(k8sClient.Update(context.TODO(), translator)).ToNot(HaveOccurred())

		})

		By("primary translator should no longer select tenant (verify status)", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator1.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).To(BeNil(), "Tenant condition should not be nil")
		})

		By("verify translated approject (subtract primary translator)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Expected Translation
			expected := &argocdv1alpha1.AppProjectSpec{
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
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
					{DefaultServiceAccount: "custom-serviceaccount", Server: "custom-server"},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator2":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
			}

			expectedAnnotations := map[string]string{
				"translator2": "annotation",
				"structured":  "exclusive",
				"override":    "structured",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedLabels), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedAnnotations), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers := []string{
				meta.TranslatorFinalizer(translator2.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

		By("users can add additional approject specifications and they are preserved", func() {
			// Get Current Project
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Update Current Spec with customs stuff
			approject.Spec.Roles = []argocdv1alpha1.ProjectRole{
				{
					Name:        "ci-role",
					Description: "ci integration for teams",
					Groups: []string{
						"some-dev-team",
					},
				},
			}
			approject.Spec.SyncWindows = argocdv1alpha1.SyncWindows{
				&argocdv1alpha1.SyncWindow{
					Kind:         "allows",
					Duration:     "30m",
					Schedule:     "* * * * SUN",
					Applications: []string{"*"},
				},
			}

			// Add Stuff to translated Spec
			approject.Spec.ClusterResourceWhitelist = append(approject.Spec.ClusterResourceWhitelist, []metav1.GroupKind{
				{
					Group: "tenant.specific.crd",
					Kind:  "ApplicationCR",
				},
			}...)

			// Attempt to overwrite Nested translation spec
			approject.Spec.NamespaceResourceBlacklist = []metav1.GroupKind{
				{
					Group: "*",
					Kind:  "Overwritten",
				},
			}

			// Attempt to overrwrite primitive config controlled by translators
			approject.Spec.Description = "My new description"

			// Apply Approject and see what happened
			Expect(k8sClient.Update(context.Background(), approject)).To(Succeed())
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Expected Translation
			expected := &argocdv1alpha1.AppProjectSpec{
				Description:                     "My new description",
				PermitOnlyProjectScopedClusters: false,
				SyncWindows: argocdv1alpha1.SyncWindows{
					&argocdv1alpha1.SyncWindow{
						Kind:         "allows",
						Duration:     "30m",
						Schedule:     "* * * * SUN",
						Applications: []string{"*"},
					},
				},
				Roles: []argocdv1alpha1.ProjectRole{
					{
						Name:        "ci-role",
						Description: "ci integration for teams",
						Groups: []string{
							"some-dev-team",
						},
					},
				},
				Destinations: []argocdv1alpha1.ApplicationDestination{
					{
						Name:      "some-other-server",
						Namespace: "tenant-*",
					},
				},
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
					{DefaultServiceAccount: "custom-serviceaccount", Server: "custom-server"},
				},
				SourceNamespaces: []string{
					"a-second-place",
				},
				ClusterResourceWhitelist: []metav1.GroupKind{
					{
						Group: "vcluster.alhpa.com",
						Kind:  "Cluster",
					},
					{
						Group: "tenant.specific.crd",
						Kind:  "ApplicationCR",
					},
				},
				NamespaceResourceBlacklist: []metav1.GroupKind{
					{
						Group: "*",
						Kind:  "Overwritten",
					},
					{
						Group: "*",
						Kind:  "Pods",
					},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator2":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
			}

			expectedAnnotations := map[string]string{
				"translator2": "annotation",
				"structured":  "exclusive",
				"override":    "structured",
			}

			// Compare Metadata
			Expect(approject.Labels).To(Equal(expectedLabels), "AppProject should have the correct labels")
			Expect(approject.Annotations).To(Equal(expectedAnnotations), "AppProject should have the correct annotations")

			// Check for finalizers (assuming a finalizer example)
			expectedFinalizers := []string{
				meta.TranslatorFinalizer(translator2.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

		By("remove secondary translator", func() {
			Expect(k8sClient.Delete(context.TODO(), translator2)).ToNot(HaveOccurred())
		})

		By("ensure tenant assets are removed", func() {
			// Assets which should be absent
			expectedResources := []struct {
				object client.Object
				key    client.ObjectKey
				desc   string
			}{
				{
					object: &argocdv1alpha1.AppProject{},
					key:    client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace},
					desc:   "AppProject",
				},
				{
					object: &corev1.Secret{},
					key:    client.ObjectKey{Name: solar.Name, Namespace: argoaddon.Spec.Argo.Namespace},
					desc:   "Cluster Secret",
				},
				{
					object: &corev1.Service{},
					key:    client.ObjectKey{Name: solar.Name, Namespace: argoaddon.Spec.Proxy.CapsuleProxyServiceNamespace},
					desc:   "Service",
				},
				{
					object: &corev1.ServiceAccount{},
					key:    client.ObjectKey{Name: solar.Name, Namespace: argoaddon.Spec.Argo.ServiceAccountNamespace},
					desc:   "ServiceAccount",
				},
			}

			// Iterate through each resource to check it is no longer present
			for _, res := range expectedResources {
				By("Verifying " + res.desc + " resource is no longer present")
				err := k8sClient.Get(context.Background(), res.key, res.object)
				Expect(k8serrors.IsNotFound(err)).To(BeTrue(), "%s should not be found", res.desc)
			}

			// Tenant should no longer contain finalizer
			tnt := &capsulev1beta2.Tenant{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: solar.Name}, tnt)).ToNot(HaveOccurred())

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(tnt)).To(BeFalse(), "AppProject should contain translator finalizer")
		})

		By("Verify serviceaccount was removed as tenant owner", func() {
			tnt := &capsulev1beta2.Tenant{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: solar.Name}, tnt)).To(Succeed())

			// Expected Owners
			owners := []capsulev1beta2.OwnerSpec{
				{
					Name: "solar-users",
					Kind: capsulev1beta2.GroupOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
				{
					Name: "alice",
					Kind: capsulev1beta2.GroupOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
			}

			Expect(tnt.Spec.Owners).To(Equal(capsulev1beta2.OwnerListSpec(owners)), "Tenant should have serviceaccount as owner")
		})

	})

	It("Respect Read-Only", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.Proxy.Enabled = false
			argoaddon.Spec.Argo.DestinationServiceAccounts = false
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create single translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator1)).ToNot(HaveOccurred())
		})

		By("create matching tenant", func() {
			solar.SetAnnotations(map[string]string{
				meta.AnnotationProjectReadOnly: "true",
			})
			Expect(k8sClient.Create(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("verify translated spec (primary translator)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

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
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator1":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
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
				meta.TranslatorFinalizer(translator1.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})

		By("verify primary translator status", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator1.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("user can update approject (but changes are overwritten)", func() {
			// Get Current Project
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Update Current Spec with customs stuff
			approject.Spec.Roles = []argocdv1alpha1.ProjectRole{
				{
					Name:        "ci-role",
					Description: "ci integration for teams",
					Groups: []string{
						"some-dev-team",
					},
				},
			}
			approject.Spec.SyncWindows = argocdv1alpha1.SyncWindows{
				&argocdv1alpha1.SyncWindow{
					Kind:         "allows",
					Duration:     "30m",
					Schedule:     "* * * * SUN",
					Applications: []string{"*"},
				},
			}

			// Add Stuff to translated Spec
			approject.Spec.ClusterResourceWhitelist = append(approject.Spec.ClusterResourceWhitelist, []metav1.GroupKind{
				{
					Group: "tenant.specific.crd",
					Kind:  "ApplicationCR",
				},
			}...)

			// Attempt to overwrite Nested translation spec
			approject.Spec.NamespaceResourceBlacklist = []metav1.GroupKind{
				{
					Group: "*",
					Kind:  "Overwritten",
				},
			}

			// Attempt to overrwrite primitive config controlled by translators
			approject.Spec.Description = "My new description"

			// Apply Approject and see what happened
			Expect(k8sClient.Update(context.Background(), approject)).To(Succeed())
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

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
				DestinationServiceAccounts: []argocdv1alpha1.ApplicationDestinationServiceAccount{
					{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Server: argoaddon.Spec.Argo.Destination},
				},
			}

			// Compare the Spec
			Expect(approject.Spec).To(Equal(*expected), "AppProject spec should match the expected spec")

			// Finalizer should not contain the translator finalizer
			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")

			// Expected labels
			expectedLabels := map[string]string{
				"translator1":           "label",
				"structured":            "exclusive",
				"override":              "structured",
				meta.ManagedByLabel:     meta.ManagedByLabelValue,
				meta.ManagedTenantLabel: solar.Name,
				meta.ProvisionedByLabel: meta.ManagedByLabelValue,
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
				meta.TranslatorFinalizer(translator1.Name),
			}

			// Compare finalizers
			Expect(approject.Finalizers).To(Equal(expectedFinalizers), "AppProject should have the expected finalizers")
		})
	})
})
