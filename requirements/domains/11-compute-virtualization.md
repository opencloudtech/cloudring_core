# 11 — Compute & Virtualization

This domain covers the virtual-machine compute substrate of the Cloud
Infrastructure Pod: the tenant-facing instance lifecycle (create, start, stop,
restart, resize, delete) executed as durable, idempotent long-running
operations with status polling; the pluggable virtualization-backend contract
with KubeVirt on upstream Kubernetes as the primary OSS driver; bare-metal
provisioning through out-of-band management; the golden-image pipeline and
image distribution mechanics (per-zone base disks, overlay provisioning,
content-addressed storage); placement/affinity and live-migration policy;
console and serial access; the instance metadata and workload-identity
service; compute quota reservation; and garbage collection of half-created
resources.

Block-storage durability semantics, virtual networking, Kubernetes cluster
products, and usage billing are sibling domains and are deliberately excluded
here (see Coverage notes).

**Domain contract.** Every mutating compute operation is asynchronous,
durable, idempotent, and observable; a synchronous API never hides a
long-running workflow. Driver-observed state wins over control-plane
bookkeeping, and ambiguity fails closed. No tenant data is destroyed outside
an explicit, audited deletion path. Quota is reserved before capacity is
consumed and released on failure. Secrets reach guests only through the
brokered metadata/identity path, never through configuration. No capability —
migration, overlay provisioning, accelerators — is claimed beyond what its
evidence (conformance suites, drills, benchmarks) currently proves.

## Requirements

### CR-CMP-010 — Instance lifecycle API with explicit state machine
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent
- **Problem:** Tenants need one stable, vendor-neutral API to run virtual
  machines. Without an explicit, versioned lifecycle contract, clients cannot
  distinguish transitional from terminal states or automate safely against
  them.
- **Requirement:** The platform MUST expose a compute API covering create,
  get, list, start, stop, restart, resize (CPU/memory), and delete of
  instances, governed by an explicit, versioned state machine (e.g.
  PROVISIONING, RUNNING, STOPPING, STOPPED, STARTING, UPDATING, ERROR,
  DELETING) in which user-triggered and system-triggered transitions are
  distinguishable. Instance status MUST reflect driver-observed reality, never
  optimistic bookkeeping; states that cannot be determined MUST surface as
  ERROR/UNKNOWN. All lifecycle requests MUST accept a client-supplied
  idempotency key. Stop SHOULD default to graceful guest shutdown with a
  documented timeout before power-off; forced power-off MUST be an explicit,
  recorded choice.
- **Acceptance evidence:** API contract tests covering every legal and
  illegal state transition; idempotency-replay tests proving duplicate
  create/start/delete requests return the original operation without duplicate
  side effects; e2e suite exercising the full lifecycle on the reference
  installation; status-truthfulness tests injecting driver failures.
- **Non-goals:** autoscaling groups, instance fleets, and scheduling policy
  products; guest OS internals.
- **Non-claims:** validated so far against a single backend driver; resize
  without reboot is not guaranteed for all shapes.
- **Stop conditions:** any delete path MUST halt and escalate when the
  instance has attached volumes whose deletion was not explicitly requested,
  when policy-required backup evidence is missing, or when driver-observed
  state contradicts control-plane records (deletion, data risk).
- **Traceability:** legacy-platform-a (instance FSM and lifecycle API),
  legacy-platform-b (status taxonomy, one operation per mutation),
  current-core (upstream Kubernetes runtime policy), vision-deck.

### CR-CMP-020 — Long-running operation model
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent
- **Problem:** Compute operations take seconds to hours. Synchronous APIs
  force clients to guess, time out, or retry blindly, producing duplicates
  and orphaned work.
- **Requirement:** Every mutating compute RPC MUST return an Operation object
  immediately, carrying a stable operation ID, the target resource, typed
  metadata (phase, progress where measurable, ETA where computable), and a
  typed result or error on completion. Operations MUST be listable and
  pollable per resource and per project, retained for an auditable window,
  and cancellable up to an explicitly declared point of no return per
  operation type. Read APIs stay synchronous.
