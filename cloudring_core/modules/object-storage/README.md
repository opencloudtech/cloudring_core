# Object Storage Module

This module package declares the public OCSv3 ownership surface for generic
S3-compatible object storage. It describes bucket lifecycle, access grants by
reference, billing, UI, support, readiness, durability, degraded, denied, and
blocked states without embedding provider endpoints or credential values.

CloudRING consumes `module-package.json` metadata only. Runtime backends,
provider-specific overlays, live endpoints, and raw access material remain
outside CloudRING.
