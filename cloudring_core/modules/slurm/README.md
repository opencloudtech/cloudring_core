Slurm is owned as an optional OCSv3 module package.

The manifest keeps Slurm partition, account, QOS, accounting, readiness, and
support ownership outside CloudRING. Foundation-only profiles must be able
to report Slurm as `not-installed` without failing readiness.

Validate the manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/slurm/module-package.json
```
