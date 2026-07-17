// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeadm

import (
	"io"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestKubeadmRenderStackedEtcdDualStackConfig(t *testing.T) {
	spec := BootstrapSpec{
		ClusterName:                 "cloudring-synthetic-site",
		KubernetesVersion:           "v1.35.6",
		ControlPlaneEndpoint:        "api.synthetic.example:6443",
		StableAPIIPv4:               "192.0.2.20",
		StableAPIIPv6:               "2001:db8::20",
		APIServingCertificateSANs:   []string{"api.synthetic.example", "192.0.2.20", "2001:db8::20"},
		ControlPlaneReplicas:        3,
		EtcdTopology:                "stacked",
		PodCIDRs:                    []string{"192.0.2.0/24", "2001:db8:244::/56"},
		ServiceCIDRs:                []string{"198.51.100.0/24", "2001:db8:96::/108"},
		SurviveUnavailableServers:   1,
		CiliumVersionRef:            "oci://registry.example.invalid/cilium@sha256:redacted",
		CiliumDualStackMode:         "native-routing",
		CiliumAPIEndpoint:           "api.synthetic.example:6443",
		ControlPlaneTransportDevice: "transport0",
		CiliumDevices:               []string{"public0", "transport0"},
		ServingCertificateLifecycle: ServingCertificateLifecycle{
			RolloutStrategy:            "one-node-at-a-time",
			ReconfigurationPlanRef:     "operations.api-certificate.reconfigure",
			RollbackPlanRef:            "operations.api-certificate.rollback",
			OneServerLossAcceptanceRef: "resilience.api.one-server-loss",
		},
		ContainerRuntimeSocket:     "unix:///run/containerd/containerd.sock",
		CoreDNSMinReplicas:         3,
		CoreDNSTopologyKey:         "kubernetes.io/hostname",
		RequiredPodDisruptionNames: []string{"coredns", "provider-controller", "provider-portal"},
		Nodes: []NodeSpec{
			{
				Name:          "node-a",
				AdvertiseIPv4: "192.0.2.11",
				AdvertiseIPv6: "2001:db8::11",
				Roles:         []string{"control-plane", "etcd", "worker", "storage", "ingress"},
				Labels:        map[string]string{"node.cloudring.io/failure-domain": "server-a"},
				Taints:        []string{"node-role.kubernetes.io/control-plane:NoSchedule"},
			},
			{
				Name:          "node-b",
				AdvertiseIPv4: "192.0.2.12",
				AdvertiseIPv6: "2001:db8::12",
				Roles:         []string{"control-plane", "etcd", "worker", "storage", "ingress"},
				Labels:        map[string]string{"node.cloudring.io/failure-domain": "server-b"},
				Taints:        []string{"node-role.kubernetes.io/control-plane:NoSchedule"},
			},
			{
				Name:          "node-c",
				AdvertiseIPv4: "192.0.2.13",
				AdvertiseIPv6: "2001:db8::13",
				Roles:         []string{"control-plane", "etcd", "worker", "storage", "ingress"},
				Labels:        map[string]string{"node.cloudring.io/failure-domain": "server-c"},
				Taints:        []string{"node-role.kubernetes.io/control-plane:NoSchedule"},
			},
		},
	}

	bundle, err := RenderStackedEtcdDualStackConfig(spec)
	if err != nil {
		t.Fatalf("expected renderer to accept upstream stacked-etcd dual-stack spec: %v", err)
	}

	for _, want := range []string{
		"kind: ClusterConfiguration",
		"controlPlaneEndpoint: api.synthetic.example:6443",
		"apiServer:",
		"- api.synthetic.example",
		"- 192.0.2.20",
		"- 2001:db8::20",
		"kubernetesVersion: v1.35.6",
		"podSubnet: 192.0.2.0/24,2001:db8:244::/56",
		"serviceSubnet: 198.51.100.0/24,2001:db8:96::/108",
		"name: node-ip",
		"value: 192.0.2.11,2001:db8::11",
		"criSocket: unix:///run/containerd/containerd.sock",
		"certificateKey: REDACTED_CERTIFICATE_KEY_SECRET_REF",
	} {
		if !strings.Contains(bundle.InitYAML, want) {
			t.Fatalf("rendered init config missing %q:\n%s", want, bundle.InitYAML)
		}
	}
	assertValidYAMLDocuments(t, bundle.InitYAML, 2)
	if len(bundle.ControlPlaneJoinYAML) != 2 {
		t.Fatalf("expected two secondary control-plane join configs, got %d", len(bundle.ControlPlaneJoinYAML))
	}
	if !strings.Contains(bundle.ControlPlaneJoinYAML[0].YAML, "apiServerEndpoint: api.synthetic.example:6443") {
		t.Fatalf("join config must target the HA endpoint:\n%s", bundle.ControlPlaneJoinYAML[0].YAML)
	}
	if !strings.Contains(bundle.ControlPlaneJoinYAML[0].YAML, "value: 192.0.2.12,2001:db8::12") {
		t.Fatalf("join config must carry dual-stack kubelet node-ip:\n%s", bundle.ControlPlaneJoinYAML[0].YAML)
	}
	assertValidYAMLDocuments(t, bundle.ControlPlaneJoinYAML[0].YAML, 1)
	assertAction(t, bundle.Actions, "validate-host-prerequisites")
	assertAction(t, bundle.Actions, "configure-container-runtime")
	assertAction(t, bundle.Actions, "configure-dual-stack-host-networking")
	assertAction(t, bundle.Actions, "kubeadm-init-upload-certs")
	assertAction(t, bundle.Actions, "kubeadm-join-control-plane")
	assertAction(t, bundle.Actions, "maintain-api-serving-certificates")
	assertAction(t, bundle.Actions, "apply-cilium-dual-stack")
	assertAction(t, bundle.Actions, "verify-control-plane-api-failover")
	assertAction(t, bundle.Actions, "verify-coredns-topology-spread")
	assertAction(t, bundle.Actions, "verify-pod-disruption-budgets")
	assertAction(t, bundle.Actions, "verify-all-nodes-ready")
	if !bundle.Cilium.DualStack {
		t.Fatalf("expected Cilium readiness metadata to require dual-stack: %#v", bundle.Cilium)
	}
	if bundle.Cilium.APIEndpoint != spec.CiliumAPIEndpoint ||
		bundle.Cilium.ControlPlaneTransportDevice != spec.ControlPlaneTransportDevice ||
		!containsString(bundle.Cilium.Devices, spec.ControlPlaneTransportDevice) {
		t.Fatalf("Cilium readiness omits the HA endpoint or transport device: %#v", bundle.Cilium)
	}
	if bundle.ServingCertificateLifecycle != spec.ServingCertificateLifecycle {
		t.Fatalf("bundle lost serving-certificate lifecycle references: %#v", bundle.ServingCertificateLifecycle)
	}
	assertActionRefs(
		t,
		bundle.Actions,
		"maintain-api-serving-certificates",
		spec.ServingCertificateLifecycle.ReconfigurationPlanRef,
		spec.ServingCertificateLifecycle.RollbackPlanRef,
		"",
	)
	assertActionRefs(
		t,
		bundle.Actions,
		"verify-control-plane-api-failover",
		"",
		"",
		spec.ServingCertificateLifecycle.OneServerLossAcceptanceRef,
	)
}

func TestKubeadmRendererCanonicalizesIPAddresses(t *testing.T) {
	spec := syntheticBootstrapSpec()
	spec.ControlPlaneEndpoint = "[2001:0db8:0:0:0:0:0:30]:6443"
	spec.CiliumAPIEndpoint = spec.ControlPlaneEndpoint
	spec.StableAPIIPv6 = "2001:0db8:0:0:0:0:0:20"
	spec.APIServingCertificateSANs = []string{
		"2001:0db8:0:0:0:0:0:30",
		spec.StableAPIIPv4,
		spec.StableAPIIPv6,
	}
	spec.Nodes[0].AdvertiseIPv6 = "2001:0db8:0:0:0:0:0:11"
	spec.Nodes[1].AdvertiseIPv6 = "2001:0db8:0:0:0:0:0:12"
	spec.Nodes[2].AdvertiseIPv6 = "2001:0db8:0:0:0:0:0:13"

	bundle, err := RenderStackedEtcdDualStackConfig(spec)
	if err != nil {
		t.Fatalf("expected equivalent IPv6 spellings to normalize: %v", err)
	}
	for _, want := range []string{
		"controlPlaneEndpoint: '[2001:db8::30]:6443'",
		"- 2001:db8::20",
		"value: 192.0.2.11,2001:db8::11",
	} {
		if !strings.Contains(bundle.InitYAML, want) {
			t.Fatalf("normalized kubeadm config missing %q:\n%s", want, bundle.InitYAML)
		}
	}
	if bundle.Cilium.APIEndpoint != "[2001:db8::30]:6443" {
		t.Fatalf("Cilium endpoint was not canonicalized: %q", bundle.Cilium.APIEndpoint)
	}
}

func TestKubeadmRendererRejectsUnsafeControlPlaneAPIContracts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*BootstrapSpec)
	}{
		{
			name: "node-bound endpoint",
			mutate: func(spec *BootstrapSpec) {
				spec.ControlPlaneEndpoint = "192.0.2.11:6443"
				spec.CiliumAPIEndpoint = spec.ControlPlaneEndpoint
				spec.APIServingCertificateSANs[0] = "192.0.2.11"
			},
		},
		{
			name: "node-bound equivalent IPv6 endpoint",
			mutate: func(spec *BootstrapSpec) {
				spec.ControlPlaneEndpoint = "[2001:0db8:0:0:0:0:0:11]:6443"
				spec.CiliumAPIEndpoint = spec.ControlPlaneEndpoint
				spec.APIServingCertificateSANs = append(
					spec.APIServingCertificateSANs,
					"2001:0db8:0:0:0:0:0:11",
				)
			},
		},
		{
			name: "unrelated stable API IPv4",
			mutate: func(spec *BootstrapSpec) {
				spec.StableAPIIPv4 = "192.0.2.99"
			},
		},
		{
			name: "node-bound stable API IPv4",
			mutate: func(spec *BootstrapSpec) {
				spec.StableAPIIPv4 = spec.Nodes[0].AdvertiseIPv4
				spec.APIServingCertificateSANs = append(spec.APIServingCertificateSANs, spec.StableAPIIPv4)
			},
		},
		{
			name: "node-bound equivalent stable API IPv6",
			mutate: func(spec *BootstrapSpec) {
				spec.StableAPIIPv6 = "2001:0db8:0:0:0:0:0:11"
				spec.APIServingCertificateSANs = append(spec.APIServingCertificateSANs, spec.StableAPIIPv6)
			},
		},
		{
			name: "missing IPv6 SAN",
			mutate: func(spec *BootstrapSpec) {
				spec.APIServingCertificateSANs = []string{"api.synthetic.example", "192.0.2.20"}
			},
		},
		{
			name: "invalid extra SAN",
			mutate: func(spec *BootstrapSpec) {
				spec.APIServingCertificateSANs = append(spec.APIServingCertificateSANs, "safe\nkind: Injected")
			},
		},
		{
			name: "Cilium endpoint mismatch",
			mutate: func(spec *BootstrapSpec) {
				spec.CiliumAPIEndpoint = "api-other.synthetic.example:6443"
			},
		},
		{
			name: "Cilium transport device missing",
			mutate: func(spec *BootstrapSpec) {
				spec.CiliumDevices = []string{"public0"}
			},
		},
		{
			name: "parallel certificate rollout",
			mutate: func(spec *BootstrapSpec) {
				spec.ServingCertificateLifecycle.RolloutStrategy = "parallel"
			},
		},
		{
			name: "missing certificate rollback",
			mutate: func(spec *BootstrapSpec) {
				spec.ServingCertificateLifecycle.RollbackPlanRef = ""
			},
		},
		{
			name: "duplicate Cilium device",
			mutate: func(spec *BootstrapSpec) {
				spec.CiliumDevices = []string{"transport0", "transport0"}
			},
		},
		{
			name: "duplicate equivalent IPv6 SAN",
			mutate: func(spec *BootstrapSpec) {
				spec.APIServingCertificateSANs = append(
					spec.APIServingCertificateSANs,
					"2001:0db8:0:0:0:0:0:20",
				)
			},
		},
		{
			name: "duplicate node name",
			mutate: func(spec *BootstrapSpec) {
				spec.Nodes[1].Name = spec.Nodes[0].Name
			},
		},
		{
			name: "duplicate node IPv4",
			mutate: func(spec *BootstrapSpec) {
				spec.Nodes[1].AdvertiseIPv4 = spec.Nodes[0].AdvertiseIPv4
			},
		},
		{
			name: "duplicate node IPv6",
			mutate: func(spec *BootstrapSpec) {
				spec.Nodes[1].AdvertiseIPv6 = spec.Nodes[0].AdvertiseIPv6
			},
		},
		{
			name: "duplicate equivalent node IPv6",
			mutate: func(spec *BootstrapSpec) {
				spec.Nodes[1].AdvertiseIPv6 = "2001:0db8:0:0:0:0:0:11"
			},
		},
		{
			name: "cluster name YAML injection",
			mutate: func(spec *BootstrapSpec) {
				spec.ClusterName = "safe\nkind: Injected"
			},
		},
		{
			name: "node name YAML injection",
			mutate: func(spec *BootstrapSpec) {
				spec.Nodes[0].Name = "safe\nkind: Injected"
			},
		},
		{
			name: "non-semantic Kubernetes version",
			mutate: func(spec *BootstrapSpec) {
				spec.KubernetesVersion = "stable"
			},
		},
		{
			name: "unsafe runtime socket",
			mutate: func(spec *BootstrapSpec) {
				spec.ContainerRuntimeSocket = "unix:///run/runtime.sock?override=true"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := syntheticBootstrapSpec()
			test.mutate(&spec)
			if _, err := RenderStackedEtcdDualStackConfig(spec); err == nil {
				t.Fatal("unsafe control-plane API contract passed renderer validation")
			}
		})
	}
}

