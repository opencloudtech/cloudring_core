# SC-06 — One-server-loss continuity

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the charter's definition of done item 2 continuity clause: the
reference installation loses one server — exercised both as a
control-plane member loss and as a workload-host loss — without
interrupting the API, without tenant-visible loss beyond documented
limits, and with recovery-time measurements recorded as live evidence.
`blocked` drill outcomes are honest terminal states, never converted into
claims.

## Actors

- operator — executes the drill under change control
- provider — owns the installation
- tenant — workloads run during the drill
- auditor — verifies drill evidence and recovery-time measurements

## Preconditions

- The reference installation runs a production-class topology: at least
  three control-plane members across failure domains behind a
  load-balanced, health-checked API endpoint (CR-K8S-020).
- Continuous synthetic monitoring and active dataplane probing are in
  place (CR-OBS-130, CR-NET-210).
- The drill is on the disaster-recovery drill calendar and runs under a
  change record with a rollback plan (CR-OPS-180, CR-OPS-090).
- Tenant workloads include migratable and declared non-migratable
  instances to exercise policy honesty (CR-CMP-120).

## Steps

1. **Baseline the stand.** Record API availability probes, synthetic
   journey results, connectivity matrix state, and workload inventory.
   - **Expected outcome:** a timestamped pre-drill baseline exists;
     golden signals are green; the connectivity matrix is fresh and
     verified.
   - **Requirements:** CR-OBS-130, CR-OBS-030, CR-NET-190

2. **Dry-run the dependency impact.** For the target server, compute
   dependent instances and control-plane impact before any action.
   - **Expected outcome:** the dry-run dependency report matches observed
     state; evacuation feasibility (live-migrate versus governed
     stop-and-restart) is decided per instance policy.
   - **Requirements:** CR-CMP-110

3. **Fail one control-plane member.** Remove or isolate one
   control-plane server while API availability is continuously probed.
   - **Expected outcome:** API service is not interrupted; etcd keeps
     quorum; no admission or scheduling outage is observed; alerts fire
     and route.
   - **Requirements:** CR-K8S-020, CR-OBS-080, CR-OBS-100

4. **Fail one workload host.** Cordon and evacuate a loaded hypervisor
   host, then remove it.
   - **Expected outcome:** instances evacuate per policy — live-migrated
     where supported, governed stop-and-restart with recorded approval
     otherwise; forced power-off is an explicit, audited decision, never
     a default; progress is reported per instance.
   - **Requirements:** CR-CMP-110, CR-CMP-120

5. **Hold storage safety lines.** During host loss, volumes re-attach on
   surviving hosts under fencing.
   - **Expected outcome:** no double-attach occurs; ambiguous ownership
     halts IO rather than risking corruption; storage-dependent drain
     steps honor the maintenance interlock.
   - **Requirements:** CR-STO-040, CR-STO-150

6. **Hold instance-state honesty.** Throughout the event, instance
   states reflect driver-observed reality.
   - **Expected outcome:** instances whose state cannot be determined
     surface as ERROR/UNKNOWN, never as optimistic RUNNING; destructive
     actions on contradictory-state instances are blocked.
   - **Requirements:** CR-CMP-010, CR-CMP-040

7. **Measure tenant-visible impact.** Synthetic journeys and the
   connectivity matrix quantify tenant-visible behavior during the event.
   - **Expected outcome:** any tenant-visible degradation is within the
     documented limits for the topology; a failing matrix cell halts any
     readiness claim refresh and triggers review.
   - **Requirements:** CR-OBS-130, CR-NET-190

8. **Classify and respond as an incident.** The drill's alerting flows
   through severity classification and on-call response as if real.
   - **Expected outcome:** severity is assigned per the classification
     scheme; on-call acknowledges within the response SLO; notifications
     are delivered through the declared channels.
   - **Requirements:** CR-OPS-050, CR-OPS-060, CR-OBS-100

