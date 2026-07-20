# 15 — Identity, Access Management & Security

This domain covers the platform's trust substrate: the hierarchical resource
model and its inherited access bindings, the fine-grained permission and role
model, subjects (users, service accounts, groups), token issuance and rotation,
federation and MFA, workload identity, break-glass emergency access,
authorization availability and caching limits, key management and envelope
encryption, the secrets manager, the audit log, tenant-isolation proof,
fail-closed behavior, security scanning gates, and compliance mapping hooks.
It defines WHAT every installation must prove about identity and security, not
the UX of login screens, the operations incident process, or network-edge
defense, which live in their own domains.

**Domain contract.** Every authentication and authorization path fails closed:
on error, timeout, ambiguity, or missing evidence the answer is deny.
Authorization is fine-grained (`service.resource.verb`), decided against a
hierarchical resource tree with inherited bindings, and every effective access
inside a container is visible and revocable at its root. Identities are
short-lived: workloads authenticate through platform-issued identity, static
credentials are exception-only, secrets are brokered and referenced — never
configuration. Encryption at rest uses managed keys under a hardware root of
trust; the secrets manager is honest about its sealed state. Every security
decision, key operation, and emergency access leaves an immutable, queryable
audit record, tenant isolation is proven by continuously run tests, and no
release ships with a disabled security gate. `blocked` and `unverified` are
honest states here; no security claim is made without fresh evidence.

## Requirements

### CR-IAM-010 — Hierarchical resource model with inherited bindings

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, service-team, operator
- **Problem:** Every authorization decision needs one consistent containment
  tree. Without it each service invents its own scoping, and revocation,
  listing, and audit break in per-service ways.
- **Requirement:** The platform MUST define a hierarchical resource model
  (organization → project/account → folder → service resources) in which an
  access binding on a container inherits to everything beneath it. Container
  lifecycle states (active, suspended, deleting, recovering) MUST be
  first-class and evaluated on the authorization path. Resource identity MUST
  be stable, typed, and provider-neutral. Bindings per resource MUST be
  bounded by an explicit, visible quota.
- **Acceptance evidence:** contract test suite proving inheritance semantics
  (grant at container is effective at leaf; revoke at container denies at
  leaf); lifecycle-state denial tests; binding-quota enforcement tests; the
  suite green in CI and in the authorization conformance gate.
- **Non-goals:** a central registry of every service's resources (services own
  their resources); modeling an organization's internal HR structure.
- **Non-claims:** deny-bindings (negative grants) are not supported and no
  semantics are claimed for them; an intermediate grouping level between
  organization and project is anticipated but unspecified.
- **Stop conditions:** any change to inheritance semantics halts rollout if
  revocation or isolation tests regress; containers in deleting state MUST
  reject new bindings, and a violation blocks release (data, trust).
- **Traceability:** legacy-platform-b; legacy-platform-a; req-history;
  current-core (tenant/project dimensions of the authorization-decision
  contract). Related: CR-IAM-050, CR-IAM-060, CR-IAM-170.

### CR-IAM-020 — Fine-grained permission grammar

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator, auditor
- **Problem:** Coarse permissions force over-privileged roles and make audit
  meaningless: if one permission covers many actions, "who could do what" has
  no precise answer.
- **Requirement:** Permissions MUST follow `<service>.<resource>.<verb>` with
  standard verbs (get/create/update/delete/list/use) and `list*` forms for
  non-resource listings. Permissions MUST be linearly independent; bundled
  read-write composite permissions are forbidden. Internal platform operations
  MUST sit behind internal permissions not grantable to tenants. Permission
  catalogs MUST be declared as versioned fixtures in code, with typed
  constants generated for service integrations and a published catalog diff
  per release.
- **Acceptance evidence:** permission-fixture schema validation in CI;
  generated-constant drift check; per-service integration test proving an
  operation is denied without its exact permission; role/permission matrix
  diff artifact published per release.
- **Non-goals:** attribute-based or Rego-class general policy languages;
  user-defined custom permission verbs.
- **Non-claims:** cross-service permission consistency is enforced by review
  and code generation, not yet by one compiled platform-wide catalog.
- **Stop conditions:** a proposed wildcard, composite, or tenant-grantable
  internal permission halts merge; any new permission that mutates money,
  data, or keys requires security review before acceptance (trust, keys).
- **Traceability:** legacy-platform-b; current-core (policy decision refs in
  connector packages). Related: CR-IAM-030, CR-IAM-190.

### CR-IAM-030 — Service roles over primitive roles

- **Priority:** P1
- **Status:** proposed
- **Actors:** provider, tenant, service-team, operator
- **Problem:** Broad primitive roles (viewer/editor/admin) were a launch
  shortcut in reference platforms and became an unfixable over-privilege
  liability; retrofitting granularity onto them never succeeds.
- **Requirement:** The platform MUST ship granular service roles
  (`<service>.<resource>.<purpose>`) as the recommended grant unit, with
  role→permission mappings declared as reviewed fixtures and published diffs.
  Primitive coarse roles MAY exist only as deprecated compatibility roles,
  MUST NOT be required by any platform service, and MUST be excluded from
  new-service onboarding and from system-account grants on platform-root
  scopes.
- **Acceptance evidence:** role-fixture review gate in CI; static check that
  no new service or system account depends on primitive roles; documentation
  mapping common tasks to the recommended granular role; role-matrix diff per
  release.
