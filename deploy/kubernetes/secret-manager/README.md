# CloudRING runtime secret-manager profile

This profile installs the generic CloudRING runtime secret boundary. It is
provider-neutral and contains no credentials, public endpoints, tenant data, or
site-specific storage classes.

The profile is intentionally split into three reconciliation stages:

1. `controllers` installs pinned trust-manager and External Secrets Operator
   controllers and their CRDs, disables cluster-wide service-account token
   creation, and grants the controller permission to request only its own
   service-account token.
2. `runtime` creates a private CA/trust bundle and a three-member OpenBao Raft
   service with TLS, anti-affinity, retained data volumes, a declarative local
   audit device backed by audit storage, and a disruption budget.
3. `store` publishes the `platform-secrets` `ClusterSecretStore` only after an
   operator has initialized and unsealed OpenBao and the protected supervisor
   has create-only initialized or exactly reused the shared KV-v2 mount,
   enabled dedicated Kubernetes auth, created the least-privilege
   `cloudring-external-secrets` role, and loaded the first versioned secret
   paths. The store is usable only by `ExternalSecret`
   objects in the privileged `external-secrets` namespace; it is not a tenant
   credential gateway.

`consumer-example` is a validated, non-default template for a service trust
boundary. It is not part of the three-stage reconciliation chain. It contains
no `ExternalSecret` or target `Secret`, and applying it does not configure
OpenBao or prove that synchronization works.

Reconcile the stages as separate Flux `Kustomization` objects with explicit
`dependsOn` ordering. Do not collapse them into one apply: `Bundle`,
`ClusterSecretStore`, and the OpenBao workload depend on controllers or
operator bootstrap state that must already exist.

The default private CA is a reproducible bootstrap profile. A production
provider may replace it with an approved external issuer, but must preserve the
separate `openbao-client-ca` trust bundle. Clients must never read the serving
certificate Secret because it also contains the server private key.

## Mandatory operator gates

- Patch durable storage class, volume size, topology, and encryption policy for
  the installation before promotion.
- Initialize exactly one OpenBao member through a non-logging operator channel;
  distribute unseal/recovery shares to independent custodians and never store
  them in Git, Kubernetes manifests, command arguments, or evidence.
- Join and unseal all three Raft members, then verify one active plus two healthy
  standby members on different nodes.
- Verify that the declarative `audit "file" "persistent"` device writes to the
  retained audit volume before enabling clients. This local device is only the
  bootstrap audit baseline; a second independent sink, log rotation proof, and
  fail-closed audit failure proof remain production promotion gates.
- Keep StatefulSet pod management `Parallel` so all three sealed members can be
  bootstrapped without an `OrderedReady` deadlock. Keep the advertised HA API
  address on the TLS-covered `openbao-active.openbao.svc` service so standby
  redirects never expose an unverifiable Pod IP.
- Configure Kubernetes auth with a bounded audience and a role restricted to
  the External Secrets service account. Do not use a static OpenBao token.
- Give every tenant or service its own namespaced `SecretStore`, dedicated
  Kubernetes service account, OpenBao auth role and backend path policy. Bind
  that role to the exact service-account namespace/name and the `openbao`
  audience; do not reuse the privileged `platform-secrets` trust domain.
- Keep the readiness probe on the pod-specific
  `<pod>.openbao-internal.openbao.svc` name. The Flux post-renderer replaces the
  chart's insecure default probe and requires the mounted CA plus an explicit
  TLS server name; `-tls-skip-verify` is forbidden.
- Prove ExternalSecret synchronization, rotation, revocation, denied
  cross-namespace access, Raft snapshot export to an off-cell immutable target,
  and a real snapshot restore before producing readiness evidence.
- Record only sanitized status, hashes, counts, timestamps, and opaque object
  identities. A rendered manifest or a Ready HelmRelease is not secret-manager
  readiness.

## Kubernetes authentication ownership contract

