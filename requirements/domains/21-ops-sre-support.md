# 21 — Operations, SRE and Support

Day-2 operation of a CloudRING installation and of the provider's support
business: operations-automation safety (idempotency, plan/preview, action
risk classes), flow control and event-delivery honesty, incident management
(severity scale, on-call, postmortems, decision memory), change management
(dated evidence, rollback plans, maintenance windows), quota-versus-limit
administration, support tooling (tickets, SLA metrics, diagnostics,
impersonation), version-skew policy, capacity management, runbooks as code,
and disaster-recovery drill scheduling. Applies to platform internals and to
OCS services participating in platform operations.

**Domain contract.** No silent mutation and no silent loss. Every mutating
operation is idempotent, previewable, risk-classed, evidenced, and reversible
—or explicitly declared irreversible with proof of backup/export. Every event
path that carries lifecycle, usage, or audit truth is bounded,
pressure-explicit, and never drops silently. Every privileged support action
is ticket-bound, impersonated, time-boxed, and fully audited. Readiness
claims are earned: change records, drills, and postmortem follow-through gate
what may be called operable, and `blocked` is an honest, non-promotable
state. Where money, data, keys, trust, exposure, deletion, migration, or
settlement are at stake, the platform fails closed and escalates.

### CR-OPS-010 — Idempotent operations with plan/preview
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, service-team
- **Problem:** Repeated or retried administrative actions against
  infrastructure and services produce divergent, partial, or duplicated state
  when idempotency is left to convention. Operators and agents must see what
  an action will do before it does it.
- **Requirement:** All platform operations automation that mutates state MUST
  be idempotent, or MUST declare its repeated-execution consequence
  explicitly in its contract. Mutating actions MUST provide a plan/preview
  mode that shows intended changes without applying them. Plans, results, and
  audit artifacts MUST be stored as structured, dated records. Execution MUST
  refuse to proceed when live state has drifted from the state the plan was
  computed against.
- **Acceptance evidence:** Contract-test suite proving convergence under
  repeated execution for the reference operation set; recorded plan/apply
  artifacts from a test-stand execution; drift-detection negative tests in
  which execution refuses a stale plan.
- **Non-goals:** Does not prescribe a specific IaC engine or plan format;
  does not cover read-only diagnostics (see CR-OPS-130).
- **Non-claims:** Only a subset of operations (installation preflight and
  deploy planning in current-core) demonstrates this discipline today;
  fleet-wide coverage is not yet evidenced.
- **Stop conditions:** Halt and escalate on any validation contradiction
  between plan and live state; on a destructive step without a
  rollback/compensation note; on missing or stale plan artifacts for an
  action about to mutate production (data/deletion/migration risk).
- **Traceability:** req-history; legacy-platform-b; current-core (operator
  CLI preflight and deploy-plan flows). Related: CR-OPS-020, CR-OPS-090.

### CR-OPS-020 — Action risk classes visible before execution
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Without an explicit effect-based classification, a log read
  and a disk wipe travel the same execution path, and destructive power
  becomes visible only after the mistake.
- **Requirement:** Every operational action MUST be classified before
  execution by effect into at least read-only, planned-change, and
  destructive classes. The class MUST be computed from the action's effect,
  environment, and blast radius — never from the identity requesting it — and
  MUST be displayed identically across CLI, API, portal, and agent plan
  surfaces. Destructive actions MUST require explicit confirmation plus
  backup/export proof and MUST record a closure audit. Classification rules
  MUST be versioned as code.
- **Acceptance evidence:** Classification registry as code with unit tests;
  end-to-end evidence that a destructive action lacking confirmation and
  backup/export proof is refused; audit-record schema checks.
- **Non-goals:** The full agent approval matrix (agent-governance domain);
  per-tenant RBAC (IAM domain).
- **Non-claims:** Approval workflows beyond CLI confirmation are not yet
  implemented or evidenced; portal surfacing of classes is undesigned.
- **Stop conditions:** Refuse and escalate any destructive action lacking
  confirmation, backup/export proof, or an approval within its validity
  window; stale approval fails closed (deletion/trust risk).
- **Traceability:** req-history; legacy-platform-b. Related: CR-OPS-010,
  CR-OPS-090.

