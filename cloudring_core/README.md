# CloudRING Core

CloudRING Core is an Apache-2.0 open source project for reusable cloud-service
orchestration contracts. It gives service teams, platform operators, and
contributors a portable foundation for building cloud modules around OCSv3
(Open Cloud Standard 3).

## Core ownership

The public core owns shared contracts and validation surfaces:

- OCSv3 registry and package validation.
- IAM, policy, and admission contracts.
- GitOps and bootstrap abstractions.
- Evidence, readiness, module lifecycle, BOM, and rollback gates.
- Portal shell and extension contracts.
- Provider adapter interfaces.
- Developer SDK documentation.

Core does not own service implementation code, deployment-specific values,
customer data, credentials, or live infrastructure endpoints.

## Service ownership

Each cloud service is owned by an independent module team. A module publishes
OCSv3 metadata for its API/controller, portal extension, billing meters, support
diagnostics, evidence, durability, lifecycle, rollback, delete/export, backup,
restore, denied, degraded, and retry contracts. The core consumes that metadata
instead of importing a service implementation.

## Developer entry points

- Read `docs/public-boundary.md` before adding public-core material.
- Follow `docs/developer-guide.md` for the external developer path.
- Use `docs/conformance.md` to validate OCSv3 module packages.
- Use `cloudring_core/examples/synthetic-service-module/connector-package.json`
  as the smallest connector-package example.
- Use `cloudring_core/reference/synthetic-service/module-package.json` as the
  reference module for conformance.
- Use provider adapter interfaces for infrastructure integration.
- Use portal shell contracts for user-interface extensions.
- Use evidence and readiness contracts before making operational claims.

## Fresh clone validation

Prerequisites: Git and a supported Go toolchain. From a fresh public clone:

```bash
gh repo clone opencloudtech/cloudring_core
cd cloudring_core
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

The example package is synthetic. It is intended for local validation and does
not include live provider values, tenant records, credentials, or deployment
receipts.

## Contribution path

Contributions should arrive through a pull request against the public
repository. Before opening a PR, run the fresh-clone validation commands above,
read `CONTRIBUTING.md`, sign the required CLA/DCO attestations, and keep the
change limited to public-safe contracts, docs, examples, SDK code, or OCSv3
validation behavior. Contributions must satisfy source-safety review before
they can be considered for merge.

Do not include local host paths, tenant/customer records, live provider
endpoints, tokens, cookies, kubeconfigs, credentials, or deployment evidence in
code, docs, examples, tests, commits, or pull request text.

## Non-claims

These documents establish project metadata and public validation paths. They do
not claim live production readiness for any particular deployment.
