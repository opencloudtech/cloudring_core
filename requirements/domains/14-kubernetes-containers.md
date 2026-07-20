# 14 — Kubernetes & Containers

Scope: the Kubernetes substrate of the CloudRING platform and the
Kubernetes-based products built on it. This domain covers upstream
kubeadm-semantics cluster provisioning with HA control planes and etcd
backup, declarative Cluster-API-style lifecycle management for platform
and tenant clusters, version support and upgrade policy, the container
image registry and artifact provenance, GitOps-managed add-ons, admission
policy with default-deny baselines, ingress/Gateway-API exposure, the CNI
contract, node lifecycle operations, multi-cluster fleet management, and
cluster audit logging. Tenant workload data protection, host/bare-metal
provisioning, and identity/secrets machinery are owned by sibling domains
and referenced, not redefined, here.

Domain contract: every cluster is unmodified upstream Kubernetes,
reproducible from versioned declarations alone; no cluster state exists
only in someone's shell history. Control planes survive single-server
loss, etcd state is backed up off-cluster and restore is drilled, and
upgrades pass through a verified backup barrier. Admission fails closed:
default-deny policy, approved registries, and explicit public exposure
are baselines, not options. Maintenance never silently kills workloads —
continuity evidence is recorded, and `blocked` is an honest state that
stops the wave rather than being converted into a claim.

## Requirements

### CR-K8S-010 — Upstream-only Kubernetes substrate
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team
- **Problem:** The charter mandates upstream-first runtime semantics, yet
  legacy experience shows lightweight or vendor-wrapped distributions
  create behavioral drift, parallel generations, and upgrade debt that
  break the replaceability contract.
- **Requirement:** All platform and tenant-facing clusters MUST be built
  on unmodified upstream Kubernetes using kubeadm-consistent semantics
  for bootstrap, join, and upgrade. Alternative distributions MAY be used
  only behind the documented cluster contract and MUST NOT become the
  default or a hidden dependency. Every cluster MUST pass an upstream
  conformance test suite before it hosts production workloads.
- **Acceptance evidence:** conformance suite results per cluster and per
  supported version; provisioning code review showing upstream binaries
  and images; clean-host install runbook evidence for the reference
  installation.
- **Non-goals:** forbidding downstream packaging for air-gapped image
  mirrors; fixing a single CNI/CSI choice in this requirement.
- **Non-claims:** conformance across the full supported version window is
  not yet demonstrated on a live reference stand.
- **Stop conditions:** trust/migration — if any required component cannot
  run on unmodified upstream Kubernetes, halt adoption and escalate to
  architecture review; never fork substrate behavior to unblock.
- **Traceability:** `current-core` (Go-first, upstream Kubernetes, kubeadm
  semantics runtime policy), `legacy-platform-a` (cost of long-lived
  parallel distribution generations), `vision-deck`. Related:
  `domains/22-deployment-iac-cicd.md`.

### CR-K8S-020 — Highly available control plane
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator
- **Problem:** A single control-plane node makes the whole platform
  fragile; routine operations and individual hardware failures must not
  interrupt the API, scheduling, or admission chain.
- **Requirement:** Production clusters MUST run a control plane of at
  least three members spread across failure domains, behind a
  load-balanced, health-checked API endpoint. etcd MUST run as a quorum
  of at least three members with documented latency, heartbeat, and
  election parameters. Control-plane certificate rotation MUST be
  automated and MUST NOT require API downtime. Loss of any single
  control-plane server MUST NOT interrupt API service.
- **Acceptance evidence:** one-server-loss drill evidence (a
  control-plane member removed while API availability is continuously
  probed); automated certificate-rotation test with a zero-downtime
  assertion; structural verifier output for control-plane topology.
- **Non-goals:** mandating external versus stacked etcd topology for all
  cluster classes; small-footprint zones MAY use stacked etcd with
  documented availability limits.
- **Non-claims:** one-server-loss evidence does not yet exist for a
  production-grade reference installation.
- **Stop conditions:** data/trust — if etcd quorum health degrades below
  quorum or certificate rotation fails on a staging stand, halt the
  change, page the on-call operator, and do not proceed with dependent
  upgrades.
