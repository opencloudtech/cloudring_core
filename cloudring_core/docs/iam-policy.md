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
