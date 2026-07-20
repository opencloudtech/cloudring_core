# 16 — Billing & FinOps

Scope: the metering, rating, charging, and commercial-account backbone of the
platform. This domain covers the versioned usage-event schema; the
authenticated, idempotent meter ingest API; the staged mediation and enrichment
pipeline; the declarative SKU catalog (pricing formulas, unit conversion,
tiered price history, self-applying per-installation bundles); decimal-precise
deterministic rating; billing accounts (prepaid/postpaid, credit limits,
suspension timelines); grants and credits; advisory budgets; per-second VM
usage evidence; invoices and EDO-class billing documents; the OCS billing
connector surface; billing-vs-observed reconciliation evidence; the public
billing API; tenant cost visibility; and revenue-sharing readiness. It does
not cover payment acquiring, marketplace product packaging, federation
settlement execution, or analytics warehouses (see Coverage notes).

Domain contract: money paths fail closed. Usage is never silently dropped —
every rejection is dead-lettered with a reason and replayable at every pipeline
hop. All money arithmetic is fixed-point decimal with explicit rounding
boundaries; all pricing is catalog-driven data, changeable without code
deploys, and reproducible for any historical window. Budgets notify and never
throttle. Suspension and deletion follow declared, auditable timelines. No
billing-correctness or revenue claim may be made for a period without verified
reconciliation evidence. Metering is a launch gate: an unmetered service is an
unfinished service, and "free" is a catalog price of zero, not an absence of
measurement.

## Requirements

### CR-BIL-010 — Versioned usage-event schema
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, provider
- **Problem:** Without a single versioned metering contract, every service
  invents its own usage records. Rating then cannot be catalog-driven,
  producers and consumers drift silently, and reconciliation against observed
  usage becomes impossible.
- **Requirement:** The platform MUST define one versioned usage-event schema
  used by every usage-bearing service. Each event MUST carry: a unique event
  id; the schema version; tenant/resource-hierarchy ids (organization,
  project, resource); service and meter identity; quantity with unit; the
  usage window as UTC start/finish timestamps; the usage type (`delta` or
  `cumulative`); structured labels; and free-form tags. Consumers MUST reject
  invalid events into the dead-letter path (CR-BIL-020), never drop them
  silently. The pipeline MUST be able to convert `cumulative` events to deltas
  deterministically, handling producer restarts, counter resets, and migration
  timestamps.
- **Acceptance evidence:** published schema document; conformance test suite
  with valid/invalid event fixtures executed in CI; producer/consumer
  version-negotiation contract tests; cumulative-to-delta conversion tests
  covering restart, reset, and proration cases.
- **Non-goals:** per-service business semantics of individual meters; the
  choice of event-bus storage technology.
- **Non-claims:** no producer implementations are validated against the schema
  yet; the schema is not yet frozen as v1.
- **Stop conditions:** money / trust — if unversioned or unvalidated events
  are found entering the rating path, halt onboarding of new producers and
  reconcile affected windows before any billing-correctness claim.
- **Traceability:** vision-deck; legacy-platform-a; legacy-platform-b.
  Related: CR-BIL-020, CR-BIL-030, CR-BIL-130, observability domain (20).

### CR-BIL-020 — Meter ingest API: authenticated, idempotent, replayable, dead-lettered
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, provider, operator, auditor
- **Problem:** Usage events arrive from many services and installations;
  retries, restarts, and redeliveries are normal. Without idempotent ingest,
  duplicates double-charge tenants; without durable dead-lettering, failures
  silently lose revenue.
- **Requirement:** The platform MUST provide an authenticated ingest API for
  usage events, authorized per producer with scoped service credentials.
  Ingest MUST be idempotent on a producer-supplied idempotency key: a
  duplicate submission returns the original acceptance and never creates a
  second charge. Raw events MUST be durably persisted before acknowledgement.
  Events failing validation MUST land in a dead-letter destination with
  operator-visible counts and per-event failure reasons. Operators MUST be
  able to replay dead-lettered events and historical windows into the pipeline
  without manual data edits. A validation-only dry-run mode MUST be available
  for producer integration testing, with a published test meter/SKU pair.
- **Acceptance evidence:** duplicate-submission dedup integration tests;
  crash-recovery test proving no acknowledged event is lost; dead-letter
  replay drill evidence; dry-run mode contract test.
