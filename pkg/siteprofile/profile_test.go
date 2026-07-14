// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package siteprofile

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSyntheticProfilePassesAndPlanIsDeterministic(t *testing.T) {
	profile := parseFixture(t, validProfileYAML)
	report := Validate(profile)
	if report.Status != "ready" || len(report.Blockers) != 0 || len(report.ProfileDigest) != 64 {
		t.Fatalf("unexpected report: %#v", report)
	}
	first, err := BuildPlan(profile)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPlan(profile)
	if err != nil {
		t.Fatal(err)
	}
	if first.ProfileDigest != second.ProfileDigest || len(first.Phases) != 7 || first.Phases[5].RollbackRef == "" {
		t.Fatalf("plan is incomplete or non-deterministic: %#v %#v", first, second)
	}
	if !contains(first.Phases[0].InputRefs, profile.Spec.ProviderAdapterRef) || !contains(first.Phases[0].InputRefs, profile.Spec.RegionRef) {
		t.Fatalf("inventory phase omits provider context: %#v", first.Phases[0].InputRefs)
	}
}

func TestProfileFailsClosedOnRequiredProductionInputs(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		replacement string
		blocker     string
	}{
		{name: "single stack", old: "dualStack: true", replacement: "dualStack: false", blocker: "dual_stack_network"},
		{name: "no immutable backup", old: "offCellBackup: true", replacement: "offCellBackup: false", blocker: "snapshot_and_off_cell_storage"},
		{name: "no rollback", old: "rollbackPlanRef: operations.rollback", replacement: "rollbackPlanRef: ''", blocker: "bootstrap_upgrade_rollback"},
		{name: "weak availability", old: "minimumFailureDomains: 3", replacement: "minimumFailureDomains: 2", blocker: "availability_policy"},
		{name: "unsafe reference", old: "regionRef: regions.synthetic", replacement: "regionRef: regions..synthetic", blocker: "provider_binding"},
		{name: "readiness overclaim", old: "nonClaim: preflight-and-plan-only", replacement: "nonClaim: production-ready", blocker: "readiness_non_claim"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile := parseFixture(t, strings.Replace(validProfileYAML, test.old, test.replacement, 1))
			report := Validate(profile)
			if report.Status != "blocked" || !contains(report.Blockers, test.blocker) {
				t.Fatalf("missing blocker %q: %#v", test.blocker, report)
			}
			if _, err := BuildPlan(profile); !errors.Is(err, ErrProfileBlocked) {
				t.Fatalf("blocked profile produced a plan: %v", err)
			}
		})
	}
}

func TestInventoryRequiresCriticalRolesAcrossDeclaredFailureDomains(t *testing.T) {
	profile := parseFixture(t, validProfileYAML)
	for index := range profile.Spec.Inventory.Nodes {
		profile.Spec.Inventory.Nodes[index].FailureDomain = "zone-a"
	}
	profile.Spec.Inventory.Nodes = append(profile.Spec.Inventory.Nodes,
		Node{
			ID:                     "gateway-b",
			FailureDomain:          "zone-b",
			Roles:                  []string{"gateway"},
			ProviderResourceRef:    "inventory.nodes.gateway-b",
			ManagementAddressRef:   "inventory.addresses.gateway-b.management",
			ProvisioningAddressRef: "inventory.addresses.gateway-b.provisioning",
		},
		Node{
			ID:                     "gateway-c",
			FailureDomain:          "zone-c",
			Roles:                  []string{"gateway"},
			ProviderResourceRef:    "inventory.nodes.gateway-c",
			ManagementAddressRef:   "inventory.addresses.gateway-c.management",
			ProvisioningAddressRef: "inventory.addresses.gateway-c.provisioning",
		},
	)

	report := Validate(profile)
	if report.Status != "blocked" || !contains(report.Blockers, "inventory") {
		t.Fatalf("gateway-only domains satisfied critical-role HA: %#v", report)
	}
}

func TestInventoryRejectsDuplicateProviderAndAddressReferences(t *testing.T) {
	tests := []struct {
		name string
		copy func(nodes []Node)
	}{
		{name: "provider resource", copy: func(nodes []Node) { nodes[1].ProviderResourceRef = nodes[0].ProviderResourceRef }},
		{name: "management address", copy: func(nodes []Node) { nodes[1].ManagementAddressRef = nodes[0].ManagementAddressRef }},
		{name: "provisioning address", copy: func(nodes []Node) { nodes[1].ProvisioningAddressRef = nodes[0].ProvisioningAddressRef }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			profile := parseFixture(t, validProfileYAML)
			test.copy(profile.Spec.Inventory.Nodes)
			report := Validate(profile)
			if report.Status != "blocked" || !contains(report.Blockers, "inventory") {
				t.Fatalf("duplicate reference passed inventory validation: %#v", report)
			}
		})
	}
}

