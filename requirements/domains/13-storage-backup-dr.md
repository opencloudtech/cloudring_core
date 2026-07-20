# 13 — Storage, Backup & Disaster Recovery

This domain defines how CloudRING stores tenant and platform data and how it
proves that data survives: block, file, and object storage exposed only behind
versioned contracts; a media taxonomy with deterministic performance models and
per-volume QoS; mount fencing; crash-consistent checkpoints with changed-block
tracking on a content-addressed chunk store; the platform backup control plane
with a signed backup barrier before any mutation; restore drills as promotion
blockers with published per-service RPO/RTO; cross-tenant and cross-namespace
isolation; control-plane state (etcd) snapshots; offsite immutable copies;
first-class durability evidence states; and the maintenance interlock between
fleet operations and storage.

**Domain contract.** (1) No storage surface is consumed except through a
versioned, provider-neutral contract — implementations are replaceable.
(2) No mutating operation on stateful data proceeds without a fresh, signed
backup-barrier receipt; the gate fails closed. (3) No stateful service is
declared ready to serve production workloads without fresh `verified` restore-drill evidence;
`blocked`, `stale`, and `synthetic` evidence never promotes. (4) Fencing and
isolation fail closed: ambiguous ownership or a double-attach attempt halts IO
rather than risking corruption or exposure. (5) Durability evidence is
append-only, UTC-stamped, and honest — blocked evidence stays visible and is
never laundered into a claim.

---

### CR-STO-010 — Storage classes behind versioned contracts
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, tenant, vendor
- **Problem:** Storage implementations churn faster than platform lifetimes.
  Wiring block/file/object services directly into platform internals locks
  every tenant workload to one backend and makes replacement a migration
  crisis.
- **Requirement:** All block, file, and object storage MUST be exposed through
  versioned, provider-neutral contracts declaring capabilities, performance
  classes, topology (zone/region), durability semantics, and data-lifecycle
  actions. Platform core and service modules MUST consume storage only through
  these contracts and their Kubernetes bindings; no service-specific storage
  driver, UI, billing, or automation logic may be wired into the platform
  core. Contract versions MUST carry deprecation and migration guidance.
- **Acceptance evidence:** contract schema plus validator enforced in CI;
  conformance check rejects a module that references an implementation-specific
  storage product; integration test suite provisions block/file/object volumes
  through the contract on the reference installation; contract changelog and
  deprecation policy published.
- **Non-goals:** choosing a single storage engine for all providers; exposing
  raw engine administration APIs to tenants.
- **Non-claims:** no production storage engine is yet integrated behind the
  contract; contract stability across a real engine swap is unproven.
- **Stop conditions:** exposure/trust — halt rollout if any platform component
  bypasses the contract to reach a storage implementation directly; keys —
  halt if a contract would pass raw credentials instead of workload-identity
  references.
- **Traceability:** `current-core`; `legacy-platform-b`; `vision-deck`.
  Related: CR-STO-020, CR-STO-190.

### CR-STO-020 — Media taxonomy and deterministic performance model
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Tenants cannot choose storage, and the platform cannot price,
  place, or throttle it, unless every media kind has a deterministic model
  that is queryable before creation.
- **Requirement:** The platform MUST define an enumerated media taxonomy
  (replicated SSD/HDD, non-replicated, mirrored, host-local, and successors)
  where kind + size + block size deterministically compute a performance model
  (IOPS/bandwidth envelope, allocation granularity) exposed through a
  describe-model API before resource creation. Pricing inputs, QoS limits, and
  placement rules MUST derive from that same model, never from per-service
  side channels.
- **Acceptance evidence:** conformance tests assert model determinism (same
  inputs produce the same model across control-plane instances); API tests
  verify describe-before-create for every offered kind; fixtures prove billing
  and QoS subsystems consume the model; the taxonomy is published as
  user-facing documentation.
- **Non-goals:** guaranteeing that observed performance equals the model under
  all failure modes — the model is an envelope, not a measurement.
- **Non-claims:** model accuracy against measured real-world performance is
  unvalidated; no production pricing derivation exists yet.
- **Stop conditions:** money — halt any billing derivation not traceable to
  the published model; trust — halt the launch of a media kind whose model
  cannot be computed deterministically.
- **Traceability:** `legacy-platform-b`; `current-core`. Related: CR-STO-030,
  CR-STO-170.

