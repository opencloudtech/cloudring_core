# 19 — Portal, UX, and Self-Service

Domain scope: the unified cloud console and every surface through which tenants,
operators, vendors, and agents see and change platform state. Covers
cross-surface parity (UI / API / CLI / IaC / agent), the canonical resource
state vocabulary, information architecture (navigation areas, project home,
create hub), pre-commit review of cost/defaults/exposure, risk-classed
confirmations, the structured error envelope, honest availability and readiness
states, reference-only secrets, inert rendering of user content, deletion and
exit semantics, party-scoped evidence bundles, the micro-frontend host with a
runtime registry, support handoff, internationalization, accessibility, and the
journey scenarios that acceptance-test the whole surface.

Domain contract: the console never lies. Every fact shown to a human is equally
available, in structured form, to API, CLI, IaC, and agent clients; every state
shown is the real state with its freshness; every cost, default, and exposure
consequence is disclosed before commit, never after; every destructive,
commercial, or security-relevant action is confirmed by risk class and leaves
an audit record and a redacted evidence bundle. Secrets are references, user
content is inert, gated capabilities stay visible with their reason and next
step, and management surfaces fail closed. A capability that exists only as a
backend API is not a shipped feature: it ships when its journey, states,
errors, evidence, and support path are all present and testable.

---

### CR-CUX-010 — Cross-surface parity contract

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent, auditor
- **Problem:** When UI, API, CLI, IaC, and agent outputs disagree about names,
  states, permissions, or cost semantics, self-service breaks silently and
  automation acts on stale or wrong facts. Legacy platforms accumulated
  divergent per-surface behavior that made agent operation and support
  reconstruction unreliable.
- **Requirement:** Every console capability MUST be reachable through UI, API,
  CLI, and declarative IaC with the same resource identity, the same state
  vocabulary (CR-CUX-020), the same action and risk classes, the same required
  permissions, the same policy results, the same billing/cost semantics, the
  same support owner, the same evidence references, and the same non-claims.
  Agent-facing structured output MUST expose the same facts a human sees
  (state, allowed actions, risk class, validation result, evidence, redaction
  status) rather than screenshots or UI-only affordances. Any parity gap MUST
  be declared explicitly per capability, never discovered by the user.
- **Acceptance evidence:** automated parity test suite that performs the same
  operation set through UI (browser e2e), API, CLI, and IaC plan and diffs
  resource identity, states, errors, and cost fields; contract checks that
  every console route has a backing API operation; a published parity-gap
  registry with zero undeclared gaps.
- **Non-goals:** pixel-level sameness across clients; requiring IaC coverage
  for purely presentational actions (saved views, UI preferences).
- **Non-claims:** full IaC/agent parity for every optional capability at first
  release is not claimed; gaps are permitted only as declared registry entries.
- **Stop conditions:** halt and escalate when a surface is found to mutate
  state, disclose cost, or report readiness differently from another surface
  for the same resource — treat as a trust defect, not a UX bug.
- **Traceability:** vision-deck (production honesty principle), req-ccp
  (parity dimensions and must-not-have gates), legacy-platform-a (divergent
  per-brand surfaces as anti-pattern), current-core (self-service contract,
  portal browser e2e parity checks). Related: CR-CUX-020, CR-CUX-060.

### CR-CUX-020 — Canonical resource state vocabulary

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent
- **Problem:** Ad-hoc per-service status strings ("created", "active",
  "pending", "error") make readiness ambiguous, hide degraded and disputed
  states, and prevent agents and support tooling from reasoning uniformly
  across resource families.
- **Requirement:** The platform MUST define one canonical lifecycle state
  vocabulary of twelve states — ready, provisioning, activating,
  request-pending, degraded, blocked, stale, disputed, retired, deleting,
  deleted, unsupported — used identically by every resource family and every
  surface. Each state MUST define its meaning, allowed and prohibited actions,
  required UX (e.g. `stale` shows last-updated time, a refresh action, and
  locks freshness-dependent operations; `disputed` shows frozen scope, allowed
  operations, owner, and evidence), and the support path. Resource-specific
  sub-states MAY exist but MUST map onto the canonical vocabulary.
- **Acceptance evidence:** a versioned state-vocabulary contract document plus
  machine-readable schema; conformance tests verifying every resource type in
  the console reports only canonical states; UI snapshot tests showing the
  required affordances per state; parity tests confirming identical state
  values across UI/API/CLI/IaC/agent output.
- **Non-goals:** forbidding internal fine-grained states inside a service, as
  long as the published state maps to the canonical set.
- **Non-claims:** the twelve-state set is proposed as complete for known
  families; new families may extend sub-states, and any extension of the
  canonical set requires a contract revision.
- **Stop conditions:** halt a release when any resource family publishes a
  state outside the canonical vocabulary without a contract revision, or when
  `deleted`/`deleting` resources remain actionable beyond their declared
  recovery window.
