# Portal Shell Contract

CloudRING Core owns the generic portal shell contract. The shell creates the
authenticated workspace frame and renders service modules from OCSv3 connector
metadata. Service modules own their domain screens, actions, lifecycle APIs,
and implementation code behind the connector boundary.

## Core Shell Responsibilities

The core shell contract may define:

- Identity and session state, including authentication status, selected tenant,
  selected project, management visibility, and session assurance references.
- Organization and project switcher state with visible organizations, tenants,
  teams, projects, namespaces, and selected entries.
- Navigation groups, stable labels, placement, route references, and empty,
  loading, degraded, or denied states.
- Permission checks, policy decision references, required actions, and audit-safe
  denial metadata.
- Module slots loaded from OCSv3 connector metadata, including slot name, route
  owner, host reference, mount reference, and API reference.
- Microfrontend mount references with host authority, runtime reference,
  scoped context fields, and certification evidence links.
- Analytics event references for shell navigation, module mount lifecycle,
  support diagnostics, and product funnels.
- Audit and evidence links that point to decision receipts, freshness metadata,
  support evidence, readiness checks, and non-claim records.

The core shell contract must not define service action handlers, concrete
service catalogs, service lifecycle decisions, billing adapters, or module
implementation routes. Those surfaces are provided by connector packages and
are mounted through declared module slots.

## OCSv3 Metadata Flow

The shell reads catalog entries, module slots, host references, mount
references, permissions, analytics references, readiness checks, durability
evidence, and support links from OCSv3 connector package metadata. A shell
renderer can decide placement and authorization, but it must not embed a fixed
service catalog or a service-specific action route in core material.

The fixture in `fixtures/synthetic-portal-shell.json` uses symbolic IDs only.
It demonstrates a generic service detail slot mapped from connector metadata
without naming a concrete service implementation.
