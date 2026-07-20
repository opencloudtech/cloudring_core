# Matrix — Risk Classes × Domains

Generated: 2026-07-20 UTC.
Scope: the 8 risk classes defined by `01-requirement-schema.md` (money, data,
keys, trust, exposure, deletion, migration, settlement) against the 16 domain
files, computed from the `risk_classes` field of every record in
`registry/requirements.json` (330 requirements; zero drift at time of writing).

## Method

- A requirement carries a risk class when the class is listed in its registry
  `risk_classes` field, which is generated from the Stop conditions field of
  the requirement block.
- One requirement may carry several classes; class assignments therefore sum
  to more than the requirement count (671 assignments over 330 requirements).
- 26 requirements carry no risk class and legitimately declare `n/a` stop
  conditions per the schema rule (stop conditions are mandatory only for
  requirements with risk-class impact). These are counted in the
  "no class" column for completeness.
- The generic stop-condition patterns below are synthesized from the stop
  conditions of the requirements carrying each class; the cited IDs are
  representative, not exhaustive.

## Risk-class definitions (recap)

| Class | What is at risk |
|---|---|
| money | Charging, metering, quota, credits, revenue — wrong amounts in either direction |
| data | Tenant or platform data — loss, corruption, divergence, unprovable state |
| keys | Secrets, credentials, tokens, key material — leakage or misuse |
| trust | Evidence integrity, honest states, contract fidelity — claims exceeding proof |
| exposure | Cross-tenant visibility, public attack surface, unreviewed disclosure |
| deletion | Irreversible removal of resources or data |
| migration | Version skew, substrate or standard change, jurisdiction or provider moves |
| settlement | Period close, invoicing, revenue share, federation money movement |

## Count matrix (requirements carrying each class, per domain)

| Domain | money | data | keys | trust | exposure | deletion | migration | settlement | no class | Total reqs |
|---|---|---|---|---|---|---|---|---|---|---|
| FND | 1 | 2 | 2 | 15 | 5 | — | 6 | — | 2 | 21 |
| CMP | 3 | 11 | 4 | 11 | 3 | 3 | 2 | — | — | 18 |
| NET | 9 | 4 | 2 | 8 | 11 | 6 | 4 | 1 | — | 22 |
| STO | 2 | 21 | 4 | 11 | 6 | 6 | 4 | — | — | 23 |
| K8S | 2 | 4 | 2 | 9 | 6 | 2 | 2 | — | — | 18 |
| IAM | 4 | 12 | 13 | 22 | 10 | — | — | — | — | 24 |
| BIL | 21 | 6 | 2 | 10 | 4 | 1 | 1 | 7 | — | 21 |
| OCS | 4 | 8 | 4 | 12 | 7 | 6 | 6 | 2 | — | 20 |
| MKT | 10 | 3 | 3 | 9 | 6 | 2 | — | 3 | — | 19 |
| CUX | 6 | 6 | 3 | 2 | 5 | 4 | — | — | 11 | 21 |
| OBS | 4 | 12 | 6 | 13 | 7 | 3 | — | 2 | — | 23 |
| OPS | 5 | 13 | 2 | 8 | 5 | 5 | 3 | 2 | — | 19 |
| DPL | — | 13 | 9 | 13 | 9 | 4 | 8 | — | 1 | 22 |
| FED | 6 | 11 | 4 | 13 | 10 | 1 | 4 | 4 | — | 18 |
| DAT | 3 | 15 | 5 | 5 | 7 | 4 | 2 | — | — | 20 |
| AGT | — | 6 | — | 4 | — | 1 | 1 | 1 | 12 | 21 |
| **Total** | **80** | **147** | **65** | **165** | **101** | **48** | **43** | **22** | **26** | **330** |

Of which P0: money 39, data 83, keys 39, trust 85, exposure 57, deletion 28,
migration 28, settlement 6.

## Generic stop-condition pattern per class

Each pattern is the corpus-wide shape of "when to halt and escalate" for the
class. Individual requirements bind the pattern to their own concrete triggers;
the schema requires the binding, this matrix fixes the pattern.

### money — halt the value path; never guess amounts

Halt acceptance, creation, or publication of the affected meter, quota,
price, or charge and escalate when: usage attribution is missing or
duplicated; consumption cannot be proven for an interval (unproven intervals
are not billed); unit, meter, or rate-card drift appears between declaration
and transmission; replay double-counts; quota state is unavailable (fail
closed, never admit unreserved); or reconciliation diverges beyond tolerance.
Suspend the affected meters or reservations until reconciled. Representative:
CR-BIL-110, CR-BIL-140, CR-CMP-160, CR-OCS-060, CR-NET-060, CR-NET-090.

### data — halt the mutation; reconcile before proceeding

Halt the operation and escalate when: driver-observed state contradicts
control-plane records (mark affected objects UNKNOWN and block destructive
actions until reconciled); a task or operation store shows corruption, replay
ambiguity, or idempotency-key conflict with differing payloads (halt admission
of new work on the affected shard — never guess which side effects already
ran); backup or export evidence for the affected data classes is absent,
stale, or failed before a data-bearing mutation; or a data-bearing state is
misrepresented to users. Representative: CR-CMP-010, CR-CMP-020, CR-CMP-030,
CR-CMP-040, CR-OCS-030, CR-OCS-040, CR-STO-080.

### keys — treat leakage as incident; rotate before continuing