### CR-OPS-030 — Backpressure with explicit pressure states
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, vendor
- **Problem:** Ingestion gateways and operations pipelines that accept work
  without bounds silently drop, delay, or reorder billable, lifecycle, and
  audit events under load — the least honest failure mode a commercial
  platform has.
- **Requirement:** Every ingestion point and gateway MUST enforce bounded
  queues and apply explicit backpressure. Every unit of work MUST land in
  exactly one declared pressure state: accepted, delayed, rejected,
  retryable, quarantined, or dropped-with-proof. Silent accept and silent
  drop MUST be treated as defects; for billable, lifecycle, audit, or
  federation events a drop MUST be impossible without a durable proof record.
  Pressure states MUST be observable as metrics and surfaced to operators.
- **Acceptance evidence:** Fault-injection test suite exercising each
  pressure state at a reference gateway; overload tests showing bounded
  behavior and operator-visible state transitions; audit tests proving no
  silent-drop code path exists for billable/audit event classes.
- **Non-goals:** Message-bus technology selection; per-service SLO
  definitions (observability domain).
- **Non-claims:** Pressure behavior under real multi-tenant production load
  is unproven; existing evidence is synthetic and test-stand only.
- **Stop conditions:** Halt and escalate on detection of any silently dropped
  or silently accepted event in billable, audit, lifecycle, or federation
  flows; on queue growth beyond declared bounds without a state transition
  (money/data/settlement risk).
- **Traceability:** req-history; legacy-platform-b (flow-controlled stream
  consumers); legacy-platform-a. Related: CR-OPS-040.

### CR-OPS-040 — Drain, recovery, and replay safety
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team
- **Problem:** Services that buffer work lose or duplicate events on
  shutdown, deploy, or crash unless flush and recovery semantics are designed
  in, and a single poison message can stall an entire pipeline.
- **Requirement:** Buffered services MUST implement graceful drain on
  shutdown with completeness accounting: what was flushed, what remains, and
  where consumption resumes. Event identity MUST be deterministic or
  idempotency-key based, so retries produce explicit already-processed,
  rejected, or conflict results. Recovery MUST resume without loss or
  duplication of lifecycle, usage, or audit events. Quarantine paths MUST
  exist for poison messages so one malformed item cannot block a stream.
  These behaviors MUST be covered by chaos/fault-injection tests.
- **Acceptance evidence:** Chaos drill evidence (kill mid-flush, crash before
  commit, duplicate delivery, replay after recovery) with post-drill
  completeness audits; quarantine isolation tests; replay-outcome
  classification tests.
- **Non-goals:** Mandating a single queue or streaming technology.
- **Non-claims:** Recovery completeness has not been demonstrated on a
  production-class stand; drill cadence is not yet established (see
  CR-OPS-180).
- **Stop conditions:** Halt a rollout or recovery and escalate when replay
  gaps, unaccounted sequences, or duplicate side effects on billable/audit
  events are detected (data/money risk).
- **Traceability:** req-history; legacy-platform-b (head-of-line blocking
  lesson; cursor/cookie commit pattern). Related: CR-OPS-030, CR-OPS-180.

### CR-OPS-050 — Incident severity classification
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Without a shared severity scale, incidents are under- or
  over-escalated, and postmortem, notification, and SLA obligations become
  inconsistent across teams.
- **Requirement:** The platform MUST define a single incident severity scale
  as versioned documentation: a bounded numeric scale scored on independent
  axes — at minimum parameter deviation, automation-versus-human
  intervention, process violation, user/neighbor impact, data compromise, and
  data loss — with the final level set by the highest axis score. Incidents
  at or above a defined threshold MUST automatically become formal incidents
  with mandatory postmortem (CR-OPS-070) and notification duties. The scale
  MUST include worked examples and MUST be used consistently by alerting,
  on-call, and support tooling.
- **Acceptance evidence:** Severity-scale document in the repository with
  classifier unit tests; escalation-routing configuration tests; tabletop
  exercise records demonstrating consistent scoring across teams.
- **Non-goals:** Prescribing organizational staffing; regulatory reporting
  formats beyond breach-notification timing hooks.
- **Non-claims:** The scale is unexercised by real incident history;
  threshold calibration is an initial policy choice, not a measured one.
- **Stop conditions:** Any confirmed data compromise, data loss, or
  authentication-material exposure halts non-emergency change activity,
  escalates to security duty immediately, and starts breach-notification
  clocks (trust/data risk).
