# 20 — Observability

Domain scope: the metrics, logs, traces, and alerting substrate for the whole
platform — a Prometheus-compatible metrics stack with operator-managed custom
resources, one service instrumentation library, golden signals with
outcome-aware metric classes, structured correlated logging, per-application
log pipelines over the platform bus, Jaeger-class distributed tracing with
declared retention, multi-tenant tenant-facing query and alerting surfaces,
alert rules and dashboards as code, central alert routing via GitOps,
notification delivery, synthetic probing, SLO/error-budget definitions,
UTC-canonical records, monitoring of the monitoring stack itself, capacity and
overbooking signals, observability data lifecycle, and observability evidence
as a hard readiness gate.

## Domain contract

Observability is a ship-blocking platform capability, not an add-on. Every
platform service is measurable, traceable, loggable, and alertable through
declared, Git-managed configuration before it may be called ready. Tenancy
isolation on every observability surface fails closed: an unresolvable scope
denies, never degrades to open. Telemetry pipelines never drop silently —
backlog, lag, and loss are first-class measured states, and any record carrying
secrets or unsanitized tenant payloads halts the pipeline. All operational
records are UTC-canonical. The observability stack is itself monitored with an
independent last-mile alert path. Blocked, stale, or synthetic observability
evidence is an honest state that must never be promoted into a readiness claim.

### CR-OBS-010 — Unified metrics platform

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** A provider cannot operate what it cannot measure. Ad-hoc
  per-service monitoring fragments operations, hides fleet-wide state, and
  makes every incident a bespoke archaeology exercise.
- **Requirement:** The platform MUST provide a unified, operator-managed
  metrics stack with Prometheus-compatible scrape, rule evaluation, and query
  semantics. Scrape targets, recording rules, and alert rules MUST be declared
  as Kubernetes custom resources reconciled by an operator. The stack MUST
  ingest platform-component, node, and Kubernetes control-plane signals by
  default on every installation. Long-term storage MUST be decoupled from
  per-cluster collection so that loss of one cluster does not erase fleet
  history. Metric endpoints of platform workloads MUST require authentication
  in shared environments.
- **Acceptance evidence:** conformance test suite that scrapes a reference
  service, evaluates a recording rule, and queries instant and range endpoints
  through the Prometheus-compatible API; live-stand evidence class showing
  node, control-plane, and service metrics for a reference installation with a
  freshness timestamp; drill evidence that a single-cluster loss preserves
  historical series in central storage.
- **Non-goals:** Not a mandate for one specific TSDB product — the contract is
  the Prometheus-compatible surface behind which implementations are
  replaceable. Does not cover tenant-facing query access (CR-OBS-070) or
  billing-grade metering (CR-OBS-230).
- **Non-claims:** No sustained-cardinality or retention-cost evidence at
  production scale yet; long-term storage sizing and compaction behavior
  unproven; no multi-zone disaster evidence for the storage tier.
- **Stop conditions:** Stop rollout if any platform metrics endpoint is
  reachable without authentication in a shared environment (exposure). Stop
  onboarding a service whose label cardinality degrades stack health beyond
  declared guards until bounded (data/capacity).
- **Traceability:** legacy-platform-a (operator-managed metrics stacks and
  per-platform TSDBs), legacy-platform-b (tiered metrics pipeline with
  agent/gateway/storage tiers), req-history (standardized metrics per service),
  current-core.

### CR-OBS-020 — Single supported instrumentation library

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator
- **Problem:** Services historically accumulate divergent observability
  libraries; every additional variant multiplies dashboard conventions, alert
  rule shapes, and onboarding cost, and divergence never converges on its own.
- **Requirement:** The platform MUST ship exactly one supported Go
  instrumentation library providing one-call setup for: the standard metric
  catalog (HTTP/gRPC/queue latency, counts, status classes, and runtime
  metrics), trace-context propagation, structured logging, and Kubernetes
  health probes. Platform services MUST adopt it; third-party OCS services
  SHOULD adopt it and MAY use alternative instrumentation only behind the same
  metric-naming, label, and trace-propagation conventions. The library MUST be
  incapable of emitting configured secret values into telemetry.
- **Acceptance evidence:** contract test suite asserting catalog metric names,
  labels, and probe endpoints emitted by a reference service; integration
  evidence that trace context propagates across an HTTP call and a message-bus
  hop in a two-service fixture without custom wiring; conformance check that
  every platform service exposes the catalog.
- **Non-goals:** Not a ban on additional service-specific metrics; not a client
  library for tenant application workloads; not a logging or tracing backend
  itself.
- **Non-claims:** The library does not yet exist as a published, versioned
  module with adoption evidence; runtime overhead benchmarks on hot paths are
  absent.
- **Stop conditions:** Stop any release of the library that can emit secret
  values, credentials, or unsanitized payload content into telemetry
  (keys/data); treat a shipped occurrence as a security incident with rotation.
- **Traceability:** legacy-platform-a (two divergent shared libraries lesson —
  ship one), req-history (standardized metrics, logs, tracing per service).

