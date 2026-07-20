# SC-08 — Tenant exit and data portability

End-to-end acceptance scenario for the CloudRING OSS requirements corpus.
Requirement IDs reference `domains/*.md`; scenario discipline follows
`00-product-charter.md` and `01-requirement-schema.md`.

## Purpose

Prove the anti-lock-in promise: a tenant exits the platform (or moves to
another provider's installation) with a complete, verifiable export of
data, configuration, and commercial records; honest disclosure of what
remains and for how long; deletion that follows declared timelines with
evidence; and a completeness-and-residency evidence bundle both sides can
audit. Exit is a product capability, not an obstruction.

## Actors

- tenant — exiting party
- provider — source installation operator
- receiving provider — target installation operator (when moving)
- auditor — verifies exit evidence and deletion timelines

## Preconditions

- The tenant has active resources, billing history, and IAM subjects on
  the source installation.
- Jurisdiction and residency attributes are recorded and queryable for
  the tenant's workloads, data sets, backups, identity, and billing
  records (CR-FED-090).
- The tenant's billing account is settled or has a declared arrears/
  retention timeline (CR-BIL-100).

## Steps

1. **Request exit and receive the disclosure.** Before any confirmation,
   the console presents the full deletion/exit disclosure.
   - **Expected outcome:** dependents and incompatible items, residual
     data and retained backups with retention periods, billing stop time,
     recovery window and appeal path, and the export/portability path
     including non-exportable boundaries are all shown; disclosure happens
     before confirmation, never after.
   - **Requirements:** CR-CUX-100

2. **Review jurisdiction impact.** Any cross-jurisdiction movement in the
   exit plan is surfaced before execution.
   - **Expected outcome:** jurisdiction attributes are queryable; actions
     moving data or metadata across a boundary show their jurisdiction
     impact up front and fail closed when policy disallows.
   - **Requirements:** CR-FED-090, CR-FND-070

3. **Choose the export scope and format.** The tenant selects what to
   export; each capability contract declares what is exportable, in what
   format, and which portability gaps exist.
   - **Expected outcome:** the exit flow states scope, format/protocol
     class, cost, duration estimate, and policy checks; non-exportable
     boundaries are disclosed, not hidden.
   - **Requirements:** CR-CUX-100, CR-FND-070

4. **Export commercial records.** Invoices, billing documents, and usage
   history are exported.
   - **Expected outcome:** documents are exportable by the tenant;
     document integrity (hash and version per issued document) travels
     with the export; the final invoice and billing stop time are
     unambiguous.
   - **Requirements:** CR-BIL-120, CR-BIL-100

5. **Build the migration package (provider-to-provider case).** A
   portable package of data, configuration, policies, identity mappings,
   billing/usage history, and recovery instructions is produced.
   - **Expected outcome:** the package is integrity- and completeness-
     checkable; the receiving provider verifies it before import; a
     dry-run compatibility report (capability gaps, quota differences,
     version windows) precedes any mutation.
   - **Requirements:** CR-FED-080

6. **Confirm the backup barrier before deletion.** Deletion of stateful
   resources is gated on fresh, signed backup-barrier receipts.
   - **Expected outcome:** no stateful resource is deleted without a
     barrier receipt inside the freshness window; a missing or invalid
     barrier halts the deletion and records a `blocked` state.
   - **Requirements:** CR-STO-090

7. **Execute staged deletion.** Deletion proceeds through staged barriers:
   export window first, purge and metadata barriers second.
   - **Expected outcome:** staged deletion follows the declared order;
     lifecycle ambiguity (unknown or disputed owner) fails closed to
     retention, never to deletion.
   - **Requirements:** CR-DAT-060, CR-STO-120

8. **Honor the retention timeline.** If the account was in arrears or
   under policy retention, the declared sequence applies.
   - **Expected outcome:** resources stop but are not deleted before the
     grace period; data-retention countdown with a declared minimum
     window precedes any deletion; every step is event-sourced, reversible
     on settlement, and visible to the tenant.
   - **Requirements:** CR-BIL-100

9. **Hold source-side deletion until confirmation.** In the provider-to-
   provider case, the source provider does not delete until the tenant
   confirms success or the declared retention window expires.
   - **Expected outcome:** source-side deletion without tenant
     confirmation or retention-window expiry is forbidden; when deletion
     finally executes, it produces evidence.
   - **Requirements:** CR-FED-080

10. **Revoke identity completely.** Tenant subjects, groups, and service
    accounts are removed from the resource hierarchy.
    - **Expected outcome:** container-membership join semantics guarantee
      full revocation at the root; every effective access inside the
      tenant's containers is visible and revoked; revocation is auditable.
    - **Requirements:** CR-IAM-050, CR-IAM-010

11. **Produce the exit evidence bundle.** Completeness-and-residency
    evidence is assembled: what moved, what remained, what was deleted,
    with verification proof.
    - **Expected outcome:** the bundle checker validates completeness
      across data, configs, policies, audit summary, billing, and recovery
      instructions; residency attributes are correct at the destination;
      the bundle is party-scoped and redacted.
    - **Requirements:** CR-FED-090, CR-CUX-110

12. **Close with audit and post-deletion view.** Final audit records are
    written; the tenant sees the post-deletion view.
    - **Expected outcome:** the post-deletion view shows what remains and
      until when (deleting → deleted states); the immutable audit log
      records the full exit sequence with actor, scope, and timestamps.
    - **Requirements:** CR-CUX-100, CR-CUX-020, CR-IAM-150

13. **Exercise the fail-closed paths.** Attempts to skip disclosure,
    delete inside the retention window, or transfer across a denied
    jurisdiction boundary are made in test.
    - **Expected outcome:** each attempt fails closed with a typed denial
      naming the missing precondition; halt-and-escalate fires when
      dependent billing, retained-backup, or recovery-window data is
      unavailable at confirmation time.
    - **Requirements:** CR-CUX-100, CR-BIL-100, CR-FED-090

## Requirement coverage

| Requirement | Title | Steps |
|---|---|---|
| CR-CUX-020 | Canonical resource state vocabulary | 12 |
| CR-CUX-100 | Deletion and exit semantics on every resource | 1, 3, 12, 13 |
| CR-CUX-110 | Operation evidence bundles, party-scoped and redacted | 11 |
| CR-FND-070 | Portability and jurisdiction freedom by design | 2, 3 |
| CR-FED-080 | Tenant migration between providers | 5, 9 |
| CR-FED-090 | Jurisdiction portability and exit evidence | 2, 11, 13 |
| CR-BIL-100 | Postpaid credit limit, arrears, suspension, retention timeline | 4, 8, 13 |
| CR-BIL-120 | Invoices and billing documents (EDO-class) | 4 |
| CR-STO-090 | Signed backup barrier before any mutation | 6 |
| CR-STO-120 | Tenant and namespace storage isolation | 7 |
| CR-DAT-060 | Staged deletion with purge and metadata barriers | 7 |
| CR-IAM-010 | Hierarchical resource model with inherited bindings | 10 |
| CR-IAM-050 | Container membership join semantics guaranteeing full revocation | 10 |
| CR-IAM-150 | Immutable, queryable, SLA-bound audit log | 12 |

## Gaps found

- **Marketplace entitlement termination at exit.** CR-MKT-120 governs
  license *expiry* semantics, but no requirement explicitly governs the
  termination, transfer, or refund handling of active marketplace
  entitlements and subscriptions when a tenant exits. A requirement is
  needed covering entitlement revocation/transfer on tenant exit and its
  billing cut-off evidence.
- **Per-primitive tested exit packages.** CR-FND-070 requires every
  capability contract to declare exportability, but its own non-claims
  record that tested exit per primitive is not yet proven; there is no
  requirement mandating a per-primitive exit-package *conformance test*
  suite. The scenario accepts contract-level declaration plus drill
  evidence as intermediate proof.

## Evidence required

- e2e deletion/exit journey evidence asserting pre-confirmation disclosure
  content (CR-CUX-100).
- Exit-evidence bundle and checker output: completeness across data,
  configs, policies, audit summary, billing, recovery instructions
  (CR-FED-090).
- Migration package integrity/completeness checker output and dry-run
  compatibility report, when moving providers (CR-FED-080).
- Signed backup-barrier receipts for all deleted stateful resources
  (CR-STO-090).
- Staged-deletion execution log with export window, purge, and metadata
  barrier records (CR-DAT-060).
- Retention-timeline event log with timestamps proving no deletion before
  the declared window (CR-BIL-100).
- Commercial record export: documents with integrity hashes and the final
  invoice (CR-BIL-120).
- IAM revocation audit query proving complete removal of tenant access
  (CR-IAM-050, CR-IAM-150).
- Typed denials from the fail-closed path tests (CR-CUX-100, CR-FED-090).