- **Traceability:** req-ccp (12-state vocabulary with per-state UX),
  legacy-platform-a (hub-side vs connector-side status taxonomies that had to
  be reconciled), current-core (experience-standard and self-service contract
  lessons). Related: CR-CUX-010, CR-CUX-070.

### CR-CUX-030 — Information architecture: navigation, project home, create hub

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team
- **Problem:** Consoles that grow feature-by-feature bury primary tasks, mix
  organization-level and project-level facts, and hide the entry point for
  creating resources, producing dead ends and mis-scoped actions.
- **Requirement:** The console MUST provide: (a) top-level navigation grouped
  by user mental model — home, create, compute, network, storage, containers,
  data, AI workloads, security, marketplace, billing, support, organization,
  governance — with organization-level and project-level areas visually and
  semantically distinct and context transitions disclosed; (b) a project home
  showing project identity and state, primary next action, billing/cost
  status, health and readiness gaps, create entry, access and quota links,
  documentation, and exits to organization/billing/support — which MUST NOT
  hide missing billing, quota, or access prerequisites, present
  organization-level mutations as project-local actions, expose unauthorized
  management controls, or imply readiness when the project is gated or
  blocked; (c) a search-first create hub whose cards name the user outcome
  (never internal object names), state purpose in one sentence, show
  availability state, required permission class, and cost/readiness warnings,
  and cover governance actions (quota request, policy exception, approval) as
  first-class cards.
- **Acceptance evidence:** navigation and page-structure conformance tests;
  browser e2e journeys reaching any primary task in at most three actions from
  project home; create-hub contract tests asserting card anatomy fields are
  populated from connector/capability metadata rather than hard-coded text;
  negative tests proving project home surfaces missing prerequisites instead
  of hiding them.
- **Non-goals:** a final visual design system; this requirement fixes
  structure and disclosure, not aesthetics.
- **Non-claims:** the nav area list is the proposed baseline; optional areas
  (batch/accelerators) may appear as gated entries (CR-CUX-070) before their
  backends exist.
- **Stop conditions:** halt when a page exposes an organization-scope mutation
  under project scope, or when project home reports a healthy/ready summary
  while billing, quota, or access prerequisites are unmet.
- **Traceability:** req-ccp (15 nav areas, project-home must/must-not lists,
  create-hub card anatomy), legacy-platform-a (project-centric hub UX as the
  proven model). Related: CR-CUX-070, CR-CUX-140.

### CR-CUX-040 — Pre-commit review of cost, defaults, and exposure

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent
- **Problem:** Hidden defaults, generated names, implicit public exposure, and
  undisclosed cost are the most repeated sources of tenant surprise and
  dispute in cloud consoles. Review-after-the-fact is not acceptable for
  commercial or security-relevant choices.
- **Requirement:** Every material create or change flow MUST end in a
  pre-commit review step showing: resource identity and scope, placement
  (region/zone), all applied defaults and generated names, cost estimate with
  billing start/stop conditions, quota impact, policy evaluation result, data
  touched, secret and dependency impact, access and exposure changes (public
  reachability MUST be an explicit, opt-in, visibly flagged choice — never a
  side effect), rollback/delete/export path, and support owner. Progressive
  disclosure MUST NOT hide risk, cost, or security fields under "advanced" UI.
  Cancel and browser-back MUST never mutate state. IaC and CLI clients MUST
  expose the equivalent as a dry-run/plan.
- **Acceptance evidence:** browser e2e tests asserting review-step content for
  representative create flows across families; API contract tests for
  dry-run/plan endpoints returning the same cost/policy/exposure fields;
  negative tests proving public exposure requires explicit opt-in; audit
  records linking each commit to the reviewed summary.
- **Non-goals:** guaranteeing price accuracy beyond the estimate class
  published by the billing domain; the review shows the estimate and its
  basis, not a settlement-grade invoice.
- **Non-claims:** cost estimates for third-party marketplace offers are shown
  with their provenance and are not claimed settlement-grade.
- **Stop conditions:** halt the flow and escalate when cost, placement, policy
  result, or exposure data is unavailable or stale at review time — block
  commit rather than let the user accept unknowns (fail closed on money and
  exposure).
- **Traceability:** vision-deck (no hidden cost/defaults/public exposure
  before commit), req-ccp (review-step requirement, hidden-defaults lesson),
  legacy-platform-b (budgets advisory, cost disclosure conventions).
  Related: CR-CUX-050, CR-BIL (billing semantics), CR-IAM (policy result).

### CR-CUX-050 — Risk-classed confirmations

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent
- **Problem:** Uniform "Are you sure?" dialogs train users to click through
  genuinely destructive, irreversible, commercial, or cross-boundary actions,
  while low-risk actions drown in confirmation noise.
- **Requirement:** Every action MUST carry a risk class (read-only, safe,
  controlled, risky, destructive). Destructive, irreversible, commercial,
  security-relevant, cross-project/cross-provider, and data-moving actions
  MUST require explicit confirmation that states the risk class, the affected
  resources and their count/scope, the consequence summary, the compensation
  or rollback path, and — where appropriate — typed confirmation (resource
  name or phrase). The confirmation and its result MUST be audit-recorded.
  The same risk classes MUST drive agent approval requirements and appear in
  API/CLI output.
