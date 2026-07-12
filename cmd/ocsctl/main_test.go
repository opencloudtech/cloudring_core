package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/opencloudtech/CloudRING/sdk/ocsv3"
)

func TestRunValidateSuccessAndFailure(t *testing.T) {
	validPath := validConnectorFixture(t)
	invalidPath := writeTestFile(t, []byte("{}\n"))

	t.Run("success", func(t *testing.T) {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		if err := runWithIO([]string{"validate", validPath}, &stdout, &stderr); err != nil {
			t.Fatalf("run validate: %v", err)
		}
		if !strings.Contains(stdout.String(), "ocs_connector_package_valid input-sha256:") {
			t.Fatalf("stdout does not report the valid package: %q", stdout.String())
		}
		assertTextOmitsPath(t, stdout.String(), validPath)
		if strings.Contains(stdout.String(), "synthetic-service-module-package") {
			t.Fatalf("stdout leaks the declared package name: %q", stdout.String())
		}
		if stderr.Len() != 0 {
			t.Fatalf("unexpected stderr: %q", stderr.String())
		}
	})

	t.Run("failure still checks every operand", func(t *testing.T) {
		var stdout bytes.Buffer
		err := runWithIO([]string{"validate", invalidPath, validPath}, &stdout, &bytes.Buffer{})
		if err == nil || err.Error() != "one or more connector packages are invalid" {
			t.Fatalf("expected aggregate validation failure, got %v", err)
		}
		output := stdout.String()
		if !strings.Contains(output, "ocs_connector_package_invalid input-sha256:") {
			t.Fatalf("stdout does not report the invalid package: %q", output)
		}
		if !strings.Contains(output, "ocs_connector_package_valid input-sha256:") {
			t.Fatalf("validation stopped before the valid operand: %q", output)
		}
		assertTextOmitsPath(t, output, invalidPath)
		assertTextOmitsPath(t, output, validPath)
	})
}

func TestReadOperatorSelectedFileConfinesSymlinksToSelectedParent(t *testing.T) {
	sandbox := t.TempDir()
	selectedParent := filepath.Join(sandbox, "selected")
	outsideParent := filepath.Join(sandbox, "outside")
	if err := os.Mkdir(selectedParent, 0o700); err != nil {
		t.Fatalf("create selected parent: %v", err)
	}
	if err := os.Mkdir(outsideParent, 0o700); err != nil {
		t.Fatalf("create outside parent: %v", err)
	}
	insidePath := writeTestFileInDirectory(t, selectedParent, []byte("inside\n"))
	outsidePath := writeTestFileInDirectory(t, outsideParent, []byte("outside\n"))
	insideLink := filepath.Join(selectedParent, "inside-link.json")
	escapeLink := filepath.Join(selectedParent, "escape-link.json")
	if err := os.Symlink(filepath.Base(insidePath), insideLink); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join("..", filepath.Base(outsideParent), filepath.Base(outsidePath)), escapeLink); err != nil {
		t.Skipf("symlink traversal fixture is unavailable: %v", err)
	}

	inside, err := readOperatorSelectedFile(insideLink)
	if err != nil {
		t.Fatalf("read symlink confined to selected parent: %v", err)
	}
	if string(inside) != "inside\n" {
		t.Fatalf("confined symlink payload = %q, want inside payload", inside)
	}
	if _, err := readOperatorSelectedFile(escapeLink); err == nil {
		t.Fatal("read through parent-escaping symlink succeeded")
	}
	if got := string(readTestFile(t, outsidePath)); got != "outside\n" {
		t.Fatalf("outside file changed after rejected traversal: %q", got)
	}
}

