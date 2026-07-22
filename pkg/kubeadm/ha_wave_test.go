// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeadm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
	"github.com/opencloudtech/CloudRING/pkg/resilience/oneserverloss"
)

func TestHAWaveSequencesOneToTwoThenFreshBackupThenTwoToThree(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	first := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID:    "synthetic-cell-001",
		Current:           haWaveCounts(1),
		BackupBarrierPath: filepath.Join("private", "backup-wave-1.json"),
		Now:               fixedHAWaveClock(base),
		ValidateBackup:    passingHAWaveBackup(base.Add(-time.Minute), "backup-1"),
	})
	if first.Status != "ready" || first.Phase != HAWavePhaseOneToTwo || first.Target != haWaveCounts(2) || first.MutationAllowed {
		t.Fatalf("unexpected first plan: %+v", first)
	}
	if got := first.Checks[1].Artifact; got != "backup-wave-1.json" {
		t.Fatalf("private backup path was not sanitized: %q", got)
	}
	firstPlanPath := writeHAWaveJSON(t, "first-plan.json", first)
	firstVerification := VerifyHAWave(haWaveVerifyOptions(t, firstPlanPath, "backup-wave-1.json", base.Add(time.Minute), 2, nil))
	if firstVerification.Status != "ready" || firstVerification.FinalOneServerLossVerified || firstVerification.MutationAttempted {
		t.Fatalf("unexpected first verification: %+v", firstVerification)
	}
	firstVerificationPath := writeHAWaveJSON(t, "first-verification.json", firstVerification)

	second := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID:    "synthetic-cell-001",
		Current:           haWaveCounts(2),
		BackupBarrierPath: "backup-wave-2.json",
		PreviousWavePath:  firstVerificationPath,
		Now:               fixedHAWaveClock(base.Add(3 * time.Minute)),
		ValidateBackup:    passingHAWaveBackup(base.Add(2*time.Minute), "backup-2"),
	})
	if second.Status != "ready" || second.Phase != HAWavePhaseTwoToThree || second.Target != haWaveCounts(3) ||
		!hasPassingHAWaveCheck(second.Checks, "previous_wave") {
		t.Fatalf("unexpected second plan: %+v", second)
	}
	secondPlanPath := writeHAWaveJSON(t, "second-plan.json", second)
	receipt := testOneServerLossReceipt()
	finalOptions := haWaveVerifyOptions(t, secondPlanPath, "backup-wave-2.json", base.Add(4*time.Minute), 3, &receipt)
	finalOptions.FinalOneServerLossPath = filepath.Join("protected", "one-server-loss.json")
	finalOptions.LoadOneServerLossReceipt = func(string, time.Time) (*oneserverloss.Receipt, string, error) {
		return &receipt, testSHA256("receipt-artifact"), nil
	}
	final := VerifyHAWave(finalOptions)
	if final.Status != "ready" || !final.FinalOneServerLossVerified || final.MutationAttempted {
		t.Fatalf("unexpected final verification: %+v", final)
	}
	if got := final.Checks[len(final.Checks)-1].Artifact; got != "one-server-loss.json" {
		t.Fatalf("private receipt path was not sanitized: %q", got)
	}
	finalPath := writeHAWaveJSON(t, "final-verification.json", final)
	if _, _, err := ReadHAWaveVerification(finalPath); err != nil {
		t.Fatalf("ready final verification did not round trip: %v", err)
	}
}

func TestHAWaveBlocksTwoToThreeWithoutNewerInterwaveBackup(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	firstPlan := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(1), BackupBarrierPath: "backup-1.json",
		Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base.Add(-time.Minute), "backup-1"),
	})
	firstPlanPath := writeHAWaveJSON(t, "first-plan.json", firstPlan)
	firstVerification := VerifyHAWave(haWaveVerifyOptions(t, firstPlanPath, "backup-1.json", base.Add(time.Minute), 2, nil))
	previousPath := writeHAWaveJSON(t, "first-verification.json", firstVerification)
	second := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(2), BackupBarrierPath: "stale-backup.json",
		PreviousWavePath: previousPath, Now: fixedHAWaveClock(base.Add(2 * time.Minute)),
		ValidateBackup: passingHAWaveBackup(base.Add(30*time.Second), "stale-backup"),
	})
	assertHAWaveBlocker(t, second.Blockers, "ha_wave_interwave_backup_not_fresh")
}