- **Non-goals:** eliminating a small, documented starter role set for tiny
  installations; per-tenant custom role composition (federation-stage
  candidate).
- **Non-claims:** the full operator duty taxonomy is separate (CR-IAM-200);
  marketplace-vendor role packs are not yet specified.
- **Stop conditions:** granting a primitive role on a platform-root scope to a
  system or service account halts deployment until replaced by a minimal
  custom role (trust, exposure).
- **Traceability:** legacy-platform-b (primitive-role lesson); req-history.
  Related: CR-IAM-020, CR-IAM-200.

### CR-IAM-040 — Subject model: users, service accounts, groups, pseudo-subjects

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, service-team, operator
- **Problem:** Authorization, audit, and revocation need consistent subject
  types across human, machine, and collective actors, and every subject must
  remain interpretable in audit history after deletion.
- **Requirement:** The platform MUST support: user accounts; service accounts
  that are simultaneously subjects and resources bound to a container; groups
  with explicit membership and declared limits (per-organization group count,
  members per group, effective groups per subject); and pseudo-subjects for
  public-access classes (all users / all authenticated users) whose semantics
  are explicit and auditable. Platform-internal service accounts MUST use
  registered well-known IDs. Every subject MUST remain resolvable in audit
  records after deletion (tombstone identity). Anonymous write access MUST NOT
  exist.
- **Acceptance evidence:** subject CRUD and binding tests per kind; group
  limit enforcement tests; pseudo-subject expansion tests showing public
  access is explicit and enumerable; audit query resolving a deleted subject
  to its tombstone record.
- **Non-goals:** operating a consumer identity provider; unbounded group
  nesting.
- **Non-claims:** cross-organization groups are not supported; federated
  group/role mapping (SCIM-class) is not yet specified (see CR-IAM-100).
- **Stop conditions:** any pseudo-subject grant on money, data, or keys
  surfaces halts rollout; a group-limit bypass or an unresolvable audit
  subject is a release blocker (trust, exposure).
- **Traceability:** legacy-platform-b; legacy-platform-a (separate customer
  and internal identity realms); current-core (subject kinds in the
  authorization-decision contract). Related: CR-IAM-050, CR-IAM-150.

### CR-IAM-050 — Container membership join semantics guaranteeing full revocation

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, provider, operator, auditor
- **Problem:** Unless "holds any access inside the tree" is joined to "visible
  member at the root", revocation and member listing silently fail — a
  reference platform shipped exactly this broken guarantee and discovered it
  only in production.
- **Requirement:** Any subject holding any effective right inside a top-level
  container MUST hold a membership role (member or owner) on that container,
  maintained transactionally with binding changes. Membership MUST power
  "list my containers" and MUST enable complete revocation of a subject at
  the container root. The guarantee MUST be proven by tests, never assumed.
  Membership listing MUST NOT replace per-item `get` checks.
- **Acceptance evidence:** invariant test suite applying randomized binding
  mutations and verifying membership consistency; full-revocation drill
  (grant deep access, revoke membership at root, verify deny everywhere
  within the propagation bound); fault-injection test showing failed binding
  writes leave no orphan access.
- **Non-goals:** membership as an authorization shortcut for item-level
  reads; membership semantics across organizations.
- **Non-claims:** the guarantee under cross-container service-account grants
  additionally requires the choke-point permission of CR-IAM-190; behavior
  under deny-bindings is out of scope because deny-bindings do not exist.
- **Stop conditions:** a membership-invariant test failure halts any IAM
  release; a discovered revocation gap is a severity-1 security incident with
  mandatory audit of affected containers (trust, data).
- **Traceability:** legacy-platform-b (broken-membership incident lesson);
  req-history. Related: CR-IAM-010, CR-IAM-040, CR-IAM-190.

### CR-IAM-060 — Central authorization API with batch decisions and list semantics

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, tenant, operator
- **Problem:** Every service must make consistent, fast, testable decisions;
  a reference platform shipped without batch authorize and forced painful
  per-item fan-out into every service's list path.
- **Requirement:** The platform MUST provide a central authorization API
  `authorize(subject, permission, resource-path) → allow | deny | degraded`
  with typed denial reasons, plus a batch-authorize operation from the start.
  List semantics MUST filter by per-item `get`, MUST NOT pre-check the
  container to skip per-item checks, MUST NOT emit short pages, and MUST
  return NotFound for a missing container. Every decision MUST emit an audit
  envelope with correlation ID, and degraded MUST never be treated as allow.
- **Acceptance evidence:** conformance against the existing
  authorization-decision contract (subject/action/target/context envelope,
  typed outcomes, reason codes); batch-authorize load test at the defined
  latency SLO; list-semantics conformance tests shared by all services;
  audit-envelope linkage test per decision.
- **Non-goals:** per-tenant custom policy upload; a general policy DSL beyond
  role/permission evaluation.
- **Non-claims:** decision latency SLOs are design targets verified in test
  environments, not yet under production-scale multi-tenant load.
- **Stop conditions:** authorization API degradation beyond SLO halts
  dependent-service deployments; ambiguous or missing denial reasons on
  money, data, or keys operations block release (trust, data).
- **Traceability:** current-core (authorization-decision contract);
  legacy-platform-b (missing-batch-check lesson). Related: CR-IAM-120,
  CR-IAM-150, CR-IAM-160.

