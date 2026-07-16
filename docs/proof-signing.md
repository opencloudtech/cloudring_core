# Proof signing

`cloudring-proof` provides a provider-neutral Ed25519 trust boundary for
release, restore, continuity, and conformance evidence. It signs the canonical
JSON payload selected by the owning proof schema and emits a detached signature
that can be verified offline.

Private material is never accepted in an argument, environment variable, or
regular file. Key generation writes the only private-key document to an
inherited pipe or socket. Signing reads that document from an inherited pipe or
socket. The calling secret broker is responsible for storing it and replaying
it without exposing the value in logs.

## Provision a key

The secret consumer owns file descriptor 3 in this example. The command writes
the public rotation policy as a new, non-overwriting owner-only artifact and
prints only a status marker:

```sh
cloudring-proof key generate \
  --key-id backup-proof-2026-01 \
  --secret-output-fd 3 \
  --trust-policy /protected/backup-proof-trust-policy.json
```

If private-key delivery fails, the newly written public policy is removed. The
generated private document has the closed `ProofSigningKey` schema and contains
a base64-encoded Ed25519 seed. Never commit or retain that document outside the
configured secret manager.

Derive the same public policy after a fresh clone by replaying the stored key
through a protected pipe. This is a read-only key operation and never prints
private material:

```sh
cloudring-proof key public \
  --key-fd 3 \
  --trust-policy /protected/derived-backup-proof-trust-policy.json
```

## Sign and verify

```sh
cloudring-proof sign \
  --payload /protected/canonical-proof-payload.json \
  --key-fd 3 \
  --signature /protected/proof-signature.json

cloudring-proof verify \
  --payload /protected/canonical-proof-payload.json \
  --signature /protected/proof-signature.json \
  --trust-policy /protected/backup-proof-trust-policy.json
```

The CLI strictly rejects duplicate JSON fields, trailing JSON, unknown fields
in keys, signatures, and policies, oversized payloads, duplicate trust-key IDs,
non-canonical base64, untrusted keys, digest mismatches, signature tampering,
and output replacement. Whitespace and object-key ordering in the input payload
are normalized before both signing and verification.

An aggregate proof format remains responsible for selecting every signed field,
binding domain and key metadata, and validating the evidence before promotion.
A valid cryptographic signature does not turn incomplete or stale evidence into
a readiness claim.
