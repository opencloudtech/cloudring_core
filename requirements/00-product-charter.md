# 00 — Product Charter

## Vision

CloudRING is a distributed, open, and scalable cloud platform that separates the
*technology* of providing cloud services from the *business* of providing them.
It unifies public cloud providers, private cloud owners, service vendors,
resellers, EDGE operators, and end users in one peer-to-peer network — without
dependence on any single provider, technology, or jurisdiction.

CloudRING ships in two layers:

- **CloudRING OpenSource** (this repository, Apache-2.0): the complete platform
  substrate — installation, infrastructure pod, services pod, the Open Cloud
  Standard (OCS), and everything an independent company needs to build and run
  its own cloud provider business.
- **CloudRING Business** (outside this repository): commercial enterprise
  extensions, support, and federation settlement services.

## Ecosystem roles

| Role | Description |
|---|---|
| Provider | Operates a CloudRING installation and sells services to clients |
| Vendor | Develops products/services on OCS and publishes them to marketplaces |
| Reseller / integrator / agent | Sells and integrates on non-discriminatory terms |
| Private cloud proprietor | Runs CloudRING (optionally with enterprise extensions) in own DCs |
| EDGE operator | Runs small-footprint CloudRING zones connected or disconnected |
| ISV / service team | Builds first-class citizen services via OCS connectors |
| Tenant / client partner | Consumes services; "all clients are partners" |
| Platform contributors | Community members with equal rights in development and evolution |

## Principles

1. **Contract before technology.** No immutable components: every layer
   (virtualization, network, storage, runtime) is replaceable behind a
   contract. We abstract the provider's business from technology so it survives
   the 5–10 year technology cycle.
2. **Open Cloud Standard as the extensibility heart.** Third-party services are
   first-class citizens: a service team can register, announce capabilities,
   implement the mandatory lifecycle APIs, and sell through the platform
   without platform-team involvement.
3. **Real open baseline.** The OSS layer must be genuinely sufficient to deploy
   and operate a provider — not a teaser for the commercial tier.
4. **Evidence over claims.** Readiness, durability, security, and delivery
   claims require fresh, verifiable evidence. `blocked` and `unverified` are
   honest, non-promotable states.
5. **Fail closed.** Authentication, authorization, secrets, money, deletion,
   and exposure paths deny on error, ambiguity, or missing evidence.
6. **Secrets are never configuration.** Secrets are referenced, brokered, and
   rotated through an approved secrets workflow; never committed, never plain.
7. **Stateful is first-class.** Data durability, backup, restore, and migration
   are designed-in, with restore drills as readiness gates.
8. **Portability and jurisdiction freedom.** A tenant can exit, move, or
   federate; no vendor lock-in by design.
9. **Upstream first.** Platform runtime is Go-first on upstream Kubernetes
   semantics; no legacy lightweight-distribution defaults.
10. **Production honesty.** No hardcoded success, no fixture-only UI presented
    as production, no in-memory persistence for production state, no hidden
    cost/defaults/public exposure before user commit.

## Scope of this corpus

These requirements cover the full platform:

- Cloud Infrastructure Pod: compute/virtualization, network, storage,
  Kubernetes/containers, bare metal, images
- Cloud Services Pod: OCS service connectors, marketplace/catalog, IAM,
  billing/finops, observability, portal/UX, data services
- Platform operations: deployment/IaC (abstract + provider profiles), CI/CD,
  upgrades, SRE, support, agent-governed operations
- Federation: P2P data bus, global portal, cross-cloud connect, settlement

Out of scope for this corpus: CloudRING Business commercial components
(enterprise extensions, settlement implementation), though the OSS
*interfaces* they plug into are in scope.

## Definition of done for the platform

The platform is "ready to deploy and build upon" when:

1. Every P0 requirement has fresh linked evidence (`accepted` + verified).
2. A reference installation on at least one real provider is deployed entirely
   from IaC, upgraded, backed up, restored, and one-server-loss tested.
3. A third-party service team can onboard a new OCS service end-to-end using
   only public documentation and the SDK.
4. CI/CD gates (source-safety, security, conformance) are green and enforced.
5. The requirements registry has zero drift and zero `unverified` P0 items.
