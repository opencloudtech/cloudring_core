# Universal IaaS Module

Universal IaaS is owned as an OCSv3 module package, not as CloudRING Core
runtime implementation. The package manifest declares the module controller API
contract, portal UI extension, billing meters, support diagnostics, evidence
bundles, durability expectations, lifecycle actions, rollback, delete/export,
backup/restore, and denied/degraded/retry states.

KubeVirt is represented only as a portable runtime dependency contract in
`module-package.json`. CloudRING Core does not own KubeVirt controllers, service
actions, or virtualization runtime implementation.

Validate the module manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/universal-iaas/module-package.json
```