- **Non-goals:** the event bus implementation itself; rating and charging.
- **Non-claims:** no production-scale load or failure-injection evidence yet.
- **Stop conditions:** money / data / keys / exposure — halt and escalate if
  duplicates produce double-rated usage, if events are acknowledged before
  durable persistence, or if ingest accepts unauthenticated producers. Ingest
  credentials MUST come only from the approved secrets workflow; any credential
  found in configuration halts the pipeline.
- **Traceability:** legacy-platform-a; legacy-platform-b; current-core.
  Related: CR-BIL-010, CR-BIL-030, CR-BIL-130.

### CR-BIL-030 — Mediation and enrichment pipeline
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Raw usage events lack commercial context (billing account,
  catalog entry, price). Staged mediation with enrichment joins is required,
  and a failure mid-pipeline must be re-processable — never terminal for the
  event.
- **Requirement:** The platform MUST implement a staged money pipeline:
  durable raw store (mediation) -> enrichment (tenant, billing-account, and
  resource-catalog joins) -> catalog/tariff resolution -> rating -> charging ->
  report/invoice feed. Each stage MUST durably persist its inputs and outputs,
  MUST emit processing-lag and error metrics, and MUST support replay of
  failed events from that stage without re-ingesting upstream stages. Every
  rejection at any stage MUST land in a queryable dump with a machine-readable
  reason and a disposition workflow. There MUST be exactly one logical
  pipeline; parallel or shadow billing stacks are prohibited.
- **Acceptance evidence:** staged-replay drill (failures injected at each
  stage, replayed, output verified identical); lag/error metric contract
  checks; audit query demonstrating every rejected event carries a reason and
  a recorded disposition.
- **Non-goals:** business analytics and warehousing (deferred to data
  domains); real-time per-second end-to-end latency targets.
- **Non-claims:** pipeline stages are not yet implemented in the current
  codebase; no production lag or replay evidence exists.
- **Stop conditions:** money / data — halt onboarding of new producers if
  dead-letter volume grows without disposition, or if any stage acknowledges
  events it did not durably persist.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-010,
  CR-BIL-020, CR-BIL-140.

### CR-BIL-040 — Declarative SKU catalog: formulas, unit conversion, resolving rules
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, vendor, operator, service-team
- **Problem:** Pricing logic embedded in service code turns every price change
  into a deploy and every audit into a code review. Reference platforms
  converged on catalog-driven pricing because code-driven pricing does not
  scale past a handful of services.
- **Requirement:** The platform MUST define the SKU catalog as declarative,
  versioned data: service and SKU identities; customer-facing names
  (multi-language capable); reporting grouping; usage unit versus pricing unit
  with an explicit unit-conversion table; per-SKU resolving rules mapping
  usage events to SKUs (including null/empty/numeric/string matching
  semantics); and pricing formulas expressed in a documented expression
  dialect over usage quantities and tags. Formula evaluation MUST use the
  platform's decimal money arithmetic (CR-BIL-060). Catalog changes MUST be
  applicable without code deploys, versioned, and auditable. Every catalog
  version MUST ship with machine-run resolving test cases (usage event ->
  expected SKU and pricing quantity) that pass before the version may apply.
- **Acceptance evidence:** catalog conformance suite (schema validation plus
  resolving test cases executed per version); drill evidence of a SKU or
  formula change applied without any service redeploy; append-only catalog
  version history with UTC timestamps.
- **Non-goals:** marketplace product-catalog UX (MKT domain); storage of
  per-tenant negotiated pricing contracts.
- **Non-claims:** the expression dialect is not yet specified; no catalog
  version has shipped.
- **Stop conditions:** money — halt catalog rollout if any resolving test case
  fails, or if a new catalog version cannot reproduce prior-period rated
  output under replay.
- **Traceability:** legacy-platform-a; legacy-platform-b; current-core.
  Related: CR-BIL-050, CR-BIL-060, CR-BIL-130.

### CR-BIL-050 — Price history and per-installation price bundles
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, auditor
- **Problem:** Prices change over time and may differ per installation;
  sovereign or offline installations cannot pull reference data from a central
  origin. Rating must reproduce any historical invoice exactly as it was rated
  at the time.
