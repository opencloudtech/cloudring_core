# Billing FinOps Module

Billing FinOps is modeled as an OCSv3 module package. CloudRING consumes
usage ledger, allocation, invoice-preview, evidence, support, and marketplace
settlement contracts from the manifest.

Payment processors, tax systems, invoice engines, and provider accounting
adapters remain outside core. This directory exposes only portable connector
metadata and review references.

Validate the module manifest with:

```sh
go run ./cmd/ocsctl validate ./modules/billing-finops/module-package.json
```