- **Traceability:** legacy-platform-b (multi-axis emergency scale and duty
  handbook); req-history. Related: CR-OPS-060, CR-OPS-070.

### CR-OPS-060 — On-call model with response SLO
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Alert delivery without an explicit duty model produces missed
  pages, unclear ownership, and slow engagement during incidents.
- **Requirement:** Every production service MUST have a declared on-call
  rotation with a primary and a backup at all times, defined as code and
  pushed to the alerting/duty system. Critical alerts MUST have a declared
  engagement SLO (default target: acknowledgment within 10 minutes,
  configurable per installation) with automatic escalation to the backup and
  then to leadership on breach. Duty priority ordering (availability above
  incidents above support above releases and ad-hoc work) MUST be documented,
  and formal handover between rotations MUST be recorded.
- **Acceptance evidence:** Rotation definitions as code with CI validation
  (no gaps, backup always set); paging integration tests against test
  channels; escalation drill evidence showing backup engagement on primary
  timeout; retained handover records.
- **Non-goals:** Chat-bot user experience (CR-OPS-190); human-resources
  scheduling.
- **Non-claims:** SLO attainment is unmeasured; no live paging evidence
  exists yet.
- **Stop conditions:** An unacknowledged critical page beyond SLO MUST
  auto-escalate; repeated escalation failure halts release activity for the
  affected service until staffing is corrected (trust risk).
- **Traceability:** legacy-platform-b (primary/backup duty, 10-minute
  engagement, duty-as-code rosters); legacy-platform-a; req-history.
  Related: CR-OPS-050.

### CR-OPS-070 — Structured postmortems
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, auditor, provider
- **Problem:** Incidents that end without structured analysis recur;
  link-based evidence rots; recommendations without tracking die.
- **Requirement:** Formal incidents MUST produce a postmortem following the
  platform template: chronological investigation with captured evidence
  (snapshots and exports, not live links); analysis covering all hypotheses
  including disproven ones with their disproof facts; conclusions with direct
  and systemic root causes; recommendations as concrete, ticket-linked,
  problem-focused follow-ups. Postmortems MUST pass a review cadence (default
  twice-weekly until released), and recommendation completion MUST be tracked
  (default twice-monthly follow-up review). Postmortem artifacts MUST be
  redacted before broad circulation: no secrets, credentials, or tenant
  payloads.
- **Acceptance evidence:** Postmortem template with lint checks in CI
  (required sections, ticket links); redaction scan wired into the
  source-safety gate; cadence and recommendation tracking reports; at least
  one completed exercise (game-day or real) using the full template.
- **Non-goals:** Blame assignment; legal discovery workflows.
- **Non-claims:** No real postmortem corpus exists; the template is inherited
  discipline, not battle-tested on this platform.
- **Stop conditions:** Circulation of a postmortem containing unredacted
  tenant data, secrets, or credentials halts distribution and triggers the
  security incident path (exposure risk).
- **Traceability:** legacy-platform-b (postmortem structure and review
  cadence); legacy-platform-a (postmortem artifact directories);
  req-history. Related: CR-OPS-050, CR-OPS-080.

### CR-OPS-080 — Operational decision memory (ADRs)
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, agent
- **Problem:** Operational choices made in chats and tickets are lost;
  superseded decisions get re-litigated or silently violated; readiness
  claims quietly depend on unaccepted decisions.
- **Requirement:** Significant operational decisions MUST be recorded as
  architecture decision records with an explicit status lifecycle (proposed,
  accepted, superseded, deprecated, rejected) and supersession chains.
  Operational churn that invalidates a document, runbook, or ADR MUST either
  update the artifact or record a no-change rationale. Readiness reports MUST
  mark dependencies on proposed (unaccepted) ADRs, and high-risk proposed
  dependencies (data, money, trust, destructive, cross-provider) MUST block
  readiness claims absent an accepted decision or a scoped owner waiver.
  Records MUST be source-safe: no copied vendor text, endpoints, or tenant
  data.
- **Acceptance evidence:** ADR directory with schema and status validation in
  CI; supersession-chain validator; readiness-gate tests demonstrating
  proposed-dependency marking and blocking; linkage checks between change
  records and updated artifacts.
- **Non-goals:** Product-requirement ADR backlog management (governance
  artifacts outside this domain).
