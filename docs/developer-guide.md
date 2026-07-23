# CloudRING Developer Guide

CloudRING exposes OCSv3 (Open Cloud Standard 3) contracts for service
modules, billing connectors, portal UI extensions, and reviewable evidence. A
service module is accepted by metadata and API contracts, not by importing
platform internals or depending on a provider-specific implementation.

Use this guide with the public docs in this repository:

- `docs/service-module-authoring.md`
- `docs/billing-connector-authoring.md`
- `docs/ui-extension-authoring.md`
- `docs/evidence-authoring.md`
- `docs/what-is-ocsv3.md`
- `docs/module-authoring.md`
- `docs/publishing-a-module.md`
- `docs/public-boundary.md`

## Fresh Clone Path

External developers should be able to validate the public tree without access
to any private repository or live provider account:

```bash
gh repo clone opencloudtech/CloudRING
cd CloudRING
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./examples/synthetic-service-module/connector-package.json
```

Use `./evidence/` or another local scratch directory for optional generated
receipts. Do not commit generated receipts unless they are synthetic,
source-safe, and intentionally part of a public example.

## Service Team Role Path

The service team role path starts with an OCSv3 connector package and ends with
a package that validates without importing platform internals.

1. Copy the synthetic package shape from
   `examples/synthetic-service-module/connector-package.json`.
2. Replace names, capability classes, routes, meters, policy refs, support refs,
   and evidence refs with module-owned values.
3. Keep implementation code in the service repository. The CloudRING package
   declares capabilities, lifecycle actions, API refs, tenant permissions,
   microfrontend mount refs, support diagnostics, and durability evidence.
4. Validate the package and run conformance against that same artifact:

```bash
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./examples/synthetic-service-module/connector-package.json
```

Expected output:

```text
ocs_connector_package_valid input-sha256:<digest>
```

`ocsctl` intentionally identifies every selected input with a stable opaque
digest. It does not emit the selected path or the package-declared name to
stdout or conformance evidence. Re-running the command with the same selected
file and unchanged content produces the same identity.

The reference module is a larger example. Validate and check conformance on that
same package as well:

```bash
go run ./cmd/ocsctl validate ./reference/synthetic-service/module-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
```

## Platform Operator Role Path

The platform operator role path consumes validated OCSv3 metadata. Operators
review connector package status, catalog visibility, idempotent lifecycle
actions, tenant access, support owner, readiness refs, and durability refs
before admitting a module to a registry or marketplace flow.

Operators must not wire service-specific controllers, database schemas,
frontend bundles, or billing logic directly into CloudRING. The platform
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
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./examples/synthetic-service-module/connector-package.json
```

Repository maintainers also run source-safety and publication-boundary checks
before merging CloudRING public changes.

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

The enterprise/downstream maintainer role path extends CloudRING by keeping
commercial packaging, proprietary modules, customer-specific policy, concrete
installation values, and live evidence outside CloudRING. Downstream
maintainers consume the pinned public core, contribute reusable platform
changes upstream, and may publish or privately operate their own OCSv3 modules.
Reusable adapters and services may be proposed to CloudRING; private internals
and customer data may not.

This guide does not claim production readiness. It documents the public authoring
path and the local validation commands required before a module can be reviewed.
