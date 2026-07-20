# 24 — Data Services

Managed data services: relational databases, streaming platforms, and caches
operated by the platform on behalf of tenants.
Scope covers the shared control-plane pattern (desired-state metadb, durable
task execution, explicit lifecycle state machine, provider adapters), the
PostgreSQL reference engine (HA, wal-g-class backups, in-place upgrades),
Kafka-class streaming, Redis-class caching, and object-storage-backed
backup/restore.
Scope also covers supporting data-platform services (connection registry,
schema registry) and the metering of data-service usage, including backup
interval billing.
All data services onboard to the platform exclusively through OCS connector
packages: the platform dogfoods its own standard.
Tenant data is the highest-value asset in the system; durability, secrecy,
and deletion safety dominate every requirement in this domain.

## Domain contract

A data service is not ready unless all of the following hold: desired state
and every mutation flow through a revisioned metadb and a durable task queue
with a hard single-active-operation invariant per cluster; lifecycle is an
explicit state machine with paired error states and offline paths; backups
are scheduled, encrypted, object-storage-backed, and restore-tested on a
recurring drill; credentials and keys are brokered through the approved
secrets workflow and never stored or transmitted in plaintext; deletion is
staged with backup and confirmation barriers; usage is metered through a
durable outbox from the first day of public availability; and the service
reaches tenants only through a validated OCS connector package with no
privileged first-party hooks. On any doubt involving data, keys, money,
trust, exposure, deletion, or migration, the platform fails closed and
escalates to a human.

### CR-DAT-010 — Desired-state metadb with layered configuration and revision history

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, operator
- **Problem:** Managed data services need one authoritative record of what each tenant cluster should look like. Ad-hoc mutation of live infrastructure makes drift detection, partial-failure recovery, and audit impossible.
- **Requirement:** Every managed data service MUST persist desired state in a control-plane metadb as structured documents with layered inheritance (platform defaults → engine-class defaults → role → individual cluster, resolving to a computed target document), and MUST record every change to every entity in append-only revision tables with UTC timestamps. The metadb MUST itself be highly available and covered by the platform backup policy. The control plane MUST treat metadb content — not live infrastructure probes — as the source of truth for intent.
- **Acceptance evidence:** schema review plus unit tests of inheritance resolution; integration test proving each mutation writes a revision row and that history is never rewritten; metadb backup/restore drill evidence.
- **Non-goals:** Does not prescribe a specific metadb engine or document schema. Does not require exposing raw revision history to tenants.
- **Non-claims:** No production-grade metadb failover drill evidence yet. Layered-inheritance conflict behavior is unproven at fleet scale.
- **Stop conditions:** data / migration — halt and escalate on any metadb schema migration lacking a tested, reversible data-migration path; halt on detected divergence between recorded revisions and applied state.
- **Traceability:** legacy-platform-b; related CR-DAT-020, CR-DAT-030.

### CR-DAT-020 — Durable task queue with dependencies and single-active-operation invariant

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, operator
- **Problem:** Cluster operations (create, resize, restore, upgrade) are long-running, multi-step, and failure-prone. Without a durable queue providing ordering and mutual exclusion, concurrent or retried operations corrupt tenant clusters.
- **Requirement:** The control plane MUST execute all mutating operations through a durable, metadb-backed task queue supporting delayed execution, task dependencies, timeouts, restarts with history, cancellation, and idempotency. The queue MUST enforce — at the storage layer, not by convention — at most one active mutating task per tenant cluster. Worker crashes MUST leave tasks either resumable or safely retryable without duplicate side effects.
- **Acceptance evidence:** unit and integration tests for dependency resolution, delay, timeout, restart, and cancellation; a storage-level uniqueness test attempting two concurrent mutating tasks on one cluster; a worker-kill chaos test proving safe mid-task resumption.
- **Non-goals:** Not a general-purpose workflow engine. Read-only operations need not enter the queue.
- **Non-claims:** Queue throughput limits at large fleet scale are unmeasured. Poison-task quarantine policy is undefined.
- **Stop conditions:** data — halt and escalate on any bypass of the single-active invariant, including manual operator mutations outside the queue; halt on task restart counts beyond a configured threshold.
- **Traceability:** legacy-platform-b, legacy-platform-a; related CR-DAT-010.

