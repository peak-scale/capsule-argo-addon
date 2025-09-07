//nolint:all
package e2e_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	capsuleapi "github.com/projectcapsule/capsule/pkg/api"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
)

var _ = Describe("Argo Repository Test", Label("repository"), func() {
	suiteSelector := e2eLabels("e2e_repository")

	// Resources
	argoaddon := &v1alpha1.ArgoAddon{}
	originalArgoAddon := &v1alpha1.ArgoAddon{}

	tnt := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "solar-e2e-repo",
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

	BeforeEach(func() {
		// Save the current state of the argoaddon configuration
		Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, originalArgoAddon)).To(Succeed())
		argoaddon = originalArgoAddon.DeepCopy()
	})

	AfterEach(func() {
		Expect(CleanTenants(e2eSelector("e2e_repository"))).ToNot(HaveOccurred())
		Expect(CleanTranslators(e2eSelector("e2e_repository"))).ToNot(HaveOccurred())

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

	It("Verify Repository Synchronization", func() {
		By("set corresponding settings", func() {
			_ = k8sClient.Get(context.Background(), client.ObjectKey{Name: e2eConfigName()}, argoaddon)
			argoaddon.Spec.AllowRepositoryCreation = true
			Expect(k8sClient.Update(context.Background(), argoaddon)).To(Succeed())
		})

		By("create matching tenant", func() {
			Expect(k8sClient.Create(context.TODO(), tnt)).ToNot(HaveOccurred())
		})

		ns := NewNamespace("")
		compareSecret := &corev1.Secret{}

		By("create Secret in Namespace", func() {
			NamespaceCreation(ns, tnt.Spec.Owners[0], defaultTimeoutInterval).Should(Succeed())

			time.Sleep(2000 * time.Millisecond)
		})

		By("creating an Argo repository Secret in the source namespace", func() {
			compareSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-repo",
					Namespace: ns.Name,
					Labels: map[string]string{
						"argocd.argoproj.io/secret-type": "repository",
					},
				},
				StringData: map[string]string{
					"url":      "https://github.com/org/repo.git",
					"username": "user",
					"password": "pass",
					"project":  "default",
				},
				Type: corev1.SecretTypeOpaque,
			}
			Expect(k8sClient.Create(context.TODO(), compareSecret)).To(Succeed())

			Eventually(func() error {
				var got corev1.Secret
				return k8sClient.Get(context.TODO(), client.ObjectKeyFromObject(compareSecret), &got)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})

		By("verifying the Secret was replicated into the target namespace", func() {
			replicaKey := types.NamespacedName{
				Namespace: argoaddon.Spec.Argo.Namespace,
				Name:      compareSecret.GetNamespace() + "-" + compareSecret.GetName(),
			}

			compare := compareSecret.Data
			compare["project"] = []byte(tnt.Name)

			Eventually(func() map[string][]byte {
				var replica corev1.Secret
				if err := k8sClient.Get(context.TODO(), replicaKey, &replica); err != nil {
					return nil
				}
				return replica.Data
			}, defaultTimeoutInterval, defaultPollInterval).Should(Equal(compareSecret.Data))
		})

		By("deleting the replica Secret and verifying it is recreated", func() {
			replicaKey := types.NamespacedName{
				Namespace: argoaddon.Spec.Argo.Namespace,
				Name:      compareSecret.GetNamespace() + "-" + compareSecret.GetName(),
			}
			Expect(k8sClient.Delete(context.TODO(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: replicaKey.Namespace,
					Name:      replicaKey.Name,
				},
			})).To(Succeed())

			time.Sleep(2000 * time.Millisecond)

			Eventually(func() error {
				var replica corev1.Secret
				return k8sClient.Get(context.TODO(), replicaKey, &replica)
			}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
		})

		By("deleting the source Secret and verifying the replica is also deleted", func() {
			Expect(k8sClient.Delete(context.TODO(), &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      compareSecret.Name,
					Namespace: compareSecret.Namespace,
				},
			})).To(Succeed())

			replicaKey := types.NamespacedName{
				Namespace: argoaddon.Spec.Argo.Namespace,
				Name:      compareSecret.GetNamespace() + "-" + compareSecret.GetName(),
			}
			Eventually(func() bool {
				var replica corev1.Secret
				err := k8sClient.Get(context.TODO(), replicaKey, &replica)
				return k8serrors.IsNotFound(err)
			}, defaultTimeoutInterval, defaultPollInterval).Should(BeTrue())
		})

	})
})
