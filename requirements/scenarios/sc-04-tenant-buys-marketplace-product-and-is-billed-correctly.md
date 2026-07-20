# SC-04 — Tenant buys a marketplace product and is billed correctly

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the commercial loop end to end with money-path honesty: a tenant
discovers an active, signed, compatibility-checked product in the catalog,
sees the full price before commit, places an idempotent order, is
provisioned through the product's connector, and is charged **only** on
provisioning-success evidence — with every usage record rated in
fixed-point decimal, reconciled against observed usage, and presented
identically on every surface.

## Actors

- tenant — buyer
- vendor / service-team — publisher of the product
- provider — operates the installation, billing, and marketplace
- auditor — verifies reconciliation and charge evidence

## Preconditions

- The publisher completed commercial onboarding and the product exists in
  the catalog as publisher → product → family → version (CR-MKT-010,
  CR-MKT-090).
- The product version is active and its artifacts are signed and
  verifiable (CR-MKT-020, CR-MKT-060).
- The product is bound to approved SKUs; the SKU catalog carries pricing
  formulas, unit conversion, and resolving rules as data (CR-MKT-030,
  CR-BIL-040).
- The tenant holds a billing account (prepaid or postpaid) separated from
  IAM identity (CR-BIL-070).

## Steps

1. **Discover the product.** The tenant browses the catalog; the product
   card shows lifecycle state, publisher, versions, and pricing.
   - **Expected outcome:** only active, signed, compatibility-declared
     versions are purchasable; draft or deprecated versions follow the
     visibility rules.
   - **Requirements:** CR-MKT-010, CR-MKT-020

2. **Estimate cost before commit.** The tenant uses the price calculator
   and sees the pre-commit review of cost, defaults, and exposure.
   - **Expected outcome:** pricing is reproducible from the catalog for
     the chosen configuration; no hidden cost, default, or public exposure
     is introduced before explicit commit.
   - **Requirements:** CR-BIL-210, CR-BIL-040, CR-CUX-040, CR-FND-120

3. **Pass the installation-compatibility gate.** The marketplace evaluates
   the product version's compatibility metadata against the buyer's
   installation descriptor.
   - **Expected outcome:** an incompatible order fails closed with a
     machine-readable explanation before any payment step; if the
     installation descriptor cannot be obtained or verified, purchasing
     halts.
   - **Requirements:** CR-MKT-040

4. **Accept terms and receive entitlement.** The tenant accepts the
   product terms; licensed products issue a signed entitlement.
   - **Expected outcome:** entitlement issuance is recorded; license
     class, deployment tier, and expiry semantics are visible to the
     tenant.
   - **Requirements:** CR-MKT-110, CR-MKT-070

5. **Place the order.** The order is submitted asynchronously with an
   idempotency key.
   - **Expected outcome:** the order follows compatibility check → terms
     acceptance → entitlement → provisioning handoff; a double submission
     yields exactly one provisioned instance.
   - **Requirements:** CR-MKT-170

6. **Provision through the connector.** The order hands off to the
   product's connector lifecycle APIs.
   - **Expected outcome:** provisioning executes through the mandatory
     lifecycle APIs with idempotency and rollback references; order status
     is trackable end to end.
   - **Requirements:** CR-MKT-170, CR-OCS-030

7. **Start charging only on delivery evidence.** Billing activation waits
   for provisioning-success evidence.
   - **Expected outcome:** failed provisioning leaves the tenant
     uncharged; if provisioning status is ambiguous after the declared
     timeout, billing activation halts and operators are alerted — the
     platform never bills on assumed success.
   - **Requirements:** CR-MKT-170, CR-FND-120

8. **Transmit usage through the metering path.** The product reports
   usage through the publisher metering write API / billing connector in
   batches with idempotency keys.
   - **Expected outcome:** per-record accept/reject results with reasons;
     a repeated key returns the original result without double-counting;
     accepted records flow into the metering pipeline with dead-letter
     handling — usage is never silently dropped.
   - **Requirements:** CR-MKT-050, CR-OCS-060, CR-BIL-020

9. **Mediate and enrich usage.** The pipeline stages raw records into
   ratable usage.
   - **Expected outcome:** every hop is replayable; every rejection is
     dead-lettered with a reason; no usage is silently altered or dropped.
   - **Requirements:** CR-BIL-030

10. **Rate with decimal precision.** Usage is rated against the SKU
    catalog and the per-installation price bundle in force for the usage
    window.
    - **Expected outcome:** all money arithmetic is fixed-point decimal
      with explicit rounding boundaries; historical windows re-rate
      reproducibly from price history.
    - **Requirements:** CR-BIL-060, CR-BIL-040, CR-BIL-050

11. **Apply grants and credits in order.** Prepaid balances, grants, and
    credits are consumed expiry-ordered with charge decomposition.
    - **Expected outcome:** the tenant can see which grant paid for which
      charge; credits never pay arrears.
    - **Requirements:** CR-BIL-080, CR-BIL-070

12. **Reconcile billed versus observed usage.** For VM-class products,
    heartbeat-derived usage reconciles against power-state events; for all
    products, rated usage reconciles against observed consumption.
    - **Expected outcome:** reconciliation evidence covers the billing
      window; divergence beyond declared tolerance halts invoicing for the
      affected accounts until reconciled.
    - **Requirements:** CR-BIL-140, CR-BIL-110

