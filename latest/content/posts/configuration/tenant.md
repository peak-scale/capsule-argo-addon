---
title: "FAQs / How To's Guide"
summary: We'll try to answer frequently asked qestions by users.
date: 2021-01-20
tags: ["PaperMod", "Docs"]
author: ["PaperMod Contributors"]
draft: true
aliases: [/posts/papermod/papermod-how-to]
weight: 3
---

The following configurations can be done via **[annotations](#annotations)** or **[labels](#labels)** on tenant basis.

## Annotations

Add the following annotations to the `Tenant` resource to configure the tenant.

| Annotation | Description | Default |
|:------------|:-------------|:---------|
| `tenant.openshift.io/tenant-type` | Type of the tenant. | `user` |
