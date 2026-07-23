# Backup Module

Backup is modeled as an OCSv3 service module. CloudRING owns the reusable
backup/restore contracts, typed decoders, collectors, validators, lifecycle
interfaces, and source-safe provider adapter protocols. The current public
runtime slice includes the Velero 1.18.2 CSI data-mover restore-proof collector
documented in `docs/restore-proof-collector.md`.

The provider-neutral Goal01 transaction engine is documented in
`docs/backup-drill-executor.md`. It adds strict plan, approval, adapter,
journal, and receipt contracts plus `cloudring-backup drill
preflight|apply|recover|rollback`. A concrete live provider adapter remains a
downstream responsibility.

Storage-service implementations and scheduler workers remain module concerns.
Reusable provider adapters may be contributed to CloudRING; account
credentials, endpoints, installation values, tenant data, and live evidence
remain downstream. This directory contains the OCSv3 module manifest, while
generic runtime code lives under `pkg/backup` and `cmd/cloudring-backup`.

Validate the module manifest with:

```sh
go run ./cmd/ocsctl validate ./modules/backup/module-package.json
```

The module manifest and restore-proof collector do not claim that an object
store, backup schedule, immutable retention policy, or live restore drill is
already deployed.
