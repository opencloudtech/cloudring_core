# 23 — Federation & Global Portal

Domain scope:

- Peer-to-peer and bilateral cross-provider data synchronization (catalog,
  availability, pricing/terms metadata).
- Federation participant registry, federated identity trust, and federation
  governance (admission, suspension, disputes, certification).
- Cross-cloud connect interconnects and cross-provider resilience scenarios
  (DR, replication, backup, CDN-class delivery).
- Tenant migration between providers and jurisdiction portability with
  exit/migration evidence.
- EDGE zones in connected, disconnected, and deferred-sync modes.
- Global white-label cloud portal and requirement-based multi-provider
  placement (closest/cheapest/residency-constrained).
- Cross-provider settlement hooks, revenue sharing, and cross-licensing
  interfaces (OSS hooks only; implementations are Business-layer).

**Domain contract.** Federation is the platform's founding vision and its
least proven capability, and this domain treats both facts as binding. Nothing
in this domain may be presented as ready, enabled by default, or promised to a
tenant without fresh linked evidence; `blocked` and `declared` are first-class
honest states. No federation mechanism may require a universal trusted
operator, reintroduce single-provider assumptions into base schemas, or move
tenant data, keys, money, or identity across participant or jurisdiction
boundaries without explicit consent, fail-closed policy gates, and auditable
evidence. Settlement and licensing surfaces in the OSS core are hooks and
contracts only — the commercial implementations live outside this repository.
Where this domain conflicts with a desire to ship the vision early, the
evidence discipline wins.

## Requirements

### CR-FED-010 — Federation-ready base schemas (no single-provider assumptions)
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, vendor
- **Problem:** Platforms built for one owner hardcode single-provider
  assumptions into identity, catalog, metering, and billing schemas;
  retrofitting multi-participant scope later is prohibitively expensive and
  was a dominant legacy failure mode. Federation cannot be added afterward if
  the base model has no place to record which participant owns what.
- **Requirement:** Base identity, resource, catalog, metering, and billing
  schemas MUST carry an explicit participant-scope attribute (a stable
  installation/federation identifier) from the first release, even when only
  one installation exists. Internal APIs and event envelopes MUST propagate
  this scope. Hardcoding a single implicit provider identity anywhere in core
  schemas, token claims, or meter records MUST be treated as a defect class.
  Core code SHOULD remain fully functional as a standalone single-installation
  deployment with a locally generated participant identity; no federation
  machinery is required to run standalone.
- **Acceptance evidence:** Go schema contract tests asserting participant-scope
  fields exist and are required in base identity/catalog/metering/billing
  record schemas; event-envelope tests showing scope propagation; a source-scan
  check that fails on single-provider sentinel values in core schemas; a
  standalone-installation conformance test proving the scope attribute defaults
  locally without any federation configuration.
- **Non-goals:** implementing federation sync, trust, or settlement machinery;
  requiring standalone installations to register anywhere; prescribing one
  specific identifier scheme beyond uniqueness and stability requirements.
- **Non-claims:** no multi-participant deployment has ever been run; the scope
  model's sufficiency for real federation is unproven and remains to be
  validated by the federation waves.
- **Stop conditions:** any change that would remove, optionalize, or hardcode
  the participant scope in base schemas MUST halt and escalate to owner review
  — it silently forecloses migration and federation (migration/trust risk).
  Discovery of persisted tenant or meter data lacking participant scope blocks
  all federation-related claims until remediated with a data-migration plan.
- **Traceability:** legacy-platform-a (single-tenant assumptions flagged as the
  top federation blocker), vision-deck (participant taxonomy), current-core
  (identity/resource model conventions). Related: domains/15-iam-identity-security.md,
  domains/16-billing-finops.md.

### CR-FED-020 — Federation honesty boundary
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, agent, auditor
- **Problem:** The federation vision (P2P bus, global portal, settlement) is
  the most marketable and least proven part of the platform. Presenting intent
  as capability — in the portal, catalog, documentation, or agent answers —
  would violate the evidence-over-claims principle and mislead adopters into
  building on surfaces that do not exist.
- **Requirement:** Every federation capability MUST carry machine-readable
  maturity metadata (declared, preview, verified, blocked) and MUST NOT be
  presented as available, ready to serve production workloads, or default-enabled without linked
  verified evidence. Federation features MUST default to disabled or hidden in
  portal and catalog surfaces until their evidence gates pass. Documentation
  and agent-facing capability answers MUST distinguish "designed/declared"
  from "verified with live evidence"; `blocked` is a first-class reportable
  state and MUST NOT be laundered into readiness. Claims in release notes,
  marketplace listings, and portal UI MUST pass the same evidence gate.
- **Acceptance evidence:** capability-metadata contract checks (maturity fields
  present and enum-locked); a CI source-safety-class scan over docs, portal,
  and catalog artifacts rejecting ungated federation readiness claims;
  portal/catalog rendering tests proving federation surfaces are hidden below
  the maturity threshold; an audit trail showing any maturity promotion links
  verified evidence.
- **Non-goals:** forbidding design documents, roadmap statements, or
  declared-status listings; requiring the federation itself to exist before
  the honesty metadata contract ships.
- **Non-claims:** no federation capability is claimed verified today; all
  content in this domain file is proposed intent, and most of it has no live
  evidence of any kind.
