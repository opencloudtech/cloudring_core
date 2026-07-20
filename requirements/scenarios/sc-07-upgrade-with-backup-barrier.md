# SC-07 — Upgrade with backup barrier

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the mutation-gated upgrade path: a platform release advances to
production only through the promotion ladder, only from an immutable tag,
and only after the backup control plane has issued a fresh, signed
backup-barrier receipt for every stateful resource the upgrade touches —
and prove that a missing, stale, forged, or failed-backup receipt halts
the upgrade closed.

## Actors

- operator / release owner — drives the upgrade
- provider — owns the installation
- agent — may execute governed steps (see SC-10)
- auditor — replays barrier decisions from the append-only record

## Preconditions

- The installation runs a prior release; the candidate release is tagged,
  signed, and carries a BOM plus lower-stage promotion receipts
  (CR-DPL-100, CR-DPL-070).
- A change record with dated evidence and a rollback plan exists, and a
  maintenance window is registered where required (CR-OPS-090,
  CR-OPS-100).
- Fresh restore-drill evidence exists for in-scope stateful services (see
  SC-05; CR-STO-100).

## Steps

1. **Open the change record.** The upgrade is registered with scope,
   risk class, dated evidence, and a rollback plan.
   - **Expected outcome:** the change record links the release tag, the
     target environment, and the rollback path; the maintenance window
     interlocks with the change register.
   - **Requirements:** CR-OPS-090, CR-OPS-100

2. **Verify ladder eligibility.** The candidate holds promotion receipts
   from every lower stage.
   - **Expected outcome:** any production-targeted change lacking
     lower-stage receipts or carrying uncommitted state is halted.
   - **Requirements:** CR-DPL-070

3. **Verify backup freshness across the blast radius.** For every
   stateful resource the upgrade touches, a completed backup exists inside
   the mutation's freshness window.
   - **Expected outcome:** backup completion evidence is queryable per
     resource; any failed or stale backup blocks the affected scope.
   - **Requirements:** CR-STO-080, CR-STO-140

4. **Take a fresh etcd snapshot.** Control-plane state is snapshotted
   before any control-plane component upgrades.
   - **Expected outcome:** the snapshot completes with integrity
     verification; no control-plane upgrade or member replacement proceeds
     without it.
   - **Requirements:** CR-K8S-030, CR-STO-130

5. **Issue signed backup-barrier receipts.** The backup control plane's
   workload identity signs receipts binding resource identity, backup
   reference with integrity hash, and a UTC timestamp.
   - **Expected outcome:** each in-scope resource holds a machine-
     checkable receipt within the freshness window; the receipts are
     recorded append-only.
   - **Requirements:** CR-STO-090

6. **Dry-run the apply.** The GitOps apply path produces a recorded
   dry-run receipt.
   - **Expected outcome:** unexpected destroy/replace of stateful
     resources in the dry-run halts the apply; the dry-run receipt is
     attached to the change record.
   - **Requirements:** CR-DPL-040

7. **Gate on the barrier.** The mutation gate verifies each receipt
   before mutating its resource.
   - **Expected outcome:** missing, stale, unsigned, forged, or
     failed-backup receipts deny the mutation closed; any bypass attempt
     is treated as a security event.
   - **Requirements:** CR-STO-090

8. **Roll out progressively.** The release deploys canary-by-default
   with health-gated promotion between waves.
   - **Expected outcome:** rollout is atomic with bounded history and
     automatic rollback on failure; canary analysis uses declared golden
     signals.
   - **Requirements:** CR-DPL-140, CR-DPL-100, CR-OBS-030

9. **Run database migrations under stage discipline.** Schema migrations
   execute as a distinct, ordered stage.
   - **Expected outcome:** a failed migration stage blocks deploy stages
     automatically; migrations are reversible or carry an explicit,
     evidenced irreversibility note.
   - **Requirements:** CR-DPL-150

10. **Upgrade the Kubernetes layer within policy.** Cluster versions
    advance within the declared support and skew windows.
    - **Expected outcome:** the upgrade follows the version support
      policy; component version skew stays inside the compatibility
      window; node upgrades preserve workload continuity with evidence.
    - **Requirements:** CR-K8S-060, CR-OPS-150, CR-K8S-130

11. **Upgrade managed data-service engines behind their own barrier.**
    In-place engine version upgrades proceed only with per-cluster backup
    evidence.
    - **Expected outcome:** an engine upgrade without fresh backup
      evidence is refused; the single-active-operation invariant per
      cluster is preserved.
    - **Requirements:** CR-DAT-090, CR-DAT-020

