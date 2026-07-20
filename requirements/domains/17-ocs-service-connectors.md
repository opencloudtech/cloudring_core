# 17 — OCS Service Connectors

This domain covers the Open Cloud Standard (OCS): the contract layer by which
every service — first-party, partner, or third-party — integrates with the
platform. It includes the connector package model (registration and capability
announcement), the mandatory lifecycle API set with idempotency and rollback
references, typed user-visible states, optional and inter-service dependency
APIs, the billing connector for usage-metrics transmission, the microfrontend
contract, the module registry with versioning, the conformance suite, the SDK
and reference implementation, the service-team onboarding journey, versioning
and compatibility gateways, workload-identity-only secrets, declared durability
surfaces, and the evidence/receipt format.

**Domain contract.** Services integrate exclusively through versioned,
declarative connector packages; the platform core never carries service-specific
wiring. Every lifecycle action is idempotent, and mutating delete/retry/repair
actions carry rollback evidence. Secrets appear in packages only as workload
identity references — raw material is a hard blocker. States are typed,
user-visible, and carry remediation. Nothing reaches a catalog without billing,
tenant-access, support, readiness, and durability surfaces plus a passing
conformance report and a recorded owner review. Evidence is fresh and honest:
blocked, stale, synthetic-only, or redacted evidence never promotes into any
readiness claim. These rules are non-negotiable.

### CR-OCS-010 — Connector package as the sole service-integration unit
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider, operator
- **Problem:** If the platform core hard-wires service-specific UI, billing,
  automation, or lifecycle logic, every new service demands core changes and
  independent service teams can never ship on their own.
- **Requirement:** The platform core MUST consume every service — first-party
  or third-party — exclusively through a versioned, declarative connector
  package describing the service's metadata and surfaces. The core MUST NOT
  contain business logic, UI routes, billing rules, or automation specific to
  any individual service. The connector package MUST be the single publishable
  unit from which catalog, portal, billing, identity, automation, and support
  surfaces are derived.
- **Acceptance evidence:** CI contract checks proving core packages carry no
  service-implementation references, provider endpoints, or hardcoded catalog
  entries; conformance suite run over the example connector packages;
  architecture contract test demonstrating that catalog, portal, and billing
  surfaces render only from package metadata.
- **Non-goals:** Does not prescribe a service's internal implementation
  language, framework, or architecture behind its declared contract.
- **Non-claims:** Verified today for package-metadata consumption in the OSS
  core; not yet proven across surfaces that are not yet implemented
  (marketplace operation, federation).
- **Stop conditions:** trust, exposure — halt and escalate on any
  service-specific wiring, hardcoded catalog entry, implementation reference,
  provider endpoint, or credential discovered inside core artifacts.
- **Traceability:** vision-deck; current-core; legacy-platform-a. Related:
  CR-OCS-020, CR-OCS-090, CR-OCS-160.

### CR-OCS-020 — Registration, capability announcement, and one service identity
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider, operator
- **Problem:** A service that cannot self-register and announce what it offers
  forces manual platform-team onboarding; disconnected per-surface identities
  (catalog versus billing versus support) make entitlement and audit
  inconsistent.
- **Requirement:** Every service MUST register through its connector package,
  announcing user-visible capabilities (class, name, description) without
  exposing implementation detail. Registration MUST assign one stable service
  identity reused across catalog, entitlements, billing, observability,
  automation, and support surfaces. Capability declarations MUST be portable
  and implementation-neutral.
- **Acceptance evidence:** registration contract tests; fixture set
  demonstrating identity reuse across all consuming surfaces; conformance
  check rejecting implementation-specific capability declarations.
- **Non-goals:** Does not define catalog browsing UX (CUX) or marketplace
  economics (MKT).
- **Non-claims:** Multi-surface identity propagation is specified; live
  cross-surface audit correlation for one identity is not yet demonstrated.
- **Stop conditions:** trust — halt registration on identity collision, on
  ownership that cannot be verified, or on capabilities that misrepresent data
  handling; keep the listing unpublished until resolved.
- **Traceability:** vision-deck; current-core; legacy-platform-a. Related:
  CR-OCS-010, CR-OCS-160.