### CR-DAT-030 — Explicit lifecycle state machine with error and offline paths

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, tenant, operator
- **Problem:** Tenants and platform automation need an unambiguous answer to "what is this cluster doing right now." Implicit or UI-inferred status hides failures and corrupts billing, support, and safety decisions.
- **Requirement:** Every managed data-service cluster MUST be governed by an explicit, documented state machine covering at minimum creating, running, modifying, stopping/stopped, starting, deleting, purging, restoring (online and offline), and maintaining, each with paired error states. Derived predicates (visible / active / billable / in-error) MUST be defined once in the control plane and consumed consistently by API, billing, and portal surfaces. State transitions MUST be driven only by queued tasks per CR-DAT-020.
- **Acceptance evidence:** state-machine contract tests including illegal-transition rejection; API conformance checks that all status values come from the defined enum; predicate tests proving billable exactly when active.
- **Non-goals:** Does not mandate identical internal substates for every engine; engine-specific substates may exist behind the shared top-level machine.
- **Non-claims:** No cross-surface audit yet proving portal, billing, and API never diverge on status in practice.
- **Stop conditions:** data / money — halt on any transition into deletion or purge states not issued through the queue; halt on clusters observed in an undefined state; freeze billing changes while the billable predicate is disputed.
- **Traceability:** legacy-platform-b, legacy-platform-a; related CR-DAT-020, CR-DAT-100.

### CR-DAT-040 — Provider-adapter isolation of platform primitives

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team
- **Problem:** Engine automation that calls compute, network, DNS, load-balancer, object-storage, identity, or secrets APIs directly hard-wires data services to one substrate and one vendor, violating the platform's replaceable-behind-contract principle.
- **Requirement:** Engine task logic MUST NOT call substrate APIs directly; it MUST compose versioned provider adapters for compute, network, DNS, load balancing, object storage, identity, secrets, and certificates. Adapters MUST be replaceable per installation profile without changes to engine logic, and each adapter contract MUST ship with a test double usable in CI.
- **Acceptance evidence:** adapter contract test suites plus a reference double implementation; an architecture/static check proving engine task packages import only adapter interfaces, never substrate clients.
- **Non-goals:** Does not require one installation to run multiple substrates simultaneously.
- **Non-claims:** A second substrate adapter family is not yet implemented; portability is an architectural property, not yet a demonstrated one.
- **Stop conditions:** keys / exposure — halt on any engine code path that embeds substrate credentials instead of obtaining them through the secrets adapter; halt on adapter changes that weaken authorization or scoping checks.
- **Traceability:** legacy-platform-b, current-core; related CR-DAT-050.

### CR-DAT-050 — Secrets brokering for engine credentials, TLS, and backup keys

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, auditor
- **Problem:** Managed engines generate and consume high-value secrets — superuser credentials, replication keys, TLS material, backup encryption keys. Storing these in the metadb, task payloads, configuration files, or host disks in plaintext is an existential exposure risk.
- **Requirement:** All data-service secrets MUST be referenced, brokered, and rotated through the platform's approved secrets workflow, with KMS-class envelope encryption at rest where persistence is unavoidable. Metadb rows and task payloads MUST carry secret references or ciphertext, never plaintext. Distribution of secrets to dataplane hosts MUST be mutually authenticated, encrypted in transit, and audited; dataplane components MUST receive the minimum secret scope their function requires.
- **Acceptance evidence:** source-safety scans clean across data-service code and docs; integration test proving no plaintext secrets in metadb or task payloads; a rotation drill for engine superuser and backup encryption keys; audit-log evidence of every secret distribution event.
- **Non-goals:** Does not prescribe the secrets engine. Does not cover tenant-managed in-engine user passwords (engine-native ACL concern).
- **Non-claims:** Zero-downtime rotation of replication and backup keys is not yet drill-proven.
- **Stop conditions:** keys / exposure — halt and escalate immediately on detection of any plaintext secret in any store, log, or task payload; halt secret distribution to any host lacking authenticated, encrypted delivery; block service public launch on a failed rotation drill.
- **Traceability:** legacy-platform-a, legacy-platform-b, current-core; related CR-DAT-080.

