# OCSv3 Architecture

OCSv3 separates platform core from service implementation.

```text
platform core
  reads module package
  mounts portal extension
  evaluates IAM/tenant access
  routes billing/support/evidence events
  reconciles lifecycle through declared Kubernetes and adapter contracts

service module
  owns CRDs/controllers
  owns service API
  owns billing connector
  owns support diagnostics
  owns portal microfrontend
  owns evidence receipts
```

The platform core depends on metadata and contracts, not implementation code.
Provider-specific code lives behind adapters outside `cloudring_core`. Public
core may include only synthetic/mock adapters.

This provider-specific boundary is mandatory: a module can depend on a public
adapter interface, but CloudRING public must not import the implementation.

## Request Path

1. User selects a catalog plan.
2. IAM evaluates tenant, project, entitlement, and permission.
3. Portal renders the module extension if allowed.
4. The service controller reconciles a Kubernetes claim.
5. Billing events are emitted with idempotency keys.
6. Readiness and support evidence are recorded.
7. Denied, degraded, blocked, retryable, and ready states are visible.

No production readiness claim is valid without current evidence.
