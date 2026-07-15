# IAM And Policy

OCSv3 modules declare permissions and entitlement references. The platform IAM
service decides whether a tenant, organization, team, or project may access a
surface.

Modules must declare:

- `tenantAccess.scope`;
- entitlements and entitlement refs;
- read/write permissions;
- policy decision refs;
- evidence refs for IAM decisions.

Denied states must be user-visible and explain the next action. A denied
decision is fail-closed and must not reveal provider-only surfaces. Provider-only
surfaces must stay hidden unless IAM allows them.

## Runtime packages

`pkg/iam` is the reusable authorization decision runtime. It covers
organization, tenant and project boundaries, tenant lifecycle, role and token
scope, expiring support grants, last-owner protection, ticket-and-reason-bound
break-glass, and an audit-sink boundary. A failed audit append converts an
otherwise allowed decision into a denial.

Owner-removal target validation and the last-owner guard are structural
invariants evaluated before interactive, API-token, support-grant, or
break-glass permission paths; no credential class can bypass them.

A subject ID or API-token reference is not authentication evidence. Every
`Policy` requires a trusted `AuthenticationVerifier`; the default without one
is deny. The verifier proves the referenced session or credential out of band
through `AuthorizeContext` transport context and returns the authenticated
subject, credential class, MFA result, and session assurance. Bearer material
never enters the authorization request or audit record. Policy then enforces
fresh/non-revoked sessions and MFA for privileged, support, and break-glass
actions and records the assurance result in the decision and audit event. The
configured policy maps, clock, verifier, and audit sink are trusted
initialization inputs and must not be mutated while the policy serves
concurrent requests.

Tenant recovery always requires the explicit break-glass credential class,
reason, and ticket; a platform-admin role alone cannot bypass that path.

`pkg/identity` validates HTTPS-bound OIDC configuration, publishes discovery
and JWKS documents, verifies asymmetric `RS256` and `ES256` JWTs, rejects
`none` and symmetric downgrade attempts, enforces issuer, audience, required
claims, lifetime and key-rotation windows, and provides secure-cookie and
session-bound CSRF primitives. Management visibility remains denied until
authentication, token validation, and IAM all allow it.

The verifier uses an explicit JWT profile: exact JOSE `typ`, JWT-class claim,
authorized party, audience, issuer, and asymmetric algorithm. It rejects
duplicate JSON fields, ambiguous audience/authorized-party combinations, and
tokens issued before key activation or after key retirement. Rotation overlap
keeps a retired public key available for the maximum token lifetime plus
verification skew, while each successor key is published for one full JWKS
cache TTL before it may sign. Signing windows and token lifetime are exact;
clock skew is used only when comparing verified claims with verifier time.

`cmd/cloudring-id` exposes two provider-neutral operator paths:

```sh
go run ./cmd/cloudring-id contract \
  --issuer https://id.example.invalid \
  --audience cloudring-console

go run ./cmd/cloudring-id verify-token \
  --issuer https://id.example.invalid \
  --audience cloudring-console \
  --jwks ./operator-input/jwks.json \
  --token-file ./operator-input/token.jwt
```

The verifier reads JWT material from a bounded protected regular file rather
than argv; on Unix the file must be owner-only and final-component symlinks,
FIFOs, and devices are rejected. It writes sanitized optional evidence through
a stable trusted directory handle with an atomic no-overwrite publish. Evidence
creation fails for an existing destination, a symlinked or writable parent, or
a final-component symlink. It reports validation booleans and claim counts, not
identity values, groups, namespaces, JWTs, or private issuer URLs.

`cloudring-id contract` is explicitly a synthetic in-process contract check.
Its status is `contract-valid`, `syntheticOnly` is true, and
`readinessClaimed` is false; it neither contacts nor approves an installation.
Deployment-specific keys, JWTs, bootstrap credentials, session stores, and
live evidence remain outside the public repository.

## Operational boundary

These packages provide reusable in-process authorization and identity
validation primitives. They are not a claim that an identity provider,
distributed policy service, or durable regional account database is already
deployed. A production consumer must supply durable audit and session storage,
external-secret-backed bootstrap references, issuer/JWKS retrieval with bounded
network policy, HA, observability, backup/restore, and the live
allow/deny/revocation evidence required by its installation profile.
