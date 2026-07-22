# CloudRING Product Architecture Invariants

CloudRING is a provider-neutral cloud platform and the reference implementation
of the project-defined Open Cloud Standard (OCS). These invariants keep the
single-cell platform, future products, and future provider federation compatible
without making later capabilities part of the current release.

## Platform and product boundary

The platform owns identity, authorization, policy, product admission, the
resource and operation model, catalog composition, usage transport, billing
integration, the provider shell, and evidence. A cloud product owns its service
implementation, product-specific API, lifecycle controller, user experience,
support diagnostics, and durability behavior.

Products integrate only through versioned OCS contracts. They may run in the
platform cell, in another cell, or in an independently operated installation.
The core must not import product implementation details or require a product to
use Kubernetes, Go, or the platform's storage internally.

Management-plane dependencies such as Kubernetes, networking, certificates,
secret management, databases, and backup are substrate components. A
tenant-visible compute, network, storage, database, license, external-account,
or other offering is an OCS product even when maintained by the platform team.

## Stable identity namespace

Every resource and event can carry provider, cell, region, organization, tenant,
project, service, and resource identity without assuming a single provider or
central control plane. Identifiers are opaque, stable, non-secret, and portable.
They never embed a provider hostname, database key, or implementation language.

Local availability must not depend on a future federation coordinator. A cell
continues serving local workloads when regional, global, marketplace, or peer
control layers are unavailable.

## One operation across every surface

Every mutating capability has one canonical `actionId`. The same action binds:

- the versioned API operation;
- generated CLI and infrastructure-as-code operations;
- the provider-shell or product-microfrontend action;
- the IAM decision and policy inputs;
- the idempotency key and long-running operation;
- the audit event, usage events, evidence, and rollback reference.

No UI-only, CLI-only, or administrator-only mutation path may bypass the same
plan, policy, apply, audit, and evidence semantics.

## OCS product contract

An installable product package declares, with digests and compatibility ranges:

- service identity, ownership, versions, capabilities, dependencies, and regions;
- API schema and transport, workload identity, health, readiness, and discovery;
- lifecycle actions for validate, plan, provision, hold, resize, suspend, resume,
  deprovision, export, retry, cancel, and status where applicable;
- asynchronous operation, idempotency, rollback, cleanup, and failure semantics;
- IAM permissions, tenant and project scoping, quota dimensions, placement, and
  availability objectives;
- catalog plans, limits, billing meters, usage delivery, corrections, and evidence;
- support, diagnostics, durability, backup, restore, deletion, and data exit;
- signed, sandboxed microfrontend modules that use only the provider-shell SDK;
- moderation state and the provider decision that permits publication.

Unknown required fields, unsupported major versions, digest mismatches, and
incompatible lifecycle changes fail closed. Minor-version evolution must be
backward compatible for the declared support window. A connector cannot silently
downgrade security, billing, durability, or tenant-isolation semantics.

The full remote connector protocol, generated multi-language SDKs, moderation
runtime, and dynamic microfrontend host are later deliverables. Current packages
must nevertheless preserve this boundary.

## Durable control and billing events

Acknowledged control-plane state is durable and concurrency-safe. Schema changes
use expand/contract migrations that support mixed application versions during a
progressive rollout. An accepted write is never rolled back to an older state
backend after subsequent writes have occurred.

Audit records are append-only. Product control and usage events are idempotent,
replayable, and designed for a transactional-outbox boundary so a future billing
backend can be replaced without losing or double-charging usage.

A usage event can identify its producer service, consuming service, subject
resource, tenant/project, provider/cell/region, meter, quantity, observation time,
idempotency key, correction target, and signature/evidence policy. This supports
service-on-service infrastructure charging and later cross-provider settlement
without changing product APIs.

## Provider experience and product experience

The provider shell supplies consistent navigation, identity context, project and
region selection, accessibility, error handling, operation progress, support, and
cost visibility. Product teams ship independently versioned microfrontends and
APIs without modifying the shell. The shell loads only admitted packages and shows
only products, plans, regions, and actions allowed for the current identity.

Provider administrators can approve, suspend, target, limit, price, and withdraw a
product through versioned policy. Availability may be global, provider-wide,
region-specific, tenant-specific, invite-only, or disabled, and is always explicit.

## Availability and change safety

The control plane, identity path, product registry, billing ingestion, provider
shell, state stores, secret path, and reconciliation path declare replicas, failure
domains, stable endpoints, failover, retry behavior, RPO/RTO, backup coverage, and
evidence. A component is not called highly available merely because it has multiple
pods.

Supported updates preserve acknowledged operations and stay within the declared
SLO. Database and API changes must be compatible across the rollout window.
Progressive delivery, health gates, rollback to the prior signed revision, and a
tested restore path are required before a production mutation can be promoted.

## Federation boundary

Federation is a peer protocol, not a mandatory central service. Independent
providers retain their identities, policy, data, catalogs, customers, and business
relationships. Peers exchange versioned and signed trust, catalog, entitlement,
usage, settlement, and portability artifacts through pull-based reconciliation.

A peer outage or federation partition cannot take down local service. Providers
choose which products and regions to publish or consume. Cross-provider access is
short-lived, scoped, auditable on both sides, and revocable. No CloudRING vendor,
jurisdiction, database, message bus, or marketplace is a permanent root of control.

The federation runtime and economy are later goals. Single-cell work is accepted
only when it does not introduce an incompatible identifier, event, operation, or
service boundary that would require a second platform implementation later.

## Goal 01 binding subset

Goal 01 binds the following foundations now:

- the stable identity namespace and canonical operation/audit identity;
- strict, additive state migration and append-only audit behavior;
- declared critical-path survivability for one three-server cell;
- public reusable backup, restore, HA, and reconciliation contracts separated
  from deployment-specific values and live evidence;
- explicit non-claims for multi-cell, multi-region, marketplace, and federation.

Later goals implement the remote connector protocol, product admission, regional
offers, quotas, billing, marketplace, service-on-service economics, multi-cell
operation, and provider federation as independently accepted vertical releases.
