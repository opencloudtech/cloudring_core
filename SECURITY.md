# Security Policy

CloudRING is a public contract surface. Security work focuses on keeping
the core portable, tenant-safe, and free of private deployment material.

## Reporting

Report suspected vulnerabilities confidentially to `yuri@trukhin.com`.

Do not include secrets, tenant data, private endpoints, cookies, kubeconfigs, or
exploit material that is not needed to describe the issue. A good report
includes the affected component, expected impact, reproduction steps, and any
safe synthetic evidence.

## Platform ownership

CloudRING owns security contracts for IAM decisions, policy evaluation,
admission behavior, audit evidence, module readiness, rollback gates, and
source-safety checks.

## Service ownership

Service module owners are responsible for their runtime security, data handling,
backup and restore behavior, support access, billing linkage, UI extension
behavior, and incident evidence. CloudRING validation should fail closed when a
module omits required security metadata.

## Enterprise and private boundary

Private identity configuration, deployment values, customer records, secrets,
and enterprise-only controls must stay outside this subtree. Public examples
must use synthetic identifiers and portable contract language.

## Developer entry points

Security reviewers should inspect OCSv3 package metadata, IAM and policy
contracts, source-safety evidence, and readiness evidence before accepting a
core change or module publication.
