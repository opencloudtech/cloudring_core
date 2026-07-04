# Marketplace Module

Marketplace is modeled as an OCSv3 module package. CloudRING consumes
catalog listing, entitlement, order handoff, revenue-share, evidence, support,
and billing contracts from the manifest.

Provider storefronts, reseller systems, payment collection, and fulfillment
adapters remain outside core. This directory contains only portable module
metadata.

Validate the module manifest with:

```powershell
go run ./cmd/ocsctl validate ./cloudring_core/modules/marketplace/module-package.json
```
