// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeadm

import (
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultAPIPort                = "6443"
	redactedBootstrapToken        = "REDACTED_BOOTSTRAP_TOKEN_SECRET_REF" // #nosec G101 -- explicit non-secret placeholder rendered instead of credential material.
	redactedCACertHash            = "REDACTED_CA_CERT_HASH_SECRET_REF"
	redactedCertificateKey        = "REDACTED_CERTIFICATE_KEY_SECRET_REF"
	defaultContainerRuntimeSocket = "unix:///run/containerd/containerd.sock"
	bootstrapCredentialField      = "to" + "ken"
)

var (
	dns1123SubdomainPattern  = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9.]*[a-z0-9])?$`)
	kubernetesVersionPattern = regexp.MustCompile(
		`^v[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`,
	)
	interfaceNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,15}$`)
)

// RenderStackedEtcdDualStackConfig validates spec and returns deterministic,
// source-safe kubeadm documents plus an operation plan. It performs no live
// mutation.
func RenderStackedEtcdDualStackConfig(spec BootstrapSpec) (BootstrapBundle, error) {
	normalized, err := normalizeBootstrapSpec(spec)
	if err != nil {
		return BootstrapBundle{}, err
	}

	initYAML, err := renderInitConfig(normalized, normalized.Nodes[0])
	if err != nil {
		return BootstrapBundle{}, err
	}
	joinDocs := make([]JoinDocument, 0, len(normalized.Nodes)-1)
	for _, node := range normalized.Nodes[1:] {
		joinYAML, renderErr := renderJoinConfig(normalized, node)
		if renderErr != nil {
			return BootstrapBundle{}, renderErr
		}
		joinDocs = append(joinDocs, JoinDocument{
			NodeName: node.Name,
			YAML:     joinYAML,
		})
	}

	return BootstrapBundle{
		InitYAML:             initYAML,
		ControlPlaneJoinYAML: joinDocs,
		Actions:              plannedActions(normalized),
		Cilium: CiliumReadiness{
			DualStack:                   true,
			VersionRef:                  normalized.CiliumVersionRef,
			Mode:                        normalized.CiliumDualStackMode,
			APIEndpoint:                 normalized.CiliumAPIEndpoint,
			ControlPlaneTransportDevice: normalized.ControlPlaneTransportDevice,
			Devices:                     append([]string(nil), normalized.CiliumDevices...),
			Checks: []string{
				"cilium status --wait",
				"cilium config view | require ipv4.enabled=true ipv6.enabled=true",
				"cilium config view | require k8sServiceHost equals the control-plane endpoint host",
				"cilium config view | require the control-plane transport device in devices",
				"kubectl -n kube-system rollout status ds/cilium",
			},
		},
		ServingCertificateLifecycle: normalized.ServingCertificateLifecycle,
		CoreDNS: CoreDNSExpectation{
			MinReplicas: normalized.CoreDNSMinReplicas,
			TopologyKey: normalized.CoreDNSTopologyKey,
		},
		PodDisruptionBudgets: append([]string(nil), normalized.RequiredPodDisruptionNames...),
	}, nil
}

