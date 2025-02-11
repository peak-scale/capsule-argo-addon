# Templating

For templating you have [Go Sprig](https://masterminds.github.io/sprig/) available. The following custom functions are additionally available:

- `toYaml`
- `fromYaml`
- `toJson`
- `fromJson`
- `toToml`
- `fromToml`

## Context

The follwing data context is available for templating:

```yaml
Config:
    Argo:
        Destination: ""
        DestinationServiceAccounts: false
        Namespace: argocd
        RBACConfigMap: argocd-rbac-cm
        ServiceAccountNamespace: ""
    Decouple: false
    Force: false
    ReadOnly: false
Endpoint: ""
Tenant:
    Name: example-tenant
    Namespaces:
        - namespace1
        - namespace2
    Object:
        ObjectMeta:
            Annotations: {}
            CreationTimestamp:
                Time: {}
            Finalizers: []
            GenerateName: ""
            Generation: 0
            Labels: {}
            ManagedFields: []
            Name: example-tenant
            Namespace: ""
            OwnerReferences: []
            ResourceVersion: ""
            SelfLink: ""
            UID: ""
        Spec:
            AdditionalRoleBindings: []
            Cordoned: false
            ImagePullPolicies: []
            IngressOptions:
                AllowWildcardHostnames: false
                HostnameCollisionScope: ""
            LimitRanges:
                Items: []
            NetworkPolicies:
                Items: []
            NodeSelector: {}
            Owners:
                - ClusterRoles: []
                  Kind: User
                  Name: example-user
                  ProxyOperations: []
                - ClusterRoles: []
                  Kind: Group
                  Name: example-group
                  ProxyOperations: []
            PreventDeletion: false
            ResourceQuota:
                Items: []
                Scope: ""
        Status:
            Namespaces:
                - namespace1
                - namespace2
            Size: 0
            State: ""
        TypeMeta:
            APIVersion: ""
            Kind: ""
```

You can access them via their Map-Path (eg. `.Config.Argo.Namespace`)