### CR-STO-030 — Per-volume QoS with token-bucket limits and burst credits
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Without enforced per-volume limits, one tenant's IO storm
  degrades every neighbor; without documented burst mechanics, tenants
  experience throttling as arbitrary platform behavior.
- **Requirement:** Every volume MUST carry token-bucket QoS limits (IOPS and
  bandwidth) derived from its performance model and enforced in the data path.
  Burst/boost credit mechanics (accrual rate, spend, cap) SHOULD be supported
  for burstable classes and, when offered, MUST be documented with precise
  user-visible semantics. Throttle events MUST be observable per volume
  through metrics and counters.
- **Acceptance evidence:** load tests prove enforcement at configured limits
  within a defined measurement window; a noisy-neighbor test shows bounded
  degradation for co-located tenants; burst-credit tests verify accrual/spend
  math against the published documentation; per-volume throttle metrics exist
  and are scraped.
- **Non-goals:** application-level caching or queue management; per-request
  fairness inside a single volume.
- **Non-claims:** enforcement overhead under full line-rate load is
  unmeasured; burst parameters are not yet tuned against real workloads.
- **Stop conditions:** trust/exposure — halt a media-class launch if QoS
  cannot be enforced in its data path; data — halt immediately if the
  throttling implementation is shown to drop or corrupt IO instead of delaying
  it.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-020, CR-STO-200.

### CR-STO-040 — Mount fencing, access modes, and session lifecycle
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** Double-attaching a block volume from two writers corrupts data
  silently, while stale sessions after a crash can block legitimate remounts
  indefinitely.
- **Requirement:** Attach and mount MUST be fenced with monotonic sequence
  numbers so that a higher sequence invalidates all prior sessions. Mounts
  MUST support access modes (read-write, read-only, administrative/repair
  read-only) and MUST bind volume to instance with a short-lived token issued
  by the control plane. Sessions MUST have an inactivity timeout with
  deterministic reclaim, and clients SHOULD automatically re-establish
  sessions across transient failures. Wrong-key or wrong-mode mounts MUST fail
  fast and be audited.
- **Acceptance evidence:** fault-injection tests attempt double-attach and
  verify the fenced writer is rejected; session-timeout tests verify reclaim
  and clean remount; access-mode tests verify read-only enforcement at the
  data path; chaos tests show client auto-remount within a bounded time.
- **Non-goals:** multi-writer shared block attach (shared access is provided
  only through the filesystem class, CR-STO-190).
- **Non-claims:** fencing correctness is proven only in design and tests; no
  production double-attach drill has been executed.
- **Stop conditions:** data — any detected double-attach or fencing regression
  halts the affected volume class and triggers incident review before further
  mounts; keys — mounts presenting mismatched encryption-key identities MUST
  be denied and audited.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-050, CR-STO-120.

### CR-STO-050 — Encryption at rest with envelope keys
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, auditor
- **Problem:** Multi-tenant media, retired devices, and backup chunks all leak
  tenant data unless storage is encrypted at rest with properly separated and
  referenced keys.
- **Requirement:** Volumes, snapshots, and backup chunks MUST be encrypted at
  rest with per-resource data-encryption keys; data keys MUST be wrapped by
  key-encryption keys held in the approved key-management workflow (envelope
  encryption), referenced by identity, and never stored alongside the data.
  Mount and restore paths MUST verify key identity (e.g., a key hash) and fail
  closed on mismatch. Decommissioned media MUST be cryptographically or
  securely erased, with verification evidence, before reuse or release.
- **Acceptance evidence:** contract tests reject any configuration carrying
  raw key material; integration tests verify wrong-key mount/restore denial;
  a key-rotation drill re-wraps keys without data unavailability; secure-erase
  verification records exist for decommissioned devices.
- **Non-goals:** the key-management service itself (owned by the identity and
  security domain); in-guest or application-level encryption.
- **Non-claims:** encryption performance overhead is unmeasured; rotation at
  fleet scale is undrilled.
- **Stop conditions:** keys — any code path that would persist, log, or
  transmit an unwrapped data key halts immediately and is treated as a
  security incident; data — an erase-verification failure blocks device reuse
  or release.
- **Traceability:** `legacy-platform-b`; `current-core`. Related: CR-STO-120,
  CR-STO-150.

