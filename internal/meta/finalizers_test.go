package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Mock object implementing client.Object interface
type mockObject struct {
	metav1.ObjectMeta
}

func (m *mockObject) GetObjectKind() schema.ObjectKind {
	return schema.EmptyObjectKind
}

func (m *mockObject) DeepCopyObject() runtime.Object {
	return &mockObject{
		ObjectMeta: *m.ObjectMeta.DeepCopy(),
	}
}

// Test for GetTranslatingFinalizers function
func TestGetTranslatingFinalizers(t *testing.T) {
	// Create a mock object with finalizers
	obj := &mockObject{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{
				"something-else",
				TranslatorFinalizer("solar"),
				TranslatorFinalizer("oil"),
				"some-other-finalizer",
			},
		},
	}

	// Call the function
	translators := GetTranslatingFinalizers(obj)

	// Verify the result
	expectedTranslators := []string{"solar", "oil"}
	assert.Equal(t, expectedTranslators, translators, "Expected translators to match")
}

func TestContainsTranslatorFinalizer(t *testing.T) {
	// Create a mock object with finalizers
	obj := &mockObject{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{
				"something-else",
				TranslatorFinalizer("solar"),
				"some-other-finalizer",
			},
		},
	}

	// Call the function
	translators := ContainsTranslatorFinalizer(obj)
	assert.Equal(t, true, translators, "Expected translators to match")
}

func TestContainsTranslatorFinalizerNegative(t *testing.T) {
	// Create a mock object with finalizers
	obj := &mockObject{
		ObjectMeta: metav1.ObjectMeta{
			Finalizers: []string{
				"something-else",
				"some-other-finalizer",
			},
		},
	}

	// Call the function
	translators := ContainsTranslatorFinalizer(obj)
	assert.Equal(t, false, translators, "Expected translators to not match")
}
