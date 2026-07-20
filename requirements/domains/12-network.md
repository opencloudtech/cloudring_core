# 12 — Network

Scope: tenant-facing virtual networking (VPCs, subnets, routing, security
groups, floating addresses), the SDN layer behind a replaceable
network-profile contract, load balancing at L4 and L7 with a health-check
plane and capacity model, edge/NAT gateways, DNS hosting and platform
DNS/IPAM automation, global traffic steering, anti-DDoS and egress-control
integration profiles, private interconnect wiring, and the network evidence
required to call any installation ready. This domain covers the Cloud
Infrastructure Pod network substrate and the network products tenants
consume; it does not cover service mesh, federation data bus, or
observability pipelines (see Coverage notes).

**Domain contract.** Networking is a first-class tenant resource model, not
an implementation side effect: every network object is owned by a tenant
project, created and mutated asynchronously, and isolated by default. The
SDN is replaceable behind a network-profile contract — the platform must
not hard-code a data plane. Tenant traffic is denied unless explicitly
allowed; public exposure is always an explicit, billed, evidenced act. No
installation may be called network-ready without a verified connectivity
matrix, dual-stack operation, and demonstrated failure of one network
component without tenant-visible loss beyond documented limits.

---

### CR-NET-010 — Virtual networks as first-class tenant resources
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator, service-team
- **Problem:** Tenants need isolated private address domains to place
  workloads into; if networks are implicit per-cluster or per-hypervisor
  constructs, isolation, portability, and later federation become
  impossible to retrofit.
- **Requirement:** The platform MUST provide a tenant-owned virtual network
  resource (VPC) scoped to a tenant project, with create/read/update/delete
  and an explicit lifecycle state machine exposed through the public API.
  Networks MUST be isolated from each other by default: traffic between
  networks MUST NOT flow unless an explicit peering, route, or public
  address path is configured. Mutable operations MUST be asynchronous with
  an operation id and status polling. Networks MUST carry labels, region
  affinity, and quota accounting like every other tenant resource.
- **Acceptance evidence:** API contract tests for the network resource
  lifecycle and state machine; an isolation test proving two tenant
  networks cannot exchange traffic without explicit configuration; e2e
  provisioning evidence on a reference stand; quota enforcement tests.
- **Non-goals:** Cross-cloud network federation (FED domain); service-mesh
  east-west policy (K8S domain); physical fabric design.
- **Non-claims:** Multi-region network portability and peering at scale are
  not yet proven on any live stand.
- **Stop conditions:** Exposure — halt and escalate if any path allows
  tenant-to-tenant traffic without explicit configuration. Deletion —
  network deletion MUST be refused while dependent resources (subnets,
  addresses, balancer targets) exist, and MUST require the standard
  deletion confirmation flow.
- **Traceability:** legacy-platform-a (compute/network facade services),
  legacy-platform-b (folder-scoped L3 network model), current-core (OCS
  resource conventions). Related: CR-NET-020, CR-NET-050, CR-NET-190.

### CR-NET-020 — Subnets and IPAM contract
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Address allocation that is ad-hoc per hypervisor or per
  cluster leads to overlapping ranges, stranded leases, and un-auditable
  usage — the exact failure modes that make migration and peering
  impossible later.
- **Requirement:** Subnets MUST be zonal, non-overlapping CIDR blocks
  within a network, with well-known addresses (gateway, DNS) reserved and
  documented. IP address management MUST be a platform service with a
  pool/zone/region model, allocation commit-and-confirm semantics, and a
  grace period before an unconfirmed allocation is released. The IPAM
  contract MUST be infrastructure-owned and isolated from billing and
  tenant-business entities. Every allocation MUST be auditable (who, when,
  for which resource).
- **Acceptance evidence:** IPAM contract tests (commit/confirm/release,
  grace-period expiry, overlap rejection); audit-log inspection; subnet
  validation tests rejecting overlapping or out-of-range CIDRs; e2e
  allocation evidence on the reference stand.
- **Non-goals:** Tenant-facing IPAM marketplace features; BYOIP
  (bring-your-own-prefix) announcement — extension wave.
- **Non-claims:** Integration with an external enterprise IPAM system of
  record is specified only as an adapter port; no live integration is
  claimed.
- **Stop conditions:** Data — halt allocation changes if IPAM state and
  the SDN's actual allocations diverge; require reconciliation before new
  allocations. Migration — any CIDR renumbering plan MUST be gated on
  backup evidence and a rollback procedure.
- **Traceability:** legacy-platform-a (standalone IPAM with commit grace,
  external-IPAM adapter, registration-race lesson), legacy-platform-b
  (zonal subnet model with reserved addresses). Related: CR-NET-010,
  CR-NET-140.

