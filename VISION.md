# CloudRING Vision

## The north star

CloudRING's ambition is to make cloud computing independent of any single
provider, company, technology stack, or jurisdiction.

The desired end state is a global network of independently operated private and
public clouds that share open technology and interoperable service contracts.
Users can choose where a service runs, combine services from different
providers, and leave a provider without rebuilding every integration. Providers
remain autonomous, compete on service quality, and collaborate on the platform
capabilities that do not differentiate them.

No single operator should possess a global kill switch. That does not make a
deployment exempt from law: each provider remains accountable in its own
jurisdictions, and every federation relationship is explicit and policy
controlled.

## The problem

Today's cloud model often combines four different choices into one package:

1. the infrastructure provider;
2. the cloud control plane;
3. the catalog of available services;
4. the legal and commercial jurisdiction governing them.

That vertical integration can be convenient, but it also concentrates failure,
policy, and switching power. Users accumulate provider-specific APIs and data
paths. Smaller providers and enterprise infrastructure teams repeatedly build
similar control planes. Independent service developers face a separate
integration and distribution problem for every cloud.

CloudRING separates these concerns.

## The product model

CloudRING is a cloud operating system with two durable layers:

- **The provider platform** supplies identity, tenancy, IAM, policy, APIs,
  portal, catalog, billing foundations, audit, installation, lifecycle, and
  operations.
- **Cloud products** supply compute, networking, storage, databases, queues,
  Kubernetes, AI infrastructure, or any other customer-facing capability.

The platform consumes products through OCSv3 contracts. It must not import a
service's private implementation, assume its programming language, or give a
built-in service an undocumented privilege unavailable to another conforming
module.

Infrastructure used by the platform itself is an implementation dependency;
infrastructure offered to a tenant is a cloud product. Keeping that boundary
clear allows technology choices to evolve without forcing the business or the
ecosystem to start over.

## Open Cloud Standard

OCSv3 is intended to make a cloud service independently buildable, testable,
installable, operable, and replaceable. It describes the complete product
surface—not only provisioning:

- identity, tenancy, policy, and service-to-service access;
- synchronous APIs, asynchronous events, idempotent lifecycle operations, and
  compatibility;
- portal and automation extensions;
- catalog, entitlement, quota, capacity, meters, and billing inputs;
- observability, support, evidence, and incident boundaries;
- durability, backup, restore, upgrade, rollback, export, and deletion;
- distribution, signing, moderation, federation, and commercial metadata.

The specification must remain language-neutral even when CloudRING ships a Go
reference implementation. Conformance must include negative and failure cases,
not only schema acceptance. OCSv3 remains a CloudRING project specification
until any external standards status is established separately.

## A provider and developer economy

The common platform should reduce duplicated engineering, not eliminate
competition.

Providers can use the open core, contribute generic improvements, build their
own adapters or services, and select which third-party products they admit.
Their competitive advantage comes from reliability, location, support,
compliance, cost, and services that solve customer problems well.

Independent developers can build a service for one provider, a group of private
clouds, or—when federation and distribution are ready—a wider CloudRING network.
A service may be open source or independently licensed. Publication, pricing,
licensing, and revenue sharing are governed by explicit provider and module
policy rather than hidden platform assumptions.

Users gain choice only when exit works in practice. Products therefore need
declared data export, deletion, compatibility, and migration behavior. A common
API does not by itself prove that workloads or data can move safely.

## Private, public, edge, and federated operation

CloudRING should support the same conceptual platform in several settings:

- an enterprise private cloud that unifies internal infrastructure and approved
  external services;
- a regional or specialist public provider that needs a complete open cloud
  foundation;
- edge or constrained sites that run a policy-approved subset of products;
- a federation in which independent providers exchange signed catalogs,
  offers, orders, entitlements, usage summaries, and settlement records without
  sharing a database or root administrator.

The first useful deployment is deliberately smaller: one provider, one region,
one bounded cell, and a complete evidence-backed customer journey. Federation
comes after the single-provider platform is secure, durable, operable, and
portable.

## Design principles

1. **No global central dependency.** Federation must continue to fail safely
   when another provider or coordinator is unavailable.
2. **Open core, replaceable edges.** Reusable capabilities belong upstream;
   deployment values and independently owned modules stay with their owners.
3. **First-class by contract.** Every service uses the same documented security,
   lifecycle, billing, support, and evidence surfaces.
4. **Provider and technology neutrality.** Core APIs describe capabilities,
   not one vendor's products or one deployment topology.
5. **Exit is a product feature.** Export, deletion, compatibility, and recovery
   are designed and tested, not deferred to documentation.
6. **Evidence before claims.** Static configuration and green unit tests do not
   prove a live installation, upgrade, recovery, or failure boundary.
7. **Local autonomy and lawful operation.** Providers retain admission,
   commercial, security, and jurisdictional control.
8. **One control surface.** Humans, CLIs, APIs, automation, and AI agents use the
   same policy-controlled operations and audit trail.
9. **Evolution without ecosystem rewrites.** Implementations may change behind
   stable versioned contracts.
10. **Solve user problems.** The platform is infrastructure; value comes from
    reliable services that help users accomplish real work.

## What success means

CloudRING succeeds when all of the following are demonstrated from released
artifacts and clean installations:

- independent providers can install and operate the same open platform without
  a private upstream dependency;
- independent product teams can ship conforming services without modifying the
  platform core;
- users receive consistent IAM, lifecycle, billing, support, and evidence
  across products;
- supported upgrades and declared failures preserve the promised service and
  data boundaries;
- users can export or move supported workloads and data through tested paths;
- providers can federate without a shared root, mandatory central coordinator,
  or universal database;
- no documentation, fixture, or historical deployment is presented as current
  proof of readiness.

This document states direction. The [public roadmap](roadmap/README.md) defines
the sequence and the evidence required before each part of that direction can
become a product claim.
