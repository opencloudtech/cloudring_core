# G09 — Orders, subscriptions, lifecycle, quota and capacity

## Outcome

Deliver the generic commercial and technical lifecycle engine used by every OCS
product. A customer can request a product and reliably reach the intended state
despite retries, connector outages, process crashes or partial failure.

All reusable implementation must merge into public OSS and the exact accepted
change must be deployed and live-verified at `https://hub.cloudring.org`.

## Scope

- Define durable Order, OrderItem, Subscription, Entitlement, Allocation and
  Operation resources with explicit state machines and invariants.
- Implement validate/quote/authorize/commit boundaries and asynchronous
  provision, hold/suspend, resume, resize and deprovision orchestration, each
  declared `supported` or `not_applicable` with a machine-readable reason.
- Use reconciliation and sagas with explicit compensation, point of no return,
  retry classification, deadlines, cancellation and orphan detection.
- Guarantee idempotency from public request through connector execution and event
  delivery. Prevent double provisioning after timeout or replay.
- Add dependency declarations so one product can consume another only through a
  scoped infrastructure-user entitlement with region compatibility,
  quota/capacity reservation and billing attribution; detect cycles before
  activation.
- Enforce tenant, offering, audience, region, quota and IAM checks before work;
  recheck authorization at destructive execution boundaries.
- Implement hierarchical quotas, transactional reservations/release, capacity
  pools, overcommit policy, placement inputs, exhaustion behavior and fair
  scheduler-facing decisions. Money and usage rating remain G10.
- Provide API, CLI, portal timelines, operator intervention, retry/repair and
  immutable audit.

## Required journeys

- complete each supported lifecycle action with retries and concurrent duplicate
  requests, and reject an undeclared or unjustified not-applicable action;
- crash before and after external side effect and reconcile to exactly one
  logical resource;
- lose connector connectivity, expose honest degraded state, recover and resume;
- reserve capacity, resize/release it, reject exhaustion before side effects and
  reject unauthorized, unavailable and cyclic dependency orders;
- cancel before point of no return and reject cancellation afterward;
- compensate partial multi-step creation and surface any manual residue clearly;
- restore orchestration state and continue accepted operations after database
  failover and backup restore.

## Hub and downstream delivery

Deploy the engine to the reference platform with no public product. Exercise it
against the isolated G07 connector and fault-injection harness, then clean all
resources. The harness stays in tests/sandbox only. Both downstreams pin the same
engine and run conformance.

## Acceptance

- Public issue #29 and replacement requirements are fully proved.
- No synchronous request blocks for provider-side completion.
- Every terminal and non-terminal state has owner, timeout, recovery and audit.
- Restoration and replay cannot double-create or double-delete a resource.
- Quota/capacity reservations remain balanced through retry, cancellation,
  failure, restore and concurrent demand.