### CR-IAM-070 — Per-service access-binding APIs proxied to the IAM core

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, service-team, vendor, provider
- **Problem:** Tenants manage access on the resources they actually use, and
  there is deliberately no central resource registry — so each resource-owning
  service must expose ACL management itself, consistently.
- **Requirement:** Every resource-owning service MUST expose
  List/Set/UpdateAccessBindings that verify the caller's
  `listAccessBindings` / `updateAccessBindings` permission and proxy to the
  IAM core's private binding service. Binding changes MUST be atomic per
  resource, audited, and quota-bounded. Services MUST NOT maintain parallel
  private ACL stores for platform-managed access.
- **Acceptance evidence:** OCS connector conformance check requiring the
  binding surface; per-service contract tests (unauthorized caller denied,
  quota enforced, audit event emitted); static check finding no parallel ACL
  persistence in service code.
- **Non-goals:** a unified cross-service ACL browsing API (container
  membership listing is covered by CR-IAM-050).
- **Non-claims:** long-running-operation semantics for Set/Update are
  specified but not yet load-tested at scale.
- **Stop conditions:** a service shipping without the binding surface blocks
  its catalog publication; a binding change without an audit event halts
  release (trust, exposure).
- **Traceability:** legacy-platform-b; current-core (tenantAccess permissions
  in connector packages). Related: CR-IAM-020, CR-IAM-150.

### CR-IAM-080 — Short-lived token model with signed-JWT exchange and signer rotation

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, service-team, operator, agent
- **Problem:** Long-lived bearer credentials are the top credential-theft
  exposure; the platform needs one uniform, verifiable, short-lived token for
  all API access.
- **Requirement:** The platform MUST issue short-lived access tokens (default
  lifetime ≤12h) carrying principal (kind, ID, container binding for service
  accounts), scope, and expiry, signed by rotating platform signer keys.
  Service-account asymmetric keys MUST be exchanged through a signed JWT
  (PS256 preferred, RS256 restricted, enforced maximum JWT lifetime and fixed
  audience) for an access token at a token-exchange endpoint. Signer keys
  MUST rotate with overlap (reference: rotate every 24h, valid 48h) and MUST
  have a documented emergency revocation procedure. Verification MUST be
  centralized or use centrally published verification keys; unknown key IDs
  and restricted algorithms MUST deny.
- **Acceptance evidence:** token lifecycle test suite (issue, verify, expiry,
  overlap rotation, emergency revocation); negative tests (wrong audience,
  expired JWT, restricted algorithm, unknown key ID all deny); signer
  rotation drill evidence; fail-closed verification tests.
- **Non-goals:** mandating a token wire format to services (tokens are opaque
  to them); interactive consent flows (portal domain).
- **Non-claims:** the 12h default is a target not yet tuned by operator
  experience; hardware custody of signer keys is governed by CR-IAM-130 and
  not claimed here.
- **Stop conditions:** signer-key compromise triggers immediate rotation and
  forced re-issue with a drill, halting deploys until complete; acceptance of
  a restricted algorithm or over-long JWT lifetime is a release blocker
  (keys, trust).
- **Traceability:** legacy-platform-b; legacy-platform-a (signed access and
  service-account tokens). Related: CR-IAM-090, CR-IAM-230.

### CR-IAM-090 — Workload identity by default; static credentials exception-only

- **Priority:** P0
- **Status:** proposed
- **Actors:** tenant, service-team, vendor, operator
- **Problem:** Static credentials caused repeated privilege-escalation
  incidents in reference platforms; workloads need identity without stored
  secrets.
- **Requirement:** Workloads MUST obtain identity through platform
  mechanisms: virtual machines via a link-local metadata service issuing
  tokens for a bound service account; bare-metal hosts via a host-attested
  signing agent (TPM-class) with per-OS-user role bindings over a local
  socket that denies foreign users. Static credentials (API keys,
  S3-compatible access keys with request signing) MUST be exception-gated:
  documented approval, named owner, rotation procedure, expiry, and
  throttling/blacklisting support. Connector packages MUST reference secrets
  via workload identity; raw secret material is rejected by validation.
- **Acceptance evidence:** metadata-service integration tests (token issued
  for the bound service account only; spoofed identity denied); host-agent
  attestation tests including permission-denied for foreign OS users; the
  static-credential exception register with a rotation drill; validator
  rejection of raw secret material (existing core gate); production-image
  check that test-only attestation fallbacks are absent.
- **Non-goals:** customer-brought workload identity providers
  (federation-stage); universal S3-compatible signing beyond object-storage
  needs.
- **Non-claims:** TPM-class attestation coverage across all supported
  hardware is unverified; soft attestation paths are test-only and must be
  provably absent from production images.
- **Stop conditions:** a static-credential exception without expiry and
  rotation blocks deployment; metadata service exposure of another tenant's
  identity is a severity-1 incident (keys, trust, exposure).
- **Traceability:** legacy-platform-b; current-core (secrets
  workloadIdentityRef; source-safety); req-history (brokered secret
  capabilities for agents). Related: CR-IAM-140, CR-IAM-170.

### CR-IAM-100 — Identity federation and multi-factor authentication

- **Priority:** P1
- **Status:** proposed
- **Actors:** tenant, provider, operator
- **Problem:** Enterprise customers bring their own directories, and staff
  access needs phishing-resistant second factors; a provider platform cannot
  ship enterprise-ready with only local passwords.
