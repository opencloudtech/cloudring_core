// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeadm

import "errors"

var (
	// ErrInvalidBootstrapSpec marks an invalid or incomplete bootstrap input.
	ErrInvalidBootstrapSpec = errors.New("kubeadm bootstrap spec invalid")
	// ErrStandBlocked marks captured stand state which failed acceptance.
	ErrStandBlocked = errors.New("upstream stand readiness blocked")
)

// BootstrapSpec is the provider-resolved input to the deterministic renderer.
type BootstrapSpec struct {
	ClusterName                 string
	KubernetesVersion           string
	ControlPlaneEndpoint        string
	StableAPIIPv4               string
	StableAPIIPv6               string
	APIServingCertificateSANs   []string
	ControlPlaneReplicas        int
	EtcdTopology                string
	PodCIDRs                    []string
	ServiceCIDRs                []string
	SurviveUnavailableServers   int
	CiliumVersionRef            string
	CiliumDualStackMode         string
	CiliumAPIEndpoint           string
	ControlPlaneTransportDevice string
	CiliumDevices               []string
	ServingCertificateLifecycle ServingCertificateLifecycle
	ContainerRuntimeSocket      string
	CoreDNSMinReplicas          int
	CoreDNSTopologyKey          string
	RequiredPodDisruptionNames  []string
	Nodes                       []NodeSpec
}

// ServingCertificateLifecycle binds sequential certificate maintenance to
// rollback and one-server-loss acceptance workflows.
type ServingCertificateLifecycle struct {
	RolloutStrategy            string
	ReconfigurationPlanRef     string
	RollbackPlanRef            string
	OneServerLossAcceptanceRef string
}

// NodeSpec describes one control-plane node without credentials.
type NodeSpec struct {
	Name          string
	AdvertiseIPv4 string
	AdvertiseIPv6 string
	Roles         []string
	Labels        map[string]string
	Taints        []string
}

// BootstrapBundle contains rendered documents and non-executing operation
// metadata.
type BootstrapBundle struct {
	InitYAML                    string
	ControlPlaneJoinYAML        []JoinDocument
	Actions                     []NodeAction
	Cilium                      CiliumReadiness
	ServingCertificateLifecycle ServingCertificateLifecycle
	CoreDNS                     CoreDNSExpectation
	PodDisruptionBudgets        []string
}

// JoinDocument binds one node name to its source-safe join configuration.
type JoinDocument struct {
	NodeName string
	YAML     string
}

// NodeAction describes an ordered operation without executing it.
type NodeAction struct {
	Name                   string
	Description            string
	Operation              string
	Nodes                  []string
	ReadOnlyPlanned        bool
	Evidence               string
	ReconfigurationPlanRef string
	RollbackPlanRef        string
	AcceptancePlanRef      string
}

// CiliumReadiness captures the required API endpoint, device set, and checks.
type CiliumReadiness struct {
	DualStack                   bool
	VersionRef                  string
	Mode                        string
	APIEndpoint                 string
	ControlPlaneTransportDevice string
	Devices                     []string
	Checks                      []string
}

// CoreDNSExpectation describes the minimum replicated DNS topology.
type CoreDNSExpectation struct {
	MinReplicas int
	TopologyKey string
}

// StandInventory is sanitized observed state consumed by the independent
// verifier.
type StandInventory struct {
	Distribution                           string
	Bootstrap                              string
	ServerVersion                          string
	ControlPlaneEndpoint                   string
	StableAPIIPv4                          string
	StableAPIIPv6                          string
	APIServingCertificateSANs              []string
	ControlPlaneReplicas                   int
	EtcdTopology                           string
	PodCIDRs                               []string
	ServiceCIDRs                           []string
	SurviveUnavailableServers              int
	CiliumDualStackReady                   bool
	CiliumAPIEndpoint                      string
	ControlPlaneTransportDevice            string
	CiliumDevices                          []string
	ServingCertificateRolloutStrategy      string
	ServingCertificateReconfigurationReady bool
	ServingCertificateRollbackReady        bool
	ControlPlaneAPIFailoverReady           bool
	CoreDNSMinReplicas                     int
	CoreDNSSpreadReady                     bool
	PodDisruptionBudgetsReady              bool
	Nodes                                  []NodeInventory
	WorkflowContinuity                     EvidenceInventory
	DataDurability                         EvidenceInventory
	SinglePointOfFailure                   EvidenceInventory
}

// NodeInventory is sanitized observed state for one control-plane node.
type NodeInventory struct {
	Name         string
	Ready        bool
	ControlPlane bool
	EtcdMember   bool
	NodeIPv4     string
	NodeIPv6     string
}

// EvidenceInventory records source-safe evidence identifiers and a summary.
type EvidenceInventory struct {
	Summary string   `json:"summary"`
	Items   []string `json:"items"`
}

// StandReport is the fail-closed verification result.
type StandReport struct {
	Status               string            `json:"status"`
	Blockers             []Blocker         `json:"blockers,omitempty"`
	WorkflowContinuity   EvidenceInventory `json:"workflowContinuity"`
	DataDurability       EvidenceInventory `json:"dataDurability"`
	SinglePointOfFailure EvidenceInventory `json:"singlePointOfFailure"`
	Observed             StandInventory    `json:"observed"`
}

// Blocker identifies one failed acceptance condition.
type Blocker struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	Message  string `json:"message"`
}
