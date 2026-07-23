# Goal01 backup drill contracts

These draft 2020-12 JSON Schemas define the closed public transaction boundary
for the provider-neutral backup drill. Runtime decoding additionally rejects
duplicate keys, trailing documents, unknown fields, excessive size, and unsafe
depth. The synthetic fixture contains only invented identifiers and digests.

The adapter reads one `adapter-request` from stdin and writes one
`adapter-response` to stdout. It receives only the literal `drill` argument and
the fixed `LANG=C`, `LC_ALL=C` environment. Requests carry no credentials.
Executables, requests, responses, plans, approval scope, and the previous
journal head are SHA-256 bound. Provider-specific authentication belongs behind
the adapter boundary and must not be placed in argv, environment, responses, or
repository fixtures.

Restore observation and creation are intentionally one protocol phase:
`restore-watch-create-observe-complete`. The adapter process must keep the
observer alive from readiness through four Restore creations and durable result
capture; split watch-ready/create operations are not valid protocol steps.
The four Restores carry exactly five namespace mappings with `1,1,2,1`
cardinality. The Namespace Restore owns distinct `platform-system` and
`flux-system` destinations, and every destination is observation- and
cleanup-bound.
