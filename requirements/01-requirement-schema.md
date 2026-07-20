# 01 — Requirement Schema

Every requirement in `domains/*.md` MUST follow this schema exactly.
Machine parsing of the corpus depends on it.

## Identity

- ID format: `CR-<DOMAIN>-NNN`, e.g. `CR-NET-010`.
- Numbering is sparse and increments by 10 (`-010`, `-020`, …) to allow
  insertions. IDs are never renumbered and never reused after retirement.
- Domain codes:

| Code | Domain file |
|---|---|
| FND | `domains/10-platform-foundation.md` |
| CMP | `domains/11-compute-virtualization.md` |
| NET | `domains/12-network.md` |
| STO | `domains/13-storage-backup-dr.md` |
| K8S | `domains/14-kubernetes-containers.md` |
| IAM | `domains/15-iam-identity-security.md` |
| BIL | `domains/16-billing-finops.md` |
| OCS | `domains/17-ocs-service-connectors.md` |
| MKT | `domains/18-marketplace-catalog.md` |
| CUX | `domains/19-portal-ux-selfservice.md` |
| OBS | `domains/20-observability.md` |
| OPS | `domains/21-ops-sre-support.md` |
| DPL | `domains/22-deployment-iac-cicd.md` |
| FED | `domains/23-federation-global-portal.md` |
| DAT | `domains/24-data-services.md` |
| AGT | `domains/25-agent-governance.md` |

## Requirement block format

```markdown
### CR-NET-010 — <short title>
- **Priority:** P0 | P1 | P2
- **Status:** proposed | accepted | blocked | retired
- **Actors:** provider | vendor | service-team | tenant | operator | agent | auditor
- **Problem:** the concrete problem or gap (1–3 sentences).
- **Requirement:** normative statements using MUST / SHOULD / MAY.
- **Acceptance evidence:** what verifiable proof demonstrates satisfaction
  (tests, live evidence class, contract checks, documentation).
- **Non-goals:** what is explicitly NOT required (prevents scope creep).
- **Non-claims:** what is NOT yet proven/claimed (honesty boundary; may be
  empty only when acceptance evidence is already linked).
- **Stop conditions:** MANDATORY for risk classes money / data / keys / trust /
  exposure / deletion / migration / settlement: when to halt and escalate.
  Write `n/a` only for requirements with no risk-class impact.
- **Traceability:** sources as provenance classes (see below); related
  requirement IDs.
```

## Priorities

- **P0** — required for a deployable, operable, honest production platform
  (the "real open baseline"). The platform MUST NOT be called ready with any
  P0 unmet.
- **P1** — required for the full product vision (federation, marketplace
  economics, advanced services).
- **P2** — valuable extensions and optimizations.

## Statuses

- `proposed` — authored, not yet accepted by owner review.
- `accepted` — approved for implementation.
- `blocked` — cannot proceed; the blocker MUST be stated in Non-claims.
  Blocked is a first-class honest state, never silently converted.
- `retired` — withdrawn; the block stays in place with the reason recorded.

## Honesty and evidence rules

1. A requirement's status changes to `accepted` only with owner review.
2. Delivery claims live outside this corpus (implementation evidence links
   back). This corpus states WHAT and WHY, with acceptance criteria — not
   unverified delivery claims.
3. Evidence states used across the project: `verified`, `blocked`, `stale`,
   `synthetic`. Only `verified` non-synthetic evidence promotes.
4. UTC timestamps everywhere; all records append-only.

## Provenance classes (for Traceability fields)

- `vision-deck` — the CloudRING product vision presentation
- `req-ccp` — control-panel requirements package
- `req-acr-singular` / `req-acr-plural` — agent-native cloud requirement packages
- `req-history` — historical CloudRING requirements corpus
- `legacy-platform-a` — owner's previous production cloud platform (private sources)
- `legacy-platform-b` — a hyperscaler-class cloud monorepo snapshot (private sources)
- `current-core` — existing CloudRING OSS contracts/modules/docs

No proprietary text, brand names of reference platforms, endpoints, tenant
data, or secrets may appear in requirement text (see `02-source-and-method.md`).

## Registry

`registry/requirements.json` is GENERATED from the Markdown corpus and
contains one record per requirement: id, title, domain, priority, status,
actors, file, risk classes, traceability. Drift between corpus and registry
fails validation.