### CR-OCS-030 — Mandatory lifecycle APIs: idempotent, rollback-referenced
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, tenant, operator, agent
- **Problem:** Unsafe lifecycle operations are the top tenant-data risk:
  partially completed provisioning, deletion without a recovery reference, or
  non-idempotent retries corrupt tenant state and billing attribution.
- **Requirement:** Every connector MUST implement the mandatory lifecycle
  action set: provision, backup, restore, export, delete, retry, and rollback.
  Every action MUST be idempotent and accept a caller-supplied idempotency key;
  repeated invocation with the same key MUST return the original result without
  re-executing side effects. Mutating delete, retry, and repair actions MUST
  carry a rollback reference to prior rollback or backup evidence. Long-running
  mutating operations MUST return an operation handle immediately and expose
  progress through the typed state model; read operations MAY be synchronous.
- **Acceptance evidence:** conformance lifecycle-surface check; contract tests
  replaying duplicate idempotency keys against the reference implementation
  proving single-effect execution; negative fixtures missing a rollback
  reference rejected by the validator; operation-log audit trail inspection.
- **Non-goals:** Does not mandate a specific internal state-machine
  implementation, storage technology, or queuing mechanism inside the service.
- **Non-claims:** Idempotency and rollback linkage are proven against the
  synthetic reference controller; behavior under real provider failure
  injection at scale is not yet load-tested.
- **Stop conditions:** data, deletion, migration — halt the operation and
  escalate when a mutating action lacks a rollback reference, when repeated
  execution with one idempotency key produces divergent results, or when delete
  proceeds while export or backup evidence for the affected data classes is
  absent or stale.
- **Traceability:** current-core; legacy-platform-a; req-acr-plural. Related:
  CR-OCS-040, CR-OCS-140, CR-OCS-150.

### CR-OCS-040 — Typed user-visible states with remediation
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent, service-team
- **Problem:** Opaque or service-specific status vocabularies make it
  impossible for UI, API, and agents to render one truthful picture of a
  service or to tell the user what to do next.
- **Requirement:** Every connector MUST declare typed user-visible states
  covering at minimum ready, denied, degraded, blocked, and retryable. Each
  state MUST carry a machine-readable reason, an evidence reference, and a
  remediation naming the next user or operator action. The vocabulary MUST be
  shared identically across UI, API, and agent surfaces; services MUST NOT
  expose divergent public states without mapping them to the typed set.
- **Acceptance evidence:** conformance states-surface check; contract test that
  every declared state carries reason, evidence reference, and remediation;
  rendering test proving UI and API present the same state payload; negative
  fixture lacking remediation rejected.
- **Non-goals:** Does not fix user-facing copy (CUX owns wording); does not
  enumerate a service's internal state machine beyond the public typed set.
- **Non-claims:** The shared vocabulary is enforced at contract level;
  consistent rendering across every portal surface is not yet verified
  end-to-end.
- **Stop conditions:** trust, data — halt and escalate if a state
  misrepresents a data-bearing operation (for example reporting deletion
  complete while data persists) or if a denied or blocked state reaches users
  without remediation.
- **Traceability:** current-core; legacy-platform-a; req-acr-singular.
  Related: CR-OCS-030, CR-OCS-160.

### CR-OCS-050 — Optional APIs and portable inter-service dependencies
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, vendor, provider
- **Problem:** Services need to consume one another (for example a database
  service using a backup service), but hard-coded backend coupling destroys
  portability and independent deployability.
- **Requirement:** Beyond the mandatory lifecycle set, a connector MAY declare
  optional APIs. Dependencies between services MUST be declared as portable
  dependency roles — a capability class plus required portability guarantees
  such as export/import — never as a fixed backend product or implementation.
  Dependency declarations MUST be versioned, and a consumer MUST degrade
  honestly into a typed denied or degraded state when a dependency is
  unavailable rather than failing silently.
- **Acceptance evidence:** conformance dependency-surface check rejecting
  implementation-specific backends; integration drill withdrawing a dependency
  and verifying the consumer surfaces a typed degraded state with remediation;
  contract tests for dependency version negotiation.
- **Non-goals:** Does not provide runtime service discovery or mesh plumbing
  (platform infrastructure concern); does not guarantee any specific dependency
  is present in a given installation.
- **Non-claims:** Portable dependency roles are specified for known classes;
  the role catalog is incomplete for future service types, and cross-service
  failure drills have not yet run.