- **Requirement:** The platform MUST support federated identity via SAML 2.0
  and OIDC/OAuth2 for tenant organizations and for the operator's staff
  realm, with cached federation metadata on the authentication path and
  explicit fail-closed behavior on metadata staleness. MFA MUST be available
  via TOTP at minimum and MUST be enforceable per organization and per role
  (staff mandatory). Step-up authentication for sensitive operations SHOULD
  be supported; WebAuthn-class phishing-resistant factors SHOULD be on the
  documented roadmap.
- **Acceptance evidence:** federation login/logout/refresh e2e tests per
  protocol against synthetic IdP fixtures; TOTP enrollment, verification, and
  recovery tests; federation-metadata staleness tests proving fail-closed
  behavior; MFA-required policy enforcement tests.
- **Non-goals:** building a consumer identity provider; social-login
  brokerage.
- **Non-claims:** WebAuthn-class factors are direction, not committed scope;
  federated group/role mapping (SCIM-class) is unspecified; sovereign or
  national-cloud identity modes are forward-looking and unproven.
- **Stop conditions:** federation metadata fetch or validation failure MUST
  deny new logins rather than fail open; any MFA bypass for staff or
  break-glass accounts is a release blocker (trust).
- **Traceability:** legacy-platform-b; legacy-platform-a (OIDC SSO library,
  separate customer and internal realms); vision-deck (multi-tenant provider
  vision). Related: CR-IAM-040, CR-IAM-110.

### CR-IAM-110 — Break-glass emergency access with expiring credentials and full audit

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Operators need emergency access when normal paths fail, but
  standing privileged access is an unacceptable exposure; the emergency path
  itself must be expiring, ticket-bound, and fully audited.
- **Requirement:** Emergency access MUST use short-lived (≤24h) credentials
  or certificates issued per incident ticket, with a named requester and
  approver, hardware-backed issuance where available, automatic expiry,
  source restrictions where applicable, and complete audit of issuance, use,
  and revocation. Standing emergency accounts MUST NOT exist. Any fallback
  long-lived emergency key MUST be sealed, monitored, and covered by a dated
  revocation drill. Every break-glass use MUST trigger an alert and a
  post-incident review.
- **Acceptance evidence:** issuance-to-expiry drill evidence; audit query
  showing who/when/why for every break-glass event; automated alert on any
  break-glass authentication; negative test proving expired break-glass
  credentials deny; post-incident review records linked to tickets.
- **Non-goals:** tenant-side account recovery (tenants own their identity
  recovery); self-service break-glass without human approval.
- **Non-claims:** hardware-token-backed issuance depends on operator
  procurement and is not verified for all installation profiles.
- **Stop conditions:** any unaudited or non-expiring emergency credential
  halts release; break-glass use without a linked incident ticket is itself
  a security incident requiring review (trust, keys, exposure).
- **Traceability:** legacy-platform-b (ticket-bound expiring-certificate
  workflow); req-history (emergency action class governance). Related:
  CR-IAM-150, CR-IAM-190.

### CR-IAM-120 — Authorization availability: read-path split and bounded caching

- **Priority:** P0
- **Status:** proposed
- **Actors:** operator, service-team, provider
- **Problem:** Authorization must survive control-plane write storms and
  database load; reference incident history shows auth-path database overload
  is the recurring existential failure of a cloud's identity stack.
- **Requirement:** The authentication/authorization read path MUST be
  separated from the control plane: independent scaling, read-consistent
  storage mode, and in-memory caches. Data-plane caching MUST obey: cache
  successes only, never denials; TTL ≤15s with refresh-ahead; no
  authentication or authorization caching inside control-plane components;
  caches MUST fail closed on corruption. The authorization service MUST have
  self-protecting load controls (bounded queues, concurrency permits, bounded
  parallel cache loads against storage) and MUST expose cache/storage
  staleness in its health signal.
- **Acceptance evidence:** soak test where control-plane writes saturate
  while the read path holds its latency SLO; cache-policy unit tests (denial
  never cached, TTL bound enforced); failure drills (slow storage engages
  limiters without error storms); health-endpoint staleness tests consumed by
  balancer checks.
- **Non-goals:** strongly consistent grant propagation on the data plane
  (eventual within the TTL is accepted); write-path caching.
- **Non-claims:** latency and availability SLOs are verified in test
  environments, not yet under production-scale load; incremental invalidation
  (segment-tree-class) is an optimization, not baseline behavior.
- **Stop conditions:** a control-plane write storm degrading the auth read
  path beyond SLO halts releases and triggers incident review; discovery of
  cached denials requires immediate fix plus audit of affected decisions
  (trust, data).
- **Traceability:** legacy-platform-b (overload incidents and caching rules);
  current-core (degraded outcome class). Related: CR-IAM-060, CR-IAM-150.

### CR-IAM-130 — Key management with envelope encryption and hardware root of trust

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, auditor
- **Problem:** All at-rest data protection and all secrets workflows depend on
  managed keys with a defensible root of trust; an ad-hoc key store collapses
  every downstream security claim.
- **Requirement:** The platform MUST provide a key management service with
  symmetric encrypt/decrypt APIs (with authenticated additional data), key
  versioning, per-key usage allow-lists of service accounts, and audited key
  use. Envelope/streaming encryption MUST be the standard library for
  large-object and backup encryption: chunked AEAD (AES-256-GCM default,
  ChaCha20-Poly1305 option), HKDF key derivation, per-chunk nonces, a
  self-describing metadata envelope, and range decryption. Master keys MUST
  be wrapped by hardware roots of trust (TPM/HSM class) across at least three
  failure domains, with separate backup root keys, a documented rewrap and
  recovery procedure, and multi-party authorization for root-key operations.
