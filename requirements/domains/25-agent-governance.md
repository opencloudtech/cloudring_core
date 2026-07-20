# 25 — Agent Governance

This domain governs how autonomous and semi-autonomous agents (human-delegated
automation, AI operators, and service accounts) exercise power inside a
CloudRING installation: how every action is risk-classified by its effect, what
evidence each class demands, how approvals are issued, expired, replay-proofed,
and revoked, how secrets are brokered instead of exposed, when agents must stop
and escalate, how emergency authority is contained, and how every action is
journaled so it can be audited and reproduced. It also covers the machine-readable
context agents act upon, the least-privilege automation identities they use, and
the learning loop that converts operational failure into better requirements.
This domain defines the governance contract only; the underlying control-plane,
identity, observability, and backup mechanics live in their own domains.

**Domain contract.** Agents are first-class operators with bounded authority,
never invisible superusers. Authority derives from the *effect* of an action —
money, data, keys, trust, exposure, deletion, migration, settlement — never from
agent identity, and no agent may classify down, self-approve, or self-escalate.
Every mutation is preceded by a plan and the evidence its risk class demands;
every approval is scoped, expiring, fail-closed, and consumed exactly once;
secrets reach agents only as brokered capabilities, never as values in context.
Anything the platform cannot prove is denied, anything ambiguous stops, and every
attempt — success, denial, or stop — lands in an append-only, restart-safe
journal from which the action can be reconstructed. Blocked stays blocked:
governance state is never laundered into readiness.

## Requirements

### CR-AGT-010 — Effect-based risk classification of agent actions
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** If authority is attached to agent identity or role labels, any
  compromised or over-permissioned agent becomes an unbounded operator. Risk
  must be a property of what an action *does*, and the same action can be
  harmless in one scope and destructive in another.
- **Requirement:** Every agent action MUST be classified before execution by
  its effect, never by the identity or self-declaration of the requesting
  agent. The classification MUST cover the effect risk classes money, data,
  keys, trust, exposure, deletion, migration, and settlement, crossed with an
  operational severity ladder (read-only, safe-change, controlled-change,
  risky-change, destructive, emergency). Classification MUST account for
  environment, target scope, and blast radius, so the same operation may land
  in different classes in different contexts. The assigned class MUST be
  visible identically across API, CLI, portal, audit, and agent plan views.
  Agents MUST NOT be able to downgrade a classification; only a governed policy
  change may alter class boundaries.
- **Acceptance evidence:** Conformance test suite that feeds a fixed action
  corpus (including boundary and disguise cases, e.g. a "read" that exports
  data, a "config change" that opens public exposure) through the classifier
  and asserts expected classes; contract check that every executable agent
  action has a declared class; parity test proving the class renders
  identically on all surfaces; negative tests proving agents cannot reclassify
  or bypass classification.
- **Non-goals:** Defining the internal implementation of the classifier engine;
  classifying human-interactive flows (covered by IAM and portal domains);
  per-service business-logic risk rules beyond the platform effect classes.
- **Non-claims:** No classifier implementation exists yet; class-boundary
  completeness against real incident history is unproven; the taxonomy has not
  been validated by an external audit.
- **Stop conditions:** Stop and escalate when an action cannot be classified,
  when classification surfaces disagree, when an action spans multiple effect
  classes and no governing policy resolves precedence, or when any agent
  attempts to influence its own classification.
- **Traceability:** `req-history` (effect-based risk taxonomy, classification
  before execution), `req-acr-singular` (risk classes and approval matrix in
  agent context), `req-acr-plural` (agent overreach risk class). Related:
  CR-AGT-020, CR-AGT-030, CR-AGT-080.

### CR-AGT-020 — Evidence ladder per risk class
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, auditor
- **Problem:** Uniform "approval" for every action either over-burdens trivial
  reads or under-protects destructive change. The proof demanded before an
  action must scale with its classified effect, and each rung must be explicit
  so agents can prepare evidence before asking.
- **Requirement:** The platform MUST define and enforce an evidence ladder:
  read-only actions record goal, scope, resource classes, and redaction status;
  safe-change actions additionally require a diff or preview plus a validation
  and rollback/compensation note; controlled-change actions require a plan,
  impact assessment, named owner, validation, rollback path, and a monitoring
  window; risky-change actions require explicit owner or policy approval;
  destructive actions require confirmation, backup or export proof, an
  irreversibility warning, and a closure audit; emergency actions follow the
  containment path of CR-AGT-070. Higher rungs MUST include all lower-rung
  artifacts. An action MUST NOT execute while any required rung artifact is
  missing, stale, or contradictory.
- **Acceptance evidence:** Ladder contract tests asserting per-rung required
  artifact sets; end-to-end scenario evidence showing one action per rung
  executed with complete artifacts and one per rung denied on a missing
  artifact; audit sampling proof that journaled actions carry the artifacts
  their class demands; chaos/negative tests showing stale artifacts fail
  closed.
- **Non-goals:** Specifying the backup/restore machinery itself (storage
  domain); defining monitoring-window durations per service (service-specific
  policy); mandating human approval for rungs the ladder assigns to automated
  checks.
- **Non-claims:** Ladder rungs are not yet wired to live enforcement; artifact
  formats are undefined; no drill has demonstrated a destructive action gated
  end-to-end by backup/export proof.
- **Stop conditions:** Stop when a required artifact is absent, expired, or
  contradicts the plan; when backup/export proof cannot be produced for a
  destructive class; when validation evidence disagrees with the declared
  intent; when an agent presents evidence outside its redaction authority.
- **Traceability:** `req-history` (per-class evidence escalation ladder),
  `req-acr-singular` (plan/apply/validate/rollback-compensation receipts),
  `req-acr-plural` (no blind mutation, evidence loop). Related: CR-AGT-010,
  CR-AGT-030, CR-AGT-090, CR-AGT-110.