func TestHAWaveBlocksPreviousVerificationFromAnotherInstallation(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	firstPlan := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(1), BackupBarrierPath: "backup-1.json",
		Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base.Add(-time.Minute), "backup-1"),
	})
	firstPlanPath := writeHAWaveJSON(t, "first-plan.json", firstPlan)
	previous := VerifyHAWave(haWaveVerifyOptions(t, firstPlanPath, "backup-1.json", base.Add(time.Minute), 2, nil))
	previous.InstallationID = "synthetic-cell-002"
	previousPath := writeHAWaveJSON(t, "other-cell-verification.json", previous)
	second := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(2), BackupBarrierPath: "backup-2.json",
		PreviousWavePath: previousPath, Now: fixedHAWaveClock(base.Add(3 * time.Minute)),
		ValidateBackup: passingHAWaveBackup(base.Add(2*time.Minute), "backup-2"),
	})
	assertHAWaveBlocker(t, second.Blockers, "ha_wave_previous_verification_invalid")
}

func TestHAWaveFinalOneServerLossEvidenceIsForbiddenBeforeThreeAndRequiredAfter(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	firstPlan := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(1), BackupBarrierPath: "backup-1.json",
		Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base.Add(-time.Minute), "backup-1"),
	})
	firstPath := writeHAWaveJSON(t, "first-plan.json", firstPlan)
	earlyOptions := haWaveVerifyOptions(t, firstPath, "backup-1.json", base, 2, nil)
	earlyOptions.FinalOneServerLossPath = "too-early.json"
	early := VerifyHAWave(earlyOptions)
	assertHAWaveBlocker(t, early.Blockers, "ha_wave_final_one_server_loss_too_early")

	secondPlan := readyHAWavePlanForTest(base, HAWavePhaseTwoToThree)
	secondPath := writeHAWaveJSON(t, "second-plan.json", secondPlan)
	missing := VerifyHAWave(haWaveVerifyOptions(t, secondPath, "backup.json", base, 3, nil))
	assertHAWaveBlocker(t, missing.Blockers, "ha_wave_final_one_server_loss_invalid")
}

func TestHAWaveFinalReceiptRemainsAuthoritativeAndFresh(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	planPath := writeHAWaveJSON(t, "second-plan.json", readyHAWavePlanForTest(base, HAWavePhaseTwoToThree))
	receipt := testOneServerLossReceipt()
	tests := []struct {
		name string
		now  time.Time
		load HAWaveOneServerLossLoader
	}{
		{
			name: "loader cannot replace receipt with a digest",
			now:  base,
			load: func(string, time.Time) (*oneserverloss.Receipt, string, error) {
				return nil, testSHA256("receipt-artifact"), nil
			},
		},
		{
			name: "invalid artifact digest",
			now:  base,
			load: func(string, time.Time) (*oneserverloss.Receipt, string, error) {
				return &receipt, "not-a-digest", nil
			},
		},
		{
			name: "stale valid receipt",
			now:  base.Add(maximumHAWaveEvidenceAge + time.Hour),
			load: func(string, time.Time) (*oneserverloss.Receipt, string, error) {
				return &receipt, testSHA256("receipt-artifact"), nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := haWaveVerifyOptions(t, planPath, "backup.json", test.now, 3, &receipt)
			opts.FinalOneServerLossPath = "receipt.json"
			opts.LoadOneServerLossReceipt = test.load
			report := VerifyHAWave(opts)
			assertHAWaveBlocker(t, report.Blockers, "ha_wave_final_one_server_loss_invalid")
		})
	}
}

