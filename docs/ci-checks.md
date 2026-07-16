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
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
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
| PostgreSQL integration | The transactional-state CAS, migrations, digest, and concurrent-writer behavior must pass against a real digest-pinned PostgreSQL service with synthetic, non-secret test configuration. |
| Windows | The same unit suite is run on `windows-latest` as a portability signal. Native Windows support is not a release-readiness blocker for the current goal. |
| OCS validation | The synthetic connector package must pass `go run ./cmd/ocsctl validate`. |
| OCS conformance | The reference synthetic module must pass `go run ./cmd/ocsctl conformance`. |
| Source-safety | The Go scanner must approve the complete tree and pre-push commit range, including intermediate commits and reviewed non-text artifacts. |
| Security | CodeQL, govulncheck, gosec, and both current-tree and Git-history secret scans must pass without broad exclusions. |
| Supply chain | Actions must be commit-pinned; workflows must be syntax-checked and must not request unexpected write permissions or PR secrets. The release-only workflow builds the Linux CLI bundle, generates a CycloneDX 1.6 SBOM, uploads both as one short-retention artifact set, and creates GitHub/Sigstore build and SBOM attestations. |
| License and contribution docs | `LICENSE`, `NOTICE`, `CONTRIBUTING.md`, `SECURITY.md`, `GOVERNANCE.md`, `CLA.md`, `DCO.md`, and `TRADEMARKS.md` must exist in the public root. |

The PostgreSQL service is an isolated CI dependency. This does not claim that
a provider database or its backup and failover have been verified live.

This contract does not require live provider credentials, secret environment
variables, network mutation, or live Kubernetes access. Passing it only means
the CloudRING public tree is locally safe to publish and validate; it is not a
production readiness claim.

## SafePush trust boundary

`.github/workflow-policy.json` binds every required workflow to its exact
canonical YAML/JSON semantic surface. The supply-chain job also rejects an
unexpected workflow or job, event/permission expansion, mutable action or
runtime image, credential context, conditional skip, and `continue-on-error`.
Changing a reviewed workflow requires changing its recorded digest in the same
review.

This repository check is defense in depth, not a self-authenticating policy: a
pull request controls the workflow revision that evaluates that pull request.
The server-side trust control for external contributors is protected `main`
with an up-to-date branch, one approving owner review, stale-review dismissal,
last-push approval, required conversations and checks, and no force-push or
branch deletion. The project founder and lead maintainer, `@trukhinyuri`, is
the final acceptance authority. Contributor-authored changes require that
owner review; automated or AI-assisted review does not replace the founder's
decision.

GitHub does not let a pull-request author submit the approval required by the
same branch rule. While the repository has a single founder/administrator, a
founder-authored pull request may therefore use the administrator merge path
after a recorded founder review of the exact head SHA. This narrow owner path
still requires every configured status check to succeed on the current head,
an up-to-date base, resolved conversations, source-safety approval, and a
post-merge `main` verification. It does not authorize direct pushes, skipped
checks, force-pushes, branch deletion, or bypasses for other contributors.
Adding another administrator changes this trust assumption and requires an
explicit governance and branch-protection review before that access is
granted.

## Release provenance

`.github/workflows/release-provenance.yml` runs only for a version tag or an
explicit workflow dispatch. It never runs for a pull request. The job builds
all public Go commands for Linux AMD64 with read-only modules and embedded VCS
metadata, packages `LICENSE`, `NOTICE`, and the CycloneDX SBOM, records the
bundle checksum, and creates a GitHub artifact attestation for the bundle and
SBOM. The only write permissions are job-local `id-token` and `attestations`
permissions for this exact reviewed workflow and job.

After downloading the bundle from its workflow run, verify the provenance and
SBOM attestation against this repository:

```bash
gh attestation verify cloudring-linux-amd64.tar.gz \
  --repo opencloudtech/CloudRING
```

An attestation binds an artifact to its accepted source and build workflow; it
does not replace vulnerability scanning, code review, release policy, or live
service validation.