### CR-AGT-030 — Approval lifecycle: scoped, expiring, fail-closed
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Standing or ambient approval turns one granted "yes" into
  permanent authority. Approvals must be bounded objects with a lifecycle, and
  any doubt about validity must resolve to denial.
- **Requirement:** Every approval MUST be an explicit record carrying approver,
  actor, action class, target scope, reason, issue time, expiry or review
  trigger, and a revocation path, with lifecycle states active, expired,
  revoked, superseded, and out-of-scope. Stale, expired, ambiguous, or
  unverifiable approvals MUST fail closed. Waivers MUST be strictly narrower
  than the rule they bypass and carry their own expiry. Approval of an action
  MUST NOT itself grant access to secrets, tenant data, or unrelated evidence.
  Approvals SHOULD be linkable to the requirements, decisions, policies, or
  incidents that motivate them. Approval state MUST be inspectable by the
  requester and auditor at any time.
- **Acceptance evidence:** Lifecycle state-machine tests covering all
  transitions including expiry, revocation, and supersession; fail-closed
  tests for expired, forged, and out-of-scope approvals; negative tests proving
  an action approval does not unlock secret or evidence access; audit-log
  evidence that every executed gated action references an active approval
  record.
- **Non-goals:** Defining who qualifies as an approver per organization
  (authority matrix, CR-AGT-080); building an approval UX (portal domain);
  encoding every organization's delegation policy as platform defaults.
- **Non-claims:** No approval store or lifecycle enforcement exists; expiry
  semantics under clock skew are undesigned; integration with external
  ticketing or change-management systems is unproven.
- **Stop conditions:** Stop when approval state cannot be retrieved or
  verified, when an approval's scope does not cover the requested action and
  target, when expiry or review triggers fire mid-plan, or when a waiver would
  be broader than the bypassed rule.
- **Traceability:** `req-history` (approval lifecycle, stale approval fails
  closed, waivers narrower than rules), `req-acr-singular` (approval records
  before risky action), `req-acr-plural` (approval gate, fail-closed denial).
  Related: CR-AGT-040, CR-AGT-060, CR-AGT-090.

### CR-AGT-040 — Non-self-escalation and consume-once approval tuples
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Two classic failure modes void approval integrity: an agent
  minting or elevating its own authority, and a captured approval being
  replayed to authorize later or broader actions. Approvals must be
  single-use, bound to a specific plan, and never issuable by the requester.
- **Requirement:** Agents MUST NOT approve, escalate, or widen their own
  authority; approver identity MUST be distinct from requester identity for
  risky-change and above. Each granted approval MUST materialize as an
  anti-replay tuple binding approver, requester, action class, target scope, a
  digest of the approved plan, expiry, and a unique nonce. A tuple MUST be
  consumed exactly once at execution time; reuse, alteration of the plan after
  approval, or presentation outside the bound scope MUST be denied. Executed
  tuples MUST be retained in the journal as consumed. Emergency paths MUST NOT
  bypass tuple consumption (see CR-AGT-070 for the retrospective mechanism).
- **Acceptance evidence:** Security test suite demonstrating replay of a
  consumed tuple is rejected, plan mutation after approval invalidates the
  tuple, and self-issued approvals are refused at the policy layer; journal
  contract check that every gated execution records tuple identity and
  consumption; penetration-test evidence class covering forgery and replay.
- **Non-goals:** Specifying cryptographic primitives or token formats
  (implementation concern); replacing enterprise change-advisory-board
  workflows (integration surface only).
- **Non-claims:** Tuple format and consumption protocol are undesigned; replay
  resistance has not been adversarially tested; multi-approver quorum rules
  are unspecified.
- **Stop conditions:** Stop on any replay attempt, any plan-digest mismatch,
  any approval presented outside its bound scope, any detected self-approval
  path, or any inability to durably record tuple consumption (journal
  unavailable).
- **Traceability:** `req-history` (no self-escalation, approval integrity,
  anti-pattern catalog), `req-acr-singular` (no risky action without approval
  record), assignment scope (anti-replay tuples consumed once). Related:
  CR-AGT-030, CR-AGT-070, CR-AGT-090.

### CR-AGT-050 — Brokered secrets for agents, never in context
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, service-team, auditor
- **Problem:** Agents assemble broad context by design; if raw secrets can
  enter prompts, plans, logs, diagnostics, or evidence, every context channel
  becomes a leak multiplier that no perimeter can recall.
- **Requirement:** Agents MUST NEVER receive secret values in context, prompts,
  plans, diagnostics, logs, evidence, or journal records. Secret use MUST go
  through a broker issuing scoped, short-lived capabilities bound to action,
  target, and expiry; agents see metadata and redacted references only.
  Secret-like material discovered in any agent-visible artifact MUST trigger
  redaction and a stop, and its presence MUST be journaled as a security
  event. Rotation MUST be supported without plaintext round-trips, and broker
  issuance, use, and revocation MUST be auditable. Plaintext reveal, where a
  product surface requires it at all, is a human-only, separately gated flow
  outside agent authority.
- **Acceptance evidence:** Secret-scan gate over agent context, plans, logs,
  evidence stores, and journal fixtures proving zero plaintext secret values;
  broker contract tests for scoping, expiry, revocation, and audit events;
  redaction-trigger drill where planted canary secrets in context cause stop +
  security event; negative tests proving diagnostics and error paths cannot be
  used to exfiltrate secrets.
- **Non-goals:** Implementing the secrets manager or KMS itself (IAM/security
  domain); tenant-managed external secret stores beyond the brokering
  contract; human break-glass reveal UX.
- **Non-claims:** No broker implementation exists; coverage of all agent
  context channels (tool outputs, memory, retrieved documents) is unproven;
  canary-drill cadence is undefined.
- **Stop conditions:** Stop on any secret-like value detected in agent
  context or output, on broker unavailability when a secret capability is
  required, on any request for plaintext by an agent, or on detection that a
  brokered capability outlived its bound action.
