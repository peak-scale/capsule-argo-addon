# Monitoring

Via the `/metrics` endpoint and the dedicated port you can scrape Prometheus Metrics. Amongst the standard [Kubebuilder Metrics](https://book-v1.book.kubebuilder.io/beyond_basics/controller_metrics) we provide metrics, which show the current state of translators or tenants. This way you can always be informed, when something is not working as expected. Our custom metrics are prefixed with `caa_`:

```shell
# HELP cca_translator_condition The current condition status of a Translator.
# TYPE cca_translator_condition gauge
cca_translator_condition{name="default-onboarding",status="NotReady"} 0
cca_translator_condition{name="default-onboarding",status="Ready"} 1
cca_translator_condition{name="dev-onboarding",status="NotReady"} 0
cca_translator_condition{name="dev-onboarding",status="Ready"} 1
```

The Helm-Chart comes with a [ServiceMonitor](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#servicemonitor) and [PrometheusRules](https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#monitoring.coreos.com/v1.PrometheusRule)
