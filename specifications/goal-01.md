# Goal 01 Public Requirements Specification

These requirements define the provider-neutral behavior owned by Goal 01. They do
not contain provider credentials, topology values, tenant data, or live evidence.
`MUST` is release-blocking for the declared Goal 01 scope.

## Platform foundation

| ID | Normative requirement |
| --- | --- |
| `CR-FND-040` | State, backup, reconciliation, and provider integrations MUST use replaceable, versioned capability boundaries rather than private implementation coupling. |
| `CR-FND-050` | Reusable platform behavior, contracts, SDKs, tests, and runbooks MUST live in the public core; concrete provider configuration and live receipts MUST remain downstream. |
| `CR-FND-120` | Readiness claims MUST be limited to the explicitly proved cell, critical path, failure model, and observation window. |
| `CR-FND-130` | Missing, stale, unsigned, mismatched, or blocked evidence MUST fail closed and MUST NOT be converted into a readiness claim. |
| `CR-FND-140` | A canonical non-secret operation identity MUST bind accepted state changes to API results, audit events, durability probes, evidence, and rollback. |

## Storage, backup, and recovery

| ID | Normative requirement |
| --- | --- |
| `CR-STO-080` | The platform MUST back up every declared Goal 01 state class through one observable backup control plane. |
| `CR-STO-090` | Every production mutation wave MUST require a fresh, signed backup barrier for its exact blast radius. |
| `CR-STO-100` | Promotion MUST be blocked until an isolated restore drill proves exact object versions, checksums, application consistency, tenant isolation, and cleanup. |
| `CR-STO-110` | Every critical state class MUST declare measurable RPO, RTO, owner, backup location class, and restore procedure. |
| `CR-STO-130` | The etcd member set MUST have off-cell snapshots and a tested exact restore procedure bound to the accepted cluster identity. |
| `CR-STO-140` | Durability evidence MUST be fresh and machine-readable and MUST distinguish ready, blocked, stale, failed, and not-applicable states. |
| `CR-STO-150` | A storage-dependent drain or server-loss wave MUST be interlocked with the backup barrier, health checks, abort conditions, and recovery result. |
| `CR-STO-160` | At least one accepted copy MUST be off-cell and retention-protected; the restore drill MUST prove deletion denial for the protected window. |

## Kubernetes continuity

| ID | Normative requirement |
| --- | --- |
| `CR-K8S-020` | The three-member upstream Kubernetes control plane and etcd quorum MUST survive sequential loss of any one server within the declared SLO. |
| `CR-K8S-030` | Kubernetes recovery MUST include a tested etcd backup and restore path, not only workload manifests. |
| `CR-K8S-090` | Every declared cluster add-on and production root MUST be owned by the accepted GitOps inventory or explicitly blocked; the accepted Goal 01 state has no unmanaged allowlist. |
| `CR-K8S-130` | Node lifecycle operations MUST preserve the declared workload and acknowledged-state continuity contract and MUST stop on a failed probe. |

## Product durability and operations

| ID | Normative requirement |
| --- | --- |
| `CR-OCS-140` | Every stateful OCS product MUST declare its data classes, backup policy, RPO/RTO, restore-test objective, and evidence gate before enablement. |
| `CR-OPS-090` | Every mutation wave MUST record sanitized pre-state, exact intended scope, rollback and abort triggers, post-state, and cleanup result. |
| `CR-OPS-180` | The disaster-recovery schedule MUST begin with a successful real isolated restore and define its next due date and owner. |

## Delivery and reconciliation

| ID | Normative requirement |
| --- | --- |
| `CR-DPL-040` | Desired-state changes MUST be dry-run-first and MUST prove reconciliation of a controlled, reversible drift fixture before prune or promotion. |
| `CR-DPL-050` | The reference installation MUST reconcile from an accepted signed Git revision and MUST fail closed on repository, revision, or inventory mismatch. |
| `CR-DPL-150` | Database changes MUST use expand/contract migrations compatible with mixed application versions and MUST preserve accepted writes across rollout and rollback. |

Goal 01 acceptance is governed by
[`roadmap/GOAL-01-hub-survivability.md`](../roadmap/GOAL-01-hub-survivability.md)
and current status by [`roadmap/COVERAGE.md`](../roadmap/COVERAGE.md).
