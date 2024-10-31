# Annotations

You can define annotations on tenant basis to influence the behavior for tenants (opt-out). The following annotations are supported:

## `argo.addons.projectcapsule.dev/name`

By default the appproject's name is the same as the tenant name. If you want to change the appproject name, you can set the `argo.addons.projectcapsule.dev/name` annotation to the desired name.

## `argo.addons.projectcapsule.dev/force`

For this tenant overwrite any other resources which may already be present. If resources are already present whey won't be overwritten until this is specified for the affected tenant or for all tenants via [configuration](config.md). This is `false` by default.

## `argo.addons.projectcapsule.dev/service-account-namespace`

By default the service account used for the capsule proxy is created in the same namespace as the appproject. If you want to change the namespace, you can set the `argo.addons.projectcapsule.dev/service-account-namespace` annotation to the desired namespace. This is useful if you have different tenants with different privileges on the cluster. Since you can [bind service accounts as capsule-users only via group (namespace)](https://projectcapsule.dev/docs/tenants/permissions/#group-scope) you can use different namespaces to seperate different ServiceAccount privileges.

> This is only relevant if the proxy or registration is enabled

## `argo.addons.projectcapsule.dev/register-proxy`

By default, if the proxy integration is enabled, for each tenant a dedicated argo cluster is registered, which uses a dedicated service account via the capsule-proxy. If you want to disable the registration of the proxy, you can set the `argo.addons.projectcapsule.dev/register-proxy` annotation to `false`.

## `argo.addons.projectcapsule.dev/read-only`

By default, if a [subject] is promoted as [appproject owner] they can update project properties like adding [SyncWIndows](https://argo-cd.readthedocs.io/en/stable/user-guide/sync_windows/) or [Roles](https://argo-cd.readthedocs.io/en/stable/user-guide/projects/#project-roles).

If you want to prevent this behavior, you can set the `argo.addons.projectcapsule.dev/read-only` annotation to `true`. This overwrites any changes not made by [translators](./translators.md).

## `argo.addons.projectcapsule.dev/decouple`

By default the argo appproject is bound to the tenant (via Ownerreference). This indicates, that if a tenant is deleted, the appproject is also deleted (including all proxy assets, if enabled).

If you want to decouple the appproject from the tenant, you can set the `argo.addons.projectcapsule.dev/decouple` annotation to `true`. This will prevent the deletion of the appproject if the tenant is deleted.