### CR-OBS-030 — Golden signals and outcome-aware metrics

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator, provider
- **Problem:** Error counters that cannot distinguish caller mistakes from
  platform failures produce misleading alerts and dishonest reliability claims;
  duplicated derivable metrics add cost without information.
- **Requirement:** Every key service MUST expose golden signals (latency,
  traffic, errors, saturation) and MUST classify operation outcomes at least as
  success / user error / platform error / retryable / policy-blocked. Metrics
  derivable from existing series by simple arithmetic SHOULD NOT be introduced
  as new families. Dashboards and alert rules MUST build on these outcome
  classes, and alert rules on unclassified error counters MUST fail linting.
- **Acceptance evidence:** per-service contract tests asserting the outcome
  label value set on operation counters; dashboard review checklist evidence
  verifying golden-signal coverage per key service; alert-rule linting evidence
  rejecting rules on unclassified counters.
- **Non-goals:** Does not set per-service SLO targets (CR-OBS-140); does not
  impose outcome classes on tenant-internal application metrics.
- **Non-claims:** The outcome taxonomy is not yet validated against real
  incident data; classification cost on latency-sensitive hot paths is
  unmeasured.
- **Stop conditions:** Stop promotion of any billing-relevant or
  readiness-relevant metric whose outcome classes are unverified
  (money/trust); stop on alerts in readiness evidence routed from unclassified
  error counters.
- **Traceability:** req-history (outcome-distinguished metrics, golden signals
  per key service, no duplicate derivable metrics), legacy-platform-a.

### CR-OBS-040 — Structured logs with correlation

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator, agent
- **Problem:** Free-text logs without correlation fields cannot be traced
  across services, become unsearchable at fleet scale, and multiply failure
  records for a single fault.
- **Requirement:** Platform services MUST emit structured logs (JSON-class)
  with mandatory fields: UTC timestamp, severity, service identity, and
  trace/request correlation identifiers; outcome class SHOULD be present where
  applicable. A failure SHOULD have a single owning log point that records it
  authoritatively, with downstream services logging propagation, not
  re-diagnosis. Log schema and field policy MUST be enforced by CI linting in
  platform repositories.
- **Acceptance evidence:** log-schema contract tests for a reference service;
  CI lint gate evidence rejecting non-conforming log statements; integration
  evidence that one failing request yields correlated lines across two services
  via a shared correlation identifier; failure-path evidence showing a single
  owning log point.
- **Non-goals:** Not a tenant application logging SDK; does not define log
  transport or storage (CR-OBS-050).
- **Non-claims:** Schema not yet frozen; no evidence on per-service-class log
  volume or the cost of the mandatory field set at scale.
- **Stop conditions:** Stop on detection of secret values, credentials, or
  unsanitized tenant payload content in any log stream (keys/data): halt the
  emitting service's pipeline, scrub, rotate, and audit before resuming.
- **Traceability:** req-history (structured logs with correlation fields,
  single owning log point), legacy-platform-a (shared logger conventions).

### CR-OBS-050 — Per-application log pipelines

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator, provider
- **Problem:** Hand-configured log shipping per service drifts between
  environments, drops records silently, and cannot be audited or reproduced.
- **Requirement:** Each service MUST be able to enable log shipping through one
  declarative toggle in its deployment contract. The platform MUST render
  per-application collection flows — with parsing rules, sidecar/proxy
  container exclusion, and environment-specific destinations — onto the durable
  message bus and from the bus to indexed storage. Delivery MUST be
  at-least-once with visible backlog and drop counters; flow definitions MUST
  be Git-managed and change with the service they describe.
- **Acceptance evidence:** end-to-end test that enables the toggle on a
  reference service and asserts parsed records in storage with correct
  application labels; negative evidence that excluded proxy containers are
  absent from the stream; backlog/drop metrics present on the pipeline
  dashboard; disruption evidence of no silent loss across a broker restart.
- **Non-goals:** Not a general-purpose data streaming product; tenant-facing
  log products are an OCS service concern, not this requirement.
- **Non-claims:** Bus-mediated log delivery is not yet load-tested at provider
  scale; storage/index cost model unproven; parse-failure quarantine behavior
  not yet specified.
- **Stop conditions:** Stop on sustained backlog growth or non-zero drop
  counters without an operator-visible alert firing (data). Stop immediately on
  any flow configuration that routes one tenant's records toward another
  tenant's destination (exposure); contain and audit.
- **Traceability:** legacy-platform-a (per-app collection flows with parsing
  and sidecar exclusion to bus-then-storage destinations), req-history
  (backpressure and no-silent-drop ingestion contract).

### CR-OBS-060 — Distributed tracing with declared retention

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator, provider
- **Problem:** Without traces, cross-service latency and failure attribution
  across many services and regions devolves into guesswork and prolonged
  incidents.