func TestHAWaveVerificationReopensExactBackupAndRequiresInventoryReceipt(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	planPath := writeHAWaveJSON(t, "first-plan.json", readyHAWavePlanForTest(base, HAWavePhaseOneToTwo))

	missingBackup := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
	missingBackup.ValidateBackup = nil
	report := VerifyHAWave(missingBackup)
	assertHAWaveBlocker(t, report.Blockers, "ha_wave_backup_revalidation_invalid")

	wrongBackup := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
	wrongBackup.ValidateBackup = passingHAWaveBackup(base.Add(-time.Minute), "different-backup")
	report = VerifyHAWave(wrongBackup)
	assertHAWaveBlocker(t, report.Blockers, "ha_wave_backup_revalidation_invalid")

	missingInventory := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
	missingInventory.LoadInventory = nil
	report = VerifyHAWave(missingInventory)
	assertHAWaveBlocker(t, report.Blockers, "ha_wave_inventory_receipt_invalid")
	if report.Observed != (HAWaveCounts{}) {
		t.Fatalf("counts were accepted without an inventory receipt: %+v", report.Observed)
	}

	fabricatedInventory := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
	fabricatedInventory.LoadInventory = func(string, time.Time) (*HAWaveInventoryReceipt, string, error) {
		inventory := testHAWaveInventoryReceipt(base, "synthetic-cell-001", 2, nil)
		inventory.Members[0].Ready = false // ReceiptSHA256 is deliberately not recomputed.
		return inventory, testSHA256("inventory-artifact"), nil
	}
	report = VerifyHAWave(fabricatedInventory)
	assertHAWaveBlocker(t, report.Blockers, "ha_wave_inventory_receipt_invalid")
}

func TestHAWaveFinalReceiptMustMatchInventoryDigestAndIdentityBinding(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	planPath := writeHAWaveJSON(t, "second-plan.json", readyHAWavePlanForTest(base, HAWavePhaseTwoToThree))
	receipt := testOneServerLossReceipt()
	opts := haWaveVerifyOptions(t, planPath, "backup.json", base, 3, &receipt)
	opts.FinalOneServerLossPath = "receipt.json"
	opts.LoadOneServerLossReceipt = func(string, time.Time) (*oneserverloss.Receipt, string, error) {
		return &receipt, testSHA256("receipt-artifact"), nil
	}
	originalLoader := opts.LoadInventory
	opts.LoadInventory = func(path string, now time.Time) (*HAWaveInventoryReceipt, string, error) {
		inventory, sha, err := originalLoader(path, now)
		if err != nil {
			return nil, "", err
		}
		copy := *inventory
		copy.OneServerLossReceiptBinding.TargetNodeUIDSHA256 = testSHA256("different-node")
		copy.ReceiptSHA256 = HAWaveInventoryReceiptSHA256(copy)
		return &copy, sha, nil
	}
	report := VerifyHAWave(opts)
	assertHAWaveBlocker(t, report.Blockers, "ha_wave_final_one_server_loss_invalid")
}

func TestHAWavePlanFailsClosedOnInvalidInputs(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	tests := []struct {
		name    string
		opts    HAWavePlanOptions
		blocker string
	}{
		{
			name: "mismatched topology",
			opts: HAWavePlanOptions{InstallationID: "synthetic-cell-001", Current: HAWaveCounts{ControlPlane: 1, Etcd: 1, APIServer: 0},
				BackupBarrierPath: "backup.json", Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base, "backup")},
			blocker: "ha_wave_current_topology_invalid",
		},
		{
			name: "unsafe installation id",
			opts: HAWavePlanOptions{InstallationID: "https://private.invalid/cell", Current: haWaveCounts(1),
				BackupBarrierPath: "backup.json", Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base, "backup")},
			blocker: "ha_wave_installation_id_invalid",
		},
		{
			name: "missing backup validator",
			opts: HAWavePlanOptions{InstallationID: "synthetic-cell-001", Current: haWaveCounts(1),
				BackupBarrierPath: "backup.json", Now: fixedHAWaveClock(base)},
			blocker: "ha_wave_backup_validator_missing",
		},
		{
			name: "stale backup",
			opts: HAWavePlanOptions{InstallationID: "synthetic-cell-001", Current: haWaveCounts(1),
				BackupBarrierPath: "backup.json", Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base.Add(-maximumHAWaveEvidenceAge-time.Second), "backup")},
			blocker: "ha_wave_backup_invalid",
		},
		{
			name: "unsafe backup artifact basename",
			opts: HAWavePlanOptions{InstallationID: "synthetic-cell-001", Current: haWaveCounts(1),
				BackupBarrierPath: ".backup.json", Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base, "backup")},
			blocker: "ha_wave_backup_artifact_invalid",
		},
		{
			name: "invalid clock",
			opts: HAWavePlanOptions{InstallationID: "synthetic-cell-001", Current: haWaveCounts(1),
				BackupBarrierPath: "backup.json", Now: fixedHAWaveClock(time.Time{}), ValidateBackup: passingHAWaveBackup(base, "backup")},
			blocker: "ha_wave_clock_invalid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := BuildHAWavePlan(test.opts)
			if plan.Status != "blocked" || plan.MutationAllowed {
				t.Fatalf("invalid input did not fail closed: %+v", plan)
			}
			assertHAWaveBlocker(t, plan.Blockers, test.blocker)
		})
	}
}