- **Acceptance evidence:** a risk-class registry mapping every console action
  to its class; e2e tests proving typed confirmation is enforced for
  destructive-class actions and that confirmation content includes scope,
  consequence, and compensation; audit-trail verification that confirmations
  are recorded with actor, target, reason, timestamp, and result.
- **Non-goals:** preventing all destructive actions; the goal is informed,
  attributable consent, not prohibition.
- **Non-claims:** the initial risk-class assignment per action family is
  proposed and requires owner review; misclassification policy is defined but
  not yet exercised.
- **Stop conditions:** halt and block the action when risk class is unknown or
  unclassified for a money/data/keys/exposure/deletion action — fail closed to
  the destructive-class confirmation; escalate repeated confirmation bypass
  attempts as a security event.
- **Traceability:** req-ccp (explicit confirmation requirement, agent approval
  matrix linkage), legacy-platform-a (surprising destructive semantics such as
  rename-equals-delete as anti-pattern), current-core (agent approval matrix
  requirement). Related: CR-CUX-100, CR-AGT (approval matrix).

### CR-CUX-060 — Structured, actionable error envelope

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent, provider
- **Problem:** Opaque errors ("something went wrong") kill self-service,
  multiply support tickets, and make agent operation impossible, because
  neither humans nor machines can decide what to do next.
- **Requirement:** Every error on every surface MUST use one structured
  envelope: stable machine-readable code, cause, impact, next step, validation
  state, retryability (safe/unsafe/unknown), owner/support path, and
  correlation identity. The envelope MUST be identical in UI rendering, API
  responses, CLI output, and agent-facing structured results. Degraded and
  blocked states MUST list what remains possible (read-only actions, safe
  retries, blocked mutations, owner, escalation path). Internal error detail
  MUST be redacted to a safe public form on tenant-facing surfaces, with full
  detail available only to authorized operator scopes.
- **Acceptance evidence:** error-envelope schema with contract tests across
  all public APIs; UI tests rendering envelope fields (including
  retryability and next step); parity tests diffing envelope content across
  surfaces; e2e tests for degraded-state pages showing remaining possible
  actions; redaction tests proving internal detail never leaks to
  unauthorized scopes.
- **Non-goals:** a single global error-code namespace for all services;
  per-domain code registries are allowed if codes are stable and documented.
- **Non-claims:** full next-step coverage for every backend failure mode is
  not claimed at first release; unknown failures MUST still carry the
  envelope with retryability `unknown` and a support path.
- **Stop conditions:** halt a release when any public mutation path returns an
  unstructured or unenveloped error, or when an error message discloses
  internal endpoints, stack traces, credentials, or tenant data.
- **Traceability:** req-ccp (error envelope acceptance gate), legacy-platform-a
  (nested vendor error tree as the proven shape), legacy-platform-b (semantic
  error mapping with internal-only detail). Related: CR-CUX-010, CR-CUX-120.

### CR-CUX-070 — Honest availability and readiness states, no dead ends

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, service-team, provider
- **Problem:** Capabilities that silently disappear from the console when not
  installed, gated, or request-only create dead ends for users and agents;
  conversely, presenting gated or preview capabilities as ready manufactures
  false readiness claims.
- **Requirement:** Every capability and page MUST carry an explicit
  availability state — available, activation-required, request-only, preview,
  gated, not-installed, degraded, blocked, retired — with the reason, the
  allowed and denied actions, the expected next state, documentation, and the
  support path. The capability matrix MUST keep visible entries for optional
  or future surfaces instead of letting them vanish. Activation and request
  gates MUST state who can activate, who gains access, the price/free/stage
  boundary, readiness non-claims, and the expected next state, and MUST write
  an audit record. Readiness claims MUST derive from operational evidence
  (health, freshness, drills), never from the mere existence of a resource.
- **Acceptance evidence:** capability-matrix registry with per-entry state and
  reason, validated in CI against installed components; gate-page contract
  tests asserting the required content items; audit verification for
  activation/request events; e2e tests proving gated entries remain visible
  and navigable with their state and next step.
- **Non-goals:** requiring every optional capability to be installable in
  every installation; honest absence is the requirement.
- **Non-claims:** the availability model does not claim any optional
  capability (marketplace, managed data services, batch/accelerators) is
  ready to serve production workloads; their entries exist with explicit states.
- **Stop conditions:** halt and escalate when a capability is shown as
  available while its readiness evidence is missing, stale, or blocked —
  downgrade to the honest state rather than carry an unproven claim.
- **Traceability:** vision-deck (production honesty, evidence over claims),
  req-ccp (availability vocabulary, hidden-readiness lesson), current-core
  (fail-closed management access checks). Related: CR-CUX-020, CR-CUX-030,
  CR-CUX-210.

