# Ownership

CloudRING Core separates public platform contracts from independently owned
service and private implementation work.

## Core ownership

CloudRING Core owns:

- OCSv3 registry and validation contracts.
- IAM, policy, and admission interfaces.
- GitOps, bootstrap, evidence, readiness, lifecycle, BOM, and rollback
  abstractions.
- Portal shell contracts and module extension slots.
- Provider adapter interfaces.
- Developer SDK and public-boundary documentation.

## Service ownership

Each service module owns its implementation, API/controller runtime, UI
extension, billing connector, support workflow, durability model, lifecycle
actions, rollback path, delete/export behavior, backup/restore evidence, and
denied/degraded/retry state handling.

## Enterprise and private boundary

Private adapters, company overlays, enterprise-only modules, customer
deployments, live configuration, and support evidence remain outside the public
core. Public core artifacts may describe interfaces for those systems, but not
their private implementation details.

## Developer entry points

Module developers publish OCSv3 metadata. Platform developers extend public
contracts. Adapter developers implement infrastructure integrations behind
interfaces. Portal developers mount service experiences through shell extension
contracts.