func TestHAWaveVerificationRejectsUnsafeArtifactBasenames(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	planPath := writeHAWaveJSON(t, "first-plan.json", readyHAWavePlanForTest(base, HAWavePhaseOneToTwo))
	t.Run("plan", func(t *testing.T) {
		payload, err := os.ReadFile(planPath) // #nosec G304 -- planPath is created by this test beneath t.TempDir.
		if err != nil {
			t.Fatal(err)
		}
		hiddenPlan := filepath.Join(t.TempDir(), ".plan.json")
		if err := os.WriteFile(hiddenPlan, payload, 0o600); err != nil { // #nosec G703 -- hiddenPlan is a fixed filename beneath this test's private t.TempDir.
			t.Fatal(err)
		}
		opts := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
		opts.PlanPath = hiddenPlan
		assertHAWaveBlocker(t, VerifyHAWave(opts).Blockers, "ha_wave_plan_invalid")
	})
	t.Run("backup", func(t *testing.T) {
		opts := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
		opts.BackupBarrierPath = ".backup.json"
		assertHAWaveBlocker(t, VerifyHAWave(opts).Blockers, "ha_wave_backup_revalidation_invalid")
	})
	t.Run("inventory", func(t *testing.T) {
		opts := haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil)
		opts.InventoryPath = ".inventory.json"
		assertHAWaveBlocker(t, VerifyHAWave(opts).Blockers, "ha_wave_inventory_receipt_invalid")
	})
	t.Run("final receipt", func(t *testing.T) {
		finalPlanPath := writeHAWaveJSON(t, "second-plan.json", readyHAWavePlanForTest(base, HAWavePhaseTwoToThree))
		receipt := testOneServerLossReceipt()
		opts := haWaveVerifyOptions(t, finalPlanPath, "backup.json", base, 3, &receipt)
		opts.FinalOneServerLossPath = ".receipt.json"
		opts.LoadOneServerLossReceipt = func(string, time.Time) (*oneserverloss.Receipt, string, error) {
			return &receipt, testSHA256("receipt-artifact"), nil
		}
		assertHAWaveBlocker(t, VerifyHAWave(opts).Blockers, "ha_wave_final_one_server_loss_invalid")
	})
}

func TestReadHAWaveEvidenceRejectsUnknownDuplicateTrailingAndMutation(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	plan := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(1), BackupBarrierPath: "backup.json",
		Now: fixedHAWaveClock(base), ValidateBackup: passingHAWaveBackup(base.Add(-time.Minute), "backup"),
	})
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	unknown := append([]byte(nil), payload[:len(payload)-1]...)
	unknown = append(unknown, []byte(`,"unexpected":true}`)...)
	duplicate := []byte(`{"schemaVersion":"cloudring.kubeadm.ha-wave/v1","schemaVersion":"cloudring.kubeadm.ha-wave/v1"}`)
	trailing := append(append([]byte(nil), payload...), []byte(`{}`)...)
	mutatedPlan := plan
	mutatedPlan.MutationAllowed = true
	mutated, err := json.Marshal(mutatedPlan)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		payload []byte
	}{
		{"unknown", unknown},
		{"duplicate", duplicate},
		{"trailing", trailing},
		{"mutation allowed", mutated},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "plan.json")
			if err := os.WriteFile(path, test.payload, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := ReadHAWavePlan(path); err == nil {
				t.Fatal("unsafe HA-wave evidence was accepted")
			}
		})
	}
}

