# Longhorn three-node storage profile

This provider-neutral profile supplies replicated CSI block storage for a
compact three-node CloudRING cell that has sufficient filesystem capacity but
no dedicated empty disks for Rook-Ceph. Sites with three independent raw
devices should use the Rook-Ceph RBD profile instead.

The `runtime` stage pins Longhorn `1.12.0`, defines a non-default three-replica
`WaitForFirstConsumer` StorageClass, and publishes exactly one Velero-selected
Retain VolumeSnapshotClass for `driver.longhorn.io`. It also installs the
single cluster-wide, two-replica CSI snapshot controller that an upstream
Kubernetes distribution requires. The controller version matches Longhorn's
CSI snapshotter (`v8.5.0`), and its multi-architecture image is pinned by
digest. The HelmRelease is suspended by default and the Longhorn UI is not
exposed by an Ingress.

The Kubernetes CSI VolumeSnapshot v1 CRDs must already exist before this stage
is activated. A cluster bootstrap or another explicitly selected storage
profile may own those cluster-wide CRDs, but it must use schemas compatible
with external-snapshotter `v8.5.0`. Do not install a second snapshot controller:
one deployment serves every CSI driver in the cluster.

Before a downstream overlay activates the release, it must prove all intended
storage nodes are Linux nodes with adequate independent capacity, mount
propagation, supported ext4 or XFS storage, and the Longhorn host tools. On
Debian-family hosts the V1 data engine requires `open-iscsi`, `nfs-common`, a
usable `iscsiadm`, an enabled and running `iscsid.service`, and the required
kernel modules including `iscsi_tcp` and `dm_crypt`. An active
`iscsid.socket` alone is not a readiness gate. Disable `multipathd` when it is
not required; otherwise configure it to exclude Longhorn devices according to
the upstream Longhorn guidance and prove the warning is absent. The default
data path is `/var/lib/longhorn`; do not activate this profile where that path
shares an undersized operating-system filesystem.

The compact profile keeps node- and disk-level replica anti-affinity hard, so
the three replicas require three schedulable nodes and disks. Zone-level
anti-affinity is soft because compact provider cells commonly place all three
nodes in one Kubernetes zone, and Longhorn treats nodes without a
`topology.kubernetes.io/zone` label as the same zone. Multi-zone sites should
label their real failure domains; Longhorn will still prefer separate zones.

Activation is not readiness. Before promotion, prove the v1 snapshot CRDs and
both snapshot-controller replicas are Ready, the exact CSIDriver, three healthy
replicas on separate nodes, PVC checksum continuity, Retain snapshot and
isolated restore, off-cell Velero backup and restore, cleanup, and continuity
through one storage-node loss and recovery. Keep the source HelmRelease
suspended for templates and activate it only in an audited site Flux root after
those host prerequisites are checked.
