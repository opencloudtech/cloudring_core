# cert-manager HA controller source

This directory provides a provider-neutral Flux source contract for cert-manager
on upstream Kubernetes. It installs the upstream Jetstack chart with three
controller, webhook, and CA injector replicas, host-level topology spreading,
PodDisruptionBudgets, retained CRDs, and digest-pinned runtime images.

The base `HelmRelease` is intentionally suspended. Rendering or structurally
validating this source does not prove that a cluster has sufficient capacity,
that the Kubernetes API server can reach the webhook, or that issuance and
renewal work. A downstream operator must unsuspend the release only after
checking those conditions and establishing a rollback window.

The source installs no `Issuer` or `ClusterIssuer`. Issuer policy, DNS or ACME
configuration, credentials, endpoints, and live evidence belong to the
downstream deployment.

## Pinned upstream inputs

- Jetstack OCI chart `oci://quay.io/jetstack/charts/cert-manager` version
  `v1.21.0`, reconciled by Flux from manifest digest
  `sha256:cd55fea42658e54abc25e85a0bc1de229925a5006445f916bfd2c6dc80ac3613`.
  Its reviewed chart content and provenance layer digests are recorded in
  [`../runtime-chart-supply-chain.json`](../runtime-chart-supply-chain.json).
- Controller, webhook, CA injector, startup API check, and ACME solver images
  use the `v1.21.0` tag together with the upstream multi-architecture manifest
  digest recorded in the Helm values.
- Helm owns the CRDs with `CreateReplace` on install and upgrade, while the
  chart's keep policy protects custom resources from an accidental uninstall.

The OCI manifest digest, not a mutable Helm repository index or annotation, is
the reconciled source of truth. The digests were resolved from the official
Quay distribution manifest. A version bump must refresh every pin, the shared
supply-chain receipt, and the strict verifier in the same change.

## Local validation

```bash
kubectl kustomize deploy/kubernetes/cert-manager/controllers >/dev/null
go test ./internal/platformmanifest -run CertManager -count=1
go run ./cmd/cloudring-manifestcheck --root .
```

Before activation, a downstream runbook must verify at least three schedulable
failure domains, webhook reachability from the API server, all three deployments
Available, all three PodDisruptionBudgets healthy, CRDs Established, a synthetic
issuance and renewal, and recovery after one controller node is unavailable.
Those checks are live deployment evidence and are not claimed by this source.
