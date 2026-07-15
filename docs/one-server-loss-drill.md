# One-server-loss continuity drill

`cloudring-resilience one-server-loss` is the provider-neutral, read-only
observer for proving that one CloudRING installation survives loss and
recovery of one server. It does not stop, reboot, fence, or modify a server and
does not create, update, patch, or delete Kubernetes resources.

The public core owns the observation contract, strict Kubernetes decoding,
content-pinned adapter boundary, continuous timeline, and offline verifier.
Each downstream installation owns its private resource bindings, approved
fault procedure, provider API, rollback, generated evidence, and signature.

## Preconditions

Run the drill only when all of these are true:

- at least three control-plane nodes, three ready etcd members, and three ready
  API-server members exist;
- the selected server hosts a ready etcd member and API-server member;
- every declared critical workload already satisfies its minimum ready-pod and
  failure-domain counts;
- the exact declared KubeVirt VM is running on the selected server and can
  migrate within the declared maximum unavailability;
- the data-probe adapter executes a real deterministic application or database
  query and returns a digest of its canonical complete result;
- a separately reviewed downstream procedure can fault and recover only the
  selected server, with pre-state, rollback, and post-state recording; and
- the operator has a protected owner-only artifact directory and unique output
  names for this run.

A single-replica database on node-local storage, a non-migratable singleton VM,
or a ConfigMap that merely claims application health cannot satisfy these
preconditions. Fix that architecture first; do not weaken the request or
substitute a synthetic self-report.

## Private request

Start from
`contracts/one-server-loss/fixtures/synthetic-request.json` and keep the real
copy outside the repository. Workload entries bind a private namespace and
exact `matchLabels`. IDs and `queryRef` are safe opaque identifiers. The
adapter owns the installation-private mapping from `queryRef` to a canonical
read-only query.

Durations must use Go's canonical duration form. The runtime requires a poll
interval from one to thirty seconds, a stable pre-state covering at least two
poll intervals, a loss and recovery-stability window of at least two poll
intervals, bounded fault/recovery timeouts, and a bounded VM unavailability
SLO. The request is rejected if its worst-case sample count could exceed the
receipt bound.

## Data-probe adapter

The adapter is an absolute executable path. The observer snapshots and hashes
that exact executable once and never resolves it through `PATH` again. Every
sample invokes it without command-line arguments and supplies one strict JSON
request on stdin according to
`contracts/one-server-loss/probe-protocol.schema.json`.

The adapter receives a fresh anonymous replay of the same pipe-backed
kubeconfig and an allowlisted environment. It must execute the real read-only
query selected by `queryRef`, serialize the complete logical result
canonically, and return only:

- implementation and version;
- the exact request and adapter-executable SHA-256 bindings;
- `hashAlgorithm: sha256`;
- the data SHA-256 and validated byte count; and
- canonical UTC start/completion timestamps.

It must not print credentials, names, addresses, query text, rows, tenant data,
or child-process errors. The data digest and byte count must remain identical
in every pre-loss, loss, and recovered sample. Adapter failure, malformed or
unknown JSON, a different executable binding, or data drift aborts the run.

## Run

Create an owner-only directory and choose three new paths: private request,
ready marker, and unsigned receipt. Supply kubeconfig through an anonymous
pipe on descriptor 3; regular kubeconfig files are rejected by the replay
boundary.

```sh
credential-broker kubeconfig --stdout | \
  cloudring-resilience one-server-loss observe \
    --request /protected/run/request.json \
    --ready-marker /protected/run/ready-for-fault.json \
    --output /protected/run/unsigned-receipt.json \
    --kubectl /usr/local/bin/kubectl \
    --probe-adapter /opt/cloudring/bin/business-state-probe \
    --kubeconfig-fd 3 3<&0
```

The process samples continuously. It first requires the selected server's
stable UID, healthy `/readyz?verbose` result including `[+]etcd ok`, full
baseline quorum, workload minima, the VM on the selected server, and a real
data-probe result. It then atomically creates the non-overwriting `0600`
`ready-for-fault` marker.

The downstream fault procedure must validate the marker schema, recompute its
digest, compare its request/run bindings with protected state, and obtain any
required live mutation approval. Only then may it fault the exact selected
server. Keep the observer process running. A marker from another request, a
stale marker, or an exited observer is not authorization to act.

During loss, the observer requires:

- the selected node is not Ready;
- Kubernetes verbose readiness and etcd check still pass;
- a majority of the baseline control-plane, etcd, and API-server members stay
  ready;
- every declared workload remains above its ready-pod and distinct-node
  minima;
- the exact VM object UID remains stable and its running VMI becomes ready off
  the failed server within the declared SLO; and
- the real data digest and byte count remain unchanged on every sample.

Recovery must restore the same node name and UID, full baseline quorum,
control-plane pods, workloads, VM readiness, and data result for the complete
stability window. A replacement node with the same name fails closed. Missing
samples or a gap greater than two poll intervals also fail closed.

Only a successful complete timeline creates the receipt. Verify it without
cluster access:

```sh
cloudring-resilience one-server-loss verify \
  --receipt /protected/run/unsigned-receipt.json
```

The offline verifier recomputes sample, phase, baseline, and receipt digests;
checks sequence/timestamps/windows; and re-evaluates quorum, workload, VM,
node-identity, adapter, and data-continuity conditions. The receipt deliberately
contains only safe IDs, counts, booleans, timestamps, and hashes. It is still
an unsigned deployment-private artifact and must be signed by the downstream
evidence workflow before supporting a release decision.

## Kubernetes permissions

The observer needs only:

- non-resource `GET /readyz`;
- `get/list` Nodes and Pods;
- `get` on the exact VirtualMachine and VirtualMachineInstance; and
- whatever separate read-only permission the installation probe needs for its
  real query.

It does not need Kubernetes mutation verbs. The downstream fault procedure has
no place in this RBAC identity.

## Failure and recovery

If the observer exits before marker publication, do not cause the fault. If it
exits after the fault, use the already-approved downstream recovery procedure
to restore the selected server, confirm the platform manually, and classify
the drill as failed; never fabricate or merge partial samples. Preserve the
private logs needed for diagnosis, remove stale markers only after all fault
automation has stopped, and repeat with a new nonce and new paths.

The tool proves only the exact declared scope. It does not prove an undeclared
service, regional disaster recovery, multi-region failover, provider fencing,
or backup correctness. Those require their own fresh evidence.
