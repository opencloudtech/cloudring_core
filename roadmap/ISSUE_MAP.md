# Issue and requirement ownership

This map prevents G00 from absorbing the whole roadmap and prevents later goals
from closing issues on partial evidence. G00 must refresh issue state, bind each
row to the stable goal requirement IDs already declared in `roadmap.yaml`, and
preserve any issue whose complete acceptance remains open. Narrower requirement
IDs may be added only through a reviewed roadmap change.

## Public provider-control-plane issues

| Issue | Owning goal | Required closure evidence |
| --- | --- | --- |
| #23 Provider control-plane epic | G27 | Every child requirement is delivered or explicitly superseded by a reviewed equivalent. |
| #24 Durable control-plane records | G03 | Crash/restart, concurrency, PostgreSQL failover, migration and event/audit replay proof. |
| #25 Organization hierarchy/sync | G03/G04 | Durable organization graph and directory synchronization including partial failure, tombstones and scale profile. |
| #26 IAM enforcement/audit | G05 | Cross-tenant denial and consistent live enforcement at gateway and executor. |
| #27 Inventory/reconciliation/adoption | G06 | Real discovery, drift, explicit adoption and ownership proof. |
| #28 Safe operation planning/execution | G06 | Idempotent plan/apply/reconcile/rollback against two provider adapters. |
| #29 OCS self-service lifecycle | G09 | Complete lifecycle, retry, compensation and capacity-safe journeys. |
| #30 FinOps/billing | G10 | Usage-to-balanced-ledger-to-invoice trace and replay/correction proof. |
| #31 Portal experience | G11 | Real backend parity across portal/API/CLI and negative authorization paths. |
| #32 Repeatable installation | G01/G02 | Fresh disposable and production-HA clean-clone installs, upgrade and rollback. |
| #33 Backup/recovery/supportability | G18/G22/G23 | Off-cell restore plus integrated failure/recovery and supported operations. |
| #34 Identity/secrets/tenant security | G04/G05/G23/G27 | Runtime identity and isolation first; final adversarial review and fixes last. |
| #35 Complete reference provider/service | G07-G12 | Clean-room provider plus first real OCS Network product, not a synthetic-only demo. |
| #36 SafePush/full conformance | G00 | Positive/negative required gates and healthy isolated signed receipt. |
| #37 Reference deployment | every goal/G27 | Exact accepted goal artifact on hub with cumulative live regression; final release proof at G27. |
| #38 Operator/developer contract | every goal/G27 | Docs match each delivered slice; final independent walkthrough and signed release docs. |

## Current defect issues

| Issue | Owning goal | Notes |
| --- | --- | --- |
| #81, #82 one-server-loss observer/readiness | G23 | G00 may fix only a delivery-gate regression; live continuity closure belongs to G23. |
| #83 executable kubeadm output | G01/G02 | No unresolved secret placeholders in runnable output. |
| #84 OpenBao version/upgrade contract | G02/G23 | Supported-version negotiation and real upgrade/rollback proof. |
| #85 worker/gateway joins | G01/G02 | Real disposable and production joins. |
| #86 secret executor/OpenBao safety | G02/G27 | Runtime fix early; broad final security retest at G27. |
| #87 synthetic service conformance | G07 | Fixed in the OCS release-candidate suite. |
| #88 dead contracts | G00/G07 | Remove already-dead accepted contracts in G00; OCS runtime-consumer proof in G07. |
| #89 production IAM | G05 | No memory-audit or offline-verifier readiness claim. |
| #90 CSI snapshot components | G13/G18 | Volume snapshot lifecycle first; backup integration later. |
| #91 backup proof robustness | G18 | Multi-upload, retry, skew and isolated restore. |
| #92 independent installation docs | G01/G02 | Clean-room execution of exact docs. |
| #93 site schema/topology hard-coding | G01/G02/G06 | Schema-driven validation and provider inventory. |
| #94 source-safety/SafePush/CLA | G00 | Real secret patterns, private endpoints and genuine contribution gate behavior. |
| #95 OCS validator strictness | G07 | Unknown fields, enums, duplicates and references reject. |
| #96 Rook/Ceph CSI deployment | G13 | Fresh production profile provisions and restores real PVCs. |
| #97 kubeadm validation gaps | G01/G02 | No dropped nodes and verified topology/survival inputs. |

## Closure rule

An owning goal comments with the accepted public SHA, regression test, downstream
pins and applicable live evidence before closure. A PR title, contract, fixture or
later roadmap assignment is not closure evidence. If one issue spans two goals,
the first goal records partial delivery and the last owning goal closes it.
