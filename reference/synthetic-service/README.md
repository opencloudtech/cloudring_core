# Synthetic Service Reference

This reference module is a provider-neutral OCSv3 implementation for local
developer validation. It is shaped like a production service module, but it
uses only a mock provider adapter and synthetic receipts.

Run:

```bash
go run ./cmd/ocsctl validate ./reference/synthetic-service/module-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
go test ./... -count=1
```

The module demonstrates the local execution profile, a versioned public product
API, lifecycle applicability for provision, hold, resume, resize, and
deprovision, plus product-specific backup, restore, export, retry, and rollback.
It also declares ready, denied, degraded, blocked, retryable, deleting, and
failed states; billing events; support diagnostics; evidence receipts; and an
optional signed, integrity-checked, sandboxed portal extension. It does not call
a live provider API or claim live service launch approval.
