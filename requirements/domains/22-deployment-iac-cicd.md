# 22 — Deployment, IaC and CI/CD

Scope: how the platform is installed, upgraded, and delivered as code. Covers the
downloadable installer (single-host to multi-node), the abstract provider-agnostic
deployment profile in the OSS repository with concrete provider profiles downstream,
the versioned bootstrap dependency graph (seed → infrastructure → platform waves),
everything-as-code discipline (provisioning tooling, Kubernetes manifests, GitOps),
golden image pipelines, environment topology and naming, environment promotion
regulation, CI gates and merge gating on the public repository, release discipline,
progressive delivery, database-migration staging, secrets and state handling, and
supply-chain integrity.

Domain contract: every environment is reproducible from version-controlled
declarations alone; no production change reaches production without passing the
promotion ladder, the required CI gates, and an explicit tag; no secret, credential,
state file, or tenant artifact is ever committed; every mutating deploy is
idempotent, dry-run-first, gated by pre-mutation evidence, and reversible. The
platform MUST NOT be called deployable or operable while any of these are unmet.

---

### CR-DPL-010 — Downloadable installer from single host to multi-node
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** The founding vision promises that a user downloads a distribution
  and deploys an installation on their own server or servers. Without a
  supported installer path the OSS layer is not genuinely sufficient to run a
  provider, and adoption reduces to reading source code.
- **Requirement:** The OSS repository MUST ship a downloadable installer that
  deploys a working installation from a single host upward to multi-node
  topologies on upstream Kubernetes semantics. The installer MUST run
  fail-closed preflight before any mutation, MUST produce a verifiable install
  report, and MUST be the same artifact family used by CI image/profile builds
  rather than a hand-maintained side path. Multi-node expansion MUST be an
  incremental, idempotent re-application of the same profile, not a separate
  installer.
- **Acceptance evidence:** clean-machine install test suite (single host and
  3+ node) executed in CI on every release candidate; recorded install reports
  with preflight results; structural verifier covering installer, profile
  schema, and documentation coherence.
- **Non-goals:** bare-metal OS provisioning for arbitrary hardware fleets;
  automatic capacity planning; commercial support packaging.
- **Non-claims:** unattended bare-metal discovery and zero-touch fleet
  enrollment are not yet designed or proven; single-host installations make no
  high-availability claim.
- **Stop conditions:** halt and escalate if preflight reports missing
  resources, unsupported substrate, or unresolved identity/secret inputs —
  never proceed to mutation on partial preflight (keys / exposure / migration).
- **Traceability:** vision-deck; legacy-platform-a; current-core.

### CR-DPL-020 — Abstract deployment profile in OSS, concrete provider profiles downstream
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** A deployment description hard-wired to one vendor recreates the
  lock-in the platform exists to remove. Provider-specific installation code in
  the public repository also leaks private operational detail and couples the
  OSS release cadence to one provider.
- **Requirement:** The OSS repository MUST define an abstract, provider-agnostic
  installation profile schema (versions, waves, components, capability and
  evidence expectations) with a machine validator. Concrete provider profiles
  (including the reference installation) MUST live downstream, implement the
  abstract profile through declared provider-adapter capability classes, and
  MUST NOT be required for OSS validation. The public core MUST NOT contain
  provider endpoints, credentials, live inventory, or provider-specific apply
  logic.
- **Acceptance evidence:** profile schema plus valid/invalid fixtures with a
  fail-closed validator in CI; a downstream provider profile validated against
  the abstract schema; source-safety checks proving no provider endpoints or
  implementation references in the public profile layer.
- **Non-goals:** a universal lowest-common-denominator API across providers;
  shipping any vendor's proprietary automation in OSS.
- **Non-claims:** only one concrete provider profile lineage exists today;
  portability of the abstract schema across a second provider is unproven.
- **Stop conditions:** halt publication if a provider-specific path, endpoint,
  or credential class is detected in the public profile layer (keys / trust /
  exposure).
- **Traceability:** vision-deck; legacy-platform-a; current-core.

### CR-DPL-030 — Bootstrap DAG as a versioned, verified dependency graph
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Stand bring-up order that lives in operator memory produces
  irreproducible environments and undebuggable partial installs. Deployment
  waves must be designed, versioned, and machine-checkable, not emergent.
