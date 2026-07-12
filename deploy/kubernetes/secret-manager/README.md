# CloudRING runtime secret-manager profile

This profile installs the generic CloudRING runtime secret boundary. It is
provider-neutral and contains no credentials, public endpoints, tenant data, or
site-specific storage classes.

The profile is intentionally split into three reconciliation stages:

1. `controllers` installs pinned trust-manager and External Secrets Operator
   controllers and their CRDs.
2. `runtime` creates a private CA/trust bundle and a three-member OpenBao Raft
   service with TLS, anti-affinity, retained data volumes, audit storage, and a
   disruption budget.
3. `store` publishes the `platform-secrets` `ClusterSecretStore` only after an
   operator has initialized and unsealed OpenBao, enabled Kubernetes auth,
   created the least-privilege `cloudring-external-secrets` role, and loaded the
   first versioned secret paths. The store is usable only by `ExternalSecret`
   objects in the privileged `external-secrets` namespace; it is not a tenant
   credential gateway.

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
- Join and unseal all three Raft members, enable an audit device, and verify one
  active plus two healthy standby members on different nodes.
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

The OpenBao, External Secrets, and trust-manager images are digest-pinned. Chart
versions remain explicit so a provider can review chart templates and image
changes independently.