- **Traceability:** `legacy-platform-a` (versioned cluster templates with
  etcd tuning parameters), `current-core` (one-server-loss evidence
  gate), `req-history`. Related: `domains/21-ops-sre-support.md`.

### CR-K8S-030 — etcd backup and restore
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** etcd holds all cluster state; without scheduled,
  off-cluster, restorable backups, a control-plane corruption event is
  unrecoverable and every cluster mutation is a gamble.
- **Requirement:** Every cluster MUST take scheduled etcd snapshots
  (default cadence at least every 6 hours, retention at least 60
  snapshots unless overridden per environment) written to off-cluster
  object storage with integrity metadata. Snapshot credentials MUST be
  referenced from the approved secrets workflow and MUST never be
  committed to any repository. Restore procedures MUST be documented and
  drilled: a restore-to-fresh-cluster drill MUST succeed at least once
  per quarter for the reference installation before upgrade readiness is
  claimed.
- **Acceptance evidence:** backup schedule and retention configuration in
  the versioned cluster template; restore drill evidence class
  (timestamped restore log plus post-restore conformance smoke test);
  source-safety scan proof that no snapshot credentials exist in git.
- **Non-goals:** tenant workload backup and per-application data
  protection (owned by the storage/backup domain); etcd performance
  tuning beyond documented quorum parameters.
- **Non-claims:** the quarterly restore-drill cadence is a target; no
  completed drill evidence is linked yet.
- **Stop conditions:** data — if snapshot jobs or integrity verification
  fail for more than one cadence interval, halt cluster mutations and any
  durability claims until backups are green again and the gap window is
  recorded.
- **Traceability:** `legacy-platform-a` (template-level scheduled etcd
  backup to object storage; committed backup credentials as an observed
  anti-pattern), `current-core` (backup/restore readiness gates).
  Related: `domains/13-storage-backup-dr.md`.

### CR-K8S-040 — Declarative cluster lifecycle provisioning
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, agent
- **Problem:** Hand-assembled clusters drift and cannot be reproduced;
  every cluster must be reconstructible from versioned declarations
  alone, and every mutation must flow through reconciliation.
- **Requirement:** Cluster lifecycle (create, scale, upgrade, delete)
  MUST be driven by declarative, versioned cluster specifications
  reconciled by controllers (Cluster-API-style infrastructure and machine
  abstractions). Environment repositories MUST carry only deltas over a
  versioned base cluster template (deep-merge customization model).
  Bootstrap MUST be self-referencing: the management tooling reconciles
  each environment from its own repository. Manual node edits are treated
  as drift and MUST be detected and reverted or reconciled back into
  declarations.
- **Acceptance evidence:** destroy-and-recreate drill (a non-production
  cluster deleted and rebuilt from repository declarations alone);
  drift-detection and self-heal evidence; base template plus
  per-environment delta layout versioned in git.
- **Non-goals:** mandating a specific upstream provisioning project;
  forbidding imperative escape hatches for emergency recovery (their use
  MUST be logged and folded back into declarations).
- **Non-claims:** the full recreate-from-git drill has not yet passed on
  a reference stand.
- **Stop conditions:** keys/exposure — if provisioning appears to require
  embedding private keys or cloud credentials in repositories, halt and
  redesign around the secrets workflow; never commit credentials to
  unblock a bring-up.
- **Traceability:** `legacy-platform-a` (self-referencing GitOps
  environment repositories; Cluster-API provider line replacing a legacy
  imperative aggregator), `legacy-platform-b` (dependency-ordered
  bootstrap graphs kept as code), `current-core`. Related:
  `domains/22-deployment-iac-cicd.md`, `domains/10-platform-foundation.md`.

### CR-K8S-050 — Tenant Kubernetes-as-a-Service
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Tenants need self-service clusters isolated from platform
  internals and from each other, delivered through the same contract
  surfaces as every other platform service.