The pinned External Secrets Operator 2.7.0 native OpenBao provider does not
support Kubernetes authentication. Until that support is released and passes
parity tests, the profile deliberately uses the Vault provider as a versioned
OpenBao compatibility seam. Every supported OpenBao and External Secrets
version pair still requires live positive and negative compatibility proof.

The controller chart's cluster-wide `serviceaccounts/token` permission is
explicitly disabled. Each consumer namespace must instead contain one `Role`
that permits `create` on `serviceaccounts/token` for exactly one
`resourceName`, plus one `RoleBinding` whose only subject is the External
Secrets controller service account. The login service account must set
`automountServiceAccountToken: false`, must not reference a legacy token Secret,
and must not receive `system:auth-delegator`.

The OpenBao server service account is the only TokenReview delegate. Give each
independently managed trust boundary a dedicated, non-generic Kubernetes auth
mount and configure it to use the server pod's rotating local service-account
token and CA; do not persist a reviewer JWT. The platform store uses
`kubernetes-platform-secrets`; the synthetic consumer uses
`kubernetes-consumer-example`. A site may choose different dedicated names,
but must never collapse them to the shared `kubernetes` mount. For every
consumer, create a separate policy and role with:

- one exact service-account name and namespace, with no wildcards;
- the exact `openbao` audience used by the `SecretStore`;
- `serviceaccount_uid` as the alias source;
- no default policy and short, bounded token TTLs;
- read-only KV v2 access to one service prefix, adding metadata/list access only
  when the service contract requires it.

The `consumer-example` `SecretStore` intentionally omits a namespace override
from `serviceAccountRef`: its identity is local to its own namespace. Its CA is
the non-secret `openbao-client-ca` ConfigMap distributed by trust-manager, never
a serving-key Secret. Site overlays own the real dedicated auth-mount name,
role, namespace, service-account identity, and backend prefix.

Source presence is not activation. Before a service uses this template, an
operator must apply the v2 dedicated-create-owned lifecycle: configure an
absent auth config only on a mount created by that same execution, and block an
existing mount whose config is absent, incomplete, or drifted. The operator
must then install the policy and role,
prove successful login and exact capabilities, prove denial for wrong audience,
service account, namespace, and backend path, wait for `SecretStore` and a
synthetic canary `ExternalSecret` to become Ready, and verify rotation and
revocation. None of those live gates is claimed by this profile.

Use the Go renderer, plan, and protected supervisor workflow in
`contracts/openbao-kubernetes-auth/README.md`. The renderer turns one
non-secret site profile into the exact temporary executor identities, empty
Lease, and resourceName-bounded RBAC. The workload and negative namespaces are
persistent consumer prerequisites and are never part of temporary cleanup;
the negative namespace has an exact ownership label and restricted Pod Security
labels. A provider stores only that profile, performs a create-only server dry
run, captures each successful create-response UID, and verifies exact live
semantics. Cleanup requires an empty captured-UID Lease followed by individual
Kubernetes API deletes with UID preconditions; name-based `kubectl delete -f`
is forbidden. Bindings must be deleted before Roles, ServiceAccounts after all
RBAC, and the empty same-UID Lease last; any failure stops cleanup. The provider
must not copy or hand-edit the generated RBAC. The tracked
`consumer-example/bootstrap-executor.yaml` is only the byte-verified golden
output for the synthetic profile and is not a deployment input for real sites.

The planner validates the typed contract without network access. The
supervisor creates and cleans the exact temporary OpenBao delegation in memory;
the apply executor uses the rendered Lease and RBAC to perform pre-state
capture, conditional mutation, exact readback, live authorization tests, and
fail-closed recovery. Production seed creation is a commit point because
OpenBao KV-v2 has no CAS metadata delete. Its `applied` status proves only the
dedicated workload identity and one KV-v2 value; it does not replace the ESO
synchronization, rotation, recovery, audit, and production gates above.

The OpenBao, External Secrets, and trust-manager images are digest-pinned. Chart
versions remain explicit so a provider can review chart templates and image
changes independently.
