# G11 — Unified human and agent experience

## Outcome

Deliver one coherent provider experience for customers, administrators,
developers, automation and AI agents. Every supported action has one public
semantic contract and consistent authorization, validation, operations and audit.

All reusable implementation must merge into public OSS and the exact accepted
change must be deployed and live-verified at `https://hub.cloudring.org`.

## Scope

- Complete the provider portal information architecture: tenant/project switch,
  product catalog, resources, operations, access, costs, support, audit and admin.
- Complete public API and CLI coverage for every G03-G10 capability. Generate
  clients and reference docs from the running contract; add shell completion and
  machine-readable output.
- Implement product navigation and extensions from G08 with accessible loading,
  empty, degraded, error, retry and permission-denied states.
- Add safe automation primitives: discovery, dry-run/plan, idempotency, approval
  references, operation polling/events, rate-limit feedback and resumable tasks.
- Publish an optional MCP-compatible adapter generated over the same public API.
  It uses scoped short-lived credentials and does not become a core dependency.
- Provide explicit destructive-action previews, cost/quota impact, rollback
  information and confirmation.
- Instrument frontend and CLI failures with privacy-safe correlation to backend
  traces and audit.
- Meet WCAG-oriented accessibility, keyboard navigation, localization-ready text,
  responsive layout and stable browser support.

## Required journeys

- perform the same tenant, project, policy, package, order and billing operations
  through portal, CLI, API and agent adapter and compare resulting state/audit;
- resume an interrupted CLI/agent operation without duplication;
- hide or disable inaccessible actions and still reject direct unauthorized calls;
- handle empty provider, degraded connector, stale client version, rate limiting,
  validation errors and partial read-model outage honestly;
- run browser accessibility and visual regression plus API/CLI compatibility tests.

## Hub and downstream delivery

Roll out the full shell and control surfaces to `hub.cloudring.org`. Run sanitized
customer, provider-admin, product-developer and scoped-agent journeys with test
tenants, then clean up. Brand/theme differences remain downstream configuration;
workflow code remains OSS.

## Acceptance

- Public issue #31 and portal portions of Task 24 are complete for capabilities
  delivered so far.
- No fixture-only, static or hidden private workflow remains in a claimed path.
- An API-only provider remains fully operable, and no UI-only capability exists.
- Agent actions are no more privileged than their identity and are completely
  attributable and revocable.
