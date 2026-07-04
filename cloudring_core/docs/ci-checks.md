# CloudRING Core CI Checks

Public CI for a clean clone should run these read-only checks:

```bash
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

The public CI contract covers these checks:

| Check | Contract |
| --- | --- |
| Go tests | The exported public module must pass `go test ./... -count=1`. |
| OCS validation | The synthetic connector package must pass `go run ./cmd/ocsctl validate`. |
| OCS conformance | The reference synthetic module must pass `go run ./cmd/ocsctl conformance`. |
| Source-safety | Public docs, examples, and generated evidence must not include private paths, live provider values, tenant data, or credentials. |
| License and contribution docs | `LICENSE`, `NOTICE`, `CONTRIBUTING.md`, `SECURITY.md`, `GOVERNANCE.md`, `CLA.md`, `DCO.md`, and `TRADEMARKS.md` must exist in the public root. |

This contract does not require live provider credentials, secret environment
variables, network mutation, or live Kubernetes access. Passing it only means
the public-core tree is locally safe to publish and validate; it is not a
production readiness claim.
