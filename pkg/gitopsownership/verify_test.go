// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package gitopsownership

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVerifyProvesUniqueInventoryOwnershipWithoutChangingPrune(t *testing.T) {
	contract := testContract()
	target := contract.Spec.DriftProof.Target
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	snapshot := Snapshot{
		Kustomizations: []KustomizationSnapshot{
			testRoot(t, "foundation", false, target, service),
		},
		Resources: map[string][]ResourceRef{
			"gateways": {target},
			"services": {service},
		},
		SourceArtifact:           SourceArtifactSnapshot{Kind: "GitRepository", Namespace: "flux-system", Name: "source", Generation: 1, ObservedGeneration: 1, Ready: true, Revision: testSourceRevision, Digest: testArtifactDigest},
		AcceptedSourceRevision:   testSourceRevision,
		AcceptedArtifactDigest:   testArtifactDigest,
		ObservedPublicGitlinkSHA: contract.Spec.SourceArtifact.PublicGitlinkSHA,
	}

	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if code != 0 || report.Status != "ready" {
		t.Fatalf("Verify code/status = %d/%s, want 0/ready; blockers=%+v", code, report.Status, report.Blockers)
	}
	if report.APIVersion != "cloudring.gitops.ownership/v1" || !report.Scope.Complete || !report.Scope.AllCriticalFamiliesDeclared || !report.Scope.AllSelectedRootsObserved {
		t.Fatalf("report envelope/scope is not gate-consumable: %+v", report)
	}
	if !report.PruneEligible || report.PruneChanged || report.LiveMutationPerformed || report.DriftProof.LiveMutationPerformed {
		t.Fatalf("unexpected mutation/prune report: %+v", report)
	}
	if !report.DriftProof.TargetInventoryOwned {
		t.Fatal("drift target ownership was not proven")
	}
}