func TestRunConformanceSuccessAndFailureEvidence(t *testing.T) {
	validPath := validConnectorFixture(t)

	t.Run("success", func(t *testing.T) {
		evidencePath := filepath.Join(privateEvidenceTestDirectory(t), "private", "passed.json")
		var stdout bytes.Buffer
		if err := runWithIO([]string{"conformance", validPath, "--evidence", evidencePath}, &stdout, &bytes.Buffer{}); err != nil {
			t.Fatalf("run conformance: %v", err)
		}
		if !strings.Contains(stdout.String(), "ocs_conformance_passed input-sha256:") {
			t.Fatalf("stdout does not report conformance success: %q", stdout.String())
		}
		assertTextOmitsPath(t, stdout.String(), validPath)
		if strings.Contains(stdout.String(), "synthetic-service-module-package") {
			t.Fatalf("stdout leaks the declared package name: %q", stdout.String())
		}

		var report ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
			t.Fatalf("decode success evidence: %v", err)
		}
		if !report.Passed || !strings.HasPrefix(report.PackageName, "input-sha256:") || report.PackageVersion != "v0.1.0" {
			t.Fatalf("unexpected success report: %+v", report)
		}
		if strings.Contains(string(readTestFile(t, evidencePath)), "synthetic-service-module-package") {
			t.Fatal("success evidence leaks the declared package name")
		}
	})

	t.Run("failure", func(t *testing.T) {
		invalidPath := writeTestFile(t, []byte("{\n"))
		evidencePath := filepath.Join(privateEvidenceTestDirectory(t), "failed.json")
		var stdout bytes.Buffer
		err := runWithIO([]string{"conformance", invalidPath, "--evidence=" + evidencePath}, &stdout, &bytes.Buffer{})
		if err == nil || err.Error() != "conformance failed for 1 connector package(s)" {
			t.Fatalf("expected conformance failure, got %v", err)
		}
		if !strings.Contains(stdout.String(), "ocs_conformance_failed input-sha256:") {
			t.Fatalf("stdout does not report conformance failure: %q", stdout.String())
		}
		assertTextOmitsPath(t, stdout.String(), invalidPath)

		var report ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
			t.Fatalf("decode failure evidence: %v", err)
		}
		if report.Passed || !strings.HasPrefix(report.PackageName, "input-sha256:") || report.Summary == "" {
			t.Fatalf("unexpected failure report: %+v", report)
		}
		assertTextOmitsPath(t, string(readTestFile(t, evidencePath)), invalidPath)
	})
}

func TestOperatorSelectedPathsNeverReachOutputOrEvidence(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	inputPath := writeTestFileInDirectory(t, directory, []byte("{\n"))
	evidencePath := filepath.Join(directory, "failure-evidence.json")

	var validateOutput bytes.Buffer
	validateErr := runWithIO([]string{"validate", inputPath}, &validateOutput, &bytes.Buffer{})
	if validateErr == nil {
		t.Fatal("validate unexpectedly passed malformed input")
	}
	assertTextOmitsPath(t, validateOutput.String(), inputPath)
	assertTextOmitsPath(t, validateErr.Error(), inputPath)

	var conformanceOutput bytes.Buffer
	conformanceErr := runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &conformanceOutput, &bytes.Buffer{})
	if conformanceErr == nil {
		t.Fatal("conformance unexpectedly passed malformed input")
	}
	assertTextOmitsPath(t, conformanceOutput.String(), inputPath)
	assertTextOmitsPath(t, conformanceErr.Error(), inputPath)
	evidence := string(readTestFile(t, evidencePath))
	assertTextOmitsPath(t, evidence, inputPath)
	if !strings.Contains(evidence, `"packageName": "input-sha256:`) {
		t.Fatalf("failure evidence lacks an opaque input identity: %s", evidence)
	}
}

func TestUnsafeDeclaredPackageNameCannotReintroduceSelectedPath(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	inputPath := filepath.Join(directory, "operator-selected-package.json")
	var fixture map[string]any
	if err := json.Unmarshal(readTestFile(t, validConnectorFixture(t)), &fixture); err != nil {
		t.Fatalf("decode valid fixture: %v", err)
	}
	metadata, ok := fixture["metadata"].(map[string]any)
	if !ok {
		t.Fatal("valid fixture metadata is not an object")
	}
	metadata["name"] = inputPath
	data, err := json.Marshal(fixture)
	if err != nil {
		t.Fatalf("encode unsafe-name fixture: %v", err)
	}
	if err := writeTestFileWithinParent(inputPath, data, 0o600); err != nil {
		t.Fatalf("write unsafe-name fixture: %v", err)
	}
	evidencePath := filepath.Join(directory, "unsafe-name-evidence.json")

	var stdout bytes.Buffer
	err = runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &stdout, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("conformance with an otherwise valid package failed: %v", err)
	}
	assertTextOmitsPath(t, stdout.String(), inputPath)
	evidence := string(readTestFile(t, evidencePath))
	assertTextOmitsPath(t, evidence, inputPath)
	if !strings.Contains(evidence, `"packageName": "input-sha256:`) {
		t.Fatalf("unsafe package name was not replaced with an opaque identity: %s", evidence)
	}
}