- **Traceability:** `req-history` (brokered secret capabilities, redaction
  before memory), `req-acr-singular` (secrets as references/capabilities,
  redaction triggers), `req-acr-plural` (agent gets secret metadata only).
  Related: CR-AGT-060, CR-AGT-110, CR-AGT-150.

### CR-AGT-060 — Mandatory stop conditions and escalation
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Guardrails that warn but proceed are decorations. The platform
  must name, once and normatively, the conditions under which an agent halts
  and escalates, and a stop must be terminal for that action — never
  downgradable to a warning.
- **Requirement:** The agent runtime MUST embed these ten stop conditions and
  halt the action, record the stop in the journal, and escalate to the owning
  human or policy authority when any fires: (1) stale, expired, or unprovable
  approval; (2) secret material requested or present outside the broker; (3)
  tenant data access without purpose, policy basis, and redaction; (4)
  commercial state change (billing, settlement, credit, invoice) without its
  required approval class; (5) jurisdiction, residency, key-custody, or
  provider-chain change without policy-aware approval; (6) trust,
  certification, suspension, revocation, or federation change without
  governance approval; (7) validation results contradicting the plan; (8)
  conflict with a requirement, decision record, policy, or conformance gate;
  (9) missing rollback or compensation path for a mutating action; (10) failed
  source-safety, tenant-data, or secret checks. A stop MUST NOT be
  auto-retried into success; resolution requires the named gap to be closed
  and re-approved where applicable. Denials and stops MUST name what is
  missing.
- **Acceptance evidence:** One automated drill per stop condition proving the
  halt, the journal record, the escalation, and the named-gap message;
  regression suite proving stops cannot be suppressed by configuration outside
  governed policy; audit evidence that stopped actions never appear as
  succeeded or warned-away.
- **Non-goals:** Defining per-organization escalation routing beyond the
  contract (authority matrix); guaranteeing escalation delivery channels
  (notification surfaces are a separate domain).
- **Non-claims:** The ten conditions are not yet enforced in any runtime;
  escalation routing is undesigned; drill cadence and ownership are undefined.
- **Stop conditions:** This requirement IS the stop-condition contract; any
  attempt to weaken, bypass, or silence a stop condition is itself a trust
  change requiring governance approval (condition 6).
- **Traceability:** `req-history` (execution gates and the ten stop
  conditions), `req-acr-singular` (mandatory stopConditions on risk classes),
  `req-acr-plural` (fail-closed everywhere, denials name what is missing).
  Related: CR-AGT-010, CR-AGT-030, CR-AGT-160.

### CR-AGT-070 — Emergency containment authority
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Incidents demand speed that normal approval latency cannot
  give, yet "emergency mode" historically becomes blanket permission. Agents
  need bounded emergency power that prefers reversible containment and always
  settles its debt to governance afterwards.
- **Requirement:** Emergency agent authority MUST be limited to a predefined
  catalog of containment scenarios (isolate, disable, revoke, quarantine,
  freeze), each with declared scope and blast radius. Containment MUST be
  preferred over irreversible mutation; destructive remediation remains
  gated. Every emergency action MUST be journaled immediately with scenario
  identity and justification, MUST obtain retrospective human or policy
  approval within a defined deadline, and its authority MUST auto-expire at
  incident closure or deadline, whichever comes first. Emergency authority
  MUST NOT be usable as a general permission path, and repeated emergency use
  MUST feed the learning loop (CR-AGT-170).
- **Acceptance evidence:** Scenario catalog contract tests; live drill per
  containment scenario proving execution, immediate journal record, expiry,
  and retrospective approval capture; negative tests proving non-catalog
  "emergency" actions are refused; audit evidence that overdue retrospective
  approvals block further emergency authority.
- **Non-goals:** Defining incident management and paging workflows (ops
  domain); guaranteeing containment success for arbitrary failure modes;
  customer-facing incident communication.
- **Non-claims:** No containment catalog exists; retrospective-approval
  deadline values are unchosen; drill evidence is absent; behavior under
  partial control-plane outage is undesigned.
- **Stop conditions:** Stop when an action is outside the containment catalog,
  when containment would itself require deletion, migration, or settlement
  effects, when journaling is unavailable, or when a prior emergency's
  retrospective approval is overdue.
- **Traceability:** `req-history` (emergency governance: predefined scenarios,
  retrospective approval, containment over irreversibility, auto-expiry),
  `req-acr-singular` (scenario stop conditions), `req-acr-plural` (emergency
  policy anti-patterns). Related: CR-AGT-040, CR-AGT-060, CR-AGT-170.

### CR-AGT-080 — Actor authority matrix (human, agent, service)
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, service-team, auditor
- **Problem:** "Who may do what by default" cannot be tribal knowledge.
  Humans, agents of different duties, and service/runtime identities need
  separate, explicit default authority, or delegation happens ad hoc and
  invisibly.
- **Requirement:** The platform MUST publish a versioned authority matrix
  covering the three actor kinds — human, agent, and service/runtime — with
  distinct default authority per agent duty class (at minimum: personal,
  service-owner, platform-operator, provider, governance, support, billing,
  and certification agents). Authority MUST default to deny outside the
  matrix, MUST be scoped to owned resources unless an explicit, journaled
  delegation record says otherwise, and cross-party actions MUST declare the
  affected-participant views and notification duties. Support and diagnostic
  agents MUST default to redacted evidence. Every action MUST resolve to one
  matrix row before admission; unknown actor classes fail closed.
- **Acceptance evidence:** Matrix-as-data contract tests (every actor class
  and duty resolves to a default; no orphan classes); admission tests proving
  deny-by-default for unlisted combinations; delegation-record lifecycle
  tests; drill evidence that cross-party actions emit the declared
  notifications; version-pinning check that plans reference the matrix
  version they were authorized under.
