//nolint:all
package e2e_test

import (
	"context"
	"fmt"
	"reflect"
	"time"

	argocdv1alpha1 "github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	. "github.com/onsi/gomega"
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultTimeoutInterval = 20 * time.Second
	defaultPollInterval    = time.Second
	e2eLabel               = "argo.addons.projectcapsule.dev/e2e"
	suiteLabel             = "e2e.argo.addons.projectcapsule.dev/suite"
)

func e2eConfigName() string {
	return "default"
}

// Returns labels to identify e2e resources.
func e2eLabels(suite string) (labels map[string]string) {
	labels = make(map[string]string)
	labels["argo.addons.projectcapsule.dev/e2e"] = "true"

	if suite != "" {
		labels["e2e.argo.addons.projectcapsule.dev/suite"] = suite
	}

	return
}

// Returns a label selector to filter e2e resources.
func e2eSelector(suite string) labels.Selector {
	return labels.SelectorFromSet(e2eLabels(suite))
}

// Pass objects which require cleanup and a label selector to filter them.
func cleanResources(res []client.Object, selector labels.Selector) (err error) {
	for _, resource := range res {
		err = k8sClient.DeleteAllOf(context.TODO(), resource, &client.MatchingLabels{"argo.addons.projectcapsule.dev/e2e": "true"})

		if err != nil {
			return err
		}
	}

	return nil
}

func NewNamespace(name string, labels ...map[string]string) *corev1.Namespace {
	if len(name) == 0 {
		name = rand.String(10)
	}

	var namespaceLabels map[string]string
	if len(labels) > 0 {
		namespaceLabels = labels[0]
	}

	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: namespaceLabels,
		},
	}
}

func NamespaceCreation(ns *corev1.Namespace, owner capsulev1beta2.OwnerSpec, timeout time.Duration) AsyncAssertion {
	cs := ownerClient(owner)
	return Eventually(func() (err error) {
		_, err = cs.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
		return
	}, timeout, defaultPollInterval)
}

func TenantNamespaceList(t *capsulev1beta2.Tenant, timeout time.Duration) AsyncAssertion {
	return Eventually(func() []string {
		Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Name: t.GetName()}, t)).Should(Succeed())
		return t.Status.Namespaces
	}, timeout, defaultPollInterval)
}

func CleanTranslators(selector labels.Selector) error {
	res := &v1alpha1.ArgoTranslatorList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list translators: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete translator %s: %w", app.GetName(), err)
		}
	}

	return nil
}

func CleanTenants(selector labels.Selector) error {
	res := &capsulev1beta2.TenantList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete tenant %s: %w", app.GetName(), err)
		}
	}

	return nil
}

func CleanAppProjects(selector labels.Selector, namespace string) error {
	res := &argocdv1alpha1.AppProjectList{}

	listOptions := client.ListOptions{
		LabelSelector: selector,
	}

	// If a namespace is provided, set it in the list options
	if namespace != "" {
		listOptions.Namespace = namespace
	}

	// List the resources based on the provided label selector
	if err := k8sClient.List(context.TODO(), res, &listOptions); err != nil {
		return fmt.Errorf("failed to list resources: %w", err)
	}

	for _, app := range res.Items {
		if err := k8sClient.Delete(context.TODO(), &app); err != nil {
			return fmt.Errorf("failed to delete resource %s: %w", app.GetName(), err)
		}
	}

	return nil
}

func DeepCompare(expected, actual interface{}) (bool, string) {
	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	// If the kinds differ, they are not equal.
	if expVal.Kind() != actVal.Kind() {
		return false, fmt.Sprintf("kind mismatch: %v vs %v", expVal.Kind(), actVal.Kind())
	}

	switch expVal.Kind() {
	case reflect.Slice, reflect.Array:
		// Convert slices to []interface{} for ElementsMatch.
		expSlice := make([]interface{}, expVal.Len())
		actSlice := make([]interface{}, actVal.Len())
		for i := 0; i < expVal.Len(); i++ {
			expSlice[i] = expVal.Index(i).Interface()
		}
		for i := 0; i < actVal.Len(); i++ {
			actSlice[i] = actVal.Index(i).Interface()
		}
		// Use a dummy tester to capture error messages.
		dummy := &dummyT{}
		if !assert.ElementsMatch(dummy, expSlice, actSlice) {
			return false, fmt.Sprintf("slice mismatch: %v", dummy.errors)
		}
		return true, ""
	case reflect.Struct:
		// Iterate over fields and compare recursively.
		for i := 0; i < expVal.NumField(); i++ {
			fieldName := expVal.Type().Field(i).Name
			ok, msg := DeepCompare(expVal.Field(i).Interface(), actVal.Field(i).Interface())
			if !ok {
				return false, fmt.Sprintf("field %s mismatch: %s", fieldName, msg)
			}
		}
		return true, ""
	default:
		// Fallback to reflect.DeepEqual for other types.
		if !reflect.DeepEqual(expected, actual) {
			return false, fmt.Sprintf("expected %v but got %v", expected, actual)
		}
		return true, ""
	}
}

// dummyT implements a minimal TestingT for testify.
type dummyT struct {
	errors []string
}

func (d *dummyT) Errorf(format string, args ...interface{}) {
	d.errors = append(d.errors, fmt.Sprintf(format, args...))
}