- **Acceptance evidence:** contract tests asserting every mutating endpoint
  returns an Operation with typed metadata/response; cancellation tests per
  operation type proving pre-point-of-no-return cancels leave no partial
  state and post-point cancels are rejected with a documented error;
  operation-retention checks; portal/agent integration tests polling
  operations to terminal state.
- **Non-goals:** a general tenant-facing workflow engine for arbitrary
  pipelines.
- **Non-claims:** ETA accuracy is uncalibrated for large-disk and cross-zone
  operations.
- **Stop conditions:** an operation exceeding its declared deadline without
  progress MUST transition to a halted/attention state and page an operator
  rather than retrying silently; repeated identical failures MUST trip a
  circuit breaker instead of amplifying side effects (migration, data risk).
- **Traceability:** legacy-platform-a (async facade, operations log, status
  endpoints), legacy-platform-b (operation-per-RPC pattern, cancellability
  classes), req-acr-singular (evidence envelopes), req-history.

### CR-CMP-030 — Durable, idempotent task execution framework
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, agent
- **Problem:** Control-plane crashes mid-provisioning must neither leak
  resources nor corrupt state; every retry of a partially executed workflow
  must be safe.
- **Requirement:** All multi-step compute workflows (create-from-image,
  resize, migrate, delete-with-attachments, GC) MUST execute as durable tasks
  with serialized state in a reliable store, resumable step-wise execution
  with per-step save points, end-to-end idempotency keys propagated from
  client headers, task dependencies, and declarable non-cancellable points.
  After a control-plane failover, in-flight tasks MUST resume or roll forward
  without operator intervention. Task identity MUST be journaled onto
  resource metadata (creating task ID, request fingerprint) so orphaned work
  is provably attributable (see CR-CMP-170).
- **Acceptance evidence:** fault-injection suite killing the control plane at
  each step boundary and proving resume-or-reconcile with zero duplicate side
  effects; idempotency-key conflict tests; chaos-drill evidence on the
  reference installation; persistence checks proving no in-memory-only
  production task state.
- **Non-goals:** exposing the task framework as a tenant-facing product.
- **Non-claims:** proven only for the primary driver; composite tasks
  spanning two different backends are not yet exercised.
- **Stop conditions:** on task-store corruption, replay ambiguity, or an
  idempotency-key conflict with differing payloads, the affected shard MUST
  halt admission of new tasks and escalate — never guess which side effects
  already ran (data, money risk).
- **Traceability:** legacy-platform-b (durable task framework, step save
  points, non-cancellable error classes), legacy-platform-a (async resource
  manager lessons), current-core (no in-memory production state).

### CR-CMP-040 — Virtualization backend abstraction, KubeVirt primary
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** Hard coupling to one hypervisor stack locks providers to a
  single vendor and violates the contract-before-technology principle that
  lets the platform survive technology rotation.
- **Requirement:** Compute MUST isolate all hypervisor interactions behind a
  versioned, vendor-neutral driver contract: instance lifecycle, disk/network
  attachment hooks, status observation, console proxy, migration hooks, and a
  capability descriptor. The OSS reference driver MUST be KubeVirt on
  upstream Kubernetes semantics; additional drivers MAY exist behind the same
  contract. The tenant-facing resource model (instance, flavor, image
  reference, attachments) MUST NOT expose driver-specific concepts; driver
  specifics live only inside driver configuration.
- **Acceptance evidence:** interface contract tests runnable against any
  driver; the KubeVirt reference driver passing the full lifecycle e2e suite
  on the reference installation; static checks that tenant-facing APIs carry
  no driver-specific fields; a second (minimal or null) driver compiling and
  passing contract tests, proving substitutability.
- **Non-goals:** parity for every hypervisor on day one; container workload
  orchestration itself (K8S domain).
