# Reference

Packages:

- [addons.projectcapsule.dev/v1alpha1](#addonsprojectcapsuledevv1alpha1)

# addons.projectcapsule.dev/v1alpha1

Resource Types:

- [ArgoAddon](#argoaddon)

- [ArgoTranslator](#argotranslator)




## ArgoAddon






ArgoAddon is the Schema for the ArgoAddons API

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | ArgoAddon | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#argoaddonspec)** | object | ArgoAddonSpec defines the desired state of ArgoAddon | false |
| **[status](#argoaddonstatus)** | object | ArgoAddonStatus defines the observed state of ArgoAddon | false |


### ArgoAddon.spec



ArgoAddonSpec defines the desired state of ArgoAddon

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[argo](#argoaddonspecargo)** | object | ArgoCD configuration | true |
| **force** | boolean | When force is enabled, approjects which already exist with the same name as a tenant will be adopted
and overwritten. When disabled the approjects will not be changed or adopted.
This is true for any other resource as well<br/><i>Default</i>: false<br/> | true |
| **[proxy](#argoaddonspecproxy)** | object | Capsule-Proxy configuration for the controller<br/><i>Default</i>: map[]<br/> | false |


### ArgoAddon.spec.argo



ArgoCD configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **serviceAccountNamespace** | string | Default Namespace to create ServiceAccounts used by arog-cd
The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
It's best to have a dedicated namespace for these serviceaccounts | true |
| **destination** | string | If you are not using the capsule-proxy integration this destination is registered
for each appproject.<br/><i>Default</i>: https://kubernetes.default.svc<br/> | false |
| **destinationServiceAccounts** | boolean | This is a feature which will be released with argocd +v2.13.0
If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha<br/><i>Default</i>: false<br/> | false |
| **namespace** | string | Namespace where the ArgoCD instance is running<br/><i>Default</i>: argocd<br/> | false |
| **rbacConfigMap** | string | Name of the ArgoCD rbac configmap (required for the controller)<br/><i>Default</i>: argocd-rbac-cm<br/> | false |


### ArgoAddon.spec.proxy



Capsule-Proxy configuration for the controller

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enabled** | boolean | Enable the capsule-proxy integration.
This automatically creates services for tenants and registers them as destination
on the argo appproject.<br/><i>Default</i>: false<br/> | false |
| **serviceName** | string | Name of the capsule-proxy service<br/><i>Default</i>: capsule-proxy<br/> | false |
| **serviceNamespace** | string |  Namespace where the capsule-proxy service is running<br/><i>Default</i>: capsule-system<br/> | false |
| **servicePort** | integer | Port of the capsule-proxy service<br/><i>Format</i>: int32<br/><i>Default</i>: 9001<br/> | false |
| **tls** | boolean | Port of the capsule-proxy service<br/><i>Default</i>: true<br/> | false |


### ArgoAddon.status



ArgoAddonStatus defines the observed state of ArgoAddon

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[loaded](#argoaddonstatusloaded)** | object | Last applied valid configuration | false |


### ArgoAddon.status.loaded



Last applied valid configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[argo](#argoaddonstatusloadedargo)** | object | ArgoCD configuration | true |
| **force** | boolean | When force is enabled, approjects which already exist with the same name as a tenant will be adopted
and overwritten. When disabled the approjects will not be changed or adopted.
This is true for any other resource as well<br/><i>Default</i>: false<br/> | true |
| **[proxy](#argoaddonstatusloadedproxy)** | object | Capsule-Proxy configuration for the controller<br/><i>Default</i>: map[]<br/> | false |


### ArgoAddon.status.loaded.argo



ArgoCD configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **serviceAccountNamespace** | string | Default Namespace to create ServiceAccounts used by arog-cd
The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
It's best to have a dedicated namespace for these serviceaccounts | true |
| **destination** | string | If you are not using the capsule-proxy integration this destination is registered
for each appproject.<br/><i>Default</i>: https://kubernetes.default.svc<br/> | false |
| **destinationServiceAccounts** | boolean | This is a feature which will be released with argocd +v2.13.0
If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha<br/><i>Default</i>: false<br/> | false |
| **namespace** | string | Namespace where the ArgoCD instance is running<br/><i>Default</i>: argocd<br/> | false |
| **rbacConfigMap** | string | Name of the ArgoCD rbac configmap (required for the controller)<br/><i>Default</i>: argocd-rbac-cm<br/> | false |


### ArgoAddon.status.loaded.proxy



Capsule-Proxy configuration for the controller

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enabled** | boolean | Enable the capsule-proxy integration.
This automatically creates services for tenants and registers them as destination
on the argo appproject.<br/><i>Default</i>: false<br/> | false |
| **serviceName** | string | Name of the capsule-proxy service<br/><i>Default</i>: capsule-proxy<br/> | false |
| **serviceNamespace** | string |  Namespace where the capsule-proxy service is running<br/><i>Default</i>: capsule-system<br/> | false |
| **servicePort** | integer | Port of the capsule-proxy service<br/><i>Format</i>: int32<br/><i>Default</i>: 9001<br/> | false |
| **tls** | boolean | Port of the capsule-proxy service<br/><i>Default</i>: true<br/> | false |

## ArgoTranslator






ArgoTranslator is the Schema for the argotranslators API

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | ArgoTranslator | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#argotranslatorspec)** | object | ArgoTranslatorSpec defines the desired state of ArgoTranslator | false |
| **[status](#argotranslatorstatus)** | object | ArgoTranslatorStatus defines the observed state of ArgoTranslator | false |


### ArgoTranslator.spec



ArgoTranslatorSpec defines the desired state of ArgoTranslator

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **customPolicy** | string | In this field you can define custom policies. It must result in a valid argocd policy format (CSV)
You can use Sprig Templating with this field | false |
| **[roles](#argotranslatorspecrolesindex)** | []object | Application-Project Roles for the tenant | false |
| **[selector](#argotranslatorspecselector)** | object | Selector to match tenants which are used for the translator | false |
| **[settings](#argotranslatorspecsettings)** | object | Additional settings for the argocd project | false |


### ArgoTranslator.spec.roles[index]



Define Permission mappings for an ArogCD Project

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **clusterRoles** | []string | TenantRoles selects tenant users based on their cluster roles to this Permission | false |
| **name** | string | Name for permission mapping | false |
| **owner** | boolean | Define if the selected users are owners of the appproject. Being owner allows the users
to update the project and effectively manage everything. By default the selected users get
read-only access to the project.<br/><i>Default</i>: false<br/> | false |
| **[policies](#argotranslatorspecrolesindexpoliciesindex)** | []object | Roles are reflected in the argocd rbac configmap | false |


### ArgoTranslator.spec.roles[index].policies[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **action** | []string | Allowed actions for this permission. You may specify multiple actions. To allow all actions use "*"<br/><i>Default</i>: [get]<br/> | false |
| **path** | string | You may specify a custom path for the resource. The available path for argo is <app-project>/<app-ns>/<app-name>
however <app-project> is already set to the argocd project name. Therefor you can only add <app-ns>/<app-name><br/><i>Default</i>: *<br/> | false |
| **resource** | string | Name for permission mapping | false |
| **verb** | string | Verb for this permission (can be allow, deny)<br/><i>Default</i>: allow<br/> | false |


### ArgoTranslator.spec.selector



Selector to match tenants which are used for the translator

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[matchExpressions](#argotranslatorspecselectormatchexpressionsindex)** | []object | matchExpressions is a list of label selector requirements. The requirements are ANDed. | false |
| **matchLabels** | map[string]string | matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed. | false |


### ArgoTranslator.spec.selector.matchExpressions[index]



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **key** | string | key is the label key that the selector applies to. | true |
| **operator** | string | operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist. | true |
| **values** | []string | values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch. | false |


### ArgoTranslator.spec.settings



Additional settings for the argocd project

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[structured](#argotranslatorspecsettingsstructured)** | object | Structured Properties for the argocd project | false |
| **template** | string | Use a template to generate to argo project settings | false |


### ArgoTranslator.spec.settings.structured



Structured Properties for the argocd project

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[meta](#argotranslatorspecsettingsstructuredmeta)** | object | Project Metadata | false |
| **[spec](#argotranslatorspecsettingsstructuredspec)** | object | Application Project Spec (Upstream ArgoCD) | false |


### ArgoTranslator.spec.settings.structured.meta



Project Metadata

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **annotations** | map[string]string | Annotations for the project | false |
| **finalizers** | []string | Finalizers for the project | false |
| **labels** | map[string]string | Labels for the project | false |


### ArgoTranslator.spec.settings.structured.spec



Application Project Spec (Upstream ArgoCD)

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[clusterResourceBlacklist](#argotranslatorspecsettingsstructuredspecclusterresourceblacklistindex)** | []object | ClusterResourceBlacklist contains list of blacklisted cluster level resources | false |
| **[clusterResourceWhitelist](#argotranslatorspecsettingsstructuredspecclusterresourcewhitelistindex)** | []object | ClusterResourceWhitelist contains list of whitelisted cluster level resources | false |
| **description** | string | Description contains optional project description | false |
| **[destinationServiceAccounts](#argotranslatorspecsettingsstructuredspecdestinationserviceaccountsindex)** | []object | DestinationServiceAccounts holds information about the service accounts to be impersonated for the application sync operation for each destination. | false |
| **[destinations](#argotranslatorspecsettingsstructuredspecdestinationsindex)** | []object | Destinations contains list of destinations available for deployment | false |
| **[namespaceResourceBlacklist](#argotranslatorspecsettingsstructuredspecnamespaceresourceblacklistindex)** | []object | NamespaceResourceBlacklist contains list of blacklisted namespace level resources | false |
| **[namespaceResourceWhitelist](#argotranslatorspecsettingsstructuredspecnamespaceresourcewhitelistindex)** | []object | NamespaceResourceWhitelist contains list of whitelisted namespace level resources | false |
| **[orphanedResources](#argotranslatorspecsettingsstructuredspecorphanedresources)** | object | OrphanedResources specifies if controller should monitor orphaned resources of apps in this project | false |
| **permitOnlyProjectScopedClusters** | boolean | PermitOnlyProjectScopedClusters determines whether destinations can only reference clusters which are project-scoped | false |
| **[roles](#argotranslatorspecsettingsstructuredspecrolesindex)** | []object | Roles are user defined RBAC roles associated with this project | false |
| **[signatureKeys](#argotranslatorspecsettingsstructuredspecsignaturekeysindex)** | []object | SignatureKeys contains a list of PGP key IDs that commits in Git must be signed with in order to be allowed for sync | false |
| **sourceNamespaces** | []string | SourceNamespaces defines the namespaces application resources are allowed to be created in | false |
| **sourceRepos** | []string | SourceRepos contains list of repository URLs which can be used for deployment | false |
| **[syncWindows](#argotranslatorspecsettingsstructuredspecsyncwindowsindex)** | []object | SyncWindows controls when syncs can be run for apps in this project | false |


### ArgoTranslator.spec.settings.structured.spec.clusterResourceBlacklist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.structured.spec.clusterResourceWhitelist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.structured.spec.destinationServiceAccounts[index]



ApplicationDestinationServiceAccount holds information about the service account to be impersonated for the application sync operation.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **defaultServiceAccount** | string | DefaultServiceAccount to be used for impersonation during the sync operation | true |
| **server** | string | Server specifies the URL of the target cluster's Kubernetes control plane API. | true |
| **namespace** | string | Namespace specifies the target namespace for the application's resources. | false |


### ArgoTranslator.spec.settings.structured.spec.destinations[index]



ApplicationDestination holds information about the application's destination

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name is an alternate way of specifying the target cluster by its symbolic name. This must be set if Server is not set. | false |
| **namespace** | string | Namespace specifies the target namespace for the application's resources.
The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace | false |
| **server** | string | Server specifies the URL of the target cluster's Kubernetes control plane API. This must be set if Name is not set. | false |


### ArgoTranslator.spec.settings.structured.spec.namespaceResourceBlacklist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.structured.spec.namespaceResourceWhitelist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.structured.spec.orphanedResources



OrphanedResources specifies if controller should monitor orphaned resources of apps in this project

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[ignore](#argotranslatorspecsettingsstructuredspecorphanedresourcesignoreindex)** | []object | Ignore contains a list of resources that are to be excluded from orphaned resources monitoring | false |
| **warn** | boolean | Warn indicates if warning condition should be created for apps which have orphaned resources | false |


### ArgoTranslator.spec.settings.structured.spec.orphanedResources.ignore[index]



OrphanedResourceKey is a reference to a resource to be ignored from

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | false |
| **kind** | string |  | false |
| **name** | string |  | false |


### ArgoTranslator.spec.settings.structured.spec.roles[index]



ProjectRole represents a role that has access to a project

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name is a name for this role | true |
| **description** | string | Description is a description of the role | false |
| **groups** | []string | Groups are a list of OIDC group claims bound to this role | false |
| **[jwtTokens](#argotranslatorspecsettingsstructuredspecrolesindexjwttokensindex)** | []object | JWTTokens are a list of generated JWT tokens bound to this role | false |
| **policies** | []string | Policies Stores a list of casbin formatted strings that define access policies for the role in the project | false |


### ArgoTranslator.spec.settings.structured.spec.roles[index].jwtTokens[index]



JWTToken holds the issuedAt and expiresAt values of a token

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **iat** | integer | <br/><i>Format</i>: int64<br/> | true |
| **exp** | integer | <br/><i>Format</i>: int64<br/> | false |
| **id** | string |  | false |


### ArgoTranslator.spec.settings.structured.spec.signatureKeys[index]



SignatureKey is the specification of a key required to verify commit signatures with

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **keyID** | string | The ID of the key in hexadecimal notation | true |


### ArgoTranslator.spec.settings.structured.spec.syncWindows[index]



SyncWindow contains the kind, time, duration and attributes that are used to assign the syncWindows to apps

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **applications** | []string | Applications contains a list of applications that the window will apply to | false |
| **clusters** | []string | Clusters contains a list of clusters that the window will apply to | false |
| **duration** | string | Duration is the amount of time the sync window will be open | false |
| **kind** | string | Kind defines if the window allows or blocks syncs | false |
| **manualSync** | boolean | ManualSync enables manual syncs when they would otherwise be blocked | false |
| **namespaces** | []string | Namespaces contains a list of namespaces that the window will apply to | false |
| **schedule** | string | Schedule is the time the window will begin, specified in cron format | false |
| **timeZone** | string | TimeZone of the sync that will be applied to the schedule | false |


### ArgoTranslator.status



ArgoTranslatorStatus defines the observed state of ArgoTranslator

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **ready** | string | Ready field indicating overall readiness of the translator | false |
| **size** | integer | Amount of tenants selected by this translator | false |
| **[tenants](#argotranslatorstatustenantsindex)** | []object | List of tenants selected by this translator | false |


### ArgoTranslator.status.tenants[index]





| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[condition](#argotranslatorstatustenantsindexcondition)** | object | Conditions represent the latest available observations of an object's state | false |
| **name** | string | List of tenants selected by this translator | false |
| **uid** | string | UID of the tracked Tenant to pin point tracking | false |


### ArgoTranslator.status.tenants[index].condition



Conditions represent the latest available observations of an object's state

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |