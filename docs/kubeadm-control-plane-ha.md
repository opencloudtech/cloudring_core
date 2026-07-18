# Kubeadm Control-Plane HA Contract

The public `pkg/kubeadm` package is the provider-neutral implementation behind
the CloudRING site-profile control-plane contract. It validates a concrete
upstream Kubernetes topology, renders deterministic kubeadm configuration, and
verifies captured stand state. It does not connect to hosts or apply the
rendered plan.

The standalone site CLI exposes both sides of that contract:

```sh
go run ./cmd/cloudring-site render-kubeadm --spec ./examples/kubeadm-bootstrap-spec.json
go run ./cmd/cloudring-site verify-kubeadm --inventory ./examples/kubeadm-stand-inventory.json
```

Both commands accept `-` instead of a path to read a single JSON document from
standard input. Input is size-bounded and decoded strictly: duplicate or
unknown fields, trailing documents, and unsafe semantic values fail closed.
Rendering is deterministic for the same validated specification. Verification
always writes the sanitized report; a blocked report uses exit code `3`.
Successful validation uses `0`, invalid input uses `1`, command-line usage
errors use `2`, and an internal encoding failure uses `4`.

`RenderStackedEtcdDualStackConfig` fails closed unless the input declares:

- at least three stacked-etcd control-plane nodes with valid IPv4 and IPv6
  advertise addresses and unique names and addresses;
- a stable API endpoint that is not any control-plane node address;
- resolved stable API IPv4 and IPv6 addresses that are not node addresses;
- unique serving-certificate SANs containing the endpoint host and those exact
  stable IPv4 and IPv6 addresses;
- the same endpoint for kubeadm and the CNI bootstrap client;
- a control-plane transport device contained in the CNI device set;
- dual-stack pod and Service CIDRs;
- a one-node-at-a-time certificate lifecycle with reconfiguration, rollback,
  and one-server-loss acceptance references.

The renderer validates DNS-1123 names, an exact Kubernetes semantic version,
the runtime socket, and interface names before structurally marshaling the
kubeadm YAML. The rendered `ClusterConfiguration` includes the validated
serving-certificate SANs. Join documents always use the stable endpoint. CNI
readiness metadata retains the same endpoint and the validated device set.
The returned bundle and its planned actions carry the exact reconfiguration,
rollback, and acceptance references. Operations are provider-neutral adapter
contracts rather than distribution-specific shell commands.

`VerifyUpstreamStand` checks captured state independently of the renderer. It
blocks when the endpoint or a declared stable API address is node-bound,
certificate SAN coverage is incomplete, node identities or addresses are
duplicated, the CNI endpoint differs, the control-plane transport device is
absent, the certificate rollout is not sequential, rollback or reconfiguration
evidence is missing, or the API failover exercise has not passed. Existing
upstream, dual-stack, etcd, DNS, disruption-budget, durability, and continuity
checks remain mandatory.

Provider adapters resolve the references declared by
`ProviderSiteProfile`, supply concrete values at runtime, perform mutations,
and collect sanitized evidence. A provider implementation must:

1. preflight every node and endpoint before mutation;
2. render and review the exact kubeadm configuration;
3. install the stable endpoint and CNI device set;
4. create or rotate serving certificates one node at a time;
5. stop and restore changed nodes on the first failed health check;
6. remove temporary certificate and rollback material after convergence;
7. remove the current endpoint holder and verify authenticated IPv4 and IPv6
   API requests through the surviving nodes.

The included examples are synthetic contract fixtures. Even when the synthetic
inventory produces a ready report, it is only an example of the verifier input
shape and is not live readiness evidence. The package and site profile are an
executable validation and rendering contract, not live deployment evidence. A
downstream installation may claim a successful release only after its adapter
has executed the referenced lifecycle and the independent stand verifier has
accepted freshly captured state.
