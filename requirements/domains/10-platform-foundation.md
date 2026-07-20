# 10 — Platform Foundation

This domain fixes the mission-level and runtime-level rules every other
domain inherits: contract-before-technology, replaceable capability
profiles for the virtualization, network, storage, and runtime layers,
the Go-first runtime policy, and the upstream-Kubernetes-only substrate
target. It also governs the repository itself: the public-core /
private-workspace topology with no duplicated core, the source-safety
boundary, the gated publication path, the declared product-primitive
minimum, portability and jurisdiction freedom, documentation as a
machine-checkable contract, English-only artifacts, the Apache-2.0 legal
scaffold, and the production-honesty rules that gate every claim made
anywhere else in this corpus.

**Domain contract.** The platform core depends on contracts and metadata,
never on implementations; platform runtime work is Go-first on upstream
Kubernetes with legacy material classified, never silently reused; no
private endpoint, tenant datum, credential, or copied private text ever
enters a public artifact; publication to the public core is machine-gated
end to end; and no readiness, durability, or delivery claim is ever made
without fresh verifiable evidence — blocked stays blocked, fixtures are
labeled fixtures, and production state is never hardcoded, faked, or held
in memory alone.

## Requirements

### CR-FND-010 — Contract before technology
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, service-team, vendor, operator, agent
- **Problem:** Platforms that hard-wire one virtualization, network,
  storage, or runtime technology into their core become unmovable within
  a single technology cycle and recreate exactly the lock-in this
  platform exists to remove.
- **Requirement:** Every platform capability MUST be defined first as a
  versioned, declarative contract expressed as outcomes, states, and
  evidence — never as the mechanics of a specific vendor, framework,
  language, or distribution. The platform core MUST depend only on
  contract metadata and published APIs; any direct dependency of the core
  on a service or infrastructure implementation is a defect. Technology
  choices MUST be expressed as replaceable profiles behind those
  contracts (see CR-FND-040).
- **Acceptance evidence:** Contract schemas and validators in the public
  core with passing conformance runs; CI checks that reject
  implementation-specific references inside core contract surfaces;
  ownership-classification verifier output proving core packages carry no
  provider-specific imports.
- **Non-goals:** Prescribing one blessed implementation per layer;
  forbidding opinionated reference implementations (reference code may
  exist, but the core must not depend on it).
- **Non-claims:** Not all layers yet have complete contracts; several
  platform surfaces are still converging on the contract model.
- **Stop conditions:** Halt and escalate to owner review any change that
  wires service-specific or provider-specific logic, UI, billing, or
  automation into the platform core; block promotion of any capability
  whose contract names an implementation rather than an outcome (trust,
  migration risk).
- **Traceability:** vision-deck; req-history; req-acr-singular;
  current-core. Related: CR-FND-040, CR-FND-150.

### CR-FND-020 — Go-first platform runtime
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, agent
- **Problem:** A polyglot runtime core multiplies operational burden,
  fragments tooling and review depth, and contradicts the single
  auditable runtime policy the platform was rebuilt around.
- **Requirement:** New platform runtime work — controllers, CLIs,
  validators, control-plane services — MUST be implemented in Go. Non-Go
  runtime surfaces MUST carry an explicit classification (retained
  tooling, migration debt with a replacement path, or an approved dated
  exception) and MUST NOT grow without owner review. Tooling scripts
  MUST NOT be presented as runtime capabilities.
- **Acceptance evidence:** The runtime-policy guard test suite passing in
  CI; a dated language-inventory baseline showing no unclassified non-Go
  runtime files; policy-check output attached to readiness reports.
- **Non-goals:** Forbidding service teams from implementing connector
  services in other languages (service implementations sit outside the
  platform runtime); rewriting existing debt ahead of higher-risk work.
- **Non-claims:** Migration debt still exists and is tracked, not
  eliminated; this requirement states policy and its enforcement, not
  completion of the migration.
- **Stop conditions:** Halt any platform-runtime change that introduces
  or expands a non-Go runtime surface without an owner-approved dated
  exception; block readiness claims for core services that bypass the
  policy (migration risk).