- **Stop conditions:** any release, listing, or portal change that would expose
  a federation capability as ready without verified evidence MUST halt at the
  gate (exposure/trust risk); a discovered ungated claim is treated as a
  release-blocking defect.
- **Traceability:** vision-deck (federation vision scope), current-core
  (evidence-freshness contract, capability maturity conventions),
  req-acr-plural (false-readiness risk class). Related:
  domains/18-marketplace-catalog.md, domains/20-observability.md.

### CR-FED-030 — Federation metadata in connector packages and listings
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, vendor, provider, agent
- **Problem:** Tenants cannot make placement or residency decisions if a
  service's federation behavior — where it can run, which cross-provider
  scenarios it supports, what its portability policy is — stays implicit.
  Legacy marketplaces hid unsupported modes behind generic listings, and the
  failure was discovered only after purchase or partition.
- **Requirement:** Every OCS connector package MUST declare a federation
  profile: supported participation modes (public provider, private presence,
  edge connected, edge disconnected, degraded, or explicitly unsupported),
  the cross-provider scenarios it supports (DR, replication, backup, CDN,
  migration), and a portability policy reference. Catalog and marketplace
  listings MUST surface this profile verbatim to tenant, operator, and agent,
  including explicit unsupported modes. A listing that omits or contradicts
  the declared federation profile MUST fail validation or publication. The
  profile SHOULD be declarable per deployment target where behavior differs
  by mode.
- **Acceptance evidence:** `ocsctl validate`/`conformance` checks covering the
  federation profile (presence, mode enum, scenario references, portability
  policy reference); a publication-gate test proving a listing with hidden or
  contradictory mode support is rejected; synthetic fixtures demonstrating
  valid and invalid profiles; rendered-listing tests showing modes visible to
  tenant and agent.
- **Non-goals:** requiring every service to support federation modes (an
  explicit `unsupported` declaration is valid and respected); defining the
  sync or settlement machinery here.
- **Non-claims:** the federation profile is declared metadata only; it does
  not prove any mode actually works for a given service — live per-mode
  evidence is a separate gate that no service has passed.
- **Stop conditions:** a service whose real behavior contradicts its declared
  federation profile (for example, one that dies without central connectivity
  while declaring disconnected support) MUST be suspended from federated
  listings pending owner review (trust/exposure risk).
- **Traceability:** current-core (federation profile in connector packages,
  16-surface conformance), req-acr-plural (mode visibility failure modes),
  vision-deck (edge and federation modes). Related:
  domains/17-ocs-service-connectors.md, domains/18-marketplace-catalog.md.

### CR-FED-040 — Federation participant registry contract
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, vendor, auditor
- **Problem:** Cross-provider scenarios need a shared, verifiable answer to
  "who are the participants, where do they operate, what do they offer, and on
  what terms" — without introducing a single trusted operator that every
  participant must depend on and trust.
- **Requirement:** The platform MUST define a versioned participant-registry
  contract: each participant record carries a stable participant identifier,
  public key material for verification, jurisdiction declarations,
  catalog/availability/pricing metadata classes, capability classes, and
  freshness timestamps. Registry records MUST be signed by the issuing
  participant and verifiable offline by any other participant. The contract
  MUST define update, revocation, and suspension semantics, including how
  consumers treat stale or revoked records. The registry SHOULD be replicable
  between participants so no single registry operator is required for reads.
- **Acceptance evidence:** a versioned contract document with a machine-readable
  schema (md+json pair) and a Go verifier; synthetic participant fixtures
  (valid, expired, revoked, wrong-signature) proving fail-closed verification;
  a replication/consistency test between at least two synthetic registry
  copies; freshness-window checks with stale-record blocking tests.
- **Non-goals:** a global consensus protocol or ledger product; a centrally
  operated registry service inside the OSS core; storing tenant-identifying
  data in participant records.
- **Non-claims:** no production participant registry exists; revocation and
  suspension semantics are designed but have never been exercised under real
  multi-party operation.
- **Stop conditions:** signature verification failure, expired or revoked
  participant records, or registry data stale beyond its declared window MUST
  halt dependent operations (sync, placement, settlement) and surface a typed
  denial (trust/keys risk); ambiguous registry state MUST never fail open.
- **Traceability:** vision-deck (federation service and participant registry),
  legacy-platform-a (registry design: public key, jurisdiction, catalog,
  pricing), req-acr-plural (federation roles and authority scope). Related:
  domains/15-iam-identity-security.md.

### CR-FED-050 — Federated identity trust
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator, auditor
- **Problem:** Cross-provider use requires a principal authenticated by one
  participant to be authorized — with least privilege and full audit — at
  another participant, without either side trusting the other blindly and
  without a universal identity operator.
- **Requirement:** The platform MUST define a federated-trust contract
  supporting multiple independent trust anchors (OIDC-class federation or
  decentralized-identifier-class schemes), where each participant decides
  which anchors it accepts. Cross-provider assertions MUST be validated
  against the issuing anchor's published keys with rotation overlap, and
  validation MUST fail closed on unknown anchors, unknown keys, expired
  assertions, or discovery failure. Identity mappings across providers MUST
  preserve ownership and audit continuity — who acted, on whose behalf, under
  which anchor — without sharing secret material between participants. Tenant
  consent MUST be explicit before their identity or workloads are exposed to
  another participant. The trust-framework contract is stable; the mechanisms
  behind it SHOULD be replaceable.
