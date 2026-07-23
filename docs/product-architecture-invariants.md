# CloudRING Product Architecture Invariants

CloudRING is a provider-neutral cloud platform and the reference implementation
of the project-defined Open Cloud Standard (OCS). These invariants keep the
reference cell, future products, and future provider federation compatible without
making later capabilities part of the current release or claiming OCS 1.0 before
product dogfood and the final release gate.

## Platform and product boundary

The platform owns identity, authorization, policy, product admission, the
resource and operation model, catalog composition, usage transport, billing
integration, the provider shell, and evidence. A cloud product owns its service
implementation, lifecycle controller, support diagnostics, durability behavior,
and any product-specific API or user experience it chooses to expose.

Products integrate only through versioned OCS contracts. A local product may run
inside the provider cell behind a local adapter; a remote product may run in another
cell or independently operated installation behind the remote connector protocol.
An API-only product may expose no Kubernetes binding and no provider-shell module.
All three execution profiles expose the same versioned public product API and the
same resource, lifecycle, identity, policy, operation, usage, and evidence semantics.
Location and presentation are deployment choices, not separate product models. The
core must not import product implementation details or require a product to use
Kubernetes, Go, or the platform's storage internally.

Management-plane dependencies such as Kubernetes, networking, certificates,
secret management, databases, and backup are substrate components. A
tenant-visible compute, network, storage, database, license, external-account,
or other offering is an OCS product even when maintained by the platform team.

## Stable identity namespace

The built-in identity service is a complete provider account system rather than
only a proxy to an external identity provider. It can issue and manage provider
identities itself or federate with external identity providers. Hierarchical IAM
applies the same organization, tenant, project, resource, workload, and
administrator policy model to every human, service, CLI, portal, and agent path.

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
- exactly one execution profile: `local`, `remote`, or `api-only`;
- the mandatory versioned public product API and OCS transport, identity, health,
  readiness, and discovery contracts, plus any optional product-specific actions;
- applicability for the universal lifecycle vocabulary `provision`, `hold` or
  `suspend`, `resume`, `resize`, and `deprovision`; every action is either supported
  with an idempotent API contract or `not_applicable` with a reason;
- asynchronous operation, idempotency, rollback, cleanup, and failure semantics;
- IAM permissions, tenant and project scoping, quota dimensions, placement, and
  availability objectives;
- catalog plans, limits, billing meters, usage delivery, corrections, and evidence;
- support, diagnostics, durability, backup, restore, deletion, and data exit;
- any optional signed, sandboxed microfrontend modules, which use only the
  provider-shell SDK;
- provider moderation state and the explicit provider decision that admits,
  targets, suspends, or withdraws the product.

The execution profile controls only profile-specific surfaces. `local` may declare
Kubernetes bindings; `remote` declares a remote endpoint and trust boundary;
`api-only` needs neither Kubernetes bindings nor a microfrontend. A declared
microfrontend must be signed, integrity-pinned, sandboxed, permission-scoped, and
revocable. Federation and commercial metadata are explicit applicability profiles:
each declares `supported` or `not_applicable` with a reason; complete metadata is
validated only when supported. They are never prerequisites for a private,
non-commercial, or non-federated product. Remote and API-only packages use
source-safe endpoint, trust-policy, and health references rather than raw endpoints
or credentials.

Every service-to-service dependency binds a capability class to a target public
product API, a compatible version range, and a provider-resolved compatibility
policy. The dependency never embeds an implementation endpoint. This lets the
provider select an admitted local or remote product and attribute the consuming
service's infrastructure usage without changing either product's implementation.

Unknown required fields, unsupported major versions, digest mismatches, and
incompatible lifecycle changes fail closed. Minor-version evolution must be
backward compatible for the declared support window. A connector cannot silently
downgrade security, billing, durability, or tenant-isolation semantics.

The full local and remote connector runtime, generated multi-language SDKs,
moderation runtime, and dynamic microfrontend host are later deliverables. Current
packages must nevertheless preserve this boundary.

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

