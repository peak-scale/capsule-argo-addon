package tenant

import (
	"testing"

	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestTranslatorRequeuePredicate_IgnoresStatusOnlyUpdates(t *testing.T) {
	oldTranslator := &configv1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "project-translator",
			Generation: 1,
		},
		Status: configv1alpha1.ArgoTranslatorStatus{
			Ready: "Ready",
		},
	}
	newTranslator := oldTranslator.DeepCopy()
	newTranslator.Status.Size = 1

	if translatorRequeuePredicate().Update(event.UpdateEvent{
		ObjectOld: oldTranslator,
		ObjectNew: newTranslator,
	}) {
		t.Fatal("expected status-only update to be ignored")
	}
}

func TestTranslatorRequeuePredicate_AllowsSpecUpdates(t *testing.T) {
	oldTranslator := &configv1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "project-translator",
			Generation: 1,
		},
	}
	newTranslator := oldTranslator.DeepCopy()
	newTranslator.Generation = 2

	if !translatorRequeuePredicate().Update(event.UpdateEvent{
		ObjectOld: oldTranslator,
		ObjectNew: newTranslator,
	}) {
		t.Fatal("expected generation change to requeue tenants")
	}
}
