# Runtime chart supply chain

CloudRING reconciles Goal01 runtime charts from immutable artifacts. The exact
reviewed inputs are recorded in
[`deploy/kubernetes/runtime-chart-supply-chain.json`](../deploy/kubernetes/runtime-chart-supply-chain.json)
and enforced by `internal/platformmanifest`.

## Source contracts

- cert-manager `v1.21.0` uses the official Jetstack OCI chart and Flux
  `OCIRepository.spec.ref.digest` plus `HelmRelease.spec.chartRef`.
- CloudNativePG chart `0.29.0` / operator `1.30.0` uses the official GHCR OCI
  chart with the same digest-pinned Flux contract. The operator image remains
  digest-pinned independently of the chart.
- Barman Cloud CNPG-I plugin chart `0.7.0` / application `v0.13.0` uses the
  official GHCR OCI chart plus immutable manager and sidecar image-index
  digests. Its Helm values disable chart-generated ServiceAccount and manager
  RBAC; CloudRING declares a dedicated ServiceAccount and the exact reviewed
  upstream manager ClusterRole and binding. The manager permission remains
  cluster-wide because this upstream version has no namespace watch
  restriction. The chart's separate namespaced leader-election Role and
  RoleBinding are rendered even with `rbac.create=false` and remain an explicit
  residual. These pins and RBAC declarations do not prove a live WAL archive or
  recovery.
- Longhorn `1.12.0` has no official OCI chart. CloudRING vendors the exact
  official release archive, its Apache-2.0 license, and upstream receipt. Flux
  consumes `./charts/longhorn` from the official Git repository pinned to the
  exact reviewed release commit.

The OCI manifest digest is the reconciled identity. Content and provenance
layer digests are retained as review evidence. For Longhorn, the official
archive checksum, exact vendored file inventory, canonical tree digest,
license checksum, and upstream commit form the reviewed identity.
The Longhorn receipt also records all 13 effective runtime images using the
multi-platform registry index digest observed for the reviewed release. The
verifier requires the same complete image inventory in Helm values and rejects
mutable, missing, extra, or mismatched image references.

## Upgrade rule

For any chart upgrade, retrieve only the official upstream artifact, verify its
registry or release checksum, review the chart and license changes, refresh all
runtime image digests, regenerate the receipt, and update verifier fixtures in
the same change. Negative tests must continue to reject a changed source kind,
URL, digest, receipt, unexpected vendored file, or mutable chart selection.

For Longhorn, `longhorn-system/longhorn-charts` uses only
`spec.ref.commit: f8def0504bf3f5f26c342941c9e4532b44830ebe`. Branch, tag,
missing, or mismatched refs fail closed. This source must be reconciled before
the Longhorn HelmRelease.

These receipts and structural renders do not prove Flux reconciliation,
workload health, data durability, restore success, or one-node-loss survival.
Those remain downstream live release gates. A structurally valid PostgreSQL
profile is reported as `source-contract-ready` with live status `blocked`.
Those values are limited to source validation and leave every runtime gate open.

## Verification

```bash
go test ./internal/platformmanifest -count=1
kubectl kustomize deploy/kubernetes/cert-manager/controllers >/dev/null
kubectl kustomize deploy/kubernetes/postgresql-ha/controllers >/dev/null
kubectl kustomize deploy/kubernetes/postgresql-ha/runtime >/dev/null
kubectl kustomize deploy/kubernetes/postgresql-ha/recovery >/dev/null
kubectl kustomize deploy/kubernetes/storage/longhorn-three-node/runtime >/dev/null
helm template longhorn deploy/kubernetes/storage/longhorn-three-node/vendor/longhorn \
  --namespace longhorn-system >/dev/null
go run ./cmd/cloudring-manifestcheck --root .
```
