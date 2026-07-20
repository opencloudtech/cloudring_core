# SC-03 — Service team onboards a new OCS service

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the charter's definition of done item 3 and the Open Cloud Standard
principle: a third-party service team — with no platform-team involvement —
registers a new service, implements the mandatory lifecycle APIs, passes
conformance, and reaches the catalog using only public documentation, the
SDK, and the shared CI contract. The platform core must never carry
service-specific wiring: the connector package is the sole integration
unit.

## Actors

- service-team / vendor — builds and submits the service
- provider — operates the installation whose catalog receives the service
- auditor — verifies conformance, evidence, and the publication-gate record

## Preconditions

- Public documentation, the SDK, and the provider-neutral reference
  implementation are published (CR-OCS-100, CR-OCS-110).
- The conformance suite and validator are runnable locally and in CI
  (CR-OCS-090).
- The target installation runs a compliant module registry and catalog
  (CR-OCS-080, CR-MKT-080).
- The service team holds no privileged platform access — only a publisher
  identity and public materials.

## Steps

1. **Read the onboarding journey.** The team follows the published
   journey: documentation, skeleton generation, local validation,
   conformance, catalog submission.
   - **Expected outcome:** the journey is achievable without platform-team
     involvement; documentation states intent before mechanism and matches
     the real tool surface.
   - **Requirements:** CR-OCS-110, CR-FND-080

2. **Generate the connector package skeleton.** Tooling emits a package
   that passes structural validation out of the box.
   - **Expected outcome:** the connector package is the only integration
     artifact; nothing service-specific is added to platform internals.
   - **Requirements:** CR-OCS-010, CR-OCS-110

3. **Register the service and announce capabilities.** The package
   declares one service identity, its capabilities, and its optional APIs.
   - **Expected outcome:** registration resolves to a single service
     identity; capability announcements are machine-readable; optional
     APIs and inter-service dependencies are declared through the portable
     dependency contract.
   - **Requirements:** CR-OCS-020, CR-OCS-050

4. **Implement the mandatory lifecycle APIs.** Provision, deprovision,
   and the remaining mandatory verbs are implemented idempotently.
   - **Expected outcome:** every lifecycle action is idempotent; mutating
     delete/retry/repair actions carry rollback references; repeated calls
     produce no duplicate side effects.
   - **Requirements:** CR-OCS-030

5. **Implement typed user-visible states.** The service reports the shared
   state vocabulary with remediation hints.
   - **Expected outcome:** states are typed, user-visible, and carry
     remediation; the same object reports the same state on every surface;
     denials name the missing precondition.
   - **Requirements:** CR-OCS-040, CR-FND-140

6. **Implement the billing connector.** The service declares its usage
   meters and transmits usage through the billing connector surface.
   - **Expected outcome:** meters are declared in the package; usage
     transmission is idempotent with replay dedup; the service cannot
     reach the catalog unmetered — metering integration is a launch gate.
   - **Requirements:** CR-OCS-060, CR-BIL-130, CR-BIL-180

7. **Reference secrets only through workload identity.** The package
   declares secret needs as workload-identity references.
   - **Expected outcome:** raw secret material anywhere in the package is
     a hard blocker at validation; secrets flow only through the brokered
     path at runtime.
   - **Requirements:** CR-OCS-130, CR-IAM-090

8. **Declare durability surfaces and objectives.** The service declares
   what state it holds, its backup/restore behavior, and RPO/RTO
   objectives.
   - **Expected outcome:** stateful services publish machine-readable
     RPO/RTO and restore-test objectives; the catalog can surface these to
     tenants.
   - **Requirements:** CR-OCS-140, CR-STO-110

9. **Package the portal extension, if any.** The microfrontend declares
   integrity metadata and runs inside the sandboxed host.
   - **Expected outcome:** the microfrontend contract passes integrity and
     sandbox checks; the runtime registry renders the service's pages
     without privileged host access.
   - **Requirements:** CR-OCS-070, CR-CUX-130

10. **Adopt the shared CI contract.** The service repository includes the
    centrally maintained pipeline by reference.
    - **Expected outcome:** ordered stages (checks → build →
      database-migrate → deploy → regression) run with no pipeline fork;
      a third-party team can adopt it from public documentation alone.
    - **Requirements:** CR-DPL-080

11. **Run local validation and the conformance suite.** The team iterates
    against machine-readable problem reports.
    - **Expected outcome:** conformance output lists structured problems
      (surface, field, message, remediation); a clean run produces a
      passing conformance report.
    - **Requirements:** CR-OCS-090, CR-FND-090

12. **Exercise the minimum-useful profile path.** A small team ships first
    against the named minimum-useful subset, then upgrades to the full
    profile.
    - **Expected outcome:** the minimum-useful profile grants limited
      catalog presence only; any surface touching money, deletion, or
      secrets without its full mandatory checks halts submission.
    - **Requirements:** CR-OCS-110

