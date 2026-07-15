# OpenBao Kubernetes authentication plan and apply contract

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

The `plan` output is not an apply artifact. OpenBao 2.5.5 exposes GET and POST
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
mount, ACL policy, target role, seed metadata, and seed data. An absent KV
mount is the sole create-only platform initialization state: under the same
exclusive Lease the supervisor repeats the complete pre-state, enables exactly
`type=kv` with `options.version=2`, and requires the complete OpenBao 2.5.5
readback including a non-empty UUID and accessor. An existing non-KV-v2 mount,
any other missing prerequisite, or drift blocks before partial mutation. Every
write is also bound to the plan-wide auth-mount lifecycle mutation gate.

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
is proven to be KV version 2. A mount created by this operation is durable
platform initialization: automatic disable is forbidden because a concurrent
or later writer could make deletion destructive. Any subsequent failure is
therefore a partial result even when workload-specific auth objects are rolled
back safely. Login, capability denial, External Secrets
TokenRequest use, synchronization, rotation, revocation, and audit proof are
mandatory before readiness can be claimed.

External Secrets Operator 2.7.0 may fall back to a legacy service-account token
Secret if TokenRequest fails. Live verification must therefore prove the
TokenRequest path and the absence of a legacy token Secret; source RBAC alone is
not evidence of the runtime path. Rollback must never adopt or delete a
pre-existing auth mount.

## Executable apply path

Before running the stateful supervisor, render its temporary Kubernetes
executor boundary from a non-secret site profile:

```sh
go run ./cmd/cloudring-openbao render kubernetes-auth-executor \
  < ./contracts/openbao-kubernetes-auth/fixtures/synthetic-kubernetes-auth-executor.json
```

The renderer is the canonical source for all 10 temporary objects: the
executor and negative identities, an initially empty Lease, exact
resourceName-bounded Roles and RoleBindings, and the self-review
ClusterRole/Binding. It emits no Secret, Pod, Job, token, certificate, API
endpoint, or provider credential. The profile embeds the existing v2 auth
contract so a shared `kubernetes` mount or inconsistent positive/negative
identity relationship fails before YAML is produced.

For a real site, keep only that non-secret profile in the provider repository.
The workload namespace and a dedicated restricted negative namespace are
persistent consumer prerequisites; the renderer neither creates nor deletes
them. Prove both namespaces exist and that the negative namespace has the exact
provider ownership label
`cloudring.org/openbao-negative-identity: "true"` plus restricted Pod Security
labels before any executor operation. With pipeline failure propagation
enabled, pipe the accepted-core render through a server dry run
(`kubectl create --dry-run=server -f -`) only after proving every rendered
identity absent. Create without field adoption, capture the UID returned for
each successfully created object in memory, and immediately GET and verify its
exact rendered semantics. A collision or drift blocks the operation. A partial
multi-object create owns only objects whose successful create-response UID was
captured; matching names or content alone never prove ownership.

Cleanup is deliberately not `kubectl delete -f -`, because that is name-based
and can delete a concurrently replaced object. After a successful supervisor
run, first prove the Lease has the captured UID, has no holder, and every other
temporary object still has its captured UID and exact rendered semantics. Then
delete each of the 10 objects through the Kubernetes API with
`DeleteOptions.preconditions.uid` set to its captured UID. Deletion order is
mandatory and monotonically removes privilege: the ClusterRoleBinding and two
RoleBindings first; the ClusterRole and two Roles second; the three
ServiceAccounts third; and the already-proven-empty same-UID Lease last. Stop
immediately on any missing object, UID/content drift, non-empty Lease,
precondition conflict, delete failure, or partial manual intervention, leaving
the remaining boundary for recovery. Prove all 10 objects absent afterward.
Never delete either persistent namespace as part of executor cleanup, and
never hand-edit or copy the canonical generated YAML into a provider
implementation.

`cloudring-openbao supervise kubernetes-auth` is the public protected workflow
for one dedicated workload identity and one create-only KV-v2 secret. It
accepts a bounded v1 supervisor request only from an anonymous or named pipe.
The request contains the apply template, one initial-root credential, and
`changeAuthorized: true`; it does not contain a policy name, wrapper, child
credential, child accessor, or binding. The supervisor generates all of those
in memory, invokes the executor in-process, and cleans the temporary authority.
No secret-bearing intermediate JSON is written or emitted.