- **Requirement:** Every SKU MUST carry price history with effective-from UTC
  timestamps and optional tiered rate tables (quantity threshold -> unit
  price). Rating MUST select price by usage-window timestamp, never by
  rating-time wall clock. Price and reference data MUST be distributable as
  signed, self-verifying, self-applying per-installation bundles that apply on
  installations without connectivity to the origin repository. Applying a
  bundle MUST be an auditable, reversible operator action with the applying
  identity, timestamp, and bundle digest recorded.
- **Acceptance evidence:** time-travel rating tests (re-rate historical
  windows -> charges identical to originals); bundle signature and
  self-verification tests; tier-boundary unit tests; offline-apply drill on an
  air-gapped installation profile.
- **Non-goals:** dynamic or real-time pricing; FX rate management
  (CR-BIL-190).
- **Non-claims:** bundle tooling is not yet built; no sovereign-installation
  apply evidence exists.
- **Stop conditions:** money / settlement — halt a price change if bundle
  verification fails, if installations that should share a price list diverge
  without an explicit per-installation override record, or if re-rating a
  historical window yields different charges than the originals.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-040,
  CR-BIL-140.

### CR-BIL-060 — Rating with fixed-point decimal precision
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, auditor, operator
- **Problem:** Binary floating-point money arithmetic accumulates rounding
  drift into real billing discrepancies and destroys trust in invoices.
- **Requirement:** All money and quantity arithmetic in the billing path MUST
  use fixed-point decimal arithmetic of at least decimal(35,15)-class
  precision, with explicit, documented rounding rules applied only at defined
  boundaries (per line item, per invoice total). Unit conversions
  (bytes-to-gibibytes, seconds-to-months, currency minor units) MUST be explicit,
  table-driven operations, not inline constants. Rating MUST be deterministic:
  identical inputs plus identical catalog version produce identical output.
- **Acceptance evidence:** property-based arithmetic test suite (rounding
  determinism, conversion round-trips, boundary behavior); golden-file rating
  tests; replay-determinism test proving identical output across runs.
- **Non-goals:** tax-engine internals; asset accounting outside cloud usage.
- **Non-claims:** no decimal library or expression dialect is selected yet.
- **Stop conditions:** money — halt release of any billing component whose
  tests show binary float usage in money paths or non-deterministic rating
  output.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-040,
  CR-BIL-050.

### CR-BIL-070 — Billing account model: prepaid/postpaid, separated from IAM
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Billing identity is commercial, not infrastructural. Conflating
  billing accounts with IAM principals breaks reseller, multi-project, and
  enterprise-contract scenarios, and couples money state to authentication
  state.
- **Requirement:** The platform MUST model billing accounts as first-class
  commercial entities separate from IAM principals: many projects/resources
  bind to one billing account; account types support prepaid balance and
  postpaid settlement with per-type payment-method metadata. Every charge
  MUST attribute to exactly one billing account and exactly one
  resource-hierarchy path. Account state transitions (active, suspended,
  closed) MUST be explicit, audited, and propagated to dependent domains
  (service lifecycle, console) through versioned events. The billing account
  MUST NOT be an IAM or resource-management entity.
- **Acceptance evidence:** schema and invariant tests (exactly-one-account
  attribution enforced); account state-machine tests; event contract tests for
  state propagation to lifecycle and console consumers.
- **Non-goals:** payment acquiring and card processing; legal-entity/KYC
  verification workflows (provider integrations).
- **Non-claims:** the model is not yet implemented; no production account
  lifecycle evidence exists.
- **Stop conditions:** money / trust / data — halt provisioning of billable
  resources if any resource lacks resolvable billing-account attribution;
  never auto-close or delete an account with unsettled balance or unexpired
  retention obligations.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: IAM domain
  (15), CR-BIL-100, CR-BIL-150.

### CR-BIL-080 — Grants and credits: expiry-ordered consumption, charge decomposition
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Commercial onboarding relies on credit instruments (trial
  grants, promotional credits, committed-use discounts, volume incentives).
  Tenants must see how each charge was decomposed, and credits must be
  consumed in a deterministic, fair order — never as an opaque balance pool.
- **Requirement:** The platform SHOULD support credit instruments as
  versioned, append-only ledger entries with expiry dates and applicability
  scopes. Rated usage MUST decompose each charge into payable cost versus
  credit consumption per instrument. Consumption MUST be ordered
  deterministically by earliest expiry first unless an instrument declares
  narrower applicability. Credits MUST NOT settle pre-existing arrears and
  MUST NOT convert to cash balance. All ledger entries MUST be append-only
  with UTC timestamps; corrections are new entries, never mutations.