- **Requirement:** The platform MUST provide deploy-toggled distributed tracing
  (Jaeger-class) with trace-context propagation across synchronous APIs and the
  message bus. Collector and query components MUST be declaratively deployable
  per environment, with the query surface exposed only through authenticated
  mesh ingress. Span storage MUST have a declared, environment-scoped retention
  policy. Services using the standard library (CR-OBS-020) MUST NOT need custom
  wiring to participate.
- **Acceptance evidence:** integration test showing one trace spanning HTTP
  service A → message bus → worker B with parent/child linkage; deployment
  evidence that the toggle enables agent/collector on a reference stand;
  retention configuration rendered from environment values and verified against
  the storage backend; sampling policy documented per environment.
- **Non-goals:** Not an APM product for tenants; not full-fidelity capture of
  every request — sampling ratios are an operations decision.
- **Non-claims:** Sampling ratios and storage footprint at production traffic
  unproven; no measured overhead on latency-critical paths; no cross-region
  trace-assembly evidence.
- **Stop conditions:** Stop on span attributes containing secrets or
  unsanitized payload data (keys/data) — halt exporters, scrub, audit. Stop on
  silent failure of retention deletion jobs beyond one declared window (data).
- **Traceability:** legacy-platform-a (deploy-toggled tracing via shared
  libraries, environment-exposed query UI, declared span retention),
  legacy-platform-b (tracing as a bootstrap dependency of managed services),
  req-history (distributed tracing with context propagation).

### CR-OBS-070 — Multi-tenant metrics query API

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Tenants need programmatic visibility into their own services;
  giving them direct TSDB access breaks isolation, while no access makes the
  platform a black box they cannot operate on.
- **Requirement:** The platform MUST expose a Prometheus-compatible query API
  (instant, range, series, labels) that enforces tenant scoping per project and
  service on every request. Tenancy enforcement MUST be server-side, derived
  from authenticated identity — never from client-supplied labels or
  parameters. The API MAY fan out to per-region metrics backends behind one
  contract. Any failure to resolve tenant scope MUST fail closed. Cross-scope
  denials MUST be audit-logged.
- **Acceptance evidence:** authorization test suite proving tenant A cannot
  read tenant B series under label-injection, path-manipulation, and
  header-forgery attempts; API conformance tests against the
  Prometheus-compatible surface; fan-out evidence against two regional
  backends; audit log samples of cross-scope denials.
- **Non-goals:** Not tenant write access to platform metrics; not a
  billing-accuracy query surface (see CR-OBS-230); not alert management
  (CR-OBS-110).
- **Non-claims:** Query-path isolation validated only in test fixtures so far;
  no multi-tenant load evidence; query fairness, throttling, and cost guards
  unproven.
- **Stop conditions:** Treat any cross-tenant data return as a security
  incident: halt the API path, contain, audit, and notify per policy
  (trust/exposure). On identity-provider outage or ambiguity, the API MUST
  deny — never degrade to open or cached-broad access (trust).
- **Traceability:** legacy-platform-a (per-tenant PromQL proxy fanning out to
  per-platform backends), req-history (party-scoped visibility), current-core.

### CR-OBS-080 — Alert rules as code per service repository

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator
- **Problem:** Alert rules managed through platform-team tickets drift from the
  service they measure, lag releases, and create a permanent toil queue.
- **Requirement:** Every service MUST declare its alert rules in its own
  repository at a declared path with a deployment-contract toggle. CI/CD MUST
  render them into rule custom resources scoped to that service's identity,
  with no platform-team ticket in the path. Rendering MUST validate rule
  syntax, label conventions, and cardinality guards before apply, and rule
  changes MUST deploy together with the service version they measure. Rendered
  rules MUST carry owning-service identity labels.
- **Acceptance evidence:** pipeline evidence rendering a reference service's
  rules into custom resources and applying them; validation-failure evidence
  for a malformed rule and for a cardinality-exploding expression; audit
  evidence of service-identity labels on rendered rules; a service-team-only
  merge completing the full flow without operator action.
- **Non-goals:** Not tenant-authored alerts (CR-OBS-110); not routing or
  receiver configuration (CR-OBS-090).
- **Non-claims:** The validation rule set is initial; its false-positive rate
  against legitimate complex expressions is unknown.
- **Stop conditions:** Stop on any rendered rule that escapes the owning
  service's scope or can select another service's or tenant's series (trust).
  Stop on rules whose expressions can explode series cardinality beyond
  declared guards (data/capacity).
- **Traceability:** legacy-platform-a (repo-local rule files plus deployment
  toggle rendered to rule CRs without tickets).

### CR-OBS-090 — Central alert routing configuration via GitOps

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Hand-edited alert routing with manual version-bump steps creates
  toil, environment drift, and unreviewed changes to who gets paged — and
  partial automation guarantees both.
- **Requirement:** Alert routing and receiver configuration MUST be fully
  declarative in Git and rolled out exclusively by the platform's GitOps path.
  Pre-merge validation (configuration linter plus routing-tree dry run) MUST be
  blocking. Every routing change MUST be traceable to a reviewed commit, and
  there MUST be no manual intermediate version-bump or pin steps between merge
  and rollout. Production routing state MUST be continuously reconciled against
  Git with drift alerting.