- **Traceability:** current-core; req-history; req-acr-plural. Related:
  CR-FND-030, CR-FND-180, CR-FND-190.

### CR-FND-030 — Upstream-Kubernetes-only substrate target
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, agent
- **Problem:** Readiness evidence gathered on legacy lightweight
  Kubernetes distributions was historically over-read as proof for the
  upstream target, producing false readiness and hidden substrate
  lock-in.
- **Requirement:** The platform's only supported substrate target MUST be
  upstream Kubernetes, deployed and upgraded with kubeadm semantics.
  Evidence produced on legacy or lightweight distributions MUST be
  classified as legacy and MUST NOT satisfy upstream readiness claims.
  Substrate compatibility shims that substitute for migration to the
  upstream target MUST NOT be introduced. Substrate profiles MUST declare
  capabilities, upgrade boundary, exit path, and explicitly unsupported
  states rather than installation commands.
- **Acceptance evidence:** Substrate-policy checks in CI rejecting legacy
  references outside classified history/debt paths; readiness reports
  that distinguish target architecture from fallback planning; a recorded
  blocker evidence class for legacy-substrate sightings; deployment-wave
  evidence (owned by the deployment domain) naming the upstream target.
- **Non-goals:** Forbidding developer-local experimentation on other
  substrates; denying that legacy stands exist — they are recorded as
  classified debt.
- **Non-claims:** A live upstream-Kubernetes production stand is not
  claimed here; live stand evidence is owned by the deployment and
  operations domains.
- **Stop conditions:** Block any readiness claim whose evidence was
  produced on a legacy substrate; halt and escalate any proposal to add
  a compatibility shim or to re-label a legacy stand as production
  (migration, trust risk).
- **Traceability:** current-core; req-history; req-acr-plural. Related:
  CR-FND-020, CR-FND-180.

### CR-FND-040 — Layered replaceable capability profiles
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, agent
- **Problem:** If any infrastructure layer is defined by its current
  implementation instead of a contract, replacing that layer later
  becomes a platform rewrite and an exit-blocking lock-in.
- **Requirement:** Each infrastructure layer — virtualization, network,
  storage, and runtime — MUST be described by a capability profile
  declaring provided capabilities, limits, portability boundary, exit
  path, upgrade boundary, degraded states, and explicitly unsupported
  states. Profiles MUST express capabilities and contracts, not
  installation commands. A layer MAY ship with a single profile, but no
  layer may be accepted without its profile contract. Alternate profiles
  behind the same contract SHOULD be introducible without core changes.
- **Acceptance evidence:** Profile contract schemas with validation
  tests; invalid fixtures proving rejection of profiles that lack exit,
  upgrade, or unsupported-state declarations; readiness reports citing
  profile contracts for each layer.
- **Non-goals:** Requiring multiple working implementations per layer at
  first release; standardizing implementations rather than contracts.
- **Non-claims:** Only a subset of layers has mature profiles today;
  multi-profile interchangeability is designed but not yet demonstrated
  on all layers.
- **Stop conditions:** Halt promotion of any layer whose profile lacks a
  portability or exit declaration; block any change that makes the core
  depend on one profile's specifics (migration, trust risk).
- **Traceability:** vision-deck; req-history; req-acr-singular;
  current-core. Related: CR-FND-010, CR-FND-070.

### CR-FND-050 — Public-core / private-workspace boundary
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Copying the public core into private workspaces (or
  private material into the public core) creates drift and dual
  maintenance, and — at worst — leaks private endpoints, tenant data, or
  credentials into a public repository irreversibly.
- **Requirement:** The public open-source repository MUST be the
  canonical source of the platform core. Private workspaces MUST consume
  it via submodule or gitlink, never via copied trees; duplicated core
  source is a defect. Provider-private implementation paths MUST be
  machine-classified as non-publishable, and the classification guard
  MUST fail closed if that classification is ever flipped. Import-path
  rewrites from private module paths to the public module path MUST be
  complete at export; unresolved private paths are block-severity
  findings.
- **Acceptance evidence:** Ownership-classification manifest plus
  verifier test that fails on boundary violations; export-manifest
  denylist with block-severity rules and recorded clean runs; CI
  source-safety scans; submodule-state checks recorded in the publication
  runbook evidence.