- **Non-goals:** Replacing the IAM/RBAC enforcement layer (identity domain);
  encoding each provider's org chart; per-tenant custom matrices beyond the
  delegation mechanism.
- **Non-claims:** The matrix content is not yet authored; duty-class coverage
  for federation and marketplace actors is unproven; notification-duty wiring
  is undesigned.
- **Stop conditions:** Stop when an actor class cannot be resolved to a
  matrix row, when a delegation record is missing, expired, or broader than
  the delegator's own authority, when cross-party notification duties cannot
  be fulfilled, or when the matrix version under which a plan was approved
  has been superseded mid-execution.
- **Traceability:** `req-history` (actor authority matrix, delegation,
  cross-party notification, redacted support defaults), `req-acr-singular`
  (actor identity, scope, purpose per action), `req-acr-plural` (agent
  identity domain, least privilege). Related: CR-AGT-010, CR-AGT-030,
  CR-AGT-150, CR-AGT-190.

### CR-AGT-090 — Append-only, restart-safe operations journal
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Governance without durable memory is theater: if action
  records can be lost on restart, rewritten after the fact, or bypassed when
  the store is down, neither audit nor reproducibility is possible.
- **Requirement:** Every agent action — intent, actor identity and delegating
  principal, scope, target, risk class, plan reference, approval tuple,
  decision, result, evidence references, rollback/compensation note, and stop
  events — MUST be recorded in an append-only journal with canonical UTC
  timestamps. The journal MUST be restart-safe (no in-memory-only production
  state), tamper-evident, and redaction-safe (references and hashes, never
  secret or raw tenant payloads). Journal write failure MUST fail the action
  closed: no mutation may proceed unjournaled. Blocked and stopped actions
  MUST be journaled as blocked or stopped, and journal records MUST NOT be
  rewritten, including after disputes, refunds, or incident reviews.
- **Acceptance evidence:** Persistence tests proving records survive process
  and node restart; tamper-evidence verification tests (chain/digest
  validation); fail-closed drill where journal unavailability blocks a
  mutating action; redaction scan over journal fixtures; audit sampling that
  journaled states match outcome vocabulary (success / blocked / failed /
  stopped) with no silent gaps.
- **Non-goals:** Replacing the observability log pipeline (observability
  domain); defining long-term archive and retention economics; customer-facing
  audit export formats (portal/reporting surfaces).
- **Non-claims:** No journal implementation or schema exists; tamper-evidence
  mechanism is unchosen; performance under write-heavy agent fleets is
  unmeasured.
- **Stop conditions:** Stop any action when the journal is unavailable or
  reject new writes, when a record would require storing secret or raw tenant
  data, or when tamper-evidence verification fails (escalate as a trust event).
- **Traceability:** `req-history` (structured plan/result/audit artifacts,
  erased-audit anti-pattern, UTC canon), `req-acr-singular` (action records
  with intent, actor, risk class, evidence), `req-acr-plural` (action audit:
  intent, inputs, actor, decision, result). Related: CR-AGT-100, CR-AGT-110,
  CR-AGT-040.

### CR-AGT-100 — Action reproducibility from the journal
- **Priority:** P1
- **Status:** proposed
- **Actors:** auditor, operator, agent, provider
- **Problem:** An audit that cannot answer "what exactly did the agent see and
  do, and would it decide the same again" forces trust in the agent's private
  memory. Journal records must be sufficient to reconstruct and re-verify
  significant actions.
- **Requirement:** Journal entries for controlled-change and above MUST
  contain, or reference durably, everything needed to reproduce the decision:
  inputs, plan, context snapshot reference and freshness, policy and authority
  matrix versions, approval tuple, and result. Auditors MUST be able to replay
  an action in an isolated verification mode that performs no mutation, and to
  compare the replayed decision against the recorded one. Reproduction MUST
  NOT require the original agent's volatile memory. Replay output MUST itself
  be journaled as a review event.
- **Acceptance evidence:** Replay test suite reconstructing a sample of
  historical actions and asserting decision equivalence or explaining drift;
  isolation proof that replay performs zero mutation; auditor walkthrough
  evidence tracing one destructive action end-to-end from journal alone;
  contract check that required reproduction fields are non-empty per class.
- **Non-goals:** Deterministic replay of decisions made by non-deterministic
  models (drift explanation is the requirement, not bit-identity); replaying
  human judgment calls; long-horizon re-simulation of whole environments.
- **Non-claims:** Reproduction field set is undefined; isolated replay
  environment does not exist; handling of non-deterministic agent decisions is
  an open design question and honestly bounded.
- **Stop conditions:** Stop when required reproduction fields are missing from
  a record (treat as journal integrity failure), when replay would need
  secret or raw tenant data not present as references, or when the recorded
  policy/matrix versions can no longer be resolved.
- **Traceability:** `req-history` (traceable/explainable key actions),
  `req-acr-singular` (machine-readable evidence suitable for review/replay),
  `req-acr-plural` (agent context reviewable by another agent). Related:
  CR-AGT-090, CR-AGT-110, CR-AGT-130.

### CR-AGT-110 — Phase-separated execution: plan, apply, validate, compensate
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, auditor
- **Problem:** Mutation that is planned and executed in one undifferentiated
  step cannot be gated, previewed, or rolled back. Preflight that itself
  mutates is a recurring, catastrophic bug class in operations tooling.
- **Requirement:** Agent-executed operations MUST be structured as separate,
  evidence-producing phases: plan (preview/diff with blast radius), apply
  (only the approved scope), validate (post-action checks against expected
  evidence), and rollback/compensate (declared path executed or explicitly
  waived by the required approval class). Preflight and plan phases MUST NEVER
  mutate. Retries MUST be idempotent or explicitly forbidden and journaled.
  Every task MUST close in exactly one terminal state — success, blocked, or
  failed — with no gray zone, and partial application MUST be a first-class
  journaled state with a recovery path, never rendered as success.