Marketplace settlement can additionally identify the offer, seller, provider,
platform share, seller share, currency, correction, and signed source usage. Revenue
share is derived from the same durable ledger as customer charging; it is not a
separate best-effort counter.

## Provider experience and product experience

The provider shell supplies consistent navigation, identity context, project and
region selection, accessibility, error handling, operation progress, support, and
cost visibility. Product teams ship independently versioned microfrontends and
APIs without modifying the shell. The shell loads only admitted packages and shows
only products, plans, regions, and actions allowed for the current identity.

Provider administrators can approve, suspend, target, limit, price, and withdraw a
product through versioned policy. Availability may be global, provider-wide,
region-specific, tenant-specific, invite-only, or disabled, and is always explicit.

The first placement release binds each resource to one provider-declared region.
Cross-region placement, evacuation, and failover are later capabilities and must not
be inferred from region-shaped identifiers.

Installation and product contracts describe capabilities and failure domains rather
than a hosting brand. The same public core must remain deployable in a private
datacenter or through an independent infrastructure provider; provider-specific
credentials, topology, pricing, and policy stay in downstream adapters and profiles.

## Availability and change safety

The control plane, identity path, product registry, billing ingestion, provider
shell, state stores, secret path, and reconciliation path declare replicas, failure
domains, stable endpoints, failover, retry behavior, RPO/RTO, backup coverage, and
evidence. A component is not called highly available merely because it has multiple
pods. This inventory is cumulative: every later critical component must join the
same one-failure, off-cell-restore, and acknowledged-state continuity gate before its
goal is accepted.

Every declared supported update path is zero-downtime. CloudRING uses rolling
delivery where versions can safely coexist and blue-green delivery where isolation
or an atomic traffic switch is required. Zero-downtime means no lost acknowledged
operation and no user-visible interruption on the declared supported path, while
continuous control, identity, and billing probes remain successful. Database and API
changes must be compatible across the mixed-version rollout window. Health gates,
rollback to the prior signed revision, and a tested restore path are required before
a production mutation can be promoted.

## Federation boundary

Federation is a decentralized sovereign peer protocol, not a mandatory central
service. It has no global coordinator, mandatory marketplace, or kill switch.
Independent providers retain their identities, policy, data, catalogs, customers,
and business relationships. Peers exchange versioned and signed trust, catalog,
entitlement, usage, settlement, and portability artifacts through pull-based
reconciliation.

A peer outage or federation partition cannot take down local service or remove a
provider's local administrative autonomy. Providers choose which products and
regions to publish or consume. Cross-provider access is short-lived, scoped,
auditable on both sides, and revocable. No CloudRING vendor, jurisdiction, database,
message bus, or marketplace is a permanent root of control.

Admitted remote products may appear in the same provider shell only through the
local provider's catalog and policy decision and always retain their provider and
jurisdiction identity. Later capacity exchange may let a private cloud place
explicitly authorized peak workloads with a peer, but never silently overrides
residency, cost, tenant policy, or user intent.

The federation runtime and economy are later goals. Single-cell work is accepted
only when it does not introduce an incompatible identifier, event, operation, or
service boundary that would require a second platform implementation later.

## Legacy Goal 01 obligations in the canonical roadmap

The former Goal 01 bound the following foundations. They are now carried by the
single G00-G27 delivery graph rather than a competing roadmap:

- the stable identity namespace and canonical operation/audit identity;
- strict, additive state migration and append-only audit behavior;
- declared critical-path survivability for one three-server reference cell in one
  region;
- public reusable backup, restore, HA, and reconciliation contracts separated
  from deployment-specific values and live evidence;
- explicit non-claims for whole-cell availability, multi-cell, multi-region,
  marketplace, remote-product runtime, and federation.

Later goals deliver an OCS release candidate with local and remote runtime and
provider moderation before dependent products; then identity, offers, entitlements,
one-region placement, billing, product dogfood, portability, marketplace economics,
multi-cell operation, and sovereign peer federation as independently accepted
vertical releases. OCS 1.0 is reserved for G27 after product dogfood and the
final security review, fixes and full regression.
