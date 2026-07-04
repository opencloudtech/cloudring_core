# Backup Module

Backup is modeled as an OCSv3 module package owned outside CloudRING
runtime implementation. The core consumes connector metadata for backup policy,
restore workflow, support diagnostics, billing meters, evidence bundles, and
tenant access.

Storage targets, scheduler workers, and provider adapters are external module
implementation concerns. This directory contains only the contract manifest.

Validate the module manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/backup/module-package.json
```
