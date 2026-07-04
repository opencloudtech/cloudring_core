# CloudRING Enterprise Boundary

This document defines how downstream companies, downstream/internal enterprise
repos, and user-owned services consume or sync CloudRING while keeping
enterprise modules, provider adapters, and customer deployments separate. It is
project policy context, not legal advice.

## License Boundary

| Material | Default classification | May be open-sourced in CloudRING? |
| --- | --- | --- |
| Reusable contracts, SDK surfaces, validators, public docs, synthetic fixtures, and module metadata schemas under `cloudring_core/` | Apache-2.0 core | Yes, when it contains no private source text, deployment values, credentials, tenant data, or customer-specific assumptions. |
| Material outside `cloudring_core/` in this repository | outer repository license | Not by default. It must be reviewed, copied or moved intentionally, and made compatible with the public-core boundary before publication. |
| Enterprise/private IP, commercial packaging, proprietary support workflows, closed integrations, private policy packs, and company-specific automation | private or separately licensed enterprise work | No, unless the owner intentionally contributes a separable public-core part under Apache-2.0. |
| Concrete provider adapters, live infrastructure integration, account-specific provisioning logic, and deployment overlays | private adapter or downstream implementation | No by default. Public core may define adapter interfaces and synthetic examples only. |
| Customer deployments, customer topology, live evidence, support records, billing records, and operational runbooks for a real tenant or environment | customer-controlled or deployment-specific material | No. Keep it in the customer or operator repository and reference only public interfaces from core. |
| User-owned modules and service implementations built against OCSv3 contracts | owned by the module author or their organization | Not automatically. The module owner decides whether to publish it and under which compatible terms. |

## Sync Model

Downstream companies may mirror, vendor, subtree-sync, or package the
Apache-2.0 core for internal use. The sync direction should preserve a clean
boundary:

- Public-to-private sync may import `cloudring_core/` into internal repos.
- Private-to-public sync must contribute only generic core contracts, docs,
  tests, SDK surfaces, or synthetic fixtures.
- Enterprise modules, provider adapters, customer deployments, and private
  overlays remain in downstream repos unless explicitly extracted into a
  public, separable core contribution.

Downstream/internal enterprise repos follow the same rule: they may consume
the Apache-2.0 core, but internal enterprise/private IP does not become
Apache-2.0 merely because it integrates with or syncs from CloudRING.

## Employer/Customer Separation

Employer/customer separation is mandatory. Work created for an employer,
customer, partner, or deployment operator must not be copied into
`cloudring_core/` unless the contributor has authority to submit it and the
material is safe for public Apache-2.0 distribution.

Do not place customer names, private endpoints, real account identifiers,
tenant data, support cases, billing records, live topology, or copied private
source text in public core. Public examples must use synthetic identifiers and
interface-level behavior.

## Contribution Rules

Contributions to `cloudring_core/` are accepted only when they fit the
Apache-2.0 core boundary:

- The contribution is a reusable core contract, validator, SDK surface,
  documentation page, test, or synthetic fixture.
- The contributor has the right to submit the material for public Apache-2.0
  distribution.
- The contribution does not include enterprise/private IP, provider-specific
  implementation details, customer deployment material, secrets, private
  endpoints, tenant data, or copied private source text.
- The contribution preserves service ownership by depending on OCSv3 metadata
  and module interfaces instead of importing a service implementation.
- The contribution preserves adapter ownership by defining provider interfaces
  rather than publishing concrete private provider adapters.

If a change needs enterprise behavior, keep that behavior downstream and add
only the generic public interface or validation contract to core.

## Fresh-Reader Classification

A fresh reader should classify a proposed artifact this way:

| Question | Core answer |
| --- | --- |
| Is it reusable by multiple companies without private deployment facts? | Candidate for Apache-2.0 core. |
| Does it implement a customer deployment, proprietary module, commercial workflow, or concrete provider integration? | Keep it enterprise/private or user-owned. |
| Does it contain live evidence, tenant records, credentials, private endpoints, or employer/customer source text? | Do not publish in core. |
| Is it a service built by a user or downstream company against OCSv3? | It is a user-owned module unless the owner intentionally contributes it. |
| Is it only an interface, schema, validator, or synthetic example needed by service or adapter authors? | Candidate for public core after source-safety review. |

## Non-Claims

This boundary does not relicense the outer repository, private downstream
repositories, enterprise modules, provider adapters, customer deployments, or
user-owned modules. It also does not grant trademark rights, deployment
readiness, customer approval, or authority to submit employer/customer work.