Halt and escalate immediately on: raw secret material detected in any
package, fixture, configuration, repository, evidence bundle, log, or export
(revoke and re-issue the affected material before continuing); identity-backend
inconsistency, signing-key rotation incidents, or suspected token replay
(halt issuance for the affected scope); suspected cross-tenant metadata or
credential exposure (disable the affected path and page security); repeated
management authentication failure (quarantine the node); or any context where
workload identity cannot be attested (halt enablement). Representative:
CR-OCS-130, CR-FND-160, CR-CMP-140, CR-CMP-150, CR-CMP-060, CR-IAM-140.

### trust — blocked stays blocked; never launder evidence

Halt promotion, publication, registration, or readiness claims and escalate
when: evidence is missing, stale, blocked, synthetic-only, or redacted for a
claim of production-grade readiness; service-specific or provider-specific wiring is
discovered inside core artifacts; declared capabilities or states misrepresent
reality (for example reporting deletion complete while data persists); a
verifier accepts an invalid input (false negative); receipts show mutation or
backdating; or ownership/identity cannot be verified. Keep the item
unpublished and the blocked state visible until resolved. Representative:
CR-FND-130, CR-FND-060, CR-OCS-020, CR-OCS-040, CR-OCS-090, CR-OCS-150.

### exposure — deny by default; hide rather than permissively show

Halt and escalate on: any path allowing tenant-to-tenant traffic or
cross-tenant visibility without explicit configuration; any public exposure,
hidden cost, or hidden default introduced before explicit user commit;
integrity-verification failure or sandbox violation in third-party UI (refuse
the mount — fail closed to hidden, never permissive); unsigned or tampered
artifacts (refuse use at consumption time); or any source-safety finding on
merge, push, export, or publication — never proceed by annotation or waiver on
exposure-class findings without an explicit owner decision. Representative:
CR-NET-010, CR-OCS-070, CR-FND-050, CR-FND-170, CR-CMP-070, CR-FND-120.

### deletion — refuse irreversibility under uncertainty

Halt the delete path and escalate when: dependent or attached resources exist
whose deletion was not explicitly requested; policy-required backup or export
evidence is missing or stale; observed state contradicts records; a deletion
batch exceeds its safety threshold (require explicit operator approval); or
the standard confirmation flow has not completed. Reclamation and garbage
collection halt on journal corruption or reference-count drift — never
reclaim under uncertainty. Representative: CR-CMP-010, CR-CMP-090, CR-CMP-170,
CR-NET-010, CR-NET-060, CR-OCS-030, CR-OCS-140.

### migration — no breaking move without a proven path

Halt the rollout, promotion, or mutation and escalate when: a breaking
standard or schema change lacks published migration guidance and a
dual-validation window; a module update's compatibility window excludes the
target platform version; readiness evidence was produced on a legacy or
non-target substrate (such evidence never promotes); a compatibility shim or
re-labeling of a legacy stand as production is proposed (escalate to owner
decision); data, backups, logs, or telemetry would cross a jurisdiction
boundary without explicit policy approval; or a component lacks a declared
exit path. Representative: CR-FND-020, CR-FND-030, CR-FND-070, CR-FND-180,
CR-OCS-120, CR-K8S-060.

### settlement — blocked reconciliation is a first-class state

Halt invoicing, credit issuance, document issuance, payout, or any external
revenue claim and escalate when: source ledger and document totals diverge;
ledger conservation checks fail; re-rating a historical window yields
different charges than the originals; a period shows divergence above
tolerance (the period blocks until dispositioned); any path deletes data
before its declared retention window during suspension or collection; or a
mutated financial document would be reissued under the same number (never —
issue a correcting document). Blocked reconciliation MUST NOT be converted
into a release claim. Representative: CR-BIL-050, CR-BIL-080, CR-BIL-100,
CR-BIL-120, CR-BIL-140, CR-OCS-190.

## Observations

- **trust (165) and data (147)** are the broadest classes: evidence honesty and
  data safety touch nearly every domain — consistent with the charter's
  evidence-over-claims and stateful-is-first-class principles.
- **money concentration** is where it should be: BIL (21), MKT (10), NET (9)
  carry the densest money-classed requirements; FND and DPL carry almost none
  by design (install machinery must not create charges).
- **keys concentration**: IAM (13) and DPL (9) dominate — identity issuance
  and CI/IaC are the two places secret material historically leaks; both are
  fail-closed by requirement.
- **settlement (22)** is deliberately small and P1-heavy (only 6 P0): the OSS
  corpus fixes the honesty gates around period close and revenue claims while
  settlement execution itself remains a CloudRING Business concern.
- **IAM carries zero deletion-classed requirements** while CUX carries zero
  migration-classed ones — checked against the registry; this reflects the
  current corpus, not a rule. AGT concentrates its risk expression in the
  approval/evidence model rather than per-class stop conditions (12 of 21 AGT
  requirements carry no class).
- Stop-condition discipline is corpus-wide: every requirement carrying a risk
  class names that class in its Stop conditions field, and the 26 requirements
  with an empty registry `risk_classes` set are exactly the "no class" column
  above (FND 2, CUX 11, DPL 1, AGT 12). Three heuristic quirks are recorded
  honestly rather than corrected here: CR-NET-210, CR-IAM-240, and CR-CUX-160
  declare `n/a` stop text that mentions class names while deferring to other
  requirements, so the registry generator (which keyword-matches the stop
  text, see `registry/validate.py`) assigns them classes — CR-NET-210 carries
  all eight. Counts in this matrix reflect the registry as generated; the
  same three blocks already appear in the review list of
  `registry/validation-report.md`.