- **Acceptance evidence:** cryptographic interop and negative tests (tampered
  ciphertext or additional data rejected); key-version rotation tests;
  root-of-trust ceremony runbook plus drill evidence (unwrap from a quorum of
  domains; single-domain loss tolerated); per-key allow-list enforcement
  tests; key-use audit queries.
- **Non-goals:** confidential-computing/encryption-in-use claims (watchlist
  only); export-control build variants (deferred, P2 candidate); a
  customer-facing key marketplace.
- **Non-claims:** formal HSM certification (FIPS-class) is not claimed;
  throughput under production-scale request rates is unverified; root-KMS
  operations are partially manual by design and do not yet scale beyond a
  small dedicated host set.
- **Stop conditions:** any plaintext key material in logs, evidence, or git
  is a security incident halting the pipeline; a root-key ceremony without
  the required domain quorum or party approvals blocks key operations (keys,
  data, trust).
- **Traceability:** legacy-platform-b (root-of-trust design, streaming
  envelope library); req-history (key custody as a policy-visible dimension).
  Related: CR-IAM-140, CR-IAM-150.

### CR-IAM-140 — Secrets manager with auto-unseal and sealed-state honesty

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, tenant, operator, agent
- **Problem:** Charter principle — secrets are never configuration. Services
  need brokered, rotatable secrets whose availability state is reported
  honestly, because a secrets store that silently serves stale values (or
  fabricates availability) corrupts everything built on it.
- **Requirement:** The platform MUST provide a secrets manager (dedicated
  cloud-secrets-store class) with versioned secrets, access bound to workload
  identity, audit of every read, and usage reporting. Auto-unseal via the KMS
  MUST be supported so restarts require no human key shards. When sealed, the
  service MUST report sealed and MUST NOT serve stale or fabricated material;
  dependents MUST fail closed. Secret distribution to hosts, where required,
  MUST go through an ACL'd broker with group-scoped grants — never through
  config files, images, or documentation.
- **Acceptance evidence:** seal/unseal drill evidence (restart auto-unseals;
  loss of KMS permission seals the service and dependents deny); version and
  rotation tests with old-version denial; per-secret read-audit query;
  integration tests showing services receive references rather than raw
  material (existing core validator gate).
- **Non-goals:** a general-purpose configuration store; a plugin ecosystem
  beyond documented interfaces.
- **Non-claims:** third-party secrets-store integration (bring-your-own
  secrets-manager class) is validated only against the documented auto-unseal
  contract, not certified; usage-based economics of the secrets store are not
  yet modeled.
- **Stop conditions:** sealed-state misreporting (serving while sealed, or
  reporting unsealed when sealed) is a release blocker and an incident; any
  secret value appearing in logs, evidence, or docs triggers rotation and
  review (keys, trust, data).
- **Traceability:** legacy-platform-b (secrets store with KMS auto-unseal and
  permission-loss sealing); legacy-platform-a (secrets discipline and its
  documented pitfalls); current-core (reference-only secrets). Related:
  CR-IAM-090, CR-IAM-130.

### CR-IAM-150 — Immutable, queryable, SLA-bound audit log

- **Priority:** P0
- **Status:** proposed
- **Actors:** auditor, operator, provider, tenant
- **Problem:** Every security, billing, and trust claim depends on a
  tamper-evident record of who did what, when, and why — without it,
  incident response, dispute resolution, and compliance are all impossible.
- **Requirement:** Every mutating control-plane call, every authorization
  decision class (allow/deny/degraded with reason), every break-glass event,
  and every key and secret operation MUST emit a structured audit event:
  actor, represented subject, tenant/project, action, target, outcome,
  reason, rule reference, UTC timestamp, correlation ID. The audit store MUST
  be append-only and tamper-evident, queryable by subject, target, time, and
  correlation ID, retention-policy bound, and covered by availability and
  query-latency SLOs. Audit-pipeline failure MUST fail closed for operations
  that require audit (the degraded decision class). User authentication
  material in any log MUST be treated as a defined security-incident class
  with notification duties.
- **Acceptance evidence:** audit-coverage conformance tests per service
  (mutating operation → event); tamper-evidence verification test
  (hash-chain or WORM-class mechanism); query-latency SLO measurements;
  pipeline-failure drill (sink down → dependent mutations deny or degrade,
  never run silently); log-scrubbing gate evidence proving tokens, cookies,
  and passwords are absent from retained logs.
- **Non-goals:** a full SIEM/alerting product (hooks required here; product
  in the observability domain); customer-visible audit export formats (P2).
- **Non-claims:** multi-year retention economics are not yet modeled;
  settlement-grade cross-provider audit is federation-stage scope.
- **Stop conditions:** an unavailable audit sink halts or degrades affected
  mutating operations per contract — never silently unaudited operation; an
  authentication-material-in-logs finding triggers the incident class with
  credential rotation (trust, data, exposure).
- **Traceability:** current-core (audit envelope of the authorization-decision
  contract); legacy-platform-b (authentication log pipeline, control-plane
  audit destination, notification rule); req-history (audit gating). Related:
  CR-IAM-060, CR-IAM-110, CR-IAM-220.

