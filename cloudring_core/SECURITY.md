# Security Policy

CloudRING is a public multi-tenant cloud platform. Security work covers its
runtime, services, supply chain, contracts, and deployment interfaces while
keeping private deployment material out of the public repository.

## Reporting

Report suspected vulnerabilities through the project maintainer channel used by
the receiving repository. Do not include secrets, tenant data, private
endpoints, or exploit material that is not needed to describe the issue.

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