func TestVerifyRejectsRootSpecAndSourceReceiptDrift(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Snapshot)
		blocker string
	}{
		{"missing suspend", func(snapshot *Snapshot) { snapshot.Kustomizations[0].Suspend = nil }, "flux_kustomization_suspended"},
		{"suspended", func(snapshot *Snapshot) { value := true; snapshot.Kustomizations[0].Suspend = &value }, "flux_kustomization_suspended"},
		{"deletion policy", func(snapshot *Snapshot) { snapshot.Kustomizations[0].DeletionPolicy = "Delete" }, "flux_deletion_policy_mismatch"},
		{"missing wait", func(snapshot *Snapshot) { snapshot.Kustomizations[0].Wait = nil }, "flux_wait_mismatch"},
		{"source ref", func(snapshot *Snapshot) { snapshot.Kustomizations[0].SourceRef.Name = "other" }, "flux_source_ref_mismatch"},
		{"path", func(snapshot *Snapshot) { snapshot.Kustomizations[0].Path = "./other" }, "flux_path_mismatch"},
		{"dependencies", func(snapshot *Snapshot) {
			snapshot.Kustomizations[0].DependsOn = []FluxObjectReference{{Kind: "Kustomization", Namespace: "flux-system", Name: "other"}}
		}, "flux_dependencies_mismatch"},
		{"last applied revision", func(snapshot *Snapshot) {
			snapshot.Kustomizations[0].LastAppliedRevision = "main@sha1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}, "flux_last_applied_revision_mismatch"},
		{"artifact revision", func(snapshot *Snapshot) {
			snapshot.SourceArtifact.Revision = "main@sha1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}, "source_artifact_revision_mismatch"},
		{"artifact digest", func(snapshot *Snapshot) {
			snapshot.SourceArtifact.Digest = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		}, "source_artifact_digest_mismatch"},
		{"source not ready", func(snapshot *Snapshot) { snapshot.SourceArtifact.Ready = false }, "source_artifact_not_ready"},
		{"missing accepted receipt", func(snapshot *Snapshot) { snapshot.AcceptedSourceRevision = "" }, "accepted_source_receipt_invalid"},
		{"missing public gitlink", func(snapshot *Snapshot) { snapshot.ObservedPublicGitlinkSHA = "" }, "observed_public_gitlink_invalid"},
		{"wrong public gitlink", func(snapshot *Snapshot) { snapshot.ObservedPublicGitlinkSHA = strings.Repeat("b", 40) }, "public_gitlink_mismatch"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := testContract()
			snapshot := readyTestSnapshot(t, contract)
			test.mutate(&snapshot)
			report, code, err := Verify(contract, snapshot)
			if err != nil {
				t.Fatalf("Verify returned error: %v", err)
			}
			if code != 2 || report.PruneEligible {
				t.Fatalf("drift code/pruneEligible = %d/%t, want 2/false", code, report.PruneEligible)
			}
			assertBlocker(t, report, test.blocker)
		})
	}
}

func TestVerifyFailsClosedForUnmanagedObjectWithEmptyDefaultAllowlist(t *testing.T) {
	contract := testContract()
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	snapshot := Snapshot{
		Kustomizations: []KustomizationSnapshot{testRoot(t, "foundation", false, contract.Spec.DriftProof.Target)},
		Resources: map[string][]ResourceRef{
			"gateways": {contract.Spec.DriftProof.Target},
			"services": {service},
		},
	}

	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if code != 2 || report.PruneEligible {
		t.Fatalf("Verify code/pruneEligible = %d/%t, want 2/false", code, report.PruneEligible)
	}
	assertBlocker(t, report, "resource_not_flux_managed")
}

func TestVerifyAcceptsExactAllowlistButRejectsStaleEntry(t *testing.T) {
	contract := testContract()
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	contract.Spec.AllowUnmanaged = []ResourceRef{service}
	snapshot := Snapshot{
		Kustomizations: []KustomizationSnapshot{testRoot(t, "foundation", false, contract.Spec.DriftProof.Target)},
		Resources: map[string][]ResourceRef{
			"gateways": {contract.Spec.DriftProof.Target},
			"services": {service},
		},
	}
	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if code != 2 {
		t.Fatalf("all-allowlisted family code = %d, want 2", code)
	}
	assertBlocker(t, report, "required_family_has_no_managed_objects")

	managedServiceRoot := testRoot(t, "foundation", false, contract.Spec.DriftProof.Target, service)
	snapshot.Kustomizations = []KustomizationSnapshot{managedServiceRoot}
	report, code, err = Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify managed allowlisted object returned error: %v", err)
	}
	if code != 2 {
		t.Fatalf("managed allowlisted object code = %d, want 2", code)
	}
	assertBlocker(t, report, "managed_object_allowlisted")
	assertBlocker(t, report, "stale_unmanaged_allowlist_entry")
}

func TestVerifyRejectsDuplicateLiveResourceObservation(t *testing.T) {
	contract := testContract()
	snapshot := readyTestSnapshot(t, contract)
	snapshot.Resources["gateways"] = append(snapshot.Resources["gateways"], contract.Spec.DriftProof.Target)

	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify duplicate resource returned error: %v", err)
	}
	if code != 2 || report.PruneEligible {
		t.Fatalf("duplicate resource code/pruneEligible = %d/%t, want 2/false", code, report.PruneEligible)
	}
	assertBlocker(t, report, "duplicate_resource_observation")
}

func TestVerifyRejectsPartiallyAllowlistedCriticalFamily(t *testing.T) {
	contract := testContract()
	service := contract.Spec.RequiredFamilies[1].ExpectedObjects[0]
	second := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api-secondary"}
	contract.Spec.RequiredFamilies[1].ExpectedObjects = append(contract.Spec.RequiredFamilies[1].ExpectedObjects, second)
	contract.Spec.RequiredFamilies[1].MinimumCount = 2
	contract.Spec.AllowUnmanaged = []ResourceRef{second}
	snapshot := readyTestSnapshot(t, contract)
	snapshot.Resources["services"] = []ResourceRef{service, second}

	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify partially allowlisted family returned error: %v", err)
	}
	if code != 2 || report.PruneEligible {
		t.Fatalf("partially allowlisted family code/pruneEligible = %d/%t, want 2/false", code, report.PruneEligible)
	}
	assertBlocker(t, report, "critical_family_ownership_incomplete")
}

