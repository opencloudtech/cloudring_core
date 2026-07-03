# CloudRING Core Public Boundary

This document defines what belongs in the Apache-2.0 CloudRING Core subtree and
what must remain in independently owned modules or private downstream work.

## Core ownership

Public core material may include:

- OCSv3 registry schemas, validators, and SDK documentation.
- IAM, policy, admission, audit, and evidence contracts.
- GitOps and bootstrap abstractions that are implementation-neutral.
- Module lifecycle, BOM compatibility, rollback, readiness, and non-claim
  contracts.
- Portal shell slots and extension metadata.
- Provider adapter interfaces expressed as portable capabilities.
- Developer entry points for service teams, adapter teams, security reviewers,
  and downstream maintainers.

Public core material must be reusable without private infrastructure values or a
single service implementation.

## Service ownership

Service modules own runtime code and service-specific behavior. A module
publishes portable metadata for API/controller behavior, UI extension points,
billing meters, support diagnostics, data durability, lifecycle actions,
rollback, delete/export, backup/restore, denied, degraded, and retry states.
The core validates those declarations and orchestrates against contracts.

## Enterprise and private boundary

The following material stays outside public core:

- Credentials, tokens, private keys, session data, and secret references tied to
  a real deployment.
- Tenant, customer, billing, support, or operational records.
- Private endpoints, hostnames, infrastructure inventory, and
  deployment-specific values.
- Company-only modules, enterprise overlays, and proprietary implementation
  details.
- Copied private-source text or implementation-specific dependencies.

Public examples should use synthetic identifiers and capability names. Public
docs may describe an interface but must not include a downstream deployment's
private values or operational runbook.

## Source-safety checklist

Before publishing or opening a pull request, confirm the change contains none
of the following:

- local host paths or private repository paths;
- tenant, customer, billing, support, or operational records;
- live provider endpoints, infrastructure inventories, or deployment values;
- tokens, cookies, credentials, kubeconfigs, private keys, or secret material;
- copied proprietary source text or company-only implementation details.

Run the public validation commands in `README.md` and keep any optional evidence
synthetic and source-safe.

## Developer entry points

Service teams start with OCSv3 module metadata. Adapter teams start with
provider adapter interfaces. Portal teams start with shell extension contracts.
Security reviewers start with IAM, policy, source-safety, and evidence
contracts. Release reviewers start with BOM, readiness, rollback, and non-claim
contracts.

## Non-claims

The public boundary does not claim that any private deployment is ready, that
all future modules have been extracted, or that material outside this subtree
has the same license. Readiness must be proven by scoped evidence.