func TestUpstreamStandVerifierRejectsLegacyDistributionVersion(t *testing.T) {
	inventory := readyInventory()
	inventory.ServerVersion = "v1.35.0+k3s1"

	report, err := VerifyUpstreamStand(inventory)
	if err == nil {
		t.Fatal("expected verifier to reject +k3s server version")
	}
	if report.Status != "blocked" {
		t.Fatalf("expected blocked report, got %s", report.Status)
	}
	assertBlocker(t, report.Blockers, "legacy_k3s_version")
	if report.WorkflowContinuity.Summary == "" {
		t.Fatalf("expected workflow inventory in rejection report: %#v", report.WorkflowContinuity)
	}
	if report.DataDurability.Summary == "" {
		t.Fatalf("expected data durability inventory in rejection report: %#v", report.DataDurability)
	}
	if report.SinglePointOfFailure.Summary == "" {
		t.Fatalf("expected SPOF inventory in rejection report: %#v", report.SinglePointOfFailure)
	}
}

func TestUpstreamStandVerifierReportsReadinessInventory(t *testing.T) {
	report, err := VerifyUpstreamStand(readyInventory())
	if err != nil {
		t.Fatalf("expected ready upstream stand to verify: %v", err)
	}
	if report.Status != "ready" {
		t.Fatalf("expected ready report, got %s", report.Status)
	}
	if report.WorkflowContinuity.Summary == "" || len(report.WorkflowContinuity.Items) == 0 {
		t.Fatalf("expected workflow continuity inventory: %#v", report.WorkflowContinuity)
	}
	if report.DataDurability.Summary == "" || len(report.DataDurability.Items) == 0 {
		t.Fatalf("expected data durability inventory: %#v", report.DataDurability)
	}
	if report.SinglePointOfFailure.Summary == "" || len(report.SinglePointOfFailure.Items) == 0 {
		t.Fatalf("expected SPOF inventory: %#v", report.SinglePointOfFailure)
	}
}