func TestVerifyBlocksDuplicateOwnershipAndPrematurePrune(t *testing.T) {
	contract := testContract()
	target := contract.Spec.DriftProof.Target
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	second := testRoot(t, "second", false, target)
	contract.Spec.PruneGate.Kustomizations = append(contract.Spec.PruneGate.Kustomizations, NamespacedName{Namespace: "flux-system", Name: "second"})
	contract.Spec.Scope.SelectedRoots = append(contract.Spec.Scope.SelectedRoots, NamespacedName{Namespace: "flux-system", Name: "second"})
	active, prune, wait := false, false, true
	contract.Spec.RootSpecs = append(contract.Spec.RootSpecs, RootSpecContract{
		Namespace: "flux-system", Name: "second", Suspend: &active, Prune: &prune, DeletionPolicy: "Orphan", Wait: &wait,
		SourceRef: FluxObjectReference{Kind: "GitRepository", Namespace: "flux-system", Name: "source"}, Path: "./second", DependsOn: []FluxObjectReference{},
	})
	snapshot := Snapshot{
		Kustomizations: []KustomizationSnapshot{testRoot(t, "foundation", true, target, service), second},
		Resources: map[string][]ResourceRef{
			"gateways": {target},
			"services": {service},
		},
	}
	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if code != 2 {
		t.Fatalf("Verify code = %d, want 2", code)
	}
	assertBlocker(t, report, "multiple_flux_owners")
	assertBlocker(t, report, "prune_enabled_before_gate")
	assertBlocker(t, report, "drift_target_ownership_unproven")
}

func TestDecodeContractRejectsUnknownAndTrailingFields(t *testing.T) {
	valid := `{
		"apiVersion":"cloudring.io/v1alpha1",
		"kind":"GitOpsOwnershipContract",
		"metadata":{"name":"test"},
		"spec":{
			"kustomizationSelector":"cloudring.io/ownership-scope=ovh-alpha",
			"sourceArtifact":{"kind":"GitRepository","namespace":"flux-system","name":"source","publicGitlinkSHA":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
			"scope":{"name":"goal01","complete":true,"nonClaim":"declared scope is not ownership readiness","criticalFamilyIds":["services"],"selectedRoots":[{"namespace":"flux-system","name":"foundation"}]},
			"rootSpecs":[{"namespace":"flux-system","name":"foundation","suspend":false,"prune":false,"deletionPolicy":"Orphan","wait":true,"sourceRef":{"kind":"GitRepository","namespace":"flux-system","name":"source"},"path":"./foundation","dependsOn":[]}],
			"requiredFamilies":[{"id":"services","apiVersion":"v1","kind":"Service","resource":"services","namespaced":true,"labelSelector":"cloudring.io/installation=test","minimumCount":1,"critical":true,"sourceState":"sourced","expectedOwner":{"namespace":"flux-system","name":"foundation"},"expectedObjects":[{"apiVersion":"v1","kind":"Service","namespace":"default","name":"test"}]}],
			"pruneGate":{"kustomizations":[{"namespace":"flux-system","name":"foundation"}]},
			"driftProof":{"mode":"read-only-plan","target":{"apiVersion":"v1","kind":"Service","namespace":"default","name":"test"},"expectedOwner":{"namespace":"flux-system","name":"foundation"},"requiredObservations":["baseline","controlled-drift","reconciled"],"mutationAuthorization":"separate-explicit-approval-required"}
		}
	}`
	if _, err := DecodeContract(strings.NewReader(valid)); err != nil {
		t.Fatalf("DecodeContract valid returned error: %v", err)
	}
	if _, err := DecodeContract(strings.NewReader(strings.Replace(valid, `"metadata":{"name":"test"}`, `"metadata":{"name":"test","unknown":true}`, 1))); err == nil {
		t.Fatal("DecodeContract accepted unknown field")
	}
	if _, err := DecodeContract(strings.NewReader(strings.Replace(valid, `"metadata":{"name":"test"}`, `"metadata":{"name":"test","name":"duplicate"}`, 1))); err == nil {
		t.Fatal("DecodeContract accepted duplicate object field")
	}
	if _, err := DecodeContract(strings.NewReader(valid + `{}`)); err == nil {
		t.Fatal("DecodeContract accepted trailing JSON")
	}
	if _, err := DecodeContract(strings.NewReader(strings.Repeat(" ", (8<<20)+1))); err == nil {
		t.Fatal("DecodeContract accepted an oversized document")
	}
}

