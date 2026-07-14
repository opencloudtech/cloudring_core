# CloudRING Public Boundary

This document defines what belongs in the Apache-2.0 CloudRING subtree and
what must remain in independently owned modules or private downstream work.

## Platform ownership

CloudRING public material includes or may include:

- The provider-neutral Go control plane and APIs.
- OCSv3 schemas, validators, SDKs, conformance tests, and reference modules.
- IAM, identity, policy, admission, durable audit, catalog, billing, portal,
  and self-service implementations.
- Generic installers, GitOps bases, observability, backup/restore, lifecycle,
  upgrade, rollback, readiness, and operations tooling.
- Open source service modules and reusable provider adapters that satisfy the
  project's contribution and source-safety requirements.
- Developer and operator entry points for service teams, adapter teams,
  security reviewers, providers, and downstream maintainers.

CloudRING public material must be reusable without a particular company's
accounts, topology, private endpoints, customer data, or proprietary modules.

## Service ownership

CloudRING may distribute complete service modules under Apache-2.0. A module
developed outside the project remains owned and licensed by its author unless
it is contributed. Every module publishes portable metadata for its
API/controller behavior, UI extension points, billing meters, support
diagnostics, data durability, lifecycle actions, rollback, delete/export,
backup/restore, and denied/degraded/retry states. The platform orchestrates the
declared interfaces without importing a module's private internals.

## Enterprise and private boundary

The following material stays outside CloudRING public:

- Credentials, tokens, private keys, session data, and secret references tied to
  a real deployment.
- Tenant, customer, billing, support, or operational records.
- Private endpoints, hostnames, infrastructure inventory, and
  deployment-specific values.
- Company-only modules, enterprise overlays, and proprietary implementation
  details.
- Account-, customer-, topology-, and installation-specific provider values.
- Copied private-source text or dependencies on private implementation
  details.

Public examples should use synthetic identifiers and capability names. Generic
operator runbooks belong in public; a downstream deployment's private values,
inventory, incidents, and live evidence do not.

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

Backup operators start with `docs/restore-proof-collector.md`. Generic typed
collection and validation belong in public core. A deployment's BackupContents
archive, source baseline, unsigned or signed receipts, provider credentials,
adapter configuration, and live evidence stay downstream.

## Non-claims

The public boundary does not claim that any private deployment is ready, that
the current extraction is complete, that every third-party module is open
source, or that material outside this repository has the same license.
Readiness must be proven by scoped evidence.