### CR-CUX-080 — Secrets shown as references only

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent, auditor
- **Problem:** Consoles that can display or export raw secret values turn one
  UI compromise or one over-privileged support session into a full credential
  breach; agent surfaces multiply that risk.
- **Requirement:** The console, API, CLI, and agent output MUST present
  secrets exclusively as references: stable identifier, fingerprint, version,
  rotation/expiry status, access policy, and reconciliation/sync state. Raw
  secret value retrieval MUST be denied by default on every surface, including
  support and agent paths; secret use MUST go through brokered actions
  (inject, rotate, bind) rather than disclosure. Error messages involving
  secrets MUST be redacted. The UI MUST function fully — diagnostics, support,
  rotation, audit — without ever revealing a raw value.
- **Acceptance evidence:** contract tests proving no console/API/CLI/agent
  path returns raw secret material by default; e2e tests for reference-only
  rendering (fingerprint/version visible, value absent); negative tests for
  agent raw-value requests being denied and audit-recorded; redaction tests
  on error payloads.
- **Non-goals:** implementing the secrets engine itself (owned by the IAM/
  security domain); this requirement governs its presentation and brokered
  use.
- **Non-claims:** none beyond the domain boundary; the presentation contract
  is fully testable once the secrets backend exists.
- **Stop conditions:** halt and treat as a security incident when any surface
  returns, logs, or renders raw secret material; block the offending path
  until remediated and audit-reviewed.
- **Traceability:** vision-deck (secrets are never configuration), req-ccp
  (reference-not-raw lesson, denied agent raw access journey),
  legacy-platform-a (real secret leaks in repos and configs as the motivating
  failure), current-core (secret runtime readiness evidence requirement).
  Related: CR-IAM (secrets engine), CR-CUX-110.

### CR-CUX-090 — User-generated content rendered inertly

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, vendor, provider
- **Problem:** Resource names, tags, support case subjects, marketplace
  metadata, and other user-controlled strings are injection vectors; a console
  that renders them as markup exposes every user to stored cross-site
  scripting.
- **Requirement:** All user-generated or tenant-controlled strings MUST render
  inertly — escaped, never interpreted as HTML/script — on every surface, and
  MUST carry provenance marking in structured output so agents and support
  tooling can distinguish platform-generated from user-generated content.
  Content-security policy MUST be enforced at the shell level; third-party
  embedded UI MUST run under context isolation (CR-CUX-130).
- **Acceptance evidence:** browser e2e tests injecting script-bearing names,
  tags, subjects, and marketplace metadata and asserting inert rendering;
  CSP headers verified in e2e; structured-output contract tests asserting
  provenance fields; regression suite covering every list/detail/notification
  surface.
- **Non-goals:** sanitizing user content into "safe HTML"; the default is
  inert text, not filtered markup.
- **Non-claims:** none; this is fully testable.
- **Stop conditions:** halt release on any rendered-injection finding; treat a
  live XSS path as a security incident with exposure assessment.
- **Traceability:** req-ccp (user content as injection surface, XSS-safe
  rendering gate), current-core (portal browser e2e XSS checks).
  Related: CR-CUX-130.

### CR-CUX-100 — Deletion and exit semantics on every resource

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, provider, auditor
- **Problem:** Deletion is simultaneously a data event and a commercial event;
  consoles that hide residual data, retained backups, dependents, or the
  billing stop time produce disputes and lock-in by design.
- **Requirement:** Every deletable resource MUST present, before confirmation:
  dependents and incompatible items, residual data and retained backups with
  retention periods, billing stop time, the recovery window and appeal path,
  and the export/portability path including non-exportable boundaries.
  Deletion MUST follow CR-CUX-050 confirmation classes, write a final audit
  record, and leave a post-deletion view (deleting → deleted states per
  CR-CUX-020) showing what remains and until when. Exit and export flows MUST
  state scope, format/protocol class, cost, duration estimate, and policy
  checks — exit is a product capability, not an obstruction.
- **Acceptance evidence:** e2e deletion journeys per resource family asserting
  pre-confirmation disclosure content; audit verification of deletion records;
  contract tests for export-flow fields; billing-domain cross-checks that
  billing stop time displayed matches billing behavior.
- **Non-goals:** defining per-service retention policy (owned by
  storage/backup and billing domains); this requirement governs honest
  disclosure of whatever the policy is.
- **Non-claims:** export-path availability varies by service; non-exportable
  boundaries are disclosed, not eliminated.
- **Stop conditions:** halt deletion and escalate when dependent billing,
  retained-backup, or recovery-window data is unavailable at confirmation
  time, or when a billing account would be detached while resources still
  bill to it — fail closed on money and data loss.
- **Traceability:** vision-deck (portability and jurisdiction freedom),
  req-ccp (deletion/cleanup views, exit-as-capability), legacy-platform-b
  (suspension/data-retention timelines as the disclosure model).
  Related: CR-CUX-050, CR-BIL, CR-STO.

