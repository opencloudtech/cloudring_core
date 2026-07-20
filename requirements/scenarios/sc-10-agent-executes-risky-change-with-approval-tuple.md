# SC-10 — Agent executes a risky change with an approval tuple

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the agent-governance core loop end to end: an autonomous agent
plans a change whose effect classifies as risky-change, assembles the
evidence its rung demands, obtains a scoped, expiring, consume-once
approval tuple from a distinct human approver, executes exactly that
approved plan with brokered secrets, and lands every artifact in the
append-only journal — while replay, plan-mutation, self-approval, and
secret-in-context attacks are all denied closed.

## Actors

- agent — plans and executes the change
- operator / approver — human authority, distinct from the agent
- provider — owns the installation and policy
- auditor — replays the journal and verifies tuple consumption

## Preconditions

- The agent runs under a least-privilege automation account (CR-AGT-150).
- The governance policy (risk classes, evidence ladder, approval matrix)
  is loaded and version-controlled (CR-AGT-010, CR-AGT-020, CR-AGT-200).
- The target stateful resource has a fresh backup, enabling barrier
  issuance where the change is destructive-adjacent (CR-STO-090).
- The operations journal is writable and restart-safe (CR-AGT-090).

## Steps

1. **Receive the goal and export runtime context.** The agent is
   delegated an operational goal; the platform exports the machine-
   readable context the agent may act upon.
   - **Expected outcome:** context export follows the export contract;
     documentation is available in machine-readable form; the agent's
     authority derives from the actor authority matrix, never from
     identity labels.
   - **Requirements:** CR-AGT-130, CR-AGT-140, CR-AGT-080

2. **Plan in separated phases.** The agent produces a plan with
   plan/apply/validate/compensate phases.
   - **Expected outcome:** the plan is explicit and digestible; the
     compensation path is declared before execution; the plan references
     the risk classes it touches.
   - **Requirements:** CR-AGT-110

3. **Classify the action by effect.** The classifier assigns the change
   to risky-change on the relevant effect classes (for example data,
   money).
   - **Expected outcome:** classification accounts for environment,
     target scope, and blast radius; the class renders identically across
     API, CLI, portal, audit, and the agent plan view; the agent cannot
     downgrade or influence its own classification.
   - **Requirements:** CR-AGT-010, CR-OPS-020

4. **Assemble the evidence rungs.** The agent attaches the artifacts its
   rung demands: plan, impact assessment, named owner, validation,
   rollback path, and monitoring window — plus all lower-rung artifacts.
   - **Expected outcome:** the action cannot execute while any required
      rung artifact is missing, stale, or contradictory; stale artifacts
      fail closed.
   - **Requirements:** CR-AGT-020

5. **Obtain brokered secret capabilities.** Any secret the change needs
   is requested from the broker.
   - **Expected outcome:** the broker issues scoped, short-lived
     capabilities bound to action, target, and expiry; the agent sees
     metadata and redacted references only — never values; a secret-like
     value detected in context triggers redaction, stop, and a journaled
     security event.
   - **Requirements:** CR-AGT-050

6. **Request approval from a distinct approver.** The agent submits the
   plan and evidence for approval.
   - **Expected outcome:** approver identity is distinct from requester
     identity; the agent cannot approve, escalate, or widen its own
     authority; approval of the action grants no access to secrets,
     tenant data, or unrelated evidence.
   - **Requirements:** CR-AGT-040, CR-AGT-030

7. **Receive the consume-once approval tuple.** The granted approval
   materializes as an anti-replay tuple: approver, requester, action
   class, target scope, plan digest, expiry, nonce.
   - **Expected outcome:** the tuple is scoped, expiring, and fail-closed;
      stale, expired, ambiguous, or unverifiable approvals deny; tuple
      state is inspectable by requester and auditor at any time.
   - **Requirements:** CR-AGT-040, CR-AGT-030

8. **Pass the policy gates.** Money, data, keys, trust, and jurisdiction
   gates evaluate the approved plan.
   - **Expected outcome:** each gate's decision is recorded; a gate
     denial halts execution with a typed reason naming the missing
     precondition.
   - **Requirements:** CR-AGT-160

9. **Verify the backup barrier, if destructive-adjacent.** Where the
   change mutates stateful data, a fresh signed barrier receipt is
   attached.
   - **Expected outcome:** the mutation proceeds only with a valid
     barrier receipt inside the freshness window.
   - **Requirements:** CR-STO-090, CR-AGT-020

10. **Execute, consuming the tuple exactly once.** The agent applies the
    approved plan; the tuple is consumed at execution time.
    - **Expected outcome:** execution matches the plan digest; the
      consumed tuple is retained in the journal; validate and compensate
      phases run as planned.
    - **Requirements:** CR-AGT-040, CR-AGT-110

