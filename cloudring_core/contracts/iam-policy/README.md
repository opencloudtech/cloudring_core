# IAM And Policy Contract

CloudRING owns the public authorization contract between identity,
service modules, policy engines, audit sinks, and operator workflows. This
contract is intentionally portable: it defines the data required to ask for,
return, and audit an authorization decision without carrying private identity
state.

## Public Boundary

Core contract material may include:

- Subject references, tenant references, project references, actions, and
  targets.
- Allow, deny, and degraded decisions with typed reasons and policy rule
  references.
- Audit metadata required to prove the decision path.
- MFA, session age, reauthentication, support grant, and break-glass
  expectations.
- Synthetic examples that use symbolic references only.

Core contract material must not include:

- Bootstrap admin credentials or credential values.
- Private signing keys, private JWKS documents, token values, cookies, or
  session secrets.
- Deployment-specific issuer URLs, tenant endpoints, installation identity
  configuration, or real customer identifiers.
- Provider-specific identity wiring or live operational runbooks.

## Decision Model

Every authorization response has one of three outcomes:

- `allow`: the subject may perform the action on the target.
- `deny`: the request failed closed and the caller must not retry without a
  changed subject, target, credential, policy, or support context.
- `degraded`: the request is not allowed as a normal success path, but the
  policy engine can identify a recoverable control dependency such as an
  unavailable audit sink or stale session assurance.

Denied and degraded decisions must include an audit-safe reason code and the
policy rule that produced the outcome. They must not include raw credential
material or downstream identity-provider state.

## Request Contract

An authorization request identifies:

- `subject`: synthetic subject ID, subject kind, group references, credential
  class, MFA status, session assurance, and symbolic identity provider
  reference.
- `tenant`: tenant ID and lifecycle state relevant to the decision.
- `project`: project ID and tenant ownership.
- `action`: a dotted capability string such as `project.read`.
- `target`: resource kind, scope, and optional resource name.
- `context`: correlation ID, reason, support ticket reference, support grant,
  break-glass request, and decision time.

The request may state that a stronger session is required, but it must not
carry a bearer token, bootstrap credential, recovery code, or signing key.

## Audit Contract

Every decision must emit audit metadata with:

- Actor and represented subject references.
- Tenant, project, action, and target references.
- Decision outcome, reason code, policy rule, and timestamp.
- Correlation ID and optional ticket/support-grant references.
- MFA and session expectation results.

Audit entries record references and decisions only. Runtime audit storage,
identity-provider discovery, key rotation, bootstrap admin setup, and live
session validation remain outside CloudRING.

## Contract Artifacts

- `authorization-decision.schema.json` defines the portable request, response,
  and audit envelope.
- `examples/authorization-decision.synthetic.json` is a synthetic allow example
  using symbolic identity references.