- **Non-claims:** only the KubeVirt driver is on the production path;
  alternative drivers are unproven in production.
- **Stop conditions:** when a driver reports state contradicting
  control-plane records, or capability negotiation fails, the platform MUST
  fail closed — mark affected instances UNKNOWN and block destructive actions
  until reconciled (data, trust risk).
- **Traceability:** vision-deck (every layer replaceable behind a contract),
  legacy-platform-a (hypervisor-coupling lesson, backend-interface split),
  current-core (Go-first, upstream Kubernetes policy).

### CR-CMP-050 — Driver capability descriptor and conformance suite
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, service-team, vendor
- **Problem:** Backends differ (migration, snapshots, local disks,
  passthrough, console types). Hiding those differences creates false tenant
  expectations and outages; advertising them without proof creates false
  claims.
- **Requirement:** Each driver MUST publish a machine-readable capability
  descriptor: supported lifecycle features, migration support, disk locality
  classes, console types, passthrough classes, and per-feature maturity. The
  control plane MUST gate requests against capabilities and return explicit
  unsupported-feature errors, never silently degrade. A versioned driver
  conformance suite MUST exist and pass in CI per driver before that driver
  may be enabled for tenants; the public support matrix SHOULD be generated
  from descriptors and verified against test outcomes.
- **Acceptance evidence:** conformance results per driver per release;
  contract checks that API validation consults the descriptor; support-matrix
  documentation verified against test outcomes.
- **Non-goals:** normalizing performance tiers across drivers.
- **Non-claims:** capability maturity labels beyond "supported in tests" are
  not yet backed by production evidence.
- **Stop conditions:** a driver failing conformance, or advertising a
  capability its tests do not prove, MUST be blocked from production
  enablement (trust risk).
- **Traceability:** legacy-platform-a (per-substrate feature drift lessons),
  current-core (conformance gating), vision-deck (evidence over claims).

### CR-CMP-060 — Bare-metal provisioning profile
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Providers need to offer dedicated bare-metal instances and to
  bootstrap hypervisor capacity itself, ideally behind the same lifecycle
  contract tenants already use.
- **Requirement:** The platform SHOULD support a bare-metal profile behind
  the same instance lifecycle API, provisioned through out-of-band management
  (IPMI/Redfish-class) with network-boot installation, hardware inventory
  ingestion, and firmware/baseline checks. Bare-metal instances MUST support
  the same metadata, console, and quota paths as virtual instances.
  Out-of-band credentials MUST be referenced from the approved secrets
  workflow, never stored in configuration, and traffic to management
  controllers MUST be isolated from tenant networks. Before reassignment to a
  new tenant, a node MUST undergo verified disk sanitization with recorded
  evidence.
- **Acceptance evidence:** bare-metal provisioning e2e on at least one
  hardware class in a lab stand (live evidence class); sanitization records
  per node reassignment; source-safety verification that no management
  credentials appear in configuration or VCS; network-isolation tests for
  management interfaces.
- **Non-goals:** full datacenter asset management; fleet-wide automated
  firmware patching.
- **Non-claims:** the validated hardware catalog is small; no broad
  multi-vendor out-of-band matrix is claimed.
- **Stop conditions:** repeated management-controller authentication failure,
  missing sanitization evidence, or inventory drift MUST quarantine the node
  and block assignment (keys, data, trust risk).
- **Traceability:** vision-deck (bare metal as first-class substrate),
  legacy-platform-a (dedicated-server product precedent, provisioning
  adapters), req-history (platform lifecycle).

### CR-CMP-070 — Golden-image pipeline
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, vendor
- **Problem:** Every instance boots from an image; untraceable, hand-built
  images destroy reproducibility and tenant trust.