### CR-STO-060 — Crash-consistent checkpoints and changed-block tracking
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, operator
- **Problem:** Snapshots that stall workloads or copy full disks every time
  are unusable as a routine data-protection primitive, so teams skip them.
- **Requirement:** The storage layer MUST support cheap crash-consistent
  checkpoints that let the workload resume immediately while data copy-out
  proceeds in the background. It MUST track changed blocks between checkpoints
  so that incremental snapshots and backups copy only deltas.
  Application-consistent capture SHOULD be supported through pre/post hooks
  coordinated with the backup control plane.
- **Acceptance evidence:** tests show checkpoint latency is bounded and
  independent of volume size (workload resume measured); incremental-copy
  tests verify changed-block correctness by comparing restored data against
  the source; hook tests demonstrate an application-consistent flow for a
  reference stateful workload.
- **Non-goals:** continuous replication (deferred to the migration workflow,
  CR-STO-180, and federation scenarios).
- **Non-claims:** changed-block tracking correctness across all edge cases
  (resize, discard, encryption boundary) is not fully proven; background-copy
  impact on QoS is unmeasured.
- **Stop conditions:** data — any inconsistency found between the
  changed-block map and actual volume data halts incremental operations for
  the affected class and forces a full-copy fallback until root-caused.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-070, CR-STO-080,
  CR-STO-230.

### CR-STO-070 — Content-addressed snapshot chunk store with integrity verification
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Naive full-copy snapshots are unaffordable at fleet scale;
  deleting one snapshot must never destroy data shared with another, and
  restored data must be provably intact.
- **Requirement:** Snapshot and backup data SHOULD be stored as
  content-addressed, compressed chunks with reference counting so identical
  blocks are stored once within a tenant lineage. Every chunk MUST carry a
  checksum verified on write and on read/restore; a failed verification MUST
  fail the operation closed. Deletion MUST only unreference chunks, with
  asynchronous garbage collection that is safe against concurrent creation,
  and MUST report the reclaimed storage size.
- **Acceptance evidence:** unit and integration tests for refcount correctness
  including delete-while-creating races; restore-integrity tests inject chunk
  corruption and verify fail-closed behavior; GC tests prove no live chunk is
  collected and orphaned chunks are reclaimed; deletion size-reporting tests.
- **Non-goals:** cross-tenant deduplication (isolation forbids it); guaranteed
  compression or dedup ratios.
- **Non-claims:** achieved dedup ratios and GC behavior at production scale
  are unmeasured; corruption-injection drills have not yet run against a live
  store.
- **Stop conditions:** data — a checksum mismatch on restore halts the
  restore, quarantines the chunk set, and escalates; deletion — any GC
  safety-failure signal halts garbage collection fleet-wide until reviewed.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-060, CR-STO-120.

### CR-STO-080 — Platform backup control plane
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, tenant
- **Problem:** Per-team backup scripts rot and skip critical state; a provider
  platform needs one backup control plane covering cluster resources and
  volume data as a default, not an opt-in.
- **Requirement:** The platform MUST provide a backup control plane of the
  Velero class — scheduled and on-demand capture of Kubernetes resources plus
  volume data — with declarative schedules, retention policies, pre/post
  hooks, and storage-location references. Backup MUST be a platform default
  for every stateful workload class. All backup and restore operations MUST be
  idempotent, audited, and emit evidence receipts.
- **Acceptance evidence:** the reference installation runs scheduled backups
  of platform and tenant workloads; conformance checks verify every stateful
  module declares a backup policy reference; end-to-end tests restore a
  namespace and a volume through the control plane; audit trail and receipts
  are emitted per operation.
- **Non-goals:** database-native point-in-time recovery internals (deferred to
  the data-services domain); cross-provider replication (federation domain).
- **Non-claims:** backup success at production data volumes is unproven;
  restore behavior under control-plane degradation is untested.
- **Stop conditions:** data — backup-job failures for production namespaces
  MUST alert and block the dependent mutation barrier (CR-STO-090); trust —
  a backup location failing integrity checks is quarantined from new writes.
- **Traceability:** `current-core`; `legacy-platform-a`; `legacy-platform-b`.
  Related: CR-STO-090, CR-STO-100, CR-STO-110.

### CR-STO-090 — Signed backup barrier before any mutation
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, agent, auditor
- **Problem:** Upgrades, migrations, shrinks, and deletions destroy data when
  executed against workloads whose last backup is missing, failed, or stale.
