# 18 — Marketplace and Catalog

Scope: the marketplace and product catalog of the Cloud Services Pod.
This domain covers the publisher-facing catalog model
(publisher → product → family → version), product lifecycle and visibility
rules, product↔SKU pricing bindings, installation-compatibility gating before
purchase, marketplace license classes and deployment tiers, per-service and
per-component licensing with vision-mandated expiry semantics, the publisher
metering write API, the platform commission model, artifact signing and
publication pipelines, cross-installation catalog synchronization, the
service-store model for first-class third-party services, and partner deal
registration. Rating, invoicing, and balances live in the billing domain; the
connector contract itself lives in the OCS domain.

**Domain contract.** The catalog is the single sellable source of truth:
nothing is purchasable unless it is active, signed, compatibility-checked
against the buyer's installation, and bound to billing through an explicit,
approved SKU binding. Money paths fail closed: publisher metering is
idempotent, batched, and dry-runnable; charging for an order starts only on
provisioning evidence, never on assumed success; any divergence between rated
usage and payout reports halts settlement. Trust paths fail closed: unsigned,
tampered, or unverifiable artifacts are never published or installed, and
signing-key compromise stops all new publications. License expiry never
disables tenant functionality — it stops maintenance and updates except
documented critical compatibility fixes, exactly as the product vision
states. Third-party services are first-class citizens: one catalog, one
order path, one metering contract for first-party and vendor products alike,
with no platform-team code in a vendor's sale path.

---

### CR-MKT-010 — Catalog data model: publisher → product → family → version
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, provider, operator
- **Problem:** A marketplace needs one canonical, API-managed catalog model.
  Catalog or price data held in files, spreadsheets, or per-service copies
  cannot be audited, synchronized across installations, or sold through
  consistently.
- **Requirement:** The platform MUST provide a catalog service whose entity
  model is publisher → product → (for versioned products) family → version,
  with per-version artifact references. Product kinds MUST cover at least
  machine-image products, simple (one-step deployable) products, and
  externally fulfilled (SaaS-style) products. Catalog data MUST live in a
  versioned, migrated store managed exclusively through an API — never in
  CSV/JSON files or ad hoc service databases. Categories, stable slugs,
  end-user-license references, and localization metadata SHOULD be
  first-class fields rather than free-form blobs.
- **Acceptance evidence:** catalog API contract tests exercising CRUD across
  all four entity levels; schema-migration tests for the catalog store; a
  conformance fixture demonstrating one product of each declared kind
  end-to-end; negative tests rejecting orphan versions, orphan families, and
  duplicate slugs.
- **Non-goals:** storefront presentation and UX; price computation and rating
  (billing domain); artifact binary storage itself.
- **Non-claims:** no catalog implementation exists in the current core; the
  model is informed by proven marketplace platforms but is unproven in
  CloudRING operation.
- **Stop conditions:** money / exposure — any write path that mutates
  price-bearing or visibility-bearing catalog state without an authorized
  role and an audit record is a release blocker; halt on discovery of
  unauthenticated catalog mutation.
- **Traceability:** legacy-platform-b (marketplace catalog model and entity
  set); legacy-platform-a (catalog store service, pricing-in-files lesson);
  vision-deck (marketplace component); current-core (connector metadata as
  catalog input). Related: CR-OCS-*, CR-BIL-*.

### CR-MKT-020 — Product lifecycle state machine and visibility rules
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, operator, provider, tenant
- **Problem:** Products must pass review before public visibility and must
  deprecate without breaking tenants who already run them. Ad hoc state
  handling leads to unreviewed products going public or running instances
  being silently broken.
- **Requirement:** Every product and version MUST follow an explicit
  lifecycle state machine with at least draft, in-review, active, suspended,
  deprecated, and rejected states, plus an explicit error state. Only active
  products/versions are publicly listable and purchasable. State changes MUST
  propagate consistently from a family to its contained products and
  versions. Every transition MUST be recorded as an auditable event (actor,
  UTC timestamp, reason). Deprecation or suspension MUST NOT delete, stop, or
  mutate already-provisioned tenant instances.
- **Acceptance evidence:** state-machine unit tests including all illegal
  transitions; visibility contract tests proving only active entries are
  publicly listed; audit-event assertions on every transition; a drill that
  deprecates a product with simulated active tenants and proves zero
  disruption to running instances.