- **Requirement:** The platform SHOULD offer tenant-facing clusters
  through the OCS service-connector model: declarative cluster and
  node-pool resources, asynchronous operations with operation IDs,
  quota-aware provisioning, and kubeconfig delivery using short-lived
  credentials issued through the platform identity provider (no
  long-lived static admin kubeconfigs by default). Tenant clusters MUST
  be isolated from platform management-plane credentials and networks.
  Tenant cluster lifecycle states MUST be explicit (provisioning,
  running, suspending, upgrading, deleting, error).
- **Acceptance evidence:** OCS connector contract validation for the
  Kubernetes product; end-to-end test (tenant orders a cluster, receives
  kubeconfig, deploys a workload, deletes the cluster with cleanup
  proof); isolation test proving a tenant cluster cannot reach platform
  management endpoints.
- **Non-goals:** prescribing dedicated versus shared control-plane models
  per tenant — both are allowed behind the contract; tenant in-cluster
  application management.
- **Non-claims:** multi-tenant noisy-neighbor performance isolation is
  unproven; the suspension-on-nonpayment flow is specified but not yet
  drill-tested.
- **Stop conditions:** trust/exposure — if any path lets tenant
  credentials address the platform management plane, fail closed, revoke
  the credentials, and escalate as a security incident.
- **Traceability:** `legacy-platform-a` (two coexisting KaaS generations;
  per-region cluster routing; cluster certificate issuance from a central
  CA), `current-core` (OCSv3 connector contract). Related:
  `domains/17-ocs-service-connectors.md`, `domains/15-iam-identity-security.md`.

### CR-K8S-060 — Version support and upgrade policy
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, tenant
- **Problem:** Unplanned version drift breaks workloads and security
  posture; upgrades must be routine, tested, and gated instead of heroic
  one-off events that strand clusters on unsupported versions.
- **Requirement:** The platform MUST support the upstream N-2
  minor-version window for cluster control planes, MUST publish a tested
  upgrade path between every supported adjacent version pair, and MUST
  gate every upgrade on a fresh verified etcd/backup snapshot (backup
  barrier) plus a workload-drain preflight. Control-plane upgrades MUST
  complete without downtime for conformant workloads (replicated, with
  disruption budgets). Rollback MUST be documented; downgrade MUST NOT be
  claimed unless a restore-based path is executed and evidenced.
- **Acceptance evidence:** upgrade test suite per version pair on a
  staging stand (pre/post conformance plus workload-continuity probe);
  backup-barrier enforcement check in the upgrade tooling (refuses to
  proceed without a fresh snapshot); published version-support matrix.
- **Non-goals:** supporting minor-version skips as an upgrade path;
  automatic minor-version upgrades without an explicit operator or
  tenant trigger.
- **Non-claims:** the N-2 window is policy; continuous multi-version test
  coverage is not yet in place.
- **Stop conditions:** migration/data — if pre-upgrade backup
  verification fails, the drain preflight cannot complete, or
  post-upgrade conformance regresses, halt the wave and escalate; never
  force an upgrade through red gates.
- **Traceability:** `legacy-platform-a` (dual-generation clusters
  coexisting for years — the cost of having no enforced version policy),
  `legacy-platform-b` (written environment-ladder regulation),
  `current-core` (kubeadm upgrade semantics). Related:
  `domains/22-deployment-iac-cicd.md`, `domains/21-ops-sre-support.md`.

### CR-K8S-070 — Container image registry and mirroring
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, vendor
- **Problem:** Clusters and CI need a reliable, jurisdiction-controllable
  source of images; uncontrolled external pulls break sovereign and
  air-gapped deployments and couple platform availability to third-party
  registries.
- **Requirement:** The platform MUST provide an OCI-compliant registry
  (or registry mirror set) through which all platform and workload images
  are pulled, and clusters MUST be configurable to pull only from
  approved registries. All third-party images required for installation
  MUST be mirrored into the platform registry with pinned digests.
  Retention and garbage-collection policies (age plus keep-last-N) MUST
  be defined and executed automatically. Registry content MUST be backed
  up or be reconstructible from declared sources of truth.