- **Requirement:** Platform bootstrap MUST be expressed as a versioned
  dependency graph kept as code: a seed wave (identity roots, coordination,
  key management, storage foundations), an infrastructure wave (compute,
  network, storage services), and a platform wave (runtime, observability,
  service plane). The graph MUST be executed topologically with cycle and
  missing-dependency policies set to block, MUST emit per-node receipts, and
  MUST have a structural verifier that fails closed on drift between graph,
  profile, and deployed reality. Re-application of an already-satisfied node
  MUST be a no-op.
- **Acceptance evidence:** the graph definition under version control with
  reviewable diffs; automated topological execution logs with receipts;
  verifier runs proving cycle/missing-dependency rejection and idempotent
  re-application.
- **Non-goals:** a general-purpose workflow engine; runtime traffic management.
- **Non-claims:** the graph currently covers the foundation waves; full
  platform-wave coverage (all optional modules) is not yet asserted.
- **Stop conditions:** halt the wave when a node's dependency receipt is
  absent, stale, or failed — never force-promote a node over blocked
  dependencies (data / trust / migration).
- **Traceability:** legacy-platform-b; current-core; vision-deck.

### CR-DPL-040 — Everything-as-code with dry-run-first GitOps reconciliation
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, service-team
- **Problem:** Environments mutated by hand diverge silently and cannot be
  audited or rebuilt. All infrastructure and platform state must be
  declarative, reviewable, and reconciled.
- **Requirement:** All platform and environment state MUST be declared as code:
  provisioning tooling (Terraform/Ansible-class) for infrastructure,
  Kubernetes manifests for workloads, and a GitOps reconciler as the apply
  path for cluster state. Every apply path MUST be dry-run-first with a
  recorded dry-run receipt before mutation, MUST be idempotent, and MUST
  classify drift as evidence. Out-of-band manual mutation of managed state is
  a defect to be reconciled or escalated, not an operating mode.
- **Acceptance evidence:** repository layout proving no non-code mutation
  paths; dry-run/apply/rollback/drift receipt classes recorded per change;
  GitOps handoff contract checks (dry-run gate enforced) in CI.
- **Non-goals:** forbidding break-glass procedures (they are governed
  separately with audit); mandating one specific IaC tool for service teams.
- **Non-claims:** drift-detection coverage is partial for non-cluster
  infrastructure; reconciliation of externally mutated IaaS state is not yet
  continuously verified.
- **Stop conditions:** halt apply when dry-run receipt is absent or shows
  unexpected destroy/replace of stateful resources (data / deletion /
  migration).
- **Traceability:** legacy-platform-a; legacy-platform-b; current-core.

### CR-DPL-050 — GitOps-per-environment repository with bootstrap self-reference
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** If an environment cannot be rebuilt from a single repository,
  disaster recovery and environment cloning become manual archaeology. The
  environment definition must be the environment.
- **Requirement:** Each environment MUST be fully described by one
  version-controlled repository containing the platform core declarations
  (storage, mesh, policy, identity) and shared add-ons, and the GitOps agent
  MUST bootstrap from that same repository (self-reference). Environment
  repositories MUST carry only deltas over versioned base templates. Cluster
  rebuild from the repository alone MUST be a rehearsed operation.
- **Acceptance evidence:** rebuild drill evidence (fresh cluster reconciled to
  readiness from the environment repository alone); template-plus-delta
  structure checks; bootstrap self-reference verified in CI.
- **Non-goals:** mono-repository for all environments; storing environment
  secrets in these repositories (references only).
- **Non-claims:** full rebuild drills have been rehearsed for the reference
  environment classes only; cross-provider rebuild portability is unproven.
- **Stop conditions:** halt reconciliation if the environment repository
  contains committed secret material or the bootstrap reference points at an
  unpinned/mutable ref (keys / trust).
- **Traceability:** legacy-platform-a; current-core.

### CR-DPL-060 — Environment topology and naming convention
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Ad-hoc environment names and mixed-purpose clusters blur blast
  radius, compliance zones, and automation targeting. Topology must be a
  contract, not a habit.
- **Requirement:** Environment naming MUST follow a documented, machine-checked
  scheme encoding region/zone, environment class, and cluster role. Per
  environment, the platform SHOULD separate infrastructure/payload,
  external-ingress (DMZ), and monitoring clusters, and MUST at minimum
  document the blast-radius decision when roles are colocated. The scheme MUST
  be validated in CI against the environment registry.
- **Acceptance evidence:** naming regulation document plus a machine-readable
  environment registry validated in CI; topology declarations per environment
  showing role separation or an explicit recorded waiver.
