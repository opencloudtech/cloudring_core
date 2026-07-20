# SC-02 — Reference-provider production install (IaC bootstrap DAG)

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the deployment half of the charter's definition of done item 2: a
production-class reference installation on at least one real provider,
deployed **entirely from IaC** through the versioned bootstrap dependency
graph, with no hand-applied state, no committed secrets, and readiness
gated on connectivity, observability, and durability evidence. The
companion proof obligations of definition of done item 2 — upgrade,
backup/restore, and one-server-loss — are closed by SC-05, SC-06, and
SC-07.

## Actors

- provider — owns the reference installation
- operator — executes deployment waves through the IaC toolchain
- agent — may run wave steps under governance (see SC-10)
- auditor — verifies receipts, verifier output, and evidence states

## Preconditions

- The abstract deployment profile schema is published in the OSS
  repository with a machine validator (CR-DPL-020).
- A concrete provider profile exists downstream, outside the public core,
  implementing the abstract schema through declared provider-adapter
  capability classes (CR-DPL-020).
- The environment repository exists and contains only deltas over
  versioned base templates (CR-DPL-050, CR-DPL-190).
- CI gates on the public repository are green (CR-DPL-090); the release is
  an immutable, tagged, signed ref with a bill of materials (CR-DPL-100).

## Steps

1. **Validate the concrete provider profile.** The downstream profile is
   validated against the abstract OSS schema.
   - **Expected outcome:** the validator fails closed on invalid fixtures;
     no provider endpoint, credential, live inventory, or provider-specific
     apply logic exists in the public profile layer.
   - **Requirements:** CR-DPL-020, CR-FND-160

2. **Review the bootstrap DAG.** The versioned dependency graph (seed →
   infrastructure → platform waves) is reviewed as code.
   - **Expected outcome:** the graph is under version control with
     reviewable diffs; cycle and missing-dependency policies are set to
     block.
   - **Requirements:** CR-DPL-030

3. **Bootstrap the environment repository self-reference.** The GitOps
   agent bootstraps from the same repository that describes the
   environment.
   - **Expected outcome:** self-reference is verified in CI; the bootstrap
     reference points at a pinned, immutable ref; no committed secret
     material exists in the repository.
   - **Requirements:** CR-DPL-050, CR-DPL-110

4. **Provision infrastructure through IaC.** Provisioning tooling creates
   the provider-side resources; IaC state is remote, locked, and versioned.
   - **Expected outcome:** state is never committed to any repository;
     every apply is dry-run-first with a recorded dry-run receipt;
     unexpected destroy/replace of stateful resources halts the apply.
   - **Requirements:** CR-DPL-040, CR-DPL-120

5. **Build golden images per datacenter.** Image pipelines produce signed,
   checksummed artifacts with provenance.
   - **Expected outcome:** pipeline runs are recorded as CI evidence;
     images pass vulnerability and source-safety scans before publication.
   - **Requirements:** CR-DPL-130, CR-CMP-070

6. **Execute the seed wave.** Identity roots, coordination, key
   management, and storage foundations deploy in topological order.
   - **Expected outcome:** per-node receipts are emitted; a node whose
     dependency receipt is absent, stale, or failed halts the wave — never
     force-promoted over blocked dependencies.
   - **Requirements:** CR-DPL-030, CR-IAM-130, CR-IAM-140

7. **Execute the infrastructure wave.** Compute, network, and storage
   services deploy behind their capability profiles.
   - **Expected outcome:** each layer's profile declares capabilities,
     limits, portability boundary, exit path, upgrade boundary, and
     explicitly unsupported states; a layer without its profile contract is
     not accepted.
   - **Requirements:** CR-DPL-030, CR-FND-040

8. **Stand up the highly available control plane.** At least three
   control-plane members spread across failure domains behind a
   load-balanced, health-checked API endpoint; etcd as a quorum of at
   least three.
   - **Expected outcome:** the structural verifier confirms topology;
     certificate rotation is automated without API downtime.
   - **Requirements:** CR-K8S-020

9. **Enable scheduled etcd snapshots.** Snapshots (default cadence at
   least every 6 hours, retention at least 60 unless overridden) are
   written to off-cluster object storage with integrity metadata.
   - **Expected outcome:** snapshot jobs emit success/failure evidence;
     snapshot credentials are referenced from the approved secrets workflow
     and exist nowhere in git.
   - **Requirements:** CR-K8S-030, CR-STO-130

10. **Execute the platform wave.** Runtime, observability, and the service
    plane deploy via GitOps reconciliation.
    - **Expected outcome:** all platform services are measurable,
      traceable, loggable, and alertable through declared, Git-managed
      configuration before any readiness claim.
    - **Requirements:** CR-DPL-040, CR-OBS-210

11. **Enforce supply-chain integrity.** All deployed artifacts resolve to
    pinned digests with SBOM, provenance, and vulnerability-exploitability
    metadata.
    - **Expected outcome:** unpinned or unverifiable components fail the
      pipeline closed; the release BOM pins every component version.
    - **Requirements:** CR-DPL-160, CR-K8S-080, CR-DPL-100

12. **Bring identity and security to production posture.** Authorization
    fails closed everywhere; management surfaces stay hidden until IAM
    allows them.
    - **Expected outcome:** on any token or security error the management
      surfaces deny; tenant-isolation tests run as a standing gate.
    - **Requirements:** CR-IAM-160, CR-IAM-170