- **Acceptance evidence:** a conformance test suite with synthetic multi-anchor
  identity providers covering cross-participant allow, cross-participant deny,
  unknown-anchor deny, unknown-key deny during rotation, and discovery-failure
  deny; audit-envelope checks proving anchor and represented-subject
  attribution; a key-rotation drill with overlap window and unknown-key
  rejection evidence; tenant consent-flow tests.
- **Non-goals:** a single global identity provider; merging participants'
  internal IAM domains; automatic trust between all participants by default.
- **Non-claims:** no cross-provider identity flow has been exercised on a live
  multi-participant installation; decentralized-identifier-class anchors are
  unexplored; all evidence to date is synthetic.
- **Stop conditions:** any assertion-validation ambiguity (unknown anchor,
  unresolvable keys, skew beyond policy, missing consent record) MUST deny and
  alert (keys/trust/exposure risk); a discovered path where a foreign
  principal obtains unmapped or elevated rights halts federation enrollment
  until fixed and re-evidenced.
- **Traceability:** vision-deck (secure utilization without necessitating
  trust), legacy-platform-a (multi-trust-anchor federation design),
  req-acr-plural (identity domains, fail-closed trust boundaries). Related:
  domains/15-iam-identity-security.md.

### CR-FED-060 — Federation governance: admission, suspension, disputes, certification
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, vendor, auditor
- **Problem:** A federation without published rules for joining, behaving,
  being suspended, and resolving disputes becomes either centralized control
  in disguise or ungoverned risk; rules invented after the first incident are
  too late for the participants already harmed.
- **Requirement:** Before any federated listing or cross-provider scenario is
  published, the platform MUST publish a versioned federation-governance
  policy: participant admission criteria, good-standing obligations (evidence
  freshness, security posture, data handling), suspension and revocation
  procedures, dispute and escalation paths, compatibility certification
  requirements, and non-discrimination terms for equal-rights participants.
  Governance changes MUST be versioned, dated, and reviewable; the publication
  pipeline MUST block federation-visible features when the governance policy
  is absent or stale. The policy SHOULD define how governance decisions are
  themselves evidenced and appealed, and MUST NOT grant any single participant
  unilateral authority over another participant's operations.
- **Acceptance evidence:** a versioned governance policy document with
  machine-checkable effective dates; a pipeline-gate test proving federated
  listings and scenarios are blocked without a current policy; a synthetic
  certification checklist executed against a candidate participant; a recorded
  dispute-path walkthrough.
- **Non-goals:** operating a legal arbitration service; encoding
  jurisdiction-specific law into the OSS core (the policy declares hooks and
  obligations; counsel maps law to policy); building governance machinery
  beyond the publication gate at this stage.
- **Non-claims:** the governance policy has never been tested by a real
  dispute, suspension, or hostile participant; certification criteria are
  initial and unvalidated by third-party participants.
- **Stop conditions:** a governance violation (stale evidence, broken
  obligations, disputed participant) MUST suspend the affected federation
  surfaces until resolved per policy (trust/exposure risk); publication of
  federation features with an absent or expired governance policy MUST halt at
  the gate.
- **Traceability:** vision-deck (equal federation, non-discriminatory terms),
  req-acr-plural (suspension, disputes, certification governance before
  publication), legacy-platform-a (governance track of the federation design).
  Related: domains/18-marketplace-catalog.md.

### CR-FED-070 — Bilateral cross-provider data synchronization
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, agent
- **Problem:** The first workable federation step is pairwise: two
  installations exchanging catalog, availability, and pricing data under an
  explicit agreement, with freshness and conflict behavior defined — before
  any network-wide bus exists or is needed.
- **Requirement:** The platform MUST define and implement a bilateral
  synchronization contract between two participant installations covering
  catalog entries, availability/capacity classes, and pricing/terms metadata.
  Synchronized records MUST be signed by the source participant, carry
  freshness timestamps, and follow declared conflict-resolution semantics
  including partition behavior. Consumers MUST display record age and MUST
  NOT act on records stale beyond the declared window. Synchronization MUST be
  opt-in per participant pair and per data class, revocable, and auditable;
  tenant-identifying data MUST NOT flow through this channel.
- **Acceptance evidence:** two-installation lab integration evidence showing
  catalog/availability/pricing records syncing with verified signatures;
  freshness-window enforcement tests (stale records blocked from placement
  decisions); a partition-and-reconnect drill demonstrating declared conflict
  behavior; a revocation test (sync disabled, records expire); an audit log of
  sync agreements per participant pair.
- **Non-goals:** multi-party mesh topologies, global consistency, or real-time
  pricing; tenant workload data replication (that belongs to the cross-cloud
  scenario contracts, not the metadata channel).
- **Non-claims:** bilateral sync has never run between independently operated
  organizations; conflict semantics under real partitions are drilled only in
  lab conditions.
- **Stop conditions:** signature failures, schema-incompatible records, or
  conflicts beyond declared resolution semantics MUST halt the affected sync
  classes and alert both operators (data/trust risk); a detected flow of
  tenant-identifying data over this channel MUST suspend the pair agreement
  and escalate.
- **Traceability:** vision-deck (data synchronization across providers),
  legacy-platform-a (pairwise federation staged before the P2P bus),
  req-acr-plural (shared contracts with freshness and conflict behavior).
  Related: CR-FED-120, domains/18-marketplace-catalog.md.