- **Non-goals:** mandating a fixed cluster count for small/edge installations;
  prescribing physical network design.
- **Non-claims:** the DMZ/monitoring split is proven at reference scale;
  minimum-footprint (single-cluster) production hardening is not fully
  evidenced.
- **Stop conditions:** halt deployment of a new environment whose name or role
  set fails registry validation — ambiguous targeting risks cross-environment
  mutation (exposure / deletion / data).
- **Traceability:** legacy-platform-a; legacy-platform-b.

### CR-DPL-070 — Environment ladder and promotion regulation
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, provider
- **Problem:** Direct-to-production changes are the classic root cause of
  avoidable outages. Promotion discipline must be codified and enforced by
  tooling, not by convention documents alone.
- **Requirement:** The platform MUST define an environment ladder
  (development → pre-production → staging → production, with staging running
  production-class configuration without users) and a written, tool-enforced
  promotion regulation: no artifact or configuration reaches production without
  having been promoted through the ladder, committed before deployment, and
  evidenced per stage. Emergency paths MUST be explicit, audited, and
  retroactively reconciled into the ladder.
- **Acceptance evidence:** promotion regulation document plus pipeline
  configuration proving stage ordering is enforced; per-stage promotion
  receipts; audit trail for any emergency bypass.
- **Non-goals:** mandating identical sizing across stages; blocking
  development-environment experimentation.
- **Non-claims:** staging parity with production is asserted by configuration
  class, not yet by continuous conformance checking.
- **Stop conditions:** halt any production-targeted change lacking
  lower-stage receipts or carrying uncommitted state (data / trust /
  migration).
- **Traceability:** legacy-platform-b; legacy-platform-a; current-core.

### CR-DPL-080 — Shared layered CI pipeline contract for services
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator
- **Problem:** Per-service pipeline sprawl produces inconsistent quality gates
  and unmaintainable YAML duplication. A platform of independent service teams
  needs one maintained delivery contract adopted with near-zero boilerplate.
- **Requirement:** The platform MUST provide a shared, centrally maintained CI
  pipeline contract consumed by service repositories via inclusion, with
  ordered stages (checks → build → database-migrate → deploy per environment →
  regression tests) and convention-based job activation (a stage runs when its
  conventional script/interface exists). Per-environment deployment values
  MUST be co-versioned with the service code. The contract MUST be usable by a
  third-party service team from public documentation alone.
- **Acceptance evidence:** the shared pipeline definition under version control
  with semantic versioning; at least one external-style service adopting it by
  inclusion with no pipeline fork; contract tests proving stage ordering and
  convention activation.
- **Non-goals:** forcing one language/toolchain per stage; forbidding
  additional service-specific stages.
- **Non-claims:** adoption by a genuinely external (non-owner) service team is
  designed but not yet demonstrated.
- **Stop conditions:** halt the pipeline on any failed stage; a failed
  db-migrate stage MUST block deploy stages automatically (data / migration).
- **Traceability:** legacy-platform-a; vision-deck.

### CR-DPL-090 — CI gates and merge gating on the public repository
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, operator, auditor
- **Problem:** An open platform whose public repository can be merged without
  tests, security scans, and source-safety checks will accumulate secrets,
  vulnerabilities, and private-boundary leaks in public history.
- **Requirement:** The public repository MUST enforce required checks before
  merge: lint/test/build, security scanning (static analysis, dependency
  vulnerability scanning, secret scanning), source-safety scanning (no
  credentials, tenant data, private endpoints, host paths, copied private
  source), and contract conformance checks. Branch protection MUST require
  pull requests with green required checks; direct pushes to the default
  branch MUST be disabled. Publication from any private workspace to the
  public repository MUST go through the gated path (manifest export → verify →
  candidate branch → required checks → merge).
- **Acceptance evidence:** branch-protection configuration export; CI workflow
  definitions for all gate classes with merge-queue triggers; safe-push tool
  run records showing fail-closed behavior on missing or failed checks.
- **Non-goals:** gating on style preferences beyond automated lint; running
  provider-live tests in public CI.
- **Non-claims:** gate coverage for supply-chain provenance verification is
  defined separately (CR-DPL-160) and not yet fully wired.
- **Stop conditions:** halt and block merge on any source-safety or secret-scan
  finding — never override with a merge-time waiver (keys / trust / exposure).