func TestVerifyDeclaresMissingCriticalSourceAsBlocker(t *testing.T) {
	contract := testContract()
	contract.Spec.Scope.CriticalFamilyIDs = append(contract.Spec.Scope.CriticalFamilyIDs, "portal")
	contract.Spec.RequiredFamilies = append(contract.Spec.RequiredFamilies, ResourceFamily{
		ID: "portal", APIVersion: "apps/v1", Kind: "Deployment", Resource: "deployments.apps", Namespaced: true,
		Critical: true, SourceState: "not-sourced", MissingSourceBlocker: "provider portal is not sourced by a selected Goal 01 Flux root",
	})
	target := contract.Spec.DriftProof.Target
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	report, code, err := Verify(contract, Snapshot{
		Kustomizations: []KustomizationSnapshot{testRoot(t, "foundation", false, target, service)},
		Resources:      map[string][]ResourceRef{"gateways": {target}, "services": {service}},
	})
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if code != 2 || len(report.Scope.SourceMissingFamilies) != 1 || report.Scope.SourceMissingFamilies[0] != "portal" {
		t.Fatalf("missing source was not reported fail-closed: code=%d report=%+v", code, report)
	}
	assertBlocker(t, report, "critical_family_source_missing")
}

func TestValidateContractRejectsIncompleteOrPartiallyDeclaredScope(t *testing.T) {
	contract := testContract()
	contract.Spec.Scope.Complete = false
	if err := ValidateContract(contract); err == nil {
		t.Fatal("ValidateContract accepted scope.complete=false")
	}
	contract = testContract()
	contract.Spec.Scope.CriticalFamilyIDs = contract.Spec.Scope.CriticalFamilyIDs[:1]
	if err := ValidateContract(contract); err == nil {
		t.Fatal("ValidateContract accepted a partially enumerated critical family scope")
	}
	contract = testContract()
	contract.Spec.Scope.SelectedRoots = nil
	if err := ValidateContract(contract); err == nil {
		t.Fatal("ValidateContract accepted an empty selected root inventory")
	}
}

func TestValidateContractBindsDriftProofToOneDeclaredCriticalObject(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Contract)
	}{
		{
			name: "undeclared target",
			mutate: func(contract *Contract) {
				contract.Spec.DriftProof.Target.Name = "not-declared"
			},
		},
		{
			name: "wrong owner",
			mutate: func(contract *Contract) {
				contract.Spec.DriftProof.ExpectedOwner.Name = "other-root"
			},
		},
		{
			name: "duplicate target across families",
			mutate: func(contract *Contract) {
				contract.Spec.RequiredFamilies[1].APIVersion = contract.Spec.RequiredFamilies[0].APIVersion
				contract.Spec.RequiredFamilies[1].Kind = contract.Spec.RequiredFamilies[0].Kind
				contract.Spec.RequiredFamilies[1].Resource = contract.Spec.RequiredFamilies[0].Resource
				contract.Spec.RequiredFamilies[1].ExpectedObjects = []ResourceRef{contract.Spec.DriftProof.Target}
			},
		},
		{
			name: "noncanonical observation sequence",
			mutate: func(contract *Contract) {
				contract.Spec.DriftProof.RequiredObservations = []string{"baseline", "reconciled", "controlled-drift"}
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := testContract()
			test.mutate(&contract)
			if err := ValidateContract(contract); err == nil {
				t.Fatal("ValidateContract accepted an unbound drift proof")
			}
		})
	}
}