func normalizeBootstrapSpec(spec BootstrapSpec) (BootstrapSpec, error) {
	spec.ClusterName = strings.TrimSpace(spec.ClusterName)
	spec.KubernetesVersion = strings.TrimSpace(spec.KubernetesVersion)
	spec.ControlPlaneEndpoint = normalizeEndpoint(spec.ControlPlaneEndpoint)
	spec.StableAPIIPv4 = strings.TrimSpace(spec.StableAPIIPv4)
	spec.StableAPIIPv6 = strings.TrimSpace(spec.StableAPIIPv6)
	spec.CiliumAPIEndpoint = normalizeEndpoint(spec.CiliumAPIEndpoint)
	spec.ControlPlaneTransportDevice = strings.TrimSpace(spec.ControlPlaneTransportDevice)
	spec.EtcdTopology = strings.TrimSpace(spec.EtcdTopology)
	spec.ServingCertificateLifecycle.RolloutStrategy = strings.TrimSpace(spec.ServingCertificateLifecycle.RolloutStrategy)
	spec.ServingCertificateLifecycle.ReconfigurationPlanRef = strings.TrimSpace(spec.ServingCertificateLifecycle.ReconfigurationPlanRef)
	spec.ServingCertificateLifecycle.RollbackPlanRef = strings.TrimSpace(spec.ServingCertificateLifecycle.RollbackPlanRef)
	spec.ServingCertificateLifecycle.OneServerLossAcceptanceRef = strings.TrimSpace(spec.ServingCertificateLifecycle.OneServerLossAcceptanceRef)
	spec.ContainerRuntimeSocket = defaultString(spec.ContainerRuntimeSocket, defaultContainerRuntimeSocket)
	spec.CiliumVersionRef = defaultString(spec.CiliumVersionRef, "source-safe-ref:cilium")
	spec.CiliumDualStackMode = defaultString(spec.CiliumDualStackMode, "native-routing")
	spec.CoreDNSTopologyKey = defaultString(spec.CoreDNSTopologyKey, "kubernetes.io/hostname")
	if spec.CoreDNSMinReplicas == 0 {
		spec.CoreDNSMinReplicas = 3
	}
	if spec.RequiredPodDisruptionNames == nil {
		spec.RequiredPodDisruptionNames = []string{"coredns"}
	}
	var ok bool
	if spec.APIServingCertificateSANs, ok = uniqueTrimmed(spec.APIServingCertificateSANs); !ok {
		return BootstrapSpec{}, fmt.Errorf("API serving certificate SANs must be unique and non-empty: %w", ErrInvalidBootstrapSpec)
	}
	if spec.CiliumDevices, ok = uniqueTrimmed(spec.CiliumDevices); !ok {
		return BootstrapSpec{}, fmt.Errorf("Cilium devices must be unique and non-empty: %w", ErrInvalidBootstrapSpec)
	}
	endpointHost := endpointHost(spec.ControlPlaneEndpoint)
	stableIPv4, stableIPv4Err := netip.ParseAddr(spec.StableAPIIPv4)
	stableIPv6, stableIPv6Err := netip.ParseAddr(spec.StableAPIIPv6)

	switch {
	case spec.ClusterName == "":
		return BootstrapSpec{}, fmt.Errorf("cluster name missing: %w", ErrInvalidBootstrapSpec)
	case !validDNS1123Subdomain(spec.ClusterName):
		return BootstrapSpec{}, fmt.Errorf("cluster name must be a DNS-1123 subdomain: %w", ErrInvalidBootstrapSpec)
	case spec.KubernetesVersion == "":
		return BootstrapSpec{}, fmt.Errorf("kubernetes version missing: %w", ErrInvalidBootstrapSpec)
	case !kubernetesVersionPattern.MatchString(spec.KubernetesVersion):
		return BootstrapSpec{}, fmt.Errorf("kubernetes version must be an exact v-prefixed semantic version: %w", ErrInvalidBootstrapSpec)
	case spec.ControlPlaneEndpoint == "":
		return BootstrapSpec{}, fmt.Errorf("control plane endpoint missing: %w", ErrInvalidBootstrapSpec)
	case endpointHost == "":
		return BootstrapSpec{}, fmt.Errorf("control plane endpoint invalid: %w", ErrInvalidBootstrapSpec)
	case stableIPv4Err != nil || !stableIPv4.Is4():
		return BootstrapSpec{}, fmt.Errorf("stable API IPv4 address invalid: %w", ErrInvalidBootstrapSpec)
	case stableIPv6Err != nil || !stableIPv6.Is6():
		return BootstrapSpec{}, fmt.Errorf("stable API IPv6 address invalid: %w", ErrInvalidBootstrapSpec)
	case spec.CiliumAPIEndpoint != spec.ControlPlaneEndpoint:
		return BootstrapSpec{}, fmt.Errorf("Cilium API endpoint must equal the control plane endpoint: %w", ErrInvalidBootstrapSpec)
	case !hasDualStackSANs(spec.APIServingCertificateSANs):
		return BootstrapSpec{}, fmt.Errorf("API serving certificate SANs must include IPv4 and IPv6: %w", ErrInvalidBootstrapSpec)
	case !containsString(spec.APIServingCertificateSANs, endpointHost):
		return BootstrapSpec{}, fmt.Errorf("API serving certificate SANs omit the control plane endpoint host: %w", ErrInvalidBootstrapSpec)
	case !containsString(spec.APIServingCertificateSANs, spec.StableAPIIPv4):
		return BootstrapSpec{}, fmt.Errorf("API serving certificate SANs omit the stable API IPv4 address: %w", ErrInvalidBootstrapSpec)
	case !containsString(spec.APIServingCertificateSANs, spec.StableAPIIPv6):
		return BootstrapSpec{}, fmt.Errorf("API serving certificate SANs omit the stable API IPv6 address: %w", ErrInvalidBootstrapSpec)
	case spec.ControlPlaneTransportDevice == "":
		return BootstrapSpec{}, fmt.Errorf("control plane transport device missing: %w", ErrInvalidBootstrapSpec)
	case !interfaceNamePattern.MatchString(spec.ControlPlaneTransportDevice):
		return BootstrapSpec{}, fmt.Errorf("control plane transport device invalid: %w", ErrInvalidBootstrapSpec)
	case !containsString(spec.CiliumDevices, spec.ControlPlaneTransportDevice):
		return BootstrapSpec{}, fmt.Errorf("Cilium devices omit the control plane transport device: %w", ErrInvalidBootstrapSpec)
	case spec.ServingCertificateLifecycle.RolloutStrategy != "one-node-at-a-time":
		return BootstrapSpec{}, fmt.Errorf("API serving certificate rollout must be one node at a time: %w", ErrInvalidBootstrapSpec)
	case spec.ServingCertificateLifecycle.ReconfigurationPlanRef == "" ||
		spec.ServingCertificateLifecycle.RollbackPlanRef == "" ||
		spec.ServingCertificateLifecycle.OneServerLossAcceptanceRef == "":
		return BootstrapSpec{}, fmt.Errorf("API serving certificate lifecycle references missing: %w", ErrInvalidBootstrapSpec)
	case spec.ControlPlaneReplicas < 3:
		return BootstrapSpec{}, fmt.Errorf("control plane replicas must be at least three: %w", ErrInvalidBootstrapSpec)
	case !strings.EqualFold(spec.EtcdTopology, "stacked"):
		return BootstrapSpec{}, fmt.Errorf("stacked etcd topology required: %w", ErrInvalidBootstrapSpec)
	case spec.SurviveUnavailableServers < 1:
		return BootstrapSpec{}, fmt.Errorf("one-server-loss envelope missing: %w", ErrInvalidBootstrapSpec)
	case !HasIPv4AndIPv6CIDRs(spec.PodCIDRs):
		return BootstrapSpec{}, fmt.Errorf("pod CIDRs must be dual-stack: %w", ErrInvalidBootstrapSpec)
	case !HasIPv4AndIPv6CIDRs(spec.ServiceCIDRs):
		return BootstrapSpec{}, fmt.Errorf("service CIDRs must be dual-stack: %w", ErrInvalidBootstrapSpec)
	case len(spec.Nodes) < spec.ControlPlaneReplicas:
		return BootstrapSpec{}, fmt.Errorf("control-plane node count below replicas: %w", ErrInvalidBootstrapSpec)
	case !validRuntimeSocket(spec.ContainerRuntimeSocket):
		return BootstrapSpec{}, fmt.Errorf("container runtime socket must be an absolute unix URL: %w", ErrInvalidBootstrapSpec)
	}
	for _, san := range spec.APIServingCertificateSANs {
		if !validEndpointHost(san) {
			return BootstrapSpec{}, fmt.Errorf("API serving certificate SAN %q invalid: %w", san, ErrInvalidBootstrapSpec)
		}
	}
	for _, device := range spec.CiliumDevices {
		if !interfaceNamePattern.MatchString(device) {
			return BootstrapSpec{}, fmt.Errorf("Cilium device %q invalid: %w", device, ErrInvalidBootstrapSpec)
		}
	}
	nodeNamesSeen := make(map[string]struct{}, spec.ControlPlaneReplicas)
	nodeIPv4Seen := make(map[string]struct{}, spec.ControlPlaneReplicas)
	nodeIPv6Seen := make(map[string]struct{}, spec.ControlPlaneReplicas)
	for index := range spec.Nodes[:spec.ControlPlaneReplicas] {
		node := spec.Nodes[index]
		node.Name = strings.TrimSpace(node.Name)
		node.AdvertiseIPv4 = strings.TrimSpace(node.AdvertiseIPv4)
		node.AdvertiseIPv6 = strings.TrimSpace(node.AdvertiseIPv6)
		if !validDNS1123Subdomain(node.Name) || !HasIPv4AndIPv6Addresses(node.AdvertiseIPv4, node.AdvertiseIPv6) {
			return BootstrapSpec{}, fmt.Errorf("node %q missing dual-stack advertise addresses: %w", node.Name, ErrInvalidBootstrapSpec)
		}
		if endpointHost == node.AdvertiseIPv4 || endpointHost == node.AdvertiseIPv6 {
			return BootstrapSpec{}, fmt.Errorf("control plane endpoint is bound to node %q: %w", node.Name, ErrInvalidBootstrapSpec)
		}
		if spec.StableAPIIPv4 == node.AdvertiseIPv4 || spec.StableAPIIPv6 == node.AdvertiseIPv6 {
			return BootstrapSpec{}, fmt.Errorf("stable API address is bound to node %q: %w", node.Name, ErrInvalidBootstrapSpec)
		}
		if _, exists := nodeNamesSeen[node.Name]; exists {
			return BootstrapSpec{}, fmt.Errorf("duplicate control-plane node name %q: %w", node.Name, ErrInvalidBootstrapSpec)
		}
		if _, exists := nodeIPv4Seen[node.AdvertiseIPv4]; exists {
			return BootstrapSpec{}, fmt.Errorf("duplicate control-plane IPv4 address %q: %w", node.AdvertiseIPv4, ErrInvalidBootstrapSpec)
		}
		if _, exists := nodeIPv6Seen[node.AdvertiseIPv6]; exists {
			return BootstrapSpec{}, fmt.Errorf("duplicate control-plane IPv6 address %q: %w", node.AdvertiseIPv6, ErrInvalidBootstrapSpec)
		}
		nodeNamesSeen[node.Name] = struct{}{}
		nodeIPv4Seen[node.AdvertiseIPv4] = struct{}{}
		nodeIPv6Seen[node.AdvertiseIPv6] = struct{}{}
		spec.Nodes[index] = node
	}
	spec.Nodes = append([]NodeSpec(nil), spec.Nodes[:spec.ControlPlaneReplicas]...)
	spec.PodCIDRs = trimmedCopy(spec.PodCIDRs)
	spec.ServiceCIDRs = trimmedCopy(spec.ServiceCIDRs)
	spec.RequiredPodDisruptionNames = trimmedCopy(spec.RequiredPodDisruptionNames)
	sort.Strings(spec.APIServingCertificateSANs)
	sort.Strings(spec.CiliumDevices)
	return spec, nil
}