### CR-IAM-160 — Fail-closed authorization and hidden management surfaces

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, operator, service-team, tenant
- **Problem:** On error, ambiguity, or missing evidence the platform must
  deny; exposing management surfaces before IAM allows is a catalogued
  anti-pattern that turns a configuration mistake into a breach.
- **Requirement:** Every authentication and authorization path MUST deny on
  error, timeout, unknown key or algorithm, ambiguous scope, or missing
  evidence. Management and provider console surfaces MUST remain hidden
  until IAM explicitly allows, and gateways MUST NOT expose internal
  endpoints publicly. Decisions MUST carry typed outcomes
  (allow/deny/degraded) with reason codes, and degraded MUST NOT be treated
  as allow. Public exposure of any new endpoint MUST be an explicit, reviewed
  declaration.
- **Acceptance evidence:** negative test matrix per service (expired token,
  unknown signer, missing scope, malformed request → deny); portal-shell
  conformance proving management stays hidden without an IAM allow (existing
  core contract); endpoint-exposure review gate plus scan over gateway
  surfaces; degraded-is-not-allow tests.
- **Non-goals:** denial-screen UX (portal domain); edge WAF/DDoS policy
  (network domain).
- **Non-claims:** endpoint inventory automation covers known gateway surfaces
  only; exposure scanning completeness is partial.
- **Stop conditions:** discovery of a fail-open path or an unreviewed public
  endpoint halts release and triggers security review (exposure, trust).
- **Traceability:** current-core (hidden-management rule; fail-closed
  conformance); legacy-platform-b; req-history (anti-patterns). Related:
  CR-IAM-060, CR-IAM-180.

### CR-IAM-170 — Tenant isolation test gate

- **Priority:** P0
- **Status:** proposed
- **Actors:** provider, tenant, operator, auditor
- **Problem:** Multi-tenancy is existential: isolation must be proven
  continuously, because a single cross-tenant read invalidates every trust
  claim the platform makes.
- **Requirement:** The platform MUST maintain an automated tenant-isolation
  suite proving one tenant cannot read, modify, enumerate, or infer another
  tenant's resources, tokens, secrets, metadata, usage, or audit records
  across every service and every shared substrate (API, metadata service,
  storage, network, caches, logs). The suite MUST include confused-deputy and
  cross-tenant service-account grant cases, MUST run in CI, and MUST run
  periodically against live stands with fresh receipts; failures block
  readiness.
- **Acceptance evidence:** the named isolation suite with a per-service
  coverage matrix; live-stand isolation run receipts within the freshness
  window of the identity-policy evidence class; regression history; negative
  cross-tenant fixtures in every connector conformance run.
- **Non-goals:** micro-architectural side-channel (Spectre-class) proof —
  documented as residual risk; noisy-neighbor performance isolation
  (capacity/observability scope).
- **Non-claims:** isolation is tested at the API/service boundary;
  hypervisor- and hardware-level isolation claims are delegated and not made
  here.
- **Stop conditions:** any cross-tenant data exposure is a severity-1
  incident; readiness claims are blocked until the isolation suite is green
  on the target stand (data, trust, exposure).
- **Traceability:** current-core (tenant-isolation contract); legacy-platform-b
  (user-network trust incidents); req-history. Related: CR-IAM-090,
  CR-IAM-160.

### CR-IAM-180 — Security scanning gates in CI/CD

- **Priority:** P0
- **Status:** proposed
- **Actors:** service-team, vendor, operator, provider
- **Problem:** Vulnerable dependencies, static-analysis findings, leaked
  secrets, and private-source leakage must be stopped before merge and
  release — scanning after the fact is incident response, not prevention.
- **Requirement:** CI MUST enforce blocking gates: Go static security
  analysis (gosec-class), dependency vulnerability scanning
  (govulncheck-class), container and artifact vulnerability scanning
  (trivy-class), secret scanning (gitleaks-class), and the platform
  source-safety scan (no credentials, tenant data, private endpoints, host
  paths, or copied private source). Exceptions MUST be scoped, expiring,
  owner-approved, and recorded. Coverage MUST include docs and examples, not
  only code, and MUST run pre-push as well as in CI.
- **Acceptance evidence:** CI workflow definitions demonstrating blocking
  behavior; negative fixtures proving each gate fires (planted test secret,
  known-vulnerable dependency, private-marker string); the exception register
  with expiry; pre-push hook evidence.
- **Non-goals:** runtime intrusion detection (observability/ops); dynamic
  scanning of tenants' own workloads.
- **Non-claims:** scanner rule tuning is ongoing and zero-false-negative
  guarantees are not made; artifact signing and provenance
  (SLSA-class supply-chain evidence) is roadmap, not claimed.
- **Stop conditions:** a disabled, skipped, or bypassed gate is a release
  blocker; a secret committed to any branch triggers rotation plus history
  review before further pushes (keys, exposure, trust).
- **Traceability:** current-core (existing security and source-safety
  workflows); legacy-platform-a (secrets-in-docs lessons); req-history
  (source-safety gate). Related: CR-IAM-160, CR-IAM-140.

### CR-IAM-190 — Privilege-escalation choke points as separate audited permissions

- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, auditor, service-team
- **Problem:** Token minting, impersonation, and cross-container grants are
  the crown jewels: bundled into admin roles they become invisible escalation
  paths — a reference platform shipped an arbitrary-service-account token
  issuance vulnerability through exactly this route.