- **Requirement:** Base images MUST be produced by declarative,
  version-controlled build pipelines (build recipe and configuration as
  code), yielding immutable, checksummed, signed artifacts with recorded
  provenance (source packages, build time, builder identity). Images MUST
  pass vulnerability and source-safety scans before publication, and
  instances MUST refuse to boot unsigned or unverified images unless an
  explicit, audited provider policy allows it. Image families SHOULD resolve
  "latest" deterministically per release channel. Tenant custom images MAY be
  imported (upload or URL) subject to the same verification path.
- **Acceptance evidence:** reproducibility checks (rebuild yields matching
  digest or documented drift); signature-verification tests at instance
  create time; blocked-boot tests for unsigned or tampered images; per-image
  provenance records; pipeline runs as CI evidence.
- **Non-goals:** marketplace presentation and commercial image licensing (MKT
  domain); per-OS hardening baselines (documented separately).
- **Non-claims:** measured-boot attestation into the guest is not yet
  implemented.
- **Stop conditions:** failing signature verification, a vulnerability-gate
  failure above severity threshold, or provenance gaps MUST halt publication
  and block new instance creation from that image (trust, exposure risk).
- **Traceability:** legacy-platform-a (versioned image factory pipelines),
  legacy-platform-b (image families, integrity verification),
  current-core (source-safety gates).

### CR-CMP-080 — Per-zone image distribution and overlay provisioning
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Instance-create latency and cross-zone image availability
  define perceived platform quality; copying full images on every create does
  not scale.
- **Requirement:** Images SHOULD be regional resources replicated to per-zone
  storage, where each zone keeps a managed pool of base disks per active
  image version. Create-from-image SHOULD provision copy-on-write overlay
  disks on the zone-local base for fast starts, with an explicit full-copy
  option for isolation- or performance-sensitive workloads. Base disks and
  pools MUST be isolated from tenant tenancy (dedicated ownership, no tenant
  write path), and pool sizing and retirement SHOULD be automated. When a
  base disk is corrupted or retired, dependent overlays MUST be rebased, or
  instances migrated, without tenant data loss.
- **Acceptance evidence:** create-from-image latency benchmarks per zone for
  pooled versus full-copy paths; isolation tests proving tenants cannot
  address or mutate base disks; rebase/retire drills with running instances;
  pool autoscaling metrics from the reference installation.
- **Non-goals:** cross-region image distribution economics; tenant-visible
  base-disk management.
- **Non-claims:** create-latency targets are benchmarked in test stands, not
  at production scale; the overlay path is proven only on the primary storage
  backend.
- **Stop conditions:** base-disk integrity failure MUST quarantine the pool
  and halt new overlay provisioning; any tenancy-isolation breach of shared
  bases MUST halt provisioning fleet-wide and escalate (data, trust risk).
- **Traceability:** legacy-platform-b (per-zone base pools, overlay disks,
  shared-base isolation lesson), legacy-platform-a (image distribution
  tooling), vision-deck.

### CR-CMP-090 — Content-addressed image store
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Storing a full copy of every image version in every zone
  wastes capacity, and ad-hoc deletion breaks dependents; integrity must be
  provable, not assumed.
- **Requirement:** The image store SHOULD chunk image data content-addressed
  with compression, reference-count shared chunks across image versions and
  zones, and verify per-chunk checksums on write and on read-out for instance
  creation. Deleting an image version MUST only decrement references;
  physical reclamation MUST be asynchronous and report reclaimed space. The
  store MUST provide integrity-audit tooling able to walk any image and prove
  every referenced chunk.
- **Acceptance evidence:** dedup-ratio measurements across image-version
  families on the reference stand; corruption-injection tests proving
  tampered chunks are detected before use; deletion tests proving dependents
  remain bootable and reclaimed-space accounting is exact; audit-tool run
  reports.
- **Non-goals:** snapshot and backup chunk storage policy and retention (STO
  domain); object-storage product semantics.
- **Non-claims:** dedup efficiency at multi-zone production scale is
  unmeasured.
- **Stop conditions:** checksum mismatch on read MUST quarantine the chunk
  and block its use; reference-count drift detected by audit MUST halt
  reclamation until reconciled — never reclaim under uncertainty (deletion,
  data risk).
