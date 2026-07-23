# Service Module Authoring

Service modules integrate with CloudRING through OCSv3 (Open Cloud Standard
3) connector packages. The module source can use any implementation language or
runtime; the package handed to the platform is declarative metadata.

## Required Surfaces

Declare these surfaces in the connector package:

- service identity, owner, version, and support owner;
- capability classes and any portable dependency roles the product actually uses;
  each dependency names the target public product API, compatible version range,
  and compatibility-policy reference;
- one `local`, `remote`, or `api-only` execution profile and a versioned public
  product API; remote and API-only profiles also declare endpoint, trust-policy,
  and health references;
- explicit applicability for provision, hold or suspend, resume, resize, and
  deprovision; supported actions are idempotent and `not_applicable` actions explain
  why;
- automation actions for operator, tenant, or agent workflows;
- usage meters and billing connector refs, or an explicit non-billable profile;
- optional portal modules and UI extension metadata; when present, the
  microfrontend is signed, integrity-pinned, sandboxed, and permission-scoped;
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
4. Declare billing, federation, and commercial applicability. Link the billing
   connector only when billing is supported.
5. Link evidence refs for readiness, support, policy, data lifecycle, and recovery,
   plus UI certification only when a UI is present.
6. Validate the package:

```sh
go run ./cmd/ocsctl validate ./examples/synthetic-service-module/connector-package.json
```

## Boundary Rules

The package must be buildable and reviewable without importing platform
internals. Dependencies use `productAPIRef`, `versionRange`, and
`compatibilityPolicyRef`; `implementationRef` is not accepted. Do not name private
code paths, provider-specific APIs, service database tables, or platform
implementation packages.

This document does not claim production readiness. It defines the module
authoring contract that must pass validation before review.