- **Acceptance evidence:** GitOps evidence of a routing change flowing from
  merge to applied state with zero manual steps; pre-merge CI failing on an
  invalid configuration; applied-state verification evidence on a stand;
  rollback evidence reverting a bad route by revert commit; drift-detection
  evidence on manual out-of-band edit.
- **Non-goals:** Does not define per-channel secrets (CR-OBS-100); not
  tenant-level routing self-service; not on-call scheduling (OPS domain).
- **Non-claims:** The GitOps-only flow is not yet exercised through a real
  paging incident; validation coverage of exotic routing trees unproven.
- **Stop conditions:** Stop on any production routing state diverging from Git
  without a corresponding commit (trust). Stop rollout of a configuration whose
  dry run shows alerts silently dropped or misrouted (must fail validation).
- **Traceability:** legacy-platform-a (central routing configuration with
  manual chart-bump toil — lesson: automate end-to-end or fully declare in
  repo).

### CR-OBS-100 — Notification gateway and alert ingress

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, service-team
- **Problem:** Alerts must reach humans through chat, email, and webhook
  channels whose credentials are per-channel and sensitive; ad-hoc integrations
  leak secrets into repositories and lose alerts silently.
- **Requirement:** The platform MUST provide a notification gateway converting
  alert-manager webhooks into channel deliveries, with per-channel secrets held
  only in the approved secrets workflow and referenced, never embedded. A
  generic authenticated alert-ingress webhook endpoint MUST exist per
  environment. Delivery failures MUST be retried with bounded backoff and
  surfaced as their own metric and alert. Channel configurations MUST be
  Git-managed with secret references only.
- **Acceptance evidence:** end-to-end evidence of a test alert traversing
  alert manager → gateway → a chat channel fixture; source-safety scan evidence
  that configurations contain secret references only; retry-evidence for a
  failing channel; negative authentication tests against the ingress endpoint.
- **Non-goals:** Not a full incident-management or on-call product (OPS
  domain); not tenant-facing notification products.
- **Non-claims:** Channel adapters beyond the initial set unproven; no
  delivery-latency evidence at alert-storm volumes; channel-specific formatting
  contract not frozen.
- **Stop conditions:** Stop and rotate on any channel secret material detected
  outside the secrets workflow (keys). Stop on unauthenticated or spoofable
  alert ingress (trust) — fail closed until fixed.
- **Traceability:** legacy-platform-a (alert-to-chat bridge with per-channel
  credentials, per-environment webhook ingress, secrets-in-repo leak lesson).

### CR-OBS-110 — Tenant self-service alerting and dashboards

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider
- **Problem:** Tenants of a provider need to define their own alert rules and
  views over their own services without filing operator tickets; without this,
  the tenant-facing query API is read-only visibility with no operational
  agency.
- **Requirement:** Tenants SHOULD be able to create, update, and delete alert
  rules for their own services through an API and console surface, rendered
  into rule resources hard-scoped to the tenant's projects. Tenants SHOULD be
  able to store custom dashboard definitions through the same surface. All
  tenant objects MUST be scoped by server-side identity; rule evaluation MUST
  execute within the tenant's data scope only; per-tenant rule and dashboard
  counts MUST be quota-bounded.
- **Acceptance evidence:** API test suite for alert and dashboard CRUD
  including cross-tenant denial cases; rendered-rule inspection evidence
  proving enforced scope labels; console walkthrough evidence on a stand;
  quota enforcement test evidence.
- **Non-goals:** Not service-team platform alerts (CR-OBS-080); not arbitrary
  cluster-wide recording rules for tenants; not a tenant dashboard design
  system.
- **Non-claims:** The full CRUD surface is not yet implemented; abuse and
  denial-of-service guards on tenant rule counts are unproven; evaluation cost
  isolation between tenants unmeasured.
- **Stop conditions:** Stop on any tenant rule able to reference out-of-scope
  series (trust/exposure) — treat as isolation incident. Stop on unbounded rule
  counts per tenant without quota enforcement (capacity).
- **Traceability:** legacy-platform-a (tenant alert CRUD and stored user
  dashboards via API).

### CR-OBS-120 — Dashboards as code

- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, operator, provider
- **Problem:** Hand-built dashboards in a UI are unreviewable, unportable, and
  disappear on reinstall; naive automation with implicit semantics makes file
  renames delete production dashboards.
- **Requirement:** Dashboard definitions MUST live in Git (raw JSON or
  Jsonnet-class generators) and deploy via CI/GitOps with a dry-run diff.
  Dashboards MUST use stable explicit UIDs and declared folder ownership.
  Deletion MUST be explicit and reviewable — renames MUST NOT silently delete.
  Platform key services MUST have their golden-signal dashboards defined this
  way, and local preview with provisioned data sources SHOULD be one command.