### CR-FED-080 — Tenant migration between providers
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator, agent
- **Problem:** The anti-lock-in promise is only real if a tenant can move
  workloads, data, and configuration from one provider's installation to
  another's and keep operating — with both sides able to prove what happened
  and neither side able to hold the tenant's data hostage.
- **Requirement:** The platform MUST define a cross-provider tenant-migration
  contract: a portable migration package (data, configuration, policies,
  identity mappings, billing/usage history, recovery instructions), a
  migration state machine (plan, dry-run, execute, verify, commit,
  decommission), and evidence at every step. Migrations MUST be dry-runnable
  with a compatibility report (capability gaps, quota differences, version
  windows) before any mutation. The receiving provider MUST verify package
  integrity and completeness before import; the source provider MUST NOT
  delete tenant data until the tenant confirms success or a declared retention
  window expires, and deletion MUST then produce evidence. Billing continuity
  (final source invoice, receiving-side entitlement start) MUST be
  unambiguous. Migrations touching jurisdictions MUST pass the jurisdiction
  policy gate (CR-FED-090).
- **Acceptance evidence:** a rehearsal drill evidence chain — export from
  installation A, compatibility dry-run report, import to installation B,
  data/identity/billing continuity verification, controlled decommission with
  deletion evidence; package integrity/completeness checker output;
  failure-injection tests (interrupted transfer, partial import) showing
  rollback and retry behavior; tenant-visible migration status and evidence
  bundle.
- **Non-goals:** live zero-downtime migration of running workloads (cutover
  windows are declared per service); migration of services that declare the
  scenario unsupported; automated decommission without tenant confirmation.
- **Non-claims:** no tenant has ever migrated between independently operated
  CloudRING providers; cross-provider migration evidence does not exist — all
  drills so far are same-lab installations.
- **Stop conditions:** failed integrity or completeness verification, a
  missing backup barrier before decommission, an ambiguous billing handoff, or
  a jurisdiction-gate denial MUST halt the migration with state and evidence
  preserved (migration/data/deletion risk); source-side deletion without
  tenant confirmation or retention-window expiry is forbidden outright.
- **Traceability:** vision-deck (free migration between providers),
  req-acr-plural (exit and recovery scenario, migration replay lessons),
  legacy-platform-a (cut-over governance lessons). Related:
  domains/13-storage-backup-dr.md, CR-FED-090.

### CR-FED-090 — Jurisdiction portability and exit evidence
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, auditor, agent
- **Problem:** Jurisdiction is a founding product parameter: tenants must know
  where their data, backups, logs, identity, and billing records legally and
  physically live, must be warned before an action moves anything across a
  boundary, and must be able to prove an exit or move happened completely.
- **Requirement:** Jurisdiction and residency attributes MUST be recorded on
  tenants, workloads, data sets, backups, identity providers, billing records,
  and support events, and MUST be queryable by tenant, operator, and agent.
  Any action that would move data or metadata across a jurisdiction boundary
  (cross-provider replication, migration, backup restore elsewhere, support
  access from another region) MUST surface the jurisdiction impact before
  execution and MUST fail closed when policy disallows it. A tenant exit or
  jurisdiction move MUST produce a completeness-and-residency evidence bundle:
  what moved, what remained, what was deleted, with verification proof. The
  platform MUST distinguish at least: technical region, legal data-controller
  zone, backup location, and log-processing location.
- **Acceptance evidence:** attribute-propagation tests across all listed object
  classes; policy-gate tests (cross-boundary action without an explicit policy
  allow produces a typed denial naming the missing precondition); an
  exit-evidence bundle checker (completeness across data, configs, policies,
  audit summary, billing, recovery instructions); a restore/move drill proving
  residency attributes land correctly at the destination; an auditor-visible
  jurisdiction report.
- **Non-goals:** encoding specific national regulations into the OSS core
  (policies are data; counsel maps law to policy); guaranteeing that a
  receiving jurisdiction is legally suitable — the platform evidences facts,
  not legal advice.
- **Non-claims:** jurisdiction modeling is unvalidated by any real regulatory
  review; exit bundles have never been exercised in a genuine provider dispute
  or regulator request.
- **Stop conditions:** missing or conflicting jurisdiction attributes on
  regulated objects, or any cross-boundary transfer whose policy evaluation
  errors, MUST block the action and alert (data/exposure/migration risk);
  discovery of a backup or log copy in an undeclared jurisdiction MUST halt
  new transfers until remediated and re-evidenced.
- **Traceability:** vision-deck (jurisdiction independence), req-acr-plural
  (jurisdiction attributes, residency fail-closed, exit package), current-core
  (portability declarations). Related: domains/13-storage-backup-dr.md,
  domains/15-iam-identity-security.md.

### CR-FED-100 — EDGE zones: disconnected and autonomous operation
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Edge and private presences must keep serving tenants when cut
  off from any central or federation control plane; an edge installation that
  dies without center connectivity is a product defect, not a network event.
- **Requirement:** An EDGE zone MUST operate autonomously when disconnected:
  local catalog snapshot, local identity and authorization decisions within
  cached policy, local metering and event spooling, local support and
  diagnostics, and local audit — with explicit per-surface behavior for
  catalog freshness, billing continuity, support handoff, updates, and audit
  while partitioned. The zone MUST declare its disconnected state honestly in
  all local and (on reconnect) federated surfaces, including data staleness.
  Disconnected operation MUST NOT silently expand local privileges; what was
  denied before the partition stays denied. Reconnection MUST trigger the
  deferred-sync reconciliation contract (CR-FED-110).