- **Requirement:** Impersonation (creating tokens for another subject),
  cross-container service-account bindings, and cross-resource reference
  linking MUST each be separate, individually grantable, individually audited
  permissions (tokenCreator / crossCloudBindings / updateReferences class).
  Grants of these permissions MUST pass a review workflow, their use MUST
  emit high-visibility audit events, and platform services holding them MUST
  use dedicated well-known service accounts.
- **Acceptance evidence:** permission existence and grant tests; periodic
  audit query enumerating all impersonation and cross-container events;
  review-workflow records for every grant; negative tests proving an admin
  role without the choke-point permission is denied.
- **Non-goals:** just-in-time elevation automation (ops tooling may build it
  on these primitives).
- **Non-claims:** behavioral analytics over choke-point use is not claimed;
  the approval workflow is ticket-based process, not yet enforced in-product.
- **Stop conditions:** any token-minting path not gated by its dedicated
  permission is a severity-1 security incident; choke-point grants without a
  review record block release (trust, keys).
- **Traceability:** legacy-platform-b (token-issuance incident; choke-point
  permissions); req-history. Related: CR-IAM-020, CR-IAM-050, CR-IAM-200.

### CR-IAM-200 — Operator duty role taxonomy with impersonation-based administration

- **Priority:** P1
- **Status:** proposed
- **Actors:** operator, provider, auditor
- **Problem:** Operations and support staff need least-privilege access that
  encodes "look but don't touch user data" versus "modify", and routes admin
  power through auditable impersonation rather than standing personal
  privilege.
- **Requirement:** The platform SHOULD define per-service duty role families:
  on-call read (metadata and technical state, no user data), on-call admin
  (modifying; held only by a dedicated well-known service account that staff
  impersonate), support read, and support admin. Direct employee modification
  of tenant-facing identity or account records MUST be classified as a
  security incident. Duty access SHOULD be time-boxed with ticket linkage
  where tooling allows.
- **Acceptance evidence:** duty role fixtures per service family;
  impersonation-flow tests (staff → duty service account, fully audited);
  user-data access denial tests for read roles; access-review records showing
  duty grants with expiry and justification.
- **Non-goals:** full identity-governance (IGA-class) product integration;
  tenant-managed support-access delegation (tenant-side feature, later).
- **Non-claims:** taxonomy coverage across all OCS services is partial until
  service teams adopt it; time-boxing is process-level, not yet enforced by
  tooling.
- **Stop conditions:** any duty role holding undeclared user-data access
  blocks the service's publication; an un-audited impersonation path is an
  incident (trust, data).
- **Traceability:** legacy-platform-b (duty role families; impersonation-only
  administration); req-history. Related: CR-IAM-030, CR-IAM-190.

### CR-IAM-210 — Threat modeling and security review for every feature

- **Priority:** P1
- **Status:** proposed
- **Actors:** service-team, vendor, provider, operator
- **Problem:** Security claims without systematic analysis are marketing;
  reference practice requires every feature to pass threat modeling,
  including insider adversary classes.
- **Requirement:** Every feature touching identity, money, data, keys, or
  exposure MUST have a recorded threat model (STRIDE-class) covering external
  and insider adversary classes (staff, contractors, datacenter personnel)
  before acceptance. Outcomes MUST link to requirements and ADRs with
  explicit residual-risk statements. Managed services reachable from tenant
  networks MUST treat tenant-controlled network data (DHCP, DNS, routes) as
  untrusted by design.
- **Acceptance evidence:** threat-model records per feature with review
  sign-off; a release checklist gate linking current threat models;
  regression tests for known abuse cases (rogue DHCP/DNS class); the
  residual-risk register, dated and owned.
- **Non-goals:** formal verification; bug-bounty program operations (a
  business-layer decision).
- **Non-claims:** threat-model quality is review-dependent and not
  machine-verified; insider-threat detection is process plus audit, not
  automated.
- **Stop conditions:** a money, data, or keys feature without a current
  threat model blocks release; a modeled threat without mitigation or an
  accepted-risk record blocks readiness claims (trust, exposure).
- **Traceability:** legacy-platform-b (security handbook practice);
  req-history (ADR gating for high-risk dependencies). Related: CR-IAM-160,
  CR-IAM-170.

### CR-IAM-220 — Compliance mapping hooks (ISO/PCI-class)

- **Priority:** P1
- **Status:** proposed
- **Actors:** auditor, provider, operator
- **Problem:** Providers must answer ISO/PCI-class audits; compliance must be
  assembled continuously from living evidence, not rebuilt per audit from
  stale documents.
- **Requirement:** The platform SHOULD maintain machine-readable mappings
  from platform controls in this corpus to ISO-27001-class and PCI-DSS-class
  control identifiers, each mapped control linked to its evidence artifacts
  (tests, drills, runbooks) with freshness tracking. Mappings MUST state
  coverage and evidence — never certification. Per-service documentation
  requirements (architecture, data flows, networks and ports, user-input
  paths, admin interfaces) MUST be defined so audit artifacts are producible
  on demand.
- **Acceptance evidence:** the compliance-mapping document with
  control→evidence links, reviewed and dated; a documentation-completeness
  gate per service; a generated sample audit-evidence bundle; recorded
  non-claims for every unmapped or stale control.
