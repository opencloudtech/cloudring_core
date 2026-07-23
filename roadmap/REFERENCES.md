# Primary technical references

The roadmap uses maintained standards and primary project documentation. These
links are design inputs, not dependencies that CloudRING must copy wholesale.
Versions are pinned by the goal that implements them and are re-evaluated through
an ADR when the ecosystem changes.

## Control planes, Kubernetes and GitOps

- [Kubernetes controller pattern](https://kubernetes.io/docs/concepts/architecture/controller/)
  — desired-state reconciliation and simple, owned control loops.
- [Cluster API](https://cluster-api.sigs.k8s.io/) — declarative create, scale,
  upgrade and delete lifecycle for conformant Kubernetes clusters.
- [Gateway API](https://gateway-api.sigs.k8s.io/docs/concepts/api-overview/) —
  role-oriented, portable L4/L7 routing APIs.
- [OpenGitOps principles](https://opengitops.dev/) — declarative, versioned,
  pulled and continuously reconciled desired state.
- [Kubernetes API Priority and Fairness](https://kubernetes.io/docs/concepts/cluster-administration/flow-control/)
  — tenant/system fairness and overload protection.

## Product and extension contracts

- [OpenAPI Specification](https://spec.openapis.org/oas/) — language-neutral
  synchronous HTTP contracts and generated tooling.
- [CloudEvents](https://cloudevents.io/) and
  [AsyncAPI 3](https://www.asyncapi.com/docs/reference/specification/v3.0.0) —
  portable event envelopes and machine-readable asynchronous APIs.
- [OCI Distribution Specification](https://github.com/opencontainers/distribution-spec)
  and [OCI artifact guidance](https://github.com/opencontainers/image-spec/blob/main/manifest.md)
  — content-addressed product packages and attached metadata.
- [Sigstore verification](https://docs.sigstore.dev/cosign/verifying/verify/) —
  identity and digest-bound verification for release artifacts.
- [Crossplane package/provider concepts](https://docs.crossplane.io/latest/packages/)
  — useful lessons for independently shipped provider extensions; CloudRING does
  not inherit Crossplane's runtime or API by default.
- [Google API Improvement Proposals](https://google.aip.dev/) and
  [long-running operations](https://google.aip.dev/151) — consistent
  resource-oriented APIs and asynchronous operation resources.

## Identity and trust

- [OpenID Connect specifications](https://openid.net/developers/specs/) — human
  identity, discovery, registration, logout and federation profiles.
- [WebAuthn Level 3](https://www.w3.org/TR/webauthn-3/) — phishing-resistant
  public-key authentication and passkeys.
- [SPIFFE standard](https://spiffe.io/docs/latest/spiffe-specs/) and
  [SPIRE concepts](https://spiffe.io/docs/latest/spire-about/spire-concepts/) —
  portable, attested, short-lived workload identity and trust federation.
- [Common Expression Language](https://github.com/google/cel-spec) — bounded,
  portable policy conditions rather than a custom Turing-complete DSL.

## Data, storage and continuity

- [CloudNativePG automated failover](https://cloudnative-pg.io/documentation/current/failover/)
  and [replica clusters](https://cloudnative-pg.io/documentation/current/replica_cluster/)
  — PostgreSQL HA, failover and disaster-recovery tradeoffs.
- [Rook/Ceph storage architecture](https://www.rook.io/docs/rook/latest/Getting-Started/storage-architecture/)
  — operator-managed block, file and S3-compatible storage with production
  failure-domain requirements.
- [KubeVirt live migration](https://kubevirt.io/user-guide/compute/live_migration/)
  — VM continuity constraints, strategies and failure behavior.
- [Cilium LoadBalancer IPAM](https://docs.cilium.io/en/stable/network/lb-ipam/)
  and [BGP control plane](https://docs.cilium.io/en/stable/network/bgp-control-plane/bgp-control-plane-configuration/)
  — dual-stack address allocation and routed service advertisement.

## Observability, cost and hyperscale reliability

- [OpenTelemetry observability primer](https://opentelemetry.io/docs/concepts/observability-primer/)
  — correlated traces, metrics and logs tied to user-visible reliability.
- [FOCUS specification](https://focus.finops.org/focus-specification/) — portable
  cost and usage export taxonomy.
- [Deployment stamps/cells](https://learn.microsoft.com/en-us/azure/architecture/patterns/deployment-stamp)
  — bounded scale units and failure isolation.
- [Google SRE guidance on cascading failures](https://sre.google/sre-book/addressing-cascading-failures/)
  and [overload handling](https://sre.google/sre-book/handling-overload/) — load
  shedding, bounded retries, backoff, capacity tests and graceful degradation.

## Research hygiene

External platforms and local historical codebases may be inspected for general
lessons only. Do not copy their code, documentation, private identifiers,
distinctive wording or proprietary implementation. Every CloudRING implementation
must be original, justified against current public standards and reviewed for
licence and source-boundary compliance.
