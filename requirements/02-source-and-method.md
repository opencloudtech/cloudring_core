# 02 — Sources and Method

## Method

This corpus was synthesized in three steps:

1. **Source analysis** — each source was independently analyzed and reduced
   to a factual digest (capabilities, architecture, operations practices,
   lessons, pitfalls, candidate requirements).
2. **Synthesis** — digests were merged, deduplicated, and rewritten as
   original normative requirements following `01-requirement-schema.md`.
3. **Validation** — registry generation, drift checks, language checks
   (English only), and source-safety checks.

## Sources (provenance classes only)

| Class | Description |
|---|---|
| `vision-deck` | The CloudRING product vision presentation (11 pages): P2P federation, Open Cloud Standard, Infrastructure/Services Pods, Enterprise Marketplace, Global Cloud Portal, monetization principles |
| `req-ccp` | Control-panel requirements package (`req/ccp-requirements`): brand-neutral console capability/UX/journey/IA requirements derived from observing a live cloud console as *observed behavior* |
| `req-acr-singular` | Agent-native cloud requirements (`req/agents_cloud_requirement`): 60 atomic requirements, evidence-envelope discipline, production promotion gates |
| `req-acr-plural` | Agent-native cloud requirements (`req/agents_cloud_requirements`): 15 requirements, scenarios, risk and OCS-fit matrices |
| `req-history` | Historical CloudRING corpus (`req/history_requirements`): 600 files — vision, platform lifecycle, operations/resilience, agent governance, ADRs, evidence discipline |
| `legacy-platform-a` | The owner's previous production public cloud platform (private monorepo, ~2,200 projects): VMware-based IaaS, managed Kubernetes, DBaaS fleet, billing pipeline, marketplace, hub/backoffice, IaC/GitOps operations |
| `legacy-platform-b` | A hyperscaler-class public cloud monorepo snapshot (private): layered compute/storage architecture, IAM/KMS model, billing/metering, network/LB products, bootstrap/deploy discipline |
| `current-core` | Existing CloudRING OSS artifacts: OCSv3 connector model, contracts, 12 module manifests, SDK, reference service, policies |

## Source-safety rules (binding)

1. **No copied text.** Requirement text is original authorship. Sources inform
   WHAT and WHY, never the wording.
2. **No brands or identifiers of reference platforms** in corpus text. Use
   provenance classes and generic terms.
3. **No endpoints, hostnames, credentials, tenant data, or secrets.** Any
   discovered during analysis is treated as [REDACTED] and never carried over.
4. **Observed behavior, not design cloning.** Third-party products are
   treated as observations of product behavior, not as designs to replicate.
5. **Lessons, not liability.** Pitfalls from legacy platforms are converted
   into positive requirements (what CloudRING must do), without disparaging
   references.
6. Violations are CI-enforced where automatable (source-safety scan) and
   owner-reviewed otherwise.

## Known gaps and honesty notes

- The vision deck describes the full P2P federation and global settlement
  vision; federation requirements (FED) are necessarily forward-looking and
  mostly P1/P2 — they are honest about not being proven.
- Legacy sources predate the current Go/upstream-Kubernetes policy; where they
  conflict, the current policy (`current-core`, requirements 75) wins.
- Duplicated sibling packages (`req-acr-singular` vs `req-acr-plural`) were
  merged; where they disagreed, the stricter variant was taken.
