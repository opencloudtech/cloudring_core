# CloudRING Public Roadmap

This roadmap builds a provider-neutral, modular cloud platform in seventeen
sequential, evidence-gated goals. Each accepted goal must leave the public project
installable and the reference deployment coherent. Future capabilities do not waive
the acceptance criteria of the current goal.

The product architecture is governed by
[`docs/product-architecture-invariants.md`](../docs/product-architecture-invariants.md).
OCS is the sole product seam; all generic runtime, contracts, SDKs, tests, and
runbooks land in this Apache-2.0 repository. Concrete provider values, credentials,
private topology, tenant records, and live receipts stay in the relevant downstream
deployment repository.

## Sequence

| Goal | Outcome |
| --- | --- |
| 01 | Reference-cell critical-path survivability: the declared one-region critical path survives any one server loss, restores off-cell, reconciles through GitOps, and keeps acknowledged portal state. |
| 02 | Provider-neutral, one-command API installer for one-to-three-node private-datacenter or infrastructure-provider cells. |
| 03 | Unified observability, SLO, notification, and tamper-evident audit core. |
| 04 | Supported zero-downtime rolling and blue-green upgrade train, mixed-version compatibility, rollback, and self-healing. |
| 05 | OCS release candidate: durable control plane, universal lifecycle baseline, common local/remote/API-only connector runtime, provider moderation, conformance SDKs, mandatory versioned product APIs, and optional signed sandboxed microfrontends. |
| 06 | Provider-grade ID, IAM, workload identity, recovery, and audited break-glass. |
| 07 | Durable product lifecycle, orders, subscriptions, offers, plans, entitlements, quotas, catalog targeting, and initial one-region placement. |
| 08 | Durable metering, rating, financial ledger, invoicing, budgets, service-on-service charging, and interoperable cost export. |
| 09 | Foundational IPv6-first network, volume, and image/artifact products required by higher-level services. |
| 10 | Complete billable compute product as the first composite reference connector. |
| 11 | Hosted-control-plane tenant Kubernetes product. |
| 12 | Object storage and managed PostgreSQL dogfood the generic product lifecycle. |
| 13 | Verified local/remote connector dogfood plus product and workload portability across a private datacenter and an independent infrastructure provider. |
| 14 | Moderated marketplace with seller onboarding, offers, settlement, and auditable provider/platform/developer revenue share. |
| 15 | Multi-cell scale and disaster recovery with locally stable data planes. |
| 16 | Decentralized sovereign peer federation, remote services, cross-cloud networking, and signed settlement without a central coordinator or kill switch, preserving local autonomy. |
| 17 | OCS and provider 1.0 release gate: final security review and fixes, LTS, conformance, benchmarks, documentation, and governance. |

## Execution rules

1. Work one goal at a time. Parallel work is allowed only inside the active goal.
2. The goal contract is corrected before code when verified reality contradicts it.
3. Public reusable behavior lands before a downstream pins and deploys it.
4. A change is not accepted from fixtures alone: it needs tests, a public artifact,
   downstream reconciliation, and fresh scoped evidence where live acceptance applies.
5. Off-cell backup, a restore proof, a rollback boundary, and exact approval precede
   destructive or production mutations.
6. Blocked and unverified are valid outcomes but never readiness claims.
7. Replaced production paths are removed after cutover; dual implementations do not
   become permanent compatibility layers.
8. Human and agent operations share the same API, IAM, policy, plan/apply, audit, and
   evidence path in every goal; agent parity is cross-cutting, not a late control plane.
9. Broad security review is Goal 17. Every earlier goal still fixes correctness and
   safety defects needed for its own function and acceptance.
10. Survivability is cumulative. Each newly accepted critical identity, registry,
    lifecycle, billing, product, marketplace, or operations component expands the
    critical-path inventory and must prove failover, off-cell restore, and zero loss
    of acknowledged state before its goal is accepted.

The active contract is
[`GOAL-01-hub-survivability.md`](GOAL-01-hub-survivability.md), with exact requirement
ownership in [`COVERAGE.md`](COVERAGE.md).