func TestDeclaredPackageNamesAlwaysUseStableOpaqueIdentity(t *testing.T) {
	declaredNames := []string{
		"synthetic-service-module-package",
		"ghp_",
		"github_pat_",
		"AKIA",
		"private-api.corp.example.invalid",
		"Q7mP9xK2" + "vN4rT8cL6wF3sH1jD5bZ0aY",
	}
	for _, declaredName := range declaredNames {
		t.Run(strings.ReplaceAll(declaredName, "/", "_"), func(t *testing.T) {
			directory := privateEvidenceTestDirectory(t)
			inputPath := filepath.Join(directory, "operator-selected-package.json")
			name := declaredName
			pkg, err := ocsv3.ParseConnectorPackage(readTestFile(t, validConnectorFixture(t)))
			if err != nil {
				t.Fatalf("parse valid fixture: %v", err)
			}
			pkg.Metadata.Name = name
			data, err := json.Marshal(pkg)
			if err != nil {
				t.Fatalf("encode declared-name fixture: %v", err)
			}
			if err := writeTestFileWithinParent(inputPath, data, 0o600); err != nil {
				t.Fatalf("write declared-name fixture: %v", err)
			}

			var validateOutput bytes.Buffer
			if err := runWithIO([]string{"validate", inputPath}, &validateOutput, &bytes.Buffer{}); err != nil {
				t.Fatalf("validate declared-name fixture: %v", err)
			}
			runConformance := func(evidenceName string) (ocsv3.ConformanceReport, string) {
				t.Helper()
				evidencePath := filepath.Join(directory, evidenceName)
				var stdout bytes.Buffer
				if err := runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &stdout, &bytes.Buffer{}); err != nil {
					t.Fatalf("run declared-name conformance: %v", err)
				}
				var report ocsv3.ConformanceReport
				if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
					t.Fatalf("decode declared-name evidence: %v", err)
				}
				return report, stdout.String() + string(readTestFile(t, evidencePath))
			}
			first, firstOutput := runConformance("first-evidence.json")
			second, secondOutput := runConformance("second-evidence.json")
			if !first.Passed || !strings.HasPrefix(first.PackageName, "input-sha256:") {
				t.Fatalf("unexpected opaque report: %+v", first)
			}
			if first.PackageName != second.PackageName {
				t.Fatalf("opaque identity changed for unchanged input: %q != %q", first.PackageName, second.PackageName)
			}
			if !strings.Contains(validateOutput.String(), "ocs_connector_package_valid "+first.PackageName) {
				t.Fatalf("validate output does not use opaque identity %q: %q", first.PackageName, validateOutput.String())
			}
			allOutput := validateOutput.String() + firstOutput + secondOutput
			for _, sensitive := range []string{name, strings.ToLower(name), inputPath, strings.ToLower(inputPath)} {
				if sensitive != "" && strings.Contains(allOutput, sensitive) {
					t.Fatalf("output/evidence leaks declared input value %q: %q", sensitive, allOutput)
				}
			}
		})
	}

	t.Run("selected path as declared name", func(t *testing.T) {
		directory := privateEvidenceTestDirectory(t)
		inputPath := filepath.Join(directory, "operator-selected-package.json")
		pkg, err := ocsv3.ParseConnectorPackage(readTestFile(t, validConnectorFixture(t)))
		if err != nil {
			t.Fatalf("parse valid fixture: %v", err)
		}
		pkg.Metadata.Name = inputPath
		data, err := json.Marshal(pkg)
		if err != nil {
			t.Fatalf("encode selected-path name fixture: %v", err)
		}
		if err := writeTestFileWithinParent(inputPath, data, 0o600); err != nil {
			t.Fatalf("write selected-path name fixture: %v", err)
		}
		evidencePath := filepath.Join(directory, "evidence.json")
		var stdout bytes.Buffer
		if err := runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &stdout, &bytes.Buffer{}); err != nil {
			t.Fatalf("run selected-path name conformance: %v", err)
		}
		assertTextOmitsPath(t, stdout.String(), inputPath)
		assertTextOmitsPath(t, string(readTestFile(t, evidencePath)), inputPath)
	})
}

