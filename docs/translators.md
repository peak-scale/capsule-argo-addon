
# Translators

Translators are client objects to translate your [Capsule Tenants](https://projectcapsule.dev/docs/tenants/) to argocd [Application Projects (Appprojects)](https://argo-cd.readthedocs.io/en/stable/user-guide/projects/). You can have multiple Translators. But all toegether have the combined purpose to translate one Capsule Tenant into one Argo Project.

To translate permissions the Operator looks at Capsule Tenant with ther [Tenant Owners](https://projectcapsule.dev/docs/tenants/permissions/#ownership) and [AdditionalRoleBindings](https://projectcapsule.dev/docs/tenants/permissions/#additional-rolebindings). Based on these specs it's evaluated which [Subject (User/Group/ServiceAccount)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-subjects) is bound to which [ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole). Based on these ClusterRoles you can then translate [Argo RBAC Policies](https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/#rbac-model-structure) which are then bound to the selected Subjects.

## Configuration

See the following Topics for insights for the configuration of Translators. [View the API Spec](./reference.md)

### Tenant Selection

Each Translator selects it's tenants via [Selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/). Setting no selector results in no tenants being selected (not all tenants).

A tenant must be selected by at least one Translator, to create an Argo Project. If a tenant is not selected by any Translator, the operator will not consider it. 

**Note**: If a tenant gets unselected from Translators into a state where it's no longer selected by any Translator, it will be garbage collected. Meaning the Appproject and other assets will be deleted. This behavior can be influenced with [per-tenant Annotations](./annotations.md).

Simple Example to select specific Tenants:

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoTranslator
metadata:
    name: example
spec:
  # Match Tenants with the label "app.kubernetes.io/type" and  the value "dev" or "prod"
  selector:
    matchExpressions:
      - key: app.kubernetes.io/type
        operator: In
        values:
          - dev
          - prod
```

### Roles Translation

To translate permissions the Operator looks at Capsule Tenant with ther [Tenant Owners](https://projectcapsule.dev/docs/tenants/permissions/#ownership) and [AdditionalRoleBindings](https://projectcapsule.dev/docs/tenants/permissions/#additional-rolebindings). Based on these specs it's evaluated which [Subject (User/Group/ServiceAccount)](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#referring-to-subjects) is bound to which [ClusterRoles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole). Based on these ClusterRoles you can then translate [Argo RBAC Policies](https://argo-cd.readthedocs.io/en/stable/operator-manual/rbac/#rbac-model-structure) which are then bound to the selected Subjects.

Let's first take a like at a simple Role Translation:

```yaml
# the name is relevant in the csv, you will see all the policies with role:{tenant}:viewer
name: "viewer"

# All users which have the clusterRole "tenant-viewer" assigned (either owner or additionalRoleBindings) will be assigned the policies below. 
clusterRoles:
  - "tenant-viewer"

# Translated to argo permissions these users/groups have the privilege to (action "*")
# on (resource "application)
policies:
- resource: "applications"
  action: ["*"]

# Assigns the selected subjects the default policy "owner". Otherwise just "read-only"
owner: false
```

Let's use a fictive tenant, so we can understand how the Roles are Translated to Argo. We are using this Tenant:

```yaml
---
apiVersion: capsule.clastix.io/v1beta2
kind: Tenant
metadata:
  name: solar
  labels:
    app.kubernetes.io/type: dev
spec:
  owners:
  - name: solar-users
    kind: Group
    # This are added by Capsule, unless specified
    clusterRoles:
       - admin
       - capsule-namespace-deleter
  - name: alice
    kind: User
    # This are added by Capsule, unless specified
    clusterRoles:
       - admin
       - capsule-namespace-deleter
  additionalRoleBindings:
  - clusterRoleName: tenant-viewer
    subjects:
    - kind: User
      name: bob
  - clusterRoleName: operators
    subjects:
    - kind: Group
      name: org:operators
```

This is the Translator we are going to use:

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoTranslator
metadata:
    name: example-permissions
spec:
  selector:
    matchExpressions:
      - key: app.kubernetes.io/type
        operator: In
        values:
          - dev
          - prod

  # Role Translation happens here
  roles:

    # the name is relevant in the csv, you will see all the policies with role:{tenant}:viewer
  - name: "viewer"

    # All users which have the clusterRole "tenant-viewer" assigned (either owner or additionalRoleBindings) will be assigned the policies below. 
    clusterRoles:
      - "tenant-viewer"

    # Translated to argo permissions these users/groups have the privilege to (action "*")
    # on (resource "application)
    policies:
    - resource: "applications"
      action: ["*"]

    # the name is relevant in the csv, you will see all the policies with role:{tenant}:viewer
  - name: "owner"

    # All users which have the clusterRole "admin" assigned (either owner or additionalRoleBindings) will be assigned the policies below. (These are by default all tenant owners)
    clusterRoles:
      - "admin"

    # Assigning owner to all the subjects based on the clusterRoles. This means they have the possability to make changes to the appproject. However the specs from the translators are still considered. This Option is powerful when your argo users want to manage their own Roles or SyncWindows etc.
    owner: true

    # Policies for argo which are bound to all subjects with the clusterRole "admin".
    policies:  
    - resource: applications
      action: 
      - "action//Pod/maintenance-off"
      - "get"
      - "sync"
    - resource: repositories
      action: ["*"]
```

#### Default policies

These policies are bootstrapped for every tenant by the controller. All policies provisioned by the controller are prefixed with `caa`.

##### Read-Only

This policy allows the users to see the project (kind of necessary). Therefor it will be assigned to any subject which is referenced in any of the `ClusterRoles`:

```csv
p, caa:role:wind:read-only,projects,get,<appproject-name>,allow
p, caa:role:wind:read-only,clusters,get,<appproject-name>/*,allow
```

##### Owner

This policy allows users to make changes to the appproject (if you have some sort of shared responsibility). To assign this role the attribute `owner` must be set to `true`.

```csv
p, caa:role:wind:owner,projects,update,<appproject-name>,allow
```

### Project Settings

Often you have your own set of Argo Project-Settings, which you would like to pass over to the tenants. This is also possible with translators. You can [view here](https://argo-cd.readthedocs.io/en/stable/user-guide/projects/) to see all the possible fields for appprojects or explain it for your kubernetes cluster:

```shell
kubectl explain appproject.spec
```

What's important

- A Translator only manages the appproject specification itself defines. That means if a translator is deleted, it removes it's part from all relevant appprojects
- Multiple translators having project settings are merged together
- By default Users with `Owner` privileges can edit appproject settings. They are merged with all the translator specifications.
- If multiple translator match, Non-Slice fields are overwritten, there's not yet a concrete priority implemented. 

#### Structured

Structured gives you the possibility to configure appproject specification and additional metadata.

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoTranslator
metadata:
    name: default-onboarding
spec:
  selector:
    matchExpressions:
      - key: app.kubernetes.io/type
        operator: In
        values:
          - dev
          - prod

  settings:
    structured:
      meta:
        labels:
          "company.infra.team/managed-by": "platform"
        finalizers:
        - "custom-company/finalizer"
      spec:
        permitOnlyProjectScopedClusters: false
        clusterResourceWhitelist:
          - group: "*"
            kind: "*"
        namespaceResourceWhitelist:
          - group: "*"
            kind: "*"
```

Explain the API:

```shell
kubectl explain argotranslator.spec.settings.structured
```

#### Templated

You can use [Sprig Template](./templating.md) to express the project settings. Templated Specs are **merged over** [structured](#structured) fields.

```yaml
---
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: ArgoTranslator
metadata:
    name: example-template
spec:
  selector:
    matchExpressions:
      - key: app.kubernetes.io/type
        operator: In
        values:
          - dev
          - prod
  settings:
    template: |
      meta:
        labels:
          here-go-extra-label: "meow"
      spec:
        description: "Managed Argo Project (Replicated {{ $.Tenant.Name }}"
        {{- if $.Tenant.Object.Spec.cordoned }}
        syncWindows:
        - applications:
          - '*'
          duration: 1m
          kind: deny
          schedule: '* * * * *'
          timeZone: UTC
        {{- end }}

        {{/* Allow Argo namespace, because the ui can just create in that namespace */}}
        sourceNamespaces:
        - {{ $.Config.Argo.Namespace | quote }}

        {{/* Allow all other namespaces which are currently managed by the tenant */}}
        {{- range $_, $value := $.Tenant.Namespaces }}
        - {{ $value | quote }}
        {{- end }}
```

## Examples

See the [Examples](./examples) to get a better understanding of how the CR is implemented.

## Status

With the status you have a quick summary of the healthy of an argotranslator:

```shell
$ kubectl get argotranslators -A
NAME                 AGE   TENANTS   STATUS
default-onboarding   90m   1         Ready
dev-onboarding       90m   1         Ready
```

This Status is also reflected in [metrics](./monitoring.md)

Here we can see both of the argotranslators are marked with the `Status` set to `Ready`. This means all the tenants they are translating did have any errors. If this is false, there is something wrong with at least one tenant from the translator. You can see in more detail, what each's tenant's status is:

```shell
kubectl get argotranslators default-onboarding -o yaml

...

  status:
    ready: true
    size: 1
    tenants:
    - condition:
        lastTransitionTime: "2024-10-27T14:10:37Z"
        message: Successfully translated tenant
        observedGeneration: 3
        reason: Applied
        status: "True"
        type: Ready
      name: solar-test-decouple
      uid: 5b872c4e-478d-4461-bfb7-88e6f4d4438b
```

If you have an issue in your translator (eg. template generates wrong content, or client objects which already exist) you will encounter a Failure-Condition. This might look like this:

```shell
$ kubectl get argotranslators -A

NAME                 AGE   TENANTS   STATUS
default-onboarding   81m   2         NotReady
dev-onboarding       81m   2         NotReady
```

Inspecting what's wrong with the tenant

```yaml
status:
  ready: NotReady
  size: 2
  tenants:
  - condition:
      lastTransitionTime: "2024-10-28T09:09:03Z"
      message: object argocd/solar-test-decouple of type *v1alpha1.AppProject already
        exists
      observedGeneration: 2
      reason: ObjectAlreadyExists
      status: "False"
      type: NotReady
    name: solar-test-decouple
    uid: e530f58e-4ddf-473d-acca-60ab25e2344b
  - condition:
      lastTransitionTime: "2024-10-28T09:09:03Z"
      message: Successfully translated tenant
      observedGeneration: 3
      reason: Applied
      status: "True"
      type: Ready
    name: nani
    uid: f98e373f-f6b6-4f6c-97c5-68fd21afdf4f
```

As you can see, only one tenant had a failure. The other one is successfully applied.