The supervisor verifies the exact OpenBao 2.5.5 initial-root profile, generates
a 128-bit-random `cloudring-bootstrap-*` policy name, proves it absent, and
creates the canonical target-specific management policy with CAS `-1`. It then
creates a 10-minute, non-renewable, no-default-policy service token through
`auth/token/create` with `no_parent: true`, an explicit maximum TTL of 15
minutes, and a 60-second response wrapper. The wrapper accessor and wrapped
child accessor remain in memory. After the executor returns, the supervisor
first removes and proves absence of the exact version-1 temporary policy, then
revokes and proves absence of both accessors. A token-create ambiguity is never
retried: removing the token's sole policy deauthorizes any unknown child, and
the result remains `partial_manual_intervention_required` for audit cleanup.
Root authority has only these fixed broker operations and is never passed to a
generic target-path API.

`cloudring-openbao apply kubernetes-auth` is the lower-level v1 executor for one
dedicated workload identity and one create-only KV-v2 secret. It accepts one
bounded JSON request on an anonymous or named stdin pipe and refuses a terminal
or regular file. Providers should use `supervise`; direct `apply` exists for a
protected broker that already implements the same policy, wrapping, accessor,
binding, and cleanup contract. The
request carries the validated v2 contract, pinned HTTPS authorities and CA
roots, a short-lived Kubernetes bearer token, a response-wrapped OpenBao
management token, one secret value, the pre-created Lease identity, two
negative-probe identities, and an execution binding. An already-authorized
protected operator supervisor assembles this request in memory.
The binding detects request drift inside that trust boundary; because the
request is self-contained, it is not a signature or standalone approval
receipt. Do not put the
request, any constituent value, or a rendered command containing it in a file,
argument, environment variable, terminal transcript, or evidence.

The wrapped OpenBao token must be an orphaned, non-renewable service token with no
`default` or `root` policy, a remaining TTL no greater than 15 minutes, and
exactly one management policy. The executor requires the unwrap lookup accessor
to equal the supervisor-recorded wrapped-child accessor, compares the complete
policy body, and checks the exact capabilities needed for this contract. The
policy includes only the target operations plus token lookup-self,
revoke-self, and capabilities-self; it grants no renew operation. A wrapped
root token is rejected, left valid, and reported as manual intervention with
the Lease held because its response wrapper has already been consumed. The
supervisor never uses root authority for a target mutation.

The apply request uses schema
`cloudring.openbao-kubernetes-auth-apply/v1`. Required groups are:

| Group | Required content |
| --- | --- |
| `contract` | One valid `cloudring.openbao-kubernetes-auth-plan/v2` contract. |
| `openBao` | HTTPS address, exact TLS server name, and base64-encoded CA PEM. It must not contain a bearer token. |
| `kubernetes` | HTTPS address, exact TLS server name, base64-encoded CA PEM, and a short-lived base64-encoded bearer token for the executor ServiceAccount. |
| `lease` | Namespace/name of a pre-created empty Lease and a unique run holder identity. |
| `executorIdentity` | The exact namespace and ServiceAccount authenticated by the Kubernetes bearer token. It must be distinct from the positive and both negative workload identities. |
| `managementPolicyName` | The sole policy on the wrapped management token. |
| `managementAccessor` | The wrapped child accessor recorded by the protected supervisor; it must equal the unwrapped token lookup accessor. |
| `wrappingTokenBase64` | A response-wrapping token; never a root or durable token. |
| `seed` | One safe relative KV-v2 path and one or more uniquely named base64-encoded values. |
| `negativeIdentities` | Existing wrong-ServiceAccount and wrong-namespace identities from the executor RBAC boundary. |
| `approval` | `changeAuthorized: true` plus the exact canonical request consistency binding computed by an already-authorized protected supervisor. |

