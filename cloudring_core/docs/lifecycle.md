# Lifecycle

OCSv3 lifecycle operations must be idempotent and evidence-backed.

Required operation names:

- `provision`
- `backup`
- `restore`
- `export`
- `delete`
- `retry`
- `rollback`

Mutating repair/delete/retry operations must point to rollback evidence through
`rollbackRef`. Every operation must have an `idempotencyKey`. Re-running a
request with the same key must be safe.

## Data Lifecycle

`dataLifecycle.export` and `dataLifecycle.delete` describe portable export and
deletion receipts. They are user-visible data safety commitments, not internal
implementation details.

## States

Every module must describe at least:

- `ready`
- `denied`
- `degraded`
- `blocked`
- `retryable`

Reference implementations may also expose `pending`, `provisioning`,
`deleting`, and `failed`.