- **Non-goals:** human moderation policy and staffing; content-quality
  judgment.
- **Non-claims:** the review/moderation workflow is not yet staffed or
  evidenced; family→product propagation semantics are unproven in production.
- **Stop conditions:** deletion / exposure — any transition that would hide,
  suspend, or delete artifacts referenced by running instances halts for
  explicit operator confirmation; publication attempts on unsigned artifacts
  halt (see CR-MKT-060).
- **Traceability:** legacy-platform-b (status model with review gate,
  suspension, deprecation, and status propagation); legacy-platform-a
  (deprecation-as-process lesson). Related: CR-MKT-060, CR-MKT-150.

### CR-MKT-030 — Product↔SKU binding and publisher pricing drafts
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, operator, provider
- **Problem:** A sellable product must be joined to billing SKUs through an
  explicit, reviewable binding. Ad hoc price wiring makes marketplace charges
  unexplainable and unreconcilable.
- **Requirement:** Product↔SKU binding MUST be an explicit catalog entity
  (product, SKU, publisher) with its own lifecycle. Publishers SHOULD propose
  pricing as SKU drafts that require platform approval before becoming
  effective. Effective-date price history MUST be retained per binding.
  Billing-relevant metadata (usage unit, pricing unit) MUST be validated
  against the billing SKU catalog before a binding activates, so that every
  purchasable product resolves to known, ratable SKUs.
- **Acceptance evidence:** contract tests for the draft → approve → bind →
  effective flow; validation tests rejecting bindings to unknown or inactive
  SKUs; audit trail of approval decisions; a reconciliation test proving
  every chargeable marketplace product carries exactly its intended active
  bindings and no orphans.
- **Non-goals:** SKU definition, pricing formulas, and rating (billing
  domain); publisher payout execution.
- **Non-claims:** the approval workflow is unimplemented; no live binding or
  reconciliation evidence exists yet.
- **Stop conditions:** money — a binding change on an active product with
  running usage halts for operator/finance review before taking effect;
  detection of duplicate or orphan active bindings suspends charging for the
  affected product until resolved.
- **Traceability:** legacy-platform-b (publisher pricing drafts,
  product-to-SKU binding entity, billing synchronization tasks);
  legacy-platform-a (catalog decoupled from rating lesson). Related:
  CR-BIL-*, CR-MKT-010.

### CR-MKT-040 — Installation-compatibility gate before purchase
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, vendor
- **Problem:** The vision promises that buyers purchase only services
  compatible with their installation. Selling incompatible products breaks
  trust and creates refund and support load.
- **Requirement:** Every product version MUST declare machine-readable
  compatibility metadata: supported platform version range, required
  capabilities and prerequisite services, supported deployment tier, and
  region/architecture constraints. The marketplace MUST evaluate this
  metadata against the buyer's installation descriptor before an order is
  accepted. Incompatible orders MUST fail closed with a machine-readable
  explanation. Compatibility evaluation SHOULD be re-runnable when an
  installation is upgraded, so already-purchased products can be re-checked.
- **Acceptance evidence:** a compatibility-rule test matrix (pass/fail across
  platform versions, capabilities, and constraints); an order-flow
  integration test proving an incompatible purchase is rejected before any
  payment step; an upgrade re-check drill on a reference installation.
- **Non-goals:** automatic remediation of incompatibility; capability
  negotiation between buyer and vendor.
- **Non-claims:** the installation-descriptor format is owned by the
  deployment domain and not yet finalized; the gate is unproven on real
  installations.
- **Stop conditions:** trust / money — if the buyer's installation descriptor
  cannot be obtained or verified, purchasing halts (fail closed); the
  marketplace MUST NOT proceed on assumed compatibility.