- **Acceptance evidence:** a partition drill evidence chain — the edge zone
  runs a defined workload set for a declared window with no center connection,
  with catalog/billing/support/update/audit behavior recorded per surface;
  privilege-escalation tests during partition proving no new rights are
  granted; staleness-labeling checks; reconnect drill entry criteria met.
- **Non-goals:** unlimited-duration autonomy (declared maximum windows per
  surface); disconnected enrollment of brand-new tenants or services that
  require center-issued trust material; feature parity with connected mode.
- **Non-claims:** the maximum safe partition duration is undetermined;
  disconnected-mode billing accuracy under long partitions is unproven; no
  real telco-edge deployment exists.
- **Stop conditions:** cached trust or policy material expiring during a
  partition MUST degrade to deny, never permit (trust/keys risk); local spool
  exhaustion or metering gaps beyond declared tolerance MUST halt billable
  mutation classes and alert (money risk); any audit gap detected on reconnect
  MUST block re-admission to federation sync until resolved.
- **Traceability:** vision-deck (disconnected edge mode), req-acr-plural
  (per-surface disconnected behavior), legacy-platform-a (edge modes in the
  federation roadmap). Related: CR-FED-110, domains/16-billing-finops.md.

### CR-FED-110 — EDGE zones: connected mode and deferred sync reconciliation
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** A connected edge consumes provider and federation services and
  must eventually reconcile everything that happened while partitioned —
  usage events, audit records, catalog changes — without double-counting or
  losing money-bearing or safety-bearing records.
- **Requirement:** Connected EDGE mode MUST declare which center and federation
  services it consumes (catalog sync, updates, marketplace, support
  escalation, federation scenarios) and MUST fail per-surface to its
  disconnected behavior when those dependencies lapse. All deferred
  synchronization MUST be idempotent and reconcilable: usage events
  deduplicate by idempotency key, audit streams merge with ordering and tamper
  checks, catalog updates apply with version and freshness rules, and
  conflicts surface as typed states rather than silent overwrites.
  Reconciliation after a partition MUST produce a reconciliation report —
  applied, deduplicated, conflicted, and rejected records — visible to
  operator and auditor. Money-bearing records MUST reconcile exactly-once or
  not at all.
- **Acceptance evidence:** a deferred-sync reconciliation test suite (replayed
  usage events deduplicate by idempotency key with zero double-charge in
  billing-ledger tests); a partition-then-reconnect drill with a synthetic
  event backlog and a complete reconciliation report; conflict-fixture tests
  proving typed surfacing; audit-merge verification covering ordering,
  signatures, and gap detection.
- **Non-goals:** exactly-once semantics for non-ledger convenience data
  (declared per class); hiding reconciliation latency from users.
- **Non-claims:** reconciliation correctness is proven only against synthetic
  backlogs; real multi-week partition backlogs are untested; connected-mode
  economics (which provider services an edge may consume or resell) are
  unsettled.
- **Stop conditions:** any unreconcilable or duplicated money-bearing record
  MUST halt settlement for the affected period and escalate
  (money/settlement risk); audit-stream gaps or signature failures during
  merge MUST block the edge's return to good standing until investigated
  (trust risk).
- **Traceability:** vision-deck (connected edge mode consuming provider
  services), req-acr-plural (delayed-sync mode, irreconcilable-usage failure
  mode), legacy-platform-a (metering idempotency lessons). Related:
  CR-FED-100, domains/16-billing-finops.md.

### CR-FED-120 — P2P cross-provider data bus
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, operator, vendor
- **Problem:** The network-wide vision needs many participants synchronizing
  catalog, availability, and pricing without central brokers or a pairwise
  explosion of agreements — the deepest and least proven technical element of
  the federation.
- **Requirement:** The platform SHOULD define a peer-to-peer data-bus contract
  for cross-participant synchronization: peer discovery and admission under
  the governance policy, signed and tamper-evident records, per-data-class
  topics with freshness and conflict semantics inherited from the bilateral
  contract (CR-FED-070), and replay/audit capability. The bus MUST NOT require
  a universal trusted operator for reads or writes; availability degradation
  MUST be local and typed, never silent. The contract SHOULD allow staged
  transport realizations (federated log replication first, mesh protocols
  later) behind a stable record and verification contract. Publication of any
  record class onto the bus MUST pass source-safety and tenant-data rules —
  metadata only.
- **Acceptance evidence:** a bus conformance suite (record signing and
  verification, tamper evidence, freshness/conflict semantics, replay); lab
  topology evidence with at least three participants, admission under the
  governance policy, and partition/degradation drills; a staged-transport
  contract test proving record-format stability across two realizations;
  topic-access policy tests (non-member participant denied).
- **Non-goals:** global Byzantine-consensus ordering for all record classes
  (only settlement-grade records need stronger guarantees, per CR-FED-170);
  tenant workload traffic on the bus; replacing bilateral agreements where law
  or policy requires them.
- **Non-claims:** the P2P bus is entirely unproven beyond design; there is no
  multi-participant deployment, no production discovery or admission, and no
  scale or adversarial testing; this requirement is expected to be revised
  after the first real topology.
