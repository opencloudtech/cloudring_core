# Package Ownership

CloudRING publishes the reusable platform runtime, contracts, validation and
policy surfaces, OCSv3 SDKs, open source service modules, reusable provider
adapters, tests, and developer/operator documentation. It does not publish
enterprise overlays, concrete installation values, customer data, credentials,
private modules, or live/generated evidence as runtime code.

This document is the current normative ownership map. A machine-readable map
must be added with the first runtime extraction slice and enforced in CI before
it can be used as a publication or boundary gate.

## Classifications

| Classification | Meaning |
| --- | --- |
| `public-core` | Reusable platform runtime, contracts, SDKs, validation tools, policy interfaces, tests, and public-safe documentation. |
| `service-module` | Service implementation behind OCSv3; publishable when contributed under Apache-2.0, otherwise independently owned. |
| `provider-adapter` | Provider or substrate implementation behind portable interfaces; reusable code may be public, installation-specific values remain downstream. |
| `enterprise-private` | Private downstream, customer, employer, or company-specific material outside the public boundary. |
| `evidence-only` | Requirements, receipts, plans, and generated evidence records. |
| `legacy/migration-debt` | Mixed or transitional material that requires splitting before publication. |

## Required Public-Core Surfaces

The extraction backlog classifies these paths as publishable `public-core`:

| Path | Public role |
| --- | --- |
| `pkg/ocs` | OCSv3 connector package and validator contracts. |
| `pkg/backup/restoreproof` | Provider-neutral restore proof model, canonical digests, and fail-closed validator. |
| `pkg/backup/velero118` | Exact Velero 1.18.1 CSI data-mover decoders, archive reader, collectors, and adapter execution boundary. |
| `pkg/kubeconfigpipe` | Bounded in-memory replay of a brokered pipe-backed kubeconfig plus process-tree-contained execution for multi-query child processes. |
| `pkg/secureexec` | Content-pinned executable identity, bounded output, process-tree cleanup, and optional kubeconfig replay for downstream collectors. |
| `cmd/cloudring-backup` | Read-only baseline, collection, and offline verification workflow. |
| `internal/iam` | IAM decision and policy interface candidate. |
| `internal/migration` | Go and upstream Kubernetes runtime policy guard candidate. |
| `contracts/release/release-bom-contract.json` | Release BOM machine contract candidate. |
| `docs/what-is-ocsv3.md` | External service-team SDK documentation entry point. |

The backlog also includes `cmd/ocsctl`, `internal/releasebom`,
`internal/docscheck`, `internal/sourcecheck`, and reusable provider adapter
code. Paths not yet present in this repository are targets, not evidence that
the extraction has already happened.

## Provider Boundary Guard

The downstream `internal/ovhinstall` path combines reusable installation behavior
with one concrete installation's assumptions. It remains downstream until it
is split. Generic profile, preflight, deployment, and verification engines
should move to public packages; provider account, topology, endpoint, and live
evidence inputs must remain downstream.

The guard is intentionally fail closed while that split is incomplete. It must
not be used to keep reusable provider support private after source-safe generic
code has been separated from installation values.

## Migration Notes

The backup proof paths above are the first runtime extraction slice. Other
listed candidate paths remain classifications until code and tests are present.

Service paths such as backup, IaaS, message delivery, marketplace,
observability, resilience, and accelerator/HPC surfaces should become complete
OCSv3 modules with controller/API, portal extension, billing, support, evidence,
durability, lifecycle, rollback, delete/export, backup/restore, denied,
degraded, and retry behavior. Current project-owned implementations and tests
belong in public once source-safe; third-party modules remain independently
owned unless contributed.

Mixed paths such as `cmd`, `docs`, `internal`, `portal`, and `scripts` must be
split before publication. Generic runtime, service, installer, operator,
contract, policy, and validation pieces move into CloudRING. Concrete
installation values and private product extensions stay downstream.

Evidence paths remain evidence. A passing ownership manifest does not claim
production readiness, deployment readiness, tenant data durability, backup
coverage, or single-point-of-failure readiness.