- **Non-goals:** Preventing private extensions that consume public
  contracts; requiring private workspaces or reference-installation
  overlays to be public.
- **Non-claims:** Classification and scanning are continuously enforced
  but do not prove absolute absence of all private identifiers — that
  remains a standing non-claim of the source-safety program.
- **Stop conditions:** Halt any export, merge, or publication on a
  denylist, unresolved-import, or ownership-guard finding; never proceed
  by annotation or waiver on exposure-class findings without an explicit
  owner decision (keys, exposure risk).
- **Traceability:** current-core. Related: CR-FND-160, CR-FND-170.

### CR-FND-060 — Declared product-primitive minimum
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, service-team, agent, auditor
- **Problem:** "A cloud" is undefined without an explicit product
  surface; platforms have historically hidden missing primitives behind
  dashboards and marketing language.
- **Requirement:** The platform MUST publish a primitive coverage map
  spanning the full application lifecycle: data and migrations; identity,
  access, and sessions; object and file storage; realtime events and
  notification fan-out; short-running functions, jobs, and schedules;
  long-running workloads; frontend and site delivery; AI/model access via
  a portable gateway; secrets and configuration; observability, logs,
  audit, and diagnostics; usage, billing, and quotas; backup, restore,
  export, and exit. Each primitive MUST declare owner, lifecycle, limits,
  security boundary, portability story, evidence story, and an honest
  maturity state. A missing or immature primitive MUST be visible as such
  and MUST NOT be masked as platform readiness. Every primitive
  integration MUST go through a published portable contract or a
  documented exception.
- **Acceptance evidence:** The published primitive map with per-primitive
  status and contract links; readiness reports referencing the map;
  review checklist records proving no readiness claim hides an absent
  primitive.
- **Non-goals:** Requiring all primitives to be mature at the first
  deployable release; fixing implementation technologies per primitive.
- **Non-claims:** Several primitives are declared but not yet delivered;
  maturity states are honest declarations, not delivery claims.
- **Stop conditions:** Block any platform readiness claim that omits or
  downgrades a missing primitive; halt publication of any capability
  presentation that implies coverage the map does not show (trust risk).
- **Traceability:** req-acr-plural; req-acr-singular; vision-deck.
  Related: CR-FND-130, CR-FND-150.

### CR-FND-070 — Portability and jurisdiction freedom by design
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, auditor, agent
- **Problem:** Lock-in hides in data gravity, identity mappings, billing
  history, and silent cross-jurisdiction movement of backups, logs, and
  telemetry — not just in workload placement.
- **Requirement:** The platform MUST be designed so that no accepted
  component makes tenant exit, provider change, or jurisdiction move
  impossible by construction. Every capability contract MUST declare what
  is exportable, in what format, and which portability gaps exist.
  Jurisdiction MUST be a first-class, visible attribute of workloads,
  data sets, backups, logs, telemetry, support access, and key custody.
  Actions that move data or metadata across jurisdictions MUST be visible
  before execution, and residency-policy violations MUST fail closed.
- **Acceptance evidence:** Portability and jurisdiction sections present
  in capability contracts with schema validation; policy tests
  demonstrating fail-closed residency violations; audit evidence that
  jurisdiction attributes are queryable through API and CLI.
- **Non-goals:** Implementing per-primitive tested exit packages in this
  domain (owned by the storage, data, and federation domains); providing
  legal advice.
- **Non-claims:** Tested exit per primitive is not yet proven;
  jurisdiction coverage for support access and key custody is designed,
  not evidenced.
- **Stop conditions:** Halt any mutation that moves data, backups, logs,
  or telemetry across a jurisdiction boundary without explicit policy
  approval; block acceptance of any component whose design has no exit
  path (data, migration, trust risk).
- **Traceability:** vision-deck; req-acr-plural; req-acr-singular;
  req-history. Related: CR-FND-040, CR-FND-130.

### CR-FND-080 — Documentation as operating contract
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, agent, vendor
- **Problem:** Documentation that drifts from the real command and API
  surface becomes an operational hazard — humans and agents act on stale
  instructions and call the damage a surprise.