- **Stop conditions:** data, trust — halt activation of a dependent service
  when its declared dependency cannot be satisfied, or when dependency version
  skew breaks a data path (backup, export) without a tested migration window.
- **Traceability:** vision-deck; current-core; legacy-platform-a. Related:
  CR-OCS-020, CR-OCS-120.

### CR-OCS-060 — Billing connector and usage-metrics transmission
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider, tenant, auditor
- **Problem:** Metering is the stated basis for billing, licensing, and revenue
  sharing; without a first-class usage-transmission contract no service can be
  sold and no provider can be paid.
- **Requirement:** Every billable service MUST declare usage meters with stable
  names, units, and aggregation rules, and MUST transmit usage events through
  its declared billing connector. Events MUST carry an idempotency key,
  entitlement attribution, and a replay policy; replay MUST deduplicate by
  idempotency key. Cost meters MUST reference rate-card evidence before any
  charge may be derived from them. A non-billable service MUST publish an
  explicit non-billable policy rather than omitting the surface. Meter
  declarations MUST cross-link to the declared billing connector and fail
  closed on mismatch.
- **Acceptance evidence:** conformance billing-surface check including
  cross-link validation; contract tests proving duplicate events are
  deduplicated on replay; fixture proving a cost meter without rate-card
  evidence is rejected; non-billable policy fixture validated.
- **Non-goals:** Charging, rating, invoicing, tariff design, and settlement
  internals belong to BIL; payment-provider integration is out of scope.
- **Non-claims:** Meter transmission and replay dedup are contract-verified;
  end-to-end charge accuracy against a running billing pipeline is not yet
  proven, and no settlement semantics are claimed.
- **Stop conditions:** money, settlement, data — halt usage acceptance and
  escalate on missing or duplicated attribution, on meter or unit drift between
  the package and transmitted events, on replay producing double-counting, or
  on any charge derived without rate-card evidence; suspend the affected meters
  until reconciled.
- **Traceability:** vision-deck; current-core; legacy-platform-a. Related:
  CR-OCS-030, CR-OCS-150, CR-OCS-160.

### CR-OCS-070 — Microfrontend contract: integrity, sandbox, guidelines
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, tenant, provider
- **Problem:** Third-party services need UI inside the platform console, but
  loading unvetted third-party code into an authenticated session is a direct
  credential- and data-exposure path.
- **Requirement:** A connector MAY declare portal modules; every declared
  module MUST bind to the microfrontend host contract: host runtime, mount
  reference, version range, integrity reference, sandbox boundary, allowed
  events, and required context — all mandatory. The platform MUST refuse to
  mount a module whose integrity verification fails or whose sandbox
  requirements cannot be enforced, failing closed to a hidden module rather
  than a permissive mount. UI guidelines covering the design system,
  accessibility, and consent boundaries MUST be published and versioned.
- **Acceptance evidence:** conformance portal/UI surface check; contract tests
  proving mount refusal on integrity failure; sandbox test suite demonstrating
  no unsanctioned access to host session, storage, or tokens; published and
  versioned UI guidelines document.
- **Non-goals:** The portal shell implementation and navigation UX belong to
  CUX; the contract does not pick a frontend framework for services beyond the
  host runtime interface.
- **Non-claims:** The integrity and sandbox contract is specified and fails
  closed at the metadata layer; a standing third-party security review program
  for module certification is not yet operational.
- **Stop conditions:** keys, trust, exposure — halt mounting and escalate
  immediately on integrity mismatch, sandbox violation, unexpected access to
  host context (cookies, session storage, tokens), or an unreviewed module
  version appearing in any registry record.
- **Traceability:** current-core; legacy-platform-a; vision-deck. Related:
  CR-OCS-080, CR-OCS-160.

### CR-OCS-080 — Module registry with versioning and policy-aware lifecycle
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, vendor, agent
- **Problem:** Without a governed registry, installing, updating, or removing
  service modules is untracked, irreversible actions lack rollback, and
  dependency conflicts surface as runtime outages.
- **Requirement:** The platform MUST maintain a module registry in which every
  module has a unique stable identifier and a semantic version. Registry
  lifecycle operations — install, update, remove, suspend, deprecate — MUST be
  idempotent, policy-aware, and auditable: each operation records an operation
  identifier, an idempotency key, an audit receipt, an evidence receipt, and a
  rollback hook. Dependency resolution MUST be topological; dependency cycles
  and missing dependencies MUST block the operation. Registry records MUST NOT
  contain service-implementation references, provider endpoints, credentials,
  or mutation commands. Module states MUST include at least installable,
  installed, suspended, deprecated, and not-installed.
