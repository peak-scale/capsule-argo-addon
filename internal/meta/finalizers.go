// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package meta

import (
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// FinalizerName is the finalizer name for the translator
	TranslatorFinalizerPrefix = "translator.addons.projectcapsule.dev/"
)

func TranslatorFinalizer(name string) string {
	return TranslatorFinalizerPrefix + name
}

// Get all translators based on their finalizer
func GetTranslatingFinalizers(obj client.Object) (translators []string) {
	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, TranslatorFinalizerPrefix) {
			translators = append(translators, strings.TrimPrefix(finalizer, TranslatorFinalizerPrefix))
		}
	}

	return
}

// Get all translators based on their finalizer
func RemoveTranslatingFinalizers(obj client.Object) {
	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, TranslatorFinalizerPrefix) {
			controllerutil.RemoveFinalizer(obj, finalizer)
		}
	}
}

// Contains Translator Finalizers
func ContainsTranslatorFinalizer(obj client.Object) (contains bool) {
	contains = false

	// Iterate over the finalizers and check if any contain the specified prefix
	for _, finalizer := range obj.GetFinalizers() {
		if strings.HasPrefix(finalizer, TranslatorFinalizerPrefix) {
			return true
		}
	}

	return
}
