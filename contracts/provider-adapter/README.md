# Provider Adapter Contract

Provider adapters let CloudRING reason about infrastructure without
embedding a provider implementation. CloudRING owns the public interface shape,
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

The interface is capability-oriented. CloudRING may validate that the classes exist
and that evidence links back to declared capabilities. Core does not import an
adapter SDK, call a provider API directly, or store deployment-specific values.

## Inventory Request and Receipt

`inventory-request.schema.json` and `inventory-receipt.schema.json` define the
closed protocol for provider discovery:

1. Core sends a deterministic, bounded list of symbolic references and their
   binding classes under the fixed `provider-inventory` scope. The allowed
   classes are provider-adapter, region, provider-resource,
   management-address, and provisioning-address references.
2. The downstream adapter resolves those references in its private catalog and
   returns the exact same ordered reference/class set when ready. Each
   observation contains only a salted SHA-256 commitment to the private
   observed value.
3. The receipt is bound to the canonical request, adapter executable, and
   adapter catalog by SHA-256. Raw addresses, provider object identifiers,
   credentials, tokens, and response bodies never belong in the receipt.

The Go runtime in `pkg/provideradapter` canonicalizes requests, rejects unknown
or duplicate fields, requires an exact receipt match, and limits a request to
4,096 bindings and one MiB. A current three-node converged site has eleven
discovery references: one adapter, one region, and three each for provider
resources, management addresses, and provisioning addresses. The protocol is
generic and does not hardcode that site size.

Observation commitments are:

```text
SHA256("cloudring/provider-inventory/observation/v1" || 0x00 ||
       rawRunNonce || 0x00 || reference || 0x00 || canonicalPrivateValue)
```

`rawRunNonce` is exactly 32 random bytes supplied through protected runtime
input. Only its SHA-256 appears in the request. The raw nonce and canonical
private value stay downstream so an IP address or provider object identifier
cannot be recovered by hashing a small candidate dictionary. An adapter must
not substitute a plain, unsalted SHA-256 of the private value.

Receipt status is one of `ready`, `blocked`, `denied`, `retryable`, or `failed`.
Non-ready receipts carry sorted, source-safe blocker IDs and may include the
canonical ordered subset observed before the blocker, including an empty set.
They may not invent, duplicate, or reorder bindings. `ready` requires the full
request set and means only that provider discovery completed; it is not a
deployment or site-release decision. Both request and receipt therefore
require `productionReady: false`.

## Fixture Policy

Fixtures in this directory are synthetic. They use placeholder IDs, reserved
capability names, and public example hosts only. They must not include real
provider inventory, deployment hostnames, account identifiers, installation
profiles, support records, or operational runbooks.

The source-safety gate for `cloudring_core` rejects provider-specific
installation fields and real inventory strings before the files can be treated
as CloudRING public material.
