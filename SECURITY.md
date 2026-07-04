# Security Policy

CloudRING is a public contract surface. Security work focuses on keeping
the core portable, tenant-safe, and free of private deployment material.

## Reporting

Report suspected vulnerabilities through the project maintainer channel used by
the receiving repository. Do not include secrets, tenant data, private
endpoints, or exploit material that is not needed to describe the issue.

## Core ownership

The public core owns security contracts for IAM decisions, policy evaluation,
admission behavior, audit evidence, module readiness, rollback gates, and
source-safety checks.

## Service ownership

Service module owners are responsible for their runtime security, data handling,
backup and restore behavior, support access, billing linkage, UI extension
behavior, and incident evidence. Core validation should fail closed when a
module omits required security metadata.

## Enterprise and private boundary

Private identity configuration, deployment values, customer records, secrets,
and enterprise-only controls must stay outside this subtree. Public examples
must use synthetic identifiers and portable contract language.

## Developer entry points

Security reviewers should inspect OCSv3 package metadata, IAM and policy
contracts, source-safety evidence, and readiness evidence before accepting a
core change or module publication.
