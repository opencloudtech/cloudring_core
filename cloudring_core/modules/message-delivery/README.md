# Message Delivery Module

Message delivery is modeled as an OCSv3 module package. CloudRING Core owns the
connector contract for notification intent, retry, evidence, support, and usage
metering; transport implementations stay outside core.

Mail relays, webhook dispatchers, and other delivery adapters must integrate by
reference through the manifest. No transport endpoint, account secret, or sender
credential is stored in this module directory.

Validate the module manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/message-delivery/module-package.json
```