### CR-NET-030 — Network-profile contract with replaceable SDN
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, vendor
- **Problem:** A platform hard-bound to one SDN cannot survive the
  technology cycle or serve providers with different substrate
  constraints; contract-before-technology is a charter principle.
- **Requirement:** All platform and tenant network functions MUST be
  expressed against a versioned network-profile contract covering virtual
  networks, subnets, routing, security groups, load balancing, NAT, and
  address management. The primary production profile MUST be implemented
  on an upstream-Kubernetes-native datapath (Cilium-class eBPF SDN) per the
  platform runtime policy. Alternative profiles MAY exist behind the same
  contract. A profile MUST declare its capabilities and limits so
  unsupported features fail fast and explicitly at validation time, never
  silently at runtime.
- **Acceptance evidence:** A published network-profile contract document
  with conformance test suite; the primary profile passing the full suite
  on the reference stand; a negative test proving unsupported-capability
  requests fail with explicit errors; profile-swap drill evidence in a
  non-production stand.
- **Non-goals:** Supporting every historical SDN; emulating
  hypervisor-vendor networking features that have no contract equivalent.
- **Non-claims:** Only the primary profile is expected to be conformance
  complete in the baseline; alternative profiles are unproven until they
  publish conformance evidence.
- **Stop conditions:** Migration — switching a stand's active profile MUST
  be gated on connectivity-matrix evidence before and after, with an
  automatic halt and rollback trigger on matrix regression.
- **Traceability:** current-core (Go/upstream-Kubernetes runtime policy,
  contract-before-technology principle), legacy-platform-a (vendor-bound
  SDN lock-in lesson), legacy-platform-b (two coexisting SDN generations
  lesson). Related: CR-NET-040, CR-NET-050, CR-NET-190.

### CR-NET-040 — Tenant-controlled routing
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider
- **Problem:** Real workloads need control over next-hop behavior —
  appliance insertion, VPN egress, segmented subnets — which implicit
  subnet-only routing cannot express.
- **Requirement:** The platform SHOULD provide tenant-managed route tables
  with static routes per network, attachable to subnets, with validation
  of next-hop types (instance, gateway, address, peering). Route changes
  MUST be asynchronous, validated against loops and black holes where
  detectable, and auditable. DHCP options and MTU/MSS behavior SHOULD be
  documented and configurable per subnet.
- **Acceptance evidence:** API contract tests for route-table CRUD and
  attachment; datapath tests proving configured routes take effect;
  negative tests for invalid next-hops; audit-trail inspection.
- **Non-goals:** Dynamic routing protocols toward tenants (BGP to tenant
  appliances is an extension); SDN-internal underlay routing.
- **Non-claims:** Overlap/conflict analysis across multiple attached route
  tables is best-effort; full formal route verification is not claimed.
- **Stop conditions:** Exposure — a route change that would direct tenant
  traffic to a next-hop outside the tenant's own resources MUST require
  elevated confirmation and be flagged in audit.
- **Traceability:** legacy-platform-b (static routes, DHCP options, MTU
  documentation), legacy-platform-a (network service facade). Related:
  CR-NET-010, CR-NET-030.

### CR-NET-050 — Security groups and default-deny network policy
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator, auditor
- **Problem:** Baseline tenant isolation and workload micro-segmentation
  are the minimum credible security posture for a multi-tenant cloud;
  platforms that default to allow-all ship exposure as a feature.
- **Requirement:** The platform MUST provide stateful security groups
  attachable per network interface: multiple groups per interface,
  ingress/egress rules, group-to-group references, a self-referential rule,
  and connection tracking with automatic return traffic. Platform
  metadata/health endpoints MUST remain reachable as documented
  exceptions. All newly created tenant interfaces MUST default to deny
  inbound and SHOULD default to restricted outbound. Kubernetes network
  policies on platform clusters MUST also default-deny, with explicit
  allow lists. Security groups MUST be documented as not a DDoS-protection
  mechanism.
- **Acceptance evidence:** Contract tests for rule semantics (statefulness,
  references, precedence); a default-deny conformance test proving a fresh
  interface accepts no unsolicited traffic; platform-cluster policy audit
  showing default-deny baselines; negative tests for rule conflicts.
- **Non-goals:** L7-aware firewalling, IDS/IPS, and DDoS scrubbing
  (separate profiles); host-level firewall management.
- **Non-claims:** Rule-scale performance (tens of thousands of rules per
  interface) is not yet benchmarked on the primary profile.
