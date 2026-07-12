# What Is OCSv3

OCSv3, Open Cloud Standard 3, is a public contract for cloud service modules
that run above upstream Kubernetes APIs. Kubernetes describes resources and
controllers; OCSv3 describes how a cloud platform discovers a service, checks
IAM, renders a portal extension, connects billing, records evidence, supports
users, and safely handles lifecycle operations.

Compared with Open Service Broker, OCSv3 covers more than provision and bind:
it includes portal, billing, IAM, evidence, support, analytics, and rollback.

Use OCSv3 when a service must be added to CloudRING or another compatible
platform without hardcoding that service into the platform core.

## Why Not Only Kubernetes API

Kubernetes is the substrate. It does not define product catalog metadata,
tenant entitlements, billing meters, support diagnostics, product analytics,
portal microfrontends, or evidence required before a readiness claim. OCSv3
adds those portable cloud-product surfaces while keeping the workload contract
Kubernetes-native.

## Minimum Module

A module package must include:

- service connector and Kubernetes bindings;
- billing connector or an explicit non-billable profile;
- portal extension manifest;
- IAM and tenant access surfaces;
- support diagnostics;
- evidence and readiness surfaces;
- lifecycle, data lifecycle, durability, and state model;
- distribution, federation, and commercial metadata;
- source-safety non-claims.

Validate a module with:

```bash
go run ./cmd/ocsctl validate ./reference/synthetic-service/module-package.json
go run ./cmd/ocsctl conformance ./reference/synthetic-service/module-package.json
```
