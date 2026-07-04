# Package Ownership

CloudRING Core publishes reusable platform contracts, validation surfaces, policy
interfaces, OCSv3 SDK surfaces, release evidence contracts, and public developer
documentation. It does not publish concrete service implementations, concrete
provider implementations, enterprise overlays, or generated evidence as runtime
code.

The machine-readable ownership map is
`cloudring_core/contracts/package-ownership.json`.

## Classifications

| Classification | Meaning |
| --- | --- |
| `public-core` | Reusable contracts, SDK facades, validation tools, policy interfaces, and public-safe documentation. |
| `service-module` | Service implementation ownership that must migrate behind OCSv3 module packages. |
| `provider-adapter` | Concrete provider or substrate implementation that stays behind adapter interfaces. |
| `enterprise-private` | Private downstream, customer, employer, or company-specific material outside the public boundary. |
| `evidence-only` | Requirements, receipts, plans, and generated evidence records. |
| `legacy/migration-debt` | Mixed or transitional material that requires splitting before publication. |

## Required Public-Core Surfaces

The ownership verifier requires these paths to be classified as publishable
`public-core`:

| Path | Public role |
| --- | --- |
| `pkg/ocs` | OCSv3 connector package and validator contracts. |
| `internal/iam` | IAM decision and policy interface candidate. |
| `internal/migration` | Go and upstream Kubernetes runtime policy guard candidate. |
| `cloudring_core/contracts/release/release-bom-contract.json` | Release BOM machine contract candidate. |
| `cloudring_core/docs/what-is-ocsv3.md` | External service-team SDK documentation entry point. |

The manifest also marks `cmd/ocsctl`, `internal/releasebom`,
`internal/docscheck`, `internal/sourcecheck`, and selected provider adapter
contract documents as public-core candidates. Those paths still need extraction
or facade work before publication when they currently live under `internal/`.

## Provider Implementation Guard

`internal/ovhinstall` is classified as `provider-adapter` with
`publishableInCore:false`. It is concrete provider implementation code, so it is
not a public-core package. The runtime-gate verifier fails if this path is mapped
to `public-core` or if `publishableInCore` is set to true.

This guard is intentionally path-based and machine-checked. It prevents a future
manifest edit from accidentally making concrete provider implementation code
part of the publishable public core.

## Migration Notes

This map is a classification contract only. It does not move code.

Service-owned paths such as backup, IaaS, message delivery, marketplace,
observability, resilience, and accelerator/HPC surfaces must become OCSv3 module
packages with controller/API, portal extension, billing, support, evidence,
durability, lifecycle, rollback, delete/export, backup/restore, denied,
degraded, and retry contracts before core consumes them as modules.

Mixed paths such as `cmd`, `docs`, `internal`, `portal`, and `scripts` must be
split before any public publication. Only the generic contract, shell, policy,
and validation pieces can move into CloudRING Core.

Evidence paths remain evidence. A passing ownership manifest does not claim
production readiness, deployment readiness, tenant data durability, backup
coverage, or single-point-of-failure readiness.
