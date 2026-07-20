# Matrix — OCS Fit

Generated: 2026-07-20 UTC.
Scope: how each of the 16 requirement domains maps onto the Open Cloud
Standard (OCS) surfaces defined in `domains/17-ocs-service-connectors.md`,
and how the 12 module packages under `cloudring_core/modules/` (backup,
billing-finops, gpu, hpc, marketplace, message-delivery, network,
object-storage, scheduler, slurm, support, universal-iaas) sit on those
surfaces.

## The OCS surface set

OCS is the single contract layer through which every service — first-party,
partner, or third-party — integrates with the platform (CR-OCS-010). For this
matrix the standard is grouped into seven surfaces plus cross-cutting rules:

| Surface | Anchor requirements | What it fixes |
|---|---|---|
| Lifecycle | CR-OCS-030, CR-OCS-040, CR-OCS-050 | Mandatory action set (provision, backup, restore, export, delete, retry, rollback) — idempotent, rollback-referenced; typed user-visible states with remediation; portable inter-service dependencies |
| Billing | CR-OCS-060 | Usage meters with stable names/units, idempotent transmission, replay dedup, rate-card evidence before any charge, explicit non-billable policy |
| Microfrontend | CR-OCS-070 | Portal module host contract: integrity reference, sandbox boundary, allowed events, fail-closed mount refusal, versioned UI guidelines |
| Registry | CR-OCS-020, CR-OCS-080 | Self-registration and one stable service identity across surfaces; module registry with semantic versions, topological dependency resolution, auditable lifecycle operations with rollback hooks |
| Conformance | CR-OCS-090, CR-OCS-100, CR-OCS-110 | Fail-closed validation suite with machine-readable problems; SDK and provider-neutral reference implementation; documented service-team onboarding journey |
| Durability | CR-OCS-140 | Declared state/data classes, backup policy reference, recovery objectives, restore-test evidence gating production enablement |
| Evidence | CR-OCS-150, CR-OCS-160 | One machine receipt format (append-only, freshness-windowed, blocked/stale/synthetic as first-class non-promoting states); publication gate requiring catalog, tenant-access, readiness, support, billing, and durability surfaces plus owner review |
| Cross-cutting | CR-OCS-010, CR-OCS-120, CR-OCS-130, CR-OCS-170, CR-OCS-180, CR-OCS-190, CR-OCS-200 | Package as sole integration unit; versioning/deprecation windows; workload-identity-only secrets; distribution profiles and optional-module honesty; broker-model migration bridge; experimental federation/commercial metadata; declared automation and analytics events |

## Domain × OCS-surface fit

### FND — Platform foundation

FND supplies the rules OCS itself obeys. Contract-before-technology
(CR-FND-010) is realized for services as CR-OCS-010; evidence-over-claims
(CR-FND-130) is the semantic base of the evidence surface (CR-OCS-150);
source-safety (CR-FND-160) scans every connector tree; the Go-first,
upstream-only runtime policy (CR-FND-020, CR-FND-030) bounds connector
runtime targets; the secrets workflow that CR-OCS-130 references is
FND/IAM machinery.
Surfaces consumed: evidence, conformance, cross-cutting (secrets, versioning).
Modules: all 12 inherit these rules; no dedicated foundation module exists —
foundation capabilities are platform core, not packages.

### CMP — Compute and virtualization

Compute products integrate as module packages: instance provision/delete map
onto the lifecycle surface with caller-supplied idempotency keys (CR-OCS-030);
instance conditions render through typed states (CR-OCS-040); volume-backed
state declares durability profiles (CR-OCS-140); per-second usage heartbeats
flow through the billing surface (CR-OCS-060, CR-BIL-110).
Modules: **universal-iaas** (primary IaaS bundle), **gpu** and **hpc**
(optional packages that default to not-installed per CR-OCS-170).

### NET — Network

Tenant networking ships as a module package: virtual networks, subnets,
balancers, addresses, and DNS objects are lifecycle-managed resources
(CR-OCS-030) whose delete paths carry rollback references; billable
addresses, balancers, and capacity units declare meters (CR-OCS-060 with the
money-class stops of CR-NET-060/090); console networking views mount through
the microfrontend contract (CR-OCS-070).
Modules: **network** (primary; declares portable gateway/policy classes, no
provider balancer or endpoint); universal-iaas consumes its portable classes.

### STO — Storage, backup, DR

STO is the home domain of the durability surface: restore drills and the
signed backup barrier (CR-STO-080/140/160) are the machinery behind
CR-OCS-140's restore-test gate; checkpoint/restore/export operations are the
concrete lifecycle actions CR-OCS-030 names; drill results are emitted as
evidence receipts (CR-OCS-150).
Modules: **backup** and **object-storage** (primary); universal-iaas consumes
block-storage classes.

### K8S — Kubernetes and containers

