# Longhorn three-node storage profile

This provider-neutral profile supplies replicated CSI block storage for a
compact three-node CloudRING cell that has sufficient filesystem capacity but
no dedicated empty disks for Rook-Ceph. Sites with three independent raw
devices should use the Rook-Ceph RBD profile instead.

The `runtime` stage pins Longhorn `1.12.0` and defines two non-default,
three-replica `WaitForFirstConsumer` StorageClasses. Use
`longhorn-replicated` for ordinary RWO volumes such as PostgreSQL data and WAL.
Use `longhorn-migratable` only for KubeVirt disks whose PVC explicitly requests
`ReadWriteMany` and `volumeMode: Block`; the class enables Longhorn's
`migratable` volume mode needed by live migration. The profile also publishes
exactly one Velero-selected Retain VolumeSnapshotClass for
`driver.longhorn.io`. The provider-neutral
[`../csi-snapshot-api`](../csi-snapshot-api) source is the only owner of the
cluster-wide VolumeSnapshot CRDs and HA snapshot controller. The Longhorn
HelmRelease remains suspended by default and the Longhorn UI is not exposed by
an Ingress.

Every effective Longhorn runtime image is pinned by its multi-platform registry
index digest in the HelmRelease values: manager, engine, UI, instance manager,
share manager, backing-image manager, support-bundle kit, and all six CSI
sidecars. The shared supply-chain receipt records the exact tag-to-digest
mapping. The verifier compares that complete inventory with every image
repository/tag reference in the reviewed chart templates; mutable tags,
missing entries, extra runtime image references, or digest drift fail closed.

Longhorn does not publish an official OCI chart for this release. The exact
official `longhorn-1.12.0.tgz` release chart is therefore vendored under
[`vendor/longhorn`](vendor/longhorn). Its archive SHA-256 is
`869bb20701b154473606f1e8967b27f34f2448a2dfe6eb8970f1cae6957384f5`;
the upstream commit, Apache-2.0 license receipt, exact file count, and canonical
vendored-tree digest are recorded in [`vendor/UPSTREAM.json`](vendor/UPSTREAM.json)
and [`../../runtime-chart-supply-chain.json`](../../runtime-chart-supply-chain.json).
The verifier rejects any file, receipt, license, or tree-digest drift.

Flux renders the reviewed chart from the official Longhorn charts repository
through the namespaced `longhorn-system/longhorn-charts` `GitRepository`. The
source is pinned with `spec.ref.commit` to the exact upstream release commit
`f8def0504bf3f5f26c342941c9e4532b44830ebe`; the HelmRelease reads
`./charts/longhorn` from that immutable artifact. The verifier rejects a
branch, tag, missing commit, mismatched commit, source URL change, or mutable
HTTP HelmRepository. The vendored release archive remains the independent
review copy used to check the upstream chart identity and license.

The snapshot class fixes `parameters.type: snap`, so Velero's CSI data mover
starts from a local Longhorn snapshot and copies the volume data to the
configured off-cell backup store. Do not change this class to Longhorn's native
backup mode: that is a separate workflow which requires a Longhorn backup
target and would make this Velero restore path depend on redundant storage
configuration.

Apply the canonical CSI snapshot source through separate Flux Kustomizations:
`../csi-snapshot-api/crds` first, then its `controller`, then this `runtime`
stage. The controller Kustomization must depend on the Ready CRD Kustomization,
and this profile Kustomization must depend on the Ready controller
Kustomization. Do not install CRDs or a second snapshot controller in this
profile: one cluster-wide deployment serves every CSI driver.

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
through one storage-node loss and recovery. For a migratable VM, additionally
prove that its disk PVC is RWX Block, the corresponding Longhorn volume reports
three healthy replicas, and live migration plus node-loss recovery preserve
guest data. Keep the source HelmRelease suspended for templates and activate it
only in an audited site Flux root after those host prerequisites are checked.
