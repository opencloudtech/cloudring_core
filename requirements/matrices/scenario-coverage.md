# Matrix — Scenario Coverage

Generated: 2026-07-20 UTC.
Scope: `scenarios/sc-01 … sc-10` against `domains/*.md` (16 domains, 330
requirements) and `registry/requirements.json` (generated, zero drift at time
of writing; see `registry/validation-report.md`).

## Method

- The scenario → requirement mapping is taken from the **Requirement coverage**
  table inside each scenario file (the canonical per-scenario mapping; step-level
  `Requirements:` lines agree with it).
- Requirement metadata (domain, priority) is taken from
  `registry/requirements.json`, which is generated from the Markdown corpus.
- A requirement counts as covered by a scenario when its ID appears in that
  scenario's coverage table. An ID may be exercised by more than one scenario.
- No requirement ID appears in a scenario table that is absent from the
  registry (checked: zero dangling references).

## Scenario index

| ID | Title |
|---|---|
| SC-01 | Fresh install to first VM (single host) |
| SC-02 | Reference-provider production install (IaC bootstrap DAG) |
| SC-03 | Service team onboards a new OCS service |
| SC-04 | Tenant buys a marketplace product and is billed correctly |
| SC-05 | Backup and restore drill promotes a release |
| SC-06 | One-server-loss continuity |
| SC-07 | Upgrade with backup barrier |
| SC-08 | Tenant exit and data portability |
| SC-09 | Security incident with break-glass access |
| SC-10 | Agent executes a risky change with an approval tuple |

## Coverage matrix (requirement count per scenario × domain)

| Scenario | FND | CMP | NET | STO | K8S | IAM | BIL | OCS | MKT | CUX | OBS | OPS | DPL | FED | DAT | AGT | Total |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| SC-01 | 6 | 7 | 4 | — | 1 | 2 | 1 | — | — | 1 | 2 | — | 5 | — | — | — | 29 |
| SC-02 | 4 | 1 | 2 | 3 | 3 | 4 | — | — | — | — | 1 | — | 11 | — | — | — | 29 |
| SC-03 | 5 | — | — | 1 | — | 1 | 2 | 17 | 1 | 1 | 1 | — | 1 | — | — | — | 30 |
| SC-04 | 2 | — | — | — | — | — | 14 | 2 | 10 | 2 | — | — | — | — | — | — | 30 |
| SC-05 | 1 | — | — | 8 | 1 | — | — | 1 | — | — | 1 | 1 | 2 | — | 2 | — | 17 |
| SC-06 | 1 | 5 | 2 | 2 | 2 | — | — | — | — | — | 5 | 4 | 1 | — | — | — | 22 |
| SC-07 | 2 | — | 1 | 5 | 3 | 1 | — | — | — | — | 2 | 3 | 6 | — | 2 | — | 25 |
| SC-08 | 1 | — | — | 2 | — | 3 | 2 | — | — | 3 | — | — | — | 2 | 1 | — | 14 |
| SC-09 | — | — | — | 1 | — | 6 | — | — | — | — | 1 | 6 | — | — | — | 4 | 18 |
| SC-10 | 1 | — | — | 1 | — | — | — | — | — | — | — | 1 | — | — | — | 17 | 20 |
| **Assignments** | 23 | 13 | 9 | 23 | 10 | 17 | 19 | 20 | 11 | 7 | 13 | 15 | 26 | 2 | 5 | 21 | 234 |

234 scenario-requirement assignments exercise **167 unique requirement IDs**
of 330 (51%). Domain codes per `01-requirement-schema.md`.

## Requirement IDs exercised, per scenario and domain

### SC-01 — Fresh install to first VM (single host)

| Domain | Requirements |
|---|---|
| FND | CR-FND-030, CR-FND-060, CR-FND-080, CR-FND-120, CR-FND-130, CR-FND-160 |
| CMP | CR-CMP-010, CR-CMP-020, CR-CMP-030, CR-CMP-070, CR-CMP-130, CR-CMP-140, CR-CMP-160 |
| NET | CR-NET-010, CR-NET-040, CR-NET-050, CR-NET-190 |
| K8S | CR-K8S-010 |
| IAM | CR-IAM-010, CR-IAM-140 |
| BIL | CR-BIL-110 |
| CUX | CR-CUX-040 |
| OBS | CR-OBS-150, CR-OBS-210 |
| DPL | CR-DPL-010, CR-DPL-030, CR-DPL-040, CR-DPL-050, CR-DPL-110 |