- **Acceptance evidence:** registry contract schema with valid and invalid
  fixtures (duplicate identifier, dependency cycle, missing dependency,
  embedded implementation reference) proving fail-closed validation; operation
  receipt test proving reinstall of an installed module is a no-op under the
  same operation identifier; audit trail inspection.
- **Non-goals:** The registry does not execute deployments (DPL owns rollout
  machinery) and does not rank or merchandise modules (MKT).
- **Non-claims:** Registry semantics are contract-verified; registry operation
  at ecosystem scale (many modules across many installations) is not yet
  exercised.
- **Stop conditions:** migration, deletion, trust — halt and escalate on a
  dependency cycle or unsatisfied dependency, on any remove or update lacking a
  rollback hook and fresh pre-mutation evidence, or on registry records
  carrying implementation references, endpoints, or secrets.
- **Traceability:** current-core; legacy-platform-a. Related: CR-OCS-070,
  CR-OCS-120, CR-OCS-150.

### CR-OCS-090 — Conformance suite with machine-readable problems
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator, agent, auditor
- **Problem:** A standard without an executable gate degrades into
  documentation; humans cannot review hundreds of metadata surfaces
  consistently, and agents need structured failure output to act on.
- **Requirement:** The platform MUST ship a conformance suite validating
  connector packages against every mandatory surface and failing closed: any
  missing or invalid surface is a blocking problem. Problems MUST be
  machine-readable with at least surface, field, message, and remediation. The
  suite MUST emit a machine report and an optional evidence receipt, MUST run
  identically in local development and CI, and MUST carry its own non-claims —
  metadata conformance is not live production readiness. The suite MUST be
  versioned together with the standard it enforces.
- **Acceptance evidence:** conformance runs over example and reference packages
  passing; a negative-fixture corpus proving the rejection path of every
  surface with remediation strings; CI gate configuration blocking merge on
  failure; receipt schema validation test.
- **Non-goals:** Conformance does not deploy, mutate, or smoke-test a live
  service; performance and load characteristics are out of scope.
- **Non-claims:** The suite verifies package completeness, not a running
  service; coverage is limited to the declared standard version and lags
  proposed surfaces until they are versioned in.
- **Stop conditions:** trust — halt the publication pipeline and escalate if
  the suite is found to accept an invalid package (verifier false negative), if
  problem output loses its machine-readable shape, or if a conformance report
  is presented as live-readiness evidence.
- **Traceability:** current-core; req-acr-singular; legacy-platform-a.
  Related: CR-OCS-100, CR-OCS-110, CR-OCS-160.

### CR-OCS-100 — SDK and provider-neutral reference implementation
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, agent
- **Problem:** A contract that a team cannot code against within a day will be
  routed around; a reference implementation is the only unambiguous executable
  form of the specification.
- **Requirement:** The platform MUST publish a maintained SDK exposing the
  connector package model, parsers, and validators as its public facade, plus a
  complete provider-neutral reference implementation: a synthetic service with
  a claim resource, a phase machine, receipts, idempotent lifecycle handling,
  billing events, diagnostics, a portal module, and deployment manifests. The
  reference implementation MUST NOT call any real provider API, MUST include
  negative fixtures, and MUST pass the conformance suite it accompanies. SDK
  and reference MUST be covered by the same CI gates as the core.
- **Acceptance evidence:** SDK published with versioned releases; reference
  implementation content checklist complete; conformance pass receipt for the
  reference package; consumer test proving an external module that imports only
  the SDK validates successfully.
- **Non-goals:** The SDK does not scaffold production infrastructure or choose
  a service's datastore; client libraries for consuming individual services are
  not part of this contract.
- **Non-claims:** The reference implementation exercises the contract
  synthetically; it is not evidence that real services behave identically under
  provider failures.
- **Stop conditions:** trust, exposure — halt release of any SDK or reference
  version that drifts from the enforced standard version, or if provider
  endpoints, credentials, or copied implementation text appear in the reference
  tree.
- **Traceability:** current-core; legacy-platform-a; vision-deck. Related:
  CR-OCS-090, CR-OCS-110.

