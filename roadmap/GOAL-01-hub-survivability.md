# Goal 01 — Reference-Cell Critical-Path Survivability

**Slug:** `reference-cell-critical-path-survivability`

**Outcome:** the declared critical path of the three-server reference cell survives
loss of any one server within its SLO, with no acknowledged order, support-ticket, or
audit-event loss. Bad changes in the covered GitOps and state-migration paths are
rejected or rolled back to the prior signed revision.

This claim covers only the enumerated critical path of one three-server reference
cell in one region. It is not a claim of uninterrupted service, availability of every
cell component, whole-provider availability, multi-region operation, off-cell product
runtime, marketplace operation, or federation survivability.

## Verified starting point

A fresh downstream read-only audit on 2026-07-22 observed three Ready Kubernetes
control-plane nodes, three Ready etcd members, a three-instance PostgreSQL cluster,
available backup locations and repositories, and Ready certificate resources. These
observations are not Goal 01 acceptance evidence by themselves.

The remaining verified debt is material:

- the provider portal still selects a ConfigMap state backend;
- Flux inventory is partial, pruning is not yet safely enabled, and some
  Kustomizations are suspended;
- existing backups and restores do not yet prove the strict off-cell object,
  retention, class coverage, application checksum, isolated restore, and cleanup
  contract;
- sequential one-server-loss continuity has not been freshly proved for every server;
- the iteration gate is still bound to an older goal sequence;
- the normal management path and separate break-glass fallback need a current,
  source-safe proof.

## Scope

### Public core

- Reusable strict backup, restore, off-cell receipt, and cleanup contracts.
- Phase-aware kubeadm HA wave planning and verification without provider secrets.
- Flux ownership and drift-audit contracts with an explicit allowlist.
- PostgreSQL transactional state, additive migrations, immutable audit journal, and
  reusable continuity primitives.
- A critical-path survivability inventory and machine-readable Goal 01 gate inputs.
- Generic runbooks and tests with synthetic values.

### Reference cell

- A stable vRack/private management path independent of public SSH, plus a separately
  brokered and audited break-glass fallback.
- Off-cell object storage, reconciled secret references, backup locations,
  repositories, schedules, etcd shipping, and an isolated restore drill.
- Verification of the already observed three-member control plane and etcd; any
  necessary correction occurs one node and one wave at a time with a fresh backup
  between waves.
- Staged Flux adoption of every declared live resource family. Pruning is enabled only
  after the ownership verifier proves that family safe.
- Migration of orders, support tickets, sessions, and audit events from the production
  ConfigMap path to three-instance CloudNativePG.
- cert-manager origin issuance and renewal proof.
- Current goal/gate state and deployment-specific, sanitized live receipts.

Concrete credentials, endpoints, topology values, tenant data, and live evidence do
not belong in the public repository.

### Out of scope

- New tenant-facing features or products.
- The installer, observability overhaul, general upgrade train, or external OIDC
  cutover.
- Local or remote OCS runtime, product moderation, offers and entitlements, billing,
  marketplace, multi-cell operation, or federation. Goal 01 only preserves their
  architecture invariants.

## Critical-path inventory

The downstream gate must enumerate every accepted component with:

- replicas and independent failure domains;
- stable client endpoint and failover mechanism;
- state location, client retry/timeouts, and maximum unavailability;
- RPO/RTO, backup and restore coverage;
- one-server-loss evidence and explicit non-claims.

At minimum this includes Kubernetes API and etcd, public ingress and portal, identity
dependencies, PostgreSQL, Flux source and controllers, secret reconciliation,
certificates and DNS automation, storage control plane, backup destination, and the
operator management path. Multiple pods alone do not prove high availability.

## State and rollout rules

- State migrations use additive expand/contract phases and support mixed application
  versions during rollout.
- Writes use optimistic concurrency and stable idempotent operation identifiers.
- Audit events are append-only. The state boundary remains compatible with a future
  transactional outbox for control and billing events.
- ConfigMap fallback is allowed only before PostgreSQL accepts a post-migration write.
  After that boundary, rollback keeps the PostgreSQL state and restores a compatible
  application revision or performs a governed database recovery.
- A production rollout must keep an order/audit write-read probe running. Acknowledged
  writes may not disappear during portal restart, PostgreSQL primary failover, node
  loss, or application rollback.

## Acceptance criteria

1. **Sequential server loss:** each server is taken unavailable separately and restored
   before the next drill. The portal, public API, Kubernetes API, order write/read, and
   audit append/read probes remain within the declared SLO. Critical workloads and
   state are verified after every step.
2. **Strict restore:** etcd, PostgreSQL/portal state, Flux state required to rebuild the
   installation, and all declared platform classes are restored from exact off-cell
   object versions into an isolated target. Checksums, retention-delete denial,
   cleanup, and no cross-tenant leakage are proved.
3. **Zero unmanaged resources:** the ownership verifier reports an empty allowlist and
   no live resource outside the declared Flux inventory. A controlled, reversible
   drift fixture is reconciled to the signed desired revision.
4. **Durable state:** portal restart, PostgreSQL primary failover, and single-node loss
   lose zero acknowledged orders, tickets, and audit events. The pre/post verifier uses
   opaque synthetic test records and cleans them up.
5. **Origin certificates:** issuer, certificate, renewal time, and direct-origin TLS
   checks pass without treating an edge certificate as origin evidence.
6. **Management continuity:** the normal private path and the independently brokered
   fallback both pass; public SSH failure does not remove operator access.
7. **Unified gate:** fresh alpha, backup, HA, ownership, certificate, state, management,
   and one-server-loss evidence makes the Goal 01 gate report
   `mutationAllowed: true`. Missing, stale, unsigned, or mismatched evidence fails
   closed.
8. **Delivery:** reusable changes are accepted in public `main`, the exact public SHA is
   pinned downstream, deployment-specific changes reconcile from downstream `main`,
   CI is green, and status/runbooks reference the final sanitized evidence.

## Definition of done

- Every acceptance criterion has fresh machine-readable evidence from the reference
  deployment and an exact cleanup/rollback result.
- Production memory/ConfigMap state is removed after the rollback boundary is crossed;
  no competing production persistence path remains.
- The public requirement rows in `COVERAGE.md` are accepted only for the precise
  behavior proved here; broader product requirements remain pending.
- Any GitHub issue fixed by the delivered code is regression-tested, commented with
  the accepted SHA and evidence, and closed only after downstream deployment.