- **Acceptance evidence:** ledger property tests (append-only; conservation:
  issued - consumed = remaining, per instrument); expiry-ordering test suite;
  decomposition golden tests on rated output; arrears-exclusion test.
- **Non-goals:** loyalty/points programs; cross-provider credit portability.
- **Non-claims:** the instrument set is not finalized; no production ledger
  evidence exists.
- **Stop conditions:** money / settlement — halt credit issuance if ledger
  conservation checks fail, or if credit consumption can be applied
  retroactively to closed periods.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-060,
  CR-BIL-100, CR-BIL-170.

### CR-BIL-090 — Budgets: advisory-only, notify never throttle
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Tenants need cost-control signals, but billing-triggered
  throttling causes billing-caused outages — a worse failure than overspend.
  Reference platforms converged on notify-only budgets as an explicit product
  decision.
- **Requirement:** The platform SHOULD provide budgets with amount types
  (usage-cost, amount-payable, balance), periods (month, quarter, year,
  custom), and multiple thresholds with per-threshold recipients. Crossing a
  threshold MUST produce a notification through the platform notification
  service and MUST NOT throttle, suspend, or delete any resource. Budget
  evaluation MUST count the full period even when the budget is created
  mid-period. Enforcement, where a provider wants it, belongs exclusively to
  the credit-limit path (CR-BIL-100) — never to budgets.
- **Acceptance evidence:** contract tests proving zero coupling from budget
  evaluation to lifecycle/throttle paths; notification integration tests;
  mid-period creation evaluation tests.
- **Non-goals:** hard quota enforcement (resource-quota domains); spend
  forecasting and anomaly detection.
- **Non-claims:** no notification-service integration evidence yet.
- **Stop conditions:** trust / money — halt and roll back any change that lets
  budget evaluation mutate resource state; treat any discovered
  budget->throttle coupling as a severity-1 defect.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-100,
  CR-BIL-150, CR-BIL-160.

### CR-BIL-100 — Postpaid credit limit, arrears, suspension, retention timeline
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator, auditor
- **Problem:** Postpaid tenants consume before paying. The provider needs a
  deterministic, contract-aligned path from negative balance to suspension to
  data handling — with no ad-hoc operator judgment in the loop.
- **Requirement:** The platform SHOULD support per-account credit limits
  (individually assigned, validity-bounded) defining how far a postpaid
  balance may go negative. Crossing the limit MUST trigger a declared,
  auditable sequence: arrears notification -> grace period -> service suspension
  (resources stopped, not deleted) -> data-retention countdown with a declared
  minimum retention window before any deletion. Every step MUST be
  event-sourced, reversible on settlement, and visible to the tenant.
  Suspension propagation to services MUST use the same lifecycle events as
  administrative suspension. Deletion after retention expiry MUST follow the
  storage domain's deletion-evidence requirements. Credits MUST NOT pay
  arrears (CR-BIL-080).
- **Acceptance evidence:** state-machine tests for the full
  arrears->suspension->restore path; drill evidence of settlement-triggered
  restore; retention-expiry deletion evidence with timestamps; audit-trail
  completeness check.
- **Non-goals:** debt collection and dunning payment retries; credit scoring.
- **Non-claims:** retention windows are policy placeholders pending legal
  review; no live suspension-propagation evidence exists.
- **Stop conditions:** money / data / deletion / trust — halt the suspension
  pipeline if any path deletes data before the declared retention window, if
  suspension skips the grace period, or if restore-after-settlement is not
  demonstrably complete.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related:
  storage/backup domain (13), IAM domain (15), CR-BIL-070, CR-BIL-080.

### CR-BIL-110 — Per-second VM usage heartbeat as metering evidence
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, service-team, tenant
- **Problem:** Per-second pay-as-you-go VM billing needs trustworthy run-state
  evidence. Coarse periodic snapshots cannot prove a VM ran for exactly N
  seconds; a fixed-cadence heartbeat is the accepted evidence mechanism.
