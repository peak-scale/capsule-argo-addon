# Capsule ❤️ Argo

![Argo Capsule Addon](docs/images/capsule-argo.png)

<p align="center">
<a href="https://github.com/peak-scale/capsule-argo-addon/releases/latest">
  <img alt="GitHub release (latest SemVer)" src="https://img.shields.io/github/v/release/peak-scale/capsule-argo-addon?sort=semver">
</a>
<a href="https://artifacthub.io/packages/search?repo=capsule-argo-addon">
  <img src="https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/capsule-argo-addon" alt="Artifact Hub">
</a>
<a href="https://app.fossa.com/projects/git%2Bgithub.com%2Fpeak-scale%2Fcapsule-argo-addon?ref=badge_small" alt="FOSSA Status"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpeak-scale%2Fcapsule-argo-addon.svg?type=small"/></a>
<a href="https://codecov.io/gh/peak-scale/capsule-argo-addon">
  <img src="https://codecov.io/gh/peak-scale/capsule-argo-addon/graph/badge.svg?token=26QLMNSN54" alt="codecov">
</a>
</p>


This addon is designed for kubernetes administrators, to automatically translate their existing Capsule Tenants into Argo Appprojects. This addon adds new capabilities to the Capsule project, by allowing the administrator to create a new tenant in Capsule, and automatically create a new Argo Appproject for that tenant. This addon is designed to be used in conjunction with the Capsule project, and is not intended to be used as a standalone project.

We have chosen a very loose implementation which makes use of so called [Translators](docs/translators.md) to translate the Capsule Tenant into an Argo Appproject. This allows us to easily add new translators for different use cases and translate permissions from the Capsule Tenant into the Argo Appproject.

![Argo Capsule Addon Overview](docs/images/capsule-argo-addon.gif)

By design the Addon is designed to work by using [Impersonation provided by Argo](https://argo-cd.readthedocs.io/en/latest/operator-manual/app-sync-using-impersonation/). The solution with [capsule-proxy](https://github.com/projectcapsule/capsule-proxy) is no longer supported, because it had large performance implications.

The project's goal is to create a very generic experience for administrators. We know of different companies which already have implemented such an operator for argocd. This is our attempt to reconcile all development into one project.

## Documentation

See the [Documentation](docs/README.md) for more information on how to use this addon.

## Demo

Spin up a live demonstration of the addon on Killercoda:

- [https://killercoda.com/peakscale/course/solutions/multi-tenant-argo](https://killercoda.com/peakscale/course/solutions/multi-tenant-argo)

## Support

This addon is developed by the community. For enterprise support (production ready setup,tailor-made features) reach out to [Peak Scale](https://peakscale.ch/en/)


## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.



[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fpeak-scale%2Fcapsule-argo-addon.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fpeak-scale%2Fcapsule-argo-addon?ref=badge_large)
