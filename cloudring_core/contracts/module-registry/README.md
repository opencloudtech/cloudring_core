# Module Registry Contract

This directory defines the public CloudRING Core module registry lifecycle
contract. The registry is a source-safe, synthetic contract for module metadata,
lifecycle operations, dependency resolution, idempotent planning, audit and
evidence receipts, and rollback hooks.

Core validates registry shape and lifecycle safety only. Service teams own
service-specific controllers, data schemas, mutation commands, and provider
integration code outside this public core contract.

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

## Non-Claims

The fixtures do not prove a deployed marketplace, service installation, billing
settlement, support workflow, tenant data migration, or provider operation. They
only prove that CloudRING Core can validate the portable lifecycle contract
without importing service implementation details.