11. **Journal everything.** Goal, classification, evidence artifacts,
    approvals, tuple identity, execution, and outcome land in the
    append-only, restart-safe journal.
    - **Expected outcome:** the action is reconstructible and
      reproducible from the journal alone; every gated execution records
      tuple identity and consumption.
    - **Requirements:** CR-AGT-090, CR-AGT-100

12. **Report the honest terminal state.** The agent reports the outcome
    in the shared vocabulary.
    - **Expected outcome:** success, denial, or stop is reported with
      reason, evidence reference, and remediation; denials fail closed
      and name the missing precondition; nothing ambiguous is reported as
      success.
    - **Requirements:** CR-AGT-120, CR-FND-140

13. **Exercise the attack paths.** Security tests attempt: replaying the
    consumed tuple, mutating the plan after approval, self-issued
    approval, out-of-scope presentation, and planting a canary secret in
    context.
    - **Expected outcome:** replay is rejected; plan-digest mismatch
      invalidates the tuple; self-approvals are refused at the policy
      layer; out-of-scope presentation denies; the canary triggers
      redaction, stop, and a journaled security event.
    - **Requirements:** CR-AGT-040, CR-AGT-050, CR-AGT-060

14. **Stop and escalate on ambiguity.** A mid-run condition the agent
    cannot resolve (contradictory evidence, unverifiable approval state)
    is injected.
    - **Expected outcome:** the agent stops and escalates rather than
      guessing; the stop and its reason are journaled; attribution and
      explanation are available to the auditor.
    - **Requirements:** CR-AGT-060, CR-AGT-190

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-AGT-010 | Effect-based risk classification of agent actions | 3 |
| CR-AGT-020 | Evidence ladder per risk class | 4, 9 |
| CR-AGT-030 | Approval lifecycle: scoped, expiring, fail-closed | 6, 7 |
| CR-AGT-040 | Non-self-escalation and consume-once approval tuples | 6, 7, 10, 13 |
| CR-AGT-050 | Brokered secrets for agents, never in context | 5, 13 |
| CR-AGT-060 | Mandatory stop conditions and escalation | 13, 14 |
| CR-AGT-080 | Actor authority matrix (human, agent, service) | 1 |
| CR-AGT-090 | Append-only, restart-safe operations journal | 11 |
| CR-AGT-100 | Action reproducibility from the journal | 11 |
| CR-AGT-110 | Phase-separated execution: plan, apply, validate, compensate | 2, 10 |
| CR-AGT-120 | Fail-closed denials and honest terminal states | 12 |
| CR-AGT-130 | Agent runtime context export contract | 1 |
| CR-AGT-140 | Documentation as machine-readable agent context | 1 |
| CR-AGT-150 | Automation accounts with least privilege | preconditions |
| CR-AGT-160 | Money, data, keys, trust, and jurisdiction policy gates | 8 |
| CR-AGT-190 | Agent identity attribution and explainability | 14 |
| CR-AGT-200 | Governance matrix change control | preconditions |
| CR-OPS-020 | Action risk classes visible before execution | 3 |
| CR-STO-090 | Signed backup barrier before any mutation | 9 |
| CR-FND-140 | One product truth across surfaces | 12 |

## Gaps found

None identified. The full loop — classification, evidence ladder,
approval tuple, brokered secrets, execution, journal, and the attack
paths — maps onto existing agent-governance requirements. (CR-AGT-030 and
CR-AGT-040 record that the approval store and tuple protocol are
undesigned; closing this scenario is the evidence that retires those
non-claims.)

## Evidence required

- Classifier conformance output for the action, with boundary and
  disguise cases (CR-AGT-010).
- The complete evidence-rung artifact set for the executed change
  (CR-AGT-020).
- Approval record and anti-replay tuple with plan digest, expiry, and
  nonce (CR-AGT-030, CR-AGT-040).
- Broker issuance, use, and revocation audit events; zero plaintext
  secret values in any agent-visible artifact (CR-AGT-050).
- Policy-gate decision records (CR-AGT-160).
- Backup-barrier receipt, where the change is destructive-adjacent
  (CR-STO-090).
- Journal extract proving tuple consumption, phase execution, and
  reconstructibility (CR-AGT-090, CR-AGT-100).
- Security test evidence: replay rejected, plan-mutation invalidated,
  self-approval refused, out-of-scope denied, canary-secret stop
  (CR-AGT-040, CR-AGT-050).
- Stop-and-escalate journal record with attribution and explanation
  (CR-AGT-060, CR-AGT-190).