### CR-DAT-060 — Staged deletion with purge and metadata barriers

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Deleting a tenant cluster destroys tenant data. Single-step, unauthenticated, or barrier-free deletion paths are unrecoverable business and trust events.
- **Requirement:** Cluster deletion MUST be staged (delete → purge → metadata-delete) with a configurable retention window between stages. The first destructive step MUST require either a fresh verified backup or an explicit recorded tenant acknowledgement. Deletion MUST require confirmed tenant intent through the authenticated API path — never operator-side shortcuts. Purge (irreversible storage destruction) MUST re-verify backup posture before executing, and metadata-delete MUST preserve the billing and audit trail per retention policy.
- **Acceptance evidence:** end-to-end deletion drill proving each stage gate; negative tests (missing confirmation, failed backup barrier) proving fail-closed behavior; audit-trail completeness check across all stages.
- **Non-goals:** Does not define platform-wide retention durations (owned by data-safety policy). Does not cover tenant-initiated drops of in-engine objects (databases, tables, topics).
- **Non-claims:** Retention-window defaults are unvalidated against tenant expectations. Deletion-semantics consistency across all engines is unproven.
- **Stop conditions:** deletion / data / trust — halt and escalate on any purge executed with a failed or missing backup barrier; halt on deletion requests lacking authenticated tenant intent; freeze all bulk-deletion tooling pending human review.
- **Traceability:** legacy-platform-b, current-core; related CR-DAT-030, CR-DAT-080.

### CR-DAT-070 — PostgreSQL reference engine: automated HA and failover

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** The first managed database engine must survive single-host loss without tenant-visible data loss. Manual failover is too slow and too error-prone for a production offering.
- **Requirement:** The PostgreSQL reference engine MUST provide automated high availability: synchronous or semi-synchronous replication options, continuous health detection, automatic failover with fencing and split-brain protection coordinated through a consensus or distributed-locking mechanism, and a bounded failover time. Failover events MUST be recorded as first-class lifecycle and audit events, with the old primary's fate (fenced, rejoined, discarded) explicitly tracked.
- **Acceptance evidence:** HA conformance suite covering kill-primary, network-partition, and split-brain scenarios, run in CI and on a live stand; measured failover time and zero-committed-transaction-loss evidence for the synchronous mode; failover audit-event checks per CR-DAT-180.
- **Non-goals:** Multi-region active-active PostgreSQL. Automatic failback (may remain operator-gated).
- **Non-claims:** Failover timing targets are not yet validated under production-like load. The zero-data-loss property holds only for explicitly configured synchronous topologies.
- **Stop conditions:** data — halt automated failover and escalate whenever fencing is uncertain: a primary that cannot be provably fenced MUST NOT be automatically replaced; halt on failover flapping beyond threshold.
- **Traceability:** legacy-platform-a, legacy-platform-b; related CR-DAT-180.

