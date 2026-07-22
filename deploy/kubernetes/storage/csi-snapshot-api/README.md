# CSI snapshot API and controller source

This directory is the single provider-neutral owner of Kubernetes CSI
snapshot API CRDs and the cluster-wide snapshot controller. Storage profiles
may publish driver-specific `VolumeSnapshotClass` objects, but must not vendor
these CRDs or deploy another snapshot controller.

The source is intentionally not composed into a live GitOps root. Rendering or
structurally validating it does not prove CRDs are Established, controller
replicas are Available, any CSI driver can create a snapshot, or backup and
restore work. Downstream activation remains suspended until the site-specific
Flux Kustomizations and live evidence gates are explicitly approved.

## Pinned upstream inputs

All inputs come from `kubernetes-csi/external-snapshotter` tag `v8.5.0`, commit
`5aab051d1af135e2c852f6fb7fc27fa709d877bf`. The CRD files are byte-for-byte
vendored release assets; the strict verifier locks their SHA-256 values:

| Upstream path | SHA-256 |
| --- | --- |
| `client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml` | `75e6565aac2c0f2949ed13ea884bbaa388cb7be576b558b709cf1168e011828d` |
| `client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml` | `895a3c1e73b60f06a0deb566dd123d01bdf1b2efc5d5ff5231ff8bbcf42dafc7` |
| `client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml` | `b032116e987fb1d7d93cec7d942833adeb9d95c9bdc7e00c10b280c2fe4a6c33` |

The controller is adapted from the v8.5.0 upstream RBAC and deployment assets.
Their reviewed upstream SHA-256 values are recorded on the adapted resources.
Group-snapshot permissions and node access are deliberately omitted because
the vendored CRDs serve only the stable `snapshot.storage.k8s.io/v1` API (the
upstream dormant `v1beta1` schemas are neither served nor storage) and this
source does not enable distributed or group snapshotting. The multi-architecture
`snapshot-controller:v8.5.0` image is pinned to registry digest
`sha256:74ca61ab13e978f03cf0f336a607281d15f04cda0a38a881306365473b28a3d8`.

## Required Flux order

Use three separate Flux Kustomizations and fail closed on each dependency:

1. apply `crds` and wait for all three CRDs to report `Established=True`;
2. make `controller` depend on the Ready CRD Kustomization, then wait for both
   controller replicas and its PodDisruptionBudget;
3. make the selected Longhorn or Rook-Ceph profile Kustomization depend on the Ready controller Kustomization before unsuspending its Helm release.

The `cloudring.org/requires-stage` annotations make the source dependency
machine-verifiable, but Flux does not interpret them. The downstream GitOps
root must encode the same edges with `spec.dependsOn`, `wait: true`, and CRD or
Deployment health checks; that root owns its `sourceRef` and is intentionally
not fabricated by this provider-neutral source.

Do not combine these stages into one unordered apply and do not unsuspend a
storage release merely because the source renders. Before a readiness or
promotion claim, prove the selected CSI driver, retained snapshot, isolated
restore, cleanup, off-cell backup/restore, and continuity after one controller
node is unavailable.

## Local validation

```bash
kubectl kustomize deploy/kubernetes/storage/csi-snapshot-api/crds >/dev/null
kubectl kustomize deploy/kubernetes/storage/csi-snapshot-api/controller >/dev/null
go test ./internal/platformmanifest -run 'CSISnapshotAPI|Longhorn|RookCeph' -count=1
go run ./cmd/cloudring-manifestcheck --root .
go run ./cmd/cloudring-sourcecheck scan --scope full
```