### CR-CUX-110 — Operation evidence bundles, party-scoped and redacted

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, provider, vendor, auditor
- **Problem:** When an operation goes wrong across provider, vendor, and
  tenant boundaries, nobody can reconstruct what happened without a bundle of
  facts; unscoped bundles leak one party's data to another.
- **Requirement:** Every console operation MUST produce a support-safe
  evidence bundle: actor and represented subject, scope, resource, action,
  parameter summary, validation result, result, timestamps (canonical UTC),
  correlation identity, redaction status, and next owner. Bundles MUST be
  party-scoped — tenant, provider, vendor, and marketplace parties each see
  only their scoped, redacted view — and MUST be retrievable from the
  resource page and the support flow. Critical actions MUST additionally write
  audit records with actor/subject/target/reason/timestamp/result/correlation.
- **Acceptance evidence:** bundle schema with contract tests; e2e tests
  generating a bundle from a failed operation and opening it in the support
  flow; party-scoping tests proving cross-party fields are redacted;
  audit-trail verification for critical actions.
- **Non-goals:** the audit log storage product itself (owned by observability/
  security domains); this requirement covers the per-operation bundle and its
  presentation.
- **Non-claims:** redaction rules per party are proposed; jurisdiction-specific
  redaction profiles are not yet defined.
- **Stop conditions:** halt cross-party evidence sharing when redaction
  coverage for a field class is unverified — withhold the field rather than
  leak it.
- **Traceability:** vision-deck (evidence over claims), req-ccp (evidence
  bundle acceptance gate, multi-party scoping pitfall), legacy-platform-a
  (operations log as first-class record). Related: CR-CUX-060, CR-CUX-120,
  CR-OBS.

### CR-CUX-120 — Support handoff with audit trail

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, operator, provider, vendor
- **Problem:** Support that starts with repeated fact-gathering wastes the
  incident window; support contexts that cannot cross provider/vendor
  ownership boundaries leave marketplace and federation users stranded.
- **Requirement:** A support case MUST be openable from any resource, error
  envelope, or evidence bundle, pre-filled with category, affected resource,
  severity/impact, subject, description, attachments, and the operation
  evidence bundle with consent and redaction preview (what is included, what
  sensitive fields are excluded). Cases MUST show number, status, owner,
  SLA/SLO posture, timeline, resource links, and update/creation times, and
  MUST support ownership handoff across provider and vendor boundaries with
  the handoff audit-recorded.
- **Acceptance evidence:** e2e journey from a failed operation to a submitted
  case carrying the redacted bundle; consent/redaction preview tests;
  cross-ownership handoff tests with audit verification; support-case list
  contract tests for the required fields.
- **Non-goals:** the support/ticketing backend and SLA engine (owned by the
  ops/support domain); this covers the console handoff contract.
- **Non-claims:** SLA posture display depends on the support backend; where no
  SLA is contracted, the UI shows an explicit "no SLA" state rather than an
  implied promise.
- **Stop conditions:** halt case submission when the redaction preview cannot
  be generated — never submit evidence the user could not review.
- **Traceability:** req-ccp (support starts with context, case model),
  legacy-platform-a (support ticket surfaces in the panel). Related:
  CR-CUX-110, CR-OPS.

### CR-CUX-130 — Micro-frontend host with runtime registry

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, service-team, vendor, operator
- **Problem:** A console that must be redeployed to add, remove, or disable a
  product couples every service team's release to the platform shell;
  historically this produced mixed composition technologies and forked
  per-brand frontends that drifted irreconcilably.
- **Requirement:** The console MUST be a shell plus a runtime registry:
  products are registered as entries (name, application identity, route,
  enabled flag, bundle location) and can be enabled/disabled by configuration
  without redeploying the shell. The platform MUST standardize on ONE
  composition technology and ONE design-system package with mandatory tokens.
  Embedded service UI MUST mount through a typed descriptor (host slot,
  required context, permissions, lifecycle hooks, telemetry, support owner),
  run under context isolation with authority containment, degrade locally on
  failure (never breaking the whole console), and pass a certification cycle
  (mount/update/suspend/unmount/failure-cleanup/retry) before registration.
- **Acceptance evidence:** registry service contract tests
  (register/enable/disable/route resolution); e2e test disabling a product
  via registry change only; extension certification suite covering the full
  lifecycle and failure-injection (extension crash leaves shell functional);
  design-system conformance checks in CI for registered modules.
- **Non-goals:** prescribing a specific framework; the requirement is one
  composition technology and the registry contract, not a brand of bundler.
- **Non-claims:** third-party (vendor) UI mounting beyond first-party service
  teams is not claimed until the certification pipeline has live evidence;
  white-label/theming beyond token configuration is a later extension.
- **Stop conditions:** halt registration of an embedded UI whose descriptor
  requests context or permissions beyond its certified scope — fail closed at
  the host boundary.
- **Traceability:** legacy-platform-a (registry service with import-map as the
  durable idea; mixed single-spa/Module-Federation sprawl and per-brand forks
  as anti-patterns), req-ccp (UI extension certification requirements),
  current-core (portal skeleton, OCS metadata-driven rendering).
  Related: CR-CUX-090, CR-OCS.