### SC-02 — Reference-provider production install (IaC bootstrap DAG)

| Domain | Requirements |
|---|---|
| FND | CR-FND-030, CR-FND-040, CR-FND-130, CR-FND-160 |
| CMP | CR-CMP-070 |
| NET | CR-NET-120, CR-NET-190 |
| STO | CR-STO-130, CR-STO-140, CR-STO-160 |
| K8S | CR-K8S-020, CR-K8S-030, CR-K8S-080 |
| IAM | CR-IAM-130, CR-IAM-140, CR-IAM-160, CR-IAM-170 |
| OBS | CR-OBS-210 |
| DPL | CR-DPL-020, CR-DPL-030, CR-DPL-040, CR-DPL-050, CR-DPL-090, CR-DPL-100, CR-DPL-110, CR-DPL-120, CR-DPL-130, CR-DPL-160, CR-DPL-190 |

### SC-03 — Service team onboards a new OCS service

| Domain | Requirements |
|---|---|
| FND | CR-FND-010, CR-FND-080, CR-FND-090, CR-FND-130, CR-FND-140 |
| STO | CR-STO-110 |
| IAM | CR-IAM-090 |
| BIL | CR-BIL-130, CR-BIL-180 |
| OCS | CR-OCS-010, CR-OCS-020, CR-OCS-030, CR-OCS-040, CR-OCS-050, CR-OCS-060, CR-OCS-070, CR-OCS-080, CR-OCS-090, CR-OCS-100, CR-OCS-110, CR-OCS-120, CR-OCS-130, CR-OCS-140, CR-OCS-150, CR-OCS-160, CR-OCS-170 |
| MKT | CR-MKT-080 |
| CUX | CR-CUX-130 |
| OBS | CR-OBS-210 |
| DPL | CR-DPL-080 |

### SC-04 — Tenant buys a marketplace product and is billed correctly

| Domain | Requirements |
|---|---|
| FND | CR-FND-120, CR-FND-140 |
| BIL | CR-BIL-020, CR-BIL-030, CR-BIL-040, CR-BIL-050, CR-BIL-060, CR-BIL-070, CR-BIL-080, CR-BIL-110, CR-BIL-120, CR-BIL-130, CR-BIL-140, CR-BIL-160, CR-BIL-180, CR-BIL-210 |
| OCS | CR-OCS-030, CR-OCS-060 |
| MKT | CR-MKT-010, CR-MKT-020, CR-MKT-030, CR-MKT-040, CR-MKT-050, CR-MKT-060, CR-MKT-070, CR-MKT-090, CR-MKT-110, CR-MKT-170 |
| CUX | CR-CUX-010, CR-CUX-040 |

### SC-05 — Backup and restore drill promotes a release

| Domain | Requirements |
|---|---|
| FND | CR-FND-130 |
| STO | CR-STO-060, CR-STO-070, CR-STO-080, CR-STO-100, CR-STO-110, CR-STO-130, CR-STO-140, CR-STO-160 |
| K8S | CR-K8S-030 |
| OCS | CR-OCS-140 |
| OBS | CR-OBS-080 |
| OPS | CR-OPS-180 |
| DPL | CR-DPL-070, CR-DPL-100 |
| DAT | CR-DAT-080, CR-DAT-100 |

### SC-06 — One-server-loss continuity

| Domain | Requirements |
|---|---|
| FND | CR-FND-130 |
| CMP | CR-CMP-010, CR-CMP-040, CR-CMP-110, CR-CMP-120, CR-CMP-170 |
| NET | CR-NET-190, CR-NET-210 |
| STO | CR-STO-040, CR-STO-150 |
| K8S | CR-K8S-020, CR-K8S-130 |
| OBS | CR-OBS-030, CR-OBS-080, CR-OBS-100, CR-OBS-130, CR-OBS-210 |
| OPS | CR-OPS-050, CR-OPS-060, CR-OPS-090, CR-OPS-180 |
| DPL | CR-DPL-030 |

### SC-07 — Upgrade with backup barrier

| Domain | Requirements |
|---|---|
| FND | CR-FND-130, CR-FND-140 |
| NET | CR-NET-190 |
| STO | CR-STO-080, CR-STO-090, CR-STO-100, CR-STO-130, CR-STO-140 |
| K8S | CR-K8S-030, CR-K8S-060, CR-K8S-130 |
| IAM | CR-IAM-150 |
| OBS | CR-OBS-030, CR-OBS-210 |
| OPS | CR-OPS-090, CR-OPS-100, CR-OPS-150 |
| DPL | CR-DPL-040, CR-DPL-070, CR-DPL-100, CR-DPL-140, CR-DPL-150, CR-DPL-200 |
| DAT | CR-DAT-020, CR-DAT-090 |