func TestUntrustedVersionAndDynamicNamesAreSanitized(t *testing.T) {
	for _, reverseOverlapOrder := range []bool{false, true} {
		name := "short before long"
		if reverseOverlapOrder {
			name = "long before short"
		}
		t.Run(name, func(t *testing.T) {
			directory := privateEvidenceTestDirectory(t)
			inputPath := filepath.Join(directory, "operator-selected-package.json")
			pkg, err := ocsv3.ParseConnectorPackage(readTestFile(t, validConnectorFixture(t)))
			if err != nil {
				t.Fatalf("parse valid fixture: %v", err)
			}
			if len(pkg.Service.Spec.Lifecycle) < 5 || len(pkg.Service.Spec.States) < 4 {
				t.Fatal("valid fixture lacks lifecycle/state sanitization fixtures")
			}

			const actionShort = "OverlapActionSecret"
			const actionLong = actionShort + "-LongSuffixSecret"
			actionOverlap := []string{actionShort, actionLong}
			const stateShort = "OverlapStateSecret"
			const stateLong = stateShort + "-LongSuffixSecret"
			stateOverlap := []string{stateShort, stateLong}
			if reverseOverlapOrder {
				actionOverlap[0], actionOverlap[1] = actionOverlap[1], actionOverlap[0]
				stateOverlap[0], stateOverlap[1] = stateOverlap[1], stateOverlap[0]
			}
			const punctuatedAction = "Päth/秘密:ActionSecret"
			const punctuatedState = "Σ-state/秘密:StateSecret"
			const selectedSuffix = "SelectedPathSuffixSecret"
			actionNames := []string{actionOverlap[0], actionOverlap[1], punctuatedAction, inputPath, inputPath + "-" + selectedSuffix}
			stateNames := []string{stateOverlap[0], stateOverlap[1], punctuatedState, inputPath + "-" + selectedSuffix}
			pkg.Metadata.Version = inputPath + "-VersionSecret"
			for index, actionName := range actionNames {
				pkg.Service.Spec.Lifecycle[index].Name = actionName
				pkg.Service.Spec.Lifecycle[index].Idempotent = false
				pkg.Service.Spec.Lifecycle[index].IdempotencyKey = ""
			}
			for index, stateName := range stateNames {
				pkg.Service.Spec.States[index].Name = stateName
				pkg.Service.Spec.States[index].EvidenceRef = ""
				pkg.Service.Spec.States[index].Remediation = ""
			}
			data, err := json.Marshal(pkg)
			if err != nil {
				t.Fatalf("encode untrusted-value fixture: %v", err)
			}
			if err := writeTestFileWithinParent(inputPath, data, 0o600); err != nil {
				t.Fatalf("write untrusted-value fixture: %v", err)
			}
			evidencePath := filepath.Join(directory, "sanitized-evidence.json")

			var stdout bytes.Buffer
			err = runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &stdout, &bytes.Buffer{})
			if err == nil {
				t.Fatal("conformance unexpectedly passed invalid lifecycle/state entries")
			}
			evidence := string(readTestFile(t, evidencePath))
			sensitiveValues := append(append([]string{}, actionNames...), stateNames...)
			sensitiveValues = append(sensitiveValues, inputPath, selectedSuffix, "LongSuffixSecret", "VersionSecret")
			for label, text := range map[string]string{"stdout": stdout.String(), "error": err.Error(), "evidence": evidence} {
				assertTextOmitsPath(t, text, inputPath)
				for _, sensitive := range sensitiveValues {
					for _, variant := range []string{sensitive, strings.ToLower(sensitive)} {
						if variant != "" && strings.Contains(text, variant) {
							t.Fatalf("%s leaks untrusted dynamic value %q: %q", label, variant, text)
						}
					}
				}
			}
			var report ocsv3.ConformanceReport
			if err := json.Unmarshal([]byte(evidence), &report); err != nil {
				t.Fatalf("decode sanitized report: %v", err)
			}
			if report.PackageVersion != "redacted" {
				t.Fatalf("packageVersion = %q, want redacted", report.PackageVersion)
			}
			joinedProblems, err := json.Marshal(report.Problems)
			if err != nil {
				t.Fatalf("encode sanitized problems: %v", err)
			}
			if !bytes.Contains(joinedProblems, []byte("custom-action")) || !bytes.Contains(joinedProblems, []byte("custom-state")) {
				t.Fatalf("sanitized problem identities are missing: %s", joinedProblems)
			}
		})
	}
}

func TestRunConformanceUsesDistinctStableOpaqueIdentitiesForMultipleInputs(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	first := writeTestFileInDirectory(t, directory, []byte("{\n"))
	second := writeTestFileInDirectory(t, directory, []byte("{\n"))

	runOnce := func(evidenceName string) []ocsv3.ConformanceReport {
		t.Helper()
		evidencePath := filepath.Join(directory, evidenceName)
		err := runWithIO([]string{"conformance", first, second, "--evidence", evidencePath}, &bytes.Buffer{}, &bytes.Buffer{})
		if err == nil {
			t.Fatal("conformance unexpectedly passed malformed inputs")
		}
		var reports []ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &reports); err != nil {
			t.Fatalf("decode multi-input evidence: %v", err)
		}
		return reports
	}

	firstRun := runOnce("first.json")
	secondRun := runOnce("second.json")
	if len(firstRun) != 2 || len(secondRun) != 2 {
		t.Fatalf("report counts = %d and %d, want 2", len(firstRun), len(secondRun))
	}
	if firstRun[0].PackageName == firstRun[1].PackageName {
		t.Fatalf("different selected inputs received the same identity %q", firstRun[0].PackageName)
	}
	for index := range firstRun {
		if !strings.HasPrefix(firstRun[index].PackageName, "input-sha256:") {
			t.Fatalf("report %d identity = %q, want opaque digest", index, firstRun[index].PackageName)
		}
		if firstRun[index].PackageName != secondRun[index].PackageName {
			t.Fatalf("report %d identity changed between runs: %q != %q", index, firstRun[index].PackageName, secondRun[index].PackageName)
		}
	}
}