K8S is the substrate connector workloads deploy onto, not itself a connector:
distribution profiles declare infrastructure targets (CR-OCS-170); the
conformance suite runs against upstream-semantics clusters (CR-OCS-090);
cluster lifecycle discipline (CR-K8S-040/060) mirrors the connector lifecycle
contract one layer down. Optional workload modules must validate as
not-installed on foundation-only profiles without failing readiness.
Modules: **scheduler**, **slurm**, **hpc**, **gpu** run on this substrate as
optional packages.

### IAM — Identity and security

IAM powers the tenant-access leg of the publication gate: entitlement
decisions fail closed (CR-OCS-160); workload identity that replaces raw
secrets in packages is IAM-brokered (CR-OCS-130, CR-IAM-090); support
diagnostics inherit IAM's redaction boundary; break-glass and audit surfaces
(CR-IAM-150/160) cover registry and lifecycle operations.
Modules: all 12 declare a tenantAccess surface resolved by IAM; no
IAM-specific module (identity is platform core).

### BIL — Billing and FinOps

BIL owns everything behind the billing surface: OCS fixes only the
transmission contract (CR-OCS-060); the versioned usage-event schema
(CR-BIL-010), ingest, mediation, rating, accounts, and reconciliation
(CR-BIL-020…210) are BIL internals. Connector meters cross-link to the
declared billing connector and fail closed on mismatch.
Modules: **billing-finops** (primary — ledger, allocation, invoice-preview,
settlement contracts); all 12 declare billing meters or an explicit
non-billable policy.

### OCS — OCS service connectors

The standard domain itself: it owns all seven surfaces, the conformance gate
that validates every module package (CR-OCS-090), the SDK and reference
implementation teams code against (CR-OCS-100), and the onboarding journey
(CR-OCS-110, exercised end-to-end by scenario SC-03).
Modules: all 12 are OCSv3 connector packages (`apiVersion`/`kind` manifests)
validated by the conformance suite; each declares the same surface set —
service, billing, catalog, configuration, readiness, tenantAccess,
durability, distribution, federation, commercial.

### MKT — Marketplace and catalog

The catalog renders only from package metadata (CR-OCS-010, CR-OCS-020);
publication through the gate (CR-OCS-160) precedes any listing; product,
pricing, and license semantics (CR-MKT-*) bind to the package's stable
service identity; commercial metadata stays experimental until the owning
planes are ratified (CR-OCS-190).
Modules: **marketplace** (primary — listing, entitlement, order handoff,
revenue-share contracts).

### CUX — Portal, UX, self-service

CUX owns the portal shell that the microfrontend surface plugs into:
connector UI modules mount under integrity verification and sandboxing
(CR-OCS-070) while navigation, parity, and honesty rules remain CUX's
(CR-CUX-*); typed service states (CR-OCS-040) must render identically across
UI, API, CLI, and agent surfaces.
Modules: packages with portal extensions — universal-iaas and object-storage
declare UI surfaces today; any module MAY declare portal modules under
CR-OCS-070.

### OBS — Observability

OBS implements what the readiness and evidence surfaces declare: named
readiness checks with targets and evidence references (CR-OCS-160) are
evaluated from OBS pipelines; analytics events declared by connectors
(CR-OCS-200) respect OBS consent and redaction boundaries; the observability
evidence classes (CR-OBS-210) feed the freshness gates of CR-OCS-150.
Modules: all 12 declare readiness checks; support-diagnostics surfaces flow
through OBS pipelines.

### OPS — Operations, SRE, support

OPS operates the surfaces OCS declares for day-2: the support surface
(owner, redaction-bounded diagnostics, documentation reference — CR-OCS-160);
automation tasks with declared risk classes and rollback behavior
(CR-OCS-200) execute under OPS change and incident discipline (CR-OPS-*);
registry operations surface as auditable operational changes.
Modules: **support** (primary — cases, diagnostics, escalation);
**message-delivery** (notification intent, retry, and metering contract;
transport stays outside core).

### DPL — Deployment, IaC, CI/CD

DPL executes what the registry and distribution surfaces declare: module
install/update/remove/suspend/deprecate operations with rollback hooks and
audit receipts (CR-OCS-080) ride DPL rollout machinery; distribution
profiles, channels, and infrastructure targets (CR-OCS-170) resolve into DPL
environment topologies; the conformance suite runs identically in local
development and in DPL's CI gates (CR-OCS-090, CR-DPL-080).
Modules: all 12 declare a distribution surface consumed by DPL.

### FED — Federation and global portal

FED is the plane OCS federation metadata points at but does not yet activate:
connector federation/commercial declarations are experimental contracts that
MUST NOT drive settlement, licensing, or cross-provider behavior until FED
planes are ratified (CR-OCS-190); base schemas avoid single-provider
assumptions so today's packages remain federatable (CR-FED-010).
Modules: all 12 declare a forward-looking federation surface; no module
derives behavior from it today (honest P1/P2 state).

