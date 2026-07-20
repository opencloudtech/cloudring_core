# SC-05 — Backup and restore drill promotes a release

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the stateful promotion path: a release candidate containing stateful
services advances through the environment ladder **only** when every
stateful service holds fresh, verified restore-drill evidence inside its
declared RPO/RTO — and prove the negative path: stale, blocked, or
synthetic drill evidence stops the promotion and stays stopped.

## Actors

- operator / release owner — drives the promotion
- service-team — owns the stateful services and their durability profiles
- provider — operates the installation
- auditor — verifies evidence states and gate decisions

## Preconditions

- The reference installation runs with the platform backup control plane
  and content-addressed snapshot chunk store (CR-STO-080, CR-STO-070).
- Each stateful service publishes machine-readable RPO/RTO objectives in
  its durability profile (CR-STO-110); OCS services declare durability
  surfaces in their connector packages (CR-OCS-140).
- A disaster-recovery drill calendar exists and is current (CR-OPS-180).
- The release candidate is an immutable, tagged ref with a bill of
  materials and lower-stage promotion receipts (CR-DPL-100, CR-DPL-070).

## Steps

1. **Confirm scheduled backups are green.** The backup control plane's
   jobs for every in-scope stateful service complete inside their
   schedules.
   - **Expected outcome:** backup completions are recorded as first-class
     evidence with UTC timestamps; any failed job is visible as such.
   - **Requirements:** CR-STO-080, CR-STO-140

2. **Verify snapshot integrity.** Crash-consistent checkpoints with
   changed-block tracking land in the content-addressed chunk store.
   - **Expected outcome:** per-chunk checksums verify on write and on
     read-out; a corruption-injection test proves tampered chunks are
     detected before use.
   - **Requirements:** CR-STO-060, CR-STO-070

3. **Confirm offsite immutable copies.** Copies exist per the retention
   governance policy.
   - **Expected outcome:** copy evidence is fresh; retention and
     immutability rules are declared and verifiable.
   - **Requirements:** CR-STO-160

4. **Verify etcd snapshot freshness.** Control-plane snapshots for every
   cluster are current (default cadence at least every 6 hours).
   - **Expected outcome:** snapshot success evidence is recorded off-
     cluster with integrity metadata; a failed snapshot job blocks
     control-plane mutations.
   - **Requirements:** CR-K8S-030, CR-STO-130

5. **Execute the scheduled restore drill.** Real backups restore into an
   isolated environment; data integrity and service health are verified.
   - **Expected outcome:** drill receipts record restored-dataset
     integrity verification and health checks; drill results are recorded
     under the durability evidence states.
   - **Requirements:** CR-STO-100, CR-STO-140

6. **Check objectives against drill reality.** Measured restore behavior
   is compared to the declared RPO/RTO.
   - **Expected outcome:** backup schedule, storage topology, and drill
     cadence demonstrably satisfy the declared objectives; any service
     whose proven drill results are weaker than its marketed objective is
     halted from claiming the tighter objective.
   - **Requirements:** CR-STO-110

7. **Cover managed data services.** The database-as-a-service class
   engines restore from object-storage-backed backups with continuous
   archiving.
   - **Expected outcome:** engine restore drills verify data integrity;
     backup intervals are metered through the durable outbox.
   - **Requirements:** CR-DAT-080, CR-DAT-100

8. **Evaluate the promotion gate.** The ladder stage requires fresh
   `verified` restore-drill evidence per stateful service within the
   declared freshness window.
   - **Expected outcome:** services with fresh verified evidence pass;
     `blocked`, `stale`, and `synthetic` drill results are non-promotable
     states with typed reasons.
   - **Requirements:** CR-STO-100, CR-STO-140, CR-DPL-070

9. **Exercise the negative path.** A drill failure is injected (or a
   freshness window is allowed to lapse) for one service.
   - **Expected outcome:** the promotion gate rejects the release for that
     service; the failed drill alerts the owning team; the blocked
     evidence remains visible and is never laundered into a claim; the
     rest of the release is blocked only to the extent the failing service
     is in scope.
   - **Requirements:** CR-STO-100, CR-FND-130, CR-OBS-080

10. **Promote the release.** With gates green, the tag advances to
    production through the ladder.
    - **Expected outcome:** promotion receipts per stage are recorded; the
      release deploys from the immutable tag with atomic rollout and
      automatic rollback on failure.
    - **Requirements:** CR-DPL-070, CR-DPL-100

11. **Record the drill on the calendar.** The executed drill is logged
    against the disaster-recovery drill calendar with its evidence links.
    - **Expected outcome:** the calendar shows the drill cadence per
      service and the next due date; overdue drills downgrade the
      dependent readiness claims automatically.
    - **Requirements:** CR-OPS-180, CR-STO-140

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-STO-060 | Crash-consistent checkpoints and changed-block tracking | 2 |
| CR-STO-070 | Content-addressed snapshot chunk store with integrity verification | 2 |
| CR-STO-080 | Platform backup control plane | 1 |
| CR-STO-100 | Restore drills as promotion blockers | 5, 8, 9 |
| CR-STO-110 | Per-service RPO/RTO objectives declared and enforced | 6 |
| CR-STO-130 | Control-plane state (etcd) snapshot and restore | 4 |
| CR-STO-140 | Durability evidence states and freshness | 1, 5, 8, 11 |
| CR-STO-160 | Offsite immutable backup copies with retention governance | 3 |
| CR-K8S-030 | etcd backup and restore | 4 |
| CR-DAT-080 | Object-storage-backed backup and restore with continuous archiving | 7 |
| CR-DAT-100 | Usage metering and billable backup intervals via durable outbox | 7 |
| CR-OCS-140 | Declared durability surfaces and restore-test objectives | preconditions |
| CR-DPL-070 | Environment ladder and promotion regulation | 8, 10 |
| CR-DPL-100 | Tag-only production releases with atomic rollout | 10 |
| CR-OPS-180 | Disaster-recovery drill calendar | 11 |
| CR-OBS-080 | Alert rules as code per service repository | 9 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 9 |

## Gaps found

None identified. The drill-to-promotion path is fully covered by the
storage domain's gate requirements plus the deployment ladder.

## Evidence required

- Backup completion records with UTC timestamps and evidence states
  (CR-STO-080, CR-STO-140).
- Integrity-audit tool output and corruption-injection test results
  (CR-STO-070).
- Offsite copy and retention evidence (CR-STO-160).
- etcd snapshot job evidence and a restore-to-fresh-cluster drill log
  with post-restore smoke test (CR-K8S-030, CR-STO-130).
- Restore-drill receipts: restored-dataset integrity verification and
  service-health checks in an isolated environment (CR-STO-100).
- Promotion-gate decision records proving rejection of stale, blocked,
  and synthetic evidence with typed reasons (CR-STO-100, CR-STO-140).
- Alert record for the injected drill failure and the persisted `blocked`
  evidence state (CR-STO-100, CR-FND-130).
- Per-stage promotion receipts and the release BOM (CR-DPL-070,
  CR-DPL-100).
- Drill calendar entries linking cadence and evidence per service
  (CR-OPS-180).
