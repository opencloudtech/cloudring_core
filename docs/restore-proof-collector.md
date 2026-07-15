# Velero CSI Data-Mover Restore-Proof Collector

`cloudring-backup` is a read-only collector toward Kubernetes and the storage
provider for one volume data path through Velero 1.18.2 CSI data mover. It
writes only local private artifacts. The receipt binds the exact Backup,
Restore, source PVC/PV, restored PVC/PV, retained and archived DataUpload,
temporary DataUploadResult ConfigMap, terminal DataDownload, an independent
content comparison, and provider-side cleanup observations.

The tool exists in the public core because the contract, typed decoders,
canonicalization, archive safety, and provider-neutral collection workflow are
reusable. Provider credentials, installation configuration, live tenant data,
and generated receipts remain downstream.

## Supported scope

The first contract supports exactly:

- Velero `v1.18.2`, proven by a fresh official `ServerStatusRequest` processed
  by the running server;
- a direct Backup with `snapshotMoveData: true`;
- the built-in Velero CSI data mover;
- Linux volume data paths;
- one source PVC copied to an isolated namespace with the same PVC name;
- one terminal `DataDownload`, one restored provider handle, and one
  source-continuity handle;
- distinct source and restored PV identities and CSI handles under the same
  observed provider implementation;
- a retained DataUpload present after restore cleanup;
- two provider absence observations at least ten seconds apart, followed by a
  source-handle presence observation and a final Kubernetes/source quiet fence
  after the second target absence observation.

