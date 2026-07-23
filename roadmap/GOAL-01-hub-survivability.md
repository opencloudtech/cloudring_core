# Legacy Goal 01 transition contract

The former “Reference-Cell Critical-Path Survivability” goal is preserved as a
compatibility entry point for existing issue links and
[`specifications/goal-01.md`](../specifications/goal-01.md). It is no longer an
independent scheduling or completion authority.

Its valid obligations are carried without weakening into the canonical G00-G27
program. [`COVERAGE.md`](COVERAGE.md) is the row-level transition ledger: every
legacy `CR-FND-*`, `CR-STO-*`, `CR-K8S-*`, `CR-OCS-*`, `CR-OPS-*` and
`CR-DPL-*` identifier names its single owning goal and canonical `CR-GNN-*`
requirement. In summary:

- G00 establishes truthful source, issue, evidence and protected-delivery state;
- G02-G03 deliver the HA empty provider and durable resource/operation kernel;
- G07 preserves provider-neutral OCS durability contracts;
- G18 delivers verified off-cell backup, restore and immutability;
- G22 owns operable mutation, drift-repair and disaster-recovery routines;
- G23 executes one-server-loss, state continuity and integrated
  upgrade/failure proof;
- G27 performs the final broad review, fixes and cumulative release gate.

The authoritative graph is [`roadmap.yaml`](roadmap.yaml), historical work-item
transition is in [`LEGACY_WORK_MAP.md`](LEGACY_WORK_MAP.md), row-level
requirement transition is in [`COVERAGE.md`](COVERAGE.md), and accepted status
belongs under [`state/`](state/). No prior Goal 01 checklist state, dated
observation, legacy identifier or compatibility link can make a canonical goal
delivered.

Concrete provider values, credentials, private topology, tenant data and live
receipts remain downstream. Reusable validators, contracts, SDKs, tests and
runbooks remain public OSS.