- **Requirement:** No mutating operation on a stateful resource (upgrade,
  migration, shrink, delete, reconfiguration, restore-overwrite) MAY proceed
  without a fresh backup-barrier receipt: a machine-checkable attestation,
  signed by the backup control plane's workload identity, binding the resource
  identity, the completed backup reference with its integrity hash, and a UTC
  timestamp within the mutation's freshness window. The mutation gate MUST
  fail closed when the receipt is missing, stale, unsigned, forged, or refers
  to a failed or unverified backup.
- **Acceptance evidence:** gate tests prove mutation denial for the missing,
  stale, forged, and failed-backup receipt cases; receipt schema and verifier
  run in CI; the reference upgrade runbook demonstrates barrier issuance and
  consumption; an auditor can replay barrier decisions from the append-only
  record.
- **Non-goals:** replacing human approval for dangerous operations — the
  barrier is an additional machine gate, not a substitute.
- **Non-claims:** barrier coverage across all mutation paths is incomplete;
  clock-skew handling for freshness windows is untested in production.
- **Stop conditions:** data/migration/deletion — a missing or invalid barrier
  halts the mutation and records a `blocked` evidence state; any attempt to
  bypass the gate is treated as a security event.
- **Traceability:** `current-core`; `legacy-platform-a`. Related: CR-STO-080,
  CR-STO-100, CR-STO-140.

### CR-STO-100 — Restore drills as promotion blockers
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, auditor
- **Problem:** Backups that are never restored are hopes, not protection;
  services reach production with unverified recoverability.
- **Requirement:** Every stateful service MUST execute scheduled restore
  drills that restore real backups into an isolated environment and verify
  data integrity and service health. A stateful service MUST NOT be promoted
  to production use without fresh `verified` restore-drill evidence
  within the declared freshness window; `blocked`, `stale`, and `synthetic`
  drill results are non-promotable states. Drill results MUST be recorded
  under the durability evidence states of CR-STO-140.
- **Acceptance evidence:** drill automation exists and runs on schedule;
  promotion-gate tests reject services with missing, stale, blocked, or
  synthetic drill evidence; drill receipts record restored-dataset integrity
  verification; at least one full-platform restore drill on the reference
  installation.
- **Non-goals:** dictating drill internals per service — services define their
  verification; the platform defines the gate.
- **Non-claims:** drill cadence under production data volumes is unproven; no
  production restore drill has yet been executed.
- **Stop conditions:** data/trust — a failed drill blocks promotion and alerts
  the owning team; repeated drill failure for a production service triggers a
  provider-level risk review before the service may keep its readiness claim.
- **Traceability:** `current-core`; `req-history`; `legacy-platform-a`.
  Related: CR-STO-090, CR-STO-110, CR-STO-140.

### CR-STO-110 — Per-service RPO/RTO objectives declared and enforced
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, provider, tenant, auditor
- **Problem:** Without declared recovery objectives, backup schedules and
  drill cadences are arbitrary, and tenants cannot judge whether a service
  meets their needs.
- **Requirement:** Every stateful service MUST publish machine-readable RPO
  and RTO objectives in its durability profile, and the platform MUST verify
  that the service's backup schedule, storage topology, and drill cadence can
  satisfy the declared objectives. Tenant-facing catalogs MUST surface these
  objectives. Changes to objectives MUST be versioned and audited.
- **Acceptance evidence:** conformance checks reject stateful modules lacking
  RPO/RTO declarations; consistency tests verify schedule and drill cadence
  against the declared objectives; catalog rendering tests show objectives to
  tenants; audit trail exists for objective changes.
- **Non-goals:** guaranteeing objectives under regional disasters beyond the
  declared scope; SLA compensation mechanics (commercial layer).
- **Non-claims:** declared objectives are not yet validated against measured
  restore times at scale.
- **Stop conditions:** money/trust — marketing or selling an objective tighter
  than the proven drill results MUST be halted; data — an objective change
  that weakens protection for existing tenants requires explicit tenant notice
  before it applies.
- **Traceability:** `current-core`; `legacy-platform-b`. Related: CR-STO-080,
  CR-STO-100.

### CR-STO-120 — Tenant and namespace storage isolation
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, auditor
- **Problem:** Shared storage backends make cross-tenant references, shared
  caches, and ambiguous ownership direct paths to data exposure.