### DAT — Data services

DAT dogfoods OCS: managed data services onboard through connector packages
like any third-party service (CR-DAT-110) — the shared control-plane pattern
(metadb, task execution, state machine, provider adapters) sits behind the
lifecycle surface (CR-OCS-030); per-engine backup/restore/export declares
durability profiles and restore-test objectives (CR-OCS-140, CR-DAT-050);
engine meters transmit through the billing surface (CR-OCS-060).
Modules: **object-storage** is the storage-class package present today;
engine packages (PostgreSQL-class, streaming-class, cache-class per
CR-DAT-070/080/090) are required by DAT to follow the same package shape —
honestly noted: no dedicated database-engine module exists among the 12 yet,
and CR-DAT-110 is currently a P0 coverage gap (see
`matrices/scenario-coverage.md`).

### AGT — Agent governance

AGT governs the actors, OCS governs the task shape: connectors declare
automation tasks with explicit inputs, risk classes, and rollback behavior
(CR-OCS-200); the agent runtime may execute those tasks only under AGT's
risk-class evidence and approval-tuple discipline (CR-AGT-060/090, scenario
SC-10); agent-executed lifecycle and registry operations emit the same
evidence receipts as any actor (CR-OCS-150).
Modules: none of the 12 is agent-specific; every module's automation
declarations are consumed and governed by the AGT runtime.

## The 12 modules on the OCS surfaces

All 12 packages under `cloudring_core/modules/` share one manifest shape
(service, billing, catalog, configuration, readiness, tenantAccess,
durability, distribution, federation, commercial), which is the
CR-OCS-160 publication-gate surface set. The table maps each module to its
primary requirement domain(s) and the surfaces it most exercises.

| Module | Primary domain(s) | Lifecycle | Billing | Microfrontend | Registry | Conformance | Durability | Evidence |
|---|---|---|---|---|---|---|---|---|
| backup | STO | restore drills as lifecycle actions | usage meters | — | install profiles | package validation | restore-test gate owner | drill receipts |
| billing-finops | BIL | account/charge operations | meter ingest, rating, settlement contracts | — | versioned price bundles | package validation | ledger durability | reconciliation receipts |
| gpu | CMP | accelerator instance lifecycle | accelerator meters | — | optional profile (not-installed default) | package validation | volume-backed state | operation receipts |
| hpc | CMP, K8S | reservation lifecycle | queue/reservation meters | — | optional profile | package validation | declared receipts | compatibility windows |
| marketplace | MKT | listing/order lifecycle | revenue-share contracts | catalog UI | publisher/product versioning | package validation | — | owner-review receipts |
| message-delivery | OPS | notification intent + retry | notification meters | — | channel declarations | package validation | delivery evidence | retry/evidence contract |
| network | NET | network/balancer/address CRUD | address/balancer/capacity meters | console networking views | portable gateway/policy classes | package validation | — | connectivity matrix |
| object-storage | STO, DAT | bucket lifecycle, grants by reference | storage/transfer meters | declared UI surface | storage-class packages | package validation | durability/degraded/denied states | grant and lifecycle receipts |
| scheduler | K8S | queue admission lifecycle | queue meters | — | optional profile (Kueue-class) | package validation | — | compatibility windows |
| slurm | CMP, K8S | partition/account lifecycle | accounting meters | — | optional profile | package validation | — | readiness receipts |
| support | OPS | case/escalation lifecycle | support usage meters | — | service catalog entry | package validation | — | diagnostics with redaction boundary |
| universal-iaas | CMP, NET, STO | full instance/volume/network lifecycle | IaaS meters | portal UI extension | module controller API contract | package validation | volume durability | evidence bundles |

("—" marks surfaces a module does not emphasize in its manifest; the
publication gate still requires the mandatory surface set before any catalog
listing.)

## Honesty notes

- Module alignment is read from the 12 `module-package.json` manifests and
  their READMEs under `cloudring_core/modules/`; a module listed as
  "primary" for a domain does not mean that domain's requirements are
  implemented — all 330 requirements are `proposed` and carry no delivery
  claims in this corpus.
- gpu, hpc, scheduler, and slurm are optional packages: foundation-only
  installations must validate with them not-installed (CR-OCS-170); treating
  any of them as foundation-blocking is a stop condition.
- Four optional packages (gpu, hpc, scheduler, slurm) carry an `xCloudRING`
  extension key in their manifests; per CR-OCS-010 the core consumes only the
  portable contract, never extension internals.
- The DAT dogfooding requirement (CR-DAT-110) is the sharpest open fit gap:
  data services must ride the same connector packages as third parties, but
  no database-engine module exists among the 12 today — recorded here and in
  the scenario-coverage gap list rather than claimed as fitted.