- **Stop conditions:** signature or provenance failures, non-member writes, or
  propagation of tenant-identifying data MUST halt the affected topics and
  escalate (data/trust/exposure risk); discovery or admission anomalies
  (Sybil-class suspicion) MUST suspend new-peer admission pending governance
  review.
- **Traceability:** vision-deck (peer-to-peer data bus), legacy-platform-a
  (bus options analysis with staged-transport recommendation), req-acr-plural
  (no universal trusted operator, freshness and conflict behavior). Related:
  CR-FED-070, CR-FED-170.

### CR-FED-130 — Cross-cloud connect interconnect
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Cross-provider DR, replication, backup, and CDN scenarios need
  actual private connectivity between participant installations and tenant
  workloads spanning them — encrypted, authenticated, and replaceable at the
  mechanism level.
- **Requirement:** The platform SHOULD define a cross-cloud-connect contract:
  encrypted, mutually authenticated interconnect between participant
  installations (mesh tunneling, dedicated links, or provider peering as
  replaceable realizations), per-tenant isolation on shared links, optional
  tenant-level VPN-style access, and declared bandwidth/latency capability
  classes usable for placement. Connection establishment MUST require explicit
  bilateral agreement records (CR-FED-070) and MUST fail closed on untrusted
  peers. Key management for interconnects MUST use reference-only secrets with
  rotation evidence. The data-plane realization MUST NOT leak
  provider-specific assumptions into the contract.
- **Acceptance evidence:** two-installation interconnect lab evidence (mutual
  authentication, per-tenant isolation tests with cross-tenant traffic denial,
  a documented bandwidth/latency measurement method); untrusted-peer
  connection-refusal tests; a key-rotation drill with overlap and no plaintext
  material in logs or evidence; a contract-versus-realization substitution
  test (two different tunnel realizations pass the same contract suite).
- **Non-goals:** building a global backbone; guaranteeing internet-routable
  performance; mandating one tunnel technology; interconnect for participants
  without a bilateral agreement.
- **Non-claims:** interconnect has never been operated between independent
  organizations; performance classes are measurement methods, not achieved
  numbers; telco-edge interconnect is unexplored.
- **Stop conditions:** peer-authentication failure, key-material handling
  violations, or cross-tenant traffic leakage on shared links MUST tear down
  the affected interconnect and alert (keys/exposure/data risk); discovery of
  provider-specific assumptions hardwired into the contract blocks publication
  until refactored.
- **Traceability:** vision-deck (cross-cloud connect component),
  legacy-platform-a (mesh interconnect design with shared-LB and routing
  options), req-acr-plural (cross-provider party boundaries). Related:
  CR-FED-140, domains/12-network.md.

### CR-FED-140 — Cross-provider resilience scenarios: DR, replication, backup, CDN
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Disaster recovery, data replication, backup to another
  provider, and content delivery across participants are the headline
  federation scenarios — and the most dangerous to oversell, because they move
  tenant data across trust and jurisdiction boundaries.
- **Requirement:** Each cross-provider scenario (DR/failover, replication,
  cross-provider backup, CDN-class delivery) MUST be published as a versioned
  scenario contract before being offered: party boundaries and
  responsibilities, data classes in scope, jurisdiction gates (CR-FED-090),
  consistency and freshness expectations, ownership and cost attribution,
  failure and rollback semantics, and the evidence classes that prove the
  scenario works. A scenario MUST be offered only between participants whose
  declared capabilities and governance standing satisfy the contract, and only
  for services whose connector federation profile (CR-FED-030) declares it.
  Live scenario claims MUST be backed by drill evidence per scenario class —
  failover drill, restore drill, replication-consistency check, invalidation
  test — within freshness windows.
- **Acceptance evidence:** per-scenario contract documents with
  machine-checkable gates; drill evidence per class on at least two
  participants (a DR failover drill with a documented RTO/RPO measurement
  method, a cross-provider backup restore drill, replication-consistency
  checker output, a CDN invalidation test); capability and eligibility gate
  tests (ineligible pair or undeclared service produces a typed refusal); a
  rollback drill for a failed failover.
- **Non-goals:** promising uniform RTO/RPO across arbitrary participant pairs;
  automatic cross-provider failover without tenant policy; content-delivery
  peering economics (settlement handles money separately).
- **Non-claims:** no scenario has ever run between independent providers; all
  drill evidence to date comes from same-organization installations;
  real-world RTO/RPO figures do not exist and MUST NOT be quoted.
- **Stop conditions:** a jurisdiction-gate denial, missing current drill
  evidence, a backup-restore verification failure, or a party-boundary dispute
  MUST halt the scenario for the affected pair and surface a blocked state
  (data/migration/exposure risk); a failover executed without fresh
  reverse-direction evidence MUST NOT be marked complete.
- **Traceability:** vision-deck (DR, replication, backup, CDN federation
  scenarios), req-acr-plural (party boundaries and ownership evidence),
  legacy-platform-a (cross-provider scenario staging). Related:
  domains/13-storage-backup-dr.md, CR-FED-090, CR-FED-130.

### CR-FED-150 — Global cloud portal (white-label unified portal)
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, provider, vendor, operator
- **Problem:** The vision's capstone is one entry point where a user sees all
  participating public clouds, their own private clouds and edges, compares
  terms and prices, selects services, and pays the chosen provider — while
  each partner provider presents the portal under its own brand.