func TestRunConformanceRejectsEvidenceInputAliasesBeforePublication(t *testing.T) {
	validBytes := readTestFile(t, validConnectorFixture(t))

	t.Run("lexical canonical path", func(t *testing.T) {
		directory := privateEvidenceTestDirectory(t)
		inputPath := writeTestFileInDirectory(t, directory, validBytes)
		before := append([]byte(nil), readTestFile(t, inputPath)...)
		intermediate := filepath.Join(directory, "intermediate")
		if err := os.Mkdir(intermediate, 0o700); err != nil {
			t.Fatalf("create lexical alias component: %v", err)
		}
		evidenceAlias := directory + string(os.PathSeparator) + "intermediate" + string(os.PathSeparator) + ".." + string(os.PathSeparator) + filepath.Base(inputPath)

		err := runWithIO([]string{"conformance", inputPath, "--evidence", evidenceAlias}, &bytes.Buffer{}, &bytes.Buffer{})
		assertEvidenceAliasRejected(t, err)
		if got := readTestFile(t, inputPath); !bytes.Equal(got, before) {
			t.Fatal("input bytes changed after lexical evidence collision")
		}
		assertNoEvidenceTemporaryFiles(t, directory)
	})

	t.Run("symlink or reparse resolution", func(t *testing.T) {
		directory := privateEvidenceTestDirectory(t)
		inputPath := writeTestFileInDirectory(t, directory, validBytes)
		before := append([]byte(nil), readTestFile(t, inputPath)...)
		evidenceAlias := filepath.Join(directory, "evidence-link.json")
		if err := os.Symlink(filepath.Base(inputPath), evidenceAlias); err != nil {
			t.Skipf("symlink/reparse aliases are unavailable: %v", err)
		}

		err := runWithIO([]string{"conformance", inputPath, "--evidence", evidenceAlias}, &bytes.Buffer{}, &bytes.Buffer{})
		assertEvidenceAliasRejected(t, err)
		if got := readTestFile(t, inputPath); !bytes.Equal(got, before) {
			t.Fatal("input bytes changed after symlink evidence collision")
		}
		info, statErr := os.Lstat(evidenceAlias)
		if statErr != nil || info.Mode()&os.ModeSymlink == 0 {
			t.Fatalf("evidence alias was changed after rejection: info=%v err=%v", info, statErr)
		}
		assertNoEvidenceTemporaryFiles(t, directory)
	})

	t.Run("hardlink file identity", func(t *testing.T) {
		directory := privateEvidenceTestDirectory(t)
		inputPath := writeTestFileInDirectory(t, directory, validBytes)
		before := append([]byte(nil), readTestFile(t, inputPath)...)
		evidenceAlias := filepath.Join(directory, "evidence-hardlink.json")
		if err := os.Link(inputPath, evidenceAlias); err != nil {
			t.Skipf("hardlink aliases are unavailable: %v", err)
		}

		err := runWithIO([]string{"conformance", inputPath, "--evidence", evidenceAlias}, &bytes.Buffer{}, &bytes.Buffer{})
		assertEvidenceAliasRejected(t, err)
		if got := readTestFile(t, inputPath); !bytes.Equal(got, before) {
			t.Fatal("input bytes changed after hardlink evidence collision")
		}
		if got := readTestFile(t, evidenceAlias); !bytes.Equal(got, before) {
			t.Fatal("hardlink bytes changed after evidence collision")
		}
		assertNoEvidenceTemporaryFiles(t, directory)
	})

	t.Run("separate destination succeeds", func(t *testing.T) {
		directory := privateEvidenceTestDirectory(t)
		inputPath := writeTestFileInDirectory(t, directory, validBytes)
		evidencePath := filepath.Join(directory, "evidence.json")
		if err := runWithIO([]string{"conformance", inputPath, "--evidence", evidencePath}, &bytes.Buffer{}, &bytes.Buffer{}); err != nil {
			t.Fatalf("separate evidence destination failed: %v", err)
		}
		if got := readTestFile(t, inputPath); !bytes.Equal(got, validBytes) {
			t.Fatal("successful evidence publication changed input bytes")
		}
		var report ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
			t.Fatalf("decode separate evidence: %v", err)
		}
		if !report.Passed || !strings.HasPrefix(report.PackageName, "input-sha256:") {
			t.Fatalf("unexpected successful report: %+v", report)
		}
	})
}

func TestEvidenceInputGuardRevalidatesLateHardlinkCollision(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	inputPath := writeTestFileInDirectory(t, directory, readTestFile(t, validConnectorFixture(t)))
	evidencePath := filepath.Join(directory, "late-hardlink.json")
	input := openOperatorSelectedInput(inputPath)
	defer func() {
		if err := input.close(); err != nil {
			t.Fatalf("close selected input: %v", err)
		}
	}()
	guard, err := newEvidenceInputGuard(evidencePath, []*operatorSelectedInput{input})
	if err != nil {
		t.Fatalf("create evidence guard before collision: %v", err)
	}
	before := append([]byte(nil), input.data...)
	if err := os.Link(inputPath, evidencePath); err != nil {
		t.Skipf("hardlink aliases are unavailable: %v", err)
	}

	err = guard.verify()
	assertEvidenceAliasRejected(t, err)
	if got := readTestFile(t, inputPath); !bytes.Equal(got, before) {
		t.Fatal("late collision verification changed input bytes")
	}
}

func TestParseConformanceArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantPaths    []string
		wantEvidence string
		wantError    string
	}{
		{name: "package only", args: []string{"package.json"}, wantPaths: []string{"package.json"}},
		{name: "separate evidence", args: []string{"a.json", "--evidence", "report.json", "b.json"}, wantPaths: []string{"a.json", "b.json"}, wantEvidence: "report.json"},
		{name: "equals evidence", args: []string{"--evidence=report.json", "package.json"}, wantPaths: []string{"package.json"}, wantEvidence: "report.json"},
		{name: "dash-prefixed operand after sentinel", args: []string{"--", "-package.json"}, wantPaths: []string{"-package.json"}},
		{name: "flags stop at sentinel", args: []string{"--evidence", "report.json", "--", "-package.json", "--literal-file"}, wantPaths: []string{"-package.json", "--literal-file"}, wantEvidence: "report.json"},
		{name: "no package", wantError: "usage: ocsctl conformance"},
		{name: "missing evidence value", args: []string{"package.json", "--evidence"}, wantError: "usage: --evidence requires a path"},
		{name: "empty separate evidence", args: []string{"package.json", "--evidence", ""}, wantError: "usage: --evidence requires a path"},
		{name: "empty equals evidence", args: []string{"package.json", "--evidence="}, wantError: "usage: --evidence requires a path"},
		{name: "unknown flag", args: []string{"package.json", "--unknown"}, wantError: "unknown conformance flag"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			paths, evidence, err := parseConformanceArgs(test.args)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("expected error containing %q, got %v", test.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parse conformance args: %v", err)
			}
			if !reflect.DeepEqual(paths, test.wantPaths) {
				t.Fatalf("paths = %#v, want %#v", paths, test.wantPaths)
			}
			if evidence != test.wantEvidence {
				t.Fatalf("evidence = %q, want %q", evidence, test.wantEvidence)
			}
		})
	}
}

func TestParseConformanceArgsDoesNotEchoUnknownFlag(t *testing.T) {
	const sensitiveFlag = "--operator-secret=/private/selected/package.json"
	_, _, err := parseConformanceArgs([]string{"package.json", sensitiveFlag})
	if err == nil || err.Error() != "unknown conformance flag" {
		t.Fatalf("unknown flag error = %v, want generic refusal", err)
	}
	if strings.Contains(err.Error(), sensitiveFlag) || strings.Contains(err.Error(), "/private/selected/package.json") {
		t.Fatalf("unknown flag error leaks argv content: %q", err)
	}
}

