//nolint:all
package e2e_test

import (
	"context"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"

	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg                          *rest.Config
	k8sClient                    client.Client
	testEnv                      *envtest.Environment
	originalArgoAddon, argoAddon *v1alpha1.ArgoAddon
)

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		UseExistingCluster: ptr.To(true),
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	Expect(argocdv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(capsulev1beta2.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())
	Expect(configv1alpha1.AddToScheme(scheme.Scheme)).NotTo(HaveOccurred())

	ctrlClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(ctrlClient).ToNot(BeNil())

	k8sClient = &e2eClient{Client: ctrlClient}

	selector := e2eSelector("")
	Expect(CleanTenants(selector)).ToNot(HaveOccurred())
	Expect(CleanTranslators(selector)).ToNot(HaveOccurred())

	// Initialize originalArgoAddon before using it.
	originalArgoAddon = &v1alpha1.ArgoAddon{}
	err = k8sClient.Get(context.Background(), client.ObjectKey{Name: "default"}, originalArgoAddon)
	Expect(err).ToNot(HaveOccurred())

	argoAddon = originalArgoAddon.DeepCopy()

	// Now that argoAddon is non-nil, you can safely update its spec.
	argoAddon.Spec.Decouple = false
	Expect(k8sClient.Update(context.Background(), argoAddon)).To(Succeed())

	Eventually(CleanAppProjects(selector, "argocd"))
})

var _ = AfterSuite(func() {
	// Apply the initial configuration from originalArgoAddon to argoaddon
	Expect(k8sClient.Get(context.Background(), client.ObjectKey{Name: originalArgoAddon.Name}, argoAddon)).To(Succeed())
	argoAddon.Spec = originalArgoAddon.Spec
	Expect(k8sClient.Update(context.Background(), argoAddon)).To(Succeed())

	By("tearing down the test environment")
	Expect(testEnv.Stop()).ToNot(HaveOccurred())
})
