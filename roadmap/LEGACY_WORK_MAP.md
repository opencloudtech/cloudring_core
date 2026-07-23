# Legacy readiness work ownership

This map preserves unfinished obligations from the previous full-readiness goal
without making historical task numbers a second roadmap. G00 refreshes their
actual accepted evidence, creates any missing atomic requirement IDs under the
owning goal, and marks each item `carried`, `superseded-with-proof` or `complete`.
No historical status or dated evidence is accepted without current regression.

| Legacy item | Owning goals | Required disposition |
| --- | --- | --- |
| Task 19 — OCS fixtures, strict parser and source safety | G00, G07 | Preserve registry fixtures, separate negative conformance fixtures, keep validation strict and close source-safety findings rather than suppressing them. |
| Task 20 — release supervisor, credential rotation and GitOps source | G00, G02, G22 | Reprove supported host supervision, credential rotation and exact private-main GitOps reconciliation. Native Windows is required only if the release claims it. |
| Task 21 — runtime secret manager | G02, G22, G23, G27 | Operate replicated secret management, rotation/revocation, backup/recovery and failure-safe consumers; retest in the integrated and final campaigns. |
| Task 22 — object backup and real restore | G17, G18, G23 | Perform off-cell multi-object backup, isolated restore, tenant denial, cleanup and recovery under failure. |
| Task 23 — Gateway/Cilium and direct dual stack | G12, G23 | Prove direct-origin IPv4/IPv6 routing, DNS/TLS, failover and rollback without CDN fallback masking origin failure. |
| Task 24 — continuity and identity denial journeys | G04, G05, G15, G23 | Prove account/IAM/CSRF/break-glass negatives, VM continuity and supported server-loss behavior with sanitized live evidence. |
| Task 25 — promotion verdict | G27 | Produce a fresh fail-closed release verdict; blocked prerequisites remain blockers rather than qualified success. |
| Task 26 — full quality and supply-chain gate | G00, G27 | Run complete tests, race/vet where applicable, tracked/changed source safety, secret scanning, dependency/container/IaC scanning, provenance and boundary review. |
| Task 27 — exact handoff and scope audit | G27 | Record exact accepted files, SHAs, pins, artifacts, live revisions, support handoff, remaining non-goals and cleanup. |
| F1-F5 — final compliance, code, live and scope audits | G27 | Execute all final audit perspectives after functionality is complete, fix release blockers and repeat affected/cumulative proof. Founder review is valid when required checks pass. |

Legacy labels never relax the acceptance of their owning goal. If an old item is
obsolete, its superseding requirement, rationale, accepted SHA and regression
evidence are recorded before it is closed.
