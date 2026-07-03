# Provider Adapter Contract

Provider adapters let CloudRING Core reason about infrastructure without
embedding a provider implementation. Core owns the public interface shape,
synthetic fixtures, validation expectations, and evidence vocabulary. Provider
teams own SDK clients, installation logic, preflight execution, deployment
actions, inventory collection, and provider-specific configuration.

## Public Interface

An adapter package publishes a document that follows
`provider-adapter.schema.json`. The document has four portable classes:

- `inventoryClasses` describe normalized capacity, location, network, and host
  inventory fields.
- `preflightClasses` describe checks that prove the requested environment can
  satisfy the plan.
- `planClasses` describe dry-run and apply steps without implementation
  commands.
- `evidenceClasses` describe the receipts emitted by inventory, preflight, and
  plan runs.

The interface is capability-oriented. Core may validate that the classes exist
and that evidence links back to declared capabilities. Core does not import an
adapter SDK, call a provider API directly, or store deployment-specific values.

## Fixture Policy

Fixtures in this directory are synthetic. They use placeholder IDs, reserved
capability names, and public example hosts only. They must not include real
provider inventory, deployment hostnames, account identifiers, installation
profiles, support records, or operational runbooks.

The source-safety gate for `cloudring_core` rejects provider-specific
installation fields and real inventory strings before the files can be treated
as public core material.
