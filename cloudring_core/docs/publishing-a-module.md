# Publishing A Module

Publish only after local checks pass:

```bash
go test ./... -count=1
go run ./cmd/ocsctl validate ./path/to/module-package.json
go run ./cmd/ocsctl conformance ./path/to/module-package.json --evidence ./evidence/module-conformance.json
```

Before publication:

- keep package version stable;
- attach fresh evidence;
- document non-claims;
- verify no secrets or provider-private values are present;
- confirm rollback and restore evidence;
- confirm portal extension permissions;
- confirm billing events are idempotent.

Do not claim live production readiness from static validation alone.