- **Requirement:** The platform SHOULD meter running VMs at per-second
  granularity using a heartbeat evidence stream (guest agent and/or hypervisor
  state feed) with a declared cadence and documented gap-handling rules.
  Missing-heartbeat intervals MUST be resolved by a documented conservative
  policy (unproven intervals are not billed), never by assuming usage.
  Heartbeat-derived usage MUST be reconcilable against compute-domain
  power-state events under UTC clock discipline. The same mechanism SHOULD
  serve per-second license-sensitive image billing (marketplace PAYG
  products).
- **Acceptance evidence:** heartbeat-to-usage conversion tests including gap
  policies; reconciliation drill of the heartbeat stream against the compute
  power-state event log; clock-skew handling tests.
- **Non-goals:** per-second granularity for all resource types; guest-agent
  features beyond metering (owned by the compute domain).
- **Non-claims:** no guest agent exists in the current codebase; gap policy is
  unproven under network partition.
- **Stop conditions:** money / trust — if heartbeat evidence and power-state
  logs diverge beyond declared tolerance for a billing window, halt invoicing
  for affected accounts and reconcile before issuing charges.
- **Traceability:** vision-deck; legacy-platform-a; legacy-platform-b.
  Related: compute domain (11), CR-BIL-140, MKT domain (18).

### CR-BIL-120 — Invoices and billing documents (EDO-class)
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, auditor, operator
- **Problem:** Commercial tenants require formal documents (invoices, usage
  acts, receipts) with legal-grade integrity; regeneration must never alter
  issued history.
- **Requirement:** The platform SHOULD generate billing documents per billing
  account and period — invoices with usage summaries, payment documents, and
  closure documents — with immutable sequential numbering, content integrity
  (hash and version per issued document), and an electronic-document-exchange
  (EDO-class) interface for machine exchange with tenant accounting systems.
  Documents MUST be derived exclusively from rated and charged ledger data.
  Correcting a past period MUST issue a corrective document; an issued
  document MUST never be mutated. Document storage MUST be durable, auditable,
  and exportable by the tenant (portability principle).
- **Acceptance evidence:** document-generation golden tests from ledger
  fixtures; immutability and corrective-document flow tests; EDO-format
  conformance checks; tenant export drill.
- **Non-goals:** jurisdiction-specific fiscal integrations and e-invoicing
  mandates (provider adapters); tax computation engines.
- **Non-claims:** no EDO profile is chosen; numbering and content requirements
  vary by jurisdiction and are legally unverified.
- **Stop conditions:** money / settlement / data — halt document issuance if
  source ledger and document totals diverge; never reissue a mutated document
  under the same number.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-060,
  CR-BIL-140, CR-BIL-160.

### CR-BIL-130 — OCS billing connector surface: meters, rate-card evidence, replay dedup
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider, auditor
- **Problem:** Third-party OCS services must become billable without
  platform-team involvement. The connector contract's billing surface must
  declare what a service charges for — and prove it — or the revenue
  mechanism fails for the ecosystem's first-class citizens.
- **Requirement:** Every OCS connector package MUST declare its billing
  surface: usage meters (what is measured), cost meters (unit price per meter
  with a rate-card evidence reference), and billing events (idempotent, with
  idempotency keys, entitlement references, attribution, and replay policies).
  The platform validator MUST reject connector packages that reference unknown
  meters, declare unsafe billing linkage, or lack rate-card evidence for
  priced meters. Service-emitted usage MUST flow through the platform metering
  contract (CR-BIL-010, CR-BIL-020) with replay dedup: replayed events with
  seen idempotency keys MUST NOT double-charge.
- **Acceptance evidence:** connector validation test suite (accept/reject
  fixtures including unknown-meter and missing-evidence cases); replay-dedup
  integration test through the ingest path; end-to-end drill in which a
  third-party reference service becomes billable using only public
  documentation and the SDK.
- **Non-goals:** the technical/lifecycle connector APIs (OCS domain);
  marketplace product packaging and SKU binding UX (MKT domain).
- **Non-claims:** the validator covers structural checks today; no third-party
  service has completed the billable-onboarding drill.
- **Stop conditions:** money / trust / exposure — block connector publication
  if rate-card evidence is missing or stale; halt a service's billing if its
  events bypass the metering contract.
- **Traceability:** current-core (OCS billing connector contract);
  vision-deck. Related: OCS domain (17), MKT domain (18), CR-BIL-180.

