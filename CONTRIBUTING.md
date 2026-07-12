# Contributing To CloudRING

CloudRING accepts contributions that improve the public platform runtime,
services, adapters, contracts, tests, installers, operations tooling,
documentation, SDKs, and module interoperability.

## Platform ownership

Contributions may change CloudRING runtime and contracts for OCSv3, identity,
IAM, audit, catalog, billing, portal, self-service, GitOps, installers,
readiness evidence, module lifecycle, backup/restore, upgrades, rollback,
provider adapters, and developer/operator tooling.

## Service ownership

Complete service implementations may be contributed under Apache-2.0 and then
become part of CloudRING. Independently developed services remain owned by
their authors. Both must expose OCSv3 metadata and APIs and must not couple the
platform to private implementation packages. Reusable provider adapters are
welcome; concrete accounts, credentials, topology, and installation values are
not.

## Enterprise and private boundary

Do not contribute customer data, credentials, private endpoints, live
installation values, company-only implementation details, enterprise-only
modules, copied proprietary source text, or support records. Use synthetic
examples and portable capability names.

## Developer entry points

Start with `docs/public-boundary.md`, then add or update the
smallest complete public slice needed for the platform, service, or adapter.
Include focused tests for behavior, failure modes, tenant boundaries, and
portability; record only synthetic evidence for repository-level operational
claims.

## Pull request path

1. Clone the public repository and create a topic branch.
2. Enable the tracked pre-push source-safety hook for this clone:

```bash
git config core.hooksPath .githooks
```

   The hook scans every commit introduced by each pushed ref and the raw
   metadata of every introduced annotated tag in a bounded, validated tag
   chain. Unsupported publication objects fail closed. It is an early local
   check; protected-branch CI remains authoritative and cannot be replaced or
   bypassed by changing local hook configuration.
3. Run:

```bash
go test ./... -count=1
go run ./cmd/cloudring-sourcecheck scan --scope changed
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
```

4. Open a pull request with a concise summary, validation notes, and any
   public-safe evidence paths created inside your working tree.
5. Complete the CLA/DCO checks described in `CLA.md` and `DCO.md`.
6. Obtain the project founder's owner review. The founder may review
   founder-authored changes under the exact-head SafePush process documented in
   `GOVERNANCE.md` and `docs/ci-checks.md`; this exception does not extend to
   other contributors.

Do not paste private repository paths, local user paths, live provider data,
tenant records, secrets, cookies, kubeconfigs, or customer support material into
code, docs, tests, examples, evidence, commit messages, or pull request text.