- **Non-claims:** ADR coverage of existing operational choices is partial;
  blocking behavior is implemented only for current-core installation flows.
- **Stop conditions:** Halt a change that contradicts an accepted ADR until
  the ADR is explicitly superseded; a high-risk change depending on a
  proposed ADR blocks until acceptance or scoped waiver (trust risk).
- **Traceability:** req-history (ADR lifecycle, proposed-dependency marking);
  current-core (decision records). Related: CR-OPS-090, CR-OPS-180.

### CR-OPS-090 — Change management with dated evidence and rollback plans
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, auditor, agent
- **Problem:** Changes executed without recorded pre/post state and a
  rollback plan cannot be audited, correlated with incidents, or safely
  undone.
- **Requirement:** Every production change MUST produce a dated change record
  containing the plan (CR-OPS-010), pre-change state capture
  (resource/configuration snapshots), execution log, post-change verification
  evidence, and a rollback or compensation plan defined before execution.
  Change records MUST follow a standard, dated, append-only convention and
  MUST be linkable from incident and postmortem workflows. Rollback plans
  MUST be executable, not aspirational — automated where the change is
  automated. Changes to stateful systems MUST include data-durability
  evidence (current backup/restore status) before mutation.
- **Acceptance evidence:** Change-record convention validator in CI
  (structure and required artifacts); end-to-end evidence from a reference
  change on a test stand including an exercised rollback; incident-correlation
  queries demonstrating change-to-incident linkage; restore-status gate
  checks for stateful changes.
- **Non-goals:** Heavyweight approval bureaucracy; deployment pipeline
  mechanics (deployment domain).
- **Non-claims:** Rollback has been exercised only for installation-wave
  flows in current-core; broad automated rollback coverage is unproven.
- **Stop conditions:** Halt the change and invoke rollback/compensation when
  post-change verification fails, when evidence capture fails mid-change, or
  when a stateful change lacks current backup/restore evidence; never
  continue a change whose rollback plan is missing or unexecutable
  (data/deletion/migration risk).
- **Traceability:** legacy-platform-a (dated change-evidence and audit-output
  convention); legacy-platform-b (change regulation); req-history;
  current-core (upgrade evidence conventions). Related: CR-OPS-010,
  CR-OPS-100.

### CR-OPS-100 — Maintenance windows with change-register interlock
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, tenant
- **Problem:** Disruptive maintenance executed ad hoc collides with tenant
  workloads, other teams' changes, and support commitments.
- **Requirement:** The platform MUST provide a central change-management
  register (CMS-class): a system of record for planned maintenance windows,
  approved changes, and affected scopes. Disruptive automation (host drains,
  upgrades, network changes, service restarts) MUST interlock with the
  register before acting: execute only inside an approved window for the
  affected scope, and fail closed when the register is unreachable. Window
  schedules and tenant-visible notices MUST be published in advance (default
  minimum notice configurable). Mid-maintenance expiry or scope overrun MUST
  pause the action pending re-approval. Emergency paths MUST record
  retrospective registration.
- **Acceptance evidence:** Register API contract tests; interlock end-to-end
  evidence (action refused outside a window; action refused when the register
  is down); tenant-notice generation tests; emergency
  retrospective-registration drill.
- **Non-goals:** Coordinating tenants' own application maintenance;
  tenant-facing notice presentation (portal domain).
- **Non-claims:** No production maintenance history exists; the interlock is
  designed but not yet wired across all disruptive controllers.
- **Stop conditions:** Halt and reschedule on interlock check failure, window
  expiry, or scope overrun; a disruptive action outside any approved window
  is refused by construction; repeated violations suspend the offending
  automation pending review (trust/exposure/data risk).
- **Traceability:** legacy-platform-b (managed-service change-management API
  client; downtime scripting); legacy-platform-a; req-history. Related:
  CR-OPS-090, CR-OPS-150.

### CR-OPS-110 — Quota-versus-limit model with reserve/commit/release
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, tenant, service-team
- **Problem:** Unbounded self-service provisioning can oversubscribe shared
  capacity and create unbillable or undeliverable commitments; ad hoc quota
  edits are unaudited and race-prone.
