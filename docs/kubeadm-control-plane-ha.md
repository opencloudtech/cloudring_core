# Kubeadm Control-Plane HA Contract

The public `pkg/kubeadm` package is the provider-neutral implementation behind
the CloudRING site-profile control-plane contract. It validates a concrete
upstream Kubernetes topology, renders deterministic kubeadm configuration, and
verifies captured stand state. It does not connect to hosts or apply the
rendered plan.

The standalone site CLI exposes both sides of that contract:

```sh
go run ./cmd/cloudring-site render-kubeadm --spec ./examples/kubeadm-bootstrap-spec.json
go run ./cmd/cloudring-site verify-kubeadm \
  --inventory /protected/run/kubeadm-stand-inventory.json \
  --one-server-loss-receipt /protected/run/unsigned-one-server-loss-receipt.json
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
  advertise addresses, an odd replica count, an unavailable-server envelope
  of exactly one, and an exact node-count match;
- unique DNS-1123 node names and addresses plus Kubernetes-valid labels and
  `key[=value]:Effect` taints (`NoSchedule`, `PreferNoSchedule`, or
  `NoExecute`);
- a stable API endpoint that is not any control-plane node address;
- resolved stable API IPv4 and IPv6 addresses that are not node addresses;
- unique serving-certificate SANs containing the endpoint host and those exact
  stable IPv4 and IPv6 addresses;
- the same endpoint for kubeadm and the CNI bootstrap client;
- a control-plane transport device contained in the CNI device set;
- dual-stack pod and Service CIDRs;
- a one-node-at-a-time certificate lifecycle with reconfiguration, rollback,
  and one-server-loss acceptance references.

The renderer validates DNS-1123 names, Kubernetes label and taint syntax, an
exact Kubernetes semantic version, the runtime socket, and interface names
before structurally marshaling the kubeadm YAML. Taint values are rendered
separately from their keys. The rendered `ClusterConfiguration` includes the
validated serving-certificate SANs. Join documents always use the stable
endpoint. CNI readiness metadata retains the same endpoint and the validated
device set. The returned bundle carries the validated unavailable-server
envelope, and its planned actions carry the exact reconfiguration, rollback,
and acceptance references. Operations are provider-neutral adapter contracts
rather than distribution-specific shell commands.

`VerifyUpstreamStand` checks captured state independently of the renderer. It
blocks when the endpoint or a declared stable API address is node-bound,
certificate SAN coverage is incomplete, node identities or addresses are
duplicated, the CNI endpoint differs, the control-plane transport device is
absent, the certificate rollout is not sequential, rollback or reconfiguration
evidence is missing, or the API failover exercise has not passed. Existing
upstream, dual-stack, etcd, DNS, disruption-budget, durability, and continuity
checks remain mandatory.

The verifier also requires the exact one-server-loss receipt as a separate
owner-protected input. It validates every receipt digest and timeline and binds
the receipt hash, run nonce, target Node UID hash, kubectl executable hash, and
data-probe adapter hash to the captured stand inventory. The bootstrap contract
still declares an intended `surviveUnavailableServers: 1` envelope, but the
captured stand inventory has no authoritative caller-declared survive count.
The deprecated field is accepted only for compatibility, ignored, and omitted
from the verifier's observed output. A ready stand report derives
`verifiedSurviveUnavailableServers: 1` only from a receipt whose
loss samples prove exactly one fewer ready control-plane Node, etcd member, and
API-server member than the protected baseline. Quorum after two simultaneous
losses is insufficient. The stacked-etcd control-plane count must be odd and
the node inventory count must match it exactly.

## Sequential HA expansion waves

`BuildHAWavePlan` and `VerifyHAWave` provide a separate, provider-neutral
contract for correcting a one-member kubeadm control plane without skipping
directly to three members. They only plan and verify; both evidence documents
set their mutation field to `false`, and neither function runs kubeadm, changes
etcd membership, contacts a provider, or modifies a cluster.

The only accepted sequence is:

1. build and review a `one-to-two` plan against a fresh off-cell backup and
   restore barrier;
2. separately apply the approved deployment-owned change, then capture a ready
   two-member verification;
3. take and validate a new off-cell backup whose generation time is later than
   completion of the first verification;
4. build and review the `two-to-three` plan, separately apply it, and capture
   three healthy control-plane, etcd, and API-server members;
5. supply the final public `oneserverloss.Receipt`. The verifier recomputes its
   full offline digest chain, requires its protected baseline to match the
   recovered three-member topology exactly, and enforces evidence freshness.

Final one-server-loss evidence is forbidden in the first verification and
mandatory in the second. A self-declared survive count, boolean, artifact hash,
or loader success is not sufficient: the loader must return the receipt itself
for independent validation. Verification also reopens the exact backup barrier
recorded by the plan and requires the same generation time and SHA-256 while it
is still fresh. Healthy counts are derived only from a fresh
`HAWaveInventoryReceipt`: its self-digest binds the installation, upstream
kubeadm version, sorted member UID hashes, Ready control-plane/etcd roles, and
sorted API-server UID hashes. There is no manual-count input to `VerifyHAWave`.
At three members, that inventory receipt must also bind the exact receipt hash,
run nonce, target Node UID hash, kubectl executable hash, and probe-adapter hash
from the final one-server-loss receipt.

The installation-specific backup validator and protected inventory and receipt
loaders remain downstream adapter boundaries because concrete backup systems,
file ownership policy, and private paths do not belong in the public package.

Plans and verifications use the versioned
`cloudring.kubeadm.ha-wave/v1` schema. `ReadHAWavePlan` and
`ReadHAWaveVerification` accept exactly one bounded JSON object, reject unknown
or duplicate fields and trailing documents, and reject evidence that enables
mutation or changes the canonical plan/non-claim text. Reports retain only
source-safe installation identifiers, artifact basenames, timestamps, counts,
booleans, and lowercase SHA-256 bindings. Private paths and receipt payloads are
not copied into them.

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

The included examples are synthetic contract-shape fixtures. Their receipt
bindings are placeholders and cannot produce a ready report without a matching
valid receipt. The package and site profile are an executable validation and
rendering contract, not live deployment evidence. A downstream installation
may claim a successful release only after its adapter has executed the
referenced lifecycle and the independent stand verifier has accepted freshly
captured state plus the exact receipt.