- **Acceptance evidence:** CI evidence deploying a dashboard change with diff
  preview; negative evidence that a rename produces a create+delete plan
  requiring explicit approval rather than implicit deletion; restore evidence
  re-materializing all dashboards after a dashboard-store rebuild; drift
  detection on manual UI edits.
- **Non-goals:** Not a dashboard design system; tenant user dashboards are
  CR-OBS-110; not panel-plugin development.
- **Non-claims:** Generator toolchain not finally selected; no evidence for
  fleets of hundreds of dashboards; local-preview parity unproven.
- **Stop conditions:** Stop on any deploy path with implicit destructive
  semantics (deletion). Stop on manual UI edits overwriting Git-managed
  dashboards without drift detection and reconciliation (trust).
- **Traceability:** legacy-platform-a (tag-deployed dashboards-as-code with
  rename-deletes lesson and dashboard tooling), legacy-platform-b
  (repo-generated dashboards from specs).

### CR-OBS-130 — Synthetic monitoring service

- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, tenant
- **Problem:** Passive metrics miss gray failures — every service reports
  healthy while user journeys are broken; the provider learns of outages from
  customers.
- **Requirement:** The platform SHOULD provide a synthetic monitoring service:
  projects of probe actions (HTTP method, expected status/body, timeouts),
  severity escalation rules with delays, an incident list API, and webhook
  notifications. Platform-critical user journeys (console sign-in, core API
  create/read paths) SHOULD have continuous probes per environment. Probe
  source ranges MUST be documented and allow-listable, and probes MUST be
  read-only or operate on synthetic fixtures.
- **Acceptance evidence:** fixture project where a failing probe produces an
  incident and a notification within the declared delay; probe-source
  allow-list evidence; API conformance tests for projects, results, and
  incidents; probe-vs-real-failure drill evidence.
- **Non-goals:** Not a replacement for service golden signals; not browser-based
  end-to-end test automation (CI concern); not network dataplane probing (NET
  domain).
- **Non-claims:** Probe scheduling at fleet scale unproven; false-positive
  rates unknown; multi-region probe placement not yet designed.
- **Stop conditions:** Stop on probe definitions carrying reusable credentials
  in plain form (keys) — probes use brokered, scoped credentials or none. Stop
  on any probe that mutates tenant data (data); synthetic fixtures only.
- **Traceability:** legacy-platform-a (synthetic health-check service with
  probe actions, escalation rules, incidents), legacy-platform-b (active
  probers for gray-failure detection).

### CR-OBS-140 — SLO/SLI definitions with error budgets

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** Without declared objectives, reliability is unfalsifiable and
  alert thresholds are arbitrary; retroactively edited objectives make breaches
  invisible.
- **Requirement:** Each key service SHOULD declare SLIs (built from the
  CR-OBS-030 outcome classes), SLO targets, and error-budget policy as code
  alongside its alert rules. Burn-rate alerts SHOULD derive from these
  definitions. Every SLO document MUST state its data lineage (which metrics,
  which windows, which exclusions) and MUST NOT be retroactively edited in ways
  that mask an active breach; changes are versioned, reviewed commits.
- **Acceptance evidence:** SLO definition fixtures rendered into recording and
  burn-rate alert rules; error-budget report evidence over a rolling window on
  a stand; governance check evidence that SLO changes are reviewed versioned
  commits with lineage stated.
- **Non-goals:** Not contractual SLA terms or settlement computation with
  customers (billing/commercial domains consume SLO data; they do not define
  it here).
- **Non-claims:** Burn-rate alerting not yet validated against real incident
  timelines; SLO targets for platform services not yet derived from measured
  baselines; error-budget policy consequences (release freezes) undefined.
- **Stop conditions:** Stop on SLO data feeding any commercial settlement,
  credit, or invoicing flow without verified lineage and reconciliation
  (money/settlement). Stop on retroactive SLO redefinition during an active
  breach window (trust); escalate to owner review.
- **Traceability:** req-history (support-useful observability outcomes
  including SLO/SLA relation), legacy-platform-b (launch stages tied to SLA
  terms).

### CR-OBS-150 — UTC-canonical observability records

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, auditor, agent
- **Problem:** Mixed timezones across logs, metrics, spans, alerts, and
  incident records make correlation unreliable and audits contestable.
- **Requirement:** All observability records — log lines, metric timestamps,
  span times, alert events, incident records — MUST be stored and transmitted
  in UTC. Display localization MUST happen only at presentation boundaries.
  Clock-skew handling MUST be explicit: records with unresolvable clock
  ambiguity MUST be flagged rather than silently accepted, and skew beyond a
  declared tolerance MUST itself be an observable, alertable condition.
- **Acceptance evidence:** schema tests asserting UTC across emitted record
  types; skew-injection test evidence showing flagged records and skew alerts;
  documentation conformance check for the presentation-boundary rule; audit
  sample review of stored records.
