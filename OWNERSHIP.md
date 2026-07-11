# Ownership

CloudRING separates the reusable public platform from downstream products,
deployments, and independently owned modules.

Copyright is stated in `NOTICE`. Project ownership and trademarks are held by
Elena Trukhina ZZP.

## Platform ownership

CloudRING owns:

- The provider-neutral control-plane runtime and public APIs.
- OCSv3 contracts, SDKs, validators, and conformance tooling.
- Identity, IAM, policy, admission, audit, catalog, billing, portal, and
  self-service platform services.
- GitOps, bootstrap, evidence, readiness, lifecycle, BOM, backup/restore,
  upgrade, and rollback tooling.
- Open source service modules and reusable provider adapters accepted under
  Apache-2.0.
- Developer and operator documentation.

## Service ownership

An independently developed service module owns its implementation and license
unless its owner contributes it to CloudRING. CloudRING-distributed modules are
part of the public project. Both use the same OCSv3 integration surfaces so a
module can be developed and operated without coupling its internals to the
platform.

## Enterprise and private boundary

Company overlays, enterprise-only modules, customer deployments, live values,
credentials, customer data, and live support/evidence records remain outside
the public core. A provider adapter belongs in public when it is reusable and
source-safe; account-, topology-, or company-specific configuration remains
downstream.

## Developer entry points

Module developers publish OCSv3 metadata and APIs. Platform developers extend
the public runtime and contracts. Adapter developers implement infrastructure
integrations behind portable interfaces. Portal developers mount service
experiences through shell extension contracts.
