# Governance

CloudRING governance keeps the public platform core small, portable, and
safe for downstream companies to adopt.

## Platform ownership

Core maintainers own public contracts, compatibility policy, validation rules,
release gates, module registry behavior, evidence semantics, and public
developer documentation. Changes should preserve implementation neutrality.

## Service ownership

Service module maintainers own their own module manifests, controllers,
adapters, portal extensions, billing connectors, support diagnostics, evidence,
and lifecycle behavior. Core maintainers review whether those modules satisfy
public contracts.

## Enterprise and private boundary

Enterprise modules, private adapters, company overlays, and customer deployment
records are governed by their owning organizations. They are not part of
CloudRING unless intentionally contributed as public, source-safe material.

## Developer entry points

Compatibility discussions start from OCSv3 contracts, provider adapter
interfaces, IAM and policy decisions, portal shell slots, and evidence/readiness
requirements. Decisions should be recorded in public docs or machine-readable
contracts when they affect downstream implementers.