13. **Issue the billing document.** The invoice for the period is
    generated exclusively from rated and charged ledger data.
    - **Expected outcome:** immutable sequential numbering and content
      integrity per issued document; corrections issue corrective
      documents; issued documents are never mutated; documents are
      exportable by the tenant.
    - **Requirements:** CR-BIL-120

14. **Show one truth everywhere.** The console, API, CLI, and agent
    surfaces show the same order state, usage, balance, and documents.
    - **Expected outcome:** the tenant's cost view matches the ledger;
      surfaces never diverge; denials name the missing precondition.
    - **Requirements:** CR-BIL-160, CR-CUX-010, CR-FND-140

15. **Prove the negative paths.** Order-failure and replay drills run as
    part of acceptance.
    - **Expected outcome:** failure-injection proves no charge without
      provisioning evidence; metering replay produces no double-count;
      unmetered usage-bearing products cannot reach the catalog at all.
    - **Requirements:** CR-MKT-170, CR-BIL-130, CR-BIL-180

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-MKT-010 | Catalog data model: publisher → product → family → version | 1 |
| CR-MKT-020 | Product lifecycle state machine and visibility rules | 1 |
| CR-MKT-030 | Product↔SKU binding and publisher pricing drafts | preconditions |
| CR-MKT-040 | Installation-compatibility gate before purchase | 3 |
| CR-MKT-050 | Publisher metering write API | 8 |
| CR-MKT-060 | Artifact signing and verified distribution | preconditions |
| CR-MKT-070 | Marketplace license classes and deployment tiers | 4 |
| CR-MKT-090 | Publisher commercial onboarding | preconditions |
| CR-MKT-110 | Per-service and per-component licensing with signed entitlements | 4 |
| CR-MKT-170 | Order flow and provisioning handoff | 5, 6, 7, 15 |
| CR-OCS-030 | Mandatory lifecycle APIs: idempotent, rollback-referenced | 6 |
| CR-OCS-060 | Billing connector and usage-metrics transmission | 8 |
| CR-BIL-020 | Meter ingest API: authenticated, idempotent, replayable, dead-lettered | 8 |
| CR-BIL-030 | Mediation and enrichment pipeline | 9 |
| CR-BIL-040 | Declarative SKU catalog: formulas, unit conversion, resolving rules | 2, 10 |
| CR-BIL-050 | Price history and per-installation price bundles | 10 |
| CR-BIL-060 | Rating with fixed-point decimal precision | 10 |
| CR-BIL-070 | Billing account model: prepaid/postpaid, separated from IAM | 11 |
| CR-BIL-080 | Grants and credits: expiry-ordered consumption, charge decomposition | 11 |
| CR-BIL-110 | Per-second VM usage heartbeat as metering evidence | 12 |
| CR-BIL-120 | Invoices and billing documents (EDO-class) | 13 |
| CR-BIL-130 | OCS billing connector surface: meters, rate-card evidence, replay dedup | 15 |
| CR-BIL-140 | Billing-vs-observed-usage reconciliation evidence | 12 |
| CR-BIL-160 | Tenant cost visibility in the console | 14 |
| CR-BIL-180 | Metering integration as a launch gate for usage-bearing services | 15 |
| CR-BIL-210 | Public price calculator and cost estimation | 2 |
| CR-CUX-010 | Cross-surface parity contract | 14 |
| CR-CUX-040 | Pre-commit review of cost, defaults, and exposure | 2 |
| CR-FND-120 | Production-honesty bans | 2, 7 |
| CR-FND-140 | One product truth across surfaces | 14 |

## Gaps found

- **Payment acquiring is out of corpus scope.** The billing domain
  explicitly excludes payment acquiring and settlement execution; this
  scenario therefore closes at the rated ledger and issued invoice. If an
  acceptance run requires real money capture, that capability must be
  supplied by a provider adapter outside this corpus, and no corpus
  requirement currently governs that adapter boundary.
- No requirement currently states how marketplace **commission
  reporting** (CR-MKT-100) is reconciled against the same ledger window as
  tenant invoices; the scenario does not depend on it, but a commercial
  audit would.

## Evidence required

- Compatibility-gate test matrix and an order-flow integration test
  proving incompatible orders are rejected before payment (CR-MKT-040).
- Order end-to-end tests per product kind, idempotency proof, and
  failure-injection proof of no-charge-without-provisioning
  (CR-MKT-170).
- Metering API contract tests: batch limits, idempotent replay, per-record
  rejection reasons, dry-run validation (CR-MKT-050, CR-BIL-020).
- Rating golden tests proving fixed-point precision and reproducible
  re-rating from price history (CR-BIL-060, CR-BIL-050).
- Grant/credit consumption and decomposition evidence (CR-BIL-080).
- Reconciliation evidence for the billing window, including
  heartbeat-vs-power-state reconciliation for VM-class products
  (CR-BIL-140, CR-BIL-110).
- Invoice generation golden tests, immutability proof, and a tenant
  document-export drill (CR-BIL-120).
- Cross-surface consistency checks for order, usage, and balance views
  (CR-CUX-010, CR-BIL-160).