- **Acceptance evidence:** Static and runtime checks proving plan/preflight
  code paths perform no mutation; scenario tests walking each phase with
  per-phase artifacts; idempotency and replay tests on retried operations;
  forced-partial-failure drills proving partial states journal correctly and
  recovery runs to a terminal state.
- **Non-goals:** Mandating a specific workflow engine; requiring rollback
  automation for change classes where compensation is manual by design (must
  still be declared); defining per-service validation checks.
- **Non-claims:** Phase separation is not yet enforced on any executor;
  idempotency coverage of the action surface is unknown; partial-state
  taxonomy is undrafted.
- **Stop conditions:** Stop when a plan cannot be produced without mutation,
  when validation contradicts expected evidence, when rollback/compensation is
  missing for a mutating action, when a retry would not be idempotent, or when
  an executor cannot represent partial application honestly.
- **Traceability:** `req-acr-singular` (plan/apply/validate/rollback-compensation
  as separate phases, preflight never mutates, no gray zone), `req-history`
  (plan/preview mode, terminal/partial states), `req-acr-plural` (safe change
  lifecycle, idempotent retries). Related: CR-AGT-020, CR-AGT-060, CR-AGT-090.

### CR-AGT-120 — Fail-closed denials and honest terminal states
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, tenant, auditor
- **Problem:** Denials that say "no" without saying why teach agents to
  guess, and states that blur "unverified" into "fine" manufacture false
  readiness. Governance output must be precise enough to act on and honest
  enough to trust.
- **Requirement:** Every denial MUST name the missing element — rights,
  context, evidence, approval, or confirmation — and SHOULD indicate the safe
  next step; a denial MUST NOT suggest bypassing policy. The governance
  vocabulary MUST distinguish ready, degraded, blocked, unknown, and stale,
  and "no evidence found" MUST NEVER be rendered as "ready". Blocked is a
  first-class terminal state that persists as blocked with reason and owner;
  it MUST NOT be downgraded to a warning or converted into a readiness claim
  by any automated path. Uncertainty MUST be recordable by agents as a
  first-class outcome.
- **Acceptance evidence:** Denial-message contract tests asserting the named
  gap and safe-next-step fields; state-vocabulary conformance tests across
  surfaces; negative tests proving blocked states cannot be re-labeled by
  config, retry, or report aggregation; journal audits showing uncertainty
  outcomes recorded as such.
- **Non-goals:** User-facing copywriting and localization of denial texts;
  tenant-notification policy for provider-caused blocks (ops/support domain).
- **Non-claims:** Denial schema is undefined; vocabulary parity across all
  surfaces is unproven; no report generator yet consumes the vocabulary.
- **Stop conditions:** Stop when a denial cannot be expressed with its named
  gap (indicates a governance blind spot), when any path attempts to relabel
  blocked or unknown as ready, or when the state vocabulary and journal
  disagree.
- **Traceability:** `req-acr-plural` (denials name what is missing, blocked is
  first-class, uncertainty recording), `req-acr-singular` (blocked persists as
  blocked, false-readiness signals), `req-history` (blocked evidence never
  converted to release claims). Related: CR-AGT-060, CR-AGT-090, CR-AGT-130.

### CR-AGT-130 — Agent runtime context export contract
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, tenant, auditor
- **Problem:** Agents act on what they can read; if context is ad hoc,
  silently partial, or stale, wrong actions follow at machine speed. Context
  must be a governed product surface with explicit freshness and boundaries.
- **Requirement:** The platform MUST provide machine-readable context export at
  defined levels — platform, region/jurisdiction, organization, tenant,
  service, workload, data set, connector package, and incident/change —
  carrying meaning, not just inventory: owner, lifecycle state, risk class,
  last change, readiness evidence, active alerts, quotas, policies, known
  limitations, and declared next-safe-actions. Every context response MUST
  carry timestamp, source, freshness, and completeness metadata, MUST state
  which parts are withheld due to actor rights, and MUST be least-privilege:
  no secrets, no tenant data beyond the actor's basis, no non-public incident
  detail. Stale context MUST NOT ground a mutating action; freshness checks
  MUST gate mutation admission.
- **Acceptance evidence:** Context-schema contract tests per level;
  freshness-gate tests proving stale context blocks mutation; redaction and
  rights-scoping tests (including "withheld parts" declarations); leakage
  scans proving no secret or out-of-basis tenant data in exports; parity tests
  that context agrees with journal and authority views.
- **Non-goals:** Building the underlying inventory/topology systems
  (foundation domain); natural-language summarization quality; defining
  per-tenant data-access policy (IAM/data domains).
- **Non-claims:** No context export exists; freshness thresholds per level are
  unchosen; completeness accounting (what "full" context means) is an open
  modeling problem.
- **Stop conditions:** Stop mutation when context freshness cannot be
  verified, when completeness metadata indicates gaps material to the planned
  action, when context assembly would require data beyond the actor's basis,
  or when context sources disagree and no precedence rule resolves them.
- **Traceability:** `req-acr-plural` (runtime context at nine levels with
  freshness/completeness, least privilege), `req-acr-singular` (agent context:
  goal, role, authority, boundaries, evidence targets), `req-history`
  (freshness discipline, stale cannot ground mutation). Related: CR-AGT-010,
  CR-AGT-110, CR-AGT-140.

### CR-AGT-140 — Documentation as machine-readable agent context
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, service-team, auditor
- **Problem:** Documentation written only for humans rots silently, and agents
  then execute yesterday's procedures against today's platform. Docs must be
  loadable, verifiable runtime context whose drift from reality is detected
  and marked.