- **Requirement:** Platform documentation MUST be organized by audience,
  task, and stage, and MUST state intent and product promise before
  mechanism. Command references MUST be generated from, or validated
  against, the real command surface. Every document MUST carry visible
  ownership and freshness. Onboarding documentation MUST state the
  automated path, manual fallback, prerequisites, and stop conditions.
  Readiness-critical flows MUST NOT be called ready while their
  documentation is incomplete.
- **Acceptance evidence:** Documentation publication-gate report
  (navigation coverage, link check, freshness markers); CI validation of
  documented commands against the CLI surface; documentation-completeness
  entries inside readiness reports.
- **Non-goals:** Exhaustive prose for unreleased capabilities;
  duplicating machine-generated API reference content by hand.
- **Non-claims:** Full documentation coverage across all domains is not
  yet achieved; command-surface validation covers the core CLIs only.
- **Stop conditions:** Halt publication of documentation that fails link,
  freshness, or command-validation checks; block readiness claims for
  flows whose documentation is incomplete (trust risk).
- **Traceability:** req-history; req-acr-plural; current-core. Related:
  CR-FND-090, CR-FND-130.

### CR-FND-090 — Machine-checkable documentation and corpus
- **Priority:** P1
- **Status:** proposed
- **Actors:** agent, operator, auditor
- **Problem:** Prose-only contracts cannot be enforced, and drift between
  human documents and machine registries silently invalidates governance.
- **Requirement:** Normative platform contracts SHOULD be authored as
  human prose paired with a machine-readable twin (schema, contract
  document, or registry entry) carrying a versioned schema identifier and
  a dated review marker. Each readiness-gating contract MUST have a
  verifier with a stable expected success marker, and validation failures
  MUST be machine-readable with surface, field, message, and remediation.
  Machine registries MUST be generated from the corpus, and drift between
  corpus and registry MUST fail validation.
- **Acceptance evidence:** Prose-plus-machine contract pairs with passing
  verifier runs; conformance reports containing structured problem lists;
  registry generation with a zero-drift check in CI; invalid fixtures
  proving verifier rejection.
- **Non-goals:** Replacing human-readable rationale with JSON;
  machine-generating normative prose.
- **Non-claims:** Not all contracts yet have machine twins; verifier
  coverage is partial and expanding.
- **Stop conditions:** Block merge when corpus-to-registry drift or
  contract-validation failures exist; halt readiness claims that cite
  contracts lacking verifiers (trust risk).
- **Traceability:** current-core; req-history; req-acr-singular.
  Related: CR-FND-080, CR-FND-200.

### CR-FND-100 — English-only artifacts
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, service-team
- **Problem:** Mixed-language artifacts fragment review, exclude
  international contributors, and historically allowed unreviewed prose
  to persist outside governance.
- **Requirement:** All repository artifacts — requirements,
  documentation, contracts, shared-code comments, and commit-facing
  metadata — MUST be authored in English. Automated gates MUST reject
  non-English prose and encoding defects in artifacts. Conversation
  language between humans and agents is independent of artifact language.
- **Acceptance evidence:** Language and encoding gate output in CI; a
  dated language baseline; zero-finding scans attached to publication
  evidence.
- **Non-goals:** Translating historical archived corpora (they remain
  paraphrase sources, not artifacts of this repository); restricting
  end-user localization of shipped product UI (owned by the portal
  domain).
- **Non-claims:** Historical archives in other languages remain outside
  this repository's governance and are not claimed as governed.
- **Stop conditions:** n/a.
- **Traceability:** req-history; req-acr-singular; current-core.
  Related: CR-FND-160.

### CR-FND-110 — License and contribution clarity
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, vendor, service-team, auditor
- **Problem:** An open platform without unambiguous licensing,
  contribution terms, and trademark separation cannot safely accept
  external contributions or be adopted by businesses.
