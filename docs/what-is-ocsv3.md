# What Is OCSv3

OCSv3, Open Cloud Standard 3, is a public contract for cloud products regardless
of where or how their implementation runs. Kubernetes may describe local resources
and controllers; OCSv3 describes how a cloud platform discovers a product, checks
IAM, exposes its API, optionally renders its user experience, connects billing,
records evidence, supports users, and safely handles lifecycle operations.

Compared with Open Service Broker, OCSv3 covers more than provision and bind:
it includes portal, billing, IAM, evidence, support, analytics, and rollback.

Use OCSv3 when a service must be added to CloudRING or another compatible
platform without hardcoding that service into the platform core.

## Why Not Only Kubernetes API

Kubernetes is one supported substrate. It does not define product catalog metadata,
tenant entitlements, billing meters, support diagnostics, product analytics,
portal microfrontends, or evidence required before a readiness claim. OCSv3
adds those portable cloud-product surfaces without requiring a remote or API-only
product to adopt Kubernetes internally.

## Execution profiles

| Profile | Required connection | Profile-specific surfaces |
| --- | --- | --- |
| `local` | Versioned public product API inside the provider trust boundary | Kubernetes bindings are allowed when the implementation uses Kubernetes. |
| `remote` | Versioned public product API plus remote endpoint, workload identity, trust, health, and retry contract | No local Kubernetes binding is required. |
| `api-only` | Versioned public product API plus endpoint, workload identity, trust, health, and retry references | No Kubernetes binding or microfrontend is required. |

A signed, integrity-pinned, sandboxed microfrontend is optional in every profile.
Federation and commercial metadata are opt-in applicability profiles. Each declares
`supported` or `not_applicable` with a reason; complete metadata is validated only
when supported. Remote and API-only products declare source-safe endpoint,
trust-policy, and health references, never raw endpoints or credentials.

## Minimum Module

A module package must include:

- one execution profile and a versioned public product API;
- billing connector or an explicit non-billable profile;
- explicit applicability for `provision`, `hold` or `suspend`, `resume`, `resize`,
  and `deprovision`;
- IAM and tenant access surfaces;
- support diagnostics;
- evidence and readiness surfaces;
- lifecycle, data lifecycle, durability, and state model;
- distribution metadata and explicit federation/commercial applicability, with
  full metadata only when supported;
- source-safety non-claims.

Kubernetes bindings and a portal extension manifest are optional profile surfaces.
When declared, they must pass their complete conformance and security contracts.

Validate a module with:

```bash
go run ./cmd/ocsctl validate ./reference/synthetic-service/module-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
```
