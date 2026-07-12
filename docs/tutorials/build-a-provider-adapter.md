# Tutorial: Build A Provider Adapter

1. Start with a mock adapter and typed states.
2. Keep real provider endpoints and credentials outside `cloudring_core`.
3. Map provider responses to ready, denied, degraded, blocked, retryable, and
   failed states.
4. Emit sanitized evidence receipts.
5. Prove retry and rollback with idempotency keys.

Provider adapters are implementation boundaries. The platform core sees only
the OCSv3 contract and evidence.