### CR-DAT-080 — Object-storage-backed backup and restore with continuous archiving

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Tenant data durability depends on backups that provably restore. Engines need scheduled base backups plus continuous write-ahead-log-class archiving for point-in-time recovery, stored off-host in durable object storage.
- **Requirement:** Every stateful data service MUST support scheduled base backups and, where the engine permits, continuous WAL-class archiving (wal-g-class) to platform object storage, with per-cluster schedules, retention policy, envelope or server-side encryption, and backup dependency chains (incremental backups track their ancestors). Restore MUST support both online (new cluster from backup) and offline (in-place) paths, each represented in the lifecycle state machine. Restore drills MUST run on a schedule and their results recorded as readiness evidence.
- **Acceptance evidence:** per-engine backup/restore e2e suites; fresh non-synthetic restore-drill reports as a readiness gate; point-in-time recovery test to a specified timestamp; encryption check proving backups are unreadable without the keys.
- **Non-goals:** Cross-provider or cross-installation backup replication (federation concern). Tenant self-service backup export formats.
- **Non-claims:** Restore-time objectives at large database sizes are unmeasured. Continuous-archiving lag under write-heavy load is uncharacterized.
- **Stop conditions:** data / keys / deletion — halt schedule changes that would open an unprotected window; escalate on two consecutive failed backups for any cluster; halt in-place restores lacking a pre-restore snapshot; block purge of any cluster whose final backup failed (CR-DAT-060).
- **Traceability:** legacy-platform-a, legacy-platform-b; related CR-DAT-050, CR-DAT-100.

### CR-DAT-090 — In-place engine version upgrades gated by backup evidence

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Engine versions age out. A managed service without an in-place upgrade path strands tenants on unsupported versions or forces dump-and-reload migrations with long downtime.
- **Requirement:** The reference engine MUST support in-place upgrades between supported versions, executed as queued tasks with pre-upgrade backup verification, upgrade progress visible through the API, and a documented failure/rollback path. Upgrade tasks MUST refuse to start when no fresh restorable backup exists. Minor-version upgrades SHOULD be automatable inside maintenance windows; major-version upgrades MUST be explicitly tenant- or operator-initiated.
- **Acceptance evidence:** upgrade e2e suite covering each supported version pair (provision on N, upgrade to N+1, verify data integrity and replication); negative test proving refusal without a fresh backup; rollback-path drill evidence.
- **Non-goals:** Unconsented automatic major-version upgrades. Cross-engine migration tooling.
- **Non-claims:** Major-version in-place upgrade reliability is not yet proven at production scale; the rollback path may be restore-from-backup rather than binary downgrade.
- **Stop conditions:** data / migration — halt the upgrade task and escalate on backup-verification failure, replication-lag anomaly, or post-upgrade health-check failure; never chain a second mutating task onto a cluster mid-upgrade (CR-DAT-020 invariant).
- **Traceability:** legacy-platform-a, legacy-platform-b; related CR-DAT-020, CR-DAT-080, CR-DAT-170.

### CR-DAT-100 — Usage metering and billable backup intervals via durable outbox

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, auditor
- **Problem:** Managed data services are commercial products. Usage records (instance time, storage, backup storage) lost or duplicated between the service and the billing pipeline are direct revenue leakage or tenant overcharge.
- **Requirement:** Data services MUST track billable state as intervals per cluster and bill type (at minimum instance usage and backup storage), MUST emit usage through the platform metering contract before the service may be called generally available, and MUST stage emissions in a durable, sequenced, restart-counted outbox so worker restarts never drop or duplicate usage. Backup usage MUST be metered as intervals distinct from instance usage. Metering gaps MUST be detectable and reconcilable from the interval tables.
- **Acceptance evidence:** metering-contract conformance tests against the billing domain's usage schema; outbox restart and duplicate-suppression tests; a reconciliation report matching interval tables to emitted usage; a launch-gate check proving metering is live before public availability.
- **Non-goals:** Rating, pricing, invoicing, and budgets (billing domain). Metering of in-engine tenant activity (queries, rows, throughput) beyond the defined bill types.
- **Non-claims:** End-to-end reconciliation against the real billing pipeline is not yet demonstrated. Clock-skew impact on interval boundaries is unquantified.
- **Stop conditions:** money — halt and escalate on any detected metering gap or duplicate-emission anomaly; freeze service launch while metering conformance is unverified; stop automatic charging paths when interval records and emitted usage disagree beyond tolerance.
- **Traceability:** legacy-platform-b, legacy-platform-a; related CR-DAT-030, CR-DAT-080; metering pipeline contract owned by the BIL domain.