- **Requirement:** Volumes, claims, snapshots, and backup locations MUST be
  scoped to a tenant and namespace; cross-namespace or cross-tenant references
  MUST be denied by default and allowed only through explicit, audited share
  grants. Encryption keys and dedup domains MUST be scoped per tenant.
  Namespace or tenant deletion MUST trigger the declared data lifecycle
  (export window, then deletion) with evidence. Shared base pools and platform
  images MUST be isolated from tenant-writable paths.
- **Acceptance evidence:** admission/policy tests reject cross-namespace
  references without grants; isolation penetration tests attempt cross-tenant
  access; deletion end-to-end tests verify lifecycle evidence; key-scope tests
  verify tenant separation; base-pool isolation tests.
- **Non-goals:** tenant-visible shared datasets (a future explicit grant
  product).
- **Non-claims:** the share-grant mechanism is designed but not implemented;
  isolation is validated only in test environments.
- **Stop conditions:** trust/exposure — any observed cross-tenant data path
  halts the affected storage class fleet-wide and triggers a security
  incident; deletion — lifecycle ambiguity (unknown or disputed owner) fails
  closed to retention, never to deletion.
- **Traceability:** `current-core`; `legacy-platform-b`. Related: CR-STO-050,
  CR-STO-070.

### CR-STO-130 — Control-plane state (etcd) snapshot and restore
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Losing cluster state means losing the platform itself;
  control-plane recovery must be a drilled, routine capability rather than
  tribal knowledge.
- **Requirement:** Every cluster MUST take scheduled etcd snapshots — with
  interval and retention declared in the cluster profile — stored to durable
  object storage with integrity verification, and MUST have a documented,
  tested restore procedure including quorum recovery. Restore drills MUST run
  at a declared cadence and feed durability evidence; snapshot cadence MUST
  satisfy the platform control-plane RPO.
- **Acceptance evidence:** the cluster profile schema carries snapshot
  interval, retention, and location; automated snapshot jobs emit
  success/failure evidence; a restore drill rebuilds a cluster from snapshot
  in an isolated environment and verifies workload and catalog state; evidence
  freshness follows CR-STO-140.
- **Non-goals:** etcd operational tuning itself; application-level restores
  (covered by CR-STO-080 and CR-STO-100).
- **Non-claims:** a restore drill has not yet been executed on the reference
  installation; quorum-loss recovery is untested.
- **Stop conditions:** data/migration — no control-plane upgrade or member
  replacement without a fresh verified snapshot; a failed snapshot job blocks
  control-plane mutations until resolved.
- **Traceability:** `legacy-platform-a`; `legacy-platform-b`; `current-core`.
  Related: CR-STO-090, CR-STO-140.

### CR-STO-140 — Durability evidence states and freshness
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, auditor, agent
- **Problem:** Durability claims inflate silently: old backup successes are
  quoted months later and blocked evidence gets laundered into readiness.
- **Requirement:** All durability evidence (backup completions, restore
  drills, integrity checks, etcd snapshots, erase verifications) MUST be
  recorded as first-class, append-only records with UTC timestamps and one of
  the states `verified`, `blocked`, `stale`, or `synthetic`. Each evidence
  class MUST declare a freshness window; evidence older than its window is
  `stale` and non-promotable. Only `verified`, non-synthetic, fresh evidence
  MAY support readiness or promotion claims; `blocked` evidence MUST remain
  visible and MUST never be converted into a claim.
- **Acceptance evidence:** evidence schema plus validator in CI; promotion
  gates reject stale, blocked, synthetic, or absent evidence with typed
  reasons; an append-only evidence store or ledger supports audit queries;
  dashboards render evidence state per service.
- **Non-goals:** the general metrics/logs pipeline — this is the claim-grade
  evidence layer, not observability plumbing.
- **Non-claims:** the evidence pipeline is not yet implemented across all
  classes; freshness windows are initial estimates pending operating
  experience.
- **Stop conditions:** trust — any tooling path that rewrites or deletes
  evidence history halts release until fixed; exposure — evidence containing
  secrets, tenant data, or endpoints is rejected at ingestion.
- **Traceability:** `current-core`; `req-history`. Related: CR-STO-090,
  CR-STO-100, CR-STO-130.

### CR-STO-150 — Maintenance interlock for storage-dependent drains
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Draining a host or device that still serves volumes causes
  outages and data loss; maintenance automation must ask storage before
  evicting hardware.