- **Requirement:** The public repository MUST carry a complete legal
  scaffold: the Apache-2.0 license text, NOTICE, contribution guide,
  Developer Certificate of Origin sign-off policy, Contributor License
  Agreement, trademark policy separating the software license from the
  marks, and governance, ownership, and security-disclosure documents.
  CI MUST check the presence of the required legal documents and enforce
  DCO sign-off and CLA presence on contributions. The CLA SHOULD preserve
  the owner's ability to offer commercial or dual licensing of future
  versions while published versions keep their original license. Counsel
  MUST review the CLA and trademark texts before broad external
  contribution waves open.
- **Acceptance evidence:** Required-documents check in CI; DCO/CLA
  workflow records on pull requests; license-scan workflow output; a
  recorded counsel-review item with status tracked before contribution
  waves.
- **Non-goals:** Legal advice to adopters; guaranteeing trademark rights
  beyond the policy text.
- **Non-claims:** Counsel review of the CLA and trademark texts is
  pending; external contribution intake is not yet open at scale.
- **Stop conditions:** Halt external contribution intake and release
  tagging if any required legal document is missing or license, DCO, or
  CLA checks fail; escalate any proposed license change to an owner
  decision before merge (trust, exposure risk).
- **Traceability:** current-core. Related: CR-FND-170.

### CR-FND-120 — Production-honesty bans
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, agent, auditor
- **Problem:** Hardcoded success paths, fixture-backed UI presented as
  production, and in-memory "persistence" for production state create
  false readiness that later detonates as data loss or broken trust.
- **Requirement:** The platform MUST NOT contain hardcoded success
  results in verification, readiness, or user-facing flows. Fixture or
  synthetic data MUST be visibly labeled and MUST NOT be presented as
  production state on any surface. Production state MUST NOT be held in
  in-memory-only structures; durable persistence is mandatory for
  anything presented as durable. Hidden costs, hidden defaults, and
  hidden public exposure MUST NOT be introduced before explicit user
  commit.
- **Acceptance evidence:** Honesty-focused test suites and source checks
  detecting fixed success returns in verification paths; UI contract
  checks proving fixture labeling; contract tests for persistence of
  production state; readiness reports listing these checks by name.
- **Non-goals:** Forbidding synthetic fixtures for development and
  conformance (they are required, clearly labeled); banning caches that
  are not the system of record.
- **Non-claims:** Detection is test- and review-based; absence of all
  dishonest patterns is verified to the extent of those checks, not
  absolutely.
- **Stop conditions:** Any discovered hardcoded-success path, unlabeled
  fixture, or in-memory production state blocks the related readiness
  claim until fixed and re-verified; halt exposure of any surface with
  hidden cost, default, or public exposure (trust, exposure, data, money
  risk).
- **Traceability:** vision-deck; req-acr-singular; current-core.
  Related: CR-FND-130, CR-FND-140.

### CR-FND-130 — Evidence before readiness; blocked stays blocked
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, agent, auditor
- **Problem:** Readiness claims assembled from intent, stale artifacts,
  or laundered blocked evidence are the platform's central credibility
  risk.
- **Requirement:** Every readiness, durability, security, or delivery
  claim MUST cite fresh, verifiable evidence of the appropriate class.
  Evidence states MUST include at minimum accepted, absent, stale,
  blocked, synthetic-only, and redacted; only verified non-synthetic
  evidence promotes. Blocked and stale evidence are first-class outcomes
  and MUST NOT be converted into, or aggregated toward, readiness claims.
  When full evidence is absent, narrower honest claim labels (for example
  lab-ready, contract-ready, preview-ready, blocked-with-evidence) MUST
  be used instead. Evidence records MUST be append-only with UTC
  timestamps.
- **Acceptance evidence:** The evidence-freshness contract and its
  validator; iteration-gate reports showing pass/fail/blocked/
  not-applicable per gate with evidence paths; CI checks rejecting
  promotion on stale or blocked evidence; status-taxonomy tests.
- **Non-goals:** Defining per-domain evidence classes (owned by the
  observability, operations, and deployment domains); forbidding
  synthetic evidence for development — it is allowed but never promotes.
- **Non-claims:** Freshness windows are defined for core evidence classes
  only; full per-class coverage is still expanding.
- **Stop conditions:** One blocked, stale, or absent mandatory gate
  blocks the entire related claim; halt and record — never downgrade a
  blocked state to a warning in order to proceed (trust risk).
