# Conformance

Run validation and conformance before publishing a module. From a fresh public
clone:

```bash
gh repo clone opencloudtech/cloudring_core
cd cloudring_core
go mod download
go test ./... -count=1
go run ./cmd/ocsctl validate ./cloudring_core/examples/synthetic-service-module/connector-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

For a module under review, run the same commands against that module package:

```bash
go run ./cmd/ocsctl validate ./cloudring_core/reference/synthetic-service/module-package.json
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json
```

Use JSON evidence in CI:

```bash
mkdir -p evidence
go run ./cmd/ocsctl conformance ./cloudring_core/reference/synthetic-service/module-package.json --evidence ./evidence/reference-conformance.json
```

Conformance checks:

- canonical OCSv3 package validation;
- service connector completeness;
- billing linkage;
- lifecycle coverage and idempotency;
- backup, restore, export, delete, retry, rollback;
- denied, degraded, blocked, retryable, ready states;
- portal and microfrontend declarations;
- IAM and tenant access;
- support diagnostics;
- evidence bundles and durability;
- analytics events;
- provider-neutral public boundary.

Run source-safety before publishing to prove no local host paths, private
provider endpoints, hardcoded tokens, or implementation leaks entered public
core. Public contributors should keep generated evidence local unless a
synthetic receipt is intentionally reviewed for inclusion.