- **Acceptance evidence:** air-gap install test (installation completes
  with external egress blocked); image-mirror manifest with pinned
  digests; registry garbage-collection job evidence; admission check
  proof that unapproved registries are rejected.
- **Non-goals:** building a general artifact-management suite (language
  package proxies MAY follow later); mandating a single registry product.
- **Non-claims:** a fully air-gapped reference installation is planned
  but not yet evidenced end-to-end.
- **Stop conditions:** exposure/trust — if required images resolve to
  unpinned or unapproved external sources at install or runtime, fail
  closed and treat it as a supply-chain defect.
- **Traceability:** `legacy-platform-a` (internal registry with an
  image-mirroring tool and retention automation), `legacy-platform-b`
  (registry as an explicit bootstrap dependency of the platform-services
  wave), `vision-deck` (jurisdiction freedom). Related:
  `domains/22-deployment-iac-cicd.md`, `domains/15-iam-identity-security.md`.

### CR-K8S-080 — Artifact signing, SBOM, and provenance
- **Priority:** P1
- **Status:** proposed
- **Actors:** vendor, service-team, auditor, operator
- **Problem:** Without provenance, operators cannot answer "what is
  running and where did it come from" during incidents, audits, or
  supply-chain disclosures.
- **Requirement:** Platform-built images SHOULD be signed
  (cosign-class), carry software bills of materials (syft-class), and
  include build-provenance attestations. Admission policy SHOULD be able
  to require signature verification for platform namespaces, and the
  SBOM/provenance store SHOULD be queryable per image digest.
  Verification failures MUST fail closed wherever enforcement is
  enabled.
- **Acceptance evidence:** CI pipeline stage emitting signature, SBOM,
  and provenance for every platform image; admission-enforcement test
  (an unsigned image is denied in enforced namespaces); auditor
  walkthrough mapping running pods to digests to SBOM entries.
- **Non-goals:** requiring signature enforcement in all tenant namespaces
  at baseline (an opt-in policy class); vulnerability-scanning policy
  (owned by security tooling requirements).
- **Non-claims:** provenance coverage across all platform images is not
  yet achieved, so enforcement defaults are intentionally conservative
  until coverage matures.
- **Stop conditions:** trust — if signature-verification infrastructure
  is unavailable while enforcement is enabled, deny new admissions in
  enforced namespaces rather than silently allowing them.
- **Traceability:** `legacy-platform-a` (SBOM generation on base images;
  supply-chain hygiene tooling), `current-core` (source-safety and
  fail-closed principles). Related: `domains/15-iam-identity-security.md`,
  `domains/22-deployment-iac-cicd.md`.

### CR-K8S-090 — GitOps-managed cluster add-ons
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, agent
- **Problem:** Cluster add-ons (CNI, CSI, ingress, policy, observability
  agents) installed by hand drift between clusters and environments,
  making behavior irreproducible and recovery slow.
- **Requirement:** All cluster add-ons MUST be declared in git and
  reconciled by a GitOps agent (Flux/Argo-class) with automated sync,
  health assessment, and drift self-healing. Add-ons MUST be layered: a
  platform-core layer (CNI, CSI, policy, identity integration) applied
  before a shared-services layer. Add-on versions MUST be pinned by tag
  or digest, and per-environment values MUST be co-versioned with the
  environment repository. Prune behavior MUST be deliberate and
  documented per add-on so a reconciler defect cannot mass-delete
  workloads.
- **Acceptance evidence:** drift report proving repository inventory
  matches live cluster state; recovery drill (a manually deleted add-on
  is restored by the reconciler); staged rollout evidence for an add-on
  version change across development, staging, and production.
- **Non-goals:** prescribing a single GitOps engine; managing tenant
  application workloads through the platform add-on layer.
- **Non-claims:** prune and self-heal behavior is not yet drill-verified
  against destructive reconciler failure modes.
- **Stop conditions:** keys/deletion — if reconciler credentials are
  compromised or a reconciliation loop begins deleting out-of-scope
  resources, freeze sync, revoke the credentials, and escalate before
  resuming.
