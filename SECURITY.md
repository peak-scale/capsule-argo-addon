# Release Artifacts

[See all the available artifacts](https://github.com/orgs/peak-scale/packages?repo_name=capsule-argo-addon)

## Verifing

To verify artifacts you need to have [cosign installed](https://github.com/sigstore/cosign#installation). This guide assumes you are using v2.x of cosign. All of the signatures are created using [keyless signing](https://docs.sigstore.dev/verifying/verify/#keyless-verification-using-openid-connect).
To verify the signature of the docker image, run the following command. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/capsule-argo-addon). The value `release_tag` is a release but without the prefix `v` (eg. `0.1.0-alpha.3`).

    VERSION=<release_tag> cosign verify ghcr.io/peak-scale/capsule-argo-addon:${VERSION} \
      --certificate-identity-regexp="https://github.com/peak-scale/capsule-argo-addon/.github/workflows/docker-publish.yml@refs/tags/*" \
      --certificate-oidc-issuer="https://token.actions.githubusercontent.com" | jq

To verify the signature of the helm image, run the following command. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/charts%2Fcapsule-argo-addon). The value `release_tag` is a release but without the prefix `v` (eg. `0.1.0-alpha.3`)

    VERSION=<release_tag> cosign verify ghcr.io/peak-scale/charts/capsule-argo-addon:${VERSION} \
      --certificate-identity-regexp="https://github.com/peak-scale/capsule-argo-addon/.github/workflows/helm-publish.yml@refs/tags/*" \
      --certificate-oidc-issuer="https://token.actions.githubusercontent.com" | jq

## Verifying Provenance

We create and attest the provenance of our builds using the [SLSA standard](https://slsa.dev/spec/v0.2/provenance) and meets the [SLSA Level 3](https://slsa.dev/spec/v0.1/levels) specification. The attested provenance may be verified using the cosign tool.

Verify the provenance of the docker image. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/capsule-argo-addon). The value `release_tag` is a release but without the prefix `v` (eg. `0.1.0-alpha.3`)

```bash
cosign verify-attestation --type slsaprovenance \
  --certificate-identity-regexp="https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/*" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  ghcr.io/peak-scale/capsule-argo-addon:<release_tag> | jq .payload -r | base64 --decode | jq
```

Verify the provenance of the helm image. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/charts%capsule-argo-addon). The value `release_tag` is a release but without the prefix `v` (eg. `0.1.0-alpha.3`)

```bash
VERSION=<release_tag> cosign verify-attestation --type slsaprovenance \
  --certificate-identity-regexp="https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_container_slsa3.yml@refs/tags/*" \
  --certificate-oidc-issuer="https://token.actions.githubusercontent.com" \
  "ghcr.io/peak-scale/charts/capsule-argo-addon:${VERSION}" | jq .payload -r | base64 --decode | jq
```

## Software Bill of Materials (SBOM)

An SBOM (Software Bill of Materials) in CycloneDX JSON format is published for each release, including pre-releases.

To inspect the SBOM of the docker image, run the following command. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/capsule-argo-addon):

    COSIGN_REPOSITORY=ghcr.io/peak-scale/capsule-argo-addon cosign download sbom ghcr.io/peak-scale/capsule-argo-addon:<release_tag>

To inspect the SBOM of the helm image, run the following command. Replace `<release_tag>` with an [available release tag](https://github.com/peak-scale/capsule-argo-addon/pkgs/container/charts%2Fcapsule-argo-addon):

    COSIGN_REPOSITORY=ghcr.io/peak-scale/capsule-argo-addon cosign download sbom ghcr.io/peak-scale/charts/capsule-argo-addon:<release_tag>