13. **Configure offsite immutable backup copies.** Backup copies with
    retention governance are in place for platform and tenant data.
    - **Expected outcome:** copy jobs and retention rules are declared as
      code; copy integrity is verified and evidenced.
    - **Requirements:** CR-STO-160

14. **Regenerate and verify the network connectivity matrix.** Zone-to-
    zone, tenant-to-edge, and tenant-to-platform-services flows on both
    address families are probed with expected MTU, latency, and policy
    outcomes.
    - **Expected outcome:** a fresh verified matrix is stored; any
      unexpected reachability cell halts promotion and triggers security
      review.
    - **Requirements:** CR-NET-190, CR-NET-120

15. **Run readiness gates and publish the evidence report.** Durability
    evidence states, observability evidence, connectivity evidence, and
    deployment receipts are aggregated into the installation readiness
    report.
    - **Expected outcome:** only verified, non-synthetic, fresh evidence
      promotes; blocked states remain visible; the report distinguishes the
      upstream target architecture from any fallback planning.
    - **Requirements:** CR-STO-140, CR-OBS-210, CR-FND-130, CR-FND-030

16. **Record the honest non-claims.** The readiness report records that
    only one concrete provider profile lineage exists and that portability
    of the abstract schema across a second provider is unproven.
    - **Expected outcome:** the non-claim is carried verbatim in the
      report; no "multi-provider proven" language appears anywhere.
    - **Requirements:** CR-DPL-020, CR-FND-130

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-DPL-020 | Abstract deployment profile in OSS, concrete provider profiles downstream | 1, 16 |
| CR-DPL-030 | Bootstrap DAG as a versioned, verified dependency graph | 2, 6, 7 |
| CR-DPL-040 | Everything-as-code with dry-run-first GitOps reconciliation | 4, 10 |
| CR-DPL-050 | GitOps-per-environment repository with bootstrap self-reference | 3 |
| CR-DPL-100 | Tag-only production releases with atomic rollout | 11 |
| CR-DPL-110 | Secrets in CI and IaC: encrypted at rest, never plaintext | 3 |
| CR-DPL-120 | Remote, locked, versioned IaC state — never committed | 4 |
| CR-DPL-130 | Golden image pipeline per datacenter | 5 |
| CR-DPL-160 | Supply-chain integrity: pinned digests, SBOM, provenance, VEX | 11 |
| CR-DPL-190 | Base cluster templates with deep-merge environment customization | preconditions |
| CR-DPL-090 | CI gates and merge gating on the public repository | preconditions |
| CR-FND-030 | Upstream-Kubernetes-only substrate target | 15 |
| CR-FND-040 | Layered replaceable capability profiles | 7 |
| CR-FND-130 | Evidence before readiness; blocked stays blocked | 15, 16 |
| CR-FND-160 | Source-safety boundary for all artifacts | 1 |
| CR-K8S-020 | Highly available control plane | 8 |
| CR-K8S-030 | etcd backup and restore | 9 |
| CR-K8S-080 | Artifact signing, SBOM, and provenance | 11 |
| CR-IAM-130 | Key management with envelope encryption and hardware root of trust | 6 |
| CR-IAM-140 | Secrets manager with auto-unseal and sealed-state honesty | 6 |
| CR-IAM-160 | Fail-closed authorization and hidden management surfaces | 12 |
| CR-IAM-170 | Tenant isolation test gate | 12 |
| CR-NET-120 | Dual-stack IPv4/IPv6 as a gate | 14 |
| CR-NET-190 | Network connectivity evidence as readiness gate | 14 |
| CR-OBS-210 | Observability evidence as a readiness gate | 10, 15 |
| CR-STO-130 | Control-plane state (etcd) snapshot and restore | 9 |
| CR-STO-140 | Durability evidence states and freshness | 15 |
| CR-STO-160 | Offsite immutable backup copies with retention governance | 13 |
| CR-CMP-070 | Golden-image pipeline | 5 |

## Gaps found

None identified. The single-provider-lineage limitation is an existing
recorded non-claim of CR-DPL-020, and this scenario carries it as evidence
honesty (step 16) rather than treating it as missing coverage.

## Evidence required

- Profile validation output including invalid-fixture rejection
  (CR-DPL-020).
- Bootstrap DAG execution log with per-node receipts and
  structural-verifier runs proving cycle/missing-dependency rejection and
  idempotent re-application (CR-DPL-030).
- Dry-run/apply/drift receipt chain for every wave (CR-DPL-040).
- IaC state backend configuration proving remote, locked, versioned state
  (CR-DPL-120).
- Golden-image pipeline evidence per datacenter (CR-DPL-130, CR-CMP-070).
- Control-plane topology verifier output and certificate-rotation test
  (CR-K8S-020).
- etcd snapshot job evidence and source-safety proof that no snapshot
  credentials exist in git (CR-K8S-030).
- Supply-chain verification records: pinned digests, SBOM, provenance
  (CR-DPL-160, CR-K8S-080).
- Tenant-isolation gate results and fail-closed management-surface tests
  (CR-IAM-160, CR-IAM-170).
- Offsite immutable copy configuration and integrity evidence (CR-STO-160).
- Fresh verified network connectivity matrix (CR-NET-190).
- Observability readiness-gate report (CR-OBS-210).
- The aggregated installation readiness report with explicit evidence
  states and the recorded single-lineage non-claim (CR-FND-130,
  CR-STO-140).