### SC-08 — Tenant exit and data portability

| Domain | Requirements |
|---|---|
| FND | CR-FND-070 |
| STO | CR-STO-090, CR-STO-120 |
| IAM | CR-IAM-010, CR-IAM-050, CR-IAM-150 |
| BIL | CR-BIL-100, CR-BIL-120 |
| CUX | CR-CUX-020, CR-CUX-100, CR-CUX-110 |
| FED | CR-FED-080, CR-FED-090 |
| DAT | CR-DAT-060 |

### SC-09 — Security incident with break-glass access

| Domain | Requirements |
|---|---|
| STO | CR-STO-140 |
| IAM | CR-IAM-080, CR-IAM-110, CR-IAM-150, CR-IAM-160, CR-IAM-190, CR-IAM-200 |
| OBS | CR-OBS-100 |
| OPS | CR-OPS-050, CR-OPS-060, CR-OPS-070, CR-OPS-120, CR-OPS-140, CR-OPS-170 |
| AGT | CR-AGT-040, CR-AGT-070, CR-AGT-170, CR-AGT-200 |

### SC-10 — Agent executes a risky change with an approval tuple

| Domain | Requirements |
|---|---|
| FND | CR-FND-140 |
| STO | CR-STO-090 |
| OPS | CR-OPS-020 |
| AGT | CR-AGT-010, CR-AGT-020, CR-AGT-030, CR-AGT-040, CR-AGT-050, CR-AGT-060, CR-AGT-080, CR-AGT-090, CR-AGT-100, CR-AGT-110, CR-AGT-120, CR-AGT-130, CR-AGT-140, CR-AGT-150, CR-AGT-160, CR-AGT-190, CR-AGT-200 |

## Coverage rule and P0 gaps

**Coverage rule.** Every P0 requirement MUST appear in at least one
end-to-end scenario. A P0 requirement that no scenario exercises cannot
demonstrate its acceptance evidence inside any user journey; it is a coverage
gap to close, either by extending an existing scenario or by adding a new one.
This rule is necessary but not sufficient: presence in a scenario is not
evidence of implementation (evidence lives outside this corpus), and P1/P2
requirements may remain uncovered without violating this rule.

**Result.** Of 180 P0 requirements, **118 (66%) are exercised by at least one
scenario** and **62 (34%) are coverage gaps to close**, computed from the
scenario coverage tables above against `registry/requirements.json`.

### Coverage gaps to close (62 P0 requirements, grouped by domain)

#### FND — Platform foundation (6)

| ID | Title |
|---|---|
| CR-FND-020 | Go-first platform runtime |
| CR-FND-050 | Public-core / private-workspace boundary |
| CR-FND-100 | English-only artifacts |
| CR-FND-110 | License and contribution clarity |
| CR-FND-150 | Real open baseline |
| CR-FND-170 | Gated public publication path |

#### NET — Network (6)

| ID | Title |
|---|---|
| CR-NET-020 | Subnets and IPAM contract |
| CR-NET-030 | Network-profile contract with replaceable SDN |
| CR-NET-060 | L4 network load balancing |
| CR-NET-100 | NAT and edge gateways per function |
| CR-NET-110 | Floating and public addresses |
| CR-NET-140 | Platform DNS/IPAM automation for provisioned resources |

#### STO — Storage, backup, DR (4)

| ID | Title |
|---|---|
| CR-STO-010 | Storage classes behind versioned contracts |
| CR-STO-020 | Media taxonomy and deterministic performance model |
| CR-STO-030 | Per-volume QoS with token-bucket limits and burst credits |
| CR-STO-050 | Encryption at rest with envelope keys |

#### K8S — Kubernetes and containers (7)

| ID | Title |
|---|---|
| CR-K8S-040 | Declarative cluster lifecycle provisioning |
| CR-K8S-070 | Container image registry and mirroring |
| CR-K8S-090 | GitOps-managed cluster add-ons |
| CR-K8S-100 | Admission policy with default-deny baselines |
| CR-K8S-110 | Ingress and Gateway API exposure contract |
| CR-K8S-120 | CNI networking contract |
| CR-K8S-150 | Kubernetes audit logging |

