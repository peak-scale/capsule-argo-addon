
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

[Argo(CD)](https://artifacthub.io/packages/helm/argo/argo-cd) is required to be installed in the [v2.13.0](https://github.com/argoproj/argo-cd/releases/tag/v2.13.0) version or above. This version adds support for `destinationServiceAccounts`, which makes the appprojects much more secure. [See our Manifest](../e2e/objects/distro/argo.flux.yaml)

**You must enable `application.sync.impersonation.enabled: "true"` impersonation explicitly in `argocd-cm`, by default it wont be working**. [Read More](https://argo-cd.readthedocs.io/en/stable/proposals/decouple-application-sync-user-using-impersonation/#component-argocd-application-controller)

This is the default configuration, you may need to adjust if argocd is running in a different namespace.

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
    # Defaults
    namespace: argocd
    rbacConfigMap: argocd-rbac-cm
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