- **Traceability:** vision-deck (purchase of services compatible with the
  buyer's installation); legacy-platform-b (product/platform constraints);
  current-core (installation profiles). Related: CR-DPL-*, CR-FND-*,
  CR-MKT-170.

### CR-MKT-050 — Publisher metering write API
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, service-team, provider
- **Problem:** Third-party products must report usage through one contract or
  marketplace revenue cannot be measured. Metering transmission is the sole
  stated basis for marketplace monetization in the vision.
- **Requirement:** The platform MUST expose a metering write API for
  marketplace products, authenticated per publisher service account.
  Submissions MUST support batches of usage records, each carrying a
  client-generated idempotency key, a SKU identifier, a positive quantity in
  the SKU's usage unit, and a UTC timestamp. The API MUST return per-record
  accept/reject results with reasons. A validate-only dry-run mode MUST be
  available and MUST be exercised during product onboarding. A published test
  product with test SKUs SHOULD be provided for integration testing. Accepted
  records MUST flow into the platform metering pipeline with dead-letter
  handling — usage is never silently dropped.
- **Acceptance evidence:** API contract tests covering batch limits,
  idempotent replay (a repeated key returns the original result without
  double-counting), per-record rejection reasons, and dry-run validation
  without persistence; a pipeline end-to-end test proving accepted records
  reach rating; a load test at the declared batch and throughput envelope.
- **Non-goals:** rating, pricing, and invoice generation (billing domain);
  metering of first-party infrastructure services, which have their own
  emitters.
- **Non-claims:** the throughput envelope is not yet measured; no external
  publisher has integrated against this API; exactly-once economics under
  retries are tested synthetically only.
- **Stop conditions:** money / data — repeated malformed or unauthenticated
  submissions trigger throttling and quarantine of the offending publisher
  key, never a pipeline stall; metering-pipeline errors affecting a payout
  period halt settlement for that period until reconciled.
- **Traceability:** vision-deck (business connector metrics transmission as
  the monetization basis); legacy-platform-b (batched metering write API with
  idempotency keys, per-record results, validate-only mode, test product);
  legacy-platform-a (token-authenticated metrics ingest gateway). Related:
  CR-BIL-*, CR-OCS-*.

### CR-MKT-060 — Artifact signing and verified distribution
- **Priority:** P0
- **Status:** proposed
- **Actors:** vendor, operator, tenant, auditor
- **Problem:** Marketplace artifacts (images, packages, charts) execute
  inside tenant installations. Unsigned or tampered artifacts are a
  supply-chain trust break for the entire platform.
- **Requirement:** Every artifact published through the marketplace MUST be
  signed with platform-recognized keys before publication, and
  installation-side verification MUST fail closed on missing or invalid
  signatures. Signing keys MUST be held in an approved secrets/key-management
  workflow — never in repositories or plain configuration. Signature and
  provenance metadata (publisher, version, checksums, SBOM where available)
  SHOULD be recorded and auditable. Key rotation MUST be supported without
  invalidating previously signed artifacts.
- **Acceptance evidence:** signing-pipeline tests (sign → publish → verify
  round trip); a negative end-to-end test proving tampered or unsigned
  artifacts are refused at install time; a key-rotation drill; inspection of
  signing and verification audit logs.
- **Non-goals:** depth of vulnerability scanning (security domain); build
  reproducibility; the artifact build system itself.
- **Non-claims:** signing infrastructure is not yet deployed in the current
  core; SBOM coverage is partial; rotation procedures are untested.
- **Stop conditions:** keys / trust — signing-key compromise or any
  verification bypass halts all new publications, triggers incident response,
  and requires re-verification of already-distributed artifacts before sales
  resume.
- **Traceability:** legacy-platform-b (artifact signing before publication);
  legacy-platform-a (signed packages and SBOM in the standard's transport
  layer plan); current-core (secrets-never-configuration policy). Related:
  CR-IAM-*, CR-DPL-*, CR-MKT-150.

### CR-MKT-070 — Marketplace license classes and deployment tiers
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, provider, tenant, operator
- **Problem:** The vision defines an enterprise marketplace spanning
  open-source, SSPL, and proprietary products, with simple and enterprise
  deployment services. Without explicit classes, license obligations,
  deployment responsibility, and support expectations blur.
- **Requirement:** Every product MUST declare its license class (open-source,
  SSPL, or proprietary) and its deployment tier (simple self-service
  deployment vs enterprise deployment involving platform or vendor
  personnel). Class and tier MUST drive purchase-flow terms: license-text
  presentation, support entitlement, and who performs deployment. The
  open-source distribution MUST be able to operate the marketplace framework
  with open-source and SSPL content fully independent of any commercial tier.
- **Acceptance evidence:** catalog schema tests enforcing class and tier
  declarations on every product; purchase-flow tests showing tier-specific
  terms and deployment routing; an open-source-only installation drill that
  lists and installs open-class products with no commercial-tier dependency.
- **Non-goals:** legal review of individual licenses; the commercial-tier
  content itself, which lives outside this repository.
- **Non-claims:** the SSPL content strategy is not yet legally validated;
  enterprise deployment services are a vision-stage concept with no
  implementation evidence.
- **Stop conditions:** trust / exposure — a product with an unclear or
  conflicting license-class declaration must not be published; halt rather
  than guess obligations.
- **Traceability:** vision-deck (three license classes; simple vs enterprise
  deployment; open-source marketplace access from any installation);
  legacy-platform-a (framework/content split between open and commercial
  tiers). Related: CR-MKT-110, CR-MKT-120.

### CR-MKT-080 — Service-store model: first-class third-party services
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, service-team, provider, tenant
- **Problem:** The founding thesis makes third-party services first-class
  citizens. A marketplace that sells only platform-made products reproduces
  the second-class-citizen problem the platform exists to fix.
- **Requirement:** Any vendor service implementing the OCS connector contract
  MUST be listable in the marketplace through the same catalog, purchase, and
  provisioning path as first-party services, with no platform-team code
  changes. The service store MUST carry the service's declared capabilities,
  dependencies, and metering contract as catalog data. Purchasing a
  connector-backed service MUST result in provisioning through its connector,
  never through special-case platform logic.
- **Acceptance evidence:** an end-to-end onboarding drill in which a service
  built only from public documentation and the SDK is registered, listed,
  purchased, and provisioned on a reference installation; connector
  conformance-suite pass linked to the listing; parity tests proving first-
  and third-party services share one order path.
- **Non-goals:** vendor certification programs; editorial curation of the
  store.
- **Non-claims:** no external team has completed this path yet; it depends on
  OCS connector maturity whose evidence lives in the OCS domain.
- **Stop conditions:** trust / exposure — a connector service that fails
  conformance or security review must not become purchasable; limited
  visibility or sandbox states are used rather than rejection-free
  publishing.
- **Traceability:** vision-deck (first-class-citizen services; service
  store); current-core (OCSv3 connector packages and validation); legacy-
  platform-a (hub/connector service model). Related: CR-OCS-*, CR-MKT-170.

### CR-MKT-090 — Publisher commercial onboarding
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, operator, provider
- **Problem:** Paying publishers requires verified legal entities and
  contracts. Informal onboarding creates money-laundering, tax, and dispute
  risk.
- **Requirement:** Publisher accounts MUST capture legal-entity data and
  contract terms — contract type (offer/general), prepay or postpay, payout
  cycle (monthly or quarterly), currency, and tax treatment — before any paid
  product can be activated. Publisher account states MUST include at least
  pending, active, suspended, and terminated. Changes to payout-relevant data
  MUST be audited and SHOULD trigger re-verification before the next payout.
- **Acceptance evidence:** onboarding workflow tests (entity registration →
  contract → activation eligibility); state-machine tests for publisher
  accounts; audit-trail verification; a negative test proving a paid product
  cannot activate while the publisher contract is incomplete.
- **Non-goals:** selection of identity-proofing or company-data providers;
  tax filing; publisher credit scoring.
- **Non-claims:** legal-entity verification integrations are
  jurisdiction-specific and unimplemented; no publisher has been onboarded.
- **Stop conditions:** money / settlement — payout-relevant data changes
  mid-cycle halt payouts until re-verified; suspected fraudulent entity data
  halts onboarding and escalates to the operator.
- **Traceability:** legacy-platform-b (publisher account, legal person, and
  contract model with payout cycles); legacy-platform-a (partner legal-entity
  lookup integration). Related: CR-BIL-*, CR-IAM-*, CR-MKT-100.

### CR-MKT-100 — Platform commission model
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, vendor, operator, auditor
- **Problem:** Marketplace revenue for the platform is a commission on top of
  publisher prices. The composition rule must be explicit, uniform, and
  explainable per charge, or payouts and audits degenerate into disputes.
- **Requirement:** Plan price composition MUST be explicit data — publisher
  price plus platform commission plus applicable taxes — never a hard-coded
  constant. The commission rate MUST be configurable per installation with
  effective-date history. Every marketplace charge MUST decompose into
  publisher share, platform commission, and tax in reporting. Publisher
  payout reports MUST be generated from rated usage in the billing pipeline,
  not from estimates.
- **Acceptance evidence:** rating integration tests asserting charge
  decomposition arithmetic in fixed-point precision; configuration tests for
  commission-rate changes with effective dates; a reconciliation test between
  payout reports and billing-pipeline output over a synthetic period.
- **Non-goals:** revenue-share tuning policy such as volume-based reduction
  toward zero (vision-stage economics); cross-provider settlement, which is a
  federation concern.
- **Non-claims:** commission values are unratified business decisions;
  volume-based reduction mechanics are unproven vision economics; no payout
  has been executed.
- **Stop conditions:** money / settlement — any discrepancy between rated
  charges and payout reports halts payouts until reconciled; a commission
  rate outside declared bounds fails closed at configuration time.
- **Traceability:** vision-deck (revenue sharing and marketplace
  monetization); legacy-platform-b (publisher price plus platform fee plus
  tax composition; publisher revenue reports). Related: CR-BIL-*, CR-FED-*,
  CR-MKT-090.

### CR-MKT-110 — Per-service and per-component licensing with signed entitlements
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, vendor, operator
- **Problem:** The vision monetizes per service and per enterprise component.
  The platform needs a licensing model that issues, delivers, and checks
  entitlements without trusting client-side claims.
- **Requirement:** Entitlement MUST be represented as signed license tokens
  carrying product/component scope, feature flags, and expiry. A license
  service SHOULD manage license templates, issued instances, and
  renewal/recreation flows. License checks MUST be enforced at install and
  provision time by the platform, and SHOULD be re-checkable at runtime.
  License state MUST be auditable per tenant installation.
- **Acceptance evidence:** token issue/verify round-trip tests including
  tamper rejection; license-service CRUD and renewal-flow tests; an
  install-time gate end-to-end test (no entitlement → no enterprise
  install); audit queries demonstrating per-installation license state.
- **Non-goals:** vendor-proprietary license servers; hardware or dongle
  licensing; per-seat office-suite-style licensing.
- **Non-claims:** the token format and license service are unimplemented;
  runtime re-check cadence and its availability impact are undecided.
- **Stop conditions:** keys / trust — license-signing-key compromise halts
  new issuance and triggers re-issuance; license-verification errors fail
  closed for new enterprise installs but MUST NOT corrupt or stop already
  running workloads (expiry semantics in CR-MKT-120).
- **Traceability:** vision-deck (license fees per service and per enterprise
  component); legacy-platform-a (signed license tokens with feature flags;
  license-check middleware in the SDK); legacy-platform-b (license templates,
  instances, and recreation flows). Related: CR-IAM-*, CR-MKT-070,
  CR-MKT-120.

### CR-MKT-120 — License expiry semantics and update channels
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, vendor, operator
- **Problem:** The vision's licensing policy is unusual and legally
  load-bearing: on expiry, functionality is preserved while maintenance and
  updates stop — except critical compatibility fixes. The platform must
  implement this exactly or not claim it.
- **Requirement:** License expiry MUST NOT disable functionality of installed
  enterprise components. Expiry MUST stop delivery of maintenance and update
  channels for the expired scope. A critical-fix channel carrying only
  documented compatibility-essential fixes MUST remain deliverable after
  expiry. Update channels SHOULD distinguish at least stable,
  critical-fixes-only, and pre-release. The rules defining what qualifies as
  a critical compatibility fix MUST be documented, versioned, and applied
  consistently; expiry events and post-expiry deliveries MUST be auditable.
- **Acceptance evidence:** an expiry drill proving a component keeps
  functioning after expiry with normal updates blocked and the critical
  channel still deliverable; channel-routing tests; presence of the
  critical-fix classification policy in the release process; audit inspection
  of post-expiry deliveries.
- **Non-goals:** grace-period commercial policy; refunds; renewal sales
  flows.
- **Non-claims:** critical-fix classification is a policy judgment never yet
  exercised; no enterprise component exists to enforce against; the semantics
  are stated in the vision but operationally unproven.
- **Stop conditions:** trust — any update path that would hard-disable tenant
  functionality on expiry is a release blocker; ambiguous fix classification
  escalates to owner review before shipping on the critical channel.
- **Traceability:** vision-deck (expiry semantics stated as policy);
  legacy-platform-a (license lifecycle: expired ⇒ functionality preserved,
  updates stopped except a critical-compat channel; stable / stable-critical
  / pre-release channels). Related: CR-MKT-110, CR-DPL-*.

### CR-MKT-130 — License-rule evaluation on resource specifications
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, provider, tenant
- **Problem:** Some products carry vendor licensing constraints tied to the
  shape of the requested resource (for example per-core operating-system
  licensing). The platform must check such rules before allowing launches.
- **Requirement:** Products SHOULD be able to declare machine-readable
  license rules — category, target entity, attribute path, expected values —
  evaluated against the requested resource specification. Launches that
  violate license rules MUST be denied with an explanation. Rule evaluation
  MUST be deterministic, unit-tested, and logged. Rule sets MUST be versioned
  together with the product version they constrain.
- **Acceptance evidence:** a rule-engine test matrix over synthetic resource
  specifications; a launch-gate end-to-end denial test; evaluation-log
  inspection showing rule, input, and verdict per check.
- **Non-goals:** interpreting vendor license agreements; automatic license
  procurement.
- **Non-claims:** the rule vocabulary is minimal today; no vendor rules are
  in production use.
- **Stop conditions:** money / trust — a rule-evaluation failure (engine
  error, unknown rule kind) fails closed: the launch is denied and the vendor
  is notified; never allow on evaluation error.
- **Traceability:** legacy-platform-b (license rules evaluated against
  resource specifications; license-check service gating launches). Related:
  CR-MKT-110, CR-CMP-*.

### CR-MKT-140 — PAYG metering modes and in-guest usage agent
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, tenant, provider
- **Problem:** Pay-as-you-go products — per-second VM, cores, or RAM while
  running, or custom publisher metrics — need trustworthy usage evidence from
  inside guest workloads without over-billing tenants.
- **Requirement:** The platform SHOULD provide an in-guest agent emitting
  signed, periodic heartbeat/usage evidence over a controlled host channel.
  PAYG charge types MUST include at least free, BYOL, and metered
  (platform-measured per-second resources or publisher custom metrics).
  Guest-agent evidence MUST be correlatable with platform-side power state so
  stopped instances are not charged. Agent absence MUST degrade to a
  documented fallback behavior, never to silent billing.
- **Acceptance evidence:** a guest-agent end-to-end test (heartbeat → rated
  usage); reconciliation tests between agent evidence and platform-observed
  instance state; over-billing regression tests (stopped instance not
  charged); tamper and spoofing negative tests on the host channel.
- **Non-goals:** the agent as a general-purpose management plane; mandatory
  agents for free or BYOL products.
- **Non-claims:** the agent is not yet built; the trust model of
  guest-reported usage is a known hard problem and remains partially
  unproven.
- **Stop conditions:** money — divergence between agent-reported and
  platform-observed usage beyond a declared tolerance halts PAYG charging for
  affected products and escalates to reconciliation.
- **Traceability:** legacy-platform-b (in-guest heartbeat agent as per-second
  PAYG evidence; free/BYOL/PAYG charge types); legacy-platform-a
  (per-product metering agents). Related: CR-BIL-*, CR-CMP-*.

### CR-MKT-150 — Asynchronous artifact and image publication pipeline
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, operator
- **Problem:** Publishing a product version — building, cloning, verifying,
  publishing to regional pools — is multi-step and failure-prone. It must be
  durable, resumable, and evidence-producing, or partial publishes leak into
  the catalog.
- **Requirement:** Artifact publication MUST run as durable, resumable tasks
  with explicit stages (build, clone, verify, publish, pool sizing), retry
  and cancel semantics. Each stage MUST record evidence (logs, checksums,
  signatures). A version MUST NOT become purchasable until all stages and the
  signing step (CR-MKT-060) complete. Artifact removal MUST follow the
  lifecycle rules of CR-MKT-020, never ad hoc deletion.
- **Acceptance evidence:** pipeline task tests with failure injection proving
  resume without duplicate side effects; an end-to-end publish of a test
  product version; evidence-record inspection per stage; a gate test proving
  a version is not purchasable mid-pipeline.
- **Non-goals:** the image build system itself (image-factory pipelines live
  in the deployment/IaC domain); artifact storage engineering.
- **Non-claims:** the pipeline is unimplemented; cross-region pool management
  is unproven.
- **Stop conditions:** deletion / exposure — a failed publish MUST NOT leave
  partially public artifacts; cleanup of failed publishes is automatic or
  explicitly operator-approved, never silent.
- **Traceability:** legacy-platform-b (async image pipeline tasks: clone,
  create, publish, pool sizing); legacy-platform-a (image factory and
  cross-environment artifact movement). Related: CR-MKT-060, CR-MKT-020,
  CR-DPL-*.

### CR-MKT-160 — Cross-installation product synchronization
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, vendor
- **Problem:** CloudRING is many installations, including sovereign and
  air-gapped ones. Catalog content must replicate across installations
  without a live dependency on a single origin.
- **Requirement:** Catalog products, versions, and their artifacts SHOULD be
  synchronizable across installations through a declarative, auditable
  mechanism. Synchronization MUST carry signatures and compatibility metadata
  intact. Receiving installations MUST re-verify signatures and re-evaluate
  compatibility before listing imported products. The mechanism SHOULD
  tolerate disconnected operation through an exportable bundle path for
  sovereign installations. Divergence between installations MUST be
  detectable and reported.
- **Acceptance evidence:** a two-installation sync drill (publish on
  installation A → re-verify, list, and install on installation B); an
  offline bundle-sync test; a divergence-report test; a negative test proving
  a tampered payload is rejected by the receiver.
- **Non-goals:** global catalog search across the federation (federation /
  global portal domain); real-time synchronization.
- **Non-claims:** no multi-installation deployment exists yet to validate
  against; conflict policy for divergent edits is undefined; this requirement
  is entirely forward-looking.
- **Stop conditions:** trust / exposure — signature or metadata mismatch
  halts import of the affected product set; the receiver never auto-resolves
  by blindly preferring the source.
- **Traceability:** vision-deck (distributed and disconnected installations);
  legacy-platform-b (catalog sync agent across installations; bundle-based
  distribution for sovereign installs). Related: CR-FED-*, CR-DPL-*,
  CR-MKT-060.

### CR-MKT-170 — Order flow and provisioning handoff
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, vendor
- **Problem:** A purchase must result in actual provisioning through the
  correct path — connector for services, deployment for images — with
  trackable status and safe failure. Billing must never run ahead of
  delivery.
- **Requirement:** The marketplace MUST provide an order flow:
  compatibility check (CR-MKT-040) → terms acceptance (EULA where
  applicable) → entitlement issuance where licensed (CR-MKT-110) →
  provisioning handoff to the product's connector or deployment path.
  Orders MUST be asynchronous with trackable operation status and idempotent
  submission. Failed provisioning MUST NOT leave the tenant charged for
  undelivered resources: charging starts on provisioning-success evidence.
  Order history MUST be auditable per tenant.
- **Acceptance evidence:** order end-to-end tests per product kind;
  idempotency tests (double submission yields one provisioned instance);
  failure-injection tests proving no charge without provisioning evidence;
  audit inspection of per-tenant order history.
- **Non-goals:** shopping-cart and checkout UX (portal domain); payment
  processing and balances (billing domain).
- **Non-claims:** no order path is implemented; the charge-on-provisioning-
  evidence coupling with billing is designed but unproven.
- **Stop conditions:** money — if provisioning status is unknown or ambiguous
  after a declared timeout, billing activation halts and operators are
  alerted; the platform never bills on assumed success.
- **Traceability:** legacy-platform-b (ordering model); legacy-platform-a
  (marketplace provisioning service with explicit state machine);
  current-core (connector lifecycle APIs). Related: CR-OCS-*, CR-BIL-*,
  CR-MKT-040, CR-MKT-110.

### CR-MKT-180 — Partner and reseller deal registration
- **Priority:** P2
- **Status:** proposed
- **Actors:** agent, operator, provider, vendor
- **Problem:** The ecosystem includes resellers, integrators, and agents on
  non-discriminatory terms. Their opportunities need registration to
  attribute revenue and avoid channel conflict.
- **Requirement:** The platform SHOULD provide a deal-registration capability
  in which partners register opportunities (customer, product scope, terms)
  with validation of company data, authenticated through platform identity.
  Registrations MUST have an auditable lifecycle (submitted, validated,
  approved, rejected, expired) and SHOULD hand off asynchronously to
  backoffice or ticketing systems. Where a registered deal completes,
  partner attribution MUST feed commission and revenue-share reporting.
- **Acceptance evidence:** registration workflow tests including validation
  failures; lifecycle and audit tests; an asynchronous handoff integration
  test against a ticketing stub; an attribution report test linking a
  completed sale to its registered deal.
- **Non-goals:** CRM functionality; partner contract management; discount
  policy engines.
- **Non-claims:** partner program terms are undefined; no partner exists;
  company-data validation integrations are jurisdiction-specific and
  unproven.
- **Stop conditions:** data / money — registered deals contain partner and
  customer personal data: access is need-to-know and audited; attribution
  disputes halt associated payouts until resolved.
- **Traceability:** vision-deck (resellers, integrators, and agents on
  non-discriminatory terms); legacy-platform-a (partner deal-registration
  portal with company-data validation and asynchronous backoffice handoff).
  Related: CR-IAM-*, CR-BIL-*, CR-MKT-100.

### CR-MKT-190 — Advanced product kinds: BYOL, SaaS order forms, prebuilt stacks
- **Priority:** P2
- **Status:** proposed
- **Actors:** vendor, tenant, provider
- **Problem:** Beyond image and simple products, the marketplace vision
  includes bring-your-own-license products, SaaS-style offerings needing
  structured order input, and prebuilt multi-service stacks.
- **Requirement:** The catalog MAY support BYOL products, where the license
  is supplied by the tenant and the platform verifies its presence without
  vending licenses. SaaS-style products MAY declare order forms — structured
  input collected at purchase and validated against a schema. Prebuilt-stack
  products (compositions of multiple services) MAY be publishable with
  dependency resolution at deploy time. Every advanced kind MUST reuse the
  same lifecycle, compatibility, and signing gates as base kinds.
- **Acceptance evidence:** per-kind contract tests; a BYOL license-presence
  check test; order-form schema-validation tests; a dependency-resolution
  end-to-end test on a synthetic multi-service stack.
- **Non-goals:** stack authoring UX; a marketplace for professional services;
  license resale agreements.
- **Non-claims:** all three kinds are unimplemented and unprioritized;
  stack-composition semantics need a design before any capability claim.
- **Stop conditions:** trust — BYOL verification failures deny activation
  (fail closed); dependency cycles or unresolved dependencies block
  publication of a composed product.
- **Traceability:** legacy-platform-b (free/BYOL/PAYG charge types; SaaS
  products with order forms; prebuilt-stack blueprints); legacy-platform-a
  (external-vendor fulfillment adapters). Related: CR-MKT-110, CR-OCS-*.

---

## Coverage notes

This domain deliberately defers:

- **SKU definition, pricing formulas, rating, invoicing, balances, budgets,
  grants, and the metering pipeline internals** → billing/FinOps domain
  (`16`, CR-BIL-*). MKT owns only the product↔SKU binding and the publisher
  metering write contract.
- **The service connector contract itself** — registration, capability
  announcement, lifecycle APIs, dependency declaration, conformance suite →
  OCS domain (`17`, CR-OCS-*). MKT consumes connector metadata as catalog
  input.
- **Identity, authentication, RBAC roles for publishers/operators, and
  secrets/key-management mechanics** → IAM domain (`15`, CR-IAM-*). MKT
  states where authorization and signing keys are required, not how they are
  implemented.
- **Storefront, console screens, and purchase UX** → portal/UX domain (`19`,
  CR-CUX-*).
- **Installation descriptors, platform version detection, image-factory
  build pipelines, and update-channel delivery machinery** → deployment/IaC
  and foundation domains (`22`, `10`; CR-DPL-*, CR-FND-*). MKT consumes
  these as gate inputs.
- **Cross-provider catalog aggregation, global search, settlement between
  providers, and federation-level revenue sharing** → federation domain
  (`23`, CR-FED-*). CR-MKT-160 stops at pairwise installation sync.
- **Observability, audit-pipeline, and support-tooling requirements for
  marketplace services themselves** → observability and ops domains (`20`,
  `21`).

Honesty note: the current core implements the OCS connector substrate, not a
marketplace. Every requirement here is `proposed`; forward-looking items
(cross-installation sync, commission settlement, license-expiry enforcement,
guest-agent metering, deal registration) carry explicit Non-claims and must
not be promoted without the named evidence.
