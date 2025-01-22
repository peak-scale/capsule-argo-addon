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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
)

var _ = Describe("Argo Destination Test", func() {
	suiteSelector := e2eLabels("e2e_destination")

	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}

	solar := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "solar-e2e-dest",
			Labels:      suiteSelector,
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

	translator := &v1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "e2e-dest",
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
		},
	}

	BeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()
	})

	AfterEach(func() {
		Expect(CleanTenants(e2eSelector("e2e_destination"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_destination"))).ToNot(HaveOccurred())

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

	It("Verify DestinationServiceAccounts", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.Argo.Destination = "https://custom.server:443"
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator)).ToNot(HaveOccurred())
		})

		By("create matching tenant", func() {
			Expect(k8sClient.Create(context.TODO(), solar)).ToNot(HaveOccurred())
		})

		By("verify resources are present", func() {
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
			}
		})

		By("verify resources are absent", func() {
			expectedResources := []struct {
				object    client.Object
				desc      string
				name      string
				namespace string
			}{
				{
					object:    &corev1.Secret{},
					desc:      "Cluster Secret",
					name:      solar.Name,
					namespace: argoaddon.Spec.Argo.Namespace,
				},
			}

			for _, res := range expectedResources {
				err := k8sClient.Get(context.Background(), client.ObjectKey{Name: res.name, Namespace: res.namespace}, res.object)
				Expect(err).ToNot(Succeed(), "%s should not be present", res.desc)
			}
		})

		By("verify project destination", func() {
			approject := &argocdv1alpha1.AppProject{}
			Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: meta.TenantProjectName(solar), Namespace: argoaddon.Spec.Argo.Namespace}, approject)).To(Succeed())

			// Expected Translation
			expected := []argocdv1alpha1.ApplicationDestinationServiceAccount{
				{DefaultServiceAccount: argoaddon.Spec.DestinationServiceAccount(solar), Namespace: "*", Server: argoaddon.Spec.Argo.Destination},
			}

			// Compare the Spec
			Expect(approject.Spec.DestinationServiceAccounts).To(Equal(expected), "AppProject destinations should match the expected spec")

			// Compare the Spec
			Expect(len(approject.Spec.Destinations)).To(Equal(0), "AppProject destinations should match the expected spec")
		})

	})

	It("Does Registry cluster (Annotation)", func() {
		By("create translation", func() {
			Expect(k8sClient.Create(context.TODO(), translator)).ToNot(HaveOccurred())
		})

		By("create matching tenant", func() {
			Eventually(func() error {
				solar.SetAnnotations(map[string]string{
					meta.AnnotationDestinationRegister: "true",
				})

				return k8sClient.Create(context.TODO(), solar)
			}).Should(Succeed())
		})

		By("verify resources are present", func() {
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
					object:    &corev1.Secret{},
					desc:      "Cluster Secret",
					name:      solar.Name,
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
			}
		})
	})
})