The exact patch level is intentional: Velero 1.18.2 skips
VolumeGroupSnapshot cleanup for backups that did not use group snapshots,
avoiding the spurious restore warning fixed upstream in
[velero#9900](https://github.com/vmware-tanzu/velero/pull/9900).

CSI-native snapshot restore and filesystem backup are rejected. Supporting a
new Velero version or restore method requires a separately versioned decoder,
contract, negative tests, and downstream acceptance.

## Trust and mutation boundary

The collector uses `kubectl get --raw` against exact API paths. It does not
create, patch, or delete Kubernetes resources. It waits for the downstream
restore workflow to remove the restored PVC, PV, DataDownload, and
DataUploadResult ConfigMap. It fails closed if it observes a name under a
different UID, but polling cannot prove that a short-lived replacement never
existed between observations.

Two executable adapters are required:

- The data-probe adapter compares source and restored content and returns only
  SHA-256 digests, byte count, timestamps, implementation/version, and a hash
  of protected runtime evidence. It must hash the complete logical byte stream
  represented by the Velero DataUpload/DataDownload, and `validatedBytes` must
  equal their terminal total byte count; partial sampling is rejected.
- The provider observer checks whether the raw CSI handle exists. The handle
  is supplied on stdin, never in argv or an environment variable. The adapter
  returns only presence status and source-safe receipt metadata.

The runtime copies each adapter into a private pinned executable snapshot,
hashes it, and binds every invocation to a fresh nonce and exact request digest.
Each invocation has a hard timeout and Unix process-tree termination.
Adapters remain installation trust roots: operators must review them and
separately sign accepted receipts. An unsigned receipt proves internal
consistency; it is not sufficient by itself for a release claim.

## Workflow

Before starting the restore, capture the immutable source baseline:

```sh
go run ./cmd/cloudring-backup baseline \
  --request /protected/baseline-request.json \
  --output /protected/source-baseline.json
```

Create the private baseline request from
`contracts/backup-proof/fixtures/synthetic-baseline-request.json`. After the
Restore completes, the downstream workflow creates an official
`velero.io/v1` `ServerStatusRequest`, waits for `status.phase: Processed`, and
records its exact name plus SHA-256 of its UID in a separate collection request
derived from `synthetic-collection-request.json`. Start collection while that
object is still present: Velero 1.18.2 expires processed requests after one
minute. The collector requires `serverVersion: v1.18.2`, a processing timestamp
after Restore completion, exact UID/GVK, and the one-minute freshness bound.
The downstream workflow also generates a fresh random run nonce, puts only its
SHA-256 digest in `cleanupRunNonceSha256`, and checks that exact digest in the
cleanup-ready marker before acting. Keep all real names, UIDs, nonces, and
generated artifacts outside the repository.

After the Restore and independent target workload are ready, start collection:

```sh
go run ./cmd/cloudring-backup collect \
  --request /protected/collection-request.json \
  --baseline /protected/source-baseline.json \
  --archive /protected/backup-contents.tar.gz \
  --data-probe-adapter /opt/cloudring/bin/volume-probe \
  --provider-adapter /opt/cloudring/bin/provider-observer \
  --cleanup-ready /protected/run-unique.cleanup-ready.json \
  --cleanup-timeout 30m \
  --output /protected/unsigned-volume-receipt.json
```

The collection process first proves that the source is unchanged, validates
the live Velero object lineage, and proves that the retained DataUpload matches
its exact archived copy. It then invokes the data probe and verifies the
provider artifact is present. Only after both checks succeed, it atomically
creates the non-overwriting `0600` cleanup-ready marker. A downstream workflow
must wait for that run-unique marker, validate it against
`contracts/backup-proof/cleanup-ready.schema.json`, and compare its
`cleanupRunNonceSha256` with the current request before performing
isolated-restore cleanup. Every delete must use both the exact original UID and
resourceVersion through Kubernetes `DeleteOptions.preconditions`, stop on
either precondition conflict, and prevent re-creation of those names until the
collector emits a valid receipt or the run is abandoned. These raw UIDs and
resourceVersions belong in protected downstream runtime state, not in the
marker or repository. The downstream workflow must
also parse `readyAt` as canonical RFC 3339 UTC, not merely match the schema
pattern. The marker contains only a schema, fixed
status, timestamp, and nonce digest; it has no object identifiers, provider
handle, raw nonce, or tenant data. The collector waits; it does not initiate
cleanup. A 404 is accepted only after a successful exact-name list against the
same GVR confirms absence. When all exact Kubernetes objects are absent, it
performs two target-provider absence reads, proves the distinct source provider
handle is still present, and then runs a final source PVC/PV and target
Kubernetes fence.
Immediately before marker publication, the collector re-reads every cleanup
target and verifies its original UID, resourceVersion, and canonical state,
but that fence does not replace UID/resourceVersion preconditions or the
downstream no-recreation rule after publication.

Use a protected local directory with no concurrent writers. Artifact
publication relies on a synced temporary file plus an atomic same-directory
hard link; unsupported filesystems fail closed. The marker is a coordination
barrier, not durable release evidence and not proof that cleanup finished.

Verify a stored receipt without cluster or provider access:

```sh
go run ./cmd/cloudring-backup verify \
  --receipt /protected/unsigned-volume-receipt.json
```

Output files are created with mode `0600` and never overwrite an existing
path. CLI status output contains no object names, UIDs, provider handles,
paths, tenant content, or child-process stderr.

## Kubernetes access

For a credential broker or supervisor that supplies kubeconfig through an
anonymous pipe, pass the inherited descriptor explicitly:

```sh
cloudring-backup baseline \
  --request /protected/baseline-request.json \
  --output /protected/source-baseline.json \
  --kubeconfig-fd 3

cloudring-backup collect \
  --request /protected/collection-request.json \
  --baseline /protected/source-baseline.json \
  --archive /protected/backup-contents.tar.gz \
  --data-probe-adapter /opt/cloudring/bin/volume-probe \
  --provider-adapter /opt/cloudring/bin/provider-observer \
  --cleanup-ready /protected/run-unique.cleanup-ready.json \
  --cleanup-timeout 30m \
  --output /protected/unsigned-volume-receipt.json \
  --kubeconfig-fd 3
```

The collector consumes that pipe once, keeps the bounded kubeconfig only in
process memory, and gives every `kubectl` invocation a fresh anonymous replay
pipe. It rejects regular-file descriptors, never places kubeconfig bytes in
argv or the environment, and removes unrelated credential variables from the
`kubectl` environment. This pipe mode is supported on Linux and macOS; no
native Windows claim is made. The default mode without `--kubeconfig-fd`
continues to use the operator's normal Kubernetes client configuration.

Consumers that build additional multi-query collectors must pin the executable
once with `secureexec.Pin` and invoke it through `Executable.Run`. That boundary
does not consult `PATH` again, bounds stdout and stderr, creates a separate Unix
process group, bounds inherited-I/O waiting, kills remaining descendants, and
only then completes the replay writer. Calling `Replay.Attach` or
`kubeconfigpipe.Run` directly leaves executable identity or output bounds with
the caller and is not sufficient by itself for a credential-bearing production
collector.

The collector needs only `get` on the exact named Backup, Restore,
ServerStatusRequest, source and target PVC, and source and target PV, plus `get/list` on
PVCs, PVs, DataUploads, DataDownloads, and ConfigMaps used for exact reads,
bounded selection, and 404 confirmation. It does not need create, update,
patch, or delete permission. Creation of the short-lived ServerStatusRequest
belongs to the downstream workflow, not this collector. The BackupContents
archive is supplied as an existing protected file; its acquisition policy
remains downstream.

List pagination is bounded and must keep one resourceVersion. Duplicate
identities, repeated continuation tokens, wrong GVK/GVR, source mutation,
replacement races that are observed, unknown adapter output, and trailing or
duplicate-key JSON all fail closed. A target with `deletionTimestamp` remains present until exact
absence is confirmed; deleting inputs before cleanup are rejected. Aggregate
list bytes/items/pages and cleanup time are hard-bounded. Archive extraction
enforces compressed, expanded, entry-count, and entry-size limits and rejects
links, traversal, duplicate paths, mismatched preferred-version copies,
concatenated gzip members, and trailing tar data.

## Recovery

The collector makes no live mutation, so interruption requires no rollback.
After confirming that the collector has exited and no downstream cleanup is
still active, remove a stale cleanup-ready marker. Delete any incomplete
private output, confirm the source baseline still matches, obtain a fresh
BackupContents archive if its identity changed, and repeat collection with a
new isolated restore and new marker path. Existing output or marker paths are
never reused; a stale path makes a new run fail closed.

## Non-claims

Every receipt declares `single-volume-data-path-only`; it is not a full Restore
scope or tenant-isolation proof. This slice does not claim production
readiness. It does not deploy an object store, create a backup, sign evidence,
prove five restore targets, or validate tenant isolation beyond the selected
volume path. It proves that source and target PV identities and handles are
distinct and observes the source handle present after target cleanup, but does not replace a full source
workload/application consistency check. It also does not prove an unobserved
create-delete interval or downstream compliance with preconditioned deletion.
Those claims require the real off-cell immutable store, a complete
backup, signed multi-target restore drill, cleanup, tenant-isolation, and
downstream release gates.