- **Non-goals:** achieving certification within the OSS project;
  jurisdiction-specific regimes (GOST-class, HIPAA-class) beyond mapping
  hooks — community extensions.
- **Non-claims:** no certification is claimed or implied; mapping completeness
  is partial and review-gated; auditor acceptance is outside platform
  control.
- **Stop conditions:** compliance language implying certification without
  evidence is a publication blocker; evidence links stale beyond their
  freshness window downgrade the mapping to blocked (trust).
- **Traceability:** legacy-platform-b (documentation-as-compliance practice);
  current-core (evidence freshness contract); req-history. Related:
  CR-IAM-150, CR-IAM-180.

### CR-IAM-230 — Service-side token handling contract

- **Priority:** P2
- **Status:** proposed
- **Actors:** service-team, vendor, operator
- **Problem:** Services historically started with near-dead tokens and failed
  mid-request; a uniform handling contract avoids flaky authentication across
  dozens of services.
- **Requirement:** Platform service libraries SHOULD: refuse to start before
  the initial token is issued; refuse to start if initial validity is below
  15 minutes; use a token for at least 10% of its TTL before background
  refresh; and alert when remaining lifetime approaches 20%. Client libraries
  SHOULD expose log-safe credential fields so tokens never appear in logs.
- **Acceptance evidence:** a shared client library implementing the rules
  with unit tests (start refusal, refresh timing, alert threshold); a service
  integration checklist; log-safety test fixtures.
- **Non-goals:** enforcing the contract inside third-party services (advisory
  via the SDK).
- **Non-claims:** the thresholds are reference defaults, not yet tuned from
  production telemetry.
- **Stop conditions:** a platform service found logging tokens triggers the
  authentication-material incident class with rotation (keys).
- **Traceability:** legacy-platform-b (service token handling rules).
  Related: CR-IAM-080, CR-IAM-150.

### CR-IAM-240 — Cryptographic availability measurement

- **Priority:** P2
- **Status:** proposed
- **Actors:** operator, provider
- **Problem:** The KMS and secrets manager are platform-wide dependencies;
  unmeasured crypto SLOs are silent risk discovered only during an outage.
- **Requirement:** The platform SHOULD compute KMS and secrets-store SLIs
  (availability, latency) from request and audit logs with defined windows
  (reference: 5-minute buckets, an error-rate downtime threshold, a 30-day
  SLO window) and publish them to operators. SLO breaches SHOULD route to
  incident review.
- **Acceptance evidence:** SLI computation jobs validated against synthetic
  log fixtures; published dashboard or report artifacts; an SLO-breach alert
  drill.
- **Non-goals:** customer-facing SLO commitments (commercial layer);
  automatic remediation of crypto-path degradation.
- **Non-claims:** thresholds are reference values; production-scale validation
  is pending.
- **Stop conditions:** n/a (measurement surface; key-material risk is governed
  by CR-IAM-130 and CR-IAM-140).
- **Traceability:** legacy-platform-b (crypto SLI/SLA pipeline from audit
  logs). Related: CR-IAM-130, CR-IAM-140.

## Coverage notes

This domain deliberately defers:

- **Login, consent, and denial UX, session screens, and portal identity
  widgets** → `domains/19-portal-ux-selfservice.md`; this domain defines only
  the hidden-until-allowed and fail-closed contracts the portal consumes.
- **Metrics, tracing, alerting products, SIEM, and security-monitoring
  operations** (including fleet host-instrumentation policy) →
  `domains/20-observability.md` and `domains/21-ops-sre-support.md`; this
  domain requires the audit event stream and SLI hooks they consume.
- **Incident management, on-call duty process, postmortems, vulnerability
  response SLAs, and forced-update operations** → `domains/21-ops-sre-support.md`;
  here only the incident *classes* tied to identity (authentication material in
  logs, revocation gaps, cross-tenant exposure) are defined.
- **Network-edge defense: WAF, DDoS protection, firewall request governance,
  bastion host provisioning, and host hardening baselines (CIS-class image
  checks)** → `domains/12-network.md` and `domains/10-platform-foundation.md` /
  `domains/11-compute-virtualization.md`; this domain assumes those surfaces
  exist and consumes their evidence.
- **Cluster-level workload identity plumbing (Kubernetes service-account token
  projection, cluster admission policy)** → `domains/14-kubernetes-containers.md`;
  this domain owns the identity semantics they implement.
- **Encryption coverage per data service (which data classes are encrypted,
  backup encryption verification, restore drills)** → `domains/13-storage-backup-dr.md`
  and the data-service domains; this domain owns the KMS primitive and
  envelope library they must use.
- **Entitlement checks, commercial audit requirements, and settlement-grade
  cross-provider trust** → `domains/16-billing-finops.md` and
  `domains/23-federation-global-portal.md`; federation-stage identity
  (cross-provider subjects, federated audit) is explicitly out of scope here.
- **OCS connector security *surfaces* (what a service package must declare:
  workload identity refs, policy refs, tenantAccess permissions)** →
  `domains/17-ocs-service-connectors.md`; this domain defines the IAM
  platform those declarations bind to.
- **Agent approval matrices, risk classes, and brokered agent capabilities** →
  `domains/25-agent-governance.md`; this domain provides the token, audit, and
  secrets primitives agents are built on.
- **Artifact signing, build provenance, and supply-chain evidence
  (SLSA-class)** → `domains/22-deployment-iac-cicd.md`; the scanning gates
  here are a subset of that pipeline.
