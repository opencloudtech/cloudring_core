# Rook-Ceph RBD storage profile

This provider-neutral profile defines a staged, fail-closed Rook-Ceph RBD
installation for CloudRING. It contains no provider inventory, credentials,
tenant data, or live evidence.

The stages are deliberately separate:

1. `controllers` installs the pinned Rook `v1.20.2` operator, CRDs, and CSI
   operator in the `rook-ceph` namespace.
2. `cluster-example` defines one Ceph `v20.2.2` cluster, one replicated RBD
   pool, a non-default `WaitForFirstConsumer` StorageClass, and the single
   Velero-selected Retain VolumeSnapshotClass. Its HelmRelease is suspended by
   default.

Do not remove `spec.suspend` until a downstream site overlay replaces all three
synthetic node names and `/dev/disk/by-id` values with exact, dedicated, empty
devices verified on three independent Kubernetes hosts. The overlay must keep
`useAllNodes: false`, `useAllDevices: false`, device encryption, replica size
three, and host failure-domain placement. Automatic discovery or consumption
of unlisted disks is forbidden.

The cluster stage also requires the Kubernetes CSI snapshot CRDs and snapshot
controller to be installed before reconciliation. Apply the stages through
separate Flux Kustomizations and make the cluster Kustomization depend on the
Ready controllers Kustomization. Do not combine the stages into one apply.

This iteration intentionally exposes RBD only. CephFS, RGW, erasure-coded
pools, a default StorageClass, and a second Velero snapshot-class label are
forbidden by the source verifier. Providers can contribute those capabilities
as separate reviewed profiles when they have their own durability, isolation,
backup, and live failure evidence.

Before production promotion, prove all of the following on the target site:

- three Ready monitors and three independent encrypted OSDs with `HEALTH_OK`;
- the exact `rook-ceph.rbd.csi.ceph.com` driver and RBD StorageClass;
- PVC provisioning, write/read checksum continuity, Retain snapshot creation,
  isolated restore, and cleanup;
- off-cell Velero backup and restore without source-namespace or tenant escape;
- continued service after one storage host is unavailable, followed by clean
  recovery.

The source profile and a Ready HelmRelease are not storage-readiness evidence.
The profile does not select hardware and makes no live or production claim.
