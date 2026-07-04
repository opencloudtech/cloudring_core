# GitOps Orchestration Contract

CloudRING Core owns the public GitOps/orchestration handoff contract, not a
live environment overlay. The contract defines how a module enters GitOps,
which controller owns reconciliation for each object class, and which receipts
must exist for dry-run, apply, and rollback decisions.

## Boundary

Core MAY contain:

- bootstrap handoff metadata for a synthetic Git source and overlay root;
- reconciliation ownership declarations for GitOps, platform controllers, and
  module controllers;
- receipt classes for dry-run, apply, rollback, and drift observations;
- module deployment handoff metadata that links a module package to a portable
  overlay path.

Core MUST NOT contain live provider overlays, private Git remotes, tenant data,
credentials, DNS zones, account identifiers, external secret values, or copied
deployed-environment configuration. Provider-specific GitOps overlays and
environment values stay outside the public core tree.

## Machine Contract

`gitops-orchestration.schema.json` defines the portable document shape.
`fixtures/synthetic-module-handoff.json` is the happy-path synthetic module
handoff fixture. The fixture points at `fixtures/synthetic-overlay`, which is a
provider-neutral kustomize overlay for validation and documentation examples.

The source-safety gate for `cloudring_core` validates that this contract remains
public-core material. A copied live provider overlay path under this directory
is a blocker even if the copied file content is otherwise empty.