- **Requirement:** Before draining or decommissioning any host or device, the
  maintenance system MUST query the storage layer for dependent volumes and
  MUST support dry-run evaluation. While dependencies remain, storage MUST
  answer with a retryable refusal (try-again with backoff guidance) and SHOULD
  initiate migration of affected data where supported. Device suspend and
  resume MUST be explicit operations, and decommissioned devices MUST complete
  secure-erase verification before release.
- **Acceptance evidence:** interlock API with dry-run and try-again semantics
  covered by contract tests; a drain drill moves a loaded host into
  maintenance without workload disruption, with evidence; erase-verification
  records for decommissioned devices; maintenance-automation integration
  tests.
- **Non-goals:** the maintenance orchestration system itself (operations
  domain); live VM migration policy (compute domain).
- **Non-claims:** the interlock is proven only in design and tests; no
  production drain has been executed.
- **Stop conditions:** data/migration — a drain request against non-empty
  dependencies without a migration plan MUST be refused; an erase-verification
  failure blocks device release; deletion — forced-removal paths require
  elevated two-party approval and are audited.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-050, CR-STO-210.

### CR-STO-160 — Offsite immutable backup copies with retention governance
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, auditor, operator
- **Problem:** Backups in the same failure domain, deletable by the same admin
  plane, do not survive site loss, ransomware-class attacks, or malicious
  insiders.
- **Requirement:** The backup control plane SHOULD support copying backup sets
  to an offsite location (an independent failure domain) with immutability
  enforcement of the object-lock class for a declared window; within the
  window, deletion or modification MUST be impossible for every platform
  identity, including administrators. Retention and immutability policy MUST
  be declared per data class and audited, and restoring from the offsite tier
  MUST be drilled at a declared (possibly reduced) cadence.
- **Acceptance evidence:** policy schema for offsite and immutable classes;
  tests proving delete attempts within the window fail even with
  administrative credentials; offsite restore drill evidence; audit records of
  retention-policy changes.
- **Non-goals:** legal-hold workflows beyond retention windows;
  multi-provider backup brokering (federation domain).
- **Non-claims:** the immutable-tier implementation is not yet selected or
  tested; offsite restore RTO is unmeasured; the cost model is unvalidated.
- **Stop conditions:** data/deletion — any path allowing in-window deletion
  halts the feature; trust — immutability-configuration drift (a shortened
  window) requires elevated approval and tenant-visible audit.
- **Traceability:** `legacy-platform-a`; `legacy-platform-b`. Related:
  CR-STO-080, CR-STO-140.

### CR-STO-170 — Non-replicated and local media with enforced restrictions
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** High-performance local and non-replicated media is genuinely
  needed, but offering it without hard restrictions invites a silent mismatch
  between tenant durability expectations and reality.
- **Requirement:** The platform MAY offer non-replicated and host-local media
  classes; when offered, the API and documentation MUST enforce and disclose
  their restrictions: no platform-layer snapshots or backups, explicit
  data-loss-risk acknowledgement at creation, fixed allocation granularity
  where applicable, exclusion as the default boot or stateful class, and
  placement groups (rack/host anti-affinity) as the recommended mitigation.
  Metering MUST distinguish these classes from durable ones.
- **Acceptance evidence:** API tests reject snapshot requests on restricted
  classes; the creation flow requires explicit risk acknowledgement;
  placement-group spread tests; the restriction matrix is published in
  user-facing docs; metering class separation is verified.
- **Non-goals:** making non-replicated media durable at the platform layer —
  by definition it is not.
- **Non-claims:** the restriction UX is not yet validated with tenants; the
  local-media performance model is unmeasured.
- **Stop conditions:** data/trust — offering restricted media without enforced
  disclosure halts the launch; placement-group bypass for restricted classes
  is denied.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-020, CR-STO-030.

### CR-STO-180 — Cross-zone volume migration workflow
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Tenants need mobility between zones for resilience and
  capacity, but naive copy-and-switch loses writes or corrupts state mid-move.
- **Requirement:** Cross-zone volume migration MUST be an explicit two-phase
  workflow: replicate to the destination zone while the source keeps serving,
  then a deliberate finish/switchover step. The workflow MUST expose progress
  and ETA, MUST be safely cancellable before the declared commit point, MUST
  mark its point of no return, and MUST delete the source only after a
  confirmed successful switchover. A backup barrier (CR-STO-090) MUST precede
  the start.