- **Traceability:** current-core; req-acr-singular; req-acr-plural;
  req-history. Related: CR-FND-120, CR-FND-140.

### CR-FND-140 — One product truth across surfaces
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, operator, agent, service-team
- **Problem:** When the portal, API, CLI, and agent surfaces tell
  different stories about the same object, every downstream decision —
  human or automated — is corrupted.
- **Requirement:** All platform surfaces MUST share one typed state
  vocabulary — at minimum ready, denied, degraded, blocked, retryable —
  with reason, evidence reference, and remediation attached. The same
  object MUST report the same state, blockers, and next actions on every
  surface; divergence is a defect. Structured results MUST be the primary
  output form; human-formatted text MUST NOT be the only machine-relevant
  channel. Denials MUST name the missing rights, context, evidence, or
  approval.
- **Acceptance evidence:** State-vocabulary conformance checks in the
  connector validator; cross-surface consistency tests or recorded
  comparisons; contract tests proving denial responses name the missing
  precondition.
- **Non-goals:** Forcing identical visual presentation across surfaces;
  removing human-friendly formatting layered on top of structured truth.
- **Non-claims:** Full automated cross-surface equivalence testing is not
  yet in place; current assurance is contract- and conformance-level.
- **Stop conditions:** Halt promotion of any capability whose surfaces
  report divergent states; treat unresolved divergence as a blocking
  defect, not a documentation issue (trust risk).
- **Traceability:** req-acr-singular; req-acr-plural; current-core.
  Related: CR-FND-120, CR-FND-130.

### CR-FND-150 — Real open baseline
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, vendor, operator, auditor
- **Problem:** An open-source tier that cannot actually deploy and
  operate a provider is a teaser, and teaser economics contradict the
  platform's stated ecosystem model.
- **Requirement:** The open-source layer MUST be genuinely sufficient to
  install, deploy, and operate a working provider: installation tooling,
  the infrastructure and services substrate, the connector standard, and
  operating documentation are all in scope of the public core. Commercial
  extensions MUST plug into published open interfaces and MUST NOT be
  required for a functional baseline deployment. Optional modules MUST
  default to not-installed or disabled and MUST NOT block a
  foundation-only deployment.
- **Acceptance evidence:** Foundation-only deployment profile evidence
  deployed entirely from IaC, with optional modules reported
  not-installed without failing readiness; public documentation
  sufficient for a third party to deploy without private material;
  interface contracts that commercial extensions consume.
- **Non-goals:** Forbidding commercial extensions; requiring the baseline
  to include every optional module.
- **Non-claims:** The full reference-installation evidence loop (deploy,
  upgrade, backup, restore, one-server-loss) is owned by the deployment
  and operations domains and is not re-claimed here.
- **Stop conditions:** Block any "open baseline" claim while a baseline
  deployment requires private or commercial components; halt introduction
  of mandatory dependencies from the open core onto commercial tiers
  (trust risk).
- **Traceability:** vision-deck; req-history; current-core. Related:
  CR-FND-010, CR-FND-060.

### CR-FND-160 — Source-safety boundary for all artifacts
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Secrets, tenant data, private endpoints, host-local paths,
  or copied private source text entering public artifacts is the
  platform's highest-severity exposure risk and is irreversible once
  published.
- **Requirement:** No artifact — code, documentation, requirements,
  evidence, examples, or fixtures — may contain credentials, tenant data,
  private endpoints, host-local paths, or copied text from private
  sources. Secrets MUST appear only as brokered references. Source-safety
  scanning MUST run as an automated merge and push gate, and findings
  MUST be block-severity. Defense in depth MUST be maintained: a
  manifest-driven denylist, content-pattern scans, and ownership
  classification guards are all required; no single mechanism is
  sufficient alone. Secret-like names alone MUST trigger redaction
  handling even when contents are never opened.
- **Acceptance evidence:** Source-safety scan results in CI and pre-push
  hooks; export-manifest content rules with recorded clean runs; negative
  fixtures proving the scanner rejects seeded credential patterns;
  periodic audit of scan coverage.