### CR-CUX-140 — Scope visibility and fail-closed management surfaces

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, provider, auditor
- **Problem:** Users acting in the wrong organization or project scope cause
  cross-tenant mistakes; management surfaces mounted before authorization
  invite privilege-boundary attacks.
- **Requirement:** Current organization, project, actor, and permission scope
  MUST be visible on every page; scope switches MUST disclose the context
  transition, and a denied switch MUST leak nothing about the target scope.
  Management-only surfaces MUST stay hidden and fail closed until IAM
  explicitly allows; no privileged UI may mount on unauthorized routes.
  Organization-level navigation (projects, billing accounts, access
  management, federation/trust relationships, support) MUST be distinct from
  project-level navigation. Every critical action MUST audit-record actor,
  represented subject, target, reason, timestamp, result, and correlation
  identity.
- **Acceptance evidence:** browser e2e parity checks (login gate,
  tenant/project switch, fail-closed management access) green; negative tests
  proving denied scope switches return no target-scope data; penetration-test
  evidence that management routes deny by default; audit verification for
  critical actions.
- **Non-goals:** the IAM policy model itself (owned by the IAM domain); this
  covers its console enforcement and visibility.
- **Non-claims:** none; the fail-closed checks are already executable in the
  existing portal e2e skeleton.
- **Stop conditions:** halt and treat as a security incident when any
  management surface renders or responds for an unauthorized principal;
  block the route class until reviewed.
- **Traceability:** vision-deck (fail closed), req-ccp (IAM-gated visibility,
  denied-switch leak rule), current-core (production identity docs, console
  fail-closed policy). Related: CR-IAM, CR-CUX-030.

### CR-CUX-150 — Idempotent mutations with operation identity

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent, service-team
- **Problem:** Ambiguous failures make humans and agents retry; without
  operation identity and idempotency, retries double-create resources,
  double-charge, or corrupt state.
- **Requirement:** Every mutating console action MUST be bound to an operation
  identity: validate synchronously, register the operation, return its ID,
  and expose operation state, logs, and safe retry. Retry after an ambiguous
  failure MUST be idempotent (duplicate behavior documented, conflict and
  quarantine handling defined) or explicitly marked unsafe in the UI, API,
  CLI, and agent output. Users MUST be able to inspect and — where the domain
  allows — cancel a running operation.
- **Acceptance evidence:** contract tests asserting every mutation path
  returns an operation ID with status tracking; fault-injection e2e tests
  (kill connection mid-mutation, retry) proving no duplicate side effects;
  UI tests rendering operation state/logs/cancel; documentation checks that
  duplicate/conflict behavior is stated per operation family.
- **Non-goals:** the async execution engine (owned by platform foundation);
  this covers the console/agent contract over it.
- **Non-claims:** cancel support varies by operation family and phase;
  non-cancellable phases are disclosed, not hidden.
- **Stop conditions:** halt a release when any mutation lacks operation
  identity or when a retry of a money-moving or resource-creating operation
  is shown to double-apply — quarantine the operation family.
- **Traceability:** req-ccp (idempotent retry requirement), legacy-platform-a
  (async-by-contract with operation IDs and operations log; idempotency
  library built late as the lesson), legacy-platform-b (durable task queue
  with single-active-operation invariant). Related: CR-FND, CR-CUX-060.

### CR-CUX-160 — Accessibility baseline

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, operator, provider
- **Problem:** Consoles that only work with a mouse, full vision, and a wide
  screen exclude users and fail precisely during incident-time sessions on
  constrained devices.
- **Requirement:** The console MUST meet a defined accessibility baseline:
  landmarks, labels, focus order, full keyboard operability, error-to-field
  linkage, and minimum contrast, plus usable behavior on narrow and embedded
  viewports. Dangerous actions MUST be visually distinct from navigation.
  Accessibility MUST be a release gate for the shell and for every registered
  module (CR-CUX-130), not an after-the-fact audit.
- **Acceptance evidence:** automated accessibility checks (axe-class) in CI
  for shell and module routes with zero critical violations; manual keyboard-
  only completion evidence for the core journeys (CR-CUX-180); responsive
  behavior tests at narrow viewports; contrast verification against the
  design tokens.
- **Non-goals:** certification against a specific national accessibility
  statute at this stage; the baseline is the WCAG-class checklist above.
- **Non-claims:** assistive-technology (screen reader) full-journey
  verification is planned but not yet evidenced; no formal certification is
  claimed.
- **Stop conditions:** n/a (no money/data/keys/exposure risk class; quality
  gate handled by release blocking).
- **Traceability:** req-ccp (accessibility as hard requirement), vision-deck
  (real open baseline — usable by all operators). Related: CR-CUX-130.

### CR-CUX-170 — Governed i18n terminology

- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, operator, provider, service-team
- **Problem:** Translating a console without a governed terminology catalog
  produces per-language drift in state names, action labels, and error text —
  which then breaks support scripts, documentation, and agent parsing.