- **Traceability:** legacy-platform-b (content-addressed chunks, refcounting,
  verify-on-restore), req-history (evidence discipline).

### CR-CMP-100 — Placement groups and affinity policy
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Tenants running replicated workloads need control over
  failure-domain spreading; providers need honest capacity semantics for
  constrained placement.
- **Requirement:** The compute API SHOULD support placement groups with at
  least a spread/anti-affinity strategy across hosts and, where topology
  permits, racks or fault domains, plus affinity/anti-affinity between
  instances. Placement constraints MUST be evaluated at create and at
  migration time; when placement cannot be satisfied, creation MUST fail with
  an explicit error rather than silently relaxing constraints. The capacity
  trade-off of strict placement MUST be documented, and group membership MUST
  be inspectable with observed topology evidence.
- **Acceptance evidence:** scheduling tests asserting spread across available
  failure domains; failure-injection tests (host or rack removal) verifying
  no group violates its policy; negative tests proving honest failure on
  unsatisfiable constraints; per-group topology evidence reports.
- **Non-goals:** dedicated-host tenancy products (possible later extension);
  NUMA/CPU-pinning performance tuning.
- **Non-claims:** rack-level fault domains depend on datacenter topology
  metadata that is only partially modeled.
- **Stop conditions:** topology-metadata drift (unknown host-to-fault-domain
  mapping) MUST block new strict-placement commits until corrected — never
  guess placement (trust, data-locality risk).
- **Traceability:** legacy-platform-b (spread placement groups, capacity
  trade-off lesson), legacy-platform-a (dedicated-host demand), vision-deck
  (portability).

### CR-CMP-110 — Host maintenance and evacuation safety
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, tenant
- **Problem:** Hypervisor hosts must be patched and retired; naive drains
  kill tenant workloads and destroy trust in the platform's operability.
- **Requirement:** The control plane MUST support cordoning a host and
  computing its dependent instances before any maintenance action (dry-run),
  evacuating them per instance policy — live-migrate where supported,
  otherwise governed stop-and-restart with recorded tenant or policy
  approval. Forced power-off as a maintenance shortcut MUST be an explicit,
  audited decision, never a default. Maintenance workflows MUST report
  progress per instance and complete or roll back within a declared window.
  One-server-loss behavior MUST be exercised as a recurring drill.
- **Acceptance evidence:** dry-run dependency reports matching observed
  state; maintenance-window e2e on the reference installation evacuating a
  loaded host; one-server-loss drill records (live evidence class) with
  recovery-time measurements; audit trail of any forced actions.
- **Non-goals:** zero-downtime guarantees for non-migratable configurations;
  datacenter-wide disaster recovery (STO/DPL domains).
- **Non-claims:** evacuation-time objectives under full production load are
  not yet established.
- **Stop conditions:** if evacuation cannot complete within the declared
  window, or affected instances include non-migratable workloads without
  recorded approval, the drain MUST halt and escalate; any unexpected
  instance loss during maintenance MUST freeze further actions fleet-wide
  pending review (data, trust risk).
- **Traceability:** legacy-platform-b (maintenance interlock with dry-run and
  try-again semantics), current-core (one-server-loss readiness gate),
  req-history (operations resilience).

### CR-CMP-120 — Live migration policy and evidence
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Non-disruptive maintenance and premium availability require
  moving running instances between hosts; unproven migration claims are worse
  than none.
- **Requirement:** Where the driver advertises it, the platform SHOULD
  support control-plane-driven live migration with a bounded
  memory-dirty-rate policy, a target maximum pause budget, automatic
  abort-and-rollback on breach, and explicit exclusion classes (passthrough
  devices, local-disk-only storage, tenant-declared non-migratable). Every
  migration MUST emit an evidence record: duration, achieved pause, bytes
  transferred, result. Cross-zone moves are separate governed workflows with
  replicate-then-switchover semantics, progress and ETA, and a safe
  cancellation window — never implied by same-host migration support.
