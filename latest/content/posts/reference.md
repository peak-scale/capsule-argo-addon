---
title: "🔬 API Reference"
summary: "Complete API reference for CRDs of the addon"
ShowToc: true
TocOpen: false
weight: 4
---


Packages:

- [config.projectcapsule.dev/v1alpha1](#configprojectcapsuledevv1alpha1)

# config.projectcapsule.dev/v1alpha1

Resource Types:

- [ArgoAddon](#argoaddon)

- [ArgoTranslator](#argotranslator)




## ArgoAddon






ArgoAddon is the Schema for the ArgoAddons API

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | config.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | ArgoAddon | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#argoaddonspec)** | object | ArgoAddonSpec defines the desired state of ArgoAddon | false |
| **[status](#argoaddonstatus)** | object | ArgoAddonStatus defines the observed state of ArgoAddon | false |


### ArgoAddon.spec



ArgoAddonSpec defines the desired state of ArgoAddon

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[config](#argoaddonspecconfig)** | object | Controller Configuration | false |


### ArgoAddon.spec.config



Controller Configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[argo](#argoaddonspecconfigargo)** | object | ArgoCD configuration<br/><i>Default</i>: map[namespace:argocd rbacConfigMap:argocd-rbac-cm]<br/> | false |
| **[proxy](#argoaddonspecconfigproxy)** | object | Capsule configuration | false |


### ArgoAddon.spec.config.argo



ArgoCD configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **namespace** | string | Namespace where the ArgoCD instance is running | false |
| **rbacConfigMap** | string | Name of the ArgoCD rbac configmap (required for the controller) | false |


### ArgoAddon.spec.config.proxy



Capsule configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enabled** | boolean | Enable the capsule-proxy integration. This automatically creates ServiceAccounts for tenants and registers them as destination
on the argo appproject.<br/><i>Default</i>: true<br/> | false |
| **serviceAccountNamespace** | string | Default Namespace to create ServiceAccounts in for proxy access.
Can be overwritten on tenant-basis | false |
| **serviceName** | string | Name of the capsule-proxy service<br/><i>Default</i>: capsule-proxy<br/> | false |
| **serviceNamespace** | string |  Namespace where the capsule-proxy service is running<br/><i>Default</i>: capsule-system<br/> | false |
| **servicePort** | integer | Port of the capsule-proxy service<br/><i>Format</i>: int32<br/><i>Default</i>: 9001<br/> | false |


### ArgoAddon.status



ArgoAddonStatus defines the observed state of ArgoAddon

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[loaded](#argoaddonstatusloaded)** | object | Last applied valid configuration | false |


### ArgoAddon.status.loaded



Last applied valid configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[argo](#argoaddonstatusloadedargo)** | object | ArgoCD configuration<br/><i>Default</i>: map[namespace:argocd rbacConfigMap:argocd-rbac-cm]<br/> | false |
| **[proxy](#argoaddonstatusloadedproxy)** | object | Capsule configuration | false |


### ArgoAddon.status.loaded.argo



ArgoCD configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **namespace** | string | Namespace where the ArgoCD instance is running | false |
| **rbacConfigMap** | string | Name of the ArgoCD rbac configmap (required for the controller) | false |


### ArgoAddon.status.loaded.proxy



Capsule configuration

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **enabled** | boolean | Enable the capsule-proxy integration. This automatically creates ServiceAccounts for tenants and registers them as destination
on the argo appproject.<br/><i>Default</i>: true<br/> | false |
| **serviceAccountNamespace** | string | Default Namespace to create ServiceAccounts in for proxy access.
Can be overwritten on tenant-basis | false |
| **serviceName** | string | Name of the capsule-proxy service<br/><i>Default</i>: capsule-proxy<br/> | false |
| **serviceNamespace** | string |  Namespace where the capsule-proxy service is running<br/><i>Default</i>: capsule-system<br/> | false |
| **servicePort** | integer | Port of the capsule-proxy service<br/><i>Format</i>: int32<br/><i>Default</i>: 9001<br/> | false |

## ArgoTranslator






ArgoTranslator is the Schema for the argotranslators API

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | config.projectcapsule.dev/v1alpha1 | true |
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



Define Permission mappings for an ArogCD Project

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
| **[clusterResourceBlacklist](#argotranslatorspecsettingsclusterresourceblacklistindex)** | []object | ClusterResourceBlacklist contains list of blacklisted cluster level resources | false |
| **[clusterResourceWhitelist](#argotranslatorspecsettingsclusterresourcewhitelistindex)** | []object | ClusterResourceWhitelist contains list of whitelisted cluster level resources | false |
| **[destinations](#argotranslatorspecsettingsdestinationsindex)** | []object | Add destinations for the project | false |
| **[meta](#argotranslatorspecsettingsmeta)** | object | Project Metadata | false |
| **[namespaceResourceBlacklist](#argotranslatorspecsettingsnamespaceresourceblacklistindex)** | []object | NamespaceResourceBlacklist contains list of blacklisted namespace level resources | false |
| **[namespaceResourceWhitelist](#argotranslatorspecsettingsnamespaceresourcewhitelistindex)** | []object | NamespaceResourceWhitelist contains list of whitelisted namespace level resources | false |
| **sourceNamespaces** | []string | Namespaces where applications for this project can come from | false |
| **[syncWindows](#argotranslatorspecsettingssyncwindowsindexindex)** | [][]object | SyncWindows controls when syncs can be run for apps in this project | false |


### ArgoTranslator.spec.settings.clusterResourceBlacklist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.clusterResourceWhitelist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.destinations[index]



ApplicationDestination holds information about the application's destination

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | Name is an alternate way of specifying the target cluster by its symbolic name. This must be set if Server is not set. | false |
| **namespace** | string | Namespace specifies the target namespace for the application's resources.
The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace | false |
| **server** | string | Server specifies the URL of the target cluster's Kubernetes control plane API. This must be set if Name is not set. | false |


### ArgoTranslator.spec.settings.meta



Project Metadata

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **annotations** | map[string]string | Annotations for the project | false |
| **finalizers** | []string | Finalizers for the project | false |
| **labels** | map[string]string | Labels for the project | false |


### ArgoTranslator.spec.settings.namespaceResourceBlacklist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.namespaceResourceWhitelist[index]



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **group** | string |  | true |
| **kind** | string |  | true |


### ArgoTranslator.spec.settings.syncWindows[index][index]



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
| **[conditions](#argotranslatorstatusconditionsindex)** | []object |  | false |
| **tenants** | []string | List of tenants selected by this translator | false |


### ArgoTranslator.status.conditions[index]



Condition contains details for one aspect of the current state of this API Resource.

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