13. **Produce evidence bundles and receipts.** The service emits evidence
    in the machine receipt format for lifecycle and durability claims.
    - **Expected outcome:** evidence bundles validate against the receipt
      schema; blocked, stale, or synthetic-only evidence never promotes.
    - **Requirements:** CR-OCS-150, CR-FND-130

14. **Pass the publication gate.** Catalog, tenant-access, support,
    readiness, and durability surfaces are complete; observability
    evidence is linked.
    - **Expected outcome:** the gate requires all surfaces plus a passing
      conformance report and a recorded owner review; golden-signal
      dashboards, alert rules, tracing, and log pipeline evidence are
      attached.
    - **Requirements:** CR-OCS-160, CR-OBS-210

15. **Appear in the service store.** The service lists in the one catalog
    through the one order path, on equal terms with first-party services.
    - **Expected outcome:** first-party and vendor products share one
      catalog, one order path, and one metering contract; distribution
      profiles and optional modules are declared honestly.
    - **Requirements:** CR-MKT-080, CR-OCS-170, CR-FND-010

16. **Record versioning posture.** The package pins its contract schema
    versions and compatibility windows.
    - **Expected outcome:** an unsupported version combination resolves to
      a blocked state, never a best-effort run.
    - **Requirements:** CR-OCS-120

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-OCS-010 | Connector package as the sole service-integration unit | 2 |
| CR-OCS-020 | Registration, capability announcement, and one service identity | 3 |
| CR-OCS-030 | Mandatory lifecycle APIs: idempotent, rollback-referenced | 4 |
| CR-OCS-040 | Typed user-visible states with remediation | 5 |
| CR-OCS-050 | Optional APIs and portable inter-service dependencies | 3 |
| CR-OCS-060 | Billing connector and usage-metrics transmission | 6 |
| CR-OCS-070 | Microfrontend contract: integrity, sandbox, guidelines | 9 |
| CR-OCS-080 | Module registry with versioning and policy-aware lifecycle | preconditions |
| CR-OCS-090 | Conformance suite with machine-readable problems | 11 |
| CR-OCS-100 | SDK and provider-neutral reference implementation | preconditions |
| CR-OCS-110 | Service-team onboarding journey and minimum-useful profile | 1, 2, 12 |
| CR-OCS-120 | Versioning, deprecation, and compatibility gateways | 16 |
| CR-OCS-130 | Workload-identity-only secret references | 7 |
| CR-OCS-140 | Declared durability surfaces and restore-test objectives | 8 |
| CR-OCS-150 | Evidence bundles and machine receipts format | 13 |
| CR-OCS-160 | Publication gate: catalog, tenant-access, support, and readiness surfaces | 14 |
| CR-OCS-170 | Distribution profiles and optional-module honesty | 15 |
| CR-BIL-130 | OCS billing connector surface: meters, rate-card evidence, replay dedup | 6 |
| CR-BIL-180 | Metering integration as a launch gate for usage-bearing services | 6 |
| CR-MKT-080 | Service-store model: first-class third-party services | 15 |
| CR-CUX-130 | Micro-frontend host with runtime registry | 9 |
| CR-DPL-080 | Shared layered CI pipeline contract for services | 10 |
| CR-FND-010 | Contract before technology | 15 |
| CR-FND-080 | Documentation as operating contract | 1 |
| CR-FND-090 | Machine-checkable documentation and corpus | 11 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 13 |
| CR-FND-140 | One product truth across surfaces | 5 |
| CR-IAM-090 | Workload identity by default; static credentials exception-only | 7 |
| CR-OBS-210 | Observability evidence as a readiness gate | 14 |
| CR-STO-110 | Per-service RPO/RTO objectives declared and enforced | 8 |

## Gaps found

None identified. The journey maps fully onto the OCS domain and its
sibling gates. (CR-OCS-110 records that end-to-end validation by a fully
external team is still pending — closing this scenario with a genuinely
external team is exactly the evidence that retires that non-claim.)

## Evidence required

- Skeleton tooling output passing validation (CR-OCS-110).
- The connector package with registration, capabilities, lifecycle,
  billing, durability, and microfrontend declarations (CR-OCS-010 …
  CR-OCS-070).
- Idempotency and rollback-reference test results for all mandatory
  lifecycle verbs (CR-OCS-030).
- Conformance report with a clean structured problem list (CR-OCS-090).
- Billing connector evidence: declared meters, dry-run transmission, and
  replay-dedup proof (CR-OCS-060, CR-BIL-130, CR-BIL-180).
- Validator proof that raw secret material is a hard blocker
  (CR-OCS-130).
- Durability declaration with RPO/RTO objectives (CR-OCS-140,
  CR-STO-110).
- Shared-CI adoption evidence by inclusion without fork (CR-DPL-080).
- Publication-gate record including owner review and observability
  evidence (CR-OCS-160, CR-OBS-210).
- A recorded onboarding drill: a team external to core development
  completes documentation → catalog submission using only public materials
  (CR-OCS-110; charter definition of done item 3).
