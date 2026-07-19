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

The current pull request series is the first extraction from the existing
platform into its canonical public repository. Contracts, conformance,
backup-proof, IAM/identity, and PostgreSQL-backed transactional-state runtime
slices are available; billing, portal,
installer, operations, and service slices are being moved with their tests.
Green checks prove only the
surfaces present in a commit and must not be interpreted as a claim that the
full platform extraction or a production installation is complete.

## Developer entry points

- Read `docs/public-boundary.md` before adding public material.
- Follow `docs/developer-guide.md` for the developer path.
- Use `docs/conformance.md` to validate OCSv3 module packages.
- Use `docs/http-transport-security.md` for the provider-neutral public HTTP
  transport and response-header audit.
- Use `docs/iam-policy.md` for the importable IAM, OIDC/JWKS/JWT, secure-cookie,
  CSRF, and management-gate runtime boundary.
- Use `examples/synthetic-service-module/connector-package.json`
  as the smallest connector-package example.
- Use `reference/synthetic-service/module-package.json` as the
  reference module for conformance.
- Use provider adapter interfaces for infrastructure integration.
- Use portal shell contracts for user-interface extensions.
- Use evidence and readiness contracts before making operational claims.
- Use `docs/restore-proof-collector.md` for the public Velero 1.18.2 CSI
  data-mover proof workflow and its explicit non-claims.
- Use `docs/one-server-loss-drill.md` for the provider-neutral, read-only
  continuity observer and exact fault/recovery evidence boundary.
- Use `docs/provider-site-installation.md` for the strict provider-neutral site
  inventory preflight and deterministic installation plan contract.
- Use `docs/kubeadm-control-plane-ha.md` for the strict standalone kubeadm
  configuration renderer and captured-state HA verifier.
- Use `deploy/kubernetes/postgresql-ha/README.md` for the pinned CloudNativePG
  HA source profile, migration boundary, and required downstream live gates.

## Fresh clone validation

Prerequisites: Git and the latest security patch of Go 1.25 or 1.26. From a
fresh public clone:

```bash
gh repo clone opencloudtech/CloudRING
cd CloudRING
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
go run ./cmd/cloudring-registry validate ./contracts/module-registry/fixtures/synthetic-module-registry.json
go run ./cmd/cloudring-site render-kubeadm --spec ./examples/kubeadm-bootstrap-spec.json
go run ./cmd/cloudring-site verify-kubeadm \
  --inventory /protected/run/kubeadm-stand-inventory.json \
  --one-server-loss-receipt /protected/run/unsigned-one-server-loss-receipt.json
go run ./cmd/cloudring-manifestcheck --root .
```

The example inputs are synthetic. They are intended for local validation and
do not include live provider values, tenant records, credentials, or deployment
receipts. A ready result for a synthetic stand inventory is not live readiness
evidence.

## HTTP security audit

`cloudring-httpcheck` checks a declared browser or API surface without
following redirects, reading response bodies, or copying URLs and response
values into its report. Canary and steady modes enforce different redirect and
HSTS promotion rules. See `docs/http-transport-security.md` for the exact
contract, command line, and exit codes. Deployment-specific targets and live
evidence stay in downstream provider repositories.

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
