# Goal Coverage

This matrix records exact public requirements owned by the active goal. `accepted`
requires the final public SHA plus fresh downstream evidence; local tests or a live
observation alone do not change status.

The stable definitions and acceptance meaning of every identifier below are in
`specifications/goal-01.md`. The matrix records delivery state; it does not redefine
or weaken those requirements.

## Goal 01 — Reference-cell critical-path survivability

| Requirement | Goal 01 acceptance slice | Status | Evidence |
| --- | --- | --- | --- |
| CR-FND-040 | Replaceable state, backup, and reconciliation boundaries | in_progress | pending final public SHA and downstream receipts |
| CR-FND-050 | Public reusable core separated from concrete provider deployment | in_progress | pending boundary audit |
| CR-FND-120 | No production-readiness claim beyond the proved one-cell envelope | in_progress | pending final gate |
| CR-FND-130 | Evidence before readiness; missing/stale stays blocked | in_progress | pending final gate |
| CR-FND-140 | Canonical operation identity binds state and audit probes | in_progress | pending durable-state proof |
| CR-STO-080 | Platform backup control plane for declared Goal 01 classes | in_progress | pending strict live barrier |
| CR-STO-090 | Fresh signed backup barrier before every mutation wave | in_progress | pending strict live barrier |
| CR-STO-100 | Isolated restore drill blocks promotion | in_progress | pending restore and cleanup receipt |
| CR-STO-110 | RPO/RTO declared for every critical state class | in_progress | pending survivability inventory |
| CR-STO-130 | etcd off-cell snapshot and exact restore | in_progress | pending restore receipt |
| CR-STO-140 | Fresh, fail-closed durability evidence states | in_progress | pending final gate |
| CR-STO-150 | Storage-dependent node drains interlocked with the barrier | in_progress | pending server-loss drill |
| CR-STO-160 | Offsite immutable copies and retention-delete denial | in_progress | pending object-store proof |
| CR-K8S-020 | Three-member control-plane continuity under one-server loss | in_progress | pending sequential drill |
| CR-K8S-030 | etcd backup and restore | in_progress | pending exact restore receipt |
| CR-K8S-090 | Declared cluster add-ons owned and reconciled through GitOps | in_progress | pending zero-unmanaged audit |
| CR-K8S-130 | Node lifecycle operations preserve workload continuity | in_progress | pending sequential drill |
| CR-OCS-140 | Durability surfaces and restore objectives remain explicit | in_progress | pending inventory and restore proof |
| CR-OPS-090 | Dated pre-state, rollback, abort, post-state, and cleanup per wave | in_progress | pending final wave receipts |
| CR-OPS-180 | Disaster-recovery drill calendar starts with a real restore | in_progress | pending restore drill |
| CR-DPL-040 | Dry-run-first desired state reconciles controlled drift | in_progress | pending ownership/drift proof |
| CR-DPL-050 | Reference environment self-reconciles from its signed Git source | in_progress | pending Flux source proof |
| CR-DPL-150 | Mixed-version-safe database migration discipline | in_progress | pending PostgreSQL cutover proof |

Requirements for the local or remote OCS runtime, dynamic microfrontends, product
moderation, offers, entitlements, quota, billing, marketplace, multi-cell operation,
and federation are not owned by Goal 01 and remain pending for later roadmap goals.