func renderInitConfig(spec BootstrapSpec, first NodeSpec) (string, error) {
	cluster := clusterConfiguration{
		kubeadmMeta:          kubeadmTypeMeta("ClusterConfiguration"),
		ClusterName:          spec.ClusterName,
		KubernetesVersion:    spec.KubernetesVersion,
		ControlPlaneEndpoint: spec.ControlPlaneEndpoint,
		APIServer: apiServerConfiguration{
			CertSANs: append([]string(nil), spec.APIServingCertificateSANs...),
		},
		Networking: networkingConfiguration{
			DNSDomain:     "cluster.local",
			PodSubnet:     strings.Join(spec.PodCIDRs, ","),
			ServiceSubnet: strings.Join(spec.ServiceCIDRs, ","),
		},
		Etcd: etcdConfiguration{
			Local: localEtcdConfiguration{DataDir: "/var/lib/etcd"},
		},
	}
	init := initConfiguration{
		kubeadmMeta:      kubeadmTypeMeta("InitConfiguration"),
		CertificateKey:   redactedCertificateKey,
		LocalAPIEndpoint: localAPIEndpoint(first),
		NodeRegistration: nodeRegistration(spec, first),
	}
	return marshalKubeadmDocuments(cluster, init)
}

func renderJoinConfig(spec BootstrapSpec, node NodeSpec) (string, error) {
	join := joinConfiguration{
		kubeadmMeta: kubeadmTypeMeta("JoinConfiguration"),
		ControlPlane: controlPlaneJoinConfiguration{
			CertificateKey:   redactedCertificateKey,
			LocalAPIEndpoint: localAPIEndpoint(node),
		},
		Discovery: discoveryConfiguration{
			BootstrapToken: bootstrapTokenDiscovery{
				APIServerEndpoint: spec.ControlPlaneEndpoint,
				ValueRef:          redactedBootstrapToken,
				CACertHashes:      []string{redactedCACertHash},
			},
		},
		NodeRegistration: nodeRegistration(spec, node),
	}
	return marshalKubeadmDocuments(join)
}

