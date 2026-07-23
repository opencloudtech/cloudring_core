# G10 — Metering, rating and billing

## Outcome

Providers can define offers and safely sell any OCS product; customers can see
usage, forecast cost, set budgets and understand invoices; product teams can emit
usage without implementing provider-specific billing.

All reusable implementation must merge into public OSS and the exact accepted
change must be deployed and live-verified at `https://hub.cloudring.org`.

## Scope

- Define meters, dimensions, units, aggregation, sampling, corrections and
  quality rules in OCS. Validate meter declarations during moderation.
- Implement authenticated, idempotent usage ingestion with late data, dedup,
  replay, gap detection, quarantine and lineage to resource and operation.
- Implement immutable, effective-dated price books, tiers, commitments, proration,
  discounts, taxes boundary, currencies, exact decimal rounding and quotes.
- Separate publisher price proposals from provider-approved effective price and
  commission policy, with distinct approvers, histories and audit trails.
- Separate usage ledger, rating results and double-entry financial ledger. Never
  mutate historical rating silently; issue explicit corrections/credits.
- Implement invoice preview/finalization, credit/dispute workflow and export. Real
  payment processing remains a provider adapter, not an OSS dependency.
- Compute immutable publisher share, provider commission, corrections and
  publisher payable liability, bound to the exact G08 publisher
  commercial-eligibility revision approved for the offering and charge; emit
  reconciled instructions only while that state permits payout. Jurisdictional
  KYC, tax and payment execution remain replaceable provider adapters with
  versioned decisions and receipts, not hard-coded OSS dependencies.
- Make service-to-service consumption visible and billable to an
  infrastructure-user identity and owning product team.
- Publish FOCUS-compatible cost/usage export and cost allocation tags.
- Add budgets, alerts, anomaly signals, forecast and provider/customer API, CLI
  and portal views.
- Run ingest, rating, financial ledger, invoice reads and their transactional
  stores without a single serving replica or node dependency; support compatible
  rolling or blue/green upgrades with no lost/duplicated usage, unbalanced ledger
  or invalid invoice.

## Required journeys

- quote, consume the G09 allocation, ingest usage, rate, invoice and trace every line
  back to immutable source;
- retry and replay usage without duplicate cost;
- receive late/corrected usage after invoice and create auditable adjustment;
- resize allocation with correct proration against the G09 reservation history;
- bill product A for use of product B through infrastructure-user entitlement;
- restore billing state and prove ledger balance and invoice identity.
- kill one serving replica and perform a signed rolling upgrade while continuous
  usage ingest, rating and invoice reads remain correct.
- onboard an independently built third-party OCS product, rate its real test
  usage, reconcile tenant charge = publisher share + provider commission + tax
  and corrections, bind the payable liability to the exact approved G08
  publisher-eligibility revision, and feed the same liability through two
  jurisdictional KYC/tax/payment adapter fixtures; stale, revoked or mismatched
  eligibility blocks payout instruction without corrupting the charge ledger.

## Hub and downstream delivery

Deploy the complete pipeline to the hub with synthetic test meters only in an
isolated billing tenant. Verify and delete test financial artifacts according to
retention policy. No customer invoice is issued. Downstreams may supply tax,
payment and legal adapters but cannot fork generic meter/rating semantics.

## Acceptance

- Public issue #30 and related FinOps requirements are complete.
- All money math is exact and property-tested; ledger always balances.
- Billing remains functional with zero installed products and does not own quota
  or capacity state.
- Usage backpressure cannot overload resource lifecycle or another tenant.
- Every marketplace charge reconciles exactly to immutable publisher share,
  provider commission, tax, corrections and payout liability.
- Every payout liability and adapter receipt identifies the exact approved G08
  publisher commercial-eligibility revision; adapter replacement cannot change
  the immutable charge, share or commission history.
- Public docs explain what OSS provides and what a jurisdiction/provider must
  configure.