### CR-DAT-110 — Data services onboard through OCS connector packages (dogfooding)

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, vendor
- **Problem:** If first-party data services wire directly into platform internals, the Open Cloud Standard remains unproven and third-party service teams inherit a second-class path. The platform must dogfood its own extensibility standard.
- **Requirement:** Every managed data service MUST be packaged, validated, published, and operated as an OCS connector package implementing the mandatory lifecycle APIs and the portal, billing, tenant-access, readiness, and durability surfaces. Platform internals MUST depend only on connector metadata and APIs, never on data-service implementation details. Capabilities offered to first-party data services and to third-party vendors MUST be identical.
- **Acceptance evidence:** connector package validation passing for each data service; an end-to-end onboarding drill registering a data service using only public documentation and the SDK; architecture tests proving platform internals contain no service-specific imports.
- **Non-goals:** Does not require data services to be multi-provider on day one. Does not cover marketplace commercial mechanics (MKT/BIL domains).
- **Non-claims:** The full OCS surface set has not yet been exercised by an external team. First-party/third-party capability parity is asserted by design, not yet audited in practice.
- **Stop conditions:** trust / exposure — block publication of any data-service connector that fails validation; halt and escalate on discovery of privileged first-party-only platform hooks; never expose a data service in the portal before its readiness and durability surfaces report green.
- **Traceability:** current-core, vision-deck; related CR-DAT-100; connector contract owned by the OCS domain.

### CR-DAT-120 — Uniform dataplane engine agent contract

- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, provider, operator
- **Problem:** Control-plane tasks ultimately act on tenant hosts. Without one uniform on-host agent contract, every engine invents its own host automation, multiplying the security surface and operational variance.
- **Requirement:** Dataplane hosts SHOULD run a uniform engine agent exposing a versioned contract: in-engine user and database CRUD, engine configuration, backup configuration and execution hooks, upgrade execution with status reporting, health reporting, and self-update via authenticated binary delivery. Agent APIs MUST be authenticated, least-privilege, and fully audited. Agent self-update MUST be staged and rollback-capable.
- **Acceptance evidence:** per-engine agent contract tests; an e2e test of a control-plane task executed through the agent; a self-update drill with forced rollback; audit-log coverage of agent actions.
- **Non-goals:** The agent is not a general configuration-management replacement; host baseline hardening is owned by the deployment/IaC domain.
- **Non-claims:** Agent contract stability across engine generations is unproven. Fleet-wide agent-update blast-radius controls are untested.
- **Stop conditions:** keys / exposure / data — quarantine any agent accepting unauthenticated or downgraded control channels; pause fleet-wide self-update on the first failure cohort until human review.
- **Traceability:** legacy-platform-a, legacy-platform-b; related CR-DAT-040, CR-DAT-050.