### CR-BIL-140 — Billing-vs-observed-usage reconciliation evidence
- **Priority:** P0
- **Status:** proposed
- **Actors:** auditor, provider, operator
- **Problem:** A billing system that cannot prove billed usage equals observed
  usage is making unverifiable money claims — a direct violation of the
  platform's evidence-over-claims principle.
- **Requirement:** The platform MUST provide periodic reconciliation across
  four views, per billing account and period: (a) observed resource usage at
  the substrate (power states, allocated bytes, provisioned resources); (b)
  metered and ingested usage; (c) rated and charged usage; (d) invoiced
  totals. Reconciliation MUST run automatically at every billing-period close,
  MUST quantify divergence against declared tolerances, and MUST emit evidence
  records in the platform's standard evidence states (verified / blocked).
  Money-correctness and revenue claims MUST NOT be made for any period lacking
  verified reconciliation evidence.
- **Acceptance evidence:** reconciliation drill on synthetic stands with
  injected divergence (missing events, duplicate events, mispriced SKUs)
  proving detection; period-close evidence records; gate integration test
  showing money claims blocked without verified reconciliation.
- **Non-goals:** forensic audit tooling; cross-provider settlement
  reconciliation (CR-BIL-170, FED domain).
- **Non-claims:** reconciliation machinery is not yet implemented; tolerance
  values are placeholders pending pilot data.
- **Stop conditions:** money / settlement / trust — any period with divergence
  above tolerance MUST block invoicing and any external revenue claim until
  dispositioned; blocked reconciliation is a first-class honest state and MUST
  NOT be converted into a release claim.
- **Traceability:** legacy-platform-a; legacy-platform-b; req-history
  (evidence discipline). Related: CR-BIL-030, CR-BIL-110, CR-BIL-120.

### CR-BIL-150 — Public billing API: accounts, SKUs, operations
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, vendor, operator
- **Problem:** Tenants and automation need programmatic access to billing
  accounts, catalog prices, and long-running billing operations. Console-only
  billing breaks the platform's unified-API promise and blocks provider
  automation.
- **Requirement:** The platform MUST expose a public, authenticated, versioned
  billing API covering: billing accounts (get/list, binding of resources to
  accounts); the SKU and price catalog (list/get with price history); budgets
  (per CR-BIL-090); and long-running operations for asynchronous mutations,
  following the platform-wide operation contract. Mutations MUST be idempotent
  and audited. Errors MUST follow the platform's sanitized external error
  policy (no internal details leak). Authorization MUST go through the
  platform IAM/policy layer, never billing-local logic.
- **Acceptance evidence:** API contract tests (versioning, idempotency,
  operation lifecycle); IAM-enforcement tests proving cross-tenant access is
  denied fail-closed; parity contract check between console billing views and
  API responses.
- **Non-goals:** payment-method and acquiring APIs (provider extensions);
  cross-provider catalog aggregation (FED domain).
- **Non-claims:** the API surface is not yet specified; no conformance suite
  exists.
- **Stop conditions:** money / exposure / keys / trust — fail closed on any
  authorization ambiguity; halt the API if a cross-tenant read or write is
  observed, or if internal error details leak to external callers.
- **Traceability:** legacy-platform-a; legacy-platform-b; current-core.
  Related: IAM domain (15), portal/UX domain (19), CR-BIL-070.

### CR-BIL-160 — Tenant cost visibility in the console
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider
- **Problem:** Tenants must understand what they are paying for without
  operator involvement — per project, service, and SKU, with current-period
  accruals — or support load and disputes scale linearly with growth.
- **Requirement:** The console SHOULD present per-tenant cost visibility:
  current-period accrued cost versus credit consumption; cost breakdown by
  project, service, and SKU; the price catalog with effective dates; budget
  status (CR-BIL-090); and document download (CR-BIL-120). All figures MUST
  come from the same ledger and rating data as invoices — one source of truth.
  Preliminary (current-period, not yet finalized) values MUST be visually
  distinct from finalized ones. Cost views MUST respect IAM scoping
  fail-closed: a tenant sees only its own data.
- **Acceptance evidence:** UI/API parity contract checks; tenant-isolation
  scoping tests; drill reconciling console figures against the issued invoice
  for a closed period.
- **Non-goals:** cost-optimization recommendations; forecasting (analytics
  extensions).
- **Non-claims:** the console surface is not yet implemented; no production
  parity evidence exists.
