# API Reference

Packages:

- [addons.projectcapsule.dev/v1alpha1](#addonsprojectcapsuledevv1alpha1)

# addons.projectcapsule.dev/v1alpha1

Resource Types:

- [ArgoAddon](#argoaddon)

- [ArgoTranslator](#argotranslator)




## ArgoAddon
<sup><sup>[↩ Parent](#addonsprojectcapsuledevv1alpha1 )</sup></sup>






ArgoAddon is the Schema for the ArgoAddons API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>addons.projectcapsule.dev/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>ArgoAddon</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#argoaddonspec">spec</a></b></td>
        <td>object</td>
        <td>
          ArgoAddonSpec defines the desired state of ArgoAddon.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argoaddonstatus">status</a></b></td>
        <td>object</td>
        <td>
          ArgoAddonStatus defines the observed state of ArgoAddon.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoAddon.spec
<sup><sup>[↩ Parent](#argoaddon)</sup></sup>



ArgoAddonSpec defines the desired state of ArgoAddon.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argoaddonspecargo">argo</a></b></td>
        <td>object</td>
        <td>
          Argo configuration<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>decouple</b></td>
        <td>boolean</td>
        <td>
          When decouple is enabled, appprojects are preserved even in the case when the origin tenant is deleted.
This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>force</b></td>
        <td>boolean</td>
        <td>
          When force is enabled, appprojects which already exist with the same name as a tenant will be adopted
and overwritten. When disabled the appprojects will not be changed or adopted.
This is true for any other resource as well. This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readonly</b></td>
        <td>boolean</td>
        <td>
          All appprojects, which are collected by this controller, are set into ready-only mode
That means only properties from matching translators are respected. Any changes from users are
overwritten. This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoAddon.spec.argo
<sup><sup>[↩ Parent](#argoaddonspec)</sup></sup>



Argo configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>serviceAccountNamespace</b></td>
        <td>string</td>
        <td>
          Default Namespace to create ServiceAccounts used by arog-cd
The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
It's best to have a dedicated namespace for these serviceaccounts<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>destination</b></td>
        <td>string</td>
        <td>
          If you are not using the capsule-proxy integration this destination is registered
for each appproject.<br/>
          <br/>
            <i>Default</i>: https://kubernetes.default.svc<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>destinationServiceAccounts</b></td>
        <td>boolean</td>
        <td>
          This is a feature which will be released with argocd +v2.13.0
If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where the ArgoCD instance is running<br/>
          <br/>
            <i>Default</i>: argocd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>rbacConfigMap</b></td>
        <td>string</td>
        <td>
          Name of the ArgoCD rbac configmap (required for the controller)<br/>
          <br/>
            <i>Default</i>: argocd-rbac-cm<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoAddon.status
<sup><sup>[↩ Parent](#argoaddon)</sup></sup>



ArgoAddonStatus defines the observed state of ArgoAddon.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argoaddonstatusloaded">loaded</a></b></td>
        <td>object</td>
        <td>
          Last applied valid configuration<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoAddon.status.loaded
<sup><sup>[↩ Parent](#argoaddonstatus)</sup></sup>



Last applied valid configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argoaddonstatusloadedargo">argo</a></b></td>
        <td>object</td>
        <td>
          Argo configuration<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>decouple</b></td>
        <td>boolean</td>
        <td>
          When decouple is enabled, appprojects are preserved even in the case when the origin tenant is deleted.
This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>force</b></td>
        <td>boolean</td>
        <td>
          When force is enabled, appprojects which already exist with the same name as a tenant will be adopted
and overwritten. When disabled the appprojects will not be changed or adopted.
This is true for any other resource as well. This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readonly</b></td>
        <td>boolean</td>
        <td>
          All appprojects, which are collected by this controller, are set into ready-only mode
That means only properties from matching translators are respected. Any changes from users are
overwritten. This can also be set on a per-tenant basis via annotations.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoAddon.status.loaded.argo
<sup><sup>[↩ Parent](#argoaddonstatusloaded)</sup></sup>



Argo configuration

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>serviceAccountNamespace</b></td>
        <td>string</td>
        <td>
          Default Namespace to create ServiceAccounts used by arog-cd
The namespace must be part of capsuleUsers and have "list", "get" and "watch" privileges for the entire cluster
It's best to have a dedicated namespace for these serviceaccounts<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>destination</b></td>
        <td>string</td>
        <td>
          If you are not using the capsule-proxy integration this destination is registered
for each appproject.<br/>
          <br/>
            <i>Default</i>: https://kubernetes.default.svc<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>destinationServiceAccounts</b></td>
        <td>boolean</td>
        <td>
          This is a feature which will be released with argocd +v2.13.0
If you are not yet on that version, you can't use this feature. Currently Feature is in state Alpha<br/>
          <br/>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace where the ArgoCD instance is running<br/>
          <br/>
            <i>Default</i>: argocd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>rbacConfigMap</b></td>
        <td>string</td>
        <td>
          Name of the ArgoCD rbac configmap (required for the controller)<br/>
          <br/>
            <i>Default</i>: argocd-rbac-cm<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## ArgoTranslator
<sup><sup>[↩ Parent](#addonsprojectcapsuledevv1alpha1 )</sup></sup>






ArgoTranslator is the Schema for the argotranslators API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>addons.projectcapsule.dev/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>ArgoTranslator</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspec">spec</a></b></td>
        <td>object</td>
        <td>
          ArgoTranslatorSpec defines the desired state of ArgoTranslator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatus">status</a></b></td>
        <td>object</td>
        <td>
          ArgoTranslatorStatus defines the observed state of ArgoTranslator.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec
<sup><sup>[↩ Parent](#argotranslator)</sup></sup>



ArgoTranslatorSpec defines the desired state of ArgoTranslator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>customPolicy</b></td>
        <td>string</td>
        <td>
          In this field you can define custom policies. It must result in a valid argocd policy format (CSV)
You can use Sprig Templating with this field<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecrolesindex">roles</a></b></td>
        <td>[]object</td>
        <td>
          Application-Project Roles for the tenant<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          Selector to match tenants which are used for the translator<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettings">settings</a></b></td>
        <td>object</td>
        <td>
          Additional settings for the argocd project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.roles[index]
<sup><sup>[↩ Parent](#argotranslatorspec)</sup></sup>



Define Permission mappings for an ArogCD Project.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>clusterRoles</b></td>
        <td>[]string</td>
        <td>
          TenantRoles selects tenant users based on their cluster roles to this Permission<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name for permission mapping<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>owner</b></td>
        <td>boolean</td>
        <td>
          Define if the selected users are owners of the appproject. Being owner allows the users
to update the project and effectively manage everything. By default the selected users get
read-only access to the project.<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecrolesindexpoliciesindex">policies</a></b></td>
        <td>[]object</td>
        <td>
          Roles are reflected in the argocd rbac configmap<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.roles[index].policies[index]
<sup><sup>[↩ Parent](#argotranslatorspecrolesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>[]string</td>
        <td>
          Allowed actions for this permission. You may specify multiple actions. To allow all actions use "*"<br/>
          <br/>
            <i>Default</i>: [get]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          You may specify a custom path for the resource. The available path for argo is <app-project>/<app-ns>/<app-name>
however <app-project> is already set to the argocd project name. Therefor you can only add <app-ns>/<app-name><br/>
          <br/>
            <i>Default</i>: *<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          Name for permission mapping<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>verb</b></td>
        <td>string</td>
        <td>
          Verb for this permission (can be allow, deny)<br/>
          <br/>
            <i>Default</i>: allow<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.selector
<sup><sup>[↩ Parent](#argotranslatorspec)</sup></sup>



Selector to match tenants which are used for the translator

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorspecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#argotranslatorspecselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings
<sup><sup>[↩ Parent](#argotranslatorspec)</sup></sup>



Additional settings for the argocd project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorspecsettingsstructured">structured</a></b></td>
        <td>object</td>
        <td>
          Structured Properties for the argocd project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Use a template to generate to argo project settings<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured
<sup><sup>[↩ Parent](#argotranslatorspecsettings)</sup></sup>



Structured Properties for the argocd project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredmeta">meta</a></b></td>
        <td>object</td>
        <td>
          Project Metadata<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspec">spec</a></b></td>
        <td>object</td>
        <td>
          Application Project Spec (Upstream ArgoCD)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.meta
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructured)</sup></sup>



Project Metadata

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations for the project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>finalizers</b></td>
        <td>[]string</td>
        <td>
          Finalizers for the project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels for the project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructured)</sup></sup>



Application Project Spec (Upstream ArgoCD)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecclusterresourceblacklistindex">clusterResourceBlacklist</a></b></td>
        <td>[]object</td>
        <td>
          ClusterResourceBlacklist contains list of blacklisted cluster level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecclusterresourcewhitelistindex">clusterResourceWhitelist</a></b></td>
        <td>[]object</td>
        <td>
          ClusterResourceWhitelist contains list of whitelisted cluster level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description contains optional project description<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecdestinationserviceaccountsindex">destinationServiceAccounts</a></b></td>
        <td>[]object</td>
        <td>
          DestinationServiceAccounts holds information about the service accounts to be impersonated for the application sync operation for each destination.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecdestinationsindex">destinations</a></b></td>
        <td>[]object</td>
        <td>
          Destinations contains list of destinations available for deployment<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecnamespaceresourceblacklistindex">namespaceResourceBlacklist</a></b></td>
        <td>[]object</td>
        <td>
          NamespaceResourceBlacklist contains list of blacklisted namespace level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecnamespaceresourcewhitelistindex">namespaceResourceWhitelist</a></b></td>
        <td>[]object</td>
        <td>
          NamespaceResourceWhitelist contains list of whitelisted namespace level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecorphanedresources">orphanedResources</a></b></td>
        <td>object</td>
        <td>
          OrphanedResources specifies if controller should monitor orphaned resources of apps in this project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>permitOnlyProjectScopedClusters</b></td>
        <td>boolean</td>
        <td>
          PermitOnlyProjectScopedClusters determines whether destinations can only reference clusters which are project-scoped<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecrolesindex">roles</a></b></td>
        <td>[]object</td>
        <td>
          Roles are user defined RBAC roles associated with this project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecsignaturekeysindex">signatureKeys</a></b></td>
        <td>[]object</td>
        <td>
          SignatureKeys contains a list of PGP key IDs that commits in Git must be signed with in order to be allowed for sync<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sourceNamespaces</b></td>
        <td>[]string</td>
        <td>
          SourceNamespaces defines the namespaces application resources are allowed to be created in<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sourceRepos</b></td>
        <td>[]string</td>
        <td>
          SourceRepos contains list of repository URLs which can be used for deployment<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecsyncwindowsindex">syncWindows</a></b></td>
        <td>[]object</td>
        <td>
          SyncWindows controls when syncs can be run for apps in this project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.clusterResourceBlacklist[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.clusterResourceWhitelist[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.destinationServiceAccounts[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



ApplicationDestinationServiceAccount holds information about the service account to be impersonated for the application sync operation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultServiceAccount</b></td>
        <td>string</td>
        <td>
          DefaultServiceAccount to be used for impersonation during the sync operation<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          Server specifies the URL of the target cluster's Kubernetes control plane API.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace specifies the target namespace for the application's resources.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.destinations[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



ApplicationDestination holds information about the application's destination

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is an alternate way of specifying the target cluster by its symbolic name. This must be set if Server is not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace specifies the target namespace for the application's resources.
The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          Server specifies the URL of the target cluster's Kubernetes control plane API. This must be set if Name is not set.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.namespaceResourceBlacklist[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.namespaceResourceWhitelist[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.orphanedResources
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



OrphanedResources specifies if controller should monitor orphaned resources of apps in this project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecorphanedresourcesignoreindex">ignore</a></b></td>
        <td>[]object</td>
        <td>
          Ignore contains a list of resources that are to be excluded from orphaned resources monitoring<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>warn</b></td>
        <td>boolean</td>
        <td>
          Warn indicates if warning condition should be created for apps which have orphaned resources<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.orphanedResources.ignore[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspecorphanedresources)</sup></sup>



OrphanedResourceKey is a reference to a resource to be ignored from

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.roles[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



ProjectRole represents a role that has access to a project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a name for this role<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description is a description of the role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>groups</b></td>
        <td>[]string</td>
        <td>
          Groups are a list of OIDC group claims bound to this role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorspecsettingsstructuredspecrolesindexjwttokensindex">jwtTokens</a></b></td>
        <td>[]object</td>
        <td>
          JWTTokens are a list of generated JWT tokens bound to this role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>policies</b></td>
        <td>[]string</td>
        <td>
          Policies Stores a list of casbin formatted strings that define access policies for the role in the project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.roles[index].jwtTokens[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspecrolesindex)</sup></sup>



JWTToken holds the issuedAt and expiresAt values of a token

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>iat</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>exp</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.signatureKeys[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



SignatureKey is the specification of a key required to verify commit signatures with

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>keyID</b></td>
        <td>string</td>
        <td>
          The ID of the key in hexadecimal notation<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.spec.settings.structured.spec.syncWindows[index]
<sup><sup>[↩ Parent](#argotranslatorspecsettingsstructuredspec)</sup></sup>



SyncWindow contains the kind, time, duration and attributes that are used to assign the syncWindows to apps

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>applications</b></td>
        <td>[]string</td>
        <td>
          Applications contains a list of applications that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusters</b></td>
        <td>[]string</td>
        <td>
          Clusters contains a list of clusters that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>duration</b></td>
        <td>string</td>
        <td>
          Duration is the amount of time the sync window will be open<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind defines if the window allows or blocks syncs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>manualSync</b></td>
        <td>boolean</td>
        <td>
          ManualSync enables manual syncs when they would otherwise be blocked<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          Namespaces contains a list of namespaces that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          Schedule is the time the window will begin, specified in cron format<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeZone</b></td>
        <td>string</td>
        <td>
          TimeZone of the sync that will be applied to the schedule<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status
<sup><sup>[↩ Parent](#argotranslator)</sup></sup>



ArgoTranslatorStatus defines the observed state of ArgoTranslator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>ready</b></td>
        <td>string</td>
        <td>
          Ready field indicating overall readiness of the translator<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>size</b></td>
        <td>integer</td>
        <td>
          Amount of tenants selected by this translator<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindex">tenants</a></b></td>
        <td>[]object</td>
        <td>
          List of tenants selected by this translator<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index]
<sup><sup>[↩ Parent](#argotranslatorstatus)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorstatustenantsindexcondition">condition</a></b></td>
        <td>object</td>
        <td>
          Conditions represent the latest available observations of an object's state<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          List of tenants selected by this translator<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexserving">serving</a></b></td>
        <td>object</td>
        <td>
          Serving  Settings for this Tenant<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the tracked Tenant to pin point tracking<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].condition
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindex)</sup></sup>



Conditions represent the latest available observations of an object's state

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindex)</sup></sup>



Serving  Settings for this Tenant

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructured">structured</a></b></td>
        <td>object</td>
        <td>
          Structured Properties for the argocd project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Use a template to generate to argo project settings<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexserving)</sup></sup>



Structured Properties for the argocd project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredmeta">meta</a></b></td>
        <td>object</td>
        <td>
          Project Metadata<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspec">spec</a></b></td>
        <td>object</td>
        <td>
          Application Project Spec (Upstream ArgoCD)<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.meta
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructured)</sup></sup>



Project Metadata

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          Annotations for the project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>finalizers</b></td>
        <td>[]string</td>
        <td>
          Finalizers for the project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          Labels for the project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructured)</sup></sup>



Application Project Spec (Upstream ArgoCD)

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecclusterresourceblacklistindex">clusterResourceBlacklist</a></b></td>
        <td>[]object</td>
        <td>
          ClusterResourceBlacklist contains list of blacklisted cluster level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecclusterresourcewhitelistindex">clusterResourceWhitelist</a></b></td>
        <td>[]object</td>
        <td>
          ClusterResourceWhitelist contains list of whitelisted cluster level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description contains optional project description<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecdestinationserviceaccountsindex">destinationServiceAccounts</a></b></td>
        <td>[]object</td>
        <td>
          DestinationServiceAccounts holds information about the service accounts to be impersonated for the application sync operation for each destination.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecdestinationsindex">destinations</a></b></td>
        <td>[]object</td>
        <td>
          Destinations contains list of destinations available for deployment<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecnamespaceresourceblacklistindex">namespaceResourceBlacklist</a></b></td>
        <td>[]object</td>
        <td>
          NamespaceResourceBlacklist contains list of blacklisted namespace level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecnamespaceresourcewhitelistindex">namespaceResourceWhitelist</a></b></td>
        <td>[]object</td>
        <td>
          NamespaceResourceWhitelist contains list of whitelisted namespace level resources<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecorphanedresources">orphanedResources</a></b></td>
        <td>object</td>
        <td>
          OrphanedResources specifies if controller should monitor orphaned resources of apps in this project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>permitOnlyProjectScopedClusters</b></td>
        <td>boolean</td>
        <td>
          PermitOnlyProjectScopedClusters determines whether destinations can only reference clusters which are project-scoped<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecrolesindex">roles</a></b></td>
        <td>[]object</td>
        <td>
          Roles are user defined RBAC roles associated with this project<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecsignaturekeysindex">signatureKeys</a></b></td>
        <td>[]object</td>
        <td>
          SignatureKeys contains a list of PGP key IDs that commits in Git must be signed with in order to be allowed for sync<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sourceNamespaces</b></td>
        <td>[]string</td>
        <td>
          SourceNamespaces defines the namespaces application resources are allowed to be created in<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sourceRepos</b></td>
        <td>[]string</td>
        <td>
          SourceRepos contains list of repository URLs which can be used for deployment<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecsyncwindowsindex">syncWindows</a></b></td>
        <td>[]object</td>
        <td>
          SyncWindows controls when syncs can be run for apps in this project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.clusterResourceBlacklist[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.clusterResourceWhitelist[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.destinationServiceAccounts[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



ApplicationDestinationServiceAccount holds information about the service account to be impersonated for the application sync operation.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultServiceAccount</b></td>
        <td>string</td>
        <td>
          DefaultServiceAccount to be used for impersonation during the sync operation<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          Server specifies the URL of the target cluster's Kubernetes control plane API.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace specifies the target namespace for the application's resources.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.destinations[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



ApplicationDestination holds information about the application's destination

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is an alternate way of specifying the target cluster by its symbolic name. This must be set if Server is not set.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace specifies the target namespace for the application's resources.
The namespace will only be set for namespace-scoped resources that have not set a value for .metadata.namespace<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          Server specifies the URL of the target cluster's Kubernetes control plane API. This must be set if Name is not set.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.namespaceResourceBlacklist[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.namespaceResourceWhitelist[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



GroupKind specifies a Group and a Kind, but does not force a version.  This is useful for identifying
concepts during lookup stages without having partially valid types

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.orphanedResources
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



OrphanedResources specifies if controller should monitor orphaned resources of apps in this project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecorphanedresourcesignoreindex">ignore</a></b></td>
        <td>[]object</td>
        <td>
          Ignore contains a list of resources that are to be excluded from orphaned resources monitoring<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>warn</b></td>
        <td>boolean</td>
        <td>
          Warn indicates if warning condition should be created for apps which have orphaned resources<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.orphanedResources.ignore[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspecorphanedresources)</sup></sup>



OrphanedResourceKey is a reference to a resource to be ignored from

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.roles[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



ProjectRole represents a role that has access to a project

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a name for this role<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description is a description of the role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>groups</b></td>
        <td>[]string</td>
        <td>
          Groups are a list of OIDC group claims bound to this role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#argotranslatorstatustenantsindexservingstructuredspecrolesindexjwttokensindex">jwtTokens</a></b></td>
        <td>[]object</td>
        <td>
          JWTTokens are a list of generated JWT tokens bound to this role<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>policies</b></td>
        <td>[]string</td>
        <td>
          Policies Stores a list of casbin formatted strings that define access policies for the role in the project<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.roles[index].jwtTokens[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspecrolesindex)</sup></sup>



JWTToken holds the issuedAt and expiresAt values of a token

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>iat</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>exp</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.signatureKeys[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



SignatureKey is the specification of a key required to verify commit signatures with

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>keyID</b></td>
        <td>string</td>
        <td>
          The ID of the key in hexadecimal notation<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### ArgoTranslator.status.tenants[index].serving.structured.spec.syncWindows[index]
<sup><sup>[↩ Parent](#argotranslatorstatustenantsindexservingstructuredspec)</sup></sup>



SyncWindow contains the kind, time, duration and attributes that are used to assign the syncWindows to apps

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>applications</b></td>
        <td>[]string</td>
        <td>
          Applications contains a list of applications that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusters</b></td>
        <td>[]string</td>
        <td>
          Clusters contains a list of clusters that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>duration</b></td>
        <td>string</td>
        <td>
          Duration is the amount of time the sync window will be open<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind defines if the window allows or blocks syncs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>manualSync</b></td>
        <td>boolean</td>
        <td>
          ManualSync enables manual syncs when they would otherwise be blocked<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          Namespaces contains a list of namespaces that the window will apply to<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schedule</b></td>
        <td>string</td>
        <td>
          Schedule is the time the window will begin, specified in cron format<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeZone</b></td>
        <td>string</td>
        <td>
          TimeZone of the sync that will be applied to the schedule<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
