# ArgoAddon

The configure the controller we have a dedicated cluster-scoped resource. This is to ensure type-safety regarding the configuration values. There can be any amount of configurations. However a controller only selects one configuration at a time. Updates to the configuration are reflected without the need to restart the controller.

Here's a simple configuration which configures the argocd properties and the capsule-proxy service:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
  kind: ArgoAddon
  metadata:
    name: default
  spec:
    argo:
      namespace: argocd
      rbacConfigMap: argocd-rbac-cm
    force: false
    proxy:
      enabled: true
      serviceAccountNamespace: capsule-argo-addon
      serviceName: capsule-argo-addon-proxy
      serviceNamespace: capsule-argo-addon
      servicePort: 9001
```

The controller is then started with the name of the configuration:

```shell
capsule-argo-addon -setting-name "default"
```

[View the Reference for all possible options](../reference.md)

## Controller-Options

The following arguments can be passed to the controller

```shell
  -health-probe-bind-address string
    	The address the probe endpoint binds to. (default ":8081")
  -kubeconfig string
    	Paths to a kubeconfig. Only required if out-of-cluster.
  -leader-elect
    	Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
  -metrics-bind-address string
    	The address the metric endpoint binds to. (default ":8080")
  -setting-name string
    	The setting name to use for this controller instance (default "default")
  -zap-devel
    	Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn). Production Mode defaults(encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error) (default true)
  -zap-encoder value
    	Zap log encoding (one of 'json' or 'console')
  -zap-log-level value
    	Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', or any integer value > 0 which corresponds to custom debug levels of increasing verbosity
  -zap-stacktrace-level value
    	Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic').
  -zap-time-encoding value
    	Zap time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano'). Defaults to 'epoch'.
```
