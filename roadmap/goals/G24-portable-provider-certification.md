# G24 — Portable provider certification

## Outcome

Certify two independent downstream products on the already complete public
interfaces: OpenCloudTech on OVH and CloudLinux on an approved disposable
independent bare-metal installation. This goal adds no new generic installer or platform
engine; any discovered generic defect is fixed in its owning earlier OSS surface.

All reusable fixes must merge into public OSS and the exact accepted change must
be deployed and live-verified at `https://hub.cloudring.org`.

## Scope

- Audit Enterprise to only public gitlink, OVH/site/jurisdiction bindings,
  proprietary business/services and protected evidence; generic duplicates zero.
- Audit Provider to only public gitlink, CloudLinux adapters, protected inventory,
  site overlays, ownership, migration and private evidence; no Enterprise read.
- Complete independent-site protected inputs/bindings and execute the public G02/G22/
  G23 workflows unchanged: install, operate, backup/restore, upgrade/rollback,
  failure drill and cleanup.
- Discover/adopt or migrate one representative legacy IaaS resource through G06,
  with coexistence, rollback and explicit ownership.
- Build/install/operate/upgrade/remove one CloudLinux-owned OCS product from its
  own repo/CI using released public artifacts.
- Keep the secondary hosted-metal profile honestly `contract-ready` unless it passes the same real live
  certification; its profile/plan must still pass public conformance.
- Reprove exact gitlinks, SafePush Stage 9, source boundaries and signed release
  artifacts for all consumer mains.

## Required journeys

- clean public+Enterprise install/upgrade/rollback and full G23 hub regression;
- clean public+Provider+SafePush install on a CloudLinux-controlled independent site,
  then operate, restore, upgrade/rollback, bounded failure and cleanup;
- two reproducible protected inventory captures with no private data committed;
- independent CloudLinux product lifecycle and billing/support path;
- legacy discovery/adoption/migration with rollback;
- explicit secondary-site non-claim or equivalent full certification.

## Acceptance

- A real independent bare-metal environment is a hard gate. Missing hardware, credentials or
  authority blocks G24; synthetic evidence cannot complete it.
- Both downstreams use the same exact accepted public release and have zero
  generic duplicate/private cross-dependency.
- Provider can remove `preflight-and-plan-only` and enable production use only
  after all live gates pass.
- Independent engineers complete site and service workflows without author help.