- **Requirement:** The platform MUST distinguish quotas (soft,
  organization/account-scoped, adjustable through a governed flow) from
  limits (hard architectural ceilings documented per service and not silently
  changeable). Default quotas SHOULD align with the trial/entry tier.
  Admission control for capacity-backed resources MUST use a
  reserve → commit/release protocol: reserve quota before provisioning,
  commit on success, release on failure or cancellation; reservations MUST
  carry expiry and be reclaimable, and reservation leaks MUST be detected and
  alarmed. Quota changes MUST be audited actions bound to an actor and, for
  support-originated changes, to a ticket; the request flow MUST be available
  through console and API. Oversubscription ratios, where a product uses
  them, MUST be declared policy, enforced, and visible — never implicit.
- **Acceptance evidence:** Quota-service contract tests including concurrent
  reserve/commit/release races, expiry reclaim, and leak detection;
  quota-change audit-trail checks; end-to-end evidence of the quota-increase
  request flow on a test stand; oversubscription policy enforcement tests.
- **Non-goals:** Billing rating and charging (billing domain); per-service
  quota catalogs (each service declares its own via OCS/billing surfaces).
- **Non-claims:** The quota subsystem is specified but not implemented
  end-to-end; no production contention evidence exists; oversubscription
  policies are undeclared for current services.
- **Stop conditions:** Refuse provisioning when reservation fails or quota is
  exhausted — never provision-then-bill; halt and escalate on detected
  reservation leaks, quota drift between enforcement points, or any quota
  mutation without an audit binding (money/settlement risk).
- **Traceability:** legacy-platform-b (soft-quota/hard-limit model, quota
  tooling, request flow); legacy-platform-a (declared oversubscription
  enforcement); req-history. Related: CR-OPS-130, CR-OPS-160.

### CR-OPS-120 — Support ticket system of record and SLA metrics
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Support handled in chats and mailboxes loses history, cannot
  be measured, and breaks handoff between provider teams.
