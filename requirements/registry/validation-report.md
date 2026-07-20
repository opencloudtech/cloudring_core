# Requirements Corpus Validation Report

Generated: 2026-07-20 03:14 UTC
Tool: `registry/validate.py` (stdlib-only; run `python3 validate.py` to
regenerate `requirements.json` and re-run all checks, or
`python3 validate.py --check` as a CI drift gate that writes nothing).
Scope: `domains/*.md` (16 domain files) against `01-requirement-schema.md`
and the source-safety rules in `02-source-and-method.md`.
This report only reports; no domain file was modified.

## Totals

- Requirements parsed: **330** (all 16 schema-listed domain files present;
  no unmapped files)
- Registry records written to `registry/requirements.json`: **330**
- Registry drift at generation time: **none**

### Per domain

| Code | File | Requirements |
|---|---|---|
| FND | `domains/10-platform-foundation.md` | 21 |
| CMP | `domains/11-compute-virtualization.md` | 18 |
| NET | `domains/12-network.md` | 22 |
| STO | `domains/13-storage-backup-dr.md` | 23 |
| K8S | `domains/14-kubernetes-containers.md` | 18 |
| IAM | `domains/15-iam-identity-security.md` | 24 |
| BIL | `domains/16-billing-finops.md` | 21 |
| OCS | `domains/17-ocs-service-connectors.md` | 20 |
| MKT | `domains/18-marketplace-catalog.md` | 19 |
| CUX | `domains/19-portal-ux-selfservice.md` | 21 |
| OBS | `domains/20-observability.md` | 23 |
| OPS | `domains/21-ops-sre-support.md` | 19 |
| DPL | `domains/22-deployment-iac-cicd.md` | 22 |
| FED | `domains/23-federation-global-portal.md` | 18 |
| DAT | `domains/24-data-services.md` | 20 |
| AGT | `domains/25-agent-governance.md` | 21 |

### Per priority

| Priority | Count |
|---|---|
| P0 | 180 |
| P1 | 109 |
| P2 | 41 |

### Per status

| Status | Count |
|---|---|
| proposed | 330 |
| accepted | 0 |
| blocked | 0 |
| retired | 0 |

## Checks that passed clean

- Required fields: all 330 blocks contain all 10 schema fields, non-empty.
- ID format and spacing: every ID matches `CR-<DOMAIN>-NNN` with
  zero-padded, multiple-of-10 numbering; no duplicates; IDs ascend within
  each file; every ID's domain code matches its file per the schema table.
- Controlled vocabularies: all priorities in {P0, P1, P2}; all statuses in
  {proposed, accepted, blocked, retired}; all actor names valid.
- Banned reference-platform brand names: none found (checked: VMware /
  vSphere / vCenter, AWS / Amazon Web Services, Azure, GCP / Google Cloud,
  Alibaba Cloud / Aliyun, Yandex Cloud, Selectel, VK Cloud, OpenStack,
  Huawei Cloud, Tencent Cloud, Oracle Cloud, IBM Cloud, DigitalOcean,
  Hetzner, OVH, Hyper-V, Proxmox). Open-source ecosystem building blocks
  named as implementation classes (Kubernetes, KubeVirt, etcd, and similar)
  are not reference platforms and are not flagged.
- Traceability: every block names at least one known provenance class.
- Cross-references: 584 cross-references to 231 unique requirement IDs were
  extracted from requirement blocks; **every referenced ID exists**.
- Non-English text: no Cyrillic, CJK, or other non-Latin scripts anywhere
  in the corpus (see the single non-ASCII letter finding below).

## Violations found

### Errors (1)

1. **CR-CMP-020** (`domains/11-compute-virtualization.md:95`) — non-ASCII
   letter `ç` (U+00E7) in the word "façade" inside the Traceability field.
   The English-only rule (CR-FND-100) and the ASCII-machine-parsability
   goal require the strict spelling "facade". Report only — not fixed here.

### Warnings (10)

1. **CR-NET-020** (`domains/12-network.md:89`) — specific third-party
   product identifier "NetBox" in the Traceability field
   ("NetBox-backed adapter"). Not a reference-platform brand, but a named
   product; source-safety rule 2 prefers generic terms in corpus text.
2. **Actor-list separator inconsistency (5 files, 106 blocks)** —
   `domains/13-storage-backup-dr.md`, `domains/16-billing-finops.md`,
   `domains/19-portal-ux-selfservice.md`, `domains/21-ops-sre-support.md`,
   and `domains/22-deployment-iac-cicd.md` separate actor lists with `|`
   while the other 11 files (224 blocks) use `,`. The schema line
   `- **Actors:** provider | vendor | …` enumerates *allowed values*; one
   list separator should be used corpus-wide. The registry parser accepts
   both, so `requirements.json` is unaffected.
