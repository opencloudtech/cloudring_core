# Governance

CloudRING governance keeps the public platform cohesive, portable, and safe
for downstream companies to adopt and extend without maintaining incompatible
forks.

## Platform ownership

CloudRING maintainers own the public runtime, APIs, contracts, compatibility
policy, validation rules, release gates, module registry behavior, evidence
semantics, accepted service modules/adapters, and public documentation.
OpenCloudTech decides which contributions are accepted. Platform code should
depend on portable capabilities and OCSv3 interfaces rather than private module
internals.

## Service ownership

CloudRING maintainers govern modules distributed in this repository.
Independent module maintainers govern their own manifests, controllers,
adapters, portal extensions, billing connectors, support diagnostics, evidence,
and lifecycle behavior unless they contribute the module to CloudRING.
CloudRING maintainers review contributed and compatible modules against public
contracts.

## Enterprise and private boundary

Enterprise modules, proprietary integrations, company overlays, concrete
installation values, and customer deployment records are governed by their
owning organizations. Reusable adapters and services may become part of
CloudRING only through an intentional, licensed, source-safe contribution.

## Developer entry points

Compatibility discussions start from OCSv3 contracts, provider adapter
interfaces, IAM and policy decisions, portal shell slots, and evidence/readiness
requirements. Decisions should be recorded in public docs or machine-readable
contracts when they affect downstream implementers.
