# CloudRING CI Checks

Public CI for a clean clone runs these read-only checks on the latest security
patch of the minimum supported Go release (1.25) and runs unit tests again on
the current supported release (1.26):

```bash
go mod download
go test ./... -count=1
go test -race ./... -count=1
go vet ./...
go run ./cmd/cloudring-sourcecheck scan --scope full
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

Enable the tracked local pre-push gate once per clone:

```bash
git config core.hooksPath .githooks
```

The hook rejects unsafe intermediate commits before transport. GitHub
protected-branch rules and the `source-safety` workflow remain authoritative;
local hook configuration is never accepted as merge evidence.

The public CI contract covers these checks:

| Check | Contract |
| --- | --- |
| Go tests | The public module must pass tests on supported minimum/current Go releases, plus race, vet, read-only module graph, and build checks. |
| Windows | The same unit suite, including native ACL and replacement behavior, must pass on `windows-latest` with both supported Go releases. |
| OCS validation | The synthetic connector package must pass `go run ./cmd/ocsctl validate`. |
| OCS conformance | The reference synthetic module must pass `go run ./cmd/ocsctl conformance`. |
| Source-safety | The Go scanner must approve the complete tree and pre-push commit range, including intermediate commits and reviewed non-text artifacts. |
| Security | CodeQL, govulncheck, gosec, and both current-tree and Git-history secret scans must pass without broad exclusions. |
| Supply chain | Actions must be commit-pinned; workflows must be syntax-checked and must not request unexpected write permissions or PR secrets. |
| License and contribution docs | `LICENSE`, `NOTICE`, `CONTRIBUTING.md`, `SECURITY.md`, `GOVERNANCE.md`, `CLA.md`, `DCO.md`, and `TRADEMARKS.md` must exist in the public root. |

This contract does not require live provider credentials, secret environment
variables, network mutation, or live Kubernetes access. Passing it only means
the CloudRING public tree is locally safe to publish and validate; it is not a
production readiness claim.