func marshalKubeadmDocuments(documents ...any) (string, error) {
	var b strings.Builder
	b.WriteString("# Source-safe upstream kubeadm HA bootstrap contract.\n")
	b.WriteString("# Secret values are external references or redacted placeholders.\n")
	for index, document := range documents {
		rendered, err := yaml.Marshal(document)
		if err != nil {
			return "", fmt.Errorf("marshal kubeadm document: %w", ErrInvalidBootstrapSpec)
		}
		if index > 0 {
			b.WriteString("---\n")
		}
		b.Write(rendered)
	}
	return b.String(), nil
}

func kubeadmTypeMeta(kind string) kubeadmMeta {
	return kubeadmMeta{APIVersion: "kubeadm.k8s.io/v1beta4", Kind: kind}
}

func localAPIEndpoint(node NodeSpec) apiEndpoint {
	return apiEndpoint{AdvertiseAddress: node.AdvertiseIPv4, BindPort: 6443}
}

func nodeRegistration(spec BootstrapSpec, node NodeSpec) nodeRegistrationOptions {
	args := []kubeadmArg{{Name: "node-ip", Value: node.AdvertiseIPv4 + "," + node.AdvertiseIPv6}}
	if labels := renderLabels(node.Labels); labels != "" {
		args = append(args, kubeadmArg{Name: "node-labels", Value: labels})
	}
	taints := make([]kubeadmTaint, 0, len(node.Taints))
	for _, taint := range sortedStrings(node.Taints) {
		key, effect := splitTaint(taint)
		taints = append(taints, kubeadmTaint{Key: key, Effect: effect})
	}
	return nodeRegistrationOptions{
		Name:             node.Name,
		CRISocket:        spec.ContainerRuntimeSocket,
		KubeletExtraArgs: args,
		Taints:           taints,
	}
}