- **Traceability:** `legacy-platform-a` (dual GitOps engines for
  infrastructure and application tracks; layered core/common add-on
  bundles; app-of-apps bootstrap), `current-core` (GitOps production
  contracts). Related: `domains/22-deployment-iac-cicd.md`,
  `domains/21-ops-sre-support.md`.

### CR-K8S-100 — Admission policy with default-deny baselines
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, auditor, tenant
- **Problem:** Without enforced baselines, any workload can run
  privileged, mount host paths, or open network paths, and tenant
  isolation depends on hope rather than mechanism.
- **Requirement:** Every cluster MUST run an admission policy engine
  (Kyverno/OPA-class) enforcing the platform baseline: restricted
  pod-security profile by default, no privilege escalation, no host
  namespaces or host paths without explicit exemption, mandatory resource
  requests and limits in platform namespaces, and default-deny
  NetworkPolicy for ingress and egress in every namespace with explicit
  allows. The engine MUST fail closed: when it is unavailable, admission
  of new workloads MUST be denied except for audited break-glass system
  accounts. Baseline policies MUST be versioned and rolled out through
  the GitOps layer.
- **Acceptance evidence:** policy test suite (non-conformant workloads
  denied, conformant workloads admitted); fail-closed drill (policy
  webhook outage leads to denied admission); versioned baseline policy
  set with change history; default-deny NetworkPolicy verification across
  namespaces.
- **Non-goals:** per-tenant custom policy authoring at baseline (a later
  policy-as-a-service extension); runtime kernel-level threat detection.
- **Non-claims:** exemption-workflow ergonomics are unvalidated with real
  service teams; the performance cost of full enforcement at scale is
  unmeasured.
- **Stop conditions:** exposure/trust — if enforcement gaps are
  discovered (namespaces without default-deny, exemptions beyond the
  approved list), halt onboarding, remediate, and audit what ran during
  the gap.
- **Traceability:** `legacy-platform-a` (policy engine constraints
  deployed via GitOps; a default-allow network posture explicitly flagged
  as incomplete), `current-core` (fail-closed principle). Related:
  `domains/15-iam-identity-security.md`, `domains/12-network.md`.

### CR-K8S-110 — Ingress and Gateway API exposure contract
- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, tenant, operator
- **Problem:** Workloads need a stable, documented north-south traffic
  entry that survives controller swaps, keeps TLS universal, and makes
  every public listener explicit.
- **Requirement:** The platform MUST define a north-south exposure
  contract with Gateway-API resources as the primary model (Ingress API
  supported for compatibility), implemented by a replaceable controller
  class. TLS MUST be terminated with certificates issued and rotated by
  the platform certificate workflow; plaintext HTTP MUST redirect or be
  refused by default for platform services. External (DMZ-class) and
  internal traffic entry SHOULD be separable per cluster or zone so that
  public exposure is always an explicit declaration.
- **Acceptance evidence:** contract test suite running identical
  Gateway/Ingress fixtures against the chosen controller; certificate
  rotation drill with no dropped TLS handshakes; exposure audit listing
  every public listener with its owning declaration.
- **Non-goals:** L4 load-balancer product semantics (owned by the network
  domain); east-west mesh traffic management.
- **Non-claims:** Gateway-API coverage of advanced traffic-splitting
  features depends on controller maturity and is not yet claimed.
- **Stop conditions:** exposure — if a change would publish a service
  publicly without an explicit declaration and certificate, deny it and
  alert; any unexpected public listener triggers immediate review.
- **Traceability:** `legacy-platform-a` (dedicated external/DMZ ingress
  clusters; controller installed explicitly rather than by distribution
  default), `legacy-platform-b` (documented listener and certificate
  model for load balancing), `vision-deck`. Related:
  `domains/12-network.md`, `domains/15-iam-identity-security.md`.

### CR-K8S-120 — CNI networking contract
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, service-team
- **Problem:** Pod networking is foundational and expensive to change
  later; it must be replaceable behind a contract while guaranteeing
  enforceable NetworkPolicy from day one.