- **Non-goals:** Proving absolute absence of all private identifiers (an
  explicit standing non-claim); scanning sources outside the repository
  boundary.
- **Non-claims:** Automated scans prove generic-pattern absence only;
  owner-held private denylist terms are deliberately not embedded in the
  repository, so absolute clearance is not claimed.
- **Stop conditions:** Halt merge, push, export, or publication on any
  source-safety finding; never proceed by annotation on exposure-class
  findings without an explicit owner decision; treat any discovered
  committed secret as an incident requiring rotation (keys, exposure
  risk).
- **Traceability:** current-core; req-history; req-acr-plural. Related:
  CR-FND-050, CR-FND-170.

### CR-FND-170 — Gated public publication path
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Direct pushes to a public main branch bypass every safety
  gate and make exposure events likely; repository mutations can also
  silently break the running reference installation.
- **Requirement:** Publication to the public repository MUST follow the
  gated path: manifest-driven export, verification, candidate branch,
  pull request, required status checks (tests, contract validation,
  conformance, source-safety, documentation, license, and contribution
  checks), and merge only after all checks are green. Direct pushes to
  the public main branch MUST be technically prevented. The health of the
  reference installation MUST be verified before and after every
  publication-affecting mutation; a blocked reference installation halts
  publication. Branch protection MUST mirror the gated path.
- **Acceptance evidence:** Publication tooling runs producing export,
  verify, and push receipts; branch-protection configuration evidence;
  recorded pre- and post-mutation health snapshots of the reference
  installation; failed-run fixtures proving fail-closed behavior on
  missing checks.
- **Non-goals:** Publishing private workspace internals; automating
  merges without required checks.
- **Non-claims:** Required-check coverage depends on the hosted CI tier's
  feature set; reference-installation gating proves the current
  installation only, not arbitrary future ones.
- **Stop conditions:** Halt publication on any missing, skipped, or
  failed required check; halt on an unhealthy reference installation;
  never force-push or bypass branch protection (exposure, trust risk).
- **Traceability:** current-core. Related: CR-FND-050, CR-FND-160.

### CR-FND-180 — Classified legacy references; no substitution shims
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Unclassified legacy runtime and substrate references blur
  the target architecture and invite shortcuts that quietly become
  "production".
- **Requirement:** Every legacy runtime or substrate reference in the
  repository MUST carry an explicit classification: retained history,
  blocked migration debt, retired, or approved dated exception.
  Compatibility shims that substitute for migration to the target
  architecture MUST NOT be introduced without an owner-approved reversal.
  Legacy stands and legacy evidence MUST NOT be presented as
  target-architecture readiness.
- **Acceptance evidence:** A classification inventory with dated review;
  policy-guard tests rejecting unclassified legacy references; readiness
  reports separating target architecture from fallback planning.
- **Non-goals:** Deleting history; forbidding study of legacy material
  for lessons.
- **Non-claims:** Classification coverage is maintained for the core
  repositories; peripheral archives may contain unclassified legacy
  material and are marked as archives.
- **Stop conditions:** Block readiness claims that cite legacy-substrate
  evidence; halt introduction of shims pending owner decision; escalate
  unclassified legacy findings into a tracked classification task
  (migration risk).
- **Traceability:** current-core; req-history; req-acr-plural. Related:
  CR-FND-020, CR-FND-030.

### CR-FND-190 — Dated, rerunnable inventory baselines
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Architecture-direction claims — "the repository is
  Go-first", "legacy references are declining" — go stale and quietly
  become false.
- **Requirement:** Any architecture-direction claim about runtime
  languages, substrate references, or repository composition MUST be
  backed by a dated, rerunnable inventory baseline with a recorded
  generation command. Stale baselines MUST downgrade or withdraw the
  dependent claim. Baseline reruns SHOULD be scheduled and attached to
  iteration evidence.
- **Acceptance evidence:** Baseline generation command plus dated output
  artifacts; claim documents linking their baselines; a recorded
  downgrade or drill instance where a stale baseline withdrew a claim.
- **Non-goals:** Real-time composition monitoring; blocking unrelated
  work on unchanged baselines.