func TestUpstreamStandVerifierFailsClosedOnMissingEvidence(t *testing.T) {
	inventory := readyInventory()
	inventory.PodCIDRs = []string{"192.0.2.0/24"}
	inventory.DataDurability = EvidenceInventory{}

	report, err := VerifyUpstreamStand(inventory)
	if err == nil {
		t.Fatal("expected missing dual-stack and durability evidence to block")
	}
	assertBlocker(t, report.Blockers, "missing_dual_stack_pod_cidrs")
	assertBlocker(t, report.Blockers, "missing_data_durability_evidence")
}

func TestUpstreamStandVerifierRejectsUnsafeControlPlaneAPIState(t *testing.T) {
	tests := []struct {
		name    string
		blocker string
		mutate  func(*StandInventory)
	}{
		{
			name:    "node-bound endpoint",
			blocker: "node_bound_control_plane_endpoint",
			mutate: func(inventory *StandInventory) {
				inventory.ControlPlaneEndpoint = "192.0.2.11:6443"
				inventory.CiliumAPIEndpoint = inventory.ControlPlaneEndpoint
				inventory.APIServingCertificateSANs[0] = "192.0.2.11"
			},
		},
		{
			name:    "node-bound equivalent IPv6 endpoint",
			blocker: "node_bound_control_plane_endpoint",
			mutate: func(inventory *StandInventory) {
				inventory.ControlPlaneEndpoint = "[2001:0db8:0:0:0:0:0:11]:6443"
				inventory.CiliumAPIEndpoint = inventory.ControlPlaneEndpoint
				inventory.APIServingCertificateSANs = append(
					inventory.APIServingCertificateSANs,
					"2001:0db8:0:0:0:0:0:11",
				)
			},
		},
		{
			name:    "stable API IPv4 omitted from SANs",
			blocker: "api_serving_certificate_sans_missing",
			mutate: func(inventory *StandInventory) {
				inventory.StableAPIIPv4 = "192.0.2.99"
			},
		},
		{
			name:    "node-bound stable API IPv4",
			blocker: "node_bound_stable_api_address",
			mutate: func(inventory *StandInventory) {
				inventory.StableAPIIPv4 = inventory.Nodes[0].NodeIPv4
				inventory.APIServingCertificateSANs = append(inventory.APIServingCertificateSANs, inventory.StableAPIIPv4)
			},
		},
		{
			name:    "node-bound equivalent stable API IPv6",
			blocker: "node_bound_stable_api_address",
			mutate: func(inventory *StandInventory) {
				inventory.StableAPIIPv6 = "2001:0db8:0:0:0:0:0:11"
				inventory.APIServingCertificateSANs = append(
					inventory.APIServingCertificateSANs,
					inventory.StableAPIIPv6,
				)
			},
		},
		{
			name:    "missing IPv6 SAN",
			blocker: "api_serving_certificate_sans_missing",
			mutate: func(inventory *StandInventory) {
				inventory.APIServingCertificateSANs = []string{"api.synthetic.example", "192.0.2.20"}
			},
		},
		{
			name:    "duplicate equivalent IPv6 SAN",
			blocker: "api_serving_certificate_sans_missing",
			mutate: func(inventory *StandInventory) {
				inventory.APIServingCertificateSANs = append(
					inventory.APIServingCertificateSANs,
					"2001:0db8:0:0:0:0:0:20",
				)
			},
		},
		{
			name:    "Cilium endpoint mismatch",
			blocker: "cilium_api_endpoint_mismatch",
			mutate: func(inventory *StandInventory) {
				inventory.CiliumAPIEndpoint = "api-other.synthetic.example:6443"
			},
		},
		{
			name:    "Cilium transport device missing",
			blocker: "cilium_control_plane_device_missing",
			mutate: func(inventory *StandInventory) {
				inventory.CiliumDevices = []string{"public0"}
			},
		},
		{
			name:    "parallel certificate rollout",
			blocker: "serving_certificate_rollout_unsafe",
			mutate: func(inventory *StandInventory) {
				inventory.ServingCertificateRolloutStrategy = "parallel"
			},
		},
		{
			name:    "certificate reconfiguration unverified",
			blocker: "serving_certificate_reconfiguration_unverified",
			mutate: func(inventory *StandInventory) {
				inventory.ServingCertificateReconfigurationReady = false
			},
		},
		{
			name:    "certificate rollback unverified",
			blocker: "serving_certificate_rollback_unverified",
			mutate: func(inventory *StandInventory) {
				inventory.ServingCertificateRollbackReady = false
			},
		},
		{
			name:    "API failover unverified",
			blocker: "control_plane_api_failover_unverified",
			mutate: func(inventory *StandInventory) {
				inventory.ControlPlaneAPIFailoverReady = false
			},
		},
		{
			name:    "duplicate node name",
			blocker: "duplicate_node_name",
			mutate: func(inventory *StandInventory) {
				inventory.Nodes[1].Name = inventory.Nodes[0].Name
			},
		},
		{
			name:    "duplicate node IPv4",
			blocker: "duplicate_node_ipv4",
			mutate: func(inventory *StandInventory) {
				inventory.Nodes[1].NodeIPv4 = inventory.Nodes[0].NodeIPv4
			},
		},
		{
			name:    "duplicate node IPv6",
			blocker: "duplicate_node_ipv6",
			mutate: func(inventory *StandInventory) {
				inventory.Nodes[1].NodeIPv6 = inventory.Nodes[0].NodeIPv6
			},
		},
		{
			name:    "duplicate equivalent node IPv6",
			blocker: "duplicate_node_ipv6",
			mutate: func(inventory *StandInventory) {
				inventory.Nodes[1].NodeIPv6 = "2001:0db8:0:0:0:0:0:11"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			inventory := readyInventory()
			test.mutate(&inventory)
			report, err := VerifyUpstreamStand(inventory)
			if err == nil {
				t.Fatal("unsafe control-plane API state passed stand verification")
			}
			assertBlocker(t, report.Blockers, test.blocker)
		})
	}
}

