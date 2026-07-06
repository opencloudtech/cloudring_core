HPC is owned as an optional OCSv3 module package.

The manifest declares HPC reservations, queue-facing service metadata,
durability receipts, support diagnostics, and compatibility windows. It is
not a foundation prerequisite and can remain `not-installed` or `disabled`
without blocking foundation-only deploys.

Validate the manifest with:

```sh
go run ./cmd/ocsctl validate ./cloudring_core/modules/hpc/module-package.json
```
