# Evidence And Readiness

OCSv3 treats readiness as an evidence-backed claim.

Evidence bundles must include:

- owner;
- claim;
- freshness policy;
- redaction policy;
- evidence refs;
- non-claims;
- review path.

Readiness checks must link to evidence. Stale evidence blocks promotion. The
phrase stale evidence must be treated as a blocker, not a warning. Raw secrets,
host-local paths, private provider endpoints, cookies, tokens, tenant data, and
customer data must never be written to public evidence.
