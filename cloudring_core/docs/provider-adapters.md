# Provider Adapters

Provider adapters connect a module controller to an external or internal
provider. OCSv3 core defines the adapter contract; provider implementation is
outside CloudRING public.

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

Adapters must return typed denied, degraded, blocked, retryable, ready, and
failed states with evidence refs.
