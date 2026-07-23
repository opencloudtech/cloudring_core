# CloudRING Public Specifications

This directory contains stable, provider-neutral requirements referenced by the
public roadmap. Requirements state durable outcomes and acceptance boundaries;
deployment-specific values and live receipts remain in downstream repositories.

- [`goal-01.md`](goal-01.md) preserves the stable legacy Goal 01 requirement
  definitions as compatibility aliases; it is not an active scheduling or
  completion authority.
- [`docs/product-architecture-invariants.md`](../docs/product-architecture-invariants.md)
  is the cross-goal architecture contract.
- [`roadmap/COVERAGE.md`](../roadmap/COVERAGE.md) maps every legacy Goal 01 row
  to one owning G00-G27 goal and canonical `CR-GNN-*` identifier; accepted
  delivery state is stored only under [`roadmap/state/`](../roadmap/state/).

Canonical goal requirements are declared in
[`roadmap/roadmap.yaml`](../roadmap/roadmap.yaml). Any later atomic expansion is
added under its owning goal before work starts, without reactivating Goal 01 or
creating a second status ledger.