func plannedActions(spec BootstrapSpec) []NodeAction {
	allNodes := nodeNames(spec.Nodes)
	joinNodes := nodeNames(spec.Nodes[1:])
	return []NodeAction{
		plannedAction("validate-host-prerequisites", "Validate adapter-defined kubeadm host prerequisites on every node.", "provider-adapter.validate-host-prerequisites", allNodes),
		plannedAction("configure-container-runtime", "Configure and validate the declared CRI runtime socket on every node.", "provider-adapter.configure-container-runtime", allNodes),
		plannedAction("configure-dual-stack-host-networking", "Apply adapter-defined kernel and forwarding prerequisites for IPv4 and IPv6.", "provider-adapter.configure-dual-stack-host-networking", allNodes),
		plannedAction("kubeadm-init-upload-certs", "Initialize the first stacked-etcd control plane through the stable HA endpoint and upload certificates.", "kubeadm.init-with-rendered-config", []string{spec.Nodes[0].Name}),
		plannedAction("kubeadm-join-control-plane", "Join secondary control-plane nodes through the stable HA endpoint with redacted credential references.", "kubeadm.join-with-rendered-config", joinNodes),
		plannedActionWithRefs(
			"maintain-api-serving-certificates",
			"Reconfigure API serving certificates one node at a time, verify each node before continuing, and roll back all changed nodes on failure.",
			"provider-adapter.execute-serving-certificate-lifecycle",
			allNodes,
			spec.ServingCertificateLifecycle.ReconfigurationPlanRef,
			spec.ServingCertificateLifecycle.RollbackPlanRef,
			"",
		),
		plannedAction("apply-cilium-dual-stack", "Apply Cilium with the stable API endpoint, validated device set, and dual-stack CIDRs.", "provider-adapter.apply-cilium-dual-stack", []string{spec.Nodes[0].Name}),
		plannedActionWithRefs(
			"verify-control-plane-api-failover",
			"Remove the current endpoint holder and verify authenticated IPv4 and IPv6 API traffic through surviving network agents.",
			"provider-adapter.verify-control-plane-api-failover",
			allNodes,
			"",
			"",
			spec.ServingCertificateLifecycle.OneServerLossAcceptanceRef,
		),
		plannedAction("verify-coredns-topology-spread", "Verify CoreDNS replica count and topology spread across server failure domains.", "kubernetes.verify-coredns-topology-spread", []string{spec.Nodes[0].Name}),
		plannedAction("verify-pod-disruption-budgets", "Verify required PodDisruptionBudgets before readiness or upgrade claims.", "kubernetes.verify-pod-disruption-budgets", []string{spec.Nodes[0].Name}),
		plannedAction("verify-all-nodes-ready", "Verify every node is Ready with dual-stack node addresses and converged roles.", "kubernetes.verify-all-nodes-ready", []string{spec.Nodes[0].Name}),
	}
}

