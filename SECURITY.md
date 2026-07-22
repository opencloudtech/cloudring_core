# Security Policy

CloudRING is a public multi-tenant cloud platform. Security work covers its
runtime, services, supply chain, contracts, and deployment interfaces while
keeping private deployment material out of the public repository.

CloudRING is in early development. No release is approved for real deployments.
The current `main` branch receives security fixes, but it must not be treated as
a supported production distribution. A future support and disclosure policy
will name supported releases, response targets, and deployment profiles before
any pilot approval or real-deployment support claim.

## Reporting

Report suspected vulnerabilities confidentially to `yuri@trukhin.com`.

Do not include secrets, tenant data, private endpoints, cookies, kubeconfigs, or
exploit material that is not needed to describe the issue. A good report
includes the affected component, expected impact, reproduction steps, and any
safe synthetic evidence.

## Platform ownership

CloudRING owns the public implementations and contracts for identity and IAM
decisions, policy evaluation, admission, tenant isolation, durable audit,
module readiness, backup/restore, rollback gates, source safety, and release
integrity.

## Service ownership

CloudRING maintainers are responsible for modules distributed by this project.
Independent module owners are responsible for their runtime security, data
handling, backup and restore behavior, support access, billing linkage, UI
extension behavior, and incident evidence. Platform validation must fail closed
when any module omits required security behavior or metadata.

## Enterprise and private boundary

Private identity configuration, deployment values, customer records, secrets,
and enterprise-only controls must stay outside this repository. Reusable
security controls, tests, and runbooks belong in public; examples must use
synthetic identifiers and portable language.

## Developer entry points

Security reviewers should inspect runtime authorization paths, tenant
boundaries, OCSv3 metadata, IAM and policy behavior, durability, source-safety
evidence, supply-chain provenance, and readiness evidence before accepting a
platform or module change.

Security-relevant examples and repository checks are scoped evidence only. They
do not prove the security of a live installation, provider adapter, independent
module, or downstream configuration.
