# Provider Site Preflight And Plan

CloudRING defines a provider-neutral site profile so downstream providers can
describe inventory and installation dependencies without forking platform
code. The profile contains references to provider-owned inventory, network,
storage, identity, GitOps, observability, upgrade, and rollback records. It
must not contain credentials, tenant data, or live endpoint values.

Validate a profile and render its deterministic plan:

```sh
go run ./cmd/cloudring-site preflight --profile ./examples/provider-site-profile.yaml
go run ./cmd/cloudring-site plan --profile ./examples/provider-site-profile.yaml
```

Preflight fails closed unless the profile declares at least three control-plane
nodes, three workers, and three public Gateway nodes, with every role spread
across at least three failure domains;
unique provider-resource, management-address, and provisioning-address
references; dual-stack networking;
snapshot-capable storage; immutable off-cell backup; OIDC, workload identity,
and a provider-owned runtime-input broker backed by its secret store; GitOps
bootstrap, upgrade, and rollback inputs; metrics,
logs, traces, and alerts; and OCSv3 conformance.

The plan is stable for the exact decoded profile. It orders read-only inventory,
network, identity, storage, and observability checks before a single bootstrap
phase and binds that mutation phase to the declared rollback reference. The
CLI never applies the plan and never reads referenced values.

The JSON Schema enforces structure and the baseline three-node role counts.
The Go preflight is authoritative for cross-field rules, including declared
availability values, per-role failure-domain spread, and reference uniqueness.

Downstream repositories own concrete site-class profiles and provider adapter
implementations. They should validate their private profile in CI, bind each
reference through their own secret-safe inventory mechanism, implement the
declared phases, and retain live acceptance and rollback evidence outside the
public repository.

For upstream Kubernetes installations where an approved Gateway node is also a
kubeadm control-plane node, bootstrap must explicitly reconcile
`node.kubernetes.io/exclude-from-external-load-balancers`. Remove that label
only from the profile's declared `gateway` nodes, after a server-side dry-run
and a captured pre-state. Retain an exact rollback that restores the label.
Acceptance must fail closed unless the LoadBalancer advertises one IPv4 and one
IPv6 address per declared Gateway node, every advertised address answers from
each failure domain, all nodes remain Ready, and temporary probes are removed.

Provider-neutral storage bases live under `deploy/kubernetes/storage`. Use the
Rook-Ceph RBD profile when three independent, dedicated raw devices are
available. Use the suspended Longhorn three-node profile for a compact cell
with enough supported filesystem capacity but no dedicated Ceph devices. Both
profiles require downstream host preflight, activation, backup/restore, node
loss, recovery, and cleanup evidence; a source render is never live readiness.

The reusable database profile lives under
`deploy/kubernetes/postgresql-ha`. It pairs the compact Longhorn profile with
three hard-separated PostgreSQL instances and synchronous quorum durability.
Downstream installations still own credential injection, sizing, immutable
off-cell WAL/base backup, isolated restore, application cutover, and live
primary-node-loss evidence.

This slice is a complete preflight and planning contract. It is not an
installer, does not prove that any provider site is reachable, and does not
claim production readiness.
