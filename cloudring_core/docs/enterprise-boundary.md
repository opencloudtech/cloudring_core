# CloudRING Enterprise Boundary

This document defines how downstream company and enterprise repositories
consume and contribute to CloudRING while keeping proprietary modules,
installation values, and customer deployments separate. The public repository
is the canonical upstream for reusable platform code.

## License Boundary

| Material | Default classification | May be open-sourced in CloudRING? |
| --- | --- | --- |
| Reusable platform runtime, services, contracts, SDKs, validators, tests, installers, generic manifests, and docs in this repository | Apache-2.0 CloudRING | Already public when accepted through SafePush. |
| Enterprise/private IP, commercial packaging, proprietary support workflows, closed integrations, private policy packs, and company-specific automation | private or separately licensed enterprise work | No, unless the owner intentionally contributes a separable public-core part under Apache-2.0. |
| Reusable provider adapters and provider-neutral provisioning logic | Apache-2.0 candidate | Yes, after source-safety and conformance review. |
| Account-, topology-, endpoint-, credential-, or installation-specific provider configuration and overlays | downstream deployment material | No. Keep concrete values and live evidence downstream. |
| Customer deployments, customer topology, live evidence, support records, billing records, and operational runbooks for a real tenant or environment | customer-controlled or deployment-specific material | No. Keep it in the customer or operator repository and reference only public interfaces from core. |
| User-owned modules and service implementations built against OCSv3 contracts | owned by the module author or their organization | Not automatically. The module owner decides whether to publish it and under which compatible terms. |

## Sync Model

Downstream products should include the public repository as a pinned Git
submodule named `cloudring_core` and keep their own product/deployment material
beside it. The sync direction preserves a clean boundary:

- Public-to-downstream updates advance the submodule pointer after upstream
  SafePush passes.
- Downstream changes to generic platform behavior are first proposed to the
  public repository with code, tests, and documentation; downstream then bumps
  the accepted public commit.
- Proprietary modules, company integrations, customer deployments, private
  values, and live evidence remain beside the submodule.
- Downstream repositories must not keep copied implementations of generic
  public packages. Transitional duplicates are removed in the same migration
  slice after upstream acceptance.

Downstream/internal enterprise repos follow the same rule: they may consume
the Apache-2.0 core, but internal enterprise/private IP does not become
Apache-2.0 merely because it integrates with or syncs from CloudRING.

## Employer/Customer Separation

Employer/customer separation is mandatory. Work created for an employer,
customer, partner, or deployment operator must not be submitted to this
repository unless the contributor has authority to submit it and the material
is safe for public Apache-2.0 distribution.

Do not place customer names, private endpoints, real account identifiers,
tenant data, support cases, billing records, live topology, or copied private
source text in CloudRING public. Public examples must use synthetic identifiers and
interface-level behavior.

## Contribution Rules

Contributions are accepted only when they fit the Apache-2.0 public boundary:

- The contribution is reusable platform runtime, a service or adapter module,
  a contract, validator, SDK surface, installer, generic manifest,
  documentation page, test, or synthetic fixture.
- The contributor has the right to submit the material for public Apache-2.0
  distribution.
- The contribution does not include enterprise/private IP, provider-specific
  implementation details, customer deployment material, secrets, private
  endpoints, tenant data, or copied private source text.
- The contribution preserves module independence by integrating through OCSv3
  metadata and APIs instead of importing another module's private internals.
- Provider code separates reusable adapter behavior from concrete accounts,
  topology, credentials, endpoints, and deployment evidence.

If a change needs enterprise behavior, keep that behavior downstream and add
only the generic public interface or validation contract to core.

## Fresh-Reader Classification

A fresh reader should classify a proposed artifact this way:

| Question | Core answer |
| --- | --- |
| Is it reusable by multiple companies without private deployment facts? | Candidate for Apache-2.0 core. |
| Does it implement a customer deployment, proprietary module, commercial workflow, or installation-specific provider configuration? | Keep it enterprise/private or user-owned. |
| Does it contain live evidence, tenant records, credentials, private endpoints, or employer/customer source text? | Do not publish in core. |
| Is it a service built by a user or downstream company against OCSv3? | It is a user-owned module unless the owner intentionally contributes it. |
| Is it reusable platform/runtime/service/adapter code, or an interface, schema, validator, test, installer, runbook, or synthetic example? | Candidate for CloudRING public after source-safety and license review. |

## Non-Claims

This boundary does not relicense the outer repository, private downstream
repositories, enterprise modules, provider adapters, customer deployments, or
user-owned modules. It does not grant trademark rights, deployment readiness,
customer approval, or authority to submit employer/customer work. It is project
policy, not legal advice.