- **Requirement:** Every provider installation MUST designate a ticket system
  of record for customer support, and every tenant-affecting request and
  support-originated action MUST be bound to a ticket. The platform MUST
  compute support SLA metrics (first-response and resolution against the
  provider's published support-plan matrix) from ticket-system data, with
  scheduled collection, retained history, and exportable reports. Gaps in the
  source data MUST be flagged, not interpolated. Support plans (channels and
  response times) SHOULD be published as part of the provider's product
  documentation.
- **Acceptance evidence:** Ticket-binding enforcement tests for
  support-originated actions (unticketed privileged action rejected); SLA
  pipeline runs over fixture queues with snapshot-verified reports;
  backfill/recompute evidence; report export checks.
- **Non-goals:** Building a full ticketing product — integration with an
  external system of record is acceptable; marketplace review systems.
- **Non-claims:** SLA computation is demonstrated only on fixtures; no real
  support history exists; the plan matrix is a template until a provider
  publishes one.
- **Stop conditions:** A support-originated action without ticket binding is
  refused; SLA-relevant records MUST NOT be deleted or rewritten
  (append-only); halt and escalate on detected tampering with SLA source
  data (money/trust risk).
- **Traceability:** legacy-platform-b (ticket queue as system of record, SLA
  metrics collector, plan matrix); req-history (support-safe receipts).
  Related: CR-OPS-130, CR-OPS-140.

### CR-OPS-130 — Diagnostic CLI and aggregated tenant view
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, service-team
- **Problem:** First-line support burns incident time hopping between
  consoles and APIs to resolve a tenant-to-account-to-resource chain and
  assemble basic context.
- **Requirement:** The platform MUST ship a support diagnostic CLI (and
  equivalent API) that resolves a tenant identity to accounts, resources,
  quotas, recent operations, and billing posture in one call, with output for
  humans and machines. Diagnostic views MUST be redacted by default (no
  secret values, no tenant payload data) and scoped to the support role's
  permission level, and SHOULD offer an aggregated per-tenant summary
  (identity, quota usage, resource inventory, open incidents, entitlement
  status) to cut time-to-context. All diagnostic access MUST be logged with
  actor, ticket, and purpose.
- **Acceptance evidence:** End-to-end evidence against synthetic tenant
  fixtures (resolve to summary); redaction test suite covering every output
  mode; access-log schema checks (actor, ticket, purpose present);
  role-scoping tests.
- **Non-goals:** Tenant-facing self-service diagnostics (portal domain);
  deep per-service debug tooling (service teams own theirs behind OCS support
  surfaces).
- **Non-claims:** Coverage across all platform services is aspirational until
  services implement the OCS support surface; performance on large
  inventories is unmeasured.
- **Stop conditions:** Refuse diagnostic queries lacking purpose/ticket
  binding at privileged scopes; halt and treat as a security incident any
  redaction failure that emits secrets or tenant payloads (exposure/data
  risk).
- **Traceability:** legacy-platform-b (one-stop support CLI; aggregated
  per-customer support views); req-history; current-core (OCS support
  surface contracts). Related: CR-OPS-110, CR-OPS-120, CR-OPS-140.

### CR-OPS-140 — Audited support impersonation
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, auditor, provider
- **Problem:** Support and on-call staff sometimes must act inside a tenant
  context; shared admin accounts and standing broad rights make such access
  unaccountable and are a classic privilege-escalation incident source.
- **Requirement:** Privileged support access to tenant contexts MUST use
  impersonation, not shared credentials: a distinct, separately grantable
  permission; bound to a ticket and an explicit reason; time-boxed with
  automatic expiry; and fully audited (who, when, which tenant, what actions,
  why). Impersonated sessions MUST be visibly marked as such in audit trails
  and SHOULD be marked in tenant-visible logs. Direct modification of tenant
  data by staff outside an impersonated, ticket-bound session MUST be
  classified as a security incident. Break-glass emergency access follows the
  same audit bar with retrospective ticket binding; its mechanics are defined
  in the IAM domain.
- **Acceptance evidence:** End-to-end evidence of ticket-bound, time-boxed
  impersonation with complete audit events; permission-denial tests without
  the explicit grant; audit-gap alarm tests (any impersonation without full
  audit triggers an incident); expiry enforcement tests.
- **Non-goals:** The underlying permission model and token mechanics (IAM
  domain); tenant-admin delegation UX.
- **Non-claims:** Impersonation flows are specified but not implemented;
  marking in tenant-visible logs is a design goal without evidence.
- **Stop conditions:** Any impersonation attempt without ticket binding, a
  valid grant, or complete audit MUST be denied and escalated; any audit gap
  for a completed impersonated session is a security incident; emergency use
  auto-expires at incident closure and requires retrospective approval within
  a deadline (trust/keys risk).
- **Traceability:** legacy-platform-b (duty-role impersonation via dedicated
  accounts; break-glass discipline); req-history (impersonation-only admin
  paths). Related: CR-OPS-120, CR-OPS-130.

### CR-OPS-150 — Version-skew and compatibility-window policy
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, vendor
- **Problem:** Upgrades across control plane, nodes, CLIs, connectors, and
  service APIs fail in the gaps when supported version combinations are
  undefined; silent downgrade paths corrupt state.
- **Requirement:** The platform MUST publish, per release train, a
  version-skew policy: supported skew between control plane and
  nodes/agents, between platform and OCS connector/API versions, and between
  CLI/SDK and server. Upgrades MUST proceed only within the declared
  compatibility window, with rollouts versioned and observable. Downgrades
  MUST be blocked where state migration is not proven reversible, and
  blocked-downgrade attempts MUST fail loudly with guidance. The policy MUST
  state the supported upstream-Kubernetes minor-version window consistent
  with the substrate policy, and drift between declared and actual component
  versions MUST be detectable.
- **Acceptance evidence:** Compatibility matrix as code with CI gate tests
  (skew violations rejected); upgrade end-to-end evidence on test stands
  across the declared window; blocked-downgrade refusal evidence;
  version-drift detection reports.
- **Non-goals:** Mandating one upgrade cadence for all providers; per-service
  release pipelines (deployment domain).
- **Non-claims:** The window is defined by policy but has not been exercised
  across a real multi-version fleet; downgrade coverage exists only for
  current installation flows.
- **Stop conditions:** Refuse an upgrade outside the compatibility window;
  refuse a downgrade to a blocked version; halt a rollout on version drift or
  failed health gates mid-upgrade and invoke the rollback plan (CR-OPS-090)
  (migration/data risk).
- **Traceability:** req-history (versioned rollouts, compatibility windows);
  legacy-platform-a (supported-version ranges; toolchain skew lessons);
  legacy-platform-b (pinned-toolchain lessons); current-core (substrate
  policy). Related: CR-OPS-090, CR-OPS-100.

### CR-OPS-160 — Capacity management and hygiene automation
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Without deliberate capacity practice, platforms discover
  exhaustion during tenant incidents; forgotten test environments and stale
  artifacts quietly consume the same pool.
- **Requirement:** The platform SHOULD provide capacity reporting over fleet
  and quota signals (usage trends, headroom per failure domain, projection to
  exhaustion) with a defined review cadence and expansion runbooks
  (CR-OPS-170). Products that oversubscribe MUST declare and enforce their
  ratios (CR-OPS-110) and include oversubscription headroom in capacity
  views. The platform SHOULD automate hygiene garbage collection — stale
  test/feature environments, superseded artifacts, expired images — with
  dry-run first, allow-lists, TTLs, and audit records, since uncollected
  garbage is a capacity and exposure risk.
- **Acceptance evidence:** Capacity report generation from metrics fixtures
  with snapshot-verified outputs; review-cadence calendar artifacts;
  garbage-collection dry-run evidence and audited live-run records on a test
  stand; oversubscription headroom checks.
- **Non-goals:** Autoscaling of tenant workloads (compute domain); chargeback
  and showback reporting (billing domain).
- **Non-claims:** Forecasting accuracy is unproven — projections are advisory
  signals, not commitments; no long-run fleet telemetry exists; garbage
  collection has not run against production inventories.
- **Stop conditions:** Garbage collection MUST dry-run and require approval
  beyond small thresholds; halt collection on any deletion-candidate
  classification uncertainty; escalate when projected exhaustion falls inside
  procurement/expansion lead time; refuse oversubscription beyond declared
  policy (money/data/deletion risk).
- **Traceability:** legacy-platform-a (oversubscription enforcement;
  stale-release and registry garbage collection); legacy-platform-b
  (flow/telemetry capacity analytics); req-history. Related: CR-OPS-110,
  CR-OPS-170.

### CR-OPS-170 — Runbooks as code
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, agent
- **Problem:** Operational knowledge trapped in wiki pages and individual
  memory is unexecutable, untestable, and silently stale; agents cannot
  safely use it.
- **Requirement:** Operational runbooks MUST be authored as versioned,
  machine-readable documents: prerequisites, expected observations, step
  contracts, validation checks, and stop conditions are structured fields,
  not prose-only. Runbooks that drive automation MUST be executable against a
  test stand in CI. Every runbook MUST carry an owner and a freshness/review
  marker, and operational changes that invalidate a runbook MUST update it or
  record a no-change rationale (CR-OPS-080). Runbooks MUST declare their
  action risk class (CR-OPS-020) and SHOULD publish reproducible dependency
  manifests for the tools they use.
- **Acceptance evidence:** Runbook schema with CI validation of required
  structured fields; executed-runbook evidence from a test stand for the
  core set (deploy, upgrade, backup-restore verification, incident triage);
  freshness/ownership lint; agent-consumption dry runs.
- **Non-goals:** Replacing human judgment for novel incidents; per-service
  runbook content (service teams own theirs under the same contract).
- **Non-claims:** Only installation/preflight runbooks exist in current-core;
  the machine-readable runbook schema is not finalized; no CI-executed
  runbook corpus exists yet.
- **Stop conditions:** An automated runbook whose executed behavior drifts
  from its declared contract MUST be blocked from unattended execution; a
  runbook hitting an undeclared stop condition halts and pages rather than
  improvising (trust/data risk).
- **Traceability:** req-history (machine-readable runbooks);
  legacy-platform-b (duty handbooks); legacy-platform-a (runbook
  conventions). Related: CR-OPS-020, CR-OPS-080, CR-OPS-160.

### CR-OPS-180 — Disaster-recovery drill calendar
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Recovery capability that is never exercised is a claim, not a
  capability; drills done ad hoc slip and quietly rot.
- **Requirement:** Every production installation MUST maintain a versioned
  disaster-recovery drill calendar: scheduled, recurring exercises covering
  at minimum backup restore verification, point-in-time recovery where
  promised, node/zone failure (including one-server-loss), control-plane
  failover, and the drain/recovery completeness of CR-OPS-040. Each drill
  MUST produce dated evidence (scope, procedure, observed versus expected
  RPO/RTO, defects found, follow-up tickets) and MUST feed readiness gates: a
  failed or stale drill blocks readiness claims for the covered capability.
  Drill scope MUST be approved against blast radius before execution, and
  production-touching drills follow the change-management path (CR-OPS-090).
- **Acceptance evidence:** Drill calendar as a versioned artifact; completed
  drill evidence records for the minimum set on a reference stand (including
  one-server-loss); readiness-gate tests showing blocked claims on failed or
  stale drills; follow-up ticket linkage checks.
- **Non-goals:** Backup/restore and replication mechanics themselves (storage
  domain); chaos-engineering tooling selection.
- **Non-claims:** No drill has been executed on a production-class stand yet;
  all readiness claims depending on drills are accordingly blocked; RPO/RTO
  figures are targets, not measurements.
- **Stop conditions:** Halt any drill whose live blast radius exceeds its
  approved scope; a failed drill freezes dependent readiness claims until
  re-run or remediated; missing current backup/restore evidence blocks any
  destructive drill (data/deletion risk).
- **Traceability:** req-history (drills as readiness gates); current-core
  (one-server-loss and restore evidence conventions); legacy-platform-a.
  Related: CR-OPS-040, CR-OPS-090.

### CR-OPS-190 — Duty and support notification automation
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** On-call and support workflows lose minutes to manual status
  posting, schedule lookup, and ticket-chasing across channels.
- **Requirement:** The platform SHOULD support chat-integrated duty and
  support automation: incident push notifications to team channels,
  duty-schedule queries, escalation notices, and ticket status summaries. Bot
  and notification credentials MUST come from the platform secrets workflow —
  never from home-directory or repository-committed configs — and
  notification payloads MUST be redacted to the minimum needed (no secrets,
  no tenant payloads). Notification routing MUST be explicit configuration as
  code, auditable, and testable against non-production channels.
- **Acceptance evidence:** Routing configuration as code with CI validation;
  integration tests against test channels; secrets-workflow wiring checks
  (source-safety scan finds no bot tokens); payload redaction tests.
- **Non-goals:** Replacing the ticket system of record (CR-OPS-120);
  tenant-facing chatbots.
- **Non-claims:** No notification automation is implemented; the value is
  proven in legacy operations but unproven on this platform.
- **Stop conditions:** Any notification path leaking secrets or tenant
  payloads is disabled and treated as a security incident; missing or expired
  bot credentials fail closed with no fallback to personal tokens
  (exposure/keys risk).
- **Traceability:** legacy-platform-b (duty/support bots; bot-credential
  sprawl lesson); legacy-platform-a. Related: CR-OPS-060, CR-OPS-120.

## Coverage notes

This domain deliberately defers:

- **Observability signals** — metrics/logs/traces standards, golden signals,
  alerting engines, dashboards-as-code, host-level health-check bundles, and
  SLO computation — to domain 20 (OBS). OPS consumes those signals for
  severity classification, on-call, and drills.
- **Identity mechanics** — permission model, break-glass issuance, access
  reviews, token/certificate machinery, secrets brokering — to domain 15
  (IAM). OPS defines only the support/impersonation workflow built on them.
- **Backup, restore, replication, and PITR mechanics** and durability audits
  — to domain 13 (STO). OPS schedules the drills and gates readiness on them.
- **Deployment mechanics** — IaC state discipline, CI/CD pipelines, the
  environment ladder and direct-to-production prohibition, image/template
  pipelines, bootstrap ordering — to domain 22 (DPL). OPS governs the change
  discipline around them, not the pipelines themselves.
- **Usage metering, rating, settlement, and SLA-credit financial processing**
  — to domain 16 (BIL). OPS supplies ticket and SLA evidence.
- **Agent approval matrix and autonomous-agent risk taxonomy** — to domain 25
  (AGT). OPS action classes align with it but do not redefine it.
- **Console UX** for quota requests, support tickets, and maintenance notices
  — to domain 19 (CUX).
- **Per-service operational specifics** (database failover runbooks, queue
  partition recovery) — to service teams via OCS support/automation surfaces
  (domain 17); data-service specifics to domain 24 (DAT).
- **Federation and cross-provider operational trust**, including
  cross-provider incident duties — to domain 23 (FED).
- **Compliance and certification audit programs** (ISO/PCI-class) — to the
  security tracks of domain 15; OPS supplies the operational evidence trail.