- **Requirement:** The console SHOULD maintain a governed terminology catalog
  mapping canonical states (CR-CUX-020), actions, error codes, and resource
  names across supported languages, with structured output always carrying
  the canonical English keys alongside localized display strings. New
  languages SHOULD be addable by catalog contribution without code change.
- **Acceptance evidence:** terminology catalog in-repo with completeness
  checks per language; UI tests verifying structured output retains canonical
  keys under non-default locales; contribution workflow documentation.
- **Non-goals:** a full translation program at first release; one additional
  language beyond English is the initial target.
- **Non-claims:** translation quality coverage is community-driven and not
  guaranteed for all modules at any given time.
- **Stop conditions:** n/a.
- **Traceability:** legacy-platform-a (en/ru i18n in the console shell;
  per-partner locale forks as the anti-pattern), req-ccp (governed i18n
  catalog). Related: CR-CUX-020.

### CR-CUX-180 — Journey scenarios as acceptance contracts

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, operator, provider, agent, auditor
- **Problem:** Requirements per feature do not prove the product works; only
  end-to-end journeys — including the negative and degraded paths — expose
  the seams where cost, policy, readiness, and support contradict each other.
- **Requirement:** The platform MUST define and continuously execute a fixed
  set of journey scenarios as acceptance contracts, each with actor and
  permission scope, entry point, preconditions, primary steps, negative and
  degraded paths, evidence produced, success criteria, and explicit
  non-claims. The minimum set MUST cover: first-run onboarding from project
  home to first resource (including blocked-billing and gated-capability
  negatives), first VM create and operate, first service purchase through the
  marketplace (including request-only and missing-export-path negatives), and
  incident communications (degraded state, support handoff with evidence
  bundle, status visibility). Journeys MUST fail the release when their
  contracts break.
- **Acceptance evidence:** executable journey specs (browser e2e plus API/CLI
  variants per CR-CUX-010) run in CI and against reference installations;
  journey reports archived as evidence with pass/fail and non-claim status;
  traceability mapping from each journey to the requirements it exercises.
- **Non-goals:** exhaustive scenario coverage of every resource family; the
  journey set is the contract backbone, families add theirs over time.
- **Non-claims:** the journey set is proposed; journeys depending on
  unshipped capabilities (marketplace purchase) remain marked blocked until
  their domains land, and blocked journeys are never converted into release
  claims.
- **Stop conditions:** a failing money, deletion, exposure, or secrets path
  inside any journey halts the release train until resolved and re-evidenced.
- **Traceability:** req-ccp (12-journey acceptance suite and template),
  current-core (iteration gate discipline — blocked evidence stays blocked).
  Related: CR-CUX-010, CR-CUX-070, CR-CUX-120.

### CR-CUX-190 — Uniform page-pattern contracts with agent-readable equivalents

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, operator, service-team, agent
- **Problem:** Per-service page improvisation makes the console unlearnable
  for humans and unparseable for agents and support tooling; minimum field
  discipline per page type is what makes cross-service operation possible.
- **Requirement:** The console MUST enforce uniform page-pattern contracts:
  list pages (breadcrumb, create/activate/request action, search by stable
  fields, operationally meaningful columns mapped to stable structured field
  keys, empty state, row status and actions, support/audit link); detail
  pages answering the ten operator questions (what, owner and support owner,
  placement and policies, state and freshness, cost, data/secrets/dependencies
  touched, safe actions now, blocked/degraded/stale/disputed status,
  evidence, stop/rollback/export/migrate/delete paths); activation gates with
  the required content items (CR-CUX-070); create forms following the
  phased contract (type, identity/scope, placement and sizing, data
  protection, review, confirm, operation state and evidence). Every list and
  detail page MUST expose a structured data equivalent using the same stable
  field keys for agents.
- **Acceptance evidence:** page-pattern linting/contract tests applied to
  every registered module; minimum-column schema checks per resource family;
  structured-equivalent snapshot tests diffing rendered tables against their
  stable-key JSON form; service-team onboarding checklist referencing the
  contracts.
- **Non-goals:** visual uniformity beyond the contract fields; modules keep
  layout freedom inside the pattern.
- **Non-claims:** minimum column sets for all sixteen observed resource
  families are drafted; families beyond the observed set define theirs at
  onboarding.
- **Stop conditions:** halt module registration when a page's structured
  equivalent or rendered contract misstates cost, data/secrets touched, or
  the available stop/rollback/export/migrate/delete paths, or when a
  destructive action is presented without its declared confirmation gate —
  a wrong action map is a trust and deletion risk, not a UX bug; block the
  offending page pattern until the contract and its stable-key rendering
  agree.
- **Traceability:** req-ccp (list/detail/gate/form patterns, 10-question
  detail contract, stable field keys for agents), legacy-platform-a
  (operation-schema-driven dynamic UI serving many products).
  Related: CR-CUX-030, CR-CUX-010.