func TestReadHAWaveVerificationRequiresExactOrderedVerificationChecks(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	planPath := writeHAWaveJSON(t, "first-plan.json", readyHAWavePlanForTest(base, HAWavePhaseOneToTwo))
	ready := VerifyHAWave(haWaveVerifyOptions(t, planPath, "backup.json", base, 2, nil))
	if ready.Status != "ready" {
		t.Fatalf("test verification is not ready: %+v", ready)
	}

	missingBoundChecks := ready
	missingBoundChecks.Checks = []HAWaveCheck{
		{ID: "current_topology", Status: "pass", Detail: "fabricated topology"},
		{ID: "fresh_backup", Status: "pass", Artifact: "backup.json", SHA256: testSHA256("fabricated-backup"), Detail: "fabricated backup"},
	}
	if _, _, err := ReadHAWaveVerification(writeHAWaveJSON(t, "missing-checks.json", missingBoundChecks)); err == nil {
		t.Fatal("verification without plan, inventory, and target checks was accepted")
	}

	reordered := ready
	reordered.Checks = append([]HAWaveCheck(nil), ready.Checks...)
	reordered.Checks[0], reordered.Checks[1] = reordered.Checks[1], reordered.Checks[0]
	if _, _, err := ReadHAWaveVerification(writeHAWaveJSON(t, "reordered-checks.json", reordered)); err == nil {
		t.Fatal("verification with reordered checks was accepted")
	}
}

func TestReadHAWavePlanCannotUseVerificationShapedChecks(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	plan := readyHAWavePlanForTest(base, HAWavePhaseTwoToThree)
	plan.Checks = []HAWaveCheck{
		{ID: "plan", Status: "pass", Artifact: "plan.json", SHA256: testSHA256("plan"), Detail: "fabricated plan"},
		{ID: "fresh_backup", Status: "pass", Artifact: "backup.json", SHA256: testSHA256("backup"), Detail: "fabricated backup"},
		{ID: "inventory", Status: "pass", Artifact: "inventory.json", SHA256: testSHA256("inventory"), Detail: "fabricated inventory"},
		{ID: "target_topology", Status: "pass", Detail: "fabricated target"},
		{ID: "final_one_server_loss", Status: "pass", Artifact: "receipt.json", SHA256: testSHA256("receipt"), Detail: "fabricated receipt"},
	}
	if _, _, err := ReadHAWavePlan(writeHAWaveJSON(t, "verification-shaped-plan.json", plan)); err == nil {
		t.Fatal("plan with verification-shaped checks was accepted")
	}
}

func TestReadHAWaveEvidenceRejectsOversizedInputBeforeDecode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "oversized-plan.json")
	if err := os.WriteFile(path, make([]byte, strictjson.MaxDocumentBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadHAWavePlan(path); err == nil {
		t.Fatal("oversized HA-wave evidence was accepted")
	}
}

func TestHAWaveReportsDoNotCopyPrivatePaths(t *testing.T) {
	base := time.Date(2026, 7, 19, 12, 1, 0, 0, time.UTC)
	privatePrefix := filepath.Join("private", "installation", "records")
	plan := BuildHAWavePlan(HAWavePlanOptions{
		InstallationID: "synthetic-cell-001", Current: haWaveCounts(1),
		BackupBarrierPath: filepath.Join(privatePrefix, "backup.json"), Now: fixedHAWaveClock(base),
		ValidateBackup: passingHAWaveBackup(base.Add(-time.Minute), "backup"),
	})
	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), privatePrefix) || !strings.Contains(string(payload), `"artifact":"backup.json"`) {
		t.Fatalf("plan evidence was not source-safe: %s", payload)
	}
}

func readyHAWavePlanForTest(now time.Time, phase string) HAWavePlan {
	from := 1
	checks := []HAWaveCheck{
		{ID: "current_topology", Status: "pass", Detail: "phase=one-to-two members=1"},
		{ID: "fresh_backup", Status: "pass", Artifact: "backup.json", SHA256: testSHA256("backup"), Detail: "fresh off-cell backup and restore barrier is valid"},
	}
	if phase == HAWavePhaseTwoToThree {
		from = 2
		checks[0].Detail = "phase=two-to-three members=2"
		checks = append(checks, HAWaveCheck{ID: "previous_wave", Status: "pass", Artifact: "previous.json", SHA256: testSHA256("previous"), Detail: "one-to-two verification is ready and the inter-wave backup is newer"})
	}
	return HAWavePlan{
		SchemaVersion: HAWaveSchemaVersion, GeneratedAt: now.Format(time.RFC3339Nano), Status: "ready", Phase: phase,
		InstallationID: "synthetic-cell-001", From: haWaveCounts(from), Target: haWaveCounts(from + 1),
		BackupGeneratedAt: now.Add(-time.Minute).Format(time.RFC3339Nano), Checks: checks, Blockers: []string{},
		Steps: append([]string(nil), haWaveSteps...), Rollback: append([]string(nil), haWaveRollback...),
		MutationAllowed: false, NonClaim: haWavePlanNonClaim,
	}
}