- **Acceptance evidence:** migration end-to-end test across zones on the
  reference installation with data-integrity verification; cancellation tests
  at each pre-commit stage; progress and ETA surfaced in the operation API;
  post-switchover source-deletion audit records.
- **Non-goals:** live VM migration orchestration (compute domain);
  cross-provider migration (federation domain).
- **Non-claims:** migration of multi-hour large volumes is not yet exercised;
  write-fencing during switchover is proven only in tests.
- **Stop conditions:** migration/data — an integrity-verification failure at
  any stage aborts the migration and retains the source; exceeding the
  declared cutover window triggers rollback to the source and an incident
  review.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-060, CR-STO-090.

### CR-STO-190 — Shared filesystem service class
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, service-team, tenant
- **Problem:** Many workloads need multi-attach shared file storage; without a
  platform filesystem class, teams bolt on incompatible per-service solutions.
- **Requirement:** The platform SHOULD offer a shared filesystem class behind
  the storage contract (CR-STO-010) with multi-attach semantics, session
  management, byte-range locking where supported, online resize, and
  checkpoint integration with the snapshot/backup pipeline. Client attachment
  SHOULD use a versioned, kernel/virtio-friendly protocol path, with a gateway
  for clients lacking native support.
- **Acceptance evidence:** contract-conformant filesystem provisioning
  end-to-end; multi-client attach tests with concurrent IO; lock-semantics
  tests; resize-without-unmount tests; backup integration tests via
  checkpoints.
- **Non-goals:** POSIX completeness certification; a per-OS client support
  matrix beyond a documented baseline.
- **Non-claims:** the filesystem class is not yet implemented; lock and
  fencing behavior under network partition is untested; Windows-class clients
  are explicitly unsupported initially.
- **Stop conditions:** data — a lock-manager or fencing failure halts new
  attaches fleet-wide; exposure — filesystem exports MUST default to private
  network scopes with no public reachability.
- **Traceability:** `legacy-platform-b`; `legacy-platform-a`. Related:
  CR-STO-010, CR-STO-040, CR-STO-060.

### CR-STO-200 — Storage performance and fault-injection gates
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** Storage regressions are catastrophic and silent; performance
  and resilience claims need continuous automated proof rather than launch-day
  benchmarks.
- **Requirement:** Each media class and data path MUST have automated
  performance suites (fio-class workloads measuring latency, IOPS, and
  bandwidth under defined profiles) and fault-injection suites (node/device
  loss, network partition, control-plane outage) whose results are recorded as
  evidence. Performance or resilience regressions beyond declared thresholds
  MUST block rollout of the offending change.
- **Acceptance evidence:** performance and fault-injection suites run in CI or
  on scheduled stands; results are recorded with thresholds; rollout-gate
  tests reject regressions; at least one nemesis-class suite covers the volume
  data path.
- **Non-goals:** absolute performance leadership — thresholds are honesty
  gates, not marketing targets.
- **Non-claims:** thresholds are initial and uncalibrated against production
  load; fault-injection coverage is incomplete.
- **Stop conditions:** data — a fault-injection finding of data loss or
  corruption halts the affected class until root-caused and fixed; trust —
  disabled or skipped suites block readiness claims.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-030, CR-STO-040,
  CR-STO-140.

### CR-STO-210 — Operator storage toolset and capacity rebalancing
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Incidents need vetted tools on day one — garbage pruning,
  metadata repair, device lifecycle — and capacity hotspots need rebalancing
  without drama.
- **Requirement:** The platform SHOULD ship first-class operator tooling for
  storage garbage inspection and pruning, metadata consistency check and
  repair, device register/suspend/resume/erase, volume reallocation and
  rebalancing across nodes, and limits management. Every tool MUST default to
  dry-run, require explicit confirmation for mutation, be idempotent where
  possible, and emit audit and evidence records.
- **Acceptance evidence:** the toolset exists with dry-run-by-default tests;
  audit records accompany every mutating invocation; a rebalancing drill moves
  data between nodes without workload-impact evidence; repair tooling is
  exercised inside a fault-injection scenario.
- **Non-goals:** autonomous remediation without operator confirmation
  (agent-governed automation is a separate domain).
- **Non-claims:** toolset completeness is unproven — the first real incident
  will find gaps; rebalancing IO impact on QoS is unmeasured.
