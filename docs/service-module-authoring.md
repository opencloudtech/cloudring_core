# Service Module Authoring

Service modules integrate with CloudRING through OCSv3 (Open Cloud Standard
3) connector packages. The module source can use any implementation language or
runtime; the package handed to the platform is declarative metadata.

## Required Surfaces

Declare these surfaces in the connector package:

- service identity, owner, version, and support owner;
- capability classes and portable dependency roles;
- idempotent lifecycle actions with idempotency keys;
- automation actions for operator, tenant, or agent workflows;
- usage meters and billing connector refs;
- portal modules with route, API ref, host ref, mount ref, and permissions;
- UI extension metadata with microfrontend host and runtime contract;
- Gateway API route refs that are portable;
- workload identity and secret refs, with no raw secret values;
- policy refs for placement, data, support access, or audit decisions;
- export and delete lifecycle actions;
- degraded and blocked user-visible states;
- support diagnostics and evidence bundles;
- readiness and durability refs.

## Authoring Flow

1. Choose a module id and capability class that describe the user-visible
   service rather than the backing implementation.
2. Define the public API refs and lifecycle actions the platform may invoke.
3. Add tenant permissions and entitlement refs.
4. Link the billing connector in `service.spec.billing.connectorRef`.
5. Link evidence refs for readiness, support, UI certification, policy, data
   lifecycle, and recovery.
6. Validate the package:

```sh
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
```

## Boundary Rules

The package must be buildable and reviewable without importing platform
internals. Use public OCSv3 fields such as `implementationRef` only for portable
adapter refs owned by the service team. Do not name private code paths,
provider-specific APIs, service database tables, or platform implementation
packages.

This document does not claim production readiness. It defines the module
authoring contract that must pass validation before review.
