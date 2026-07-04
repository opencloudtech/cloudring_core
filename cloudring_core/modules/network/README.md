# Network Module

This module package declares the public OCSv3 ownership surface for tenant
networking. It depends on portable Gateway API and Cilium-compatible policy
classes, but it does not name a provider load balancer, endpoint, credential,
or deployment overlay.

The core consumes `module-package.json` for UI, billing, support, readiness,
durability, degraded, denied, and blocked-state metadata. Runtime controllers
and provider-specific implementation values remain outside CloudRING.