- **Acceptance evidence:** migration drill evidence (per release and
  periodic): measured pause distributions against the declared budget under
  defined load; abort-path tests forcing dirty-rate overrun; exclusion-class
  enforcement tests; cross-zone workflow tests proving cancellation before
  switchover leaves the source authoritative.
- **Non-goals:** storage-only migration mechanics (STO domain);
  inter-provider migration (FED domain).
- **Non-claims:** the pause budget is a policy target, not a proven SLO;
  accelerator/passthrough migration is unsupported and not claimed.
- **Stop conditions:** breaching the pause or dirty-rate budget MUST abort
  and roll back the migration; migrating an excluded instance class MUST be
  refused up front; any post-migration data-inconsistency signal MUST halt
  further migrations and escalate (migration, data risk).
- **Traceability:** legacy-platform-b (live-migration semantics, exclusion
  classes, two-phase cross-zone move), vision-deck (portability),
  req-history.

### CR-CMP-130 — Console and serial access
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, operator, auditor
- **Problem:** Tenants need out-of-band access to broken guests (network
  down, misconfigured firewall), but console paths are a standing exposure
  surface.
- **Requirement:** The platform SHOULD provide serial/console access proxied
  through the control plane, gated by IAM authorization per instance, issued
  as short-lived single-purpose tokens, fully audited (who, when, source),
  and rate-limited. Console endpoints MUST NOT be exposed directly on
  hypervisor networks, and read-only versus interactive modes SHOULD be
  distinguishable in policy. Serial output retrieval (boot logs) SHOULD be
  available without an interactive session.
- **Acceptance evidence:** authorization tests (deny without grant, allow
  with scoped grant, token expiry enforced); audit records for every session;
  network-isolation tests proving no direct hypervisor exposure; e2e
  serial-output retrieval on the reference installation.
- **Non-goals:** in-guest remote-desktop products; multi-user collaborative
  sessions.
- **Non-claims:** keystroke-level session recording is not implemented and
  not claimed.
- **Stop conditions:** any authorization-bypass signal, token-validation
  error, or discovery of an unproxied console endpoint MUST disable the
  affected access path and escalate (exposure, trust risk).
- **Traceability:** legacy-platform-a (token-gated serial-console service),
  legacy-platform-b (serial output API), current-core (fail-closed security
  policy).

### CR-CMP-140 — Instance metadata service
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Guests need self-configuration (hostname, SSH keys, user-data)
  at first boot; the metadata path is also a classic cross-tenant leak
  vector.
- **Requirement:** Every instance MUST be able to reach a link-local,
  per-instance metadata service delivering instance identity, user-data, SSH
  public keys, and network configuration, compatible with standard
  guest-initialization tooling. Responses MUST be scoped strictly to the
  requesting instance (source-bound); user-data MUST be delivered exactly as
  supplied, with no undisclosed provider injection; and the service MUST fail
  closed — serve nothing rather than risk cross-instance disclosure. Secrets
  MUST NOT be placed in general metadata fields; they flow only through the
  workload-identity path (CR-CMP-150).
- **Acceptance evidence:** guest e2e on reference images proving key
  injection and user-data execution; isolation tests attempting
  cross-instance metadata access (spoofing, forwarding) with zero leakage;
  fail-closed tests under backend errors; source-safety checks that metadata
  carries no secrets.
- **Non-goals:** a general configuration-management channel; guest-agent
  functionality beyond metadata delivery.
- **Non-claims:** compatibility is verified against the reference image set
  only, not the full ecosystem of guest tooling.
- **Stop conditions:** any suspected cross-tenant metadata exposure MUST
  disable the metadata path for the affected scope and page security;
  repeated resolution or authorization errors MUST fail closed (keys,
  exposure, trust risk).
- **Traceability:** legacy-platform-b (metadata service as the guest delivery
  channel), legacy-platform-a (guest initialization via images and keys),
  current-core (fail closed).