- **Stop conditions:** Exposure — any change widening effective access
  (new allow rule, group attachment) MUST be attributable to an
  authenticated actor; an unexplained default-allow state on any tenant
  interface halts readiness promotion. Trust — the metadata exception list
  MUST be minimal, versioned, and reviewable.
- **Traceability:** legacy-platform-b (stateful interface-level groups,
  default metadata exception, explicit non-DDoS caveat), legacy-platform-a
  (allow-all policy stub lesson). Related: CR-NET-030, CR-NET-160,
  CR-NET-190.

### CR-NET-060 — L4 network load balancing
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Tenants cannot build available services without a managed
  way to distribute TCP/UDP traffic across targets with failure detection;
  it is the minimum viable balancer product.
- **Requirement:** The platform MUST provide L4 load balancers in external
  and internal types with listeners, target groups, configurable health
  checks, and a documented session-affinity mode. Balancers MUST have an
  explicit API state machine, asynchronous mutation, and quota accounting.
  Internal balancer behavior that shadows direct target access on balanced
  ports MUST be documented as a limitation or avoided by design. Balancer
  datapath MUST survive loss of a single balancer node within documented
  recovery bounds.
- **Acceptance evidence:** API contract tests incl. state machine; traffic
  tests (distribution, affinity, health-check-driven target removal);
  one-node-loss drill evidence with measured recovery; limitation
  documentation review.
- **Non-goals:** TLS termination and HTTP semantics (CR-NET-080); global
  steering (CR-NET-150).
- **Non-claims:** Affinity consistency during target-group churn is
  best-effort; sub-second failover is not claimed for the baseline.
- **Stop conditions:** Money — external balancers and their addresses are
  billable; creation MUST pass quota and billing-context checks, and
  metering evidence MUST exist before the feature is declared publicly available. Deletion —
  balancer deletion MUST follow the deletion confirmation flow and release
  attached addresses only per the address resource's own rules.
- **Traceability:** legacy-platform-b (external/internal L4 balancer,
  target groups, affinity, port-shadowing lesson), legacy-platform-a
  (balancer lifecycle in compute facade). Related: CR-NET-070, CR-NET-090,
  CR-NET-110.

### CR-NET-070 — Centralized health-check plane
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** If health checking is done ad hoc by each datapath node,
  tenants cannot predict probe sources, false positives multiply with
  node count, and debugging target flapping becomes guesswork.
- **Requirement:** Load-balancer health checking SHOULD be executed by a
  dedicated health-check plane (controller plus checking agents) with
  stable, documented source address ranges that tenants can allow-list in
  security groups. Health-check configuration (protocol, path/port,
  intervals, thresholds) MUST be validated, and probe outcomes MUST be
  observable per target with history sufficient for incident analysis.
- **Acceptance evidence:** Architecture document; source-range
  documentation consumed by security-group validation; probe-observability
  tests; a drill demonstrating consistent behavior during checker-agent
  loss.
- **Non-goals:** Synthetic application-level monitoring beyond balancer
  health; tenant self-hosted checkers.
- **Non-claims:** Checker-plane autoscaling under fleet-wide probe load is
  not yet capacity-tested.
- **Stop conditions:** Trust — probe source ranges MUST be treated as
  platform trust boundaries; any change to them requires versioned
  documentation and tenant notice, never silent renumbering.
- **Traceability:** legacy-platform-b (separate healthcheck controller/node
  roles, fixed documented probe ranges). Related: CR-NET-050, CR-NET-060,
  CR-NET-080.

### CR-NET-080 — L7 application load balancing
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** Production HTTP/gRPC workloads need host/path routing, TLS
  termination, and multi-certificate serving — capabilities that cannot be
  bolted onto L4 balancing without a distinct product surface.
- **Requirement:** The platform SHOULD provide an L7 balancer with HTTP and
  stream listeners, TLS 1.2+ termination, SNI-based certificate selection,
  host/path routers, backend groups with health checks and locality-aware
  routing, and per-zone enable/disable. Certificates MUST be referenced
  from the platform certificate/secret store, never uploaded as inline
  material. Integration with the platform's certificate issuance workflow
  MUST exist.
- **Acceptance evidence:** API contract tests (listeners, routers, backend
  groups); TLS termination and SNI tests with referenced certificates;
  locality-routing behavior tests; per-zone disable drill evidence.
- **Non-goals:** WAF functionality; service-mesh ingress replacement;
  arbitrary TCP proxying beyond documented stream support.
- **Non-claims:** Performance under TLS resumption-heavy and
  many-certificate workloads is unbenchmarked; HTTP/3 is not claimed.
- **Stop conditions:** Keys — private key material MUST never appear in
  API payloads, logs, or state dumps; any discovery of inline key material
  halts the pipeline and triggers secret-rotation. Exposure — listeners
  bound to public addresses MUST be explicit and billed.