The binding includes both API authorities, TLS server names, SHA-256 values of
both CA roots, and the management child accessor, but excludes rotating bearer
and wrapping token values. Unknown,
duplicate, trailing, insecure, unbound, or oversized input is rejected
before network mutation. Both API transports use only the supplied trust root,
TLS 1.2 or newer, an exact TLS server name, bounded bodies, timeouts, no proxy,
no redirects, and duplicate-response-field rejection. API response bodies and
dynamic provider errors never enter the report.

The executor performs this fixed transaction:

1. Verify the exact active, initialized, unsealed OpenBao 2.5.5 health and
   built-in plugin profiles; authenticate the bounded-lifetime projected
   Kubernetes JWT as the exact executor ServiceAccount and UID through
   `SelfSubjectReview`; verify the current executor ServiceAccount UID and
   exact standard groups;
   prove the required Lease, ServiceAccount, TokenRequest, and self-review
   permissions plus denial of broader Lease, Secret, TokenReview, arbitrary
   SubjectAccessReview, and unrelated-ServiceAccount access; then read all
   positive/negative ServiceAccount UIDs.
2. Acquire the pre-created Lease with a resourceVersion-guarded full update;
   never create, patch, delete, or automatically take over a non-empty Lease.
3. Unwrap and verify the exact non-root management token, child accessor,
   canonical policy body, self-service token controls, and target capabilities.
4. Capture all mount, config, role-inventory, KV mount, policy, role, metadata,
   and secret pre-state before the first write. If and only if the KV mount is
   absent, repeat that complete pre-state under the Lease, enable the exact
   KV-v2 mount create-only, and bind its UUID/accessor into subsequent checks.
5. Apply the dedicated mount/config, create-only CAS policy, exact role, and
   CAS-0 KV-v2 seed with pre-write rereads and exact OpenBao 2.5.5 post-write
   readbacks, including the complete KV metadata version and timestamp shape.
6. Request a 600-second projected token and prove the positive login, identity
   metadata, policy/TTL, allowed data read, wrong audience/ServiceAccount/
   namespace denial, sibling/metadata denial, and post-revoke denial.
7. Revoke the management token and release the Lease with its latest
   resourceVersion.

Before the production seed exists, rollback runs in reverse order and removes
only current-run workload-specific objects after exact ownership/readback
checks. A current-run-created shared KV-v2 mount is deliberately retained and
never disabled automatically. KV-v2 also has no atomic compare-and-delete for
metadata, so successful creation of the fixed production seed is a second
commit point: later failures never automatically delete that data. Such runs
return `partial_manual_intervention_required`, revoke temporary authority where
provable, preserve the durable state, and retain or release the Lease according
to whether mutation ownership is unambiguous. Any drift, ambiguous API result,
lost Lease, failed revocation, or failed cleanup has the same fail-closed
result; automatic Lease takeover is forbidden. A fully cleaned supervisor
pre-apply failure is `rolled_back`; the other terminal statuses are
`blocked_preflight` and `applied`.

The apply report contains fixed gate names and non-claims only. `applied`
proves the dedicated OpenBao workload identity plus one secret and its live
authorization boundary. It does not claim ESO `SecretStore` readiness,
`ExternalSecret` synchronization, rotation, recovery, or release qualification.
Those are the next consumer-specific vertical slice.

The synthetic executor profile deterministically reproduces
`consumer-example/bootstrap-executor.yaml`; the platform manifest verifier
rejects any byte drift from that public renderer. The generated example is a
golden test artifact and is not activated by the generic runtime stages.

The contract is aligned with the versions pinned by the CloudRING runtime
profile. Primary references are the [OpenBao Kubernetes auth
API](https://openbao.org/api-docs/auth/kubernetes/), [OpenBao 2.5.5 config
implementation](https://github.com/openbao/openbao/blob/v2.5.5/builtin/credential/kubernetes/path_config.go), [auth mount
API](https://openbao.org/api-docs/system/auth/), [ACL policy CAS
API](https://openbao.org/api-docs/system/policies/), [Kubernetes
RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/), and the
[ESO 2.7.0 Kubernetes auth
implementation](https://github.com/external-secrets/external-secrets/blob/v2.7.0/providers/v1/vault/auth_kubernetes.go).