- **Requirement:** The platform MUST maintain canonical, machine-readable
  agent documentation: a stable discovery index, task-oriented docs linked to
  the real API/CLI/schema surfaces, explicit examples of allowed and
  forbidden actions with their risk classes, and explicit staleness and
  deprecation marking. Command and capability references MUST be generated or
  validated against the actual command surface; drift between docs and
  metadata MUST be surfaced as marked drift, never silently. Docs MUST state
  what an agent may conclude from them and what requires live verification.
  Generated documentation MUST NOT be treated as the source of truth.
- **Acceptance evidence:** Docs-drift CI gate comparing documented commands,
  parameters, and capabilities against the live surface; index completeness
  checks; fixtures proving allowed/forbidden examples carry risk classes;
  staleness-marker rendering tests; evidence that agents query the index
  before acting in scenario tests.
- **Non-goals:** Authoring the full human documentation set (documentation
  practices live with each domain); mandating one documentation toolchain;
  natural-language quality thresholds.
- **Non-claims:** No discovery index or drift gate exists; coverage of the
  command surface by validated docs is zero today; allowed/forbidden example
  corpus is unauthored.
- **Stop conditions:** Stop when an agent's planned action relies on docs
  marked stale or drifting, when the discovery index and live surface
  contradict on the action's command or capability, or when required
  context docs are absent for the target capability (treat as unknown, fail
  closed for mutation).
- **Traceability:** `req-acr-plural` (documentation as runtime context,
  discovery indexes, staleness marking), `req-history` (command/docs drift
  trap, generated docs never source of truth, docs-as-contract). Related:
  CR-AGT-130, CR-AGT-110, CR-AGT-170.

### CR-AGT-150 — Automation accounts with least privilege
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, auditor
- **Problem:** Agents acting on shared human credentials destroy attribution,
  defeat revocation, and silently inherit broad rights. Automation needs its
  own identities whose default is little or nothing.
- **Requirement:** Every agent or automation workload MUST run under a
  distinct automation identity, never a shared human credential, and SHOULD be
  scoped per duty and per environment. Automation credentials MUST be
  least-privilege, scoped, short-lived, and rotatable, and MUST start
  read-only; write authority requires an explicit, journaled grant tied to the
  authority matrix. Revocation MUST invalidate active sessions and cached
  tokens within a declared bound. Issuance, grant, rotation, and revocation of
  automation credentials MUST be journaled as trust events. Automation
  identities MUST NOT be usable to mint further identities or broaden their
  own grants.
- **Acceptance evidence:** Identity-lifecycle tests (issue, grant, rotate,
  revoke) with journal verification; revocation-latency measurements against
  the declared bound; negative tests proving automation identities cannot
  self-mint or self-grant; audit evidence that no executed action is
  attributable to a shared human credential; default-posture check that new
  automation identities are read-only.
- **Non-goals:** Implementing the underlying identity provider or token
  machinery (IAM domain); workforce identity lifecycle; tenant-managed
  automation identities beyond the contract.
- **Non-claims:** No automation-identity issuance flow exists; revocation
  bounds are unmeasured; coverage across all agent execution environments
  (local, CI, in-cluster) is unproven.
- **Stop conditions:** Stop when an action presents a shared or human
  credential for automation, when an automation identity requests a grant
  broader than its duty class permits, when revocation cannot be confirmed to
  have propagated, or when credential issuance cannot be journaled.
- **Traceability:** `req-acr-plural` (AI-agent identity as separate domain,
  agent explains on whose behalf it acts), `req-history` (maintenance vs
  runtime role separation, identity-as-authority anti-pattern),
  `req-acr-singular` (actor identity and scope per action). Related:
  CR-AGT-080, CR-AGT-190, CR-AGT-040.

### CR-AGT-160 — Money, data, keys, trust, and jurisdiction policy gates
- **Priority:** P0
- **Status:** proposed
- **Actors:** agent, operator, provider, tenant, auditor
- **Problem:** Some effect classes cannot be governed by generic approval
  alone: commercial state, tenant data, jurisdiction, and platform trust each
  demand a named, higher gate, or a single agent slip becomes a financial,
  legal, or trust catastrophe.
- **Requirement:** The governance layer MUST enforce dedicated gates: billing,
  settlement, credit, and invoice changes are classified risky-change or
  stronger and require the commercial approval class; tenant data access
  requires purpose, policy basis, redaction, and party-scoped visibility;
  jurisdiction, residency, key-custody, or provider-chain changes require
  policy-aware approval with a jurisdiction impact preview; trust,
  certification, suspension, revocation, and federation-membership changes
  require governance approval. Key custody changes MUST additionally satisfy
  the brokered-secrets contract (CR-AGT-050). Each gate MUST journal the gate
  decision with its basis, and gate outcomes MUST be inspectable by the
  affected party within their rights.
- **Acceptance evidence:** Gate matrix contract tests binding each effect
  class to its approval class; scenario drills per gate showing approval
  required, denied-without-basis, and journaled outcomes; jurisdiction
  preview tests proving data-movement impact is shown before approval;
  cross-tenant visibility tests for party-scoped data access.
- **Non-goals:** Defining billing/settlement internals (billing domain) or
  federation governance processes (federation domain); legal advice encoded in
  policy; per-jurisdiction compliance content.
- **Non-claims:** Gate definitions are unauthored; integration with actual
  billing, key-custody, and federation systems is absent; jurisdiction
  modeling (region vs controller vs backup vs log processing location) is
  undrafted.
- **Stop conditions:** Stop on any commercial state change without its
  approval class, on tenant data access lacking purpose/basis/redaction, on
  jurisdiction-impacting change without preview and policy approval, on trust
  or federation change without governance approval, or on any gate-decision
  journaling failure.
- **Traceability:** `req-history` (money/policy/trust gates, tenant data
  purpose+redaction, jurisdiction approval), `req-acr-singular` (risk matrix
  gates for money, data, keys, trust, public exposure), `req-acr-plural`
  (jurisdiction impact visible before dangerous action). Related: CR-AGT-010,
  CR-AGT-060, CR-AGT-050.