- **Traceability:** legacy-platform-b (L7 balancer with SNI routing,
  routers, backend groups, cert-store integration), current-core (secrets
  are never configuration). Related: CR-NET-060, CR-NET-070, CR-NET-090.

### CR-NET-090 — Load-balancer capacity model
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Without a capacity unit model, balancer performance is a
  surprise: tenants cannot size, providers cannot price or plan, and
  overload manifests as unexplained production degradation.
- **Requirement:** Balancer products SHOULD declare a capacity-unit model
  (LCU-class) mapping new connections, active connections, and throughput
  to units; quota, pricing inputs, and subnet sizing guidance (minimum
  free addresses per unit) SHOULD derive from it. Actual consumption MUST
  be metered and exposed to tenants with alerting before exhaustion.
- **Acceptance evidence:** Published capacity model with unit definitions;
  load-test evidence mapping measured traffic to units within a documented
  error band; metering pipeline tests; tenant-facing consumption metrics.
- **Non-goals:** Automatic per-tenant autoscaling guarantees in the
  baseline; cross-region capacity pooling.
- **Non-claims:** Unit definitions are provisional until validated against
  production-class load tests; the error band is not yet measured.
- **Stop conditions:** Money — capacity units feed billing; any unit-model
  change MUST be versioned with effective dates, and un-metered billable
  consumption halts public-availability claims.
- **Traceability:** legacy-platform-b (LCU capacity model with subnet
  sizing guidance), legacy-platform-a (per-product metering plugins).
  Related: CR-NET-060, CR-NET-080.

### CR-NET-100 — NAT and edge gateways per function
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, tenant
- **Problem:** Egress for private workloads, address-family translation,
  and provider-edge interconnection have different scaling and failure
  characteristics; one undifferentiated gateway box becomes a fragile
  single point of failure and a scaling ceiling.
- **Requirement:** Edge functions MUST be decomposed per function —
  tenant egress NAT, IPv4 egress, IPv6 egress, and provider
  interconnect/direct-connect — as separately deployable, separately
  scalable gateway roles. Tenant NAT gateways MUST support port control
  and documented limits, be zonal with explicit failure behavior, and be
  billable/quota-accounted resources. Gateway datapath state MUST NOT be a
  single point of failure for established tenant connections beyond
  documented recovery bounds.
- **Acceptance evidence:** Deployment topology documentation; per-function
  gateway role tests; one-gateway-loss drill evidence with measured
  connection impact; tenant NAT e2e tests (source NAT, port exhaustion
  behavior); quota and metering checks.
- **Non-goals:** Tenant-managed virtual router appliances (extension);
  carrier-grade NAT for the provider's own backbone.
- **Non-claims:** Connection-survival guarantees during gateway failover
  are bounded and not yet quantified beyond best-effort for short flows.
- **Stop conditions:** Money — NAT egress is metered; un-metered egress
  paths halt public-availability claims. Exposure — egress policy changes MUST be audited;
  an open unrestricted-egress default for all tenants requires explicit
  provider opt-in, never silent default.
- **Traceability:** legacy-platform-b (gateway hosts specialized per
  function: IPv4, IPv6, NAT, direct-connect), legacy-platform-a (DNAT rule
  lifecycle in network facade). Related: CR-NET-110, CR-NET-120,
  CR-NET-170, CR-NET-180.

### CR-NET-110 — Floating and public addresses
- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Tenants need stable public endpoints decoupled from
  workload lifecycle; if public reachability is tied to individual
  instances, availability and mobility are impossible.
- **Requirement:** The platform MUST provide public address resources that
  tenants can allocate, hold, and bind/unbind to interfaces or balancers,
  with DNS-name assignment per platform conventions. Public binding MUST
  be an explicit act with quota and billing checks; addresses MUST be
  reclaimable on tenant deletion and auditable throughout their lifecycle.
  The platform MUST document which address ranges are used and how abuse
  (e.g., mail-port egress policy) is handled.
- **Acceptance evidence:** Address lifecycle contract tests (allocate,
  bind, move, release, reclaim); binding audit-trail inspection; e2e
  reachability evidence on the reference stand; abuse-policy documentation
  review.
- **Non-goals:** Bring-your-own-prefix announcement (extension);
  per-tenant anycast addresses (CR-NET-150).
- **Non-claims:** Address portability across regions is not claimed;
  rebind propagation-time bounds are not yet measured fleet-wide.
- **Stop conditions:** Exposure — public binding changes MUST require
  authenticated intent and be audit-logged; unexpected public reachability
  of a previously private target halts readiness promotion. Money —
  held-but-unbound addresses are billable and MUST be metered.