- **Requirement:** Every cluster MUST run a CNI implementation that
  enforces Kubernetes NetworkPolicy and is declared, versioned, and
  deployed through the GitOps add-on layer. The CNI choice MUST sit
  behind the cluster contract — IPAM mode, MTU, and encapsulation
  documented per environment — so implementations can be swapped per
  cluster class without platform code changes. A default-allow network
  posture is forbidden: namespace creation MUST include default-deny
  NetworkPolicy (per CR-K8S-100). Two CNI implementation families MAY be
  supported as per-environment options.
- **Acceptance evidence:** NetworkPolicy enforcement test matrix (deny
  and allow cases including egress); CNI swap drill on a non-production
  cluster class; per-environment network parameter documentation
  generated from the environment repository.
- **Non-goals:** mandating eBPF data paths; multi-cluster pod networking
  or cluster mesh at baseline.
- **Non-claims:** the multi-CNI support matrix is defined, but only one
  implementation is expected to be drill-verified initially.
- **Stop conditions:** exposure — if a CNI change disables policy
  enforcement (silently failing open), roll back immediately and audit
  the traffic that flowed during the window.
- **Traceability:** `legacy-platform-a` (two CNI families as
  per-environment template options with documented parameters),
  `current-core` (contract before technology). Related:
  `domains/12-network.md`.

### CR-K8S-130 — Node lifecycle operations with workload continuity
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, agent
- **Problem:** Nodes must be patched and replaced routinely; unsafe
  drains are the classic way a platform loses tenant workloads and
  operator trust.
- **Requirement:** Node maintenance MUST follow the cordon, drain
  (honoring PodDisruptionBudgets), operate, uncordon-or-replace flow,
  automated through the declarative lifecycle layer. The platform MUST
  NOT claim maintenance safety for workloads lacking replicas or
  disruption budgets; a pre-drain preflight MUST surface such workloads
  and require explicit override. Worker node replacement MUST be routine:
  any worker replaceable from declaration without manual steps. Workload
  continuity evidence (an availability probe per protected service) MUST
  be recorded for production maintenance operations.
- **Acceptance evidence:** maintenance drill on a production-class
  cluster (nodes drained and replaced while continuous availability
  probes stay green); disruption-budget blocking test (drain refuses to
  proceed when budgets cannot be honored); preflight report fixture
  listing unprotected workloads.
- **Non-goals:** automatically remedying tenant workloads that are
  inherently non-HA (they are surfaced, not fixed); bare-metal node
  provisioning (owned by the foundation domain).
- **Non-claims:** continuity-evidence collection is specified but not yet
  wired into the maintenance tooling for all cluster classes.
- **Stop conditions:** data — if a drain would evict stateful workloads
  whose disruption budget or storage topology cannot tolerate it, halt at
  cordon, and escalate to the owning team with the preflight report.
- **Traceability:** `legacy-platform-a` (node-drain tooling in cluster
  configuration management; lifecycle guards protecting attached
  volumes), `current-core` (workflow-continuity evidence convention).
  Related: `domains/21-ops-sre-support.md`, `domains/13-storage-backup-dr.md`.

### CR-K8S-140 — Multi-cluster fleet management
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Providers will run many clusters across regions, zones,
  and roles; without fleet-level inventory and uniform governance, each
  cluster becomes a snowflake and every rollout a bespoke risk.
- **Requirement:** The platform SHOULD maintain a fleet inventory: every
  cluster registered with identity, region, zone, role (platform payload,
  external ingress, monitoring, tenant), version, and health, with
  environment repositories as the source of truth. Fleet operations
  SHOULD include uniform policy and add-on rollout waves with canary
  clusters, an aggregated compliance view (policy, version, and audit
  posture per cluster), and blast-radius-aware rollout ordering. Naming
  and topology conventions MUST make every cluster's region, zone,
  environment, and role derivable from its identity.
- **Acceptance evidence:** fleet inventory reconciliation test
  (repository truth versus registered clusters); staged wave rollout
  evidence from canary to broad wave; compliance export listing
  per-cluster posture.
- **Non-goals:** cross-cluster workload scheduling or federation (owned
  by the federation domain); a hosted multi-tenant fleet SaaS.