- **Requirement:** The platform SHOULD provide a global-portal contract and
  reference implementation: unified aggregation of participant catalogs (via
  CR-FED-070 and CR-FED-120 data), white-label theming with branding isolation
  per partner provider, unified API parity with the UI, the tenant's own
  private and edge presences shown alongside public ones, and unified
  cross-provider monitoring visibility within tenant rights. White-label
  deployments MUST NOT leak one partner's branding, tenants, or commercial
  terms into another's view. All federation content in the portal MUST obey
  the honesty boundary (CR-FED-020): stale or unverified data is labeled,
  never presented as fact. Payment flows MUST route to the tenant-selected
  provider only; settlement between providers is a separate concern
  (CR-FED-170).
- **Acceptance evidence:** portal contract tests — catalog aggregation from at
  least two synthetic participants with per-record freshness labels; a
  branding-isolation test suite (no cross-partner leakage in rendered output
  or API); API/UI parity checks; rights-scoped monitoring-visibility tests
  (tenant sees only their own cross-provider estates); a payment-routing test
  proving the charge lands at the selected provider.
- **Non-goals:** replacing per-provider portals (partners MAY embed or
  white-label); a global account system (identity stays federated per
  CR-FED-050); price negotiation or contracting flows.
- **Non-claims:** the global portal is unbuilt beyond contract scope;
  white-label licensing terms are a Business-layer concern and undefined;
  unified cross-provider monitoring has no live data source yet.
- **Stop conditions:** any cross-partner data or branding leakage, or
  presentation of stale prices and terms as current, MUST take the affected
  portal surfaces offline or into labeled-degraded mode
  (exposure/money/trust risk); payment-routing ambiguity MUST halt checkout
  and alert.
- **Traceability:** vision-deck (global cloud portal, white-label licensing,
  unified monitoring), req-acr-plural (marketplace selection factors),
  legacy-platform-a (portal lessons: metadata-driven, no hardcoded catalog).
  Related: domains/19-portal-ux-selfservice.md,
  domains/18-marketplace-catalog.md, CR-FED-160.

### CR-FED-160 — Requirement-based multi-provider placement
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, agent, provider
- **Problem:** The portal's distinctive promise is search by requirements —
  "closest to me", "cheapest", residency-constrained — which is only
  trustworthy if placement factors, data freshness, and ranking methodology
  are explicit and checkable by the person relying on them.
- **Requirement:** Placement and search MUST accept structured requirement
  inputs (location/proximity, price ceiling, jurisdiction constraints,
  capability requirements, latency class, trust/standing threshold, SLA
  class) and MUST return results with the factors that determined ranking,
  per-factor data age, and declared gaps. Ranking methodology MUST be
  published and versioned; "cheapest" and "closest" claims MUST be computable
  and recomputable from the shown inputs. Results based on data stale beyond
  declared windows MUST be labeled or excluded. Placement MUST exclude or
  downrank participants failing governance standing (CR-FED-060) and MUST
  respect tenant jurisdiction policy (CR-FED-090). Sponsored or preferential
  ranking, if ever introduced, MUST be labeled as such.
- **Acceptance evidence:** a placement-engine test suite over synthetic
  multi-participant offers (proximity ranking, price ranking, jurisdiction
  filtering, trust-threshold exclusion); recompute-check tooling reproducing
  any shown ranking from its recorded inputs; stale-data guard tests; a
  versioned methodology document with change log; an audit trail of placement
  queries and shown factors.
- **Non-goals:** automated provisioning from search results (selection then
  proceeds through normal per-provider flows); real-time price negotiation;
  global capacity guarantees.
- **Non-claims:** the ranking methodology is unvalidated by real users;
  price and availability data sources are synthetic; no claim is made that
  cross-provider prices are comparable across currencies, tax regimes, or
  contract terms beyond the declared normalization.
- **Stop conditions:** undocumentable ranking influence, a missing methodology
  version, or stale data presented as fresh MUST halt placement features and
  alert (money/trust/exposure risk); jurisdiction-policy violations in results
  are release-blocking defects.
- **Traceability:** vision-deck (requirement-based search: closest, cheapest),
  req-acr-plural (selection factors: price, location, jurisdiction, latency,
  trust, capability, SLA), legacy-platform-a (public price calculator
  precedent). Related: CR-FED-150, CR-FED-090.

### CR-FED-170 — Cross-provider settlement hooks
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, vendor, operator, auditor
- **Problem:** In the federation model a tenant pays their selected provider,
  and that provider owes shares to others whose infrastructure or services
  were consumed; without a verifiable settlement trail this becomes
  unresolvable financial dispute material.
- **Requirement:** The OSS platform MUST define settlement interfaces — not a
  settlement implementation: signed, tamper-evident usage/exchange events
  carrying participant attribution, idempotency keys, and entitlement
  references; a reconciliation contract computing who owes whom from those
  events; and a settlement-adapter boundary so the terminal settlement or
  payment system is replaceable, with commercial settlement services plugging
  in behind it. All money-bearing events MUST reconcile exactly-once;
  discrepancies MUST surface as typed dispute states with evidence, never as
  silent adjustments. Settlement records MUST be retainable and auditable for
  declared periods, and MUST NOT contain tenant-identifying data beyond what
  the reconciliation contract requires.
