# CloudRING

CloudRING is an Apache-2.0 open source cloud platform. It gives operators a
portable control plane for building a cloud provider and gives independent
service teams stable integration surfaces through OCSv3 (Open Cloud Standard
3).

## Platform ownership

The public project owns the reusable platform, including:

- OCSv3 contracts, SDKs, validators, and conformance tooling.
- Identity, IAM, policy, admission, and durable audit services.
- Catalog, billing/FinOps, portal, and self-service lifecycle surfaces.
- Provider-neutral control-plane runtimes and provider adapter interfaces.
- Installers, GitOps bases, observability, backup/restore, readiness, upgrade,
  rollback, and operations tooling.
- Open source service modules and reusable provider adapters accepted by the
  project.

Deployment-specific values, credentials, customer data, live evidence, and
company-only modules remain in downstream repositories. A downstream product
must configure and extend the public platform rather than maintain a duplicate
fork of generic core behavior.

## Service ownership

CloudRING may ship service implementations under Apache-2.0. Independently
developed modules remain owned and licensed by their authors unless they are
contributed to this repository. In either case, a module integrates through
OCSv3 metadata and APIs so the platform does not depend on its implementation
language, deployment substrate, or release cadence.

## Repository state

This directory currently contains the first extraction slice. Contracts and
conformance are available first; runtime, IAM, billing, portal, installer,
operations, and service slices are being moved with their tests. Green checks
prove only the surfaces present in a commit and must not be interpreted as a
claim that the full platform extraction or a production installation is
complete.

## Developer entry points

- Read `docs/public-boundary.md` before adding CloudRING public-safe material.
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

Prerequisites: Git and the latest security patch of Go 1.25 or 1.26. From a
fresh public clone:

```bash
gh repo clone opencloudtech/CloudRING
cd CloudRING
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
change public-safe and reusable. Platform runtime, tests, documentation,
service modules, adapters, and OCSv3 changes are all valid contributions when
they respect the public/downstream boundary. Contributions must satisfy
source-safety review before they can be considered for merge.

Do not include local host paths, tenant/customer records, live provider
endpoints, tokens, cookies, kubeconfigs, credentials, or deployment evidence in
code, docs, examples, tests, commits, or pull request text.

## Non-claims

These documents establish project metadata and public validation paths. They do
not claim live production readiness for any particular deployment.
