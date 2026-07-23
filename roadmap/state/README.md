# Roadmap state records

`roadmap.yaml` is the compact status index. When a goal starts, create
`GNN.json` here using `../state.schema.json`; accepted evidence records use
`../evidence.schema.json` and are referenced by digest rather than copied as raw
logs.

CI must reject a state file when its goal, requirement IDs, deployment targets or
status disagree with `roadmap.yaml`, when a dependency is not delivered, when an
evidence reference cannot be resolved under `../EVIDENCE_POLICY.md`, or when a
delivered goal lacks exact repository, artifact, live, rollback, predecessor-
regression and cleanup proof. Initial `not_started` goals need no state file.

Only sanitized state belongs in public OSS. Private inventory, identities,
endpoints and raw operational receipts stay in the owning downstream evidence
store; the public record contains their approved attestation digest and verdict.

The files under `../examples/` are schema fixtures with reserved placeholder
domains and values. They are never release or live evidence.
