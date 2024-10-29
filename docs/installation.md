# Installation

The Installation of the addon is only supported via Helm-Chart. Any other method is not officially supported.

## Requirements

The following is expected to be installed (including their CRDs)

- [Capsule](https://artifacthub.io/packages/helm/projectcapsule/capsule)
- [Argo(CD)](https://artifacthub.io/packages/helm/argo/argo-cd)

Without these the addon won't work.

### Capsule-Proxy

The [capsule-proxy](https://artifacthub.io/packages/helm/projectcapsule/capsule-proxy) is used to allow serviceaccounts to just see what they should see within the boundaries of your tenant. It is optional to use the proxy and it can be disabled via the [configuration](./config.md).

If you plan to use the capsule-proxy, we recommend installing a dedicated capsule-proxy instance for the addon, because Argo puts a lot of pressure on the proxy.

With the [Helm Chart](#helm) a dedicated capsule-proxy is already installed (exclusive CRDs) by default. Adjust this according to your needs and your setups.

## Helm