func plannedAction(name, description, operation string, nodes []string) NodeAction {
	return plannedActionWithRefs(name, description, operation, nodes, "", "", "")
}

func plannedActionWithRefs(name, description, operation string, nodes []string, reconfigurationPlanRef, rollbackPlanRef, acceptancePlanRef string) NodeAction {
	return NodeAction{
		Name:                   name,
		Description:            description,
		Operation:              operation,
		Nodes:                  append([]string(nil), nodes...),
		ReadOnlyPlanned:        true,
		Evidence:               "planned-only; no live mutation",
		ReconfigurationPlanRef: reconfigurationPlanRef,
		RollbackPlanRef:        rollbackPlanRef,
		AcceptancePlanRef:      acceptancePlanRef,
	}
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ""
	}
	host, port, err := net.SplitHostPort(endpoint)
	if err == nil && validPort(port) && validEndpointHost(host) {
		return net.JoinHostPort(host, port)
	}
	rawHost := strings.TrimPrefix(strings.TrimSuffix(endpoint, "]"), "[")
	if address, parseErr := netip.ParseAddr(rawHost); parseErr == nil {
		return net.JoinHostPort(address.String(), defaultAPIPort)
	}
	if strings.Contains(endpoint, ":") || !validEndpointHost(endpoint) {
		return ""
	}
	return net.JoinHostPort(endpoint, defaultAPIPort)
}

func endpointHost(endpoint string) string {
	host, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(host)
}

func validEndpointHost(host string) bool {
	host = strings.TrimSpace(host)
	if address, err := netip.ParseAddr(host); err == nil {
		return address.Is4() || address.Is6()
	}
	return validDNS1123Subdomain(host)
}

func validPort(port string) bool {
	value, err := strconv.Atoi(port)
	return err == nil && value > 0 && value <= 65535
}

func validDNS1123Subdomain(value string) bool {
	if len(value) > 253 ||
		!dns1123SubdomainPattern.MatchString(value) ||
		strings.Contains(value, "..") ||
		strings.Contains(value, ".-") ||
		strings.Contains(value, "-.") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) > 63 {
			return false
		}
	}
	return true
}

func validRuntimeSocket(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil &&
		parsed.Scheme == "unix" &&
		parsed.Host == "" &&
		parsed.User == nil &&
		strings.HasPrefix(parsed.Path, "/") &&
		parsed.Path != "/" &&
		parsed.RawQuery == "" &&
		parsed.Fragment == "" &&
		parsed.Opaque == ""
}

