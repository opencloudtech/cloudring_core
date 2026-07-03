GPU is owned as an optional OCSv3 module package.

CloudRING Core consumes only `module-package.json` metadata for GPU catalog,
readiness, billing, support, and lifecycle state. The module is not a
foundation prerequisite and defaults to `not-installed` / `disabled` until a
provider supplies reviewed hardware, driver, policy, backup, and rollback
evidence.

Compatibility and BOM windows are declared in the manifest under
`xCloudRING.compatibilityBOM`. Those windows are contracts for a future module
owner; they are not live installation readiness claims.

Validate the manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/gpu/module-package.json
```
