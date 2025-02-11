
{{- define "helm.webhooks.service" -}}
  {{- include "helm.webhooks.cabundle" $.ctx | nindent 0 }}
  {{- if $.ctx.Values.webhooks.service.url }}
url: {{ printf "%s/%s" (trimSuffix "/" $.ctx.Values.webhooks.service.url ) (trimPrefix "/" (required "Path is required for the function" $.path)) }}
  {{- else }}
service:
  name: {{ default (printf "%s-webhook-service" (include "helm.fullname" $.ctx)) $.ctx.Values.webhooks.service.name }}
  namespace: {{ default $.ctx.Release.Namespace $.ctx.Values.webhooks.service.namespace }}
  port: {{ default 443 $.ctx.Values.webhooks.service.port }}
  path: {{ required "Path is required for the function" $.path }}
  {{- end }}
{{- end }}

{{/*
Capsule Webhook endpoint CA Bundle
*/}}
{{- define "helm.webhooks.cabundle" -}}
  {{- if $.Values.webhooks.service.caBundle -}}
caBundle: {{ $.Values.webhooks.service.caBundle -}}
  {{- end -}}
{{- end -}}
