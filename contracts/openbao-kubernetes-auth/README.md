# OpenBao Kubernetes authentication plan contract

This contract lets a provider describe one least-privilege OpenBao workload
identity without embedding credentials, endpoints, tenant data, or executable
operations in Git. The Go planner validates the complete v1 profile and emits
only fixed public policy facts, operation classes, required live gates, and
explicit non-claims:

```sh
go run ./cmd/cloudring-openbao plan kubernetes-auth \
  < ./contracts/openbao-kubernetes-auth/fixtures/synthetic-kubernetes-auth-bootstrap.json
```

Input is accepted only on stdin. The sanitized stdout report contains neither
raw identifiers nor identifier-derived fingerprints: ordinary hashes of
predictable names can be reversed by enumeration. Consequently the report is
not target-binding approval evidence; a later apply workflow must keep its
exact protected receipt inside the operator trust boundary. The contract has
no fields for an OpenBao
admin token, Kubernetes bearer token, reviewer JWT, CA PEM, client Secret, API
address, or HTTP header. Unknown and duplicate fields fail closed. Use only
public or synthetic identifiers in version-controlled examples; supply real
provider/site contracts through a protected non-logging operator channel.

The fixed CloudRING v1 profile uses the in-cluster Kubernetes service DNS,
OpenBao's rotating pod-local service-account token and CA, one exact Kubernetes
service account and namespace, audience `openbao`, UID aliases, a read-only
KV-v2 prefix, no default policy, and 10-minute/30-minute token TTL limits. These
TTLs are CloudRING policy, not OpenBao defaults.

The output is a plan, not an apply artifact. A live operator must still capture
pre-state, verify the auth mount type, prove that no static reviewer JWT is
stored, validate TokenReview from the OpenBao server service account, use CAS
for a new ACL policy, stop on role or policy drift, capture rollback data, and
perform exact post-write readbacks. Kubernetes auth config and role APIs have
no CAS, so their preflight, write, and readback must run under one exclusive
operator lock. The ACL policy path must not be created until the referenced
secret mount is proven to be KV version 2. Login, capability denial, External
Secrets TokenRequest use, synchronization, rotation, revocation, and audit
proof are mandatory before readiness can be claimed.

External Secrets Operator 2.7.0 may fall back to a legacy service-account token
Secret if TokenRequest fails. Live verification must therefore prove the
TokenRequest path and the absence of a legacy token Secret; source RBAC alone is
not evidence of the runtime path. Rollback must never blindly delete a shared
auth mount.

The contract is aligned with the versions pinned by the CloudRING runtime
profile. Primary references are the [OpenBao Kubernetes auth
API](https://openbao.org/api-docs/auth/kubernetes/), [auth mount
API](https://openbao.org/api-docs/system/auth/), [ACL policy CAS
API](https://openbao.org/api-docs/system/policies/), [Kubernetes
RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and the
[ESO 2.7.0 Kubernetes auth
implementation](https://github.com/external-secrets/external-secrets/blob/v2.7.0/providers/v1/vault/auth_kubernetes.go).
