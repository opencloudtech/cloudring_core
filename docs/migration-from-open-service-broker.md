# Migration From Open Service Broker

Open Service Broker API focuses on service catalog, provision, bind, unbind,
and deprovision. OCSv3 keeps the catalog idea but adds cloud-platform surfaces
that modern Kubernetes-native services need.

Key differences:

- OCSv3 uses Kubernetes-native claims and controllers.
- Billing, IAM, support, evidence, portal extensions, analytics, durability,
  and rollback are first-class.
- Provider adapters stay outside platform core.
- Readiness requires evidence.
- Microfrontends are declared by module package, not hardcoded in platform UI.

Migration steps:

1. Map OSB service/plan to OCSv3 catalog and capability.
2. Map provision/deprovision to lifecycle actions.
3. Add backup, restore, export, retry, and rollback.
4. Add billing meters and idempotent events.
5. Add IAM, support, evidence, and portal extension surfaces.
6. Run `ocsctl conformance`.