- **Stop conditions:** exposure / money — hide the surface fail-closed if
  scoping or parity checks fail; never present fixture or synthetic data as
  live cost data.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: portal/UX
  domain (19), CR-BIL-150.

### CR-BIL-170 — Revenue-sharing readiness
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, vendor, operator, auditor
- **Problem:** The federation economy requires that stored charges can later
  be split among parties (service vendor, reselling master account,
  infrastructure provider). Retrofitting attribution after charges exist is
  infeasible — the dimensions must be persisted at rating time.
- **Requirement:** Rated usage and ledger entries SHOULD carry revenue
  attribution dimensions: originating service and vendor identity, reselling
  or master account where present, and charge decomposition sufficient to
  compute a per-party revenue split. Settlement execution stays out of the OSS
  core (Business tier / FED domain), but the platform MUST NOT persist charges
  in a shape that loses the attribution needed for later settlement.
  Attribution MUST survive replay and re-rating deterministically.
- **Acceptance evidence:** attribution schema contract tests; replay
  determinism tests for attribution; settlement dry-run on synthetic data
  proving a multi-party split is computable from stored charges alone.
- **Non-goals:** settlement execution, payout rails, cross-provider clearing;
  tax withholding.
- **Non-claims:** no settlement implementation exists or is claimed;
  attribution dimensions are design-stage and unvalidated by any real
  revenue-sharing run.
- **Stop conditions:** settlement / money — if stored charges lack sufficient
  attribution for a declared party split, halt period close for affected
  accounts until backfill or migration evidence exists.
- **Traceability:** vision-deck; legacy-platform-a. Related: federation domain
  (23), CR-BIL-080, CR-BIL-140.

### CR-BIL-180 — Metering integration as a launch gate for usage-bearing services
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, provider, operator
- **Problem:** A service that ships without metering silently forfeits revenue
  and breaks the platform's monetization contract. Reference platforms treated
  metering as a hard launch precondition across their full service fleets.
- **Requirement:** No usage-bearing service — first-party or third-party OCS —
  MUST be marked generally available without: (1) a declared billing connector
  surface (CR-BIL-130); (2) usage events validated against the metering
  contract (CR-BIL-010); (3) resolving test cases passing against the current
  SKU catalog (CR-BIL-040); and (4) a completed ingest dry-run in a
  pre-production stand (CR-BIL-020). Release gates MUST check these artifacts
  and block promotion when any is absent. Free-of-charge services MUST still
  meter: a zero price is a catalog decision, not an absence of measurement.
- **Acceptance evidence:** gate integration tests (promotion blocked when
  metering artifacts are missing); onboarding drill evidence for a reference
  service; catalog check proving zero-priced SKUs still resolve and rate.
- **Non-goals:** charging for the platform's own internal infrastructure
  consumption accounting.
- **Non-claims:** no release-gate wiring exists yet; the policy is unenforced.
- **Stop conditions:** money / trust — if an unmetered usage-bearing service
  is discovered in production, treat it as a launch-gate defect: meter
  retroactively where evidence exists and disclose unbillable windows honestly
  rather than estimating charges.
- **Traceability:** vision-deck; legacy-platform-a; legacy-platform-b.
  Related: OCS domain (17), deployment/CI-CD domain (22), CR-BIL-130.

### CR-BIL-190 — Multi-currency pricing and minor-unit conversion
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, tenant
- **Problem:** Providers in different jurisdictions price and invoice in
  different currencies. The OSS baseline ships single-currency, but the
  catalog and ledger must not hard-code one currency, or later expansion
  becomes a data migration.
- **Requirement:** The catalog MAY carry per-currency price lists for a SKU.
  Currency handling MUST use explicit minor-unit arithmetic with no fractional
  minor units, and currency MUST be declared on every price, ledger entry, and
  document. Currency selection per billing account MUST be explicit and stable
  within a period. Any FX conversion MUST use versioned rate sources with
  recorded provenance; implicit conversion between currencies is prohibited.
- **Acceptance evidence:** minor-unit conversion tests; mixed-currency
  rejection tests; rate-provenance contract checks on any conversion path.
- **Non-goals:** real-time FX semantics; crypto-denominated pricing.
- **Non-claims:** multi-currency is untested beyond conversion unit tests; no
  multi-currency installation exists.