- **Traceability:** legacy-platform-b (public address model, egress
  port-25 policy documentation), legacy-platform-a (public IP + DNAT
  entities in compute FSM). Related: CR-NET-060, CR-NET-100, CR-NET-180.

### CR-NET-120 — Dual-stack IPv4/IPv6 as a gate
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, tenant, auditor
- **Problem:** IPv6 retrofitted after IPv4-only launch historically fails:
  control planes, metering, security groups, and tenant tooling all carry
  single-stack assumptions that surface as production gaps.
- **Requirement:** All network products — networks, subnets, addresses,
  balancers, NAT, DNS, security groups — MUST support dual-stack
  IPv4/IPv6 from the first release, and platform-internal network paths
  SHOULD be dual-stack where the substrate allows. Readiness gates MUST
  include dual-stack operation: an installation that cannot demonstrate
  tenant IPv6 reachability end-to-end MUST NOT be called network-ready.
  Address-family behavior MUST be explicit in every API (no implicit
  family inference).
- **Acceptance evidence:** Dual-stack conformance test suite across all
  network resources; live IPv6 tenant-traffic evidence on the reference
  stand; API review confirming explicit family fields; security-group
  parity tests for both families.
- **Non-goals:** IPv6-only tenant networks in the baseline; translation
  (NAT64-class) services.
- **Non-claims:** Feature parity between families is the target; any
  verified gap MUST be documented as a limitation, not silently absent.
- **Stop conditions:** Exposure — security-group and egress-filter rules
  MUST apply symmetrically to both families; an IPv6 bypass of an IPv4
  deny halts promotion and triggers a security review.
- **Traceability:** legacy-platform-b (dual-stack interfaces and separate
  IPv4/IPv6 gateway roles), vision-deck (jurisdiction/provider diversity).
  Related: CR-NET-050, CR-NET-100, CR-NET-190.

### CR-NET-130 — Tenant DNS zones
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, service-team
- **Problem:** DNS hosting is an expected cloud capability; without it
  tenants must bolt on external providers, breaking the
  single-control-plane experience and complicating automation.
- **Requirement:** The platform SHOULD provide tenant DNS hosting:
  public zones with standard record types, API-driven management, and an
  authoritative backend of the PowerDNS class (or contract-equivalent).
  Zone lifecycle MUST be asynchronous, auditable, and billed/quota
  accounted where the provider prices it. Backend specifics MUST sit
  behind a DNS-provider adapter so the tenant API is backend-neutral.
  Delegation checks and zone-import convenience SHOULD exist.
- **Acceptance evidence:** DNS API contract tests (zones, records,
  TTLs); resolution evidence against the authoritative backend on the
  reference stand; adapter conformance tests; audit-trail inspection.
- **Non-goals:** Recursive resolver service for tenants; DNSSEC signing as
  a baseline feature (extension); registrar functions.
- **Non-claims:** Per-tenant isolated DNS server instances (vs. shared
  multi-tenant backend) are an open design option, not a committed claim.
- **Stop conditions:** Trust — DNS is a trust anchor; zone transfer,
  takeover, or record mutation without authenticated intent halts the
  service and triggers incident response. Deletion — zone deletion MUST
  follow the confirmation flow with a documented propagation warning.
- **Traceability:** legacy-platform-a (PowerDNS-backed DNS-as-a-service
  with sync to per-client instances, open design questions recorded),
  legacy-platform-b (DNS service in platform wave). Related: CR-NET-140,
  CR-NET-150.

### CR-NET-140 — Platform DNS/IPAM automation for provisioned resources
- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, service-team
- **Problem:** Platform VMs and services that lack automatic, idempotent
  DNS and IP registration drift into an un-addressable, un-inventoried
  estate; registration races between discovery and provisioning were a
  named production failure mode.
- **Requirement:** Every platform-provisioned compute resource MUST be
  registered idempotently in IPAM and in DNS (forward and reverse
  records) as part of its provisioning workflow, with race-safety between
  discovery scans and registration, and deregistration on decommission.
  Registration MUST be verifiable: an inventory query MUST reconcile
  DNS/IPAM state against actual resources. Service VMs MUST
  self-register through the same workflow, not through per-service ad hoc
  scripts.
- **Acceptance evidence:** Provisioning-workflow tests asserting IPAM and
  DNS records exist before readiness; reconciliation job evidence showing
  zero drift on the reference stand; race-condition regression tests;
  decommission cleanup tests.
- **Non-goals:** Tenant-facing DNS features (CR-NET-130); external
  enterprise DNS federation.