- **Non-claims:** Baselines exist for the core repositories; coverage of
  auxiliary tooling repositories is incomplete.
- **Stop conditions:** Withdraw or downgrade any architecture-direction
  claim whose baseline is stale; do not restate it until a rerun
  refreshes it (trust risk).
- **Traceability:** req-history; current-core. Related: CR-FND-020,
  CR-FND-180.

### CR-FND-200 — Requirements corpus governance
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, agent, auditor
- **Problem:** Requirements without stable identity, mandatory fields,
  and review discipline decay into unverifiable wishes.
- **Requirement:** Every requirement in this corpus MUST follow the
  schema in `01-requirement-schema.md`: stable sparse ID, priority,
  status lifecycle, actors, problem, normative requirement, acceptance
  evidence, non-goals, non-claims, stop conditions where risk classes
  apply, and traceability. Status changes to accepted MUST go through
  owner review. IDs MUST NOT be renumbered or reused after retirement.
  A machine registry MUST be generated from the corpus, and drift between
  corpus and registry MUST fail validation. Retired requirements remain
  in place with the reason recorded.
- **Acceptance evidence:** Schema lint over the corpus; the generated
  registry plus zero-drift check in CI; review records for status
  transitions.
- **Non-goals:** Capturing implementation plans (this corpus states WHAT
  and WHY); duplicating delivery evidence (implementation evidence links
  back here).
- **Non-claims:** The corpus is newly authored; all requirements are
  proposed and none are yet owner-accepted.
- **Stop conditions:** A requirement whose why, acceptance evidence, or
  non-claims cannot be filled is not ready — halt and rework rather than
  publishing hollow entries; block acceptance transitions without owner
  review (trust risk).
- **Traceability:** current-core; req-history; req-acr-singular;
  req-acr-plural. Related: CR-FND-090.

### CR-FND-210 — Cross-platform operator workflows
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, agent
- **Problem:** Operator workflows that assume one shell or operating
  system silently fail for contributors on other platforms and produce
  environment-specific behavior in safety-relevant tooling.
- **Requirement:** Operator-facing workflows and scripts SHOULD be
  runnable across mainstream development platforms. Durable,
  safety-relevant behavior MUST be implemented in Go under the runtime
  policy rather than in shell; shell wrappers, where unavoidable, MUST be
  minimal and MUST NOT carry policy logic.
- **Acceptance evidence:** CI matrix runs or recorded executions of core
  CLI verification on multiple platforms; an inventory of shell-specific
  logic with Go replacements tracked.
- **Non-goals:** Supporting every exotic environment; banning shell for
  thin wrappers.
- **Non-claims:** Some legacy tutorials still assume a POSIX environment;
  full matrix coverage is not yet scheduled.
- **Stop conditions:** n/a.
- **Traceability:** current-core; req-history. Related: CR-FND-020.

## Coverage notes

This domain deliberately defers the following to sibling domains:

- OCS connector package surfaces, the conformance engine's detailed
  checks, and the minimum-useful-connector profile → `17-ocs-service-connectors.md`.
- Identity domains, production identity gates (issuer discovery, token
  policy, break-glass), and secrets/key custody mechanics → `15-iam-identity-security.md`.
- Usage meters, billing events, settlement, and commercial order flow → `16-billing-finops.md`.
- Per-primitive capability contracts for compute, network, storage, and
  managed data services → `11-compute-virtualization.md`, `12-network.md`,
  `13-storage-backup-dr.md`, `24-data-services.md`.
- Readiness-gate execution, preflight and deploy waves, restore drills,
  and one-server-loss evidence → `22-deployment-iac-cicd.md` and `21-ops-sre-support.md`.
- Evidence freshness classes, observability bundles, and audit/UTC
  mechanics → `20-observability.md`.
- Portal shell behavior and UI rendering of the shared state vocabulary → `19-portal-ux-selfservice.md`.
- Agent approval matrix, risk classes, and agent runtime context → `25-agent-governance.md`.
- Federation, edge, and disconnected-mode behavior → `23-federation-global-portal.md`.
- Marketplace publication economics and service-card content → `18-marketplace-catalog.md`.