### CR-AGT-170 — Learning loop: from postmortem to requirement
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, agent, provider, auditor
- **Problem:** Platforms that do not metabolize failure repeat it. Incidents,
  stopped actions, and repeated manual overrides are governance signals that
  must land as changes to requirements, policy, runbooks, or checks — with
  proof the loop closed.
- **Requirement:** Every closed incident and every emergency or stop-condition
  cluster MUST produce a learning record resolving to exactly one outcome: a
  new or changed requirement, a decision record, a runbook update, a
  conformance check, or an explicit no-change rationale. Raw incident text and
  private detail MUST NOT be retained in the learning record — only
  source-safe, redacted lessons. Repeated manual approvals of the same action
  pattern MUST trigger review of whether policy should grant a narrower
  standing rule; repeated emergency or validation failures MUST create
  follow-up work with an owner. Learning records MUST link to the journal
  events that motivated them, and loop closure MUST itself be journaled.
- **Acceptance evidence:** Learning-record schema validation; sampled
  traceability from incident journal events to closed learning records with
  one of the five outcomes; redaction scan over learning records; metrics or
  audit evidence showing repeated-approval and repeated-failure patterns
  actually triggered reviews; no-change rationales spot-checked for honesty.
- **Non-goals:** Mandating a specific postmortem format or meeting ritual;
  automating requirement authorship (proposals may be drafted, acceptance
  stays human/owner); organization-specific SLA on loop closure beyond having
  one.
- **Non-claims:** No learning-record pipeline exists; pattern-detection for
  repeated approvals/failures is undesigned; loop-closure quality is
  unmeasurable today.
- **Stop conditions:** Stop (escalate) when an incident closes without a
  learning record, when a learning record would require retaining raw incident
  text or private detail, or when the same stop condition fires repeatedly
  without any follow-up owner.
- **Traceability:** `req-acr-singular` (incident closure creates learning
  record, no raw incident text), `req-history` (approval-matrix learning loop,
  operational churn updates docs or rationale), `req-acr-plural` (evidence-loop
  operations). Related: CR-AGT-060, CR-AGT-070, CR-AGT-090, CR-AGT-200.

### CR-AGT-180 — Shared-workspace and concurrency safety
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, service-team
- **Problem:** Multiple agents and humans share working state — repositories,
  worktrees, pipelines, infrastructure state. An agent that resets, force-
  overwrites, or blindly proceeds on someone else's uncommitted work destroys
  other actors' data at machine speed.
- **Requirement:** Agents MUST detect shared or dirty workspace state before
  acting, MUST NEVER reset, discard, or overwrite changes they did not make,
  and MUST stop on shared-file or shared-state conflicts rather than resolve
  them unilaterally. Where coordination is required, agents MUST use declared
  locking or leasing surfaces rather than convention. Workspace-affecting
  actions MUST record the pre-action state fingerprint in the journal, and
  concurrent-action conflicts MUST be reported with the conflicting parties
  identified within rights boundaries.
- **Acceptance evidence:** Concurrency scenario tests (two agents, overlapping
  targets) proving conflict detection, stop, and named-party reporting; dirty-
  state detection tests; negative tests proving reset/discard of foreign
  changes is refused; journal contract check for pre-action state
  fingerprints.
- **Non-goals:** Building a general distributed-locking service; merge-
  conflict resolution assistance (a separate, non-governance capability);
  human pair-working etiquette.
- **Non-claims:** No shared-state detection exists; lock/lease surfaces are
  undesigned; behavior across heterogeneous workspaces (local, CI, cluster) is
  unproven.
- **Stop conditions:** Stop on detection of foreign uncommitted changes in the
  action's footprint, on lock/lease contention that cannot be attributed, on
  state fingerprints that changed between plan and apply, or on any request to
  discard or overwrite others' work (escalate as data risk).
- **Traceability:** `req-acr-singular` (respect shared dirty worktrees, stop
  on shared-file conflict), `req-history` (conflict preflight, dirty-root
  evidence downgrades), `req-acr-plural` (safe agent change scenario stops).
  Related: CR-AGT-060, CR-AGT-110.

### CR-AGT-190 — Agent identity attribution and explainability
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, tenant, auditor
- **Problem:** If an agent cannot state on whose behalf it acts, with which
  rights, and under which approval, then accountability is fictional and every
  downstream audit, dispute, or revocation decision is guesswork.
- **Requirement:** Every agent action MUST carry an attribution triple —
  agent identity, delegating principal (human, organization, or service), and
  approval or authority reference — journaled with the action. An agent MUST
  be able to answer, through a structured interface, on whose behalf it acts,
  which authority class backs each planned action, which boundaries apply, and
  which rights it does NOT have. Role or identity mismatch (a valid identity
  used for a duty outside its class) MUST fail closed. Attribution MUST remain
  intact through delegation chains, with each hop journaled.
- **Acceptance evidence:** Attribution schema tests on journal records;
  explainability-interface contract tests (behalf-of, authority, boundaries,
  non-rights queries return structured answers); fail-closed tests for
  role-mismatched identities; delegation-chain traversal tests proving
  per-hop attribution; tenant-visible attribution within rights boundaries.
- **Non-goals:** Natural-language explanation quality; exposing internal
  delegation chains beyond each party's rights; defining org-level delegation
  policy (authority matrix).
- **Non-claims:** No explainability interface exists; attribution through
  multi-hop delegation is undesigned; tenant-visible attribution surfaces are
  undefined.
- **Stop conditions:** Stop when the delegating principal cannot be resolved,
  when an identity is presented outside its duty class, when a delegation hop
  lacks a journaled record, or when the agent cannot produce its authority
  basis for a planned action.
