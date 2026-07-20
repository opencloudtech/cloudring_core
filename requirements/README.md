# CloudRING Requirements

This directory contains the complete, detailed product and platform requirements
for the CloudRING open-source cloud platform (`opencloudtech/CloudRING`,
Apache-2.0). It is the single source of truth for what the platform must do;
implementation evidence lives elsewhere and links back here.

## Audience

- Service teams building on the Open Cloud Standard (OCS)
- Platform engineers implementing the OSS core
- Providers deploying CloudRING (public, private, EDGE, federated)
- Agents (AI or automation) executing changes against these requirements

## Reading order

1. `00-product-charter.md` — vision, roles, principles, scope
2. `01-requirement-schema.md` — requirement format, ID scheme, priorities
3. `02-source-and-method.md` — where these requirements come from, safety rules
4. `domains/*.md` — normative requirements by domain
5. `scenarios/*.md` — end-to-end acceptance scenarios
6. `matrices/*.md` — coverage, risk and OCS-fit matrices
7. `registry/requirements.json` — machine-readable index (generated)

## Rules for changing requirements

- English only.
- Every requirement follows `01-requirement-schema.md` exactly.
- Sparse IDs (`…-010`, `…-020`, …) — never renumber, never reuse a retired ID.
- The JSON registry is generated from the Markdown files; drift between files
  and registry is a defect (see `matrices/` validation notes).
- A requirement is never marked delivered without linked, fresh evidence.
  `blocked` and `unverified` are first-class honest states.
- No copied text from proprietary sources, no brand names of reference
  platforms, no endpoints, no tenant data, no secrets (see
  `02-source-and-method.md`).