func syntheticBootstrapSpec() BootstrapSpec {
	return BootstrapSpec{
		ClusterName:                 "cloudring-synthetic-site",
		KubernetesVersion:           "v1.35.6",
		ControlPlaneEndpoint:        "api.synthetic.example:6443",
		StableAPIIPv4:               "192.0.2.20",
		StableAPIIPv6:               "2001:db8::20",
		APIServingCertificateSANs:   []string{"api.synthetic.example", "192.0.2.20", "2001:db8::20"},
		ControlPlaneReplicas:        3,
		EtcdTopology:                "stacked",
		PodCIDRs:                    []string{"192.0.2.0/24", "2001:db8:244::/56"},
		ServiceCIDRs:                []string{"198.51.100.0/24", "2001:db8:96::/108"},
		SurviveUnavailableServers:   1,
		CiliumVersionRef:            "oci://registry.example.invalid/cilium@sha256:redacted",
		CiliumDualStackMode:         "native-routing",
		CiliumAPIEndpoint:           "api.synthetic.example:6443",
		ControlPlaneTransportDevice: "transport0",
		CiliumDevices:               []string{"public0", "transport0"},
		ServingCertificateLifecycle: ServingCertificateLifecycle{
			RolloutStrategy:            "one-node-at-a-time",
			ReconfigurationPlanRef:     "operations.api-certificate.reconfigure",
			RollbackPlanRef:            "operations.api-certificate.rollback",
			OneServerLossAcceptanceRef: "resilience.api.one-server-loss",
		},
		ContainerRuntimeSocket:     "unix:///run/containerd/containerd.sock",
		CoreDNSMinReplicas:         3,
		CoreDNSTopologyKey:         "kubernetes.io/hostname",
		RequiredPodDisruptionNames: []string{"coredns", "provider-controller", "provider-portal"},
		Nodes: []NodeSpec{
			{Name: "node-a", AdvertiseIPv4: "192.0.2.11", AdvertiseIPv6: "2001:db8::11"},
			{Name: "node-b", AdvertiseIPv4: "192.0.2.12", AdvertiseIPv6: "2001:db8::12"},
			{Name: "node-c", AdvertiseIPv4: "192.0.2.13", AdvertiseIPv6: "2001:db8::13"},
		},
	}
}

