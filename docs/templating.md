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
    AllowRepositoryCreation: false
    Argo:
        Destination: ""
        DestinationServiceAccounts: false
        Namespace: argocd
        RBACConfigMap: argocd-rbac-cm
        ServiceAccountClusterRoles: []
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
            Data:
                Raw: []
            GatewayOptions: {}
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
                - Annotations: {}
                  CoreOwnerSpec:
                    ClusterRoles: []
                    UserSpec:
                        Kind: User
                        Name: example-user
                  Labels: {}
                  ProxyOperations: []
                - Annotations: {}
                  CoreOwnerSpec:
                    ClusterRoles: []
                    UserSpec:
                        Kind: Group
                        Name: example-group
                  Labels: {}
                  ProxyOperations: []
            Permissions:
                AllowOwnerPromotion: false
                MatchOwners: []
            PreventDeletion: false
            ResourceQuota:
                Items: []
                Scope: ""
            Rules: []
        Status:
            Conditions: []
            Namespaces:
                - namespace1
                - namespace2
            ObservedGeneration: 0
            Owners: []
            Promotions: []
            Size: 0
            Spaces: []
            State: ""
            TenantAvailableStatus:
                Classes:
                    DeviceClasses: []
                    GatewayClasses: []
                    PriorityClasses: []
                    RuntimeClasses: []
                    StorageClasses: []
        TypeMeta:
            APIVersion: ""
            Kind: ""
```

You can access them via their Map-Path (eg. `.Config.Argo.Namespace`)
