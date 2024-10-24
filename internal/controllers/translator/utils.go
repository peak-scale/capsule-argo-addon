package translator

import (
	"strings"

	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// FinalizerName is the finalizer name for the translator
	FinalizerPrefix = "translator.addons.projectcapsule.dev/"
)

func TranslatorFinalizer(translator *v1alpha1.ArgoTranslator) string {
	return FinalizerPrefix + translator.Name
}

// Get all translators based on their finalizer
func GetTranslatingFinalizers(obj client.Object) (translators []string) {
	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, FinalizerPrefix) {
			translators = append(translators, strings.TrimPrefix(finalizer, FinalizerPrefix))
		}
	}

	return
}

// Get all translators based on their finalizer
func RemoveTranslatingFinalizers(obj client.Object) (translators []string) {
	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, FinalizerPrefix) {
			controllerutil.RemoveFinalizer(obj, finalizer)
		}
	}

	return
}

// Contains Translator Finalizers
func ContainsTranslatorFinalizer(obj client.Object) (contains bool) {
	contains = false

	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, FinalizerPrefix) {
			return true
		}
	}

	return
}