- **Non-goals:** Not NTP/clock infrastructure management (platform foundation
  domain); not a ban on localized rendering in consoles.
- **Non-claims:** Skew tolerance thresholds not yet tuned against measured
  fleet clock behavior; no evidence for behavior during clock-source outages.
- **Stop conditions:** Stop on non-UTC records entering audit, billing-adjacent,
  or incident stores (data integrity): halt the emitting pipeline until
  corrected, and quarantine the ambiguous window.
- **Traceability:** req-history (canonical UTC for all operational records with
  clock-skew handling and display-localization boundary), legacy-platform-a
  (UTC-only platform decision).

### CR-OBS-160 — Monitoring of monitoring

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** If the observability stack fails silently, the platform is blind
  exactly when it needs visibility most — and discovers it during the next real
  incident.
- **Requirement:** The observability stack MUST be monitored as a first-class
  workload: collection agents, rule evaluators, storage, alert manager, and the
  notification gateway MUST emit health metrics and alerts. A dead-man's-switch
  alert MUST fire when the alerting path itself goes silent. The last-mile
  self-alert path MUST run in an independent failure domain from the stack it
  watches. Any window during which the alerting path was unobserved MUST be
  recorded as an incident-class gap.
- **Acceptance evidence:** failure-injection drill evidence — the alert manager
  is disabled and the dead-man alert arrives through the independent path
  within a declared time; self-monitoring dashboard evidence; runbook for
  observability outage; recorded-gap audit trail.
- **Non-goals:** Not infinite recursion (meta-meta-monitoring); one independent
  last-mile path is sufficient. Not a second full observability stack.
- **Non-claims:** The independent-path topology is not yet deployed or drilled;
  dead-man timing tolerances unset.
- **Stop conditions:** Treat any unobserved-alerting window as an
  incident-class gap (trust): freeze readiness promotions for services whose
  monitoring depended on the failed stack during that window until evidence is
  re-established.
- **Traceability:** legacy-platform-b (self-hosted monitoring plane for the
  platform's own components), req-history (evidence honesty; blocked is a
  first-class state).

### CR-OBS-170 — Multi-cluster metrics fan-in over the event bus

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Edge zones and many clusters cannot all be scraped centrally,
  and direct remote-write couples edge collectors to central storage
  availability, losing metrics during outages.
- **Requirement:** The platform SHOULD support metrics fan-in in which
  per-cluster agents remote-write onto the durable message bus and a reader
  forwards into central storage. Bus buffering MUST make edge and central
  outages survivable without metric loss up to declared bounds; backlog and lag
  MUST be first-class metrics with alerts. Bus topics and ACLs for metrics MUST
  follow platform bus governance — pipeline-declared, never created by
  applications in shared environments.
- **Acceptance evidence:** end-to-end evidence of metrics surviving a central
  storage outage via bus buffering within declared bounds, with replay
  completeness verification; backlog/lag metrics on dashboards with alert
  thresholds; governance check evidence that topics and ACLs are
  pipeline-declared.
- **Non-goals:** Not a metrics-database replication protocol; not real-time
  guarantees — lag is tolerated, measured, and alerted, not hidden.
- **Non-claims:** Buffer sizing and replay throughput unproven at fleet scale;
  bus cost model for metrics volume unknown; ordering/duplication semantics
  under replay not yet characterized.
- **Stop conditions:** Stop on silent metric loss at buffer overflow — overflow
  MUST alert and account dropped ranges with proof (data). Stop on bus ACLs
  permitting cross-tenant metric reads (exposure); contain and audit.
- **Traceability:** legacy-platform-a (edge-agent → bus → central storage
  fan-in service with authenticated bus access), req-history (bounded queues
  and backpressure states).

### CR-OBS-180 — Non-Kubernetes estate discovery bridge

- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Provider estates include VMs and bare metal outside Kubernetes;
  excluding them from the monitoring plane creates blind islands where the
  oldest and riskiest workloads live.
- **Requirement:** The platform SHOULD provide declarative service discovery
  bridging non-Kubernetes estates (inventory sources, VM groupings) into
  scrape-target custom resources with per-job ports, labels, prefix filters,
  and refresh intervals. Discovery output MUST be inspectable and reconciled —
  decommissioned targets disappear without hand edits. Target credentials MUST
  come from the secrets workflow as references only.
- **Acceptance evidence:** fixture inventory source rendered into scrape-target
  resources and scraped on a stand; reconciliation evidence removing a
  decommissioned target; source-safety scan evidence of secret-reference-only
  configuration; refresh-interval behavior test.
- **Non-goals:** Not a CMDB; not agent installation or host provisioning
  (foundation domain); not application-level instrumentation of those hosts.
- **Non-claims:** Only one inventory-source class considered so far; behavior
  under high target churn unproven; duplicate/overlapping target deduplication
  unspecified.
- **Stop conditions:** Stop on scrape credentials appearing outside the secrets
  workflow (keys). Stop on discovery writing resources outside its declared
  scope or namespace set (trust).