func TestValidateContractRejectsExpectedObjectSharedAcrossFamilies(t *testing.T) {
	contract := testContract()
	contract.Spec.RequiredFamilies[1].APIVersion = contract.Spec.RequiredFamilies[0].APIVersion
	contract.Spec.RequiredFamilies[1].Kind = contract.Spec.RequiredFamilies[0].Kind
	contract.Spec.RequiredFamilies[1].Resource = contract.Spec.RequiredFamilies[0].Resource
	contract.Spec.RequiredFamilies[1].ExpectedObjects = []ResourceRef{contract.Spec.RequiredFamilies[0].ExpectedObjects[0]}
	if err := ValidateContract(contract); err == nil {
		t.Fatal("ValidateContract accepted one expected object in multiple resource families")
	}
}

func testContract() Contract {
	active, prune, wait := false, false, true
	return Contract{
		APIVersion: "cloudring.io/v1alpha1",
		Kind:       "GitOpsOwnershipContract",
		Metadata:   Metadata{Name: "test"},
		Spec: ContractSpec{
			KustomizationSelector: "cloudring.io/ownership-scope=ovh-alpha",
			SourceArtifact:        SourceArtifactContract{Kind: "GitRepository", Namespace: "flux-system", Name: "source", PublicGitlinkSHA: strings.Repeat("a", 40)},
			RootSpecs: []RootSpecContract{{
				Namespace: "flux-system", Name: "foundation", Suspend: &active, Prune: &prune, DeletionPolicy: "Orphan", Wait: &wait,
				SourceRef: FluxObjectReference{Kind: "GitRepository", Namespace: "flux-system", Name: "source"}, Path: "./foundation", DependsOn: []FluxObjectReference{},
			}},
			Scope: ScopeContract{
				Name: "goal01", Complete: true, NonClaim: "declared scope is not ownership readiness",
				CriticalFamilyIDs: []string{"gateways", "services"},
				SelectedRoots:     []NamespacedName{{Namespace: "flux-system", Name: "foundation"}},
			},
			RequiredFamilies: []ResourceFamily{
				{
					ID: "gateways", APIVersion: "gateway.networking.k8s.io/v1", Kind: "Gateway", Resource: "gateways.gateway.networking.k8s.io", Namespaced: true,
					LabelSelector: "cloudring.io/installation=test", MinimumCount: 1, Critical: true, SourceState: "sourced",
					ExpectedOwner:   NamespacedName{Namespace: "flux-system", Name: "foundation"},
					ExpectedObjects: []ResourceRef{{APIVersion: "gateway.networking.k8s.io/v1", Kind: "Gateway", Namespace: "platform-system", Name: "cloudring-public"}},
				},
				{
					ID: "services", APIVersion: "v1", Kind: "Service", Resource: "services", Namespaced: true,
					LabelSelector: "cloudring.io/installation=test", MinimumCount: 1, Critical: true, SourceState: "sourced",
					ExpectedOwner:   NamespacedName{Namespace: "flux-system", Name: "foundation"},
					ExpectedObjects: []ResourceRef{{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}},
				},
			},
			PruneGate: PruneGate{Kustomizations: []NamespacedName{{Namespace: "flux-system", Name: "foundation"}}},
			DriftProof: DriftProof{
				Mode:                  "read-only-plan",
				Target:                ResourceRef{APIVersion: "gateway.networking.k8s.io/v1", Kind: "Gateway", Namespace: "platform-system", Name: "cloudring-public"},
				ExpectedOwner:         NamespacedName{Namespace: "flux-system", Name: "foundation"},
				RequiredObservations:  []string{"baseline", "controlled-drift", "reconciled"},
				MutationAuthorization: "separate-explicit-approval-required",
			},
		},
	}
}