- **Non-claims:** Reconciliation frequency and drift-detection latency
  bounds are not yet tuned for large fleets.
- **Stop conditions:** Data — divergence between DNS/IPAM records and the
  actual resource inventory beyond a threshold halts new provisioning
  until reconciled. Deletion — deregistration MUST precede address release
  to prevent stale-record hijack.
- **Traceability:** legacy-platform-a (IPAM+DNS as first-class
  provisioning steps, discovery-race lesson), legacy-platform-b (service
  VMs self-registering in DNS). Related: CR-NET-020, CR-NET-130.

### CR-NET-150 — Global server load balancing and anycast
- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Multi-zone and multi-region tenants need health-aware global
  traffic steering; without it, region evacuation and geographic latency
  optimization fall entirely on tenant DIY DNS.
- **Requirement:** The platform SHOULD provide DNS-based global load
  balancing: health-checked endpoint pools, weighted/geographic steering
  policies, per-project quotas, and tenant notifications on endpoint state
  changes. An anycast ingress tier MAY front selected platform and tenant
  services where the provider's edge justifies it. Steering decisions MUST
  be observable and health-state changes auditable.
- **Acceptance evidence:** GSLB API contract tests; health-check-driven
  steering behavior tests incl. failover; notification delivery tests;
  quota enforcement tests; for anycast, edge-announcement drill evidence.
- **Non-goals:** Replacing regional balancers; CDN functionality;
  guaranteed sub-TTL failover (DNS caching realities documented).
- **Non-claims:** Forward-looking: no production anycast tier is claimed
  for the baseline; steering-policy richness is limited to documented
  policy types.
- **Stop conditions:** Trust — steering changes affect live tenant
  traffic; health-check false-positive storms MUST trigger damping and
  operator alerting, not oscillation. Exposure — anycast announcements
  MUST be scoped to approved prefixes with change control.
- **Traceability:** legacy-platform-a (GSLB service family with health
  checking, quotas, state-change notifications), legacy-platform-b
  (anycast/decap ingress tier). Related: CR-NET-130, CR-NET-160,
  CR-NET-070.

### CR-NET-160 — Anti-DDoS integration profile
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, tenant, vendor
- **Problem:** Public clouds are DDoS targets by definition; without an
  integration profile, each provider improvises protection wiring and
  tenants get inconsistent or absent mitigation.
- **Requirement:** The platform SHOULD define an anti-DDoS integration
  profile covering: detection-signal intake, diversion/scrubbing workflow
  (external scrubbing service or platform decapsulation tier),
  per-tenant protection products sellable through the catalog, and clear
  documentation that security groups are not DDoS protection. Protection
  state (under attack, mitigating) MUST be visible to the affected tenant
  and to operators. The profile MUST be an OCS-compatible integration, not
  hard-wired to one mitigation vendor.
- **Acceptance evidence:** Published integration profile with adapter
  contract; a reference adapter implementation; simulated-attack drill
  evidence showing detection-to-mitigation flow; tenant-visibility
  notification tests.
- **Non-goals:** Building a full scrubbing network as the baseline;
  guaranteeing mitigation capacity against arbitrary attack volumes.
- **Non-claims:** Mitigation efficacy is bounded by the provider's chosen
  backend; no specific scrubbing capacity or SLA is claimed in the OSS
  layer.
- **Stop conditions:** Trust — diversion of tenant traffic to a scrubbing
  path MUST be tenant-visible and auditable; silent diversion halts the
  integration. Money — protection products are billable; un-metered
  protection consumption blocks GA claims.
- **Traceability:** legacy-platform-a (anti-DDoS product connector as
  catalog item), legacy-platform-b (decap/anycast tier separate from
  tenant security groups). Related: CR-NET-050, CR-NET-150.

### CR-NET-170 — Private interconnect and provider wiring
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, operator
- **Problem:** Enterprise onboarding and private-cloud federation require
  dedicated connectivity between tenant premises/other clouds and the
  platform; ad hoc per-customer tunnel scripts do not scale and are
  unauditable.
- **Requirement:** The platform SHOULD define a private-interconnect
  product contract: dedicated link onboarding (tagged VLANs, link
  aggregation, per-direction rate policing), attachment of interconnects
  to tenant networks with routing integration, and a point-of-presence
  inventory. As an OSS-layer minimum, an overlay-based private wiring
  option (vRack-class L2/L3 private connectivity across provider
  infrastructure) MUST be expressible through the network-profile
  contract so providers can offer it without core changes.
- **Acceptance evidence:** Interconnect contract document; provisioning
  workflow tests for attachment and routing; an e2e private-connectivity
  demonstration on the reference stand (overlay class acceptable);
  policing/encapsulation configuration evidence.