func TestRunConformanceAcceptsDashPrefixedOperandAfterSentinel(t *testing.T) {
	validBytes := readTestFile(t, validConnectorFixture(t))
	directory := privateEvidenceTestDirectory(t)
	inputPath := filepath.Join(directory, "-package.json")
	if err := writeTestFileWithinParent(inputPath, validBytes, 0o600); err != nil {
		t.Fatalf("write dash-prefixed package: %v", err)
	}
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("read working directory: %v", err)
	}
	if err := os.Chdir(directory); err != nil {
		t.Fatalf("enter dash-prefixed package directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	var stdout bytes.Buffer
	if err := runWithIO([]string{"conformance", "--", "-package.json"}, &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("run conformance with dash-prefixed operand: %v", err)
	}
	if !strings.Contains(stdout.String(), "ocs_conformance_passed input-sha256:") {
		t.Fatalf("unexpected conformance output: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "-package.json") || strings.Contains(stdout.String(), inputPath) {
		t.Fatalf("dash-prefixed operand leaked to output: %q", stdout.String())
	}
}

func TestRunWithIORejectsInvalidCommands(t *testing.T) {
	for _, args := range [][]string{nil, []string{"unknown"}, []string{"validate"}} {
		if err := runWithIO(args, &bytes.Buffer{}, &bytes.Buffer{}); err == nil || !strings.HasPrefix(err.Error(), "usage:") {
			t.Fatalf("runWithIO(%q) error = %v, want usage error", args, err)
		}
	}
}

func TestWriteConformanceEvidenceUsesPrivatePermissions(t *testing.T) {
	root := privateEvidenceTestDirectory(t)
	directory := filepath.Join(root, "private", "nested")
	evidencePath := filepath.Join(directory, "report.json")
	if err := writeConformanceEvidence(evidencePath, []ocsv3.ConformanceReport{testConformanceReport("private")}); err != nil {
		t.Fatalf("write conformance evidence: %v", err)
	}

	if runtime.GOOS != "windows" {
		for _, path := range []string{filepath.Join(root, "private"), directory} {
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("stat evidence directory %s: %v", path, err)
			}
			if got := info.Mode().Perm(); got != 0o700 {
				t.Fatalf("evidence directory %s permissions = %04o, want 0700", path, got)
			}
		}
		info, err := os.Stat(evidencePath)
		if err != nil {
			t.Fatalf("stat evidence file: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("evidence file permissions = %04o, want 0600", got)
		}
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWriteConformanceEvidenceSafelyReplacesExistingFile(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	evidencePath := writeTestFileInDirectory(t, directory, []byte("previous evidence\n"))
	if err := makeEvidenceTestPredecessorPermissive(evidencePath); err != nil {
		t.Fatalf("make old evidence permissive: %v", err)
	}

	if err := writeConformanceEvidence(evidencePath, []ocsv3.ConformanceReport{testConformanceReport("replacement")}); err != nil {
		t.Fatalf("replace conformance evidence: %v", err)
	}
	if bytes.Contains(readTestFile(t, evidencePath), []byte("previous evidence")) {
		t.Fatal("evidence destination still contains the old payload")
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(evidencePath)
		if err != nil {
			t.Fatalf("stat replaced evidence: %v", err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Fatalf("replaced evidence permissions = %04o, want 0600", got)
		}
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWritePrivateFileSafelyPreservesDestinationOnReplaceFailure(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	evidencePath := writeTestFileInDirectory(t, directory, []byte("previous evidence\n"))
	replaceFailure := errors.New("injected replacement failure")

	err := writePrivateFileSafelyWith(evidencePath, []byte("new evidence\n"), func(source string, destination string) error {
		if filepath.Dir(source) != filepath.Dir(destination) {
			t.Fatalf("temporary and destination directories differ: source=%q destination=%q", source, destination)
		}
		if filepath.Base(destination) != filepath.Base(evidencePath) {
			t.Fatalf("replacement destination = %q, want base %q", destination, filepath.Base(evidencePath))
		}
		return replaceFailure
	})
	if !errors.Is(err, replaceFailure) {
		t.Fatalf("write error = %v, want injected replacement failure", err)
	}
	if got := string(readTestFile(t, evidencePath)); got != "previous evidence\n" {
		t.Fatalf("destination changed after failed replacement: %q", got)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestRenameWithinParentRejectsCrossParentTraversal(t *testing.T) {
	left := t.TempDir()
	right := t.TempDir()
	source := writeTestFileInDirectory(t, left, []byte("source\n"))
	destination := filepath.Join(right, "destination.json")

	err := renameWithinParent(source, destination)
	if err == nil || !strings.Contains(err.Error(), "must share a parent directory") {
		t.Fatalf("cross-parent rename error = %v, want confinement refusal", err)
	}
	if got := string(readTestFile(t, source)); got != "source\n" {
		t.Fatalf("source changed after rejected cross-parent rename: %q", got)
	}
	if _, err := lstatWithinParent(destination); !os.IsNotExist(err) {
		t.Fatalf("cross-parent destination unexpectedly exists: %v", err)
	}
}

func TestWritePrivateFileSafelyReturnsCleanupFailure(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	evidencePath := filepath.Join(directory, "cleanup.json")
	replaceFailure := errors.New("injected replacement failure")
	cleanupFailure := errors.New("injected cleanup failure")
	hooks := evidenceWriteHooks{remove: func(path string) error {
		if strings.HasPrefix(filepath.Base(path), strings.TrimSuffix(evidenceTemporaryPattern, "*")) {
			return cleanupFailure
		}
		return removeEvidenceFile(path)
	}}
	err := writePrivateFileSafelyWithHooks(evidencePath, []byte("cleanup evidence\n"), func(_, _ string) error {
		return replaceFailure
	}, hooks)
	if !errors.Is(err, replaceFailure) || !errors.Is(err, cleanupFailure) {
		t.Fatalf("write error = %v, want replacement and cleanup failures", err)
	}
	matches, globErr := filepath.Glob(filepath.Join(directory, evidenceTemporaryPattern))
	if globErr != nil {
		t.Fatalf("glob injected cleanup temporary: %v", globErr)
	}
	if len(matches) != 1 {
		t.Fatalf("injected cleanup left %d temporary files, want 1", len(matches))
	}
	if err := removeEvidenceFile(matches[0]); err != nil {
		t.Fatalf("remove injected cleanup temporary: %v", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWriteConformanceEvidenceRejectsSymlinkDestination(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	targetPath := writeTestFileInDirectory(t, directory, []byte("protected target\n"))
	linkPath := filepath.Join(directory, "evidence-link.json")
	if err := os.Symlink(filepath.Base(targetPath), linkPath); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	err := writeConformanceEvidence(linkPath, []ocsv3.ConformanceReport{testConformanceReport("symlink")})
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("write error = %v, want symbolic-link refusal", err)
	}
	if got := string(readTestFile(t, targetPath)); got != "protected target\n" {
		t.Fatalf("symlink target changed: %q", got)
	}
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("lstat evidence link: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("evidence link was replaced despite fail-closed policy")
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWriteConformanceEvidenceRejectsSymlinkedParentWithoutSideEffects(t *testing.T) {
	sandbox := privateEvidenceTestDirectory(t)
	selectedParent := filepath.Join(sandbox, "selected")
	externalParent := filepath.Join(sandbox, "external")
	if err := os.Mkdir(selectedParent, 0o700); err != nil {
		t.Fatalf("create selected evidence parent: %v", err)
	}
	if err := os.Mkdir(externalParent, 0o700); err != nil {
		t.Fatalf("create external evidence parent: %v", err)
	}
	redirect := filepath.Join(selectedParent, "redirect")
	if err := os.Symlink(externalParent, redirect); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}
	externalNested := filepath.Join(externalParent, "must-not-exist")
	evidencePath := filepath.Join(redirect, filepath.Base(externalNested), "report.json")

	err := writeConformanceEvidence(evidencePath, []ocsv3.ConformanceReport{testConformanceReport("parent-symlink")})
	if err == nil || !strings.Contains(err.Error(), "symbolic link") {
		t.Fatalf("write error = %v, want symlinked-parent refusal", err)
	}
	if _, statErr := lstatWithinParent(externalNested); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("symlink target received an unexpected directory: %v", statErr)
	}
	if _, statErr := lstatWithinParent(filepath.Join(externalNested, "report.json")); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("symlink target received an unexpected evidence file: %v", statErr)
	}
	assertNoEvidenceTemporaryFiles(t, externalParent)
}

func TestWriteConformanceEvidenceRejectsNonRegularDestination(t *testing.T) {
	directory := privateEvidenceTestDirectory(t)
	evidencePath := filepath.Join(directory, "report.json")
	if err := os.Mkdir(evidencePath, 0o700); err != nil {
		t.Fatalf("create directory destination: %v", err)
	}
	err := writeConformanceEvidence(evidencePath, []ocsv3.ConformanceReport{testConformanceReport("directory")})
	if err == nil || !strings.Contains(err.Error(), "non-regular file") {
		t.Fatalf("write error = %v, want non-regular-file refusal", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func validConnectorFixture(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "reference", "synthetic-service", "module-package.json"))
	if err != nil {
		t.Fatalf("resolve valid connector fixture: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat valid connector fixture: %v", err)
	}
	return path
}

func writeTestFile(t *testing.T, data []byte) string {
	t.Helper()
	return writeTestFileInDirectory(t, t.TempDir(), data)
}

func writeTestFileInDirectory(t *testing.T, directory string, data []byte) string {
	t.Helper()
	file, err := os.CreateTemp(directory, "ocsctl-test-*.json")
	if err != nil {
		t.Fatalf("create test file: %v", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		t.Fatalf("write test file: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close test file: %v", err)
	}
	return file.Name()
}

func writeTestFileWithinParent(path string, data []byte, mode os.FileMode) (resultErr error) {
	rooted, err := openParentRootPath(path)
	if err != nil {
		return err
	}
	defer rooted.close(&resultErr)
	return rooted.root.WriteFile(rooted.name, data, mode)
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := readOperatorSelectedFile(path)
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}
	return data
}

func assertNoEvidenceTemporaryFiles(t *testing.T, directory string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(directory, evidenceTemporaryPattern))
	if err != nil {
		t.Fatalf("glob evidence temporary files: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("evidence temporary files were not cleaned up: %v", matches)
	}
}

func assertEvidenceAliasRejected(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), "evidence destination aliases a connector package input") {
		t.Fatalf("evidence alias error = %v, want fail-closed collision refusal", err)
	}
}

func assertTextOmitsPath(t *testing.T, text string, path string) {
	t.Helper()
	variants := []string{path, filepath.Clean(path), filepath.ToSlash(path)}
	for _, variant := range variants {
		if variant != "" && strings.Contains(text, variant) {
			t.Fatalf("text leaks operator-selected path %q: %q", variant, text)
		}
	}
}

func testConformanceReport(name string) ocsv3.ConformanceReport {
	return ocsv3.ConformanceReport{
		APIVersion:  ocsv3.APIVersion,
		Kind:        "ConformanceReport",
		PackageName: name,
		Passed:      true,
		Summary:     "test conformance passed",
	}
}