func passingHAWaveBackup(generatedAt time.Time, digestSeed string) HAWaveBackupValidator {
	return func(string, string, time.Time) (time.Time, string, error) {
		return generatedAt, testSHA256(digestSeed), nil
	}
}

func haWaveVerifyOptions(t *testing.T, planPath, backupPath string, now time.Time, members int, receipt *oneserverloss.Receipt) HAWaveVerifyOptions {
	t.Helper()
	plan, _, err := ReadHAWavePlan(planPath)
	if err != nil {
		t.Fatalf("read test HA-wave plan: %v", err)
	}
	backupAt, err := parseHAWaveTime(plan.BackupGeneratedAt)
	if err != nil {
		t.Fatal(err)
	}
	backupCheck, ok := findHAWaveCheck(plan.Checks, "fresh_backup")
	if !ok {
		t.Fatal("test plan lacks fresh backup check")
	}
	inventory := testHAWaveInventoryReceipt(now, plan.InstallationID, members, receipt)
	return HAWaveVerifyOptions{
		PlanPath:          planPath,
		BackupBarrierPath: backupPath,
		ValidateBackup: func(string, string, time.Time) (time.Time, string, error) {
			return backupAt, backupCheck.SHA256, nil
		},
		InventoryPath: "inventory.json",
		LoadInventory: func(string, time.Time) (*HAWaveInventoryReceipt, string, error) {
			return inventory, testSHA256("inventory-artifact"), nil
		},
		Now: fixedHAWaveClock(now),
	}
}

func testHAWaveInventoryReceipt(now time.Time, installationID string, count int, receipt *oneserverloss.Receipt) *HAWaveInventoryReceipt {
	uids := make([]string, 0, count)
	if receipt != nil {
		uids = append(uids, receipt.TargetNodeUIDSHA256)
	}
	for index := len(uids); index < count; index++ {
		uids = append(uids, testSHA256("wave-member-"+strconv.Itoa(index)))
	}
	slices.Sort(uids)
	members := make([]HAWaveMemberIdentity, 0, count)
	for _, uid := range uids {
		members = append(members, HAWaveMemberIdentity{UIDSHA256: uid, Ready: true, ControlPlane: true, EtcdMember: true})
	}
	inventory := &HAWaveInventoryReceipt{
		SchemaVersion: HAWaveInventorySchemaVersion, CapturedAt: now.Format(time.RFC3339Nano), InstallationID: installationID,
		Distribution: "upstream", Bootstrap: "kubeadm", ServerVersion: "v1.35.6", Members: members,
		APIServerNodeUIDSHA256: slices.Clone(uids),
	}
	if receipt != nil {
		inventory.OneServerLossReceiptBinding = OneServerLossReceiptBinding{
			ReceiptSHA256: receipt.ReceiptSHA256, RunNonceSHA256: receipt.RunNonceSHA256,
			TargetNodeUIDSHA256: receipt.TargetNodeUIDSHA256, KubectlExecutableSHA256: receipt.KubectlExecutableSHA256,
			ProbeAdapterSHA256: receipt.ProbeAdapterSHA256,
		}
	}
	inventory.ReceiptSHA256 = HAWaveInventoryReceiptSHA256(*inventory)
	return inventory
}

func fixedHAWaveClock(now time.Time) func() time.Time {
	return func() time.Time { return now }
}

func writeHAWaveJSON(t *testing.T, name string, value any) string {
	t.Helper()
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func assertHAWaveBlocker(t *testing.T, blockers []string, wanted string) {
	t.Helper()
	for _, blocker := range blockers {
		if blocker == wanted {
			return
		}
	}
	t.Fatalf("missing blocker %q in %v", wanted, blockers)
}