- **Non-goals:** Operating physical cross-connect logistics; guaranteed
  bandwidth SLAs in the OSS layer.
- **Non-claims:** Physical interconnect attributes (optics, VLAN ranges,
  PoP inventory) are specified as contract fields; no live physical
  interconnect is claimed. Enterprise-grade throughput validation is
  unproven.
- **Stop conditions:** Trust — interconnect attachment crosses
  administrative boundaries; cross-tenant or cross-provider wiring MUST
  require bilateral confirmation and be fully audited. Money — dedicated
  links are billable; metering evidence precedes public-availability claims.
- **Traceability:** legacy-platform-b (dedicated L2 trunk spec: tagged
  VLANs, LACP, policing, PoP list), legacy-platform-a (per-customer
  tunnel automation lesson). Related: CR-NET-010, CR-NET-030,
  CR-NET-100.

### CR-NET-180 — Egress controls and regulatory filtering profile
- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, operator, auditor
- **Problem:** Providers face abuse (spam, scanning) and jurisdictional
  blocking obligations; without a built-in control profile, compliance
  becomes manual firewall toil that drifts and fails audits.
- **Requirement:** The platform MUST define an egress-control profile:
  default policy for abuse-prone ports (e.g., outbound mail) with a
  tenant-visible exception workflow, plus a pluggable regulatory-filtering
  pipeline (block-list intake → enforcement configuration → alerting) that
  a provider can wire to its jurisdiction's sources. Egress policy MUST be
  uniform across address families, and every automated enforcement action
  MUST be logged and reviewable.
- **Acceptance evidence:** Egress-policy conformance tests (default
  posture, exception grant flow); filtering-pipeline drill with a
  synthetic block list; enforcement audit-log inspection; dual-stack
  parity tests.
- **Non-goals:** Content inspection beyond documented filtering classes;
  tenant-facing content-filter products.
- **Non-claims:** Jurisdictional block-list formats differ; only the
  pipeline contract and a reference adapter are claimed, not per-country
  compliance coverage.
- **Stop conditions:** Trust — enforcement configuration changes MUST be
  attributable and reviewable; an unexplained enforcement divergence
  (blocking beyond the fed list, or failing to block it) halts the
  pipeline. Exposure — disabling default abuse-port filtering platform
  wide MUST require explicit provider-level approval with audit.
- **Traceability:** legacy-platform-b (default mail-port filtering,
  automated registry→config→enforce→alert pipeline), legacy-platform-a
  (compliance surfaces as isolated services). Related: CR-NET-100,
  CR-NET-110, CR-NET-120.

### CR-NET-190 — Network connectivity evidence as readiness gate
- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, auditor, agent
- **Problem:** "The network works" is unfalsifiable without a defined
  evidence artifact; platforms ship with silent partitions, MTU
  regressions, and policy gaps that only tenants discover.
- **Requirement:** Every installation MUST produce a versioned network
  connectivity matrix as readiness evidence: zone-to-zone, tenant-to-edge,
  tenant-to-platform-services, both address families, with expected MTU,
  latency, and policy outcomes recorded. The matrix MUST be regenerated
  on a schedule and after any network-profile, gateway, or security-model
  change; a failing matrix blocks readiness promotion. Evidence MUST be
  stored per platform evidence discipline (append-only, timestamped,
  non-synthetic for promotion).
- **Acceptance evidence:** The matrix generator and its test suite;
  fresh verified matrix artifacts on the reference stand; gate
  configuration proving promotion failure on a red matrix; drill evidence
  of a forced regression being caught.
- **Non-goals:** Full formal verification of the datapath; continuous
  per-flow verification.
- **Non-claims:** Matrix coverage dimensions (which flows are sampled)
  will evolve; coverage gaps MUST be listed in the artifact itself.
- **Stop conditions:** Exposure — a matrix cell showing unexpected
  reachability (isolation breach) halts promotion and triggers security
  review. Migration — profile or gateway changes MUST NOT complete
  without a post-change green matrix.
- **Traceability:** legacy-platform-b (active network-dataplane probing
  as dedicated service), req-history (evidence discipline), current-core
  (evidence-over-claims principle). Related: CR-NET-030, CR-NET-050,
  CR-NET-120, CR-NET-210.

### CR-NET-200 — Network flow telemetry
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Capacity planning, incident forensics, and egress-cost
  attribution need flow-level visibility that metrics aggregates cannot
  provide.
- **Requirement:** The platform SHOULD collect flow telemetry
  (sFlow/IPFIX-class) from the network datapath, export it to the
  observability stack for near-real-time analysis, and provide a
  long-term analytics sink. Collection MUST be privacy-scoped (headers/
  metadata class, not payload) and tenant-attributable for usage
  visibility.