- **Acceptance evidence:** a reconciliation test suite over signed synthetic
  event streams (multi-party attribution, exact-once under replay,
  dispute-state surfacing on divergence); an adapter-boundary conformance
  suite proving the terminal settlement system is replaceable (two mock
  adapters pass the same suite); retention and audit-export checks;
  event-schema validation covering signatures, idempotency, and attribution
  fields.
- **Non-goals:** moving money, invoicing, tax handling, or operating a
  settlement service (Business-layer scope); a blockchain or ledger product
  (optional tamper evidence MAY use hash-chaining, but no consensus product is
  required); real-time settlement.
- **Non-claims:** no settlement has ever been computed or paid between
  participants; reconciliation correctness is synthetic-only; currency, tax,
  and regulatory treatment is undefined and explicitly outside the OSS claim.
- **Stop conditions:** reconciliation divergence, unsigned or duplicated
  money-bearing events, or retention-policy violations MUST halt settlement
  runs and escalate to the providers' operators (money/settlement risk); any
  adapter attempting to embed provider lock-in into the boundary — an
  irreplaceable terminal system — blocks certification.
- **Traceability:** vision-deck (pay the selected provider, who settles with
  the others), legacy-platform-a (settlement over signed events; the
  file-drop money-sink lesson), req-acr-plural (settlement-aware billing).
  Related: domains/16-billing-finops.md, CR-FED-180.

### CR-FED-180 — Revenue sharing and cross-licensing interfaces
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, vendor, operator
- **Problem:** The federation economy depends on two policy-level mechanisms:
  revenue sharing among participants (and toward the platform's commercial
  layer) and cross-licensing that lets one provider sell another's services —
  both must be declarable and auditable without hardcoding one business model
  into the core.
- **Requirement:** The platform MUST define declarative commercial hooks:
  per-service and per-component revenue-share terms as data (with
  volume-based reduction of the share, potentially to zero, expressible as
  policy), and cross-licensing grants between participants (which participant
  may sell which service, under what license class and expiry behavior).
  License-expiry behavior MUST follow the platform-wide rule: functionality
  preserved, maintenance and updates stop except critical compatibility fixes.
  Revenue-share computations MUST be derivable from settlement events
  (CR-FED-170) and recomputable for audit. Terms changes MUST be versioned,
  dated, and never retroactive without explicit counterparty acceptance. The
  OSS core SHOULD ship only the hooks and policy schema; commercial terms
  themselves are Business-layer content.
- **Acceptance evidence:** revenue-share computation tests from synthetic
  settlement events (a tiered reduction-to-zero policy expressed and
  recomputed); cross-license entitlement tests (grant, scope, and expiry
  behavior: functionality preserved with update-channel gating and the
  critical-fix carve-out); terms-versioning and non-retroactivity checks;
  audit recompute tooling output matching original computations.
- **Non-goals:** setting actual revenue-share percentages or prices; legally
  enforcing contracts; building a licensing storefront (the marketplace domain
  handles purchase flows).
- **Non-claims:** no revenue share has ever been computed or paid;
  cross-licensing has never been exercised; the reduction-to-zero mechanics
  are expressible in the policy schema but economically unvalidated.
- **Stop conditions:** ambiguous or retroactive terms changes, license-state
  ambiguity (expired-but-running versus expired-and-blocked), or entitlement
  decisions without a verifiable license record MUST halt the affected
  commercial flows and escalate (money/settlement/trust risk); any expiry path
  that deletes tenant functionality or data is forbidden outright.
- **Traceability:** vision-deck (revenue sharing with reduction toward zero,
  cross-licensing, license-expiry policy), legacy-platform-a (license
  lifecycle design: signed licenses, channels, expiry semantics). Related:
  domains/16-billing-finops.md, domains/18-marketplace-catalog.md, CR-FED-170.

## Coverage notes

This domain deliberately defers:

- **Per-primitive portability contracts and the baseline tenant exit package**
  to the foundation and storage/backup/DR domains; FED owns only the
  cross-provider and jurisdiction face of portability.
- **Single-provider metering, rating, charging, invoicing, and pay-accounts**
  to the billing domain; FED owns only the cross-provider settlement hooks and
  commercial interfaces.
- **OCS connector base surfaces** (lifecycle, billing meters, readiness,
  durability) to the OCS domain; FED owns only the federation profile
  semantics declared inside connector packages.
- **IAM internals** (token issuance, JWKS handling, session policy, nine
  identity domains) to the identity domain; FED owns only multi-anchor trust
  between participants.
- **Portal shell mechanics** (identity/session ownership, module slots,
  microfrontend mounting) to the portal domain; FED owns only global
  aggregation and white-label isolation.
- **Network product surfaces** (tenant VPN products, load balancing, CDN as a
  sellable product) to the network domain; FED owns only the interconnect
  contract between participants.
- **Single-provider DR/backup** to the storage/backup/DR domain; FED owns only
  scenarios crossing participant boundaries.
- **The "global clouds exchange" concept** — named once in the vision topology
  without behavioral semantics — is intentionally not a requirement until a
  product definition exists.
- **Telco-edge specifics and endpoint-device participation** — vision-level
  mentions without behavior — are deferred pending definition.
- **Settlement implementation, actual revenue-share percentages, white-label
  licensing terms, and federation settlement services** are CloudRING
  Business commercial scope, outside this corpus per the charter; the OSS
  interfaces they plug into are specified here.