### CR-CMP-150 — Workload identity for instances
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, service-team, provider
- **Problem:** Workloads on instances need platform API access without
  tenants embedding long-lived credentials in images or user-data.
- **Requirement:** The platform SHOULD issue short-lived, per-instance
  workload credentials through the metadata path, brokered from the approved
  secrets/identity workflow, bound to instance identity and tenancy context,
  rotated automatically, and revocable per instance. Token issuance MUST fail
  closed on identity-backend errors — never fall back to static or expired
  credentials. All issuance and use SHOULD be auditable.
- **Acceptance evidence:** e2e obtaining a workload token inside a guest and
  calling a platform API with least-privilege scope; expiry and rotation
  tests; revocation drills; audit-trail verification; negative tests proving
  no issuance on backend failure.
- **Non-goals:** federation of workload identity to third-party clouds
  (IAM/FED domains); in-guest secret stores.
- **Non-claims:** process-level identity within one instance is out of scope
  and not claimed.
- **Stop conditions:** identity-backend inconsistency, signing-key rotation
  incidents, or suspected token replay MUST halt issuance for the affected
  scope and escalate (keys, trust risk).
- **Traceability:** legacy-platform-b (metadata-delivered workload
  credentials), current-core (secrets are never configuration), vision-deck.

### CR-CMP-160 — Compute quotas with two-phase reservation
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Asynchronous provisioning races capacity; without held quotas,
  concurrent creates over-commit the fleet and billable resources appear
  without an accountable reservation.
- **Requirement:** The platform MUST enforce per-project compute quotas
  (instances, vCPU, memory, disk-by-class, accelerators where applicable)
  with a two-phase protocol: reserve before provisioning begins, commit on
  success, release on failure or cancellation, and automatic release on
  reservation expiry. Quota checks MUST fail closed when the quota service is
  unavailable — creation is denied, never admitted unreserved. Default quota
  templates per project tier SHOULD exist, and every reserve/commit/release
  decision MUST be auditable with the requesting identity.
- **Acceptance evidence:** concurrency tests racing creates against quota
  proving no over-commit; failure-path tests proving release on provisioning
  failure and on expiry; fail-closed tests with the quota service down; audit
  records per transition; reconciliation reports of reservations against
  observed fleet state.
- **Non-goals:** pricing, charging, and pay-account integration (BIL domain);
  quota-increase broker workflows (portal/OPS).
- **Non-claims:** enforcement is proven for the compute control plane's own
  provisioning paths; third-party service quotas follow the OCS contract (OCS
  domain).
- **Stop conditions:** quota-store unavailability MUST fail closed;
  reconciliation drift between reservations and observed usage beyond
  threshold MUST halt new reservations and escalate — quota errors are money
  errors (money risk).
- **Traceability:** legacy-platform-a (reserve/commit/release/rollback quota
  protocol, default templates), vision-deck (multi-tenant fairness),
  req-history.

### CR-CMP-170 — Garbage collection of half-created resources
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, agent
- **Problem:** Failed or interrupted provisioning leaks instances, disks,
  interfaces, addresses, and pool slots; leaks cost money, strand capacity,
  and confuse tenants.
- **Requirement:** Every resource created during provisioning MUST carry
  journal metadata (creating task ID, request fingerprint, tenancy), and a
  continuous reconciliation loop MUST detect resources whose creating task
  failed, was cancelled, or is unknown. Half-created resources MUST be
  quarantined, then removed by a governed cleanup path with graded levels
  (single orphan up to full project teardown), each level dry-runnable and
  audited. GC MUST never delete anything lacking positive journal proof of
  failed creation; ambiguous cases go to an operator queue, not the deleter.
- **Acceptance evidence:** fault-injection tests killing provisioning at each
  step and proving full reclamation; dry-run reports matching actual
  deletions; negative tests proving healthy journaled resources are never
  collected; reconciliation-latency metrics; operator-queue workflow tests.