### CR-OCS-110 — Service-team onboarding journey and minimum-useful profile
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, vendor, provider
- **Problem:** Full connector packages are large; if the only path is
  "implement everything", small teams never start and the service ecosystem
  stalls.
- **Requirement:** The platform SHOULD publish a documented onboarding journey
  achievable without platform-team involvement: documentation, skeleton
  generation, local validation, conformance, catalog submission. A
  minimum-useful-connector profile SHOULD be defined — a named subset of
  surfaces sufficient for validation and limited catalog presence — with an
  explicit upgrade path to the full profile. The journey SHOULD include
  tutorials for the main connector types (service module, billing connector,
  portal extension) and a timed onboarding drill.
- **Acceptance evidence:** journey documentation set; skeleton tooling output
  passing validation; minimum-useful profile contract; recorded onboarding
  drill in which a team external to core development completes the path from
  documentation to catalog submission using only public materials.
- **Non-goals:** Does not guarantee marketplace publication (MKT gate); does
  not provide hosted development environments.
- **Non-claims:** The journey is documented but not yet validated end-to-end by
  a fully external service team; the minimum-useful surface subset is proposed,
  not ratified.
- **Stop conditions:** trust — halt catalog submission under the minimum-useful
  profile if it is found to permit surfaces touching money, deletion, or
  secrets without their full mandatory checks; escalate any workaround that
  bypasses conformance.
- **Traceability:** current-core; legacy-platform-a; vision-deck. Related:
  CR-OCS-090, CR-OCS-100, CR-OCS-160.

### CR-OCS-120 — Versioning, deprecation, and compatibility gateways
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, vendor, provider, operator
- **Problem:** The standard evolves; without explicit versioning and
  deprecation windows, upgrades either silently break installed services or
  freeze the standard forever.
- **Requirement:** All contract schemas and package formats MUST be versioned.
  Breaking changes MUST ship as a new version with a published deprecation
  policy: support window, migration guide, and a gateway period during which
  both versions validate. Compatibility between module versions and platform
  component versions MUST be declared as compatibility windows treated as
  contracts, not readiness claims. An unsupported version combination MUST
  resolve to a blocked state, never a best-effort run. Long parallel-run
  windows SHOULD be budgeted explicitly when deprecating widely used surfaces.
- **Acceptance evidence:** versioned schema set with namespaced identifiers;
  published deprecation policy; compatibility-window contract fixtures;
  conformance behavior test proving a blocked state on unsupported
  combinations; upgrade rehearsal drill evidence.
- **Non-goals:** Does not define platform release cadence or the release
  bill-of-materials process (DPL); does not promise indefinite backward
  compatibility.
- **Non-claims:** Versioning mechanics exist; no live deprecation cycle has yet
  been executed across real third-party modules, so window lengths are
  uncalibrated.
- **Stop conditions:** migration — halt rollout of a breaking standard change
  lacking published migration guidance and a dual-validation window; halt any
  module update whose compatibility window excludes the target platform
  version.
- **Traceability:** current-core; legacy-platform-a; req-history. Related:
  CR-OCS-050, CR-OCS-080.

### CR-OCS-130 — Workload-identity-only secret references
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator, auditor
- **Problem:** Connector packages and their evidence flow through registries,
  catalogs, and CI; any raw secret in these artifacts is an immediate
  platform-wide compromise.
- **Requirement:** Connectors MUST reference secrets exclusively through
  workload identity — a reference resolved at runtime by the platform's
  approved secrets workflow. Raw secret material (credentials, tokens, keys,
  kubeconfig content) MUST be rejected fail-closed in packages, documentation,
  fixtures, evidence, and exports. Rotation MUST be possible without package
  changes. Secret access MUST be auditable per service identity.
- **Acceptance evidence:** validator rejecting raw-material fixtures;
  source-safety scan coverage of connector trees; runtime contract test proving
  a module obtains credentials only via its workload identity; rotation drill
  evidence; audit log inspection of per-identity secret access.
- **Non-goals:** The secrets store and broker implementation and the identity
  policy model belong to IAM/FND; service-internal application secrets beyond
  brokered references are the service's own concern.
- **Non-claims:** Package- and reference-level enforcement is verified; runtime
  attestation coverage for every deployment profile (for example isolated
  cells) is not yet complete.