- **Acceptance evidence:** Collector deployment on the reference stand
  with exported metrics; sample-rate and overhead documentation;
  forensics drill using flow data to explain a synthetic incident.
- **Non-goals:** Full packet capture; tenant-facing flow products in the
  baseline.
- **Non-claims:** Flow-based per-tenant traffic accounting is a design
  input to billing but not yet a committed billing meter.
- **Stop conditions:** Money — if flow data feeds billing meters, meter
  correctness evidence precedes charging. Data — flow retention MUST
  follow platform data-retention policy; unbounded retention is a data
  risk.
- **Traceability:** legacy-platform-b (flow collectors with metrics
  export and analytics sink). Related: CR-NET-100, CR-NET-190.

### CR-NET-210 — Active synthetic dataplane probing
- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** Passive metrics miss gray failures — paths that forward at
  degraded rate, asymmetric drops, policy mismatches — which active
  probing catches before tenants do.
- **Requirement:** The platform SHOULD run synthetic probers (dedicated
  probe workloads in every zone) exercising representative tenant-class
  paths continuously, with results feeding alerting and the connectivity
  matrix (CR-NET-190). Prober placement and test scenarios SHOULD be
  declared as code.
- **Acceptance evidence:** Prober deployment manifests as code; alerting
  drill on an injected degradation; matrix integration evidence.
- **Non-goals:** Replacing tenant-side monitoring; per-tenant custom
  probes in the baseline.
- **Non-claims:** Scenario coverage is partial by design; probe blindness
  to unrepresented paths is documented, not eliminated.
- **Stop conditions:** n/a (probing generates synthetic traffic only; no
  money/data/keys/trust/exposure/deletion/migration/settlement risk class
  impact beyond what CR-NET-190 already gates).
- **Traceability:** legacy-platform-b (dedicated prober images and
  creator/API components). Related: CR-NET-190, CR-NET-200.

### CR-NET-220 — Network resource reconciliation and cleanup
- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, agent
- **Problem:** Orphaned subnets, balancers, NAT rules, and addresses
  accumulate from failed operations and force-deleted parents; they leak
  quota, money, and address space, and they block tenant off-boarding.
- **Requirement:** The platform SHOULD run reconciliation loops that
  detect and report network objects with no live owner or divergent
  desired/actual state, and provide a graded, automatable cleanup path
  (report → quarantine → delete) with dry-run and per-resource audit.
  Cleanup of tenant-billable orphans MUST emit the billing events that
  stop charges.
- **Acceptance evidence:** Reconciler test suite incl. injected orphans;
  dry-run evidence; cleanup audit-trail inspection; billing-event
  verification on cleanup of metered resources.
- **Non-goals:** Auto-deleting anything a tenant could still own without a
  quarantine window; cross-domain garbage collection (each domain owns its
  reconciler).
- **Non-claims:** Divergence-detection completeness depends on SDN
  inventory quality of the active profile; unmeasurable on profiles
  without read-back.
- **Stop conditions:** Deletion — automated cleanup MUST require the
  quarantine window and MUST NOT touch resources whose ownership is
  ambiguous; ambiguous cases escalate to an operator. Money — cleanup
  MUST stop associated billing; deleting a billable resource without
  emitting stop events halts the reconciler.
- **Traceability:** legacy-platform-a (orphan garbage-collector endpoints
  for networks/balancers; graded cleanup tooling lesson), legacy-platform-b
  (resource-inventory tooling). Related: CR-NET-010, CR-NET-060,
  CR-NET-110.

---

## Coverage notes

This domain deliberately defers:

- **Service mesh, ingress controllers, and in-cluster east-west policy** to
  the Kubernetes/containers domain (14); this file covers the substrate and
  tenant products, not workload-to-workload mesh.
- **Observability pipelines, TSDB, and alerting mechanics** to domain 20;
  here we only require the network-specific telemetry and evidence feeds.
- **Billing implementation** (tariffs, charging, settlement) to domain 16;
  network requirements reference metering evidence only as gating input.
- **IAM enforcement for network APIs** to domain 15; this domain assumes
  authenticated, attributable actors on every mutation.
- **Cross-cloud network federation, P2P data bus, and global-portal
  reachability** to domain 23; interconnect here is single-provider scope.
- **DNS/IPAM server implementation choices and host-level DNS resolver
  configuration** to the platform-foundation (10) and deployment/IaC (22)
  domains; this domain pins the contracts and automation behavior.
- **Compute NIC attachment and instance-level bandwidth** to the compute
  domain (11); this domain defines the network objects they attach to.
