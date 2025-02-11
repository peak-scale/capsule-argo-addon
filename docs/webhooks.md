# Webhooks

There are mutatingwebhooks available for both `Applications` and `ApplicationSets`. You can enable them in the helm chart via:

```
webhooks:
  enabled: true
```

They are disabled by default. These webhooks patch the `project` property for both resources to the corresponding tenant, in which they were deployed. Very useful if your users can create Applications and Appsets in any namespace.

This requires `cert-manager` to be installed on the cluster

## Argo Settings


**params**
```yaml
"applicationsetcontroller.allowed.scm.providers": "https://github.com/,https://gitlab.com/"
"application.namespaces": "*"
"applicationsetcontroller.namespaces": "*"
```

**cm**
```yaml
"application.resourceTrackingMethod": "annotation+label"
```
