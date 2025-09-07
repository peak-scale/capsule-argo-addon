//nolint:all
package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

func ownerClient(owner capsulev1beta2.OwnerSpec) (cs kubernetes.Interface) {
	c, err := config.GetConfig()
	Expect(err).ToNot(HaveOccurred())
	c.Impersonate.Groups = []string{"projectcapsule.dev", owner.Name}
	c.Impersonate.UserName = owner.Name
	cs, err = kubernetes.NewForConfig(c)
	Expect(err).ToNot(HaveOccurred())

	return cs
}

func impersonationClient(user string, groups []string, scheme *runtime.Scheme) client.Client {
	c, err := config.GetConfig()
	Expect(err).ToNot(HaveOccurred())
	c.Impersonate = rest.ImpersonationConfig{
		UserName: user,
		Groups:   groups,
	}
	cl, err := client.New(c, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	return cl
}

func withDefaultGroups(groups []string) []string {
	return append([]string{"projectcapsule.dev"}, groups...)
}
