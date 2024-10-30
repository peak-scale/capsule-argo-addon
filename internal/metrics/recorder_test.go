package metrics

import (
	"testing"

	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRecorder_RecordCondition(t *testing.T) {
	rec := NewRecorder()
	reg := prometheus.NewRegistry()
	reg.MustRegister(rec.conditionGauge)

	ref := corev1.ObjectReference{
		Kind: "ArgoTranslator",
		Name: "test",
	}

	cond := metav1.Condition{
		Type:   meta.ReadyCondition,
		Status: metav1.ConditionTrue,
	}

	rec.RecordCondition(ref, cond)

	metricFamilies, err := reg.Gather()
	require.NoError(t, err)

	require.Equal(t, len(metricFamilies), 1)
	require.Equal(t, len(metricFamilies[0].Metric), 3)

	var conditionTrueValue float64
	for _, m := range metricFamilies[0].Metric {
		for _, pair := range m.GetLabel() {
			if *pair.Name == "type" && *pair.Value != meta.ReadyCondition {
				t.Errorf("expected condition type to be %s, got %s", meta.ReadyCondition, *pair.Value)
			}
			if *pair.Name == "status" && *pair.Value == string(metav1.ConditionTrue) {
				conditionTrueValue = *m.GetGauge().Value
			} else if *pair.Name == "status" && *m.GetGauge().Value != 0 {
				t.Errorf("expected guage value to be 0, got %v", *m.GetGauge().Value)
			}
		}
	}

	require.Equal(t, conditionTrueValue, float64(1))

	// Delete metrics.
	rec.DeleteCondition(ref, cond.Type)

	metricFamilies, err = reg.Gather()
	require.NoError(t, err)
	require.Equal(t, len(metricFamilies), 0)
}
