# Billing

Billing is an OCSv3 connector surface, not platform-specific code.

A billing connector declares:

- usage meters;
- cost meters and rate-card evidence;
- idempotent billing events;
- entitlement attribution;
- replay policy.

Every event must include an idempotency key. Billing replay must deduplicate by
that key and must be safe after controller retry.

If a service has no billable usage, publish an explicit non-billable policy in
the module documentation and keep conformance evidence for that decision.
