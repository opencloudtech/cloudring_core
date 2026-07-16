# Containerized Data Importer profile

This provider-neutral profile installs KubeVirt Containerized Data Importer
(CDI) `v1.65.0` for persistent virtual-machine disks. The upstream operator
release asset is vendored from commit
`d97a33c2f063258d14c6df27d297e84e3f48b779`; its original SHA-256 is
`e96d59abdf358c5161cb96adcfdcc6107efc3fb608ec93ade11578c94a222015`.
CloudRING pins the operator and every operand image to immutable registry
digests.

The `controllers` stage is deliberately inert: the operator Deployment has
zero replicas. A reviewed site overlay explicitly selects `activation`, which
raises it to one replica. After the operator Deployment and the `CDI` CRD are
healthy, the site may apply `runtime`. Keep those as ordered Flux roots so an
operator upgrade cannot race the runtime custom resource.

The runtime enables delayed-binding support and blocks CDI removal while
managed workloads exist. It does not select a storage implementation. Each
DataVolume must name an installation-provided replicated StorageClass, an
explicit capacity and access mode, and an immutable source. Keep VM root disks
as standalone DataVolumes rather than VM-owned templates when the disk must
survive VM replacement.

Raw block imports run as the non-root QEMU identity. Every containerd-backed
node that may run a CDI importer or KubeVirt workload must therefore set
`device_ownership_from_security_context = true` in the active CRI plugin
configuration, validate the resulting TOML with `containerd config dump`, and
restart containerd one node at a time. A site preflight must reject activation
when the setting is absent or false on any eligible node. Retain the previous
containerd configuration until the node is Ready again and a real raw-block
DataVolume import succeeds. This is required for containerd config schemas v2
and v3; changing the importer to run as root is not an acceptable substitute.

Structural verification is not a live durability claim. Before promotion,
prove the controller and webhook health, a digest-pinned import, bound PVC/PV
and backend replicas, VM restart/reattach continuity, retained snapshot,
off-cell backup, isolated restore, cleanup, tenant isolation, and one-server
loss continuity.