- **Stop conditions:** keys, exposure — halt and escalate immediately on raw
  secret material detected in a package, fixture, evidence bundle, log, or
  export; revoke and re-issue the affected material before continuing; halt
  module enablement wherever workload identity cannot be attested.
- **Traceability:** current-core; legacy-platform-a; req-acr-singular.
  Related: CR-OCS-010, CR-OCS-150.

### CR-OCS-140 — Declared durability surfaces and restore-test objectives
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, tenant, operator, auditor
- **Problem:** Stateful services are first-class; a service that cannot state
  how its data is classified, backed up, exported, and verifiably restored
  cannot be trusted with tenant data.
- **Requirement:** Every connector MUST declare a durability profile: state
  class, data classes, backup policy reference, recovery objective, and
  recovery evidence references. User-visible data-safety commitments (export
  and delete actions) MUST each carry an action reference and an evidence
  reference. Objectives of the form restore-test-required-before-production
  MUST be enforceable: a stateful service MUST NOT be catalog-enabled for
  production tenants without fresh restore-test evidence inside its declared
  freshness window.
- **Acceptance evidence:** conformance durability and data-lifecycle surface
  checks; restore-drill evidence recorded for the reference implementation and
  at least one stateful module; gate test proving production enablement is
  blocked on absent, stale, or synthetic-only restore evidence.
- **Non-goals:** Backup storage implementation, schedules, retention, and DR
  execution belong to STO/DAT; this contract declares surfaces and gates, not
  mechanics.
- **Non-claims:** Declared surfaces are conformance-verified; actual
  recovery-time and recovery-point performance of any real service is not
  claimed here and requires per-service drill evidence.
- **Stop conditions:** data, deletion — halt production enablement and escalate
  when restore-test evidence is absent, stale, synthetic-only, or failed; halt
  delete-path activation while export evidence for the affected data classes is
  missing.
- **Traceability:** current-core; vision-deck; req-acr-plural. Related:
  CR-OCS-030, CR-OCS-150, CR-OCS-160.

### CR-OCS-150 — Evidence bundles and machine receipts format
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, auditor, agent, service-team, provider
- **Problem:** Evidence-over-claims is a charter principle; without one receipt
  format, readiness, lifecycle, and billing claims cannot be compared,
  freshness cannot be enforced, and blocked states get laundered into
  readiness.
- **Requirement:** The platform MUST define one machine-emittable evidence and
  receipt format used by validation, conformance, lifecycle operations,
  registry operations, and durability drills. Every receipt MUST carry the
  issuing command or tool, UTC generation time, owning identity, claim,
  freshness window or policy, redaction policy, evidence references, explicit
  non-claims, and a review path. Receipts MUST be append-only and MUST
  represent blocked, stale, synthetic-only, and redacted states as first-class
  values that can never satisfy a claim of production-grade readiness.
- **Acceptance evidence:** receipt schema with fixtures for each receipt class
  (validation, conformance, lifecycle operation, registry operation, restore
  drill); freshness-gate test proving stale, blocked, or synthetic receipts
  block promotion; append-only ledger property test.
- **Non-goals:** Long-term evidence storage, retention, and private-evidence
  handling are FND/OPS concerns; this contract fixes format and semantics only.
- **Non-claims:** The format is specified and used on conformance and registry
  paths; uniform adoption across future surfaces (marketplace, federation) is
  not yet demonstrated.
- **Stop conditions:** trust — halt any promotion, publication, or enablement
  relying on stale, blocked, synthetic-only, or redacted evidence; halt and
  escalate on receipt mutation or backdating.
- **Traceability:** current-core; req-acr-singular; req-history. Related:
  CR-OCS-030, CR-OCS-060, CR-OCS-140, CR-OCS-160.

### CR-OCS-160 — Publication gate: catalog, tenant-access, support, and readiness surfaces
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider, operator, tenant
- **Problem:** A service listed for tenants without entitlement, support,
  readiness, or billing surfaces creates unowned risk: tenants can adopt
  something nobody can bill, support, or prove healthy.
- **Requirement:** Before any catalog publication, a connector package MUST
  declare its catalog entry, its tenant-access model (scope, entitlement
  reference, permissions), readiness checks (named checks with targets,
  conditions, and evidence references), a support surface (owner, diagnostics
  with a redaction boundary, documentation reference), its billing surface per
  CR-OCS-060, and its durability profile per CR-OCS-140. Publication MUST
  require a passing conformance report plus an explicit owner review recorded
  as evidence. Tenant access MUST fail closed: without an affirmative
  entitlement decision the service is invisible and unavailable to the tenant.
