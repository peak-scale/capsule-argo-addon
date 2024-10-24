package e2e_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

var _ = Describe("Capsule-Proxy", func() {
	config

	BeforeEach(func() {
		// Tenants
		solar := tntSolar.DeepCopy()
		oil := tntOil.DeepCopy()
		tenants := []*capsulev1beta2.Tenant{solar, oil}

		translators := []*v1alpha1.ArgoTranslator{baseTranslator}

		for _, tnt := range tenants {
			Eventually(func() error {
				tnt.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), tnt)
			}).Should(Succeed())
		}

		for _, tran := range translators {
			Eventually(func() error {
				tran.ResourceVersion = ""

				return k8sClient.Create(context.TODO(), tran)
			}).Should(Succeed())
		}

	})

	JustAfterEach(func() {
		// Define Resources which are lifecycled after each test
		resourcesToClean := []client.Object{
			&capsulev1beta2.Tenant{},
		}

		Eventually(func() error {
			return cleanResources(resourcesToClean, e2eSelector())
		}, defaultTimeoutInterval, defaultPollInterval).Should(Succeed())
	})

	It("Correctly Bootstrap Proxy-Service according to configuration", func() {

	})
})