### CR-DAT-130 — Kafka-class managed streaming service lifecycle

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, service-team
- **Problem:** Streaming platforms have more moving parts than databases — brokers, a coordination ensemble, client access control, optional public endpoints — and need the same managed-lifecycle rigor as the reference database engine.
- **Requirement:** A Kafka-class managed streaming service SHOULD be provided on the shared control-plane pattern, with an explicit multi-state lifecycle FSM (including provisioning, building, checking, running, resizing, config-updating, suspending/suspended, and error states), per-topic access-control users integrated with platform IAM, optional public access via managed ingress/NAT rules, and broker-plus-coordination topology managed as one unit. Suspension on insufficient funds MUST be a first-class state with a safe resume path.
- **Acceptance evidence:** FSM conformance tests; e2e lifecycle drill (create → produce/consume → resize → suspend → resume → delete); IAM-integration tests for topic-level ACLs; exposure tests proving no unintended public listeners.
- **Non-goals:** Exactly-once tenant application semantics. Managed schema evolution policy (see CR-DAT-200). Cross-cluster replication.
- **Non-claims:** The streaming engine is not part of the first ship baseline; its durability and failover evidence does not yet exist. Multi-zone broker placement policy is unproven.
- **Stop conditions:** data / exposure — halt on any configuration change that would open unauthenticated client access; halt resize and config tasks while partitions are under-replicated; escalate repeated task restarts per CR-DAT-020.
- **Traceability:** legacy-platform-a; related CR-DAT-030, CR-DAT-140, CR-DAT-150.

### CR-DAT-140 — Continuous rebalancing for streaming clusters (cruise-control-class)

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Broker load in streaming clusters drifts with partition skew, broker replacement, and growth. Manual partition reassignment is toil and a common cause of outages.
- **Requirement:** Streaming services SHOULD integrate cruise-control-class rebalancing automation that continuously models cluster load, proposes partition and leadership movements, executes them throttled, and supports dry-run approval for large moves. Rebalancing MUST respect replication-factor invariants, MUST pause automatically on cluster instability, and MUST NOT run concurrently with lifecycle mutations (CR-DAT-020 invariant).
- **Acceptance evidence:** a rebalancing drill on a deliberately skewed test cluster showing measured improvement with zero under-replicated-partition windows; dry-run/approval flow test; mutual-exclusion test with lifecycle tasks.
- **Non-goals:** Automatic broker scaling decisions (capacity management is separate). Tenant-visible rebalancing APIs.
- **Non-claims:** Optimization-goal tuning for production workloads is unproven. Rebalancing traffic impact on client latency is uncharacterized.
- **Stop conditions:** data — halt rebalancing immediately on under-replicated partitions, controller instability, or degraded client latency; require operator approval for proposals above a movement-volume threshold.
- **Traceability:** legacy-platform-a; related CR-DAT-130.

### CR-DAT-150 — Emergency produce-disable breaker for resource exhaustion

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** A streaming cluster whose disks fill fails ungracefully and can lose data. The platform needs a deliberate, reversible emergency lever that stops client writes to save the cluster.
- **Requirement:** Streaming services MUST provide an emergency produce-disable breaker: on configurable disk-usage thresholds the platform SHOULD be able to disable produce paths while keeping consume paths alive, with engagement and release represented as explicit lifecycle states, automatic operator alerting, tenant notification, and an audited, authorization-gated manual trigger. The breaker MUST be reversible without data loss.
- **Acceptance evidence:** a drill filling a test cluster to threshold proving automatic engagement, consume continuity, and clean re-enable; authorization tests on the manual trigger; audit and notification evidence for both transitions.
- **Non-goals:** Tenant-level quotas and rate limiting (a separate capability). Automatic message deletion to free space.
- **Non-claims:** Threshold defaults are unvalidated. Client behavior under prolonged engagement (retry storms) is uncharacterized.
- **Stop conditions:** data / trust / money — the breaker deliberately degrades tenant writes: any engagement without threshold evidence or without operator alert is an incident; repeated engagements for one cluster MUST trigger a capacity review before re-enable.
- **Traceability:** legacy-platform-a; related CR-DAT-130, CR-DAT-180.