func TestInvalidProfileNameIsNotEchoedInBlockedReport(t *testing.T) {
	profile := parseFixture(t, validProfileYAML)
	profile.Metadata.Name = "invalid_name"
	report := Validate(profile)
	if report.Status != "blocked" || report.ProfileName != "" || !contains(report.Blockers, "profile_identity") {
		t.Fatalf("invalid profile identity was echoed or accepted: %#v", report)
	}
}

func TestParserRejectsUnknownDuplicateAliasAndTrailingDocuments(t *testing.T) {
	tests := []string{
		strings.Replace(validProfileYAML, "siteClass: synthetic", "siteClass: synthetic\n  unknownField: rejected", 1),
		strings.Replace(validProfileYAML, "kind: ProviderSiteProfile", "kind: ProviderSiteProfile\nkind: ProviderSiteProfile", 1),
		strings.Replace(validProfileYAML, "name: synthetic-provider-site", "name: &site synthetic-provider-site", 1),
		validProfileYAML + "\n---\n{}\n",
	}
	for index, fixture := range tests {
		if _, err := Parse(strings.NewReader(fixture)); !errors.Is(err, ErrInvalidProfile) {
			t.Fatalf("fixture %d was accepted: %v", index, err)
		}
	}
}

func TestParserRejectsOversizedInput(t *testing.T) {
	if _, err := Parse(bytes.NewReader(bytes.Repeat([]byte{'x'}, maximumBytes+1))); !errors.Is(err, ErrProfileTooLarge) {
		t.Fatalf("oversized profile was accepted: %v", err)
	}
}

func parseFixture(t *testing.T, fixture string) Profile {
	t.Helper()
	profile, err := Parse(strings.NewReader(fixture))
	if err != nil {
		t.Fatal(err)
	}
	return profile
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

const validProfileYAML = `apiVersion: cloudring.io/v1alpha1
kind: ProviderSiteProfile
metadata:
  name: synthetic-provider-site
spec:
  providerAdapterRef: adapters.synthetic
  siteClass: synthetic
  regionRef: regions.synthetic
  availability:
    minimumControlPlaneNodes: 3
    minimumWorkerNodes: 3
    minimumFailureDomains: 3
  inventory:
    nodes:
      - id: node-a
        failureDomain: zone-a
        roles: [control-plane, worker]
        providerResourceRef: inventory.nodes.node-a
        managementAddressRef: inventory.addresses.node-a.management
        provisioningAddressRef: inventory.addresses.node-a.provisioning
      - id: node-b
        failureDomain: zone-b
        roles: [control-plane, worker]
        providerResourceRef: inventory.nodes.node-b
        managementAddressRef: inventory.addresses.node-b.management
        provisioningAddressRef: inventory.addresses.node-b.provisioning
      - id: node-c
        failureDomain: zone-c
        roles: [control-plane, worker]
        providerResourceRef: inventory.nodes.node-c
        managementAddressRef: inventory.addresses.node-c.management
        provisioningAddressRef: inventory.addresses.node-c.provisioning
  network:
    dualStack: true
    managementPlaneRef: networks.management
    provisioningPlaneRef: networks.provisioning
    tenantPlaneRef: networks.tenant
    publicIngressRef: networks.ingress
  storage:
    defaultClassRef: storage.default
    snapshotClassRef: storage.snapshots
    backupLocationRef: backup.off-cell
    immutableRetentionPolicyRef: backup.retention
    offCellBackup: true
  identity:
    oidcProviderRef: identity.oidc
    workloadIdentityRef: identity.workload
    runtimeInputBrokerRef: runtime-inputs.platform
  operations:
    gitOpsSourceRef: gitops.platform
    bootstrapPlanRef: operations.bootstrap
    upgradePlanRef: operations.upgrade
    rollbackPlanRef: operations.rollback
  observability:
    metricsRef: observability.metrics
    logsRef: observability.logs
    tracesRef: observability.traces
    alertsRef: observability.alerts
  ocs:
    version: v3
    conformanceProfileRef: ocs.provider-conformance
  nonClaim: preflight-and-plan-only
`