- **Acceptance evidence:** gate test suite attempting publication with each
  mandatory surface missing (all rejected); recorded owner-review receipt;
  entitlement fail-closed contract test; catalog rendering test driven only by
  package metadata.
- **Non-goals:** Marketplace economics, pricing, and merchandising (MKT); the
  entitlement decision engine itself (IAM).
- **Non-claims:** The gate is specified and enforced at package level; a full
  rehearsal of the publication workflow against a real marketplace instance has
  not yet occurred.
- **Stop conditions:** trust, exposure, money — halt publication on any missing
  mandatory surface, on any conformance problem, or on absent owner review;
  halt tenant visibility wherever entitlement checks error or degrade.
- **Traceability:** current-core; vision-deck; legacy-platform-a. Related:
  CR-OCS-020, CR-OCS-060, CR-OCS-090, CR-OCS-140.

### CR-OCS-170 — Distribution profiles and optional-module honesty
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, vendor
- **Problem:** Installations differ (foundation-only versus full); modules that
  silently assume optional infrastructure break minimal deployments, and
  channel confusion ships the wrong artifact to the wrong installation.
- **Requirement:** A connector package SHOULD declare distribution metadata:
  deployment profiles (for example optional-module, provider-managed,
  dedicated-cell), distribution channels, infrastructure targets, and an update
  policy reference. Modules not required for a foundation deployment MUST
  default to not-installed or disabled and MUST NOT block a foundation-only
  deployment or its readiness. Profile-specific requirements such as isolation
  or attestation MUST be declared per profile rather than assumed globally.
- **Acceptance evidence:** distribution-surface conformance check; deployment
  profile fixtures; evidence that a foundation-only profile validates and
  reports optional modules as not-installed without failing readiness;
  channel and infrastructure-target consistency contract test.
- **Non-goals:** Rollout machinery, canary strategy, and environment overlays
  are DPL; marketplace channel operation is MKT.
- **Non-claims:** Profiles are declarable and validated; isolated-cell
  requirements are declared but not yet exercised on real isolated
  infrastructure.
- **Stop conditions:** exposure, migration — halt distribution of a module to
  an infrastructure target outside its declared compatibility; halt and
  escalate any deployment plan that treats an optional module as
  foundation-blocking.
- **Traceability:** current-core; vision-deck; legacy-platform-a. Related:
  CR-OCS-080, CR-OCS-120.

### CR-OCS-180 — Migration bridge from broker-style service models
- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, vendor
- **Problem:** Teams arriving from broker-style ecosystems (a catalog,
  provision, deprovision mental model) face a conceptual cliff; without a
  mapping, adoption friction and mis-implemented connectors follow.
- **Requirement:** The standard SHOULD publish and maintain a migration guide
  mapping broker-style concepts (service and plan, provision and deprovision,
  bindings) to connector-package surfaces, explicitly naming the additional
  mandatory surfaces with no broker analogue: backup, restore, export, retry,
  rollback, meters, typed states, and evidence. The guide SHOULD include a
  worked migration example validated by the conformance suite.
- **Acceptance evidence:** migration guide document; worked example package
  passing conformance; documentation freshness review per standard version.
- **Non-goals:** Automated conversion tooling from broker implementations is
  not required; runtime broker-protocol compatibility is explicitly not
  claimed.
- **Non-claims:** The mapping is documented; its sufficiency is unproven until
  exercised by teams actually migrating from broker-based platforms.
- **Stop conditions:** migration — halt any guided migration step that would
  drop data-path surfaces (backup, export, delete) to ease adoption; escalate
  rather than weaken mandatory lifecycle requirements.
- **Traceability:** current-core; req-history. Related: CR-OCS-030, CR-OCS-110.

### CR-OCS-190 — Federation and commercial metadata declarations
- **Priority:** P2
- **Status:** proposed
- **Actors:** vendor, provider, service-team
- **Problem:** The federation vision needs services to declare cross-provider
  behavior and commercial terms early, but those planes do not exist yet;
  undecided semantics must not harden into unchangeable metadata.