- **Traceability:** current-core; legacy-platform-a.

### CR-DPL-100 — Tag-only production releases with atomic rollout
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team
- **Problem:** Production deployed from floating branches makes the running
  state unanswerable to "what exactly is deployed?" and breaks rollback
  discipline.
- **Requirement:** Production deployments MUST originate only from immutable,
  tagged, signed release refs; pipelines targeting production MUST fail closed
  on any non-tag ref. Rollouts MUST be atomic (all-or-nothing with bounded
  history and automatic rollback on failure), and every release MUST map to a
  bill of materials pinning component versions. Release tags MUST be linked to
  recorded promotion evidence from lower environments.
- **Acceptance evidence:** pipeline logic rejecting non-tag refs for production
  targets with test coverage; rollout configurations showing atomic flags,
  bounded history, and automatic rollback; release BOM records per tag.
- **Non-goals:** forbidding hotfix tags (they follow the same tag discipline);
  mandating a fixed release cadence.
- **Non-claims:** signed-tag verification is required by policy; enforcement
  of signature verification at deploy time is not yet uniformly proven.
- **Stop conditions:** halt a production deploy when the tag's lower-stage
  promotion receipts are missing or the BOM references unpinned components
  (trust / data / migration).
- **Traceability:** legacy-platform-a; current-core.

### CR-DPL-110 — Secrets in CI and IaC: encrypted at rest, never plaintext
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, auditor
- **Problem:** Committed credentials and plaintext pipeline secrets were a
  repeatedly observed, real-world failure class. Secret handling must be
  structural: encrypted at rest, injected at runtime, masked in logs.
- **Requirement:** Secrets MUST NEVER be committed in plaintext anywhere in
  platform or service repositories. IaC and pipeline values carrying secrets
  MUST be encrypted at rest using SOPS/AGE-class envelope encryption with keys
  held outside the repository. CI secret variables MUST be masked, prefixed by
  convention, and never echoed. Cluster-consumed secrets MUST be referenced
  (workload identity / external secret store), not embedded. Secret scanning
  MUST be a hard merge gate.
- **Acceptance evidence:** encrypted-values fixtures demonstrating decrypt only
  with external keys; CI configuration proving masking and key-injection
  patterns; secret-scanning gate records including negative fixtures proving
  plaintext rejection; repository history scans clean.
- **Non-goals:** operating the secret store itself (covered by
  identity/security domains); mandating one encryption tool for service teams
  beyond the contract class.
- **Non-claims:** historical-secret rotation procedures are defined but
  bulk-rotation drills are not yet evidenced.
- **Stop conditions:** halt pipelines and revoke/rotate immediately on any
  detected plaintext secret in a repository, log, or artifact (keys / trust /
  exposure).
- **Traceability:** legacy-platform-a; current-core; legacy-platform-b.

### CR-DPL-120 — Remote, locked, versioned IaC state — never committed
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Committed provisioning state files leak sensitive infrastructure
  detail, merge-conflict, and silently diverge under concurrent applies.
  State is operational data, not source code.
- **Requirement:** Infrastructure provisioning state MUST live in a remote,
  access-controlled, versioned backend (object-storage or database class) with
  state locking; state files MUST NEVER be committed to any repository.
  Backend configuration SHOULD be generated per environment/module from
  templates rather than hand-edited. Access to state MUST use dedicated,
  least-privilege automation identities.
- **Acceptance evidence:** repository scans proving zero committed state files;
  backend configuration templates with locking enabled; applied-access policies
  for state backends; concurrency tests showing lock enforcement.
- **Non-goals:** prescribing one backend technology; state encryption beyond
  what the backend class provides (covered by data domains).
- **Non-claims:** migration tooling from any legacy committed-state layout is
  scoped but not yet exercised at scale.
- **Stop conditions:** halt applies when the state lock is contested or state
  shows drift inconsistent with the declared configuration (data / keys /
  migration).
- **Traceability:** legacy-platform-a; legacy-platform-b.

### CR-DPL-130 — Golden image pipeline per datacenter
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Hand-built VM images drift, carry unknown provenance, and
  reproduce vulnerabilities across the fleet. Base images must be built,
  hardened, attested, and distributed as a pipeline product.
- **Requirement:** The platform SHOULD provide a golden image pipeline
  (Packer-class) building hardened base images per datacenter from versioned
  templates: pinned installer checksums, applied hardening configuration,
  standard agents (metrics, logs, inventory, ssh baseline), generated SBOM,
  and manual-gated publication per datacenter. Images MUST be versioned by
  content/commit and traceable to their build inputs.
