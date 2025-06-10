package v1alpha1

import (
	"testing"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReconcileTranslator_WithEmptySelector_ShouldMatchAllTenants(t *testing.T) {
	translator := &ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-translator",
			Namespace: "capsule-system",
		},
		Spec: ArgoTranslatorSpec{
			Selector: nil,
		},
	}

	tenants := []capsulev1beta2.Tenant{
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-a",
			Labels: map[string]string{"team": "a"},
		}},
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-b",
			Labels: map[string]string{"team": "b"},
		}},
	}

	matchedTenants := []capsulev1beta2.Tenant{}
	for _, tenant := range tenants {
		if translator.MatchesObject(&tenant) {
			matchedTenants = append(matchedTenants, tenant)
		}
	}

	if len(matchedTenants) != 2 {
		t.Errorf("expected 2 tenants to match, got %d", len(matchedTenants))
	}
}

func TestReconcileTranslator_WithEmptySelector_ShouldMatchNoTenant(t *testing.T) {
	translator := &ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-translator",
			Namespace: "capsule-system",
		},
		Spec: ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"no-tenant-has-this": "true",
				},
			},
		},
	}

	tenants := []capsulev1beta2.Tenant{
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-a",
			Labels: map[string]string{"team": "a"},
		}},
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-b",
			Labels: map[string]string{"team": "b"},
		}},
	}

	matchedTenants := []capsulev1beta2.Tenant{}
	for _, tenant := range tenants {
		if translator.MatchesObject(&tenant) {
			matchedTenants = append(matchedTenants, tenant)
		}
	}

	if len(matchedTenants) != 0 {
		t.Errorf("expected 0 tenants to match, got %d", len(matchedTenants))
	}
}

func TestReconcileTranslator_WithEmptySelector_ShouldMatchOneTenant(t *testing.T) {
	translator := &ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-translator",
			Namespace: "capsule-system",
		},
		Spec: ArgoTranslatorSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"team": "a",
				},
			},
		},
	}

	tenants := []capsulev1beta2.Tenant{
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-a",
			Labels: map[string]string{"team": "a"},
		}},
		{ObjectMeta: metav1.ObjectMeta{
			Name:   "tenant-b",
			Labels: map[string]string{"team": "b"},
		}},
	}

	matchedTenants := []capsulev1beta2.Tenant{}
	for _, tenant := range tenants {
		if translator.MatchesObject(&tenant) {
			matchedTenants = append(matchedTenants, tenant)
		}
	}

	if len(matchedTenants) != 1 {
		t.Errorf("expected 1 tenants to match, got %d", len(matchedTenants))
	}
}
