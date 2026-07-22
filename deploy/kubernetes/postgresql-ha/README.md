# PostgreSQL HA profile

This provider-neutral profile supplies a three-instance PostgreSQL 18 cluster
for durable CloudRING control-plane and business state. CloudNativePG owns
election and failover; the database pods use hard hostname anti-affinity,
quorum synchronous replication, fail-closed durability, TLS, SCRAM, separate
replicated PGDATA and WAL volumes, and a PodDisruptionBudget.

The `controllers` stage pins the CloudNativePG chart, operator image, webhook
failure policy, security context, and two controller replicas. Its source
HelmRelease is suspended. A downstream Flux root must activate it only after
the CRDs, cert-manager, target Kubernetes version, and two-node controller
placement have been checked.

The chart is reconciled from the official
`oci://ghcr.io/cloudnative-pg/charts/cloudnative-pg` OCI artifact at manifest
digest
`sha256:209c588b902982bf283a0073db83edd422d9710a2c8a670fe57c0329abe789a4`.
The reviewed chart content and provenance layer digests are recorded in
[`../runtime-chart-supply-chain.json`](../runtime-chart-supply-chain.json).
The operator remains image-pinned to CloudNativePG `1.30.0`. A mutable Helm
repository, tag-only chart selection, or documentary checksum annotation is
not an accepted source contract.

The `runtime` stage is an explicit compact-cell reference. It consumes the
public `longhorn-replicated` StorageClass and `longhorn-retain`
VolumeSnapshotClass. Sites using another CSI implementation must patch both
names to durable equivalents that have already passed the CloudRING storage
and restore gates. Do not reduce three replicas, hard host separation,
synchronous durability, failover quorum, TLS, or the retained snapshot policy.

Before applying the runtime stage, the installation must create two
`kubernetes.io/basic-auth` Secrets in `cloudring-database` through the approved
runtime secret manager. Both must carry `cnpg.io/reload: "true"` and contain
only `username` and `password`:

- `cloudring-postgres-owner` uses username `cloudring_owner`. This database and
  schema owner is projected only into the one-shot migration workload.
- `cloudring-postgres-app` uses username `cloudring_app`. This managed role is
  the only identity projected into steady-state application workloads.

Never put either Secret, a DSN, or rendered secret material in Git, command
arguments, logs, or evidence. After the migration gate, remove the migration
workload and its owner-Secret projection; do not expose owner credentials to
application replicas. Applications must use the generated read-write service
with `sslmode=verify-full` and the operator-generated CA bundle. Same-namespace
clients require the pod label `cloudring.org/postgresql-client: "true"`.
Cross-namespace clients require both that pod label and the namespace label
`cloudring.org/postgresql-client-namespace: "true"`; neither label alone grants
access. A downstream may replace this with an equally narrow NetworkPolicy.

The included six-hour online VolumeSnapshot schedule is a local recovery
layer, not an off-cell disaster-recovery claim. Production promotion still
requires an immutable off-cell WAL/base-backup target, a real isolated restore,
retention and cleanup proof, and a one-server-loss drill with an unchanged
transactional-state digest. A downstream may add the official Barman Cloud
CNPG-I plugin for S3-compatible storage; it must pin the plugin by version and
image digest and keep its credentials in the runtime secret manager.

Run schema migration before starting application replicas by passing the DSN
through an inherited anonymous pipe:

```text
cloudring-postgres-migrate --dsn-fd 3 \
  --owner-role cloudring_owner --application-role cloudring_app
```

The pipe must contain one self-contained URL-form DSN with its username and
password. The process rejects PostgreSQL environment fallbacks, passfiles,
service files, and client-key indirection; an explicitly referenced read-only
CA bundle remains supported through `sslrootcert`. It accepts neither a DSN
argument nor a DSN environment variable nor a regular file, and it never echoes
database configuration or underlying driver errors. It accepts a completely
absent schema or the exact recorded schema, owner, constraints, and grants;
partial or pre-existing drift fails closed. The application identity receives
only schema usage and document SELECT/INSERT/UPDATE/DELETE privileges and
cannot execute DDL in the CloudRING schema.