#### IAM — Identity and security (6)

| ID | Title |
|---|---|
| CR-IAM-020 | Fine-grained permission grammar |
| CR-IAM-040 | Subject model: users, service accounts, groups, pseudo-subjects |
| CR-IAM-060 | Central authorization API with batch decisions and list semantics |
| CR-IAM-070 | Per-service access-binding APIs proxied to the IAM core |
| CR-IAM-120 | Authorization availability: read-path split and bounded caching |
| CR-IAM-180 | Security scanning gates in CI/CD |

#### BIL — Billing and FinOps (2)

| ID | Title |
|---|---|
| CR-BIL-010 | Versioned usage-event schema |
| CR-BIL-150 | Public billing API: accounts, SKUs, operations |

#### CUX — Portal, UX, self-service (9)

| ID | Title |
|---|---|
| CR-CUX-030 | Information architecture: navigation, project home, create hub |
| CR-CUX-050 | Risk-classed confirmations |
| CR-CUX-060 | Structured, actionable error envelope |
| CR-CUX-070 | Honest availability and readiness states, no dead ends |
| CR-CUX-080 | Secrets shown as references only |
| CR-CUX-090 | User-generated content rendered inertly |
| CR-CUX-140 | Scope visibility and fail-closed management surfaces |
| CR-CUX-150 | Idempotent mutations with operation identity |
| CR-CUX-210 | Stateful honesty in console presentation |

#### OBS — Observability (9)

| ID | Title |
|---|---|
| CR-OBS-010 | Unified metrics platform |
| CR-OBS-020 | Single supported instrumentation library |
| CR-OBS-040 | Structured logs with correlation |
| CR-OBS-050 | Per-application log pipelines |
| CR-OBS-060 | Distributed tracing with declared retention |
| CR-OBS-070 | Multi-tenant metrics query API |
| CR-OBS-090 | Central alert routing configuration via GitOps |
| CR-OBS-160 | Monitoring of monitoring |
| CR-OBS-200 | Observability data lifecycle: retention, isolation, deletion |

#### OPS — Operations, SRE, support (4)

| ID | Title |
|---|---|
| CR-OPS-010 | Idempotent operations with plan/preview |
| CR-OPS-030 | Backpressure with explicit pressure states |
| CR-OPS-040 | Drain, recovery, and replay safety |
| CR-OPS-110 | Quota-versus-limit model with reserve/commit/release |

#### DPL — Deployment, IaC, CI/CD (1)

| ID | Title |
|---|---|
| CR-DPL-060 | Environment topology and naming convention |

#### FED — Federation and global portal (2)

| ID | Title |
|---|---|
| CR-FED-010 | Federation-ready base schemas (no single-provider assumptions) |
| CR-FED-020 | Federation honesty boundary |

#### DAT — Data services (6)

| ID | Title |
|---|---|
| CR-DAT-010 | Desired-state metadb with layered configuration and revision history |
| CR-DAT-030 | Explicit lifecycle state machine with error and offline paths |
| CR-DAT-040 | Provider-adapter isolation of platform primitives |
| CR-DAT-050 | Secrets brokering for engine credentials, TLS, and backup keys |
| CR-DAT-070 | PostgreSQL reference engine: automated HA and failover |
| CR-DAT-110 | Data services onboard through OCS connector packages (dogfooding) |

Domains with **no** P0 gaps: CMP, OCS, MKT, AGT — every P0 requirement in
these four domains is exercised by at least one scenario.

## Honesty notes

- Coverage here means *scenario mapping*, not implementation. All 330
  requirements are `proposed`; none carry linked delivery evidence in this
  corpus.
- The gap clusters are structural, not random: console/UX honesty (CUX),
  observability pipelines (OBS), data-service control plane (DAT), and the
  network product set beyond basic tenant networking (NET) have no journey
  that exercises them end-to-end. Closing them likely needs scenario
  extensions (for example a console-honesty walkthrough, an observability
  readiness journey, a managed-database lifecycle journey, a network-product
  journey) rather than new domains.
- Several gaps are cross-cutting rules that every scenario implicitly depends
  on but none explicitly checks (CR-FND-020 Go-first runtime, CR-FND-100
  English-only artifacts, CR-FND-150 real open baseline). Explicit scenario
  steps are still required by the coverage rule — implicit dependence is not
  coverage.
- Recompute this matrix whenever a scenario coverage table or the registry
  changes; drift between scenario tables and this file is a defect of the
  same class as registry drift.