- **Stop conditions:** deletion/data — pruning and repair tools MUST refuse
  destructive action without a fresh backup barrier for affected resources
  (CR-STO-090); any forced destructive path requires elevated approval.
- **Traceability:** `legacy-platform-b`; `legacy-platform-a`. Related:
  CR-STO-150, CR-STO-090.

### CR-STO-220 — Data-path endpoint lifecycle and transport fallback
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** Hypervisor diversity and transport failures make a single
  hardwired IO path fragile, and future high-performance tiers need alternate
  transports.
- **Requirement:** The storage data path SHOULD expose endpoint lifecycle
  operations (start, stop, list, health) and support at least one fallback
  transport (a userspace block protocol) beside the primary VM attachment
  path; the transport type of each attachment SHOULD be recorded for fleet
  telemetry. High-performance transports of the RDMA/NVMe-oF class MAY be
  added later as a distinct media class with its own performance model.
- **Acceptance evidence:** endpoint lifecycle API tests; a fallback-transport
  drill disables the primary path and confirms IO continues via fallback;
  transport telemetry fields present; if a high-performance class is added,
  its model and QoS tests follow CR-STO-020 and CR-STO-030.
- **Non-goals:** guest-initiated transports; exposing raw device access to
  tenants.
- **Non-claims:** the fallback transport is not yet implemented; the
  RDMA-class path is exploration only, with no performance data.
- **Stop conditions:** data — transport failover MUST preserve fencing
  (CR-STO-040); any failover path that can split writers is blocked.
- **Traceability:** `legacy-platform-b`. Related: CR-STO-040, CR-STO-020.

### CR-STO-230 — Changed-block API for external backup integration
- **Priority:** P2
- **Status:** proposed
- **Actors:** vendor, service-team, provider
- **Problem:** Third-party backup products cannot integrate efficiently
  without a supported delta-read API, and scraping internal checkpoints
  creates fragility for everyone.
- **Requirement:** The platform MAY expose a versioned, read-only
  changed-block API enabling authorized external backup systems to consume
  checkpoint deltas, with tenant-scoped authorization, rate limits, and audit.
  The API MUST NOT allow mutation of checkpoints, volumes, or chunk state.
- **Acceptance evidence:** API contract plus conformance tests; authorization
  tests covering tenant scoping and denial paths; audit records of
  consumption; a reference integration test with a mock external consumer.
- **Non-goals:** certifying specific third-party backup products; write-path
  integration for external vendors.
- **Non-claims:** no external consumer has integrated; API stability and abuse
  resistance are unproven.
- **Stop conditions:** exposure/keys — any cross-tenant delta visibility or
  credential weakness halts the API; data — the API failing open to mutation
  is an immediate security incident.
- **Traceability:** `legacy-platform-b`; `vision-deck`. Related: CR-STO-060,
  CR-STO-120.

---

## Coverage notes

This domain deliberately defers the following to sibling domains:

- **Image lifecycle** (base images, families, per-zone image pools,
  overlay-based fast provisioning) is deferred to CMP (compute); this domain
  owns only the checkpoint/snapshot mechanics images are built on.
- **Object storage as a tenant-facing product** (S3-compatible buckets,
  lifecycle rules, quotas) is an OCS service module deferred to DAT/OCS; this
  domain owns the durability, isolation, and backup obligations those modules
  inherit.
- **Database-native protection** (point-in-time recovery, WAL archiving,
  dump/restore orchestration) is deferred to DAT; this domain owns the volume
  and backup primitives beneath it.
- **Key management, rotation policy, and secrets brokering** are deferred to
  IAM; this domain only consumes envelope keys and verifies key identity at
  mount/restore/erase boundaries.
- **Metrics, logs, traces, dashboards, and alert routing** are deferred to
  OBS; this domain defines which storage signals must exist, not the
  observability pipeline.
- **Fleet maintenance orchestration** (drain scheduling, upgrade waves) is
  deferred to OPS; this domain owns only the interlock contract the
  orchestrator must honor (CR-STO-150).
- **GitOps/IaC delivery** of storage components and cluster profiles is
  deferred to DPL.
- **Storage metering rate cards, pricing, and settlement** are deferred to BIL
  and FED; this domain guarantees metering classes exist and are
  distinguishable, not what they cost.
- **Cross-provider replication and federated backup exchange** are deferred to
  FED.