- **Traceability:** `req-acr-plural` (agent explains on whose behalf it acts
  and with what rights, role mismatch fails closed), `req-acr-singular`
  (permission decisions explain actor/subject/action/target/scope),
  `req-history` (separate default authority per actor, cross-party views).
  Related: CR-AGT-080, CR-AGT-150, CR-AGT-090.

### CR-AGT-200 — Governance matrix change control
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider, auditor, agent
- **Problem:** The risk classes, evidence ladder, stop conditions, and
  authority matrix are themselves security-critical configuration. Quiet edits
  to the rules of governance are as dangerous as breaking them.
- **Requirement:** The governance rule set (risk-class definitions, evidence
  ladder, stop conditions, authority matrix, containment catalog) MUST be
  versioned, reviewed, and journaled like code, with changes classified as
  backward-compatible or explicitly breaking. Every plan and approval MUST
  pin the governance version it was evaluated under, and breaking changes
  MUST invalidate outstanding approvals issued under superseded versions.
  Governance changes MUST themselves pass the trust-change gate (CR-AGT-160).
  Differences between versions MUST be machine-inspectable so auditors can
  answer "under which rules was this action admitted".
- **Acceptance evidence:** Version-pinning contract tests on plans and
  approvals; breaking-change drills proving outstanding approvals are
  invalidated; machine-diff evidence between governance versions; trust-gate
  workflow evidence for a governance change itself; audit query evidence
  reconstructing the governing version of historical actions.
- **Non-goals:** Freezing governance evolution (change is expected; silent
  change is forbidden); defining the human review board composition;
  per-tenant governance forks beyond declared policy.
- **Non-claims:** No versioned governance store exists; compatibility
  classification rules are undefined; migration behavior for in-flight actions
  during a governance upgrade is undesigned.
- **Stop conditions:** Stop when a governance change bypasses the trust gate,
  when a plan's pinned governance version is unresolvable, when a breaking
  change lands while unexpired approvals depend on the superseded version, or
  when version history has gaps (escalate as a trust event).
- **Traceability:** `req-history` (matrix changes versioned, reviewed,
  backward-compatible or explicitly breaking), `req-acr-singular` (registry as
  source of truth, drift as defect), `req-acr-plural` (policy versioning and
  freshness). Related: CR-AGT-010, CR-AGT-030, CR-AGT-170.

### CR-AGT-210 — Governance enforcement drills
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, auditor, provider, agent
- **Problem:** Governance that is never adversarially exercised decays into
  assumed governance. The platform must routinely prove — not merely assert —
  that stops fire, approvals expire, replays are rejected, and secrets stay
  out of context.
- **Requirement:** The platform SHOULD run scheduled and on-demand governance
  drills covering: stop-condition firing (all ten), approval expiry and
  revocation, replay of consumed approval tuples, secret-canary injection into
  context, classification boundary cases, emergency authority expiry, and
  journal unavailability fail-closed behavior. Drill results MUST be produced
  as dated, rerunnable evidence with pass/fail/blocked outcomes, and drill
  failures MUST be treated as governance regressions that block any claim of
  governed agent operation until resolved. Drill scenarios and results MUST be
  journaled and auditor-visible.
- **Acceptance evidence:** A drill suite runnable on demand with per-scenario
  evidence artifacts; dated drill reports with explicit outcomes; regression
  workflow proof that a failed drill blocks governed-operation claims;
  auditor sampling of drill journals; historical drill trend data.
- **Non-goals:** Continuous chaos engineering of production tenant workloads
  (drills target governance machinery, in scoped environments); mandating
  drill frequency beyond a declared, enforced minimum; certifying security
  beyond the drill scope.
- **Non-claims:** No drill suite exists; minimum drill cadence is undeclared;
  drill coverage of the full stop-condition list is unproven; evidence format
  is undefined.
- **Stop conditions:** Stop (and block governed-operation claims) when a
  mandatory drill fails or is overdue, when a drill cannot run without
  touching production tenant data, or when drill evidence cannot be journaled.
- **Traceability:** `req-history` (restore-drill-as-gate pattern applied to
  governance, negative-scenario evidence), `req-acr-singular` (failure
  injection or credible simulation for readiness claims), `req-acr-plural`
  (readiness gates demand live evidence classes). Related: CR-AGT-060,
  CR-AGT-040, CR-AGT-050, CR-AGT-090.

## Coverage notes

This domain deliberately specifies only the governance layer for agent-actor
operations and defers the following:

- **Structured control-plane state and surface parity** (single truth across
  UI/API/CLI/agent surfaces, service lifecycle state models) — platform
  foundation (`10-platform-foundation`) and ops (`21-ops-sre-support`).
- **Authentication, token/JWKS policy, production identity gates, RBAC
  mechanics, and the secrets-manager/KMS implementation** — identity and
  security (`15-iam-identity-security`); this domain consumes those as
  brokered capabilities and authority decisions.
- **Metrics, logs, tracing, alerting, and diagnostics pipelines** —
  observability (`20-observability`); the ops journal here is a governance
  record, not the telemetry system.
- **Incident management, runbooks, paging, and support workflows** — ops/SRE
  (`21-ops-sre-support`); this domain governs only the agent-authority
  aspects of emergency action.
- **Backup, restore, DR mechanics and restore drills** — storage
  (`13-storage-backup-dr`); the evidence ladder references backup proof but
  does not define it.
- **Migration execution mechanics and branching/preview environments for data
  services** — data services (`24-data-services`) and deployment
  (`22-deployment-iac-cicd`); here only the approval and stop semantics.
- **Billing, settlement, marketplace order lifecycle internals** — billing
  (`16-billing-finops`) and marketplace (`18-marketplace-catalog`); this
  domain gates who may change commercial state, not how it is computed.
- **OCS connector validation and service onboarding surfaces** — OCS
  (`17-ocs-service-connectors`).
- **Federation trust, settlement between providers, and cross-provider
  authority** — federation (`23-federation-global-portal`); this domain
  defines only the governance gate such changes must pass.
