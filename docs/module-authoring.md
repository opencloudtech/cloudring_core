# Module Authoring

Start from the synthetic reference service. A module package is the single
portable manifest that the platform reads:

```bash
cp -r reference/synthetic-service my-service
```

Then change names, execution profile, service class, API contract, diagnostics,
and docs references. Add CRD bindings, meters, or a portal route only when they
apply. Keep lifecycle applicability and evidence surfaces complete.

## Required Package Sections

- `metadata`: name, display name, owner, version.
- `service.spec`: execution profile, public product API, capabilities, lifecycle
  applicability, automation, billing applicability, gateway routes, workload
  identity, policies, data lifecycle, states, support, and evidence bundles;
  dependencies, analytics, portal/UI, and Kubernetes bindings only when applicable.
- `billing`: meters, cost meters, and idempotent events when billing is supported;
  otherwise declare `not_applicable` with a reason and omit the connector.
- `catalog`: class, visibility, portability, plans.
- `tenantAccess`: entitlements and permissions.
- `durability`: state class, data classes, backup policy, recovery objective.
- `distribution`: release contract metadata; `federation` and `commercial` declare
  `supported` or `not_applicable`, with complete metadata only when supported.

Run conformance before publishing:

```bash
go run ./cmd/ocsctl conformance ./my-service/module-package.json
```
