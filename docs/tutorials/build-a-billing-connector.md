# Tutorial: Build A Billing Connector

1. Define a usage meter with a stable name and unit.
2. Define a cost meter that references the usage meter.
3. Add rate-card evidence.
4. Emit idempotent billing events.
5. Link service billing profile to the billing connector.

Check:

```bash
go run ./cmd/ocsctl conformance ./my-service/module-package.json
```
