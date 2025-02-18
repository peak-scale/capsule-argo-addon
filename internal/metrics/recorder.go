// Copyright 2024 Peak Scale
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	configv1alpha1 "github.com/peak-scale/capsule-argo-addon/api/v1alpha1"
	"github.com/peak-scale/capsule-argo-addon/internal/meta"
	"github.com/prometheus/client_golang/prometheus"
	crtlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"

	capsulev1beta2 "github.com/projectcapsule/capsule/api/v1beta2"
)

type Recorder struct {
	translatorConditionGauge *prometheus.GaugeVec
	tenantConditionGauge     *prometheus.GaugeVec
}

func MustMakeRecorder() *Recorder {
	metricsRecorder := NewRecorder()
	crtlmetrics.Registry.MustRegister(metricsRecorder.Collectors()...)

	return metricsRecorder
}

func NewRecorder() *Recorder {
	return &Recorder{
		translatorConditionGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cca_translator_condition",
				Help: "The current condition status of a Translator.",
			},
			[]string{"name", "status"},
		),

		tenantConditionGauge: prometheus.NewGaugeVec( // Initialize tenantConditionGauge here
			prometheus.GaugeOpts{
				Name: "cca_tenant_condition",
				Help: "The current condition status of a Tenant.",
			},
			[]string{"name", "status"},
		),
	}
}

func (r *Recorder) Collectors() []prometheus.Collector {
	return []prometheus.Collector{
		r.translatorConditionGauge,
		r.tenantConditionGauge,
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordTranslatorCondition(translator *configv1alpha1.ArgoTranslator) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if status == translator.Status.Ready {
			value = 1
		}

		r.translatorConditionGauge.WithLabelValues(translator.Name, status).Set(value)
	}
}

// RecordCondition records the condition as given for the ref.
func (r *Recorder) RecordTenantCondition(tenant *capsulev1beta2.Tenant, condition string) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		var value float64
		if status == condition {
			value = 1
		}

		r.tenantConditionGauge.WithLabelValues(tenant.Name, status).Set(value)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteTenantCondition(tenantName string) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.tenantConditionGauge.DeleteLabelValues(tenantName, status)
	}
}

// DeleteCondition deletes the condition metrics for the ref.
func (r *Recorder) DeleteTranslatorCondition(translatorName string) {
	for _, status := range []string{meta.ReadyCondition, meta.NotReadyCondition} {
		r.translatorConditionGauge.DeleteLabelValues(translatorName, status)
	}
}
