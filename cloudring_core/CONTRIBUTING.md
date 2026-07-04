# Contributing To CloudRING Core

CloudRING Core accepts contributions that improve public platform contracts,
validation, documentation, SDK guidance, and module interoperability.

## Core ownership

Contributions may change public core contracts for OCSv3 registry and
validation, IAM and policy interfaces, GitOps abstractions, readiness evidence,
module lifecycle gates, BOM and rollback contracts, portal shell extension
points, provider adapter interfaces, and developer documentation.

## Service ownership

Service-specific controllers, billing integrations, user interfaces, support
flows, durability logic, and provider implementations belong in independently
owned modules. Public core changes should describe the metadata and contracts
those modules publish, not embed their runtime implementation.

## Enterprise and private boundary

Do not contribute customer data, credentials, private endpoints, live
installation values, company-only implementation details, enterprise-only
modules, copied proprietary source text, or support records. Use synthetic
examples and portable capability names.

## Developer entry points

Start with `docs/public-boundary.md`, then add or update the smallest public
contract needed for your module or adapter. Include focused tests or checks for
new validation behavior, and record evidence for operational claims.

## Pull request path

1. Clone the public repository and create a topic branch.
2. Run:

```bash
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

3. Open a pull request with a concise summary, validation notes, and any
   public-safe evidence paths created inside your working tree.
4. Complete the CLA/DCO checks described in `CLA.md` and `DCO.md`.

Do not paste private repository paths, local user paths, live provider data,
tenant records, secrets, cookies, kubeconfigs, or customer support material into
code, docs, tests, examples, evidence, commit messages, or pull request text.
