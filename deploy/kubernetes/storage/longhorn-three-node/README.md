# Longhorn three-node storage profile

This provider-neutral profile supplies replicated CSI block storage for a
compact three-node CloudRING cell that has sufficient filesystem capacity but
no dedicated empty disks for Rook-Ceph. Sites with three independent raw
devices should use the Rook-Ceph RBD profile instead.

The `runtime` stage pins Longhorn `1.12.0`, defines a non-default three-replica
`WaitForFirstConsumer` StorageClass, and publishes exactly one Velero-selected
Retain VolumeSnapshotClass for `driver.longhorn.io`. The HelmRelease is
suspended by default and the Longhorn UI is not exposed by an Ingress.

Before a downstream overlay activates the release, it must prove all intended
storage nodes are Linux nodes with adequate independent capacity, mount
propagation, supported ext4 or XFS storage, and the Longhorn host tools. The V1
data engine requires `open-iscsi`, a usable `iscsiadm`, and a running or
socket-activated `iscsid`. The default data path is `/var/lib/longhorn`; do not
activate this profile where that path shares an undersized operating-system
filesystem.

Activation is not readiness. Before promotion, prove the exact CSIDriver,
three healthy replicas on separate nodes, PVC checksum continuity, Retain
snapshot and isolated restore, off-cell Velero backup and restore, cleanup,
and continuity through one storage-node loss and recovery. Keep the source
HelmRelease suspended for templates and activate it only in an audited site
Flux root after those host prerequisites are checked.