9. **Recover the server.** Replace or rejoin the server through the same
   IaC and bootstrap paths used to build it.
   - **Expected outcome:** re-application of already-satisfied bootstrap
     nodes is a no-op; the node rejoins through the declarative node
     lifecycle with workload-continuity evidence; capacity is restored
     without manual snowflaking.
   - **Requirements:** CR-DPL-030, CR-K8S-130

10. **Reconcile leftovers.** The garbage-collection loop scans for
    half-created resources from any interrupted operations during the
    event.
    - **Expected outcome:** only resources with positive journal proof of
      failed creation are quarantined and cleaned; ambiguous cases go to
      the operator queue, never to the deleter.
    - **Requirements:** CR-CMP-170

11. **Record drill evidence.** Recovery-time measurements, impact
    measurements, and timeline are recorded as live evidence.
    - **Expected outcome:** the drill record carries UTC timestamps,
      measured recovery times, tenant-impact measurements, and the final
      state; a `blocked` drill outcome, if any, stays blocked and visible.
    - **Requirements:** CR-CMP-110, CR-OPS-180, CR-FND-130

12. **Feed readiness.** The drill evidence updates the installation
    readiness report.
    - **Expected outcome:** the one-server-loss obligation of the
      charter's definition of done item 2 is marked satisfied only with
      this fresh, verified, non-synthetic evidence linked.
    - **Requirements:** CR-FND-130, CR-OBS-210

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-K8S-020 | Highly available control plane | 3 |
| CR-K8S-130 | Node lifecycle operations with workload continuity | 9 |
| CR-CMP-010 | Instance lifecycle API with explicit state machine | 6 |
| CR-CMP-040 | Virtualization backend abstraction, KubeVirt primary | 6 |
| CR-CMP-110 | Host maintenance and evacuation safety | 2, 4, 11 |
| CR-CMP-120 | Live migration policy and evidence | 4 |
| CR-CMP-170 | Garbage collection of half-created resources | 10 |
| CR-STO-040 | Mount fencing, access modes, and session lifecycle | 5 |
| CR-STO-150 | Maintenance interlock for storage-dependent drains | 5 |
| CR-NET-190 | Network connectivity evidence as readiness gate | 1, 7 |
| CR-NET-210 | Active synthetic dataplane probing | preconditions |
| CR-OBS-030 | Golden signals and outcome-aware metrics | 1 |
| CR-OBS-080 | Alert rules as code per service repository | 3 |
| CR-OBS-100 | Notification gateway and alert ingress | 3, 8 |
| CR-OBS-130 | Synthetic monitoring service | 1, 7 |
| CR-OBS-210 | Observability evidence as a readiness gate | 12 |
| CR-OPS-050 | Incident severity classification | 8 |
| CR-OPS-060 | On-call model with response SLO | 8 |
| CR-OPS-090 | Change management with dated evidence and rollback plans | preconditions |
| CR-OPS-180 | Disaster-recovery drill calendar | 11 |
| CR-DPL-030 | Bootstrap DAG as a versioned, verified dependency graph | 9 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 11, 12 |

## Gaps found

None identified. Control-plane loss, host loss, storage fencing, network
verification, incident response, and recovery each map to existing
requirements.

## Evidence required

- Pre-drill baseline: availability probes, synthetic journeys, fresh
  connectivity matrix (CR-OBS-130, CR-NET-190).
- Dry-run dependency report matching observed state (CR-CMP-110).
- One-server-loss drill evidence for the control-plane member: continuous
  API availability probe log and etcd quorum health (CR-K8S-020).
- Evacuation evidence: per-instance progress, migration evidence records
  (duration, achieved pause, result), and audit of any forced actions
  (CR-CMP-110, CR-CMP-120).
- Fencing proof: no double-attach during the event (CR-STO-040).
- Alert and incident records: severity assignment, on-call acknowledgement
  times (CR-OPS-050, CR-OPS-060).
- Node replacement/rejoin evidence through the declarative node lifecycle
  (CR-K8S-130, CR-DPL-030).
- GC reconciliation report post-event (CR-CMP-170).
- The final drill record: UTC-stamped recovery-time and impact
  measurements, stored append-only, referenced by the readiness report
  (CR-CMP-110, CR-OPS-180, CR-FND-130).
