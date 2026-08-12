package v1alpha1

import (
	"reflect"
	"testing"
	"time"

	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
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

func TestUpdateTenantCondition_PreservesTransitionTimeWhenConditionIsUnchanged(t *testing.T) {
	transitionTime := metav1.NewTime(time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC))

	translator := &ArgoTranslator{}
	translator.UpdateTenantCondition(TenantStatus{
		Name: "tenant-a",
		Condition: metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 1,
			Reason:             "Applied",
			Message:            "Successfully translated tenant",
			LastTransitionTime: transitionTime,
		},
	})

	translator.UpdateTenantCondition(TenantStatus{
		Name: "tenant-a",
		Condition: metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: 1,
			Reason:             "Applied",
			Message:            "Successfully translated tenant",
			LastTransitionTime: metav1.Now(),
		},
	})

	got := translator.Status.Tenants[0].Condition.LastTransitionTime
	if !got.Equal(&transitionTime) {
		t.Fatalf("expected last transition time %s, got %s", transitionTime, got)
	}
}

func TestSyncFinalizerStatus_RemovesFinalizerWhenTranslatorIsDeleting(t *testing.T) {
	now := metav1.Now()
	translator := &ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &now,
			Finalizers:        []string{meta.ControllerFinalizer},
		},
		Status: ArgoTranslatorStatus{
			Tenants: []TenantStatus{{Name: "tenant-a"}},
		},
	}

	translator.SyncFinalizerStatus()

	for _, finalizer := range translator.Finalizers {
		if finalizer == meta.ControllerFinalizer {
			t.Fatal("expected controller finalizer to be removed while translator is deleting")
		}
	}
}

func TestPruneMissingTenantStatuses_RemovesStaleTenants(t *testing.T) {
	translator := &ArgoTranslator{
		Status: ArgoTranslatorStatus{
			Tenants: []TenantStatus{
				{
					Name: "tenant-a",
					UID:  k8stypes.UID("uid-a"),
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
				{
					Name: "tenant-b",
					UID:  k8stypes.UID("uid-b"),
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionFalse,
					},
				},
				{
					Name: "tenant-c",
					UID:  k8stypes.UID("old-uid-c"),
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
				{
					Name: "tenant-d",
					Condition: metav1.Condition{
						Type:   "Ready",
						Status: metav1.ConditionTrue,
					},
				},
			},
		},
	}
	translator.CollectStatus()

	translator.PruneMissingTenantStatuses([]capsulev1beta2.Tenant{
		{ObjectMeta: metav1.ObjectMeta{Name: "tenant-a", UID: k8stypes.UID("uid-a")}},
		{ObjectMeta: metav1.ObjectMeta{Name: "tenant-c", UID: k8stypes.UID("new-uid-c")}},
		{ObjectMeta: metav1.ObjectMeta{Name: "tenant-d", UID: k8stypes.UID("uid-d")}},
	})

	want := []string{"tenant-a", "tenant-d"}
	if got := translator.GetTenantNames(); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected tenants %v, got %v", want, got)
	}

	if translator.Status.Size != 2 {
		t.Fatalf("expected status size 2, got %d", translator.Status.Size)
	}

	if translator.Status.Ready != meta.ReadyCondition {
		t.Fatalf("expected ready status %q, got %q", meta.ReadyCondition, translator.Status.Ready)
	}

}
