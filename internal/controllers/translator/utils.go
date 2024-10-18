package translator

import (
	"github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
)

const (
	// FinalizerName is the finalizer name for the translator
	FinalizerPrefix = "translator.addons.projectcapsule.dev/"
)

func TranslatorFinalizer(translator *v1alpha1.ArgoTranslator) string {
	return FinalizerPrefix + translator.Name
}
