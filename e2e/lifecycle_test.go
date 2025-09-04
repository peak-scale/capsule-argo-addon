//nolint:all
package e2e_test

import (
	"context"
	"fmt"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("lifecycle Appproject", func() {
	selector := e2eLabels("e2e_lifecycle")
	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}

	solar := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "solar-lifecycle-e2e",
			Labels:      selector,
			Annotations: map[string]string{},
		},
		Spec: capsulev1beta2.TenantSpec{
			Owners: []capsulev1beta2.OwnerSpec{
				{
					Name: "alice",
					Kind: capsulev1beta2.GroupOwner,
				},
			},
		},
	}

	translator := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-lifecycle",
			Labels: selector,
		},
		Spec: v1alpha1.ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					suiteLabel: "e2e_lifecycle",
				},
			},
			ProjectRoles: []v1alpha1.ArgocdProjectRolesTranslator{
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
			ProjectSettings: &v1alpha1.ArgocdProjectProperties{
				Structured: &v1alpha1.ArgocdProjectStructuredProperties{
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

	JustBeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()

		// Cleanup old resources
		Expect(CleanTenants(e2eSelector("e2e_lifecycle"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_lifecycle"))).ToNot(HaveOccurred())
		Expect(CleanAppProjects(e2eSelector("e2e_lifecycle"), "argocd")).ToNot(HaveOccurred())

		for _, tran := range []*v1alpha1.ArgoTranslator{translator} {
			Eventually(func() error {
				tran.ResourceVersion = ""
				return k8sClient.Create(context.TODO(), tran)
			}).Should(Succeed())
		}
	})

	JustAfterEach(func() {
		Expect(CleanTenants(e2eSelector("e2e_lifecycle"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_lifecycle"))).ToNot(HaveOccurred())
		Expect(CleanAppProjects(e2eSelector("e2e_lifecycle"), "argocd")).ToNot(HaveOccurred())

		// Delete loose items
		expectedResources := []struct {
			object    client.Object
			desc      string
			name      string
			namespace string
		}{
			{
				object:    &argocdv1alpha1.AppProject{},
				desc:      "AppProject",
				name:      meta.TenantProjectName(solar),
				namespace: argoaddon.Spec.Argo.Namespace,
			},
			{
				object:    &corev1.ServiceAccount{},
				desc:      "ServiceAccount",
				name:      solar.Name,
				namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
			},
		}

		for _, res := range expectedResources {
			By("Deleting " + res.desc)

			// First, attempt to get the resource to ensure it exists before deletion
			err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
			if err != nil {
				continue
			}

			// Attempt to delete the resource
			err = k8sClient.Delete(context.Background(), res.object)
			Expect(err).To(Succeed(), "Expected to delete %s", res.desc)
		}

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

	It("Test lifecycle Settings (with Force)", func() {
		By("set corresponding settings", func() {
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)).To(Succeed())
			argoaddon.Spec.Force = true

			// Attempt to update the argoaddon object in Kubernetes
			err := k8sClient.Update(context.Background(), argoaddon)
			if err != nil {
				fmt.Printf("Error updating argoaddon: %v\n", err)
			}
			Expect(err).To(Succeed(), "Failed to update argoaddon")
		})

		By("Check existence resources", func() {

			// Approject
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels("e2e_lifecycle"),
				},
			}

			// Add deferred logging to capture `approject` if the test fails
			defer func() {
				if approjErr := k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, appproject); approjErr == nil {
					fmt.Printf("AppProject state at failure: %+v\n", appproject)
				}
			}()
		})

		By("precreate solar resources", func() {

			// Approject
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels("e2e_lifecycle"),
				},
			}

			// Add deferred logging to capture `approject` if the test fails
			defer func() {
				if approjErr := k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, appproject); approjErr == nil {
					fmt.Printf("AppProject state at failure: %+v\n", appproject)
				}
			}()

			Expect(k8sClient.Create(context.Background(), appproject)).To(Succeed())

			// ServiceAccount (+ Token)
			accountResource := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      solar.Name,
					Namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
				},
			}
			Expect(k8sClient.Create(context.Background(), accountResource)).To(Succeed())

			tokenResource := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      accountResource.Name,
					Namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
					Annotations: map[string]string{
						"kubernetes.io/service-account.name": accountResource.Name,
					},
				},
				Type: corev1.SecretTypeServiceAccountToken,
			}
			Expect(k8sClient.Create(context.Background(), tokenResource)).To(Succeed())
		})

		By("create tenant solar", func() {
			Eventually(func() error {
				solar.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("Verify resources were adopted (ownerreference)", func() {
			By("Verify resources were adopted (ownerreference)", func() {
				expectedResources := []struct {
					object    client.Object
					desc      string
					name      string
					namespace string
				}{
					{
						object:    &argocdv1alpha1.AppProject{},
						desc:      "AppProject",
						name:      meta.TenantProjectName(solar),
						namespace: argoaddon.Spec.Argo.Namespace,
					},
					{
						object:    &corev1.ServiceAccount{},
						desc:      "ServiceAccount",
						name:      solar.Name,
						namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
					},
				}

				for _, res := range expectedResources {
					By("Verifying " + res.desc + " contains tenant ownerreference")
					err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
					Expect(err).To(Succeed(), "%s should be present", res.desc)
					Expect(meta.HasTenantOwnerReference(res.object, solar)).To(BeTrue(), "%s should contain tenant ownerreference", res.desc)

					if err == nil {
						// Check if all expected labels match the actual labels on the resource
						labels := res.object.GetLabels()
						for key, value := range meta.TranslatorTrackingLabels(solar) {
							Expect(labels).To(HaveKeyWithValue(key, value), "%s should contain correct label %s=%s", res.desc, key, value)
						}
					}
				}
			})
		})

		By("Verify serviceaccount was added as tenant owner", func() {
			tnt := &capsulev1beta2.Tenant{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: solar.Name}, tnt)).To(Succeed())

			// Expected Owners
			owners := []capsulev1beta2.OwnerSpec{
				{
					Name: "alice",
					Kind: capsulev1beta2.GroupOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
				{
					Name: "system:serviceaccount:" + argoaddon.Spec.Argo.ServiceAccountNamespace + ":" + solar.Name,
					Kind: capsulev1beta2.ServiceAccountOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
			}

			Expect(tnt.Spec.Owners).To(Equal(capsulev1beta2.OwnerListSpec(owners)), "Tenant should have serviceaccount as owner")
		})

		By("Verify approject was adopted (finalizers)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")
		})

		By("Verify approject was adopted (translator condition)", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("Remove tenant solar", func() {
			Expect(k8sClient.Delete(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("Verify resources are no longer present (cascading deletion)", func() {
			// Define the expected resources to check for deletion
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
		})

		By("Verify tenant is no longer translated", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).To(BeNil(), "Tenant condition should not be nil")
		})
	})

	It("Test lifecycle Settings (without Force)", func() {
		By("set corresponding settings", func() {
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)).To(Succeed())
			argoaddon.Spec.Force = false

			// Attempt to update the argoaddon object in Kubernetes
			err := k8sClient.Update(context.Background(), argoaddon)
			if err != nil {
				fmt.Printf("Error updating argoaddon: %v\n", err)
			}
			Expect(err).To(Succeed(), "Failed to update argoaddon")
		})

		By("precreate solar resources", func() {
			// Approject
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels("e2e_lifecycle"),
				},
			}
			Expect(k8sClient.Create(context.Background(), appproject)).To(Succeed())

			serviceaccount := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      solar.Name,
					Namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
				},
			}
			Expect(k8sClient.Create(context.Background(), serviceaccount)).To(Succeed())
		})

		By("create tenant solar", func() {
			Eventually(func() error {
				solar.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("Verify resources were not adopted (ownerreference)", func() {
			approject := &argocdv1alpha1.AppProject{}

			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())
			Expect(meta.HasTenantOwnerReference(approject, solar)).To(BeFalse(), "Appproject should contain tenant ownerreference")
		})

		By("Verify approject was not adopted (finalizers)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeFalse(), "AppProject should contain translator finalizer")
		})

		By("Verify approject was not adopted (translator condition)", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator.Name}, tra)).To(Succeed())
			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionFalse), "Expected tenant condition status to be False")
			Expect(condition.Type).To(Equal(meta.NotReadyCondition), "Expected tenant condition type to be NotReady")
			Expect(condition.Reason).To(Equal(meta.ObjectAlreadyExistsReason), "Expected tenant condition reason to be ObjectAlreadyExists")
		})

		By("Verify serviceaccount was not added as tenant owner", func() {
			tnt := &capsulev1beta2.Tenant{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: solar.Name}, tnt)).To(Succeed())

			// Expected Owners
			owners := []capsulev1beta2.OwnerSpec{
				{
					Name: "alice",
					Kind: capsulev1beta2.GroupOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
			}

			Expect(tnt.Spec.Owners).To(Equal(capsulev1beta2.OwnerListSpec(owners)), "Tenant should not have serviceaccount as owner")
		})

	})

	It("Test lifecycle Settings (Force Annotation)", func() {
		By("set corresponding settings", func() {
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)).To(Succeed())
			argoaddon.Spec.Force = false

			// Attempt to update the argoaddon object in Kubernetes
			err := k8sClient.Update(context.Background(), argoaddon)
			if err != nil {
				fmt.Printf("Error updating argoaddon: %v\n", err)
			}
			Expect(err).To(Succeed(), "Failed to update argoaddon")
		})

		By("precreate solar resources", func() {
			// Approject
			appproject := &argocdv1alpha1.AppProject{
				ObjectMeta: metav1.ObjectMeta{
					Name:      meta.TenantProjectName(solar),
					Namespace: argoaddon.Spec.Argo.Namespace,
					Labels:    e2eLabels("e2e_lifecycle"),
				},
			}
			Expect(k8sClient.Create(context.Background(), appproject)).To(Succeed())
		})

		By("create tenant solar", func() {
			Eventually(func() error {
				solar.SetAnnotations(map[string]string{
					meta.AnnotationForce: "true",
				})

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})
		By("Verify resources were adopted (ownerreference)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())
			Expect(meta.HasTenantOwnerReference(approject, solar)).To(BeTrue(), "Appproject should contain tenant ownerreference")
		})

		By("Verify approject was adopted (finalizers)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")
		})

		By("Verify approject was adopted (translator condition)", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("Verify resources were adopted (ownerreference)", func() {
			expectedResources := []struct {
				object    client.Object
				desc      string
				name      string
				namespace string
			}{
				{
					object:    &argocdv1alpha1.AppProject{},
					desc:      "AppProject",
					name:      meta.TenantProjectName(solar),
					namespace: argoaddon.Spec.Argo.Namespace,
				},
				{
					object:    &corev1.ServiceAccount{},
					desc:      "ServiceAccount",
					name:      solar.Name,
					namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
				},
			}

			for _, res := range expectedResources {
				By("Verifying " + res.desc + " contains tenant ownerreference")
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
				Expect(err).To(Succeed(), "%s should be present", res.desc)
				Expect(meta.HasTenantOwnerReference(res.object, solar)).To(BeTrue(), "%s should contain tenant ownerreference", res.desc)

				if err == nil {
					// Check if all expected labels match the actual labels on the resource
					labels := res.object.GetLabels()
					for key, value := range meta.TranslatorTrackingLabels(solar) {
						Expect(labels).To(HaveKeyWithValue(key, value), "%s should contain correct label %s=%s", res.desc, key, value)
					}
				}
			}
		})

		By("Verify approject was adopted (finalizers)", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			Expect(meta.ContainsTranslatorFinalizer(approject)).To(BeTrue(), "AppProject should contain translator finalizer")
		})

		By("Verify approject was adopted (translator condition)", func() {
			tra := &v1alpha1.ArgoTranslator{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: translator.Name}, tra)).To(Succeed())

			condition := tra.GetTenantCondition(solar)
			Expect(condition).NotTo(BeNil(), "Tenant condition should not be nil")

			Expect(condition.Status).To(Equal(metav1.ConditionTrue), "Expected tenant condition status to be True")
			Expect(condition.Type).To(Equal(meta.ReadyCondition), "Expected tenant condition type to be Ready")
			Expect(condition.Reason).To(Equal(meta.SucceededReason), "Expected tenant condition reason to be Succeeded")
		})

		By("Verify serviceaccount was added as tenant owner", func() {
			tnt := &capsulev1beta2.Tenant{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: solar.Name}, tnt)).To(Succeed())

			// Expected Owners
			owners := []capsulev1beta2.OwnerSpec{
				{
					Name: "alice",
					Kind: capsulev1beta2.GroupOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
				{
					Name: "system:serviceaccount:" + argoaddon.Spec.Argo.ServiceAccountNamespace + ":" + solar.Name,
					Kind: capsulev1beta2.ServiceAccountOwner,
					ClusterRoles: []string{
						"admin",
						"capsule-namespace-deleter",
					},
				},
			}

			Expect(tnt.Spec.Owners).To(Equal(capsulev1beta2.OwnerListSpec(owners)), "Tenant should have serviceaccount as owner")
		})
	})

	It("Test lifecycle Settings (Decouple)", func() {
		By("set corresponding settings", func() {
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)).To(Succeed())
			argoaddon.Spec.Force = false

			// Attempt to update the argoaddon object in Kubernetes
			err := k8sClient.Update(context.Background(), argoaddon)
			if err != nil {
				fmt.Printf("Error updating argoaddon: %v\n", err)
			}
			Expect(err).To(Succeed(), "Failed to update argoaddon")
		})

		By("create tenant solar", func() {
			solar.SetAnnotations(map[string]string{
				meta.AnnotationProjectDecouple: "true",
			})
			Eventually(func() error {
				solar.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("Verify resources were adopted (ownerreference)", func() {
			expectedResources := []struct {
				object    client.Object
				desc      string
				name      string
				namespace string
			}{
				{
					object:    &argocdv1alpha1.AppProject{},
					desc:      "AppProject",
					name:      meta.TenantProjectName(solar),
					namespace: argoaddon.Spec.Argo.Namespace,
				},
				{
					object:    &corev1.ServiceAccount{},
					desc:      "ServiceAccount",
					name:      solar.Name,
					namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
				},
			}

			for _, res := range expectedResources {
				By("Verifying " + res.desc + " contains tenant ownerreference")
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
				Expect(err).To(Succeed(), "%s should be present", res.desc)
				Expect(meta.HasTenantOwnerReference(res.object, solar)).To(BeTrue(), "%s should contain tenant ownerreference", res.desc)

				if err == nil {
					// Check if all expected labels match the actual labels on the resource
					labels := res.object.GetLabels()
					for key, value := range meta.TranslatorTrackingLabels(solar) {
						Expect(labels).To(HaveKeyWithValue(key, value), "%s should contain correct label %s=%s", res.desc, key, value)
					}
				}
			}
		})

		By("Remove tenant solar", func() {
			Expect(k8sClient.Delete(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("Verify resources are still present and have no relation to tenant anymore", func() {
			expectedResources := []struct {
				object    client.Object
				desc      string
				name      string
				namespace string
			}{
				{
					object:    &argocdv1alpha1.AppProject{},
					desc:      "AppProject",
					name:      meta.TenantProjectName(solar),
					namespace: argoaddon.Spec.Argo.Namespace,
				},
				{
					object:    &corev1.ServiceAccount{},
					desc:      "ServiceAccount",
					name:      solar.Name,
					namespace: argoaddon.Spec.Argo.ServiceAccountNamespace,
				},
			}

			for _, res := range expectedResources {
				By("Verifying " + res.desc + " contains tenant ownerreference")
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
				Expect(err).To(Succeed(), "%s should be present", res.desc)

				Expect(meta.HasTenantOwnerReference(res.object, solar)).To(BeFalse(), "%s should not contain tenant ownerreference", res.desc)
				Expect(meta.ContainsTranslatorFinalizer(res.object)).To(BeFalse(), "AppProject should not contain any translator finalizer")

				if err == nil {
					// Check if all expected labels match the actual labels on the resource
					labels := res.object.GetLabels()
					for key, value := range meta.TranslatorTrackingLabels(solar) {
						if key == meta.ProvisionedByLabel {
							Expect(labels).To(HaveKeyWithValue(key, value), "%s should contain label %s=%s", res.desc, key, value)
						} else {
							Expect(labels).ToNot(HaveKeyWithValue(key, value), "%s should no longer contain label %s=%s", res.desc, key, value)

						}
					}
				}
			}
		})
	})
})