- **Non-claims:** fleet-scale behavior is extrapolated from small-count
  environments and is unproven at target scale.
- **Stop conditions:** trust — if fleet-level credentials or the
  management plane lose reachability or attestation of member clusters,
  freeze fleet-wide mutations to prevent split-brain rollouts.
- **Traceability:** `legacy-platform-a` (per-cluster environment
  repositories with role-split topologies; versioned cluster templates),
  `legacy-platform-b` (per-stand configuration layout), `vision-deck`
  (multi-zone provider model). Related: `domains/23-federation-global-portal.md`,
  `domains/21-ops-sre-support.md`.

### CR-K8S-150 — Kubernetes audit logging
- **Priority:** P0
- **Status:** proposed
- **Actors:** auditor, operator, provider
- **Problem:** Every API action on a cluster must be attributable after
  the fact; without durable audit logs, security review and incident
  forensics are impossible.
- **Requirement:** Every cluster MUST enable API-server audit logging
  with a documented policy (metadata at minimum for all requests;
  request and response bodies for secret and token-adjacent access),
  shipped off-node to the central log store with UTC-timestamped,
  append-only, tamper-evident retention. Audit coverage MUST include
  authentication and authorization decisions, admission denials, and
  credential or secret access. Retention periods MUST be defined per
  environment and meet compliance needs. Gaps in audit delivery MUST
  alert.
- **Acceptance evidence:** audit policy file in the versioned cluster
  template; end-to-end test (a scripted sequence of API actions is found
  complete in the central store); gap-detection alert test (log shipping
  stopped leads to a fired alert); retention configuration per
  environment.
- **Non-goals:** full SIEM correlation (owned by observability and
  security tooling); tenant application log collection.
- **Non-claims:** tamper-evidence properties of the log store depend on
  the chosen backend and are not yet verified.
- **Stop conditions:** trust — if audit delivery gaps exceed the alerting
  threshold on a production cluster, freeze non-emergency API-mutating
  automation until the pipeline is restored and the gap window is
  recorded.
- **Traceability:** `legacy-platform-a` (API audit logging enabled in
  cluster templates with sized retention), `current-core` (append-only,
  UTC, evidence-over-claims rules). Related: `domains/20-observability.md`,
  `domains/15-iam-identity-security.md`.

### CR-K8S-160 — Cluster autoscaling and elastic node groups
- **Priority:** P2
- **Status:** proposed
- **Actors:** tenant, operator, provider
- **Problem:** Fixed node pools waste money at low load and fail
  workloads at peak; elasticity is valuable but dangerous if unbounded.
- **Requirement:** Node groups SHOULD support automated scale-out and
  scale-in driven by pending-workload and utilization signals, integrated
  with the declarative cluster specification (minimum and maximum bounds,
  per-pool policies). Scale-in MUST honor disruption budgets and the
  node-lifecycle flow of CR-K8S-130. Autoscaling decisions MUST be
  observable through events and metrics and MUST be bounded by quota so
  runaway scaling cannot create unbounded cost.
- **Acceptance evidence:** load test (synthetic pending workload triggers
  scale-out within bounds and scale-in after); budget-honoring scale-in
  test; cost-bound test (scaling stops at quota or maximum with a
  surfaced event).
- **Non-goals:** vertical pod autoscaling as a platform default;
  predictive or ML-driven scaling.
- **Non-claims:** scale-in safety with diverse stateful workload types is
  not yet evidenced; quota coupling is designed but untested.
- **Stop conditions:** money — if scaling oscillation or unbounded
  scale-out is detected, clamp to current size, alert, and require
  operator acknowledgement before resuming automation.
- **Traceability:** `legacy-platform-a` (autoscaling conventions in the
  universal deployment chart; quota services), `legacy-platform-b`
  (autoscaling role in host configuration). Related:
  `domains/16-billing-finops.md`.

### CR-K8S-170 — Service mesh readiness
- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, operator
- **Problem:** Some services expect advanced traffic management and
  mutual TLS, but a mesh must never become an unremovable platform
  dependency.
