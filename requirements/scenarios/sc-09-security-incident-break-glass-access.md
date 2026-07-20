# SC-09 — Security incident with break-glass access

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the emergency-access contract under a realistic security incident:
normal administrative paths are unavailable or insufficient, an operator
obtains short-lived, ticket-bound break-glass credentials through a named
requester/approver flow, every use alerts and audits, credentials expire
automatically, and the incident closes with a post-incident review —
while the rest of the platform continues to fail closed.

## Actors

- operator — incident responder requesting break-glass access
- provider — security officer / approver
- auditor — verifies the audit chain and drill records
- agent — if involved in containment, acts under emergency containment
  authority (CR-AGT-070)

## Preconditions

- An incident is detected and classified with a severity per the
  classification scheme (CR-OPS-050).
- On-call rotation is staffed with a declared response SLO (CR-OPS-060).
- No standing emergency accounts exist anywhere in the installation
  (CR-IAM-110).
- The immutable audit log is queryable and within its SLA (CR-IAM-150).

## Steps

1. **Detect and classify the incident.** Monitoring and alerting surface
   the event; the responder assigns severity.
   - **Expected outcome:** alerts route through the notification gateway;
     severity classification follows the declared scheme; the incident
     ticket is opened as the system of record.
   - **Requirements:** CR-OBS-100, CR-OPS-050, CR-OPS-120

2. **Page on-call and confirm normal paths fail.** The responder verifies
   that standard administrative or identity paths are unavailable or
   insufficient for containment.
   - **Expected outcome:** on-call acknowledges within the response SLO;
     the failed normal path and the reason are recorded on the ticket;
     authorization everywhere else continues to fail closed.
   - **Requirements:** CR-OPS-060, CR-IAM-160

3. **Request break-glass access.** The operator requests emergency
   credentials bound to the incident ticket.
   - **Expected outcome:** the request names the requester, the approver,
     the scope, and the reason; self-service break-glass without human
     approval is impossible.
   - **Requirements:** CR-IAM-110

4. **Approve and issue short-lived credentials.** A distinct approver
   issues credentials or certificates valid ≤24h, source-restricted where
   applicable.
   - **Expected outcome:** issuance is per-incident, hardware-backed where
     available, and fully audited; any fallback long-lived emergency key
     is sealed, monitored, and covered by a dated revocation drill.
   - **Requirements:** CR-IAM-110, CR-IAM-080

5. **Alert on every use.** Each authentication with break-glass
   credentials triggers an immediate alert.
   - **Expected outcome:** security receives an alert per use; a
     break-glass authentication without a linked incident ticket is itself
     treated as a security incident.
   - **Requirements:** CR-IAM-110, CR-IAM-150

6. **Contain through audited choke points.** The responder performs
   containment actions that cross privilege-escalation choke points.
   - **Expected outcome:** each choke point is a separate, audited
      permission; every authorization decision is logged with actor,
      subject, target, reason, timestamp, and result.
   - **Requirements:** CR-IAM-190, CR-IAM-150

7. **Use duty roles and impersonation where tenant-side access is
   needed.** Any administration inside a tenant scope happens through
   impersonation, not shared credentials.
   - **Expected outcome:** impersonation is ticket-bound, time-boxed,
     attributed to the named operator, and visible to the tenant; the
     operator duty-role taxonomy is enforced.
   - **Requirements:** CR-OPS-140, CR-IAM-200

8. **Contain agent-driven blast radius, if agents are involved.**
   Emergency containment authority is exercised for affected agents.
   - **Expected outcome:** containment actions are journaled with their
     authority basis; retrospective review is mandatory; emergency paths
     never bypass approval-tuple consumption records.
   - **Requirements:** CR-AGT-070, CR-AGT-040

9. **Expire and revoke.** Credentials expire automatically at ≤24h;
   explicit revocation follows containment completion.
   - **Expected outcome:** expired break-glass credentials deny (negative
      test proof); revocation is audited; no credential outlives its
      incident.
   - **Requirements:** CR-IAM-110, CR-IAM-080

10. **Run the post-incident review.** A structured postmortem is produced
    and linked to the ticket.
    - **Expected outcome:** the review covers timeline, decisions,
      contributing factors, and follow-ups; every break-glass use is
      reconciled against ticket scope; follow-up actions are tracked.
    - **Requirements:** CR-OPS-070, CR-IAM-110

11. **Feed the learning loop.** Postmortem findings convert into
    requirement, policy, or runbook changes where warranted.
    - **Expected outcome:** accepted follow-ups are tracked to
      implementation; runbooks-as-code are updated; any governance-matrix
      change goes through change control.
    - **Requirements:** CR-AGT-170, CR-OPS-170, CR-AGT-200

12. **Evidence the drill.** The break-glass issuance-to-expiry path is
    exercised as a drill independent of any real incident.
    - **Expected outcome:** drill evidence shows issuance, alert-on-use,
      expiry denial, and audit query results; drill records are append-
      only and UTC-stamped.
    - **Requirements:** CR-IAM-110, CR-STO-140

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-IAM-080 | Short-lived token model with signed-JWT exchange and signer rotation | 4, 9 |
| CR-IAM-110 | Break-glass emergency access with expiring credentials and full audit | 3, 4, 5, 9, 10, 12 |
| CR-IAM-150 | Immutable, queryable, SLA-bound audit log | 5, 6 |
| CR-IAM-160 | Fail-closed authorization and hidden management surfaces | 2 |
| CR-IAM-190 | Privilege-escalation choke points as separate audited permissions | 6 |
| CR-IAM-200 | Operator duty role taxonomy with impersonation-based administration | 7 |
| CR-OPS-050 | Incident severity classification | 1 |
| CR-OPS-060 | On-call model with response SLO | 2 |
| CR-OPS-070 | Structured postmortems | 10 |
| CR-OPS-120 | Support ticket system of record and SLA metrics | 1 |
| CR-OPS-140 | Audited support impersonation | 7 |
| CR-OPS-170 | Runbooks as code | 11 |
| CR-OBS-100 | Notification gateway and alert ingress | 1 |
| CR-AGT-040 | Non-self-escalation and consume-once approval tuples | 8 |
| CR-AGT-070 | Emergency containment authority | 8 |
| CR-AGT-170 | Learning loop: from postmortem to requirement | 11 |
| CR-AGT-200 | Governance matrix change control | 11 |
| CR-STO-140 | Durability evidence states and freshness | 12 |

## Gaps found

None identified. The incident → break-glass → containment → review chain
maps fully onto the IAM, OPS, and AGT domains.

## Evidence required

- Incident ticket with severity classification and on-call response
  timestamps (CR-OPS-050, CR-OPS-060, CR-OPS-120).
- Break-glass issuance record: named requester and approver, scope,
  expiry ≤24h, per-incident binding (CR-IAM-110).
- Alert records for every break-glass authentication (CR-IAM-110).
- Audit query results showing who/when/why for every emergency action,
  including choke-point crossings (CR-IAM-150, CR-IAM-190).
- Impersonation session records, ticket-bound and time-boxed
  (CR-OPS-140).
- Negative test proof that expired break-glass credentials deny
  (CR-IAM-110).
- Post-incident review record linked to the ticket with reconciled
  break-glass usage (CR-OPS-070, CR-IAM-110).
- Issuance-to-expiry drill evidence (CR-IAM-110).
- Follow-up tracking records from the learning loop (CR-AGT-170).
