# CloudRING Core Developer Guide

CloudRING Core exposes OCSv3 (Open Cloud Standard 3) contracts for service
modules, billing connectors, portal UI extensions, and reviewable evidence. A
service module is accepted by metadata and API contracts, not by importing
platform internals or depending on a provider-specific implementation.

Use this guide with the public docs in this repository:

- `cloudring_core/docs/service-module-authoring.md`
- `cloudring_core/docs/billing-connector-authoring.md`
- `cloudring_core/docs/ui-extension-authoring.md`
- `cloudring_core/docs/evidence-authoring.md`
- `cloudring_core/docs/what-is-ocsv3.md`
- `cloudring_core/docs/module-authoring.md`
- `cloudring_core/docs/publishing-a-module.md`
- `cloudring_core/docs/public-boundary.md`

## Fresh Clone Path

External developers should be able to validate the public tree without access
to any private repository or live provider account:

```bash
gh repo clone opencloudtech/cloudring_core
cd cloudring_core
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

Use `./evidence/` or another local scratch directory for optional generated
receipts. Do not commit generated receipts unless they are synthetic,
source-safe, and intentionally part of a public example.

## Service Team Role Path

The service team role path starts with an OCSv3 connector package and ends with
a package that validates without importing platform internals.

1. Copy the synthetic package shape from
   `cloudring_core/examples/synthetic-service-module/connector-package.json`.
2. Replace names, capability classes, routes, meters, policy refs, support refs,
   and evidence refs with module-owned values.
3. Keep implementation code in the service repository. The CloudRING package
   declares capabilities, lifecycle actions, API refs, tenant permissions,
   microfrontend mount refs, support diagnostics, and durability evidence.
4. Validate the package:

```bash
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
```

Expected output:

```text
ocs_connector_package_valid cloudring_core/examples/synthetic-service-module/connector-package.json
```

Run conformance against a module package when the package includes the reference
module surfaces:

```bash
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

## Platform Operator Role Path

The platform operator role path consumes validated OCSv3 metadata. Operators
review connector package status, catalog visibility, idempotent lifecycle
actions, tenant access, support owner, readiness refs, and durability refs
before admitting a module to a registry or marketplace flow.

Operators must not wire service-specific controllers, database schemas,
frontend bundles, or billing logic directly into CloudRING Core. The platform
reads module metadata and invokes documented API or automation refs owned by the
module.

## Security Reviewer Role Path

The security reviewer role path checks that the package is source-safe and
tenant-safe:

- no secrets, tokens, private endpoints, tenant data, or copied source text;
- no internal package imports or implementation refs;
- secret fields are references only;
- tenant permissions are explicit and scoped;
- support diagnostics have a redaction boundary;
- evidence bundles name non-claims and review paths.

Run the public validation path before publishing docs or examples:

```bash
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

Repository maintainers also run source-safety and publication-boundary checks
before merging public-core changes.

## Pull Request Path

1. Create a topic branch from the public repository.
2. Keep the change scoped to public contracts, docs, examples, SDK code, or
   OCSv3 validation behavior.
3. Run the fresh-clone validation commands from this guide.
4. Open a pull request with a concise summary and public-safe validation notes.
5. Complete the CLA/DCO checks and address source-safety findings.

Never paste local machine paths, private repository paths, tenant/customer
records, live provider endpoints, tokens, cookies, kubeconfigs, credentials, or
private deployment evidence into the public repository or pull request text.

## Enterprise/Downstream Maintainer Role Path

The enterprise/downstream maintainer role path extends CloudRING Core by
keeping private adapters, commercial packaging, customer-specific policy, and
deployment material outside the public core. Downstream maintainers can consume
the OCSv3 contracts and publish their own module packages, but they own their
service implementation, extension delivery, billing settlement, support process,
and release evidence.

This guide does not claim production readiness. It documents the public authoring
path and the local validation commands required before a module can be reviewed.