- **Stop conditions:** money / settlement — halt mixed-currency invoicing if
  any implicit conversion is detected or rate provenance is missing.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-050,
  CR-BIL-060.

### CR-BIL-200 — Free-tier and trial mechanics as catalog data
- **Priority:** P2
- **Status:** proposed
- **Actors:** provider, tenant, vendor
- **Problem:** Free tiers (always-free usage floors) and trial periods are
  commercial levers that belong in catalog and policy data — not in
  service-code branches that diverge per product.
- **Requirement:** The platform MAY support free-tier definitions (per-SKU
  free quantity per period per billing account) and trial-period mechanics
  (grant-funded trial with declared end behavior) as declarative catalog and
  policy entries evaluated by the standard rating path, never by
  service-specific logic. Trial end MUST transition through the same account
  state machinery as CR-BIL-100 (stop-and-keep, never silent deletion).
- **Acceptance evidence:** rating tests proving free-floor subtraction before
  pricing; trial lifecycle tests covering end-of-trial transition and
  upgrade-to-paid; replay tests reproducing free-tier application for
  historical windows.
- **Non-goals:** marketing campaign management; anti-abuse and fraud scoring.
- **Non-claims:** no free-tier definitions are authored; abuse resistance is
  unstudied.
- **Stop conditions:** money — halt a free-tier change that cannot be
  reproduced by replay of prior windows.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-040,
  CR-BIL-080, CR-BIL-100.

### CR-BIL-210 — Public price calculator and cost estimation
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, provider, vendor
- **Problem:** Prospective tenants need pre-purchase estimates computed from
  the same prices that will actually rate their usage; a separately maintained
  price copy drifts and erodes trust.
- **Requirement:** The platform MAY expose a price-calculator API and UI that
  computes estimates from the current SKU catalog and price history — never
  from hard-coded price copies — and labels output clearly as an estimate.
  Every calculator response MUST be traceable to the catalog version it used.
- **Acceptance evidence:** parity tests between calculator output and the
  rating engine on identical synthetic usage; catalog-version traceability
  checks on responses.
- **Non-goals:** quoting/CPQ workflows and negotiated-discount handling
  (Business tier).
- **Non-claims:** the calculator is not yet specified; calculator-to-rating
  drift is untested.
- **Stop conditions:** trust / money — if calculator and rating diverge for
  identical inputs, block publication of estimates until reconciled.
- **Traceability:** legacy-platform-a; legacy-platform-b. Related: CR-BIL-040,
  CR-BIL-150.

## Coverage notes

This domain deliberately defers:

- **Payment acquiring, card processing, KYC, and legal-entity verification** —
  provider integrations and CloudRING Business extensions; the OSS core models
  accounts and payment-method metadata only.
- **Federation settlement execution, cross-provider clearing, and revenue-share
  payout** — federation domain (23) and the Business tier; this domain only
  guarantees charges are persisted with sufficient attribution (CR-BIL-170).
- **Marketplace product catalog, publisher SKU binding, and PAYG product
  packaging** — MKT domain (18); this domain consumes the resulting meters and
  prices through the connector surface (CR-BIL-130).
- **Connector lifecycle, capability registration, and technical APIs** — OCS
  domain (17); only the billing surface of the connector contract lives here.
- **IAM principals, tenancy model, and the policy engine** — IAM domain (15);
  billing accounts are commercial entities that reference, not embed, IAM
  identities.
- **Guest-agent implementation and power-state event sources** — compute
  domain (11); this domain consumes heartbeat and state evidence for metering.
- **Deletion and retention evidence mechanics** for suspended-account data —
  storage/backup domain (13); this domain declares the timelines and consumes
  the evidence.
- **Observability substrate** (metrics storage, dashboards, alerting) that
  pipeline lag and error metrics ride on — observability domain (20).
- **Consumption analytics, warehousing, forecasting, and ML** (LTV, antifraud,
  capacity) — data-services domain (24); the billing ledger is a source, not
  an analytics product.
- **Console shell and micro-frontend mechanics** — portal/UX domain (19); only
  cost-visibility content requirements live here.
- **Release-gate machinery** enforcing the metering launch gate — deployment/
  CI-CD domain (22); this domain defines the artifacts the gate checks.
- **Jurisdiction-specific fiscal adapters** (e-invoicing mandates, fiscal
  receipts) — provider extensions behind the EDO-class document interface.