func readyInventory() StandInventory {
	return StandInventory{
		Distribution:                           "upstream",
		Bootstrap:                              "kubeadm",
		ServerVersion:                          "v1.35.6",
		ControlPlaneEndpoint:                   "api.synthetic.example:6443",
		StableAPIIPv4:                          "192.0.2.20",
		StableAPIIPv6:                          "2001:db8::20",
		APIServingCertificateSANs:              []string{"api.synthetic.example", "192.0.2.20", "2001:db8::20"},
		ControlPlaneReplicas:                   3,
		EtcdTopology:                           "stacked",
		PodCIDRs:                               []string{"192.0.2.0/24", "2001:db8:244::/56"},
		ServiceCIDRs:                           []string{"198.51.100.0/24", "2001:db8:96::/108"},
		SurviveUnavailableServers:              1,
		CiliumDualStackReady:                   true,
		CiliumAPIEndpoint:                      "api.synthetic.example:6443",
		ControlPlaneTransportDevice:            "transport0",
		CiliumDevices:                          []string{"public0", "transport0"},
		ServingCertificateRolloutStrategy:      "one-node-at-a-time",
		ServingCertificateReconfigurationReady: true,
		ServingCertificateRollbackReady:        true,
		ControlPlaneAPIFailoverReady:           true,
		CoreDNSMinReplicas:                     3,
		CoreDNSSpreadReady:                     true,
		PodDisruptionBudgetsReady:              true,
		Nodes: []NodeInventory{
			{Name: "node-a", Ready: true, ControlPlane: true, EtcdMember: true, NodeIPv4: "192.0.2.11", NodeIPv6: "2001:db8::11"},
			{Name: "node-b", Ready: true, ControlPlane: true, EtcdMember: true, NodeIPv4: "192.0.2.12", NodeIPv6: "2001:db8::12"},
			{Name: "node-c", Ready: true, ControlPlane: true, EtcdMember: true, NodeIPv4: "192.0.2.13", NodeIPv6: "2001:db8::13"},
		},
		WorkflowContinuity: EvidenceInventory{
			Summary: "tenant API, console, GitOps, DNS, and certificate workflows verified",
			Items:   []string{"tenant-api-smoke", "console-smoke", "gitops-reconcile", "dns-aaaa", "certificate-renewal"},
		},
		DataDurability: EvidenceInventory{
			Summary: "etcd snapshot, Velero backup, object-store reachability, and restore drill verified",
			Items:   []string{"etcd-snapshot", "velero-backup", "object-store", "restore-drill"},
		},
		SinglePointOfFailure: EvidenceInventory{
			Summary: "three control-plane/etcd nodes and one-server-loss drill verified",
			Items:   []string{"three-control-plane", "stacked-etcd", "one-server-loss"},
		},
	}
}

