# SC-01 — Fresh install to first VM (single host)

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` (evidence over claims, fail closed, real open
baseline) and `01-requirement-schema.md`.

## Purpose

Prove the founding "download and run" promise of the real open baseline:
starting from one clean host, a provider installs the platform from the
downloadable OSS artifact on upstream Kubernetes semantics, passes the
evidence gates that may lawfully gate a single-host stand, and a tenant
creates, reaches, and is metered for a first virtual machine. Single-host
availability limits must be stated honestly — no high-availability claim is
made for this topology.

## Actors

- provider / operator — performs the installation
- tenant — consumes the first instance
- agent — may execute operator steps under agent governance (see SC-10)
- auditor — verifies the evidence chain

## Preconditions

- A clean host meeting the documented hardware and OS prerequisites; no
  pre-existing cluster or platform state.
- The downloadable installer artifact of a tagged release is available.
- Installation documentation is published, organized by audience and task,
  and validated against the real command surface (CR-FND-080).
- No high-availability expectation: single-host installations carry the
  documented no-HA non-claim of CR-DPL-010.

## Steps

1. **Obtain installer and documentation.** The operator downloads the
   tagged installer and follows the published installation guide.
   - **Expected outcome:** the guide matches the real CLI surface; the
     installer belongs to the same artifact family used by CI image/profile
     builds, not a hand-maintained side path.
   - **Requirements:** CR-DPL-010, CR-FND-080

2. **Run fail-closed preflight.** The installer validates resources,
   substrate, and identity/secret inputs before any mutation.
   - **Expected outcome:** a preflight report is produced; on missing
     resources, an unsupported substrate, or unresolved identity/secret
     inputs the installer halts and never proceeds to mutation on a partial
     preflight.
   - **Requirements:** CR-DPL-010, CR-FND-030

3. **Execute the single-host install.** The installer brings up the
   substrate on upstream Kubernetes semantics only.
   - **Expected outcome:** a verifiable install report; the upstream target
     is named in the report; no legacy-substrate evidence is accepted toward
     readiness.
   - **Requirements:** CR-DPL-010, CR-K8S-010, CR-FND-030

4. **Bootstrap DAG executes.** Seed, infrastructure, and platform waves run
   topologically with cycle and missing-dependency policies set to block.
   - **Expected outcome:** per-node receipts are emitted; re-application of
     an already-satisfied node is a no-op; the structural verifier passes
     with no drift between graph, profile, and deployed reality.
   - **Requirements:** CR-DPL-030

5. **Platform state reconciles as code.** The GitOps reconciler is the
   apply path for cluster state, bootstrapping from the environment
   repository itself.
   - **Expected outcome:** every apply is dry-run-first with a recorded
     dry-run receipt; the environment repository fully describes the stand
     and carries no secret material.
   - **Requirements:** CR-DPL-040, CR-DPL-050

6. **Secrets bootstrap by reference only.** The secrets manager comes up
   honest about its sealed state; all secret inputs are brokered references.
   - **Expected outcome:** no secret value appears in configuration, the
     installer report, or any repository; sealed/unsealed state is visible
     and fail-closed.
   - **Requirements:** CR-IAM-140, CR-DPL-110, CR-FND-160

7. **Verify installation readiness evidence.** The network connectivity
   matrix is generated and the observability evidence classes are linked
   for every shipped platform service.
   - **Expected outcome:** a versioned connectivity matrix exists and is
     green; golden-signal dashboards, alert rules, tracing, and log
     pipelines are evidenced; blocked or stale evidence stays visible and is
     never converted into a readiness claim.
   - **Requirements:** CR-NET-190, CR-OBS-210, CR-FND-130

8. **Publish a golden image.** A declarative, version-controlled pipeline
   builds a base image; the artifact is immutable, checksummed, and signed,
   and passes vulnerability and source-safety scans.
   - **Expected outcome:** the image is listed with provenance; an unsigned
     or tampered image is refused at instance-create time.
   - **Requirements:** CR-CMP-070

9. **Create tenant project and network.** A tenant project is created in
   the hierarchical resource model, and a tenant-owned virtual network is
   created as a first-class resource.
   - **Expected outcome:** access bindings inherit correctly down the
     resource tree; the network is isolated by default with default-deny
     policy until the tenant explicitly allows traffic.
   - **Requirements:** CR-IAM-010, CR-NET-010, CR-NET-050

10. **Tenant orders the first instance.** The console/API shows a
    pre-commit review of cost, defaults, and exposure; the create request
    carries a client-supplied idempotency key.
    - **Expected outcome:** no hidden cost, default, or public exposure is
      introduced before explicit commit; the API returns an Operation
      object immediately with a stable operation ID and typed metadata.
    - **Requirements:** CR-CUX-040, CR-FND-120, CR-CMP-010, CR-CMP-020

11. **Quota is reserved two-phase.** Per-project quota is reserved before
    provisioning begins.
    - **Expected outcome:** commit on success, release on failure or
      cancellation; when the quota service is unavailable the creation is
      denied, never admitted unreserved; every transition is auditable.
    - **Requirements:** CR-CMP-160

12. **Instance provisions and boots.** Provisioning executes as a durable,
    idempotent task; the metadata service delivers instance identity,
    user-data, and SSH keys.
    - **Expected outcome:** task state survives a control-plane restart
      with resume-or-roll-forward and zero duplicate side effects; metadata
      responses are source-bound to the requesting instance; no secret is
      placed in general metadata fields.
    - **Requirements:** CR-CMP-030, CR-CMP-140

13. **Tenant reaches the instance.** The tenant connects over the network
    path they opened, or through the proxied console/serial path.
    - **Expected outcome:** console access is IAM-authorized per instance,
      issued as a short-lived single-purpose token, rate-limited, and fully
      audited; no console endpoint is exposed directly on hypervisor
      networks.
    - **Requirements:** CR-CMP-130, CR-NET-040

14. **Metering begins honestly.** Running-VM usage is metered at
    per-second granularity from the heartbeat evidence stream.
    - **Expected outcome:** missing-heartbeat intervals are resolved by the
      documented conservative policy — unproven intervals are not billed;
      heartbeat usage reconciles against compute power-state events under
      UTC discipline.
    - **Requirements:** CR-BIL-110, CR-OBS-150

15. **Publish the honest readiness statement.** The primitive coverage map
    and readiness report are updated for the stand.
    - **Expected outcome:** missing or immature primitives are visible as
      such; single-host availability limits are declared; nothing
      fixture-backed or synthetic is presented as production state.
    - **Requirements:** CR-FND-060, CR-FND-120, CR-DPL-010

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-DPL-010 | Downloadable installer from single host to multi-node | 1, 2, 3, 15 |
| CR-DPL-030 | Bootstrap DAG as a versioned, verified dependency graph | 4 |
| CR-DPL-040 | Everything-as-code with dry-run-first GitOps reconciliation | 5 |
| CR-DPL-050 | GitOps-per-environment repository with bootstrap self-reference | 5 |
| CR-DPL-110 | Secrets in CI and IaC: encrypted at rest, never plaintext | 6 |
| CR-FND-030 | Upstream-Kubernetes-only substrate target | 2, 3 |
| CR-FND-060 | Declared product-primitive minimum | 15 |
| CR-FND-080 | Documentation as operating contract | 1 |
| CR-FND-120 | Production-honesty bans | 10, 15 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 7 |
| CR-FND-160 | Source-safety boundary for all artifacts | 6 |
| CR-K8S-010 | Upstream-only Kubernetes substrate | 3 |
| CR-IAM-010 | Hierarchical resource model with inherited bindings | 9 |
| CR-IAM-140 | Secrets manager with auto-unseal and sealed-state honesty | 6 |
| CR-NET-010 | Virtual networks as first-class tenant resources | 9 |
| CR-NET-040 | Tenant-controlled routing | 13 |
| CR-NET-050 | Security groups and default-deny network policy | 9 |
| CR-NET-190 | Network connectivity evidence as readiness gate | 7 |
| CR-OBS-150 | UTC-canonical observability records | 14 |
| CR-OBS-210 | Observability evidence as a readiness gate | 7 |
| CR-CMP-010 | Instance lifecycle API with explicit state machine | 10 |
| CR-CMP-020 | Long-running operation model | 10 |
| CR-CMP-030 | Durable, idempotent task execution framework | 12 |
| CR-CMP-070 | Golden-image pipeline | 8 |
| CR-CMP-130 | Console and serial access | 13 |
| CR-CMP-140 | Instance metadata service | 12 |
| CR-CMP-160 | Compute quotas with two-phase reservation | 11 |
| CR-CUX-040 | Pre-commit review of cost, defaults, and exposure | 10 |
| CR-BIL-110 | Per-second VM usage heartbeat as metering evidence | 14 |

## Gaps found

None identified. Every step of this journey maps to at least one existing
requirement. Note that CR-DPL-010 itself records the honesty boundary for
this scenario (no HA claim for single-host installs); the scenario enforces
that boundary rather than working around it.

## Evidence required

- Preflight report and verifiable install report from the clean-machine
  install test (CR-DPL-010 acceptance evidence class).
- Bootstrap DAG execution log with per-node receipts and a passing
  structural-verifier run (CR-DPL-030).
- Dry-run/apply receipts from the GitOps reconciliation path (CR-DPL-040).
- Versioned network connectivity matrix artifact, green and append-only
  stored (CR-NET-190).
- Observability readiness-gate report listing the five evidence classes
  with per-class states (CR-OBS-210).
- Golden-image pipeline runs, provenance records, and blocked-boot tests
  for unsigned images (CR-CMP-070).
- End-to-end instance lifecycle run on the stand, including
  idempotency-replay proof (CR-CMP-010, CR-CMP-020).
- Quota reservation audit records and fail-closed test output (CR-CMP-160).
- Metadata isolation test results showing zero cross-instance leakage
  (CR-CMP-140).
- Console-session audit records (CR-CMP-130).
- Heartbeat-to-usage conversion and reconciliation evidence (CR-BIL-110).
- The readiness report citing the primitive coverage map with honest
  maturity states (CR-FND-060, CR-FND-130).
