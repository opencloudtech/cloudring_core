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
| 01 | Hub survivability: one-cell critical path survives any one server loss, restores off-cell, reconciles through GitOps, and keeps acknowledged portal state. |
| 02 | One-command, API-only installer for one-to-three-node providers. |
| 03 | Unified observability, SLO, notification, and tamper-evident audit core. |
| 04 | Progressive zero-loss upgrade train and self-healing. |
| 05 | Durable provider control plane and the versioned OCS product seam. |
| 06 | Provider-grade ID, IAM, workload identity, recovery, and audited break-glass. |
| 07 | Complete billable compute product as the first reference connector. |
| 08 | Durable metering, rating, financial ledger, invoicing, budgets, and interoperable cost export. |
| 09 | IPv6-first tenant networking product. |
| 10 | Hosted-control-plane tenant Kubernetes product. |
| 11 | Remote OCS runtime, SDK, conformance, microfrontends, and generated client surfaces. |
| 12 | Object storage and managed PostgreSQL dogfood the generic product lifecycle. |
| 13 | Agent-native operation and development surfaces with the same policy path as humans. |
| 14 | Moderated marketplace, entitlements, offers, and independent developer economics. |
| 15 | Multi-cell scale and disaster recovery with locally stable data planes. |
| 16 | Peer provider federation, remote services, cross-cloud networking, and signed settlement. |
| 17 | Provider 1.0 release gate: final security review and fixes, LTS, conformance, benchmarks, documentation, and governance. |

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
8. Broad security review is Goal 17. Every earlier goal still fixes correctness and
   safety defects needed for its own function and acceptance.

The active contract is
[`GOAL-01-hub-survivability.md`](GOAL-01-hub-survivability.md), with exact requirement
ownership in [`COVERAGE.md`](COVERAGE.md).
