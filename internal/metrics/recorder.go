package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

type Recorder struct {
	conditionGauge *prometheus.GaugeVec
}

func MustMakeRecorder() *Recorder {
	metricsRecorder := NewRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)
	return metricsRecorder
}

func NewRecorder() *Recorder {
	return &Recorder{
		conditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cca_translator_condition",
				Help: "The current condition status of a Translator.",
			},
			[]string{"kind", "name", "type", "status"},
		),
	}
}

func (r *Recorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.conditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordCondition(ref corev1.ObjectReference, condition metav1.Condition) {
	for _, status := range []metav1.ConditionStatus{metav1.ConditionTrue, metav1.ConditionFalse, metav1.ConditionUnknown} {
		var value float64
		if status == condition.Status {
			value = 1
		}
		r.conditionGauge.WithLabelValues(ref.Kind, ref.Name, condition.Type, string(status)).Set(value)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteCondition(ref corev1.ObjectReference, conditionType string) {
	for _, status := range []metav1.ConditionStatus{metav1.ConditionTrue, metav1.ConditionFalse, metav1.ConditionUnknown} {
		r.conditionGauge.DeleteLabelValues(ref.Kind, ref.Name, ref.Namespace, conditionType, string(status))
	}
}
