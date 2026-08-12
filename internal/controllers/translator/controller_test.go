package translator

import (
	"context"
	"testing"

	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestPruneMissingTenantStatuses_RemovesDeletedTenants(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add config scheme: %v", err)
	}
	if err := capsulev1beta2.AddToScheme(scheme); err != nil {
		t.Fatalf("add capsule scheme: %v", err)
	}

	translator := &configv1alpha1.ArgoTranslator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "project-translator",
		},
		Status: configv1alpha1.ArgoTranslatorStatus{
			Tenants: []configv1alpha1.TenantStatus{
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
			},
			Size:  2,
			Ready: meta.NotReadyCondition,
		},
	}
	liveTenant := &capsulev1beta2.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: "tenant-a",
			UID:  k8stypes.UID("uid-a"),
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(translator, liveTenant).
		WithStatusSubresource(&configv1alpha1.ArgoTranslator{}).
		Build()
	reconciler := &Reconciler{Client: fakeClient}

	pruned, err := reconciler.pruneMissingTenantStatuses(
		context.Background(),
		client.ObjectKey{Name: "project-translator"},
	)
	if err != nil {
		t.Fatalf("prune missing tenant statuses: %v", err)
	}
	if !pruned {
		t.Fatal("expected stale tenant status to be pruned")
	}

	got := &configv1alpha1.ArgoTranslator{}
	if err := fakeClient.Get(context.Background(), client.ObjectKey{Name: "project-translator"}, got); err != nil {
		t.Fatalf("get translator: %v", err)
	}

	if gotNames := got.GetTenantNames(); len(gotNames) != 1 || gotNames[0] != "tenant-a" {
		t.Fatalf("expected only tenant-a status to remain, got %v", gotNames)
	}
	if got.Status.Size != 1 {
		t.Fatalf("expected status size 1, got %d", got.Status.Size)
	}
	if got.Status.Ready != meta.ReadyCondition {
		t.Fatalf("expected ready status %q, got %q", meta.ReadyCondition, got.Status.Ready)
	}
}
