---
title: "👋 Examples"
summary: "Different Examples on how to use the Addon (Translators)"
weight: 1
ShowToc: true
TocOpen: true
---

# Quickstart Example

```yaml
apiVersion: ArgoTranslator/v1alpha1
kind: ArgoTranslator
metadata:
    name: default-onboarding
spec:
  selector:
    matchLabels:
      app.kubernetes.io/type: dev
  settings:
    namespaceResourceWhitelist:
    - group: "*"
      kind: "*"      
    clusterResourceWhitelist:
    - group: "*"
      kind: "*"
  roles:
  # This creates the baseline role. All users with the cluster-role "admin" will be able to access all repositories
  - name: "baseline"
    clusterRoles:
      - "tenant-viewer"
    policies:
    - resource: "*"
      action: ["get"]
  - name: "admin"
    # Selects entities which are mapped to the clusterRole "admin
    clusterRoles:
      - "admin"
    # Allows the users to make changes to the appproject (just update verb)
    owner: true
    # Additional policies for the mapped entities. Allows to interact with
    policies:  
    - resource: applications
      action: ["*"]
    - resource: repositories
      action: ["*"]
```


This addon is designed for kubernetes administrators, to automatically translate their existing Capsule Tenants into Argo Appprojects. This addon adds new capabilities to the Capsule project, by allowing the administrator to create a new tenant in Capsule, and automatically create a new Argo Appproject for that tenant. This addon is designed to be used in conjunction with the Capsule project, and is not intended to be used as a standalone project.

We have choosen a very loose implementation which makes use of so called "Translators" to translate the Capsule Tenant into an Argo Appproject. This allows us to easily add new translators for different use cases and translate permissions from the Capsule Tenant into the Argo Appproject.





By design the Addon is designed to work with the capsule-proxy. Meaning each Approject gets it's own declarative and project scoped cluster. Which is finally a dedicated ServiceAccount, who is promoted as owner on the corresponding tenant.

The project's goal is to create a very generic experience for administrators. We know of different companies which already have implemented such an operator for argocd. This is our attempt to reconcile all development into one project.