func assertAction(t *testing.T, actions []NodeAction, name string) {
	t.Helper()
	for _, action := range actions {
		if action.Name == name {
			if !action.ReadOnlyPlanned {
				t.Fatalf("expected action %s to be read-only planned: %#v", name, action)
			}
			if action.Operation == "" {
				t.Fatalf("expected action %s to include operation contract: %#v", name, action)
			}
			return
		}
	}
	t.Fatalf("missing planned action %s in %#v", name, actions)
}

func assertActionRefs(t *testing.T, actions []NodeAction, name, reconfiguration, rollback, acceptance string) {
	t.Helper()
	for _, action := range actions {
		if action.Name != name {
			continue
		}
		if action.ReconfigurationPlanRef != reconfiguration ||
			action.RollbackPlanRef != rollback ||
			action.AcceptancePlanRef != acceptance {
			t.Fatalf("action %s lost lifecycle references: %#v", name, action)
		}
		return
	}
	t.Fatalf("missing planned action %s in %#v", name, actions)
}

func assertValidYAMLDocuments(t *testing.T, rendered string, expected int) {
	t.Helper()
	decoder := yaml.NewDecoder(strings.NewReader(rendered))
	count := 0
	for {
		var document map[string]any
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("rendered kubeadm document is invalid YAML: %v\n%s", err, rendered)
		}
		if len(document) == 0 {
			continue
		}
		count++
		if document["apiVersion"] != "kubeadm.k8s.io/v1beta4" {
			t.Fatalf("rendered document has unexpected apiVersion: %#v", document)
		}
	}
	if count != expected {
		t.Fatalf("expected %d YAML documents, decoded %d:\n%s", expected, count, rendered)
	}
}

func assertBlocker(t *testing.T, blockers []Blocker, id string) {
	t.Helper()
	for _, blocker := range blockers {
		if blocker.ID == id {
			return
		}
	}
	t.Fatalf("missing blocker %s in %#v", id, blockers)
}