func testRoot(t *testing.T, name string, prune bool, refs ...ResourceRef) KustomizationSnapshot {
	t.Helper()
	inventory := make([]InventoryEntry, 0, len(refs))
	for _, ref := range refs {
		entry, err := InventoryID(ref)
		if err != nil {
			t.Fatalf("InventoryID(%+v): %v", ref, err)
		}
		inventory = append(inventory, entry)
	}
	active, wait := false, true
	return KustomizationSnapshot{
		Namespace: "flux-system", Name: name, Generation: 1, ObservedGeneration: 1, Ready: true, Suspend: &active, Prune: &prune,
		DeletionPolicy: "Orphan", Wait: &wait, SourceRef: FluxObjectReference{Kind: "GitRepository", Namespace: "flux-system", Name: "source"},
		Path: "./" + name, DependsOn: []FluxObjectReference{}, LastAppliedRevision: testSourceRevision, Inventory: inventory,
	}
}

func readyTestSnapshot(t *testing.T, contract Contract) Snapshot {
	t.Helper()
	target := contract.Spec.DriftProof.Target
	service := ResourceRef{APIVersion: "v1", Kind: "Service", Namespace: "platform-system", Name: "provider-api"}
	return Snapshot{
		Kustomizations:           []KustomizationSnapshot{testRoot(t, "foundation", false, target, service)},
		Resources:                map[string][]ResourceRef{"gateways": {target}, "services": {service}},
		SourceArtifact:           SourceArtifactSnapshot{Kind: "GitRepository", Namespace: "flux-system", Name: "source", Generation: 1, ObservedGeneration: 1, Ready: true, Revision: testSourceRevision, Digest: testArtifactDigest},
		AcceptedSourceRevision:   testSourceRevision,
		AcceptedArtifactDigest:   testArtifactDigest,
		ObservedPublicGitlinkSHA: contract.Spec.SourceArtifact.PublicGitlinkSHA,
	}
}

func TestVerifyDoesNotReflectMalformedLiveSourceOrInventoryValues(t *testing.T) {
	contract := testContract()
	snapshot := readyTestSnapshot(t, contract)
	secretLike := "private-token-value_ssh://internal"
	snapshot.SourceArtifact.Namespace = secretLike
	snapshot.SourceArtifact.Name = secretLike
	snapshot.SourceArtifact.Revision = secretLike
	snapshot.SourceArtifact.Digest = secretLike
	snapshot.Kustomizations[0].LastAppliedRevision = secretLike
	snapshot.Kustomizations[0].Inventory = append(snapshot.Kustomizations[0].Inventory, InventoryEntry{ID: secretLike, Version: secretLike})
	snapshot.Kustomizations = append(snapshot.Kustomizations, KustomizationSnapshot{
		Namespace: secretLike, Name: secretLike, Generation: 1, ObservedGeneration: 1,
	})
	snapshot.Resources["services"] = append(snapshot.Resources["services"], ResourceRef{
		APIVersion: secretLike, Kind: secretLike, Namespace: secretLike, Name: secretLike,
	})

	report, code, err := Verify(contract, snapshot)
	if err != nil {
		t.Fatalf("Verify malformed live evidence returned error: %v", err)
	}
	if code != 2 || report.PruneEligible {
		t.Fatalf("malformed live evidence code/pruneEligible = %d/%t, want 2/false", code, report.PruneEligible)
	}
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), secretLike) {
		t.Fatal("report reflected malformed live evidence")
	}
}

const testSourceRevision = "main@sha1:0123456789abcdef0123456789abcdef01234567"
const testArtifactDigest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func assertBlocker(t *testing.T, report Report, id string) {
	t.Helper()
	for _, blocker := range report.Blockers {
		if blocker.ID == id {
			return
		}
	}
	t.Fatalf("blocker %q missing from %+v", id, report.Blockers)
}