- **Acceptance evidence:** image pipeline definitions with pinned checksum
  verification; build records including SBOM artifacts; publication gates
  requiring human approval per datacenter; image-to-build traceability
  samples.
- **Non-goals:** application-level images (service CI owns those); supporting
  every guest OS family at launch.
- **Non-claims:** cross-datacenter image distribution via platform-owned
  object storage is designed; end-to-end dogfooded builds on the platform's
  own compute are not yet routine.
- **Stop conditions:** halt publication of an image whose checksum
  verification, hardening profile, or SBOM generation failed (trust /
  exposure).
- **Traceability:** legacy-platform-a; legacy-platform-b.

### CR-DPL-140 — Canary-by-default progressive delivery
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, operator
- **Problem:** All-at-once rollouts turn every release into a potential
  fleet-wide incident. Progressive, metric-gated delivery must be the paved
  road, not an expert option.
- **Requirement:** The standard service deployment contract SHOULD default to
  canary delivery (Flagger-class) with automated, metric-gated promotion and
  automatic rollback on gate failure. Services MAY explicitly opt into a
  plain deployment strategy with a recorded reason. Promotion gates and
  timeouts MUST be declarative and version-controlled with the service.
- **Acceptance evidence:** standard deployment template showing canary as the
  default strategy; test releases demonstrating gate-driven promotion and
  gate-failure rollback; opt-out records with reasons.
- **Non-goals:** mandating canary for batch jobs or one-shot tasks; building
  a custom progressive-delivery controller.
- **Non-claims:** metric-gate quality depends on service-level indicators
  that not all first-party services yet define.
- **Stop conditions:** halt and automatically roll back a canary whose
  promotion gates fail or whose analysis errors (data / trust).
- **Traceability:** legacy-platform-a; current-core.

### CR-DPL-150 — Database-migration stage discipline
- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, operator
- **Problem:** Schema and data migrations applied ad hoc — or after the code
  that needs them — cause a large share of stateful-service incidents.
  Migration must be a first-class, ordered, guarded pipeline stage.
- **Requirement:** Database migrations SHOULD run as a dedicated, ordered
  pipeline stage per environment, gated before the dependent deploy stage.
  Migrations MUST be versioned with the service, MUST declare their rollback
  or forward-recovery strategy, and MUST be blocked by a fresh backup barrier
  for stateful production data. Destructive migrations MUST require explicit
  confirmation and evidence of the backup barrier.
- **Acceptance evidence:** pipeline stage definitions showing migrate-before-
  deploy ordering; migration version records per service; backup-barrier
  receipts attached to production migration runs; drill evidence of rollback
  or forward recovery.
- **Non-goals:** a platform-wide schema registry; online-migration frameworks
  for every engine class.
- **Non-claims:** automated rollback of arbitrary destructive migrations is
  not claimed; forward-recovery is the default honest posture.
- **Stop conditions:** halt a production migration when the backup barrier
  receipt is missing/stale or the migration is destructive without recorded
  confirmation (data / deletion / migration).
- **Traceability:** legacy-platform-a; legacy-platform-b; current-core.

### CR-DPL-160 — Supply-chain integrity: pinned digests, SBOM, provenance, VEX
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, auditor
- **Problem:** Unpinned dependencies and unattested artifacts make the platform
  unanswerable to "what is running and where did it come from?" — a trust
  failure for a platform whose thesis is verifiability.
- **Requirement:** Build and deployment artifacts SHOULD be pinned by digest;
  every released artifact SHOULD carry a generated SBOM and SLSA-class build
  provenance; vulnerability advisories SHOULD be tracked with OpenVEX-class
  statements so triage decisions are auditable. CI MUST reject mutable-tag
  references in release paths. Provenance and SBOM MUST be verifiable by
  consumers of the public repository.
- **Acceptance evidence:** release pipelines emitting SBOM and provenance
  attestations; CI checks rejecting unpinned references in release paths;
  verification instructions reproduced by an independent consumer; VEX
  statement records per advisory triage.
- **Non-goals:** reproducible bit-for-bit builds for all artifacts at launch;
  signing every development-stage artifact.
- **Non-claims:** full SLSA level targets are aspirational for early releases;
  provenance coverage currently spans release artifacts only.
