# Support Module

Support is modeled as an OCSv3 module package. CloudRING consumes support
case, diagnostic, escalation, evidence, and usage-meter contracts from the
manifest without owning a ticketing or contact-center implementation.

External support systems integrate through provider adapters and referenced
workload identities. This module directory stores only portable contracts and
review paths.

Validate the module manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/support/module-package.json
```
