# Provider Adapters

Provider adapters connect a module controller to an external or internal
provider. OCSv3 core defines the adapter contract; provider implementation is
outside CloudRING public.

The public inventory protocol is a narrow boundary between the site plan and a
downstream provider implementation. Core owns:

- the five binding classes and strict request/receipt JSON shapes;
- canonical ordering, bounded decoding, and exact request-to-receipt matching;
- SHA-256 binding to the site profile, run nonce, adapter catalog, adapter
  executable, and each private observation;
- the fixed `provider-inventory` protocol scope;
- fail-closed statuses and source-safe blocker IDs.

A downstream adapter owns its private binding catalog, provider API clients,
authentication, retries, and collection of actual values. The request and
receipt may carry only symbolic references, binding classes, and SHA-256
commitments. They must not carry raw IP addresses, server or network object
identifiers, credentials, tokens, or provider response bodies.

An observation commitment uses a protected 32-byte per-run nonce, its symbolic
reference, and the canonical private value under the
`cloudring/provider-inventory/observation/v1` domain. Only the nonce SHA-256 is
published. A direct unsalted hash of an address or provider identifier is not
valid because a small candidate dictionary could recover the original value.

The site plan keeps provider discovery and host preparation separate.
`inventory` contains only adapter, region, provider-resource, management-address,
and provisioning-address references. `host-baseline` contains the persistence
and verification references for host runtime capacity, and `bootstrap` depends
on that phase. This prevents a successful provider inventory lookup from being
misreported as a completed host baseline.

Allowed in `cloudring_core`:

- synthetic adapter examples;
- mock provider state;
- public schemas;
- provider-neutral interface docs.

Not allowed in `cloudring_core`:

- real provider endpoints;
- account IDs;
- credentials;
- operational live inventory;
- private runbooks.

Inventory adapters return `ready`, `blocked`, `denied`, `retryable`, or `failed`.
A non-ready receipt contains sorted blocker IDs and may echo only the canonical
ordered subset observed before failure, including no observations. It does not
echo provider error text. A ready receipt must echo the full requested set.
`productionReady` is always false because discovery is only one input to later
preflight, host-baseline, bootstrap, and live acceptance gates.