- **Non-goals:** post-deletion data-shredding policy and backup retention
  (STO domain).
- **Non-claims:** GC coverage is proven for compute-owned resources;
  cross-domain orphans (network, storage) depend on those domains' journals.
- **Stop conditions:** journal corruption, reconciliation ambiguity, or a
  deletion batch exceeding a safety threshold MUST halt GC, require explicit
  operator approval, and escalate — deletion is fail-closed (deletion, money,
  data risk).
- **Traceability:** legacy-platform-b (create-request journaling,
  clear-deleted tasks), legacy-platform-a (orphan GC endpoints, graded
  cleanup levels), current-core (production honesty).

### CR-CMP-180 — Accelerator attachment model
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Accelerator-hungry tenants need equipped instances; ad-hoc
  passthrough breaks scheduling, migration, and honesty guarantees.
- **Requirement:** The platform MAY support accelerator passthrough as an
  explicit driver capability: accelerator-aware flavors, device inventory and
  health telemetry, and documented interaction with placement and migration
  (excluded from live migration unless the driver proves otherwise).
  Accelerator allocation MUST participate in the same quota and lifecycle
  contracts as other resources.
- **Acceptance evidence:** capability-descriptor entries validated by
  conformance tests; provisioning e2e on at least one accelerator-equipped
  host class (live evidence); utilization telemetry exported to the
  observability stack; a documented per-driver support matrix.
- **Non-goals:** fractional-device virtualization, in-guest driver
  management, AI platform services (DAT/MKT domains).
- **Non-claims:** only passthrough is considered; no live migration, no
  multi-tenant sharing of a single device, and no performance-isolation
  guarantees are claimed.
- **Stop conditions:** device-health telemetry failure or inventory drift
  MUST remove the device from scheduling until revalidated (trust risk).
- **Traceability:** legacy-platform-a (accelerator substrate demand,
  utilization telemetry), legacy-platform-b (migration exclusion classes),
  vision-deck (AI as a growth driver).

## Coverage notes

This domain deliberately defers the following to sibling domains:

- **NET (12):** VPC/subnet topology, network interfaces, IPAM, address
  allocation, security groups, NAT, and load balancing. Compute consumes
  network attachment hooks only; it does not define network semantics.
- **STO (13):** block-volume durability and replication, attach/detach
  fencing internals, snapshots, backup/restore, DR, and post-deletion data
  shredding. Compute references disks but does not define their data
  protection; snapshot chunk-store policy belongs to STO, not to the image
  store of CR-CMP-090.
- **K8S (14):** Kubernetes node infrastructure, cluster products, and
  container workloads. KubeVirt here is a VM driver, not a container product.
- **IAM (15):** authentication, authorization policy, and the secrets/identity
  workflow itself. CR-CMP-140/150 consume those services; they do not define
  them.
- **BIL (16):** usage metering, pricing, charging, and pay accounts derived
  from compute lifecycle events; quota pricing and billing of reservations.
- **FND (10):** the event spine, region/zone model, service chassis, and
  shared persistence conventions that the compute control plane builds on.
- **OBS (20):** fleet metrics, alerting, and dashboards; compute emits
  telemetry and evidence but does not own the observability stack.
- **OPS (21):** capacity management, patch cadence, SRE runbooks, and support
  workflows around maintenance windows.
- **DPL (22):** installation and IaC for hypervisor capacity and zone
  bring-up; this domain states runtime behavior, not deployment mechanics.
- **MKT (18):** marketplace presentation, commercial licensing, and
  third-party publication of images built per CR-CMP-070.
- **CUX (19):** portal presentation of instances, operations, and console
  sessions.
- **FED (23):** cross-provider instance portability, migration, and identity
  federation.
- **DAT (24):** data and AI platform services that consume accelerator
  instances from CR-CMP-180.
- **OCS (17) / AGT (25):** the connector contract for third-party services
  riding on compute, and agent-governance rules for automated operational
  actions (e.g. GC approval flows).