### CR-DAT-160 — Redis-class managed cache engine

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, service-team
- **Problem:** Tenants expect a managed in-memory cache with HA in a credible data-services portfolio. It is also the second engine proving that the control-plane pattern generalizes beyond PostgreSQL.
- **Requirement:** A Redis-class managed cache SHOULD be provided through the same metadb, queue, and FSM framework, with replica-based HA driven by a sentinel-class control loop, optional persistence under the same backup contract as other engines when enabled, version upgrades, and eviction-policy/configuration surfaces exposed through the versioned engine config API.
- **Acceptance evidence:** an engine conformance suite reusing the shared framework tests (evidence of pattern reuse); HA failover drill; backup/restore e2e for persistence-enabled configurations; config-API contract tests.
- **Non-goals:** Cache-module ecosystems beyond a defined baseline. Durability guarantees for non-persistent configurations, which are explicitly volatile.
- **Non-claims:** Not part of the first ship baseline. Sentinel-class failover correctness under network partition is not yet drill-proven for this engine.
- **Stop conditions:** data — persistence-enabled clusters inherit the backup and deletion barriers of CR-DAT-080 and CR-DAT-060; any durability claim for a non-persistent configuration MUST be blocked at the API and documentation level.
- **Traceability:** legacy-platform-a, legacy-platform-b; related CR-DAT-070, CR-DAT-080.

### CR-DAT-170 — Maintenance windows and config-driven maintenance waves

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Engine fleets need rolling maintenance — patches, configuration changes, minor upgrades. Per-cluster manual execution does not scale, and execution outside tenant windows violates service expectations.
- **Requirement:** The platform SHOULD support config-driven maintenance waves: a declarative maintenance definition (cluster selection, desired-state mutation, task type and arguments) executed through the standard queue, honoring per-cluster tenant-configurable maintenance windows, with progress tracking and the ability to pause a wave. Emergency security maintenance MAY bypass windows but MUST be flagged, alerted, and audited as such.
- **Acceptance evidence:** a maintenance-wave e2e drill on a test fleet (selection → windowed execution → pause/resume); window-compliance tests proving no task starts outside a window unless flagged emergency; per-wave audit trail.
- **Non-goals:** Negotiated per-tenant scheduling beyond window preferences. Fully unsupervised execution of major changes.
- **Non-claims:** Fleet-scale wave throughput and blast-radius controls are unvalidated.
- **Stop conditions:** data / exposure — pause the wave and escalate when the in-wave failure rate exceeds threshold; any out-of-window execution without the emergency flag halts the wave immediately.
- **Traceability:** legacy-platform-b, legacy-platform-a; related CR-DAT-020, CR-DAT-090.

### CR-DAT-180 — Lifecycle event emission to the platform audit/event bus

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, auditor, tenant
- **Problem:** Tenants, auditors, and support need a trustworthy activity trail of what happened to each cluster and when. Internal task tables are not a tenant-facing or compliance-facing event stream.
- **Requirement:** Data services SHOULD emit lifecycle events — operation started/done wrapped with cluster, operation type, actor class, and outcome metadata — to the platform event/audit bus in a CloudEvents-class envelope, covering at minimum create, modify, start, stop, restore, upgrade, failover, deletion stages, and breaker transitions. Emission MUST be durable across emitter restarts and ordered per cluster.
- **Acceptance evidence:** event-schema contract tests; e2e proof that each queued mutation produces a matching started/done pair; emitter-restart durability test; per-cluster ordering checks.
- **Non-goals:** Tenant-defined subscriptions and webhooks (portal/integration domains). Long-term audit retention policy (audit domain).
- **Non-claims:** Exactly-once delivery is not claimed; consumers must tolerate duplicates. Cross-service envelope consistency is not yet audited.
- **Stop conditions:** trust — halt promotion of any service whose lifecycle mutations can complete without corresponding events; alert on event-pipeline lag beyond threshold.
- **Traceability:** legacy-platform-b, legacy-platform-a; related CR-DAT-020, CR-DAT-150; event bus owned by the observability/audit domains.

### CR-DAT-190 — Connection registry with envelope-encrypted credentials

- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, tenant, vendor
- **Problem:** Data tooling and integration services need a central, typed registry of database connection configurations. Connection strings with embedded passwords scattered across configs are unmanageable and unsafe.
- **Requirement:** The platform MAY provide a connection-registry service storing per-engine typed connection configurations (cluster references, network parameters, TLS settings) with credentials persisted only as KMS-class envelope-encrypted ciphertext, asynchronous mutation operations, and IAM-authorized read paths that SHOULD broker short-lived access rather than hand out static passwords where engine capabilities allow.
- **Acceptance evidence:** per-engine connection schema contract tests; proof that at-rest records contain only ciphertext and that decryption paths are IAM-gated and audited; async-operation conformance tests.
- **Non-goals:** Connection pooling or proxying (a separate product). Automatic rotation of engine-native user credentials.
- **Non-claims:** Forward-looking service: no production deployment or consumer integration exists yet. Short-lived-credential brokering depends on engine capabilities not yet verified.
- **Stop conditions:** keys / exposure — any registry response returning plaintext credentials to an unauthorized caller is an immediate halt-and-escalate security incident; the registry MUST fail closed when the key-management dependency is unavailable — no decryption and no degraded plaintext fallback.
- **Traceability:** legacy-platform-b; related CR-DAT-050.

### CR-DAT-200 — Schema registry for streaming ecosystems

- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, tenant, vendor
- **Problem:** Streaming producers and consumers break silently when message schemas change without control. Tenants need namespaces, versioned schemas, and compatibility enforcement.
- **Requirement:** The platform MAY provide a schema-registry service with namespaces, versioned schemas, search, and configurable compatibility modes (backward, forward, full), exposed through the same IAM and audit surfaces as other data-platform services, and integrated with the Kafka-class streaming service so producers and consumers can reference schema versions.
- **Acceptance evidence:** compatibility-check unit tests across all modes; API contract tests for namespace, schema, version, and search surfaces; an integration drill with the streaming service showing producer rejection on an incompatible change.
- **Non-goals:** Schema inference from traffic. Governance approval workflows beyond compatibility modes.
- **Non-claims:** Forward-looking: no production consumers exist yet. Interoperability depth with third-party schema tooling is unverified.
- **Stop conditions:** data / trust — registry unavailability MUST NOT silently disable compatibility enforcement on produce paths that declare schema enforcement; fail closed or reject with an explicit error, never pass through unvalidated.
- **Traceability:** legacy-platform-b; related CR-DAT-130.

## Coverage notes

This domain deliberately defers:

- **Metering pipeline, rating, pricing, invoicing, budgets** — BIL domain
  (`domains/16-billing-finops.md`). DAT owns only the emitter side
  (CR-DAT-100).
- **OCS connector package format, validation, lifecycle APIs** — OCS domain
  (`domains/17-ocs-service-connectors.md`). DAT dogfoods the contract
  (CR-DAT-110) but does not define it.
- **Marketplace catalog, publisher economics, revenue share** — MKT domain.
- **Portal surfaces and tenant self-service UX** — CUX domain. DAT supplies
  readiness and durability surfaces, not screens.
- **Object storage product itself** (bucket APIs, durability of the backup
  target) — STO domain. DAT consumes it as a backup backend.
- **IAM/KMS substrate** — IAM domain. DAT brokers through it
  (CR-DAT-050, CR-DAT-190) but does not implement it.
- **Event/audit bus, monitoring, and alerting infrastructure** — OBS domain.
  DAT emits events and standard service observability only.
- **Backup/DR of the platform control plane itself** — STO/OPS domains. DAT
  covers tenant-cluster backups only.
- **Compute, network, DNS, and load-balancer provisioning primitives** —
  CMP/NET domains, consumed exclusively through provider adapters
  (CR-DAT-040).
- **Deployment of the data-service control plane** (GitOps, images,
  environments) — DPL domain.
- **Cross-installation federation of data services** (backup shipping,
  global catalog) — FED domain; future work.
- **Consumption analytics, DWH, and ML portfolios** — out of scope here; a
  candidate for a separate analytics domain in a later wave.