### CR-CUX-200 — Empty states and documentation at point of need

- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, operator, service-team
- **Problem:** Empty lists that say nothing strand new users; documentation
  divorced from the page where a decision is made forces context switches and
  guesses.
- **Requirement:** Empty states SHOULD act as onboarding: explain the
  resource, its value, the primary action, and a documentation link.
  Documentation SHOULD be linked at every point of need (create flows, gates,
  error next-steps). Onboarding emphasis SHOULD de-escalate once a project
  has active resources while remaining discoverable.
- **Acceptance evidence:** UI tests for empty-state content on primary list
  pages; link-integrity checks for in-context documentation; journey
  evidence (CR-CUX-180) that a first-time user reaches a created resource
  without leaving the console for answers.
- **Non-goals:** the documentation platform itself.
- **Non-claims:** documentation coverage completeness is not claimed at first
  release; pages without docs must show the support path instead.
- **Stop conditions:** n/a.
- **Traceability:** req-ccp (empty states as onboarding, docs at point of
  need). Related: CR-CUX-030, CR-CUX-180.

### CR-CUX-210 — Stateful honesty in console presentation

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, provider, auditor
- **Problem:** "Backup exists" is not "restore works", "created" is not
  "ready", and transport success is not accepted usage — consoles that blur
  these distinctions manufacture readiness and recoverability claims that
  fail exactly when tenants depend on them.
- **Requirement:** The console MUST NOT present backup existence as
  recoverability without restore-validation or restore-drill evidence, or an
  explicit "unproven" warning with its non-claim. RPO/RTO figures MUST be
  shown only when claimed and backed by linked evidence, with their as-of
  time. Cluster and data-service readiness MUST derive from operational
  evidence (health, freshness, drill status), not from provisioning success.
  Billing-related display MUST distinguish operation success from accepted
  usage, invoiced, and settled states. Stale evidence MUST be labeled stale
  with last-refreshed time and a refresh action, and freshness-dependent
  operations MUST lock while evidence is stale.
- **Acceptance evidence:** backup/restore detail-page contract tests
  (source, time, retention, validation status, restore targets, drill
  evidence or unproven warning); readiness-derivation tests proving displayed
  readiness tracks linked evidence; freshness/staleness e2e tests including
  operation lockout; cross-checks with the storage/backup domain's drill
  evidence.
- **Non-goals:** producing the restore drills themselves (owned by the
  storage/backup domain); this requirement governs honest presentation.
- **Non-claims:** no RPO/RTO values are claimed by the console layer; they
  are rendered only when the owning domain claims them with evidence.
- **Stop conditions:** halt and downgrade display to `blocked`/`stale` with
  an explicit warning when linked durability or readiness evidence is missing
  or expired — never present a recoverability or readiness claim without
  fresh evidence.
- **Traceability:** vision-deck (evidence over claims, stateful is
  first-class), req-ccp (backup ≠ restore readiness, success ≠ truth),
  legacy-platform-b (billing truth separated from operation success),
  current-core (stateful restore/failover readiness requirements).
  Related: CR-STO, CR-OBS, CR-CUX-070.

---

## Coverage notes

This domain deliberately defers:

- **Billing correctness, price calculation, invoices, settlement, and dispute
  mechanics** to `16-billing-finops.md` (BIL). CUX defines only the honest
  presentation of cost, billing status, and billing stop times.
- **IAM policy model, token issuance, secrets engine internals, and identity
  provider integration** to `15-iam-identity-security.md` (IAM). CUX consumes
  IAM decisions and enforces fail-closed presentation.
- **OCS connector lifecycle APIs, connector validation, and service-team SDK**
  to `17-ocs-service-connectors.md` (OCS). CUX renders from connector
  metadata and certifies embedded UI at the host boundary only.
- **Marketplace economics, publisher onboarding, revenue share, and catalog
  governance** to `18-marketplace-catalog.md` (MKT). CUX covers the purchase
  journey and gate/gate-state presentation.
- **Metrics/logs/traces pipelines, alert routing, and audit log storage** to
  `20-observability.md` (OBS). CUX presents freshness, health categories, and
  evidence references.
- **Support backend, SLA engine, on-call, and incident management process** to
  `21-ops-sre-support.md` (OPS). CUX defines the support handoff contract.
- **Async execution engine, operation registry storage, and quota two-phase
  protocol** to `10-platform-foundation.md` (FND). CUX binds every mutation
  to operation identity but does not own the engine.
- **Deployment, IaC tooling internals, and CI/CD gates** to
  `22-deployment-iac-cicd.md` (DPL). CUX requires IaC parity of behavior, not
  the IaC engine.
- **Federation/global portal cross-installation aggregation and settlement
  presentation** to `23-federation-global-portal.md` (FED). CUX defines the
  single-installation console; FED reuses its contracts across installations.
- **Agent approval policy, risk-class assignment governance, and autonomous
  operation rules** to `25-agent-governance.md` (AGT). CUX renders agent risk
  classes, approvals, and fail-closed states defined there.
