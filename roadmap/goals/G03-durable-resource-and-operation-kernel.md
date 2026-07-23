# G03 — Durable resource and operation kernel

## Outcome

Deliver the small business kernel used by every product: providers, regions,
cells, organizations, tenants, projects, resources, operations, events and audit.
It survives crashes and database failover without losing or duplicating work.

All reusable implementation must merge into public OSS and the exact accepted
change must be deployed and live-verified at `https://hub.cloudring.org`.

## Scope

- Define canonical IDs, resource names, hierarchy, ownership, labels,
  generations/conditions, etags and finalization/deletion semantics.
- Implement organization/tenant/project lifecycle and authoritative directory
  synchronization with durable cursors, tombstones, conflict precedence,
  staleness/unmapped metrics and non-destructive partial-failure behavior.
- Implement versioned resource-oriented HTTP APIs with pagination, filtering,
  field masks, idempotency keys, optimistic concurrency and stable errors.
- Model long-running work as durable Operations with progress, result/error,
  cancellation, retry, compensation, timeout and retention.
- Use PostgreSQL HA with owned schemas, expand/migrate/contract migrations,
  transactions, connection backpressure and restore.
- Add transactional outbox/inbox delivery using CloudEvents, at-least-once
  semantics, replay and deduplication. No broker without measured need.
- Implement immutable audit correlated with bootstrap identity, organization,
  tenant, operation, request and resulting resource version.
- Provide API/CLI and read-only portal views. Until G04/G05, all access is limited
  to the G02 bootstrap operator identity; no customer self-service is exposed.

## Required journeys

- create/update/delete organizations, tenants and projects with duplicate and
  conflicting requests;
- synchronize nested organization membership at the versioned scale profile,
  interrupt it and prove no destructive partial sync;
- crash between state commit and event delivery and recover one logical result;
- cancel before and reject after an operation point of no return;
- run concurrent conflicting updates with no lost update or cross-scope read;
- migrate with mixed-version replicas and roll back before the boundary;
- fail over PostgreSQL during accepted operations and meet RPO/RTO.

## Hub and downstream delivery

Deploy with no products, create protected test organizations/tenants, compare API,
CLI and durable state, then clean up. Downstreams contain no competing resource,
organization, operation, event or audit engine.

## Acceptance

- Public issue #24 closes with scale, restart, migration and failover proof. G03
  proves the durable graph/synchronization part of #25; identity-source lifecycle
  remains open until G04.
- No production memory store, fixture or hardcoded audit sink is reachable.
- Bounded queues, retry budgets and tenant/system fairness are measured.
- Generated schemas/clients validate the running contract and prerelease.
