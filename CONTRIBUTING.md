# Contributing To CloudRING

CloudRING welcomes contributions to the public platform runtime, OCSv3,
services, provider adapters, contracts, tests, installers, operations tooling,
security, documentation, SDKs, and developer experience.

The project is early. The most valuable contributions are small, complete,
provider-neutral slices with tests and explicit failure behavior—not broad
scaffolds that imply capabilities the repository cannot yet deliver.

## Where contributions fit

### Platform core

Platform contributions may improve identity, IAM, audit, catalog, orders,
billing foundations, portal/API/CLI surfaces, durable operations, GitOps,
installation, observability, backup/restore, upgrade/rollback, evidence, or
release safety. Core code must depend on portable capabilities and versioned
contracts rather than a private service implementation or provider account.

### OCSv3 and cloud products

Service teams may contribute OCSv3 contracts, SDKs, conformance cases, reference
products, or complete open source modules. A module should expose the same
documented IAM, lifecycle, billing, support, durability, data-exit, and evidence
surfaces expected from every first-class product.

Independently developed services remain owned and licensed by their authors
unless intentionally contributed. Compatibility claims require conformance
evidence; they do not require transferring a private implementation to this
repository.

### Provider adapters

Reusable provider and site adapters are welcome when they use portable
interfaces and synthetic fixtures. Concrete accounts, credentials, private
inventory, endpoints, topology, commercial terms, and deployment values remain
with the operator.

### Product, UX, operations, and documentation

CloudRING also needs clear user journeys, accessibility, support workflows,
failure-mode tests, operability, security review, and precise documentation.
A contribution that removes an unsupported claim or makes a blocked state
visible is as valuable as new code.

## Design expectations

- Start with [VISION.md](VISION.md), the [public roadmap](roadmap/README.md), and the
  [public boundary](docs/public-boundary.md).
- Prefer the smallest vertical change that a user or operator can verify.
- Keep APIs and OCSv3 wire contracts language-neutral; Go is the current
  reference implementation, not an integration requirement.
- Use upstream Kubernetes APIs for the target runtime and hide provider details
  behind adapters.
- Include denied, degraded, retry, rollback, cleanup, and data-durability paths
  where they apply.
- Do not mark a feature ready from schemas, fixtures, manifests, or unit tests
  alone. State the exact scope of the evidence.
- Preserve a user's ability to export data, delete resources, and understand
  cost and support ownership.

## Public and private boundary

Do not contribute customer or tenant data, credentials, private endpoints,
live installation values, company-only implementation details, private
repository paths, copied proprietary text, support records, cookies,
kubeconfigs, or deployment evidence. Use synthetic identifiers and portable
capability names.

## Pull request path

1. Clone the public repository and create a topic branch.
2. Enable the tracked pre-push source-safety hook:

   ```bash
   git config core.hooksPath .githooks
   ```

   The hook scans commits and annotated-tag metadata introduced by the push.
   It is an early local check; protected-branch CI remains authoritative.
3. Run:

   ```bash
   go test ./... -count=1
   go run ./cmd/cloudring-sourcecheck scan --scope changed
   go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
   go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
   go run ./cmd/cloudring-manifestcheck --root .
   ```

4. Open a pull request with a concise problem statement, the chosen boundary,
   validation results, remaining non-claims, and any source-safe evidence.
5. Read [CLA.md](CLA.md); submitting a contribution makes the representations
   stated there. Add the `Signed-off-by` line required by [DCO.md](DCO.md) to
   every commit.
6. Obtain the review required by [GOVERNANCE.md](GOVERNANCE.md). Maintainer or
   administrator access is not permission to bypass required checks.

Do not include private or live deployment material in commit messages, pull
request text, review comments, logs, or generated evidence.