- **Stop conditions:** halt a release when SBOM or provenance generation
  fails, or when an unpinned/mutable dependency is detected in a release path
  (trust / keys).
- **Traceability:** legacy-platform-a; current-core; legacy-platform-b.

### CR-DPL-170 — Versioned module, chart, and template distribution
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, service-team, provider
- **Problem:** Shared automation consumed by floating refs makes environments
  unreproducible and breaks staged rollouts. Reusable building blocks must be
  released and consumed like products.
- **Requirement:** Shared provisioning modules, deployment charts, and pipeline
  templates SHOULD be distributed as semantically versioned artifacts consumed
  by pinned refs, with changelog discipline and deprecation notices. Consumers
  MUST be able to pin, upgrade deliberately, and roll back. Superseded
  versions SHOULD carry explicit retirement markers.
- **Acceptance evidence:** versioned artifact registry or tag-based
  distribution with changelogs; consumer repositories demonstrating pinned
  refs; retirement markers on superseded versions.
- **Non-goals:** a public artifact marketplace for third parties at this
  stage; runtime auto-upgrade of shared modules.
- **Non-claims:** independent release trains per artifact family are
  operational for first-party artifacts; third-party publishing flow is
  undefined.
- **Stop conditions:** halt consumption of an artifact version whose
  changelog records a known data-loss or security defect (trust / data).
- **Traceability:** legacy-platform-a.

### CR-DPL-180 — Environment-scoped deployment isolation for CI
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, auditor
- **Problem:** If any runner can deploy anywhere, a compromised or mis-targeted
  job becomes a cross-environment incident. Deployment capability must be
  scoped to environment and network zone.
- **Requirement:** CI jobs SHOULD execute on runners scoped to the target
  environment's network zone; production jobs MUST run only on
  production-zone runners. Cluster credentials delivered to jobs MUST be
  encrypted (SOPS/vault-class) and short-lived where the toolchain allows;
  plaintext kubeconfig-in-variable is forbidden for production. Deploy
  identity MUST be least-privilege per environment.
- **Acceptance evidence:** runner topology records binding runners to zones;
  pipeline configuration proving production jobs cannot target
  non-production-scoped runners; encrypted credential delivery fixtures;
  negative tests showing cross-zone deploy rejection.
- **Non-goals:** physical isolation of CI infrastructure; per-service runners.
- **Non-claims:** short-lived credential issuance is partially implemented;
  long-lived encrypted kubeconfigs remain in some environment classes.
- **Stop conditions:** halt and investigate any job attempting
  cross-environment credential use or cross-zone targeting (keys / trust /
  exposure).
- **Traceability:** legacy-platform-a; legacy-platform-b.

### CR-DPL-190 — Base cluster templates with deep-merge environment customization
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Copy-pasted cluster configurations per environment drift and
  accumulate silent divergence. Environments should be small, reviewable
  deltas over a governed base.
- **Requirement:** Platform clusters SHOULD be defined by versioned base
  templates (network plugin, ingress posture, control-plane backup schedule,
  audit logging, identity integration) with environment repositories carrying
  only deep-merged customization deltas. Base-template changes MUST be
  changelogged and roll out through the environment ladder like any other
  change.
- **Acceptance evidence:** template repository with versioned changelogs;
  environment repositories showing delta-only content; rendered-configuration
  diff tooling used in review; template-change promotion receipts.
- **Non-goals:** forbidding environment-specific resources outside the
  template's scope; one global template for all installation classes.
- **Non-claims:** template coverage of all cluster classes (DMZ, monitoring,
  edge) is incomplete; edge-class templates are not yet defined.
- **Stop conditions:** halt a base-template rollout when a delta merge
  invalidates declared environment invariants (e.g. disabled audit logging)
  (data / trust / exposure).
- **Traceability:** legacy-platform-a; legacy-platform-b.

### CR-DPL-200 — Change-record and post-incident evidence trail
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, auditor
- **Problem:** Operational changes without recorded pre/post evidence cannot be
  audited, and incident learnings evaporate. A cheap, conventional evidence
  trail multiplies the value of every other gate.
- **Requirement:** Environment repositories SHOULD carry a conventional
  change-record structure (dated change directories with pre/post state and
  audit output) and a post-incident artifact area. Change records SHOULD be
  produced by the same tooling that executes the change, not handwritten
  after the fact.