- **Requirement:** The platform MAY offer a service mesh (Istio/Linkerd
  class) as a versioned add-on behind the same GitOps and policy
  contract; the platform core MUST function with the mesh disabled. If
  offered, mutual TLS enablement, ingress-gateway integration with the
  north-south contract of CR-K8S-110, and canary deployment primitives
  MUST be documented and tested. Mesh enrollment per namespace MUST be
  opt-in and labeled.
- **Acceptance evidence:** install/remove drill (mesh added to and fully
  removed from a cluster without workload impact); canary promotion
  end-to-end test using mesh traffic splitting; mutual TLS verification
  test for enrolled namespaces.
- **Non-goals:** making a mesh mandatory for OCS service connectors;
  multi-cluster mesh at baseline.
- **Non-claims:** the operational cost and upgrade cadence of a supported
  mesh are not yet validated; no mesh SLA is claimed.
- **Stop conditions:** trust — if mutual TLS verification regresses or
  the mesh control plane fails open to plaintext for enrolled services,
  halt enrollment expansion and treat it as a security defect.
- **Traceability:** `legacy-platform-a` (mesh operator plus canary
  delivery as standard add-ons; ingress and egress gateway split),
  `current-core` (contract before technology). Related:
  `domains/12-network.md`, `domains/22-deployment-iac-cicd.md`.

### CR-K8S-180 — Ephemeral sandbox clusters and namespaces
- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, vendor, operator
- **Problem:** Development and CI need disposable environments; without
  time-boxing and automated cleanup they leak cost and attack surface.
- **Requirement:** The platform SHOULD support time-boxed sandbox
  clusters and namespaces with automatic expiry: TTL labels, automated
  cleanup of releases, namespaces, and clusters past expiry, and owner
  notification before deletion. Sandbox environments MUST be constrained
  by policy to non-production data and endpoints.
- **Acceptance evidence:** cleanup job evidence (expired resources
  deleted, owners notified); policy test proving sandbox resources cannot
  reach production endpoints; cost report for the sandbox estate.
- **Non-goals:** per-developer production access patterns; performance
  parity of sandboxes with production.
- **Non-claims:** cleanup reliability with long-lived custom resources
  and finalizers is a known hard case and is not yet fully validated.
- **Stop conditions:** deletion/money — cleanup automation MUST dry-run
  and report before its first destructive run in any environment; any
  deletion of non-expired resources halts the cleaner and triggers an
  incident review.
- **Traceability:** `legacy-platform-a` (feature namespaces plus
  stale-release garbage collection), `current-core` (production honesty).
  Related: `domains/22-deployment-iac-cicd.md`.

## Coverage notes

This domain deliberately defers:

- **Host, bare-metal, and OS provisioning** for cluster nodes to
  `domains/10-platform-foundation.md`; this domain consumes nodes through
  the declarative lifecycle contract.
- **Virtual machine workloads on Kubernetes** (KubeVirt-class) to
  `domains/11-compute-virtualization.md`.
- **VPC, L4/L7 load-balancer products, SDN, and interconnect** to
  `domains/12-network.md`; here only the in-cluster CNI and north-south
  exposure contracts live.
- **Persistent volume data protection, tenant workload backup, and
  disaster recovery** to `domains/13-storage-backup-dr.md`; this domain
  owns only etcd state backup.
- **Identity, OIDC cluster access, secrets brokering, and certificate
  authority operation** to `domains/15-iam-identity-security.md`; this
  domain consumes those services and cites them in stop conditions.
- **Metrics, logs, alerting, and SIEM correlation** to
  `domains/20-observability.md`; this domain owns only the cluster audit
  pipeline as a security baseline.
- **CI/CD pipelines, environment ladders, and release mechanics** to
  `domains/22-deployment-iac-cicd.md`; GitOps here covers in-cluster
  reconciliation only.
- **OCS connector contract details** for the tenant Kubernetes product to
  `domains/17-ocs-service-connectors.md`.
- **Cross-cloud cluster federation and workload placement across
  providers** to `domains/23-federation-global-portal.md`.
- **Agent authority bounds for autonomous cluster operations** to
  `domains/25-agent-governance.md`.