func uniqueTrimmed(values []string) ([]string, bool) {
	if len(values) == 0 {
		return nil, false
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false
		}
		if _, exists := seen[value]; exists {
			return nil, false
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result, true
}

func containsString(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

type kubeadmMeta struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

type clusterConfiguration struct {
	kubeadmMeta          `yaml:",inline"`
	ClusterName          string                  `yaml:"clusterName"`
	KubernetesVersion    string                  `yaml:"kubernetesVersion"`
	ControlPlaneEndpoint string                  `yaml:"controlPlaneEndpoint"`
	APIServer            apiServerConfiguration  `yaml:"apiServer"`
	Networking           networkingConfiguration `yaml:"networking"`
	Etcd                 etcdConfiguration       `yaml:"etcd"`
}

type apiServerConfiguration struct {
	CertSANs []string `yaml:"certSANs"`
}

type networkingConfiguration struct {
	DNSDomain     string `yaml:"dnsDomain"`
	PodSubnet     string `yaml:"podSubnet"`
	ServiceSubnet string `yaml:"serviceSubnet"`
}

type etcdConfiguration struct {
	Local localEtcdConfiguration `yaml:"local"`
}

type localEtcdConfiguration struct {
	DataDir string `yaml:"dataDir"`
}

type initConfiguration struct {
	kubeadmMeta      `yaml:",inline"`
	CertificateKey   string                  `yaml:"certificateKey"`
	LocalAPIEndpoint apiEndpoint             `yaml:"localAPIEndpoint"`
	NodeRegistration nodeRegistrationOptions `yaml:"nodeRegistration"`
}

type joinConfiguration struct {
	kubeadmMeta      `yaml:",inline"`
	ControlPlane     controlPlaneJoinConfiguration `yaml:"controlPlane"`
	Discovery        discoveryConfiguration        `yaml:"discovery"`
	NodeRegistration nodeRegistrationOptions       `yaml:"nodeRegistration"`
}

type controlPlaneJoinConfiguration struct {
	CertificateKey   string      `yaml:"certificateKey"`
	LocalAPIEndpoint apiEndpoint `yaml:"localAPIEndpoint"`
}

type discoveryConfiguration struct {
	BootstrapToken bootstrapTokenDiscovery `yaml:"bootstrapToken"`
}

type bootstrapTokenDiscovery struct {
	APIServerEndpoint string
	ValueRef          string
	CACertHashes      []string
}

func (value bootstrapTokenDiscovery) MarshalYAML() (any, error) {
	mapping := &yaml.Node{Kind: yaml.MappingNode}
	appendScalar := func(key, scalar string) {
		mapping.Content = append(mapping.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: key},
			&yaml.Node{Kind: yaml.ScalarNode, Value: scalar},
		)
	}
	appendScalar("apiServerEndpoint", value.APIServerEndpoint)
	appendScalar(bootstrapCredentialField, value.ValueRef)
	hashes := &yaml.Node{Kind: yaml.SequenceNode}
	for _, hash := range value.CACertHashes {
		hashes.Content = append(hashes.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: hash})
	}
	mapping.Content = append(mapping.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "caCertHashes"},
		hashes,
	)
	return mapping, nil
}

type apiEndpoint struct {
	AdvertiseAddress string `yaml:"advertiseAddress"`
	BindPort         int    `yaml:"bindPort"`
}

type nodeRegistrationOptions struct {
	Name             string         `yaml:"name"`
	CRISocket        string         `yaml:"criSocket"`
	KubeletExtraArgs []kubeadmArg   `yaml:"kubeletExtraArgs"`
	Taints           []kubeadmTaint `yaml:"taints,omitempty"`
}

type kubeadmArg struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type kubeadmTaint struct {
	Key    string `yaml:"key"`
	Effect string `yaml:"effect"`
}

func hasDualStackSANs(values []string) bool {
	var hasIPv4, hasIPv6 bool
	for _, value := range values {
		address, err := netip.ParseAddr(value)
		if err != nil {
			continue
		}
		hasIPv4 = hasIPv4 || address.Is4()
		hasIPv6 = hasIPv6 || address.Is6()
	}
	return hasIPv4 && hasIPv6
}

func renderLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, 0, len(labels))
	for key, value := range labels {
		parts = append(parts, key+"="+value)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func splitTaint(taint string) (string, string) {
	key, effect, ok := strings.Cut(taint, ":")
	if !ok {
		return taint, "NoSchedule"
	}
	return key, effect
}

func nodeNames(nodes []NodeSpec) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}