3. **`n/a` stop conditions with risk-class words in requirement text
   (4 blocks, heuristic review list)**:
   - **CR-CUX-190** (page-pattern contracts) — requirement text mentions
     cost, data/secrets touched, and migrate/delete paths; stop conditions
     are `n/a`. The most substantive of these four: the requirement is
     about *surfacing* risk classes rather than creating risk impact, but
     owner review should confirm `n/a` is intended.
   - **CR-IAM-240** (cryptographic availability measurement) — mentions
     KMS/secrets-store SLIs; stop conditions explicitly defer key-material
     risk to CR-IAM-130/CR-IAM-140 governance. Likely a correct `n/a`;
     flagged for confirmation.
   - **CR-CUX-170** (governed i18n terminology) — matched "keys" meaning
     canonical terminology keys, not cryptographic keys. False positive,
     listed for completeness.
   - **CR-FND-100** (English-only artifacts) — matched "metadata" in the
     commit-metadata sense. False positive, listed for completeness.
   The other five `n/a` blocks (CR-FND-210, CR-NET-210, CR-CUX-160,
   CR-CUX-200, CR-DPL-200) contain no risk-class words and are consistent
   with the schema's `n/a` rule.

## Cross-referenced IDs that do not exist

**None.** All 584 cross-references (231 unique IDs) resolve to requirement
blocks present in the corpus.

## Exit-code contract for CI

- `python3 validate.py` — regenerates `registry/requirements.json`, prints
  counts, exits 1 while any error-level finding exists (currently: the
  single non-ASCII letter).
- `python3 validate.py --check` — writes nothing; additionally fails if
  `requirements.json` on disk differs from the corpus (drift gate).
- Warnings are always reported but never fail the run.

## Fixes applied

Applied 2026-07-20 after the report above was generated; the report body is
kept as the record of findings. `python3 validate.py` does not regenerate
this file, so fixes are recorded here and verified by a fresh validator run.

1. **CR-CMP-020 non-ASCII letter (error)** — fixed: "façade" replaced with
   the strict ASCII spelling "facade" in
   `domains/11-compute-virtualization.md` (Traceability field).
2. **CR-NET-020 third-party product identifier (warning)** — fixed:
   "NetBox-backed adapter" replaced with the generic provenance-safe phrase
   "external-IPAM adapter" in `domains/12-network.md` (Traceability field).
   No new provenance class was introduced; the existing class
   `legacy-platform-a` remains the source.
3. **Actor-list separator inconsistency (warning, 5 files, 106 blocks)** —
   fixed: actor lists in `domains/13-storage-backup-dr.md` (23),
   `domains/16-billing-finops.md` (21), `domains/19-portal-ux-selfservice.md`
   (21), `domains/21-ops-sre-support.md` (19), and
   `domains/22-deployment-iac-cicd.md` (22) now use `, ` as the separator,
   matching the other 11 files. No actor names were changed.
4. **`n/a` stop-condition review (warning, 4 blocks)** — resolved:
   - **CR-CUX-190** — genuinely touches the trust and deletion risk classes
     (the contract governs how cost, data/secrets touched, and
     stop/rollback/export/migrate/delete paths are surfaced to humans and
     agents); `n/a` replaced with a real stop condition halting module
     registration on any misstated action map.
   - **CR-IAM-240** — kept `n/a`: measurement-only surface; key-material
     risk is explicitly governed by CR-IAM-130 and CR-IAM-140. Confirmed
     correct.
   - **CR-CUX-170** — kept `n/a`: heuristic matched "keys" meaning
     canonical terminology keys, not cryptographic keys. False positive.
   - **CR-FND-100** — kept `n/a`: heuristic matched "metadata" in the
     commit-metadata sense, not a data risk class. False positive.

Post-fix validator run (`python3 validate.py`, exit 0; `--check` drift gate
also clean): **330 requirements parsed, 0 errors, 3 warnings** — the three
remaining warnings are the heuristic `n/a` review notices for CR-FND-100,
CR-IAM-240, and CR-CUX-170 listed above as confirmed-correct `n/a` blocks;
warnings never fail the run. `registry/requirements.json` was regenerated
from the fixed corpus. No requirement IDs, priorities, statuses, or other
normative text were changed.
