# Module Registry Contract

This directory defines the public CloudRING module registry lifecycle
contract. The registry is a source-safe, synthetic contract for module metadata,
lifecycle operations, dependency resolution, idempotent planning, audit and
evidence receipts, and rollback hooks.

CloudRING validates registry shape and lifecycle safety only. Service teams own
service-specific controllers, data schemas, mutation commands, and provider
integration code outside this CloudRING public contract.

## Contract Rules

- Module IDs are unique within a registry document.
- Lifecycle states include `installable`, `installed`, `suspended`,
  `deprecated`, and `not-installed`.
- Operations include `install`, `update`, `remove`, `suspend`, and `deprecate`.
- Each operation is idempotent, policy-aware, and includes an idempotency key,
  audit receipt, evidence receipt, rollback hook, and next action.
- Install plans resolve declared module dependencies before the target module.
- Reinstalling an already installed module emits a no-op plan with the declared
  operation ID and idempotency key.
- Registry records must not include service implementation references, provider
  endpoints, deployment values, credentials, or direct mutation commands.

## Files

- `module-registry.schema.json` describes the machine-readable registry shape.
- `fixtures/synthetic-module-registry.json` is the valid registry fixture.
- `fixtures/invalid-*.json` fixtures prove fail-closed verifier behavior.

Validate the contract through the public Go consumer:

```bash
go run ./cmd/cloudring-registry validate \
  ./contracts/module-registry/fixtures/synthetic-module-registry.json
```

The validator checks the typed registry shape, source-safety flags, dependency
graph, lifecycle operations, and idempotent plan references. It never loads or
executes a service implementation and returns exit code `2` for a blocked
registry.

## Non-Claims

The fixtures do not prove a deployed marketplace, service installation, billing
settlement, support workflow, tenant data migration, or provider operation. They
only prove that CloudRING can validate the portable lifecycle contract
without importing service implementation details.
