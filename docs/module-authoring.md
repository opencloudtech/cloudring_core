# Module Authoring

Start from the synthetic reference service. A module package is the single
portable manifest that the platform reads:

```bash
cp -r reference/synthetic-service my-service
```

Then change names, CRD group, service class, meters, portal route, diagnostics,
and docs references. Keep lifecycle and evidence surfaces complete.

## Required Package Sections

- `metadata`: name, display name, owner, version.
- `service.spec`: capabilities, dependencies, lifecycle, automation, meters,
  billing, portal modules, UI host, analytics, Kubernetes bindings, gateway
  routes, secrets, policies, data lifecycle, states, support, evidence bundles.
- `billing`: meters, cost meters, idempotent events.
- `catalog`: class, visibility, portability, plans.
- `tenantAccess`: entitlements and permissions.
- `durability`: state class, data classes, backup policy, recovery objective.
- `distribution`, `federation`, `commercial`: release contract metadata.

Run conformance before publishing:

```bash
go run ./cmd/ocsctl conformance ./my-service/module-package.json
```
