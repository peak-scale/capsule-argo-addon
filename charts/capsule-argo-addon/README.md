# Capsule Argo Addon

![Capsule Argo Addon Logo](https://github.com/peak-scale/capsule-argo-addon/blob/main/docs/images/capsule-argo.png)

This addon is designed for kubernetes administrators, to automatically translate their existing Capsule Tenants into Argo Appprojects. This addon adds new capabilities to the Capsule project, by allowing the administrator to create a new tenant in Capsule, and automatically create a new Argo Appproject for that tenant. This addon is designed to be used in conjunction with the Capsule project, and is not intended to be used as a standalone project. [Read More about the Installation](https://github.com/peak-scale/capsule-argo-addon/blob/main/docs/installation.md)

## Installation

1. Install Helm Chart:

        $ helm install capsule-argo-addon oci://ghcr.io/peak-scale/charts/capsule-argo-addon  -n capsule-system

3. Show the status:

        $ helm status capsule-argo-addon -n capsule-system

4. Upgrade the Chart

        $ helm upgrade capsule-argo-addon oci://ghcr.io/peak-scale/charts/capsule-argo-addon --version 0.4.7

5. Uninstall the Chart

        $ helm uninstall capsule-argo-addon -n capsule-system

## Values

The following Values are available for this chart.

### Global Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| global.jobs.kubectl.affinity | object | `{}` | Set affinity rules |
| global.jobs.kubectl.annotations | object | `{"helm.sh/hook-delete-policy":"before-hook-creation,hook-succeeded"}` | Annotations to add to the certgen job. |
| global.jobs.kubectl.image.pullPolicy | string | `"IfNotPresent"` | Set the image pull policy of the helm chart job |
| global.jobs.kubectl.image.registry | string | `"docker.io"` | Set the image repository of the helm chart job |
| global.jobs.kubectl.image.repository | string | `"clastix/kubectl"` | Set the image repository of the helm chart job |
| global.jobs.kubectl.image.tag | string | `""` | Set the image tag of the helm chart job |
| global.jobs.kubectl.nodeSelector | object | `{}` | Set the node selector |
| global.jobs.kubectl.podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Security context for the job pods. |
| global.jobs.kubectl.priorityClassName | string | `""` | Set a pod priorityClassName |
| global.jobs.kubectl.resources | object | `{}` | Job resources |
| global.jobs.kubectl.restartPolicy | string | `"Never"` | Set the restartPolicy |
| global.jobs.kubectl.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsGroup":1002,"runAsNonRoot":true,"runAsUser":1002}` | Security context for the job containers. |
| global.jobs.kubectl.tolerations | list | `[]` | Set list of tolerations |
| global.jobs.kubectl.topologySpreadConstraints | list | `[]` | Set Topology Spread Constraints |
| global.jobs.kubectl.ttlSecondsAfterFinished | int | `60` | Sets the ttl in seconds after a finished certgen job is deleted. Set to -1 to never delete. |

### CustomResourceDefinition Lifecycle

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| crds.annnotations | object | `{}` | Extra Annotations for CRDs |
| crds.install | bool | `true` | Install the CustomResourceDefinitions (This also manages the lifecycle of the CRDs for update operations) |
| crds.keep | bool | `false` | Keep the annotations if deleted |
| crds.labels | object | `{}` | Extra Labels for CRDs |

###  Capsule-Proxy

Deploy a dedicated [capsule-proxy](https://artifacthub.io/packages/helm/projectcapsule/capsule-proxy) with the addon.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| proxy.autoscaling.enabled | bool | `true` | Specifies whether an hpa for capsule-proxy should be created. |
| proxy.autoscaling.targetCPUUtilizationPercentage | int | `75` |  |
| proxy.certManager | object | `{"generateCertificates":false}` | Get Certificate from cert-manager from Capsule-Proxy |
| proxy.enabled | bool | `true` |  |
| proxy.options | object | `{"clientConnectionBurst":100,"clientConnectionQPS":50,"extraArgs":["--feature-gates=ProxyClusterScoped=true","--feature-gates=ProxyAllNamespaced=true"],"logLevel":"10"}` | Enable Required FeatureGates |
| proxy.options.clientConnectionBurst | int | `100` | Burst to use for interacting with kubernetes API Server. |
| proxy.options.clientConnectionQPS | int | `50` | QPS to use for interacting with Kubernetes API Server. |

### General Parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Set affinity rules |
| args.extraArgs | list | `[]` | A list of extra arguments to add to the capsule-argo-addon |
| args.logLevel | int | `4` | Log Level |
| config.create | bool | `true` | Create Plugin Configuration |
| config.name | string | `"default"` | Plugin Configuration Name |
| config.spec | object | `{}` | Config Specification |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` | Set the image pull policy. |
| image.registry | string | `"ghcr.io"` | Set the image registry |
| image.repository | string | `"peak-scale/capsule-argo-addon"` | Set the image repository |
| image.tag | string | `""` | Overrides the image tag whose default is the chart appVersion. |
| imagePullSecrets | list | `[]` | Configuration for `imagePullSecrets` so that you can use a private images registry. |
| livenessProbe | object | `{"httpGet":{"path":"/healthz","port":10080}}` | Configure the liveness probe using Deployment probe spec |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | Set the node selector |
| podAnnotations | object | `{}` | Annotations to add |
| podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Set the securityContext |
| priorityClassName | string | `""` | Set the priority class name of the Capsule pod |
| rbac.enabled | bool | `true` | Enable bootstraping of RBAC resources |
| readinessProbe | object | `{"httpGet":{"path":"/readyz","port":10080}}` | Configure the readiness probe using Deployment probe spec |
| replicaCount | int | `1` | Amount of replicas |
| resources | object | `{"limits":{"cpu":"200m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` | Set the resource requests/limits |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsNonRoot":true,"runAsUser":1000}` | Set the securityContext for the container |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account. |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created. |
| serviceAccount.name | string | `""` | The name of the service account to use. |
| tolerations | list | `[]` | Set list of tolerations |
| topologySpreadConstraints | list | `[]` | Set topology spread constraints |

### Monitoring Parameters

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| monitoring.enabled | bool | `false` | Enable Monitoring of the Operator |
| monitoring.rules.annotations | object | `{}` | Assign additional Annotations |
| monitoring.rules.enabled | bool | `true` | Enable deployment of PrometheusRules |
| monitoring.rules.groups | list | `[{"name":"TranslatorAlerts","rules":[{"alert":"TranslatorNotReady","annotations":{"description":"The Translator {{ $labels.name }} has been in a NotReady state for over 5 minutes.","summary":"Translator {{ $labels.name }} is not ready"},"expr":"cca_translator_condition{status=\"NotReady\"} == 1","for":"5m","labels":{"severity":"critical"}}]}]` | Prometheus Groups for the rule |
| monitoring.rules.labels | object | `{}` | Assign additional labels |
| monitoring.rules.namespace | string | `""` | Install the rules into a different Namespace, as the monitoring stack one (default: the release one) |
| monitoring.serviceMonitor.annotations | object | `{}` | Assign additional Annotations |
| monitoring.serviceMonitor.enabled | bool | `true` | Enable ServiceMonitor |
| monitoring.serviceMonitor.endpoint.interval | string | `"15s"` | Set the scrape interval for the endpoint of the serviceMonitor |
| monitoring.serviceMonitor.endpoint.metricRelabelings | list | `[]` | Set metricRelabelings for the endpoint of the serviceMonitor |
| monitoring.serviceMonitor.endpoint.relabelings | list | `[]` | Set relabelings for the endpoint of the serviceMonitor |
| monitoring.serviceMonitor.endpoint.scrapeTimeout | string | `""` | Set the scrape timeout for the endpoint of the serviceMonitor |
| monitoring.serviceMonitor.jobLabel | string | `"app.kubernetes.io/name"` | Prometheus Joblabel |
| monitoring.serviceMonitor.labels | object | `{}` | Assign additional labels according to Prometheus' serviceMonitorSelector matching labels |
| monitoring.serviceMonitor.matchLabels | object | `{}` | Change matching labels |
| monitoring.serviceMonitor.namespace | string | `""` | Install the ServiceMonitor into a different Namespace, as the monitoring stack one (default: the release one) |
| monitoring.serviceMonitor.serviceAccount.name | string | `""` |  |
| monitoring.serviceMonitor.serviceAccount.namespace | string | `""` |  |
| monitoring.serviceMonitor.targetLabels | list | `[]` | Set targetLabels for the serviceMonitor |