- **Traceability:** legacy-platform-a (inventory-driven scrape-target bridges
  and client-cluster onboarding into central monitoring).

### CR-OBS-190 — Capacity and overbooking signals

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** Without fleet capacity and oversubscription visibility, a
  provider cannot plan growth, silently breaches performance promises, and
  discovers exhaustion through tenant incidents.
- **Requirement:** The platform SHOULD publish capacity signals per zone and
  resource class: allocated versus physical capacity, current utilization, and
  headroom, plus the declared oversubscription ratio per resource. These
  signals MUST be visible to operators with alert thresholds at declared safety
  margins. Any tenant-visible capacity or availability claim MUST be derivable
  from these signals, never asserted independently. Ratio policy changes MUST
  be reviewed, recorded configuration.
- **Acceptance evidence:** capacity dashboard evidence on a stand showing
  per-zone ratios and headroom; alert evidence on crossing a declared margin;
  ratio policy document with review history; cross-check evidence that a
  published capacity claim traces to the signals.
- **Non-goals:** Not automated capacity purchasing or provisioning (deployment
  domain); not billing metering (BIL domain); not performance benchmarking of
  individual workloads.
- **Non-claims:** Ratio policy values are not yet set from measured workload
  profiles; no evidence under a genuinely saturated fleet; headroom forecasting
  is out of scope and unproven.
- **Stop conditions:** Stop on oversubscription exceeding declared policy
  without recorded operator sign-off (trust/money-adjacent). Stop on tenant
  capacity claims published without backing signals (trust).
- **Traceability:** legacy-platform-a (mandatory resource floors and scaling
  conventions), legacy-platform-b (capacity planned and priced in resource
  units with sizing guidance).

### CR-OBS-200 — Observability data lifecycle: retention, isolation, deletion

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, tenant, auditor
- **Problem:** Telemetry accumulates unboundedly, silently mixes tenant data
  classes, and is forgotten at offboarding — creating cost, compliance, and
  trust risk.
- **Requirement:** Every observability data class (metrics, logs, traces, alert
  history, incident records) MUST have a declared retention policy per
  environment, rendered into backend configuration. Tenant-scoped telemetry
  MUST be stored with enforced isolation. Deletion of a tenant's telemetry at
  offboarding MUST follow the platform deletion contract and produce proof
  records. Retention, deletion, and hold events MUST be auditable, and deletion
  MUST honor active retention holds and open incident windows.
- **Acceptance evidence:** retention configuration rendered per environment and
  verified against backends; isolation tests across tenants on shared storage;
  offboarding drill evidence with deletion proof records; audit trail samples
  of lifecycle events including a blocked-deletion case under hold.
- **Non-goals:** Not backup/DR of telemetry as a durability promise — telemetry
  is operational data, though alert and incident history SHOULD survive their
  declared windows. Not tenant data-of-record deletion (storage domain).
- **Non-claims:** Retention cost curves unproven; the deletion-proof record
  format is not yet implemented; cross-class consistency (deleting logs but not
  traces) is undefined.
- **Stop conditions:** Stop deletion when a retention hold, open incident, or
  compliance window applies (deletion/data). Stop immediately on discovery of
  cross-tenant storage mixing (exposure): halt writes, contain, audit, and
  notify per policy.
- **Traceability:** req-history (data discipline and redaction-before-memory),
  legacy-platform-a (per-environment retention configuration practice).

### CR-OBS-210 — Observability evidence as a readiness gate

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, vendor, auditor
- **Problem:** Services historically reach "production" while unobservable, and
  the first proof of missing monitoring is a customer-visible incident with no
  data behind it.
- **Requirement:** No platform service or OCS connector service MUST be
  declared ready to serve production workloads without linked observability evidence:
  golden-signal dashboards (CR-OBS-030), alert rules as code (CR-OBS-080),
  tracing participation (CR-OBS-060), log pipeline (CR-OBS-050), and runbook
  references. Readiness reports MUST classify observability evidence as
  verified / blocked / stale / synthetic, and blocked evidence MUST NOT be
  converted into readiness claims. The gate MUST be enforced by the platform's
  iteration/conformance tooling, not by convention.
- **Acceptance evidence:** readiness-gate check failing a fixture service that
  lacks alert rules; gate report evidence listing the five evidence classes
  with per-class states; integration evidence with the platform iteration gate;
  audit trail of a blocked promotion that stayed blocked.
- **Non-goals:** It does not claim full production readiness — other domains
  gate durability, security, and other aspects. Not SLO target sign-off
  (CR-OBS-140).
- **Non-claims:** The gate is not yet wired into all service onboarding paths;
  evidence freshness windows are not yet tuned; no evidence of the gate
  operating across a full fleet release.
- **Stop conditions:** Stop any promotion attempt presenting blocked, stale, or
  synthetic observability evidence as verified (trust): halt the promotion
  pipeline and escalate to owner review; repeated attempts trigger policy
  review.
- **Traceability:** req-history (conformance status taxonomy, blocked ≠ failed,
  no blocked-to-release conversion), current-core (iteration-gate convention),
  legacy-platform-a (per-service observability block convention).

### CR-OBS-220 — Blackbox probing of internal endpoints

- **Priority:** P2
- **Status:** proposed
- **Actors:** operator
- **Problem:** Some internal endpoints (API servers, ingress paths) need
  external-style reachability probes where "up" legitimately includes expected
  authentication rejections.
- **Requirement:** The platform MAY provide blackbox probing modules for
  internal endpoints with explicit expected-status definitions (for example,
  200, or 401/403 counted as reachable-but-protected). TLS verification MUST be
  on by default; any certificate-verification skip MUST be explicit,
  environment-scoped, and flagged in review, and MUST NOT appear in production
  scopes without a recorded exception.
- **Acceptance evidence:** probe fixture evidence distinguishing
  reachable-but-protected from down; CI check evidence flagging insecure probe
  modules in production-scoped configuration; module configuration tests.
- **Non-goals:** Not a substitute for synthetic user-journey monitoring
  (CR-OBS-130); not certificate lifecycle management; not external
  availability monitoring as a tenant product.
- **Non-claims:** The module set is minimal; no broad deployment evidence;
  expected-status catalogs per endpoint class not yet defined.
- **Stop conditions:** Stop on certificate-verification-skip probe modules in
  production scopes without an explicit recorded exception (trust/exposure).
- **Traceability:** legacy-platform-a (blackbox prober modules with
  auth-accepting reachability definitions and a TLS-skip caveat lesson).

### CR-OBS-230 — Product and billing-adjacent metrics over the metrics pipeline

- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, operator, provider
- **Problem:** Product usage signals with product labels are operationally
  useful for monitoring metering health, but they must never be confused with
  settled commercial truth or become a shadow billing path.
- **Requirement:** Product metering services MAY expose usage metrics with
  product labels into the metrics pipeline for operational monitoring. Such
  metrics MUST carry an operational-only marker. Any commercial consumption
  MUST go through the billing pipeline contract with its
  accepted-for-processing semantics — never through observability queries.
  Dashboards and reports MUST NOT present observability usage metrics as
  billing truth, and divergence between operational usage metrics and billing
  pipeline records SHOULD itself be monitored.
- **Acceptance evidence:** labeled-metric fixture evidence on a stand;
  documentation check that operational-only marking is present; cross-check
  evidence that reconciliation consumes the billing pipeline rather than
  observability queries; divergence-alert fixture evidence.
- **Non-goals:** Not metering ingestion, rating, or charging (BIL domain owns
  the usage pipeline); not a replacement for mediation/replay of billable
  events.
- **Non-claims:** No evidence yet of drift behavior between operational usage
  metrics and billing pipeline records; reconciliation tooling unproven.
- **Stop conditions:** Stop immediately if observability usage metrics are
  consumed by charging, rating, invoicing, or settlement flows
  (money/settlement): treat as a money-path integrity incident and sever the
  consumption path.
- **Traceability:** legacy-platform-a (product usage metrics observed on the
  metrics pipeline with product labels), req-history (accepted-for-processing
  is never settled commercial truth).

## Coverage notes

This domain deliberately defers:

- **Usage ingestion, metering contracts, charging, rating, and cost
  visibility** — to `domains/16-billing-finops.md` (BIL). Observability
  monitors the metering pipeline's health; it never defines commercial truth.
- **Incident management, on-call rosters, escalation policy, runbook authoring,
  and support SLA measurement** — to `domains/21-ops-sre-support.md` (OPS).
  This domain supplies the alert and incident *signals*; OPS owns the response
  process.
- **CI/CD mechanics, the GitOps engine, deployment pipelines, and environment
  promotion** — to `domains/22-deployment-iac-cicd.md` (DPL). This domain
  declares *what* is rendered (rules, flows, dashboards, routing); DPL owns
  *how* rendering and rollout work.
- **Authentication, authorization, service accounts, and the secrets workflow**
  referenced by the query API, ingress endpoints, and channel credentials — to
  `domains/15-iam-identity-security.md` (IAM).
- **Console presentation** of tenant dashboards, alert management, and
  status views — to `domains/19-portal-ux-selfservice.md` (CUX); this domain
  owns the APIs and data contracts behind them.
- **OCS connector observability surface declarations** (health, readiness, and
  monitoring metadata a connector package must declare) — to
  `domains/17-ocs-service-connectors.md` (OCS); this domain consumes those
  declarations.
- **Federation-level telemetry exchange, cross-provider status, and settlement
  visibility** — to `domains/23-federation-global-portal.md` (FED).
- **Network dataplane probing depth and flow telemetry** — to
  `domains/12-network.md` (NET); this domain covers the platform-wide synthetic
  and blackbox probing contracts only.
- **Clock/NTP infrastructure and host-level base agents** — to
  `domains/10-platform-foundation.md` (FND); this domain only mandates the
  UTC-canonical record contract that depends on them.
