# OpenBao Kubernetes authentication plan contract

This contract lets a provider describe one least-privilege OpenBao workload
identity without embedding credentials, endpoints, tenant data, or executable
operations in Git. The Go planner validates the complete v2 profile and emits
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

The fixed CloudRING v2 profile requires
`authMountOwnership: dedicated-create-owned` and rejects the ambiguous generic
mount name `kubernetes`. Every independently managed workload boundary gets a
dedicated mount name. The profile uses the in-cluster Kubernetes service DNS,
OpenBao's rotating pod-local service-account token and CA, one exact Kubernetes
service account and namespace, audience `openbao`, UID aliases, a read-only
KV-v2 prefix, no default policy, and 10-minute/30-minute token TTL limits. These
TTLs are CloudRING policy, not OpenBao defaults.

The output is a plan, not an apply artifact. OpenBao 2.5.5 exposes GET and POST
for Kubernetes auth config, but no operation that restores an existing mount
to an absent-config state. The planner therefore defines this exact lifecycle:

| Pre-state under the exclusive lock | Decision |
| --- | --- |
| Mount absent | Create the dedicated mount and config; rollback may disable only the mount created by this execution after exact ownership readback. |
| Mount, contract-bound description, config, and role inventory exact, with no foreign roles | Reuse without mutating mount or config; rollback must never disable it. |
| Existing mount with config absent | Block before any mutation. |
| Existing config incomplete or drifted | Block before any mutation. |
| Existing mount has another type or description | Block before any mutation. |
| State is unknown, incomplete, unavailable, or changes during the run | Block before any mutation. |

The config POST has an explicit `auth-mount-created-by-current-run` mutation
guard. The desired mount description binds the dedicated mount to the exact
role name, and the pre-state includes the complete role inventory. Config
absence alone never authorizes a write. Before the first write, the ordered
plan reads the auth mount, complete auth config, auth role inventory, KV-v2
mount, ACL policy, and target role. Any missing prerequisite or drift blocks
before partial mutation. Every write is also bound to the plan-wide auth-mount
lifecycle mutation gate.

Post-create mount readback must match both type and contract-bound description.
Config readback must match the complete desired state and explicitly prove
`token_reviewer_jwt_set: false`; the write-only `token_reviewer_jwt` field is
never used as a substitute for that read-only fact. A live operator must
still capture pre-state, prove that no static reviewer JWT is stored, validate
TokenReview from the OpenBao server service account, use CAS for a new ACL
policy, stop on role or policy drift, capture rollback data, and perform exact
post-write readbacks. Kubernetes auth config and role APIs have no CAS, so all
preflight, write, and readback steps must run under one exclusive operator
lock. The ACL policy path must not be created until the referenced secret mount
is proven to be KV version 2. Login, capability denial, External Secrets
TokenRequest use, synchronization, rotation, revocation, and audit proof are
mandatory before readiness can be claimed.

External Secrets Operator 2.7.0 may fall back to a legacy service-account token
Secret if TokenRequest fails. Live verification must therefore prove the
TokenRequest path and the absence of a legacy token Secret; source RBAC alone is
not evidence of the runtime path. Rollback must never adopt or delete a
pre-existing auth mount.

The contract is aligned with the versions pinned by the CloudRING runtime
profile. Primary references are the [OpenBao Kubernetes auth
API](https://openbao.org/api-docs/auth/kubernetes/), [OpenBao 2.5.5 config
implementation](https://github.com/openbao/openbao/blob/v2.5.5/builtin/credential/kubernetes/path_config.go), [auth mount
API](https://openbao.org/api-docs/system/auth/), [ACL policy CAS
API](https://openbao.org/api-docs/system/policies/), [Kubernetes
RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and the
[ESO 2.7.0 Kubernetes auth
implementation](https://github.com/external-secrets/external-secrets/blob/v2.7.0/providers/v1/vault/auth_kubernetes.go).