- **Requirement:** A connector package MAY declare federation metadata
  (federation modes, message-bus reference, cross-provider scenarios,
  portability policy reference) and commercial metadata (roles, revenue model,
  license reference, expiry behavior, support reference). These declarations
  MUST be treated as forward-looking contracts: the platform MUST NOT derive
  settlement, license enforcement, or federation behavior from them until the
  corresponding planes exist, and the schema MUST mark these surfaces as
  experimental and unproven.
- **Acceptance evidence:** schema with explicit experimental marking;
  conformance treating these surfaces as optional-with-non-claims; a review
  checklist for promoting any such surface from experimental to normative.
- **Non-goals:** Federation transport, settlement mechanics, and license
  enforcement engines are FED/BIL/MKT scope; this requirement fixes only the
  declaration shape.
- **Non-claims:** Federation and commercial semantics are unproven; no
  settlement, revenue-share, or cross-provider runtime behavior is claimed or
  implied by these fields today.
- **Stop conditions:** money, settlement, trust — halt and escalate if any
  system consumes these declarations for charging, settlement, or access
  decisions before the owning plane's contracts are ratified; halt schema
  promotion of these surfaces without owner review.
- **Traceability:** vision-deck; current-core. Related: CR-OCS-120, CR-OCS-160.

### CR-OCS-200 — Automation and analytics event declarations
- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, vendor, operator, agent, tenant
- **Problem:** Services need to expose automation tasks for operators, tenants,
  and agents, plus product analytics; without declared surfaces, automation
  cannot be governed and analytics becomes an uncontrolled data-exfiltration
  path.
- **Requirement:** A connector SHOULD declare at least one automation task — a
  typed workflow for operator, tenant, or agent use — with explicit inputs, a
  risk class, and rollback behavior where mutating. Product analytics events
  SHOULD be declared by name and payload class; payloads MUST exclude
  credentials, secrets, and tenant content, and MUST respect the platform's
  consent and redaction boundaries.
- **Acceptance evidence:** automation-surface conformance check (at least one
  task with declared risk class); analytics event schema validation; redaction
  contract test scanning declared payloads for forbidden classes; negative
  fixture with a secret-bearing payload rejected.
- **Non-goals:** The agent runtime executing automation is AGT scope; analytics
  storage and pipelines are OBS scope.
- **Non-claims:** Declared surfaces are validated; governance of agent-executed
  automation against real services is not yet exercised.
- **Stop conditions:** data, exposure, keys — halt an automation task whose
  declared risk class understates its effects (for example unmarked deletion);
  halt analytics transmission on detection of tenant-content or credential
  classes in payloads.
- **Traceability:** current-core; vision-deck; req-acr-plural. Related:
  CR-OCS-030, CR-OCS-040.

## Coverage notes

This domain deliberately defers:

- **FND** — platform runtime policy (Go-first, upstream Kubernetes), evidence
  storage and retention infrastructure, source-safety scanning machinery.
- **IAM** — identity providers, the authorization decision engine, entitlement
  evaluation internals, and the secrets store/broker that resolves workload
  identity references.
- **BIL** — charging, rating, invoicing, tariff and price modeling, payment
  accounts, and settlement execution; OCS declares only the metering
  transmission contract.
- **MKT** — marketplace catalog UX, merchandising, revenue-share economics, and
  purchase flows; OCS declares the surfaces a listing must have.
- **CUX** — portal shell implementation, navigation, UX copy, and the console
  design system; OCS declares only the microfrontend host contract.
- **OBS** — the metrics, logging, tracing, and alerting stacks; OCS declares
  analytics and diagnostics surfaces, not their pipelines.
- **OPS** — SRE practice, support case operations, and incident management;
  OCS declares the support surface a service must publish.
- **DPL** — deployment machinery, GitOps overlays, canary and release trains,
  environment profiles, and the release bill-of-materials process.
- **STO / DAT** — backup and restore implementation, schedules, retention, DR
  execution, and the data services themselves; OCS declares durability surfaces
  and gates only.
- **FED** — the federation data bus, cross-provider connectivity, global
  portal, and settlement mechanics; OCS keeps federation metadata experimental.
- **AGT** — the agent runtime and its governance; OCS declares automation task
  surfaces that runtime may consume.
- **CMP / NET / K8S** — the compute, network, and container substrate that
  connector workloads run on.