12. **Exercise the negative path.** An upgrade attempt is made against
    one resource with a stale (or forged) barrier receipt.
    - **Expected outcome:** the gate denies the mutation, records a
      `blocked` evidence state, and the attempt is visible in the audit
      trail; the upgrade wave does not silently skip the resource.
    - **Requirements:** CR-STO-090, CR-FND-130, CR-IAM-150

13. **Verify post-upgrade health.** The connectivity matrix regenerates;
    observability evidence is re-linked; tenant-facing states are
    consistent across surfaces.
    - **Expected outcome:** a green post-change matrix exists; golden
      signals return to baseline; no surface reports divergent state.
    - **Requirements:** CR-NET-190, CR-OBS-210, CR-FND-140

14. **Close the change record.** Evidence links are attached and the
    rollback plan is stood down or retained per policy.
    - **Expected outcome:** the change-record and post-incident evidence
      trail is complete and replayable; barrier decisions are replayable
      from the append-only record.
    - **Requirements:** CR-OPS-090, CR-DPL-200, CR-STO-090

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-STO-080 | Platform backup control plane | 3 |
| CR-STO-090 | Signed backup barrier before any mutation | 5, 7, 12, 14 |
| CR-STO-100 | Restore drills as promotion blockers | preconditions |
| CR-STO-130 | Control-plane state (etcd) snapshot and restore | 4 |
| CR-STO-140 | Durability evidence states and freshness | 3 |
| CR-K8S-030 | etcd backup and restore | 4 |
| CR-K8S-060 | Version support and upgrade policy | 10 |
| CR-K8S-130 | Node lifecycle operations with workload continuity | 10 |
| CR-DAT-020 | Durable task queue with dependencies and single-active-operation invariant | 11 |
| CR-DAT-090 | In-place engine version upgrades gated by backup evidence | 11 |
| CR-DPL-040 | Everything-as-code with dry-run-first GitOps reconciliation | 6 |
| CR-DPL-070 | Environment ladder and promotion regulation | 2 |
| CR-DPL-100 | Tag-only production releases with atomic rollout | 8 |
| CR-DPL-140 | Canary-by-default progressive delivery | 8 |
| CR-DPL-150 | Database-migration stage discipline | 9 |
| CR-DPL-200 | Change-record and post-incident evidence trail | 14 |
| CR-OPS-090 | Change management with dated evidence and rollback plans | 1, 14 |
| CR-OPS-100 | Maintenance windows with change-register interlock | 1 |
| CR-OPS-150 | Version-skew and compatibility-window policy | 10 |
| CR-NET-190 | Network connectivity evidence as readiness gate | 13 |
| CR-OBS-030 | Golden signals and outcome-aware metrics | 8 |
| CR-OBS-210 | Observability evidence as a readiness gate | 13 |
| CR-IAM-150 | Immutable, queryable, SLA-bound audit log | 12 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 12 |
| CR-FND-140 | One product truth across surfaces | 13 |

## Gaps found

None identified. Barrier issuance, gate enforcement, ladder discipline,
progressive delivery, and the negative path all map to existing
requirements. (CR-STO-090 records that barrier coverage across all
mutation paths is incomplete — the scenario scope is limited to the
mutation classes where the gate is implemented, and any uncovered class
must be declared in the change record as a non-claim.)

## Evidence required

- Change record with scope, risk class, dated evidence, rollback plan,
  and maintenance-window interlock (CR-OPS-090, CR-OPS-100).
- Backup freshness query results for the blast radius (CR-STO-080).
- Fresh etcd snapshot with integrity verification (CR-K8S-030).
- Signed backup-barrier receipts and the append-only barrier decision
  record, replayable by an auditor (CR-STO-090).
- Dry-run receipt attached to the change record (CR-DPL-040).
- Canary rollout records with health-gate decisions and, where exercised,
  automatic rollback evidence (CR-DPL-140, CR-DPL-100).
- Migration-stage logs and failure-blocking proof (CR-DPL-150).
- Gate test evidence for missing/stale/forged/failed-backup denial cases
  (CR-STO-090).
- Post-upgrade connectivity matrix and observability gate report
  (CR-NET-190, CR-OBS-210).
- Closed change record with the full evidence trail (CR-DPL-200).