- **Acceptance evidence:** sampled change directories containing tool-generated
  pre/post artifacts; audit queries demonstrating per-change traceability;
  post-incident records linked to their changes.
- **Non-goals:** a heavyweight ITSM change-management workflow; mandating the
  convention for development environments.
- **Non-claims:** adoption is uneven across environment classes; automation
  that emits records by default is partially built.
- **Stop conditions:** n/a (record-keeping requirement; no direct risk-class
  impact).
- **Traceability:** legacy-platform-a.

### CR-DPL-210 — Garbage-collection automation for transient resources
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, service-team
- **Problem:** Forgotten test releases, stale images, and abandoned sandbox
  environments silently consume capacity and money. Cleanup must be automated
  with bounded, declared policies.
- **Requirement:** The platform SHOULD provide scheduled garbage-collection
  automation covering stale workload releases (age/TTL based), registry image
  retention (age plus keep-last-N), and ephemeral environment teardown. Every
  GC policy MUST be declarative, dry-runnable, and MUST exclude stateful or
  production-tagged resources by default.
- **Acceptance evidence:** GC policy declarations with dry-run reports;
  retention metrics showing bounded growth; exclusion tests proving protected
  resources are never collected.
- **Non-goals:** deleting tenant-owned resources (tenant data lifecycle is
  governed by data domains); cost-optimization recommendations.
- **Non-claims:** GC coverage exists for workload releases and registry
  images; orphaned IaaS-side resources (volumes, addresses) are not yet
  covered.
- **Stop conditions:** halt a GC run whose candidate set includes
  production-tagged or stateful resources (deletion / data).
- **Traceability:** legacy-platform-a.

### CR-DPL-220 — Ephemeral feature/sandbox environments with TTL
- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, vendor
- **Problem:** Shared development environments serialize teams and accumulate
  conflicting state. Per-change ephemeral environments accelerate review and
  testing when their lifecycle is bounded.
- **Requirement:** The delivery tooling SHOULD support ephemeral per-branch or
  per-change environments (namespace-class isolation, convention-derived
  names) with mandatory time-to-live and automatic teardown. Ephemeral
  environments MUST NOT hold production data or production secrets.
- **Acceptance evidence:** pipeline features creating/tearing down ephemeral
  environments; TTL enforcement records; policy checks proving no production
  data or secret classes can be attached.
- **Non-goals:** performance-test-grade environments; externally reachable
  demos from ephemeral environments.
- **Non-claims:** cost attribution per ephemeral environment is not yet
  metered.
- **Stop conditions:** halt creation of an ephemeral environment requesting
  production data sources or production secret classes (data / keys /
  exposure).
- **Traceability:** legacy-platform-a.

---

## Coverage notes

This domain deliberately defers:

- **Secret store operations, workload identity, and authorization policy** for
  deploy-time access → `domains/15-iam-identity-security.md` (IAM). Here we
  only fix delivery-side handling: encryption at rest, injection, masking,
  and scoped deploy identity.
- **Backup/restore content, retention, and DR drills** for platform and tenant
  data → `domains/13-storage-backup-dr.md` (STO). This domain consumes backup
  barriers as deploy gates but does not define backup policy.
- **What gets deployed**: service connector contracts, module registry
  lifecycle semantics, and conformance surfaces →
  `domains/17-ocs-service-connectors.md` (OCS); runtime substrate policy
  (Go-first, upstream Kubernetes) → `domains/10-platform-foundation.md` (FND)
  and `domains/14-kubernetes-containers.md` (K8S).
- **Observability content** (metrics, alerts, dashboards-as-code semantics)
  → `domains/20-observability.md` (OBS); this domain only requires that
  delivery pipelines wire standard agents and monitoring labels.
- **Release content policy beyond delivery mechanics** — BOM component windows,
  upgrade/rollback gates as product contracts → platform foundation and
  release-management concerns shared with FND; this domain pins the
  tag/BOM/atomic-rollout mechanics only.
- **Marketplace distribution of third-party services and licensing/update
  channels** → `domains/18-marketplace-catalog.md` (MKT) and federation
  domains; this domain covers first-party artifact distribution only.
- **SRE operating practices** (on-call, incident response, support tooling)
  → `domains/21-ops-sre-support.md` (OPS); the change-record convention here
  is a delivery artifact, not an incident process.
- **Provider adapter capability semantics** (inventory/preflight/plan/evidence
  classes) → OCS and foundation domains; this domain requires only that
  concrete provider profiles use them downstream.
