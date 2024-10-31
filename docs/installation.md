
# Installation

The Installation of the addon is only supported via Helm-Chart. Any other method is not officially supported.

## Requirements

The following is expected to be installed (including their CRDs). Without these the addon won't work.

### Capsule

[Capsule](https://artifacthub.io/packages/helm/projectcapsule/capsule) is required to run this addon. You must use a version greater or equal to `0.7.1`, otherwise the users can hijack system namespaces.

#### Tenants

Note that this operator adds owners to tenants (serviceaccounts). If you are managing tenants via gitops, make sure to ignore these changes.

**Argo**:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: tenants
spec:
  ...
  ignoreDifferences:
  - group: capsule.clastix.io
    kind: Tenant
    jqPathExpressions:
      - >-
        .spec.owners[] | select(.kind == "ServiceAccount")
```

**Flux**:

### ArgoCD

[Argo(CD)](https://artifacthub.io/packages/helm/argo/argo-cd) is recommended to be installed in the [v2.13.0](https://github.com/argoproj/argo-cd/releases/tag/v2.13.0) version. This version adds support for `destinationServiceAccounts`, which makes the appprojects much more secure. Here's how the addon is best configured for the corresponding argo-cd versions

#### `>=v2.13.0` (default)

This is the default configuration, this only works if the argo-cd version `>=v2.13.0` is used.

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoAddon
metadata:
  name: default
spec:
  # Specify as needed
  force: false

  # Configure argo
  argo:
    # This should point to the in-cluster api-endpoint
    destination: https://kubernetes.default.svc
    # Enables the usage of destination serviceaccounts
    destinationServiceAccounts: true

    # Defaults
    namespace: argocd
    rbacConfigMap: argocd-rbac-cm

  # Disable the proxy
  proxy:
    enabled: false
```

#### `<v2.13.0`

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoAddon
metadata:
  name: default
spec:
  # Specify as needed
  force: false

  # Configure argo
  argo:
    # This should point to the in-cluster api-endpoint
    destination: https://kubernetes.default.svc
    destinationServiceAccounts: false

    # Defaults
    namespace: argocd
    rbacConfigMap: argocd-rbac-cm

  # Disable the proxy
  proxy:
    enabled: true
```

## Capsule

For ArgoCD to work properly we need to allow `list`, `get` and `watch` within all the namespaces, where we are going to deploy serviceaccounts for the tenants. You must provision these bindings manually, since they heavily depend on how you are deploying the addon:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: list-all-resources
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["list", "watch", "get"]
```

This `ClusterRole` is deployed as part of the helm chart, you need to manually bind them to custom 

```yaml
---
# ClusterRoleBinding that binds the ClusterRole to all ServiceAccounts
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: list-all-resources-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: list-all-resources     # The ClusterRole to bind
subjects:
  # Grants permissions to all serviceaccounts in the namespace 'capsule-argo-addon' 
  - kind: Group
    name: system:serviceaccounts:capsule-argo-addon
    apiGroup: rbac.authorization.k8s.io
  # Grants permissions to all serviceaccounts in the namespace 'privileged-service-accounts' 
  - kind: Group
    name: system:serviceaccounts:privileged-service-accounts
    apiGroup: rbac.authorization.k8s.io
```

This should no raise any security concerns, since the users won't be able to access the serviceaccounts.

These Namespaces must also be part of the capsuleConfiguration, to be considered as Tenant users. You must add the namespace where the serviceaccounts are deployed to, to the capsuleUsers. For example:

```yaml
apiVersion: capsule.clastix.io/v1beta2
kind: CapsuleConfiguration
metadata:
  name: default
spec:
  ...
  # The default installation namespace
  userGroups:
  - system:serviceaccounts:capsule-argo-addon
  # Add other namespaces (etc..)
  - system:serviceaccounts:privileged-service-accounts
```

## Helm

[Artifact Hub](https://artifacthub.io/packages/helm/capsule-argo-addon/capsule-argo-addon)

Currently we support installation via Helm-Chart click the badge or [here](https://artifacthub.io/packages/helm/capsule-argo-addon/capsule-argo-addon) to view instructions and possible values on the chart.

### Capsule-Proxy

> No longer needed with Argo 2.13.0.

The [capsule-proxy](https://artifacthub.io/packages/helm/projectcapsule/capsule-proxy) is used to allow serviceaccounts to just see what they should see within the boundaries of your tenant. It is optional to use the proxy and it can be disabled via the [configuration](./config.md).

If you plan to use the capsule-proxy, we recommend installing a dedicated capsule-proxy instance for the addon, because Argo puts a lot of pressure on the proxy.

With the [Helm Chart](#helm) a dedicated capsule-proxy is already installed (exclusive CRDs) by when enabling the integration. Adjust this according to your needs and your setups.