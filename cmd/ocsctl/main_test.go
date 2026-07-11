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

	"github.com/opencloudtech/CloudRING/cloudring_core/sdk/ocsv3"
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
		if !strings.Contains(stdout.String(), "ocs_connector_package_valid "+validPath) {
			t.Fatalf("stdout does not report the valid package: %q", stdout.String())
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
		if !strings.Contains(output, "ocs_connector_package_invalid "+invalidPath) {
			t.Fatalf("stdout does not report the invalid package: %q", output)
		}
		if !strings.Contains(output, "ocs_connector_package_valid "+validPath) {
			t.Fatalf("validation stopped before the valid operand: %q", output)
		}
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
		if !strings.Contains(stdout.String(), "ocs_conformance_passed "+validPath) {
			t.Fatalf("stdout does not report conformance success: %q", stdout.String())
		}

		var report ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
			t.Fatalf("decode success evidence: %v", err)
		}
		if !report.Passed || report.PackageName != "synthetic-service-module-package" {
			t.Fatalf("unexpected success report: %+v", report)
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
		if !strings.Contains(stdout.String(), "ocs_conformance_failed "+invalidPath) {
			t.Fatalf("stdout does not report conformance failure: %q", stdout.String())
		}

		var report ocsv3.ConformanceReport
		if err := json.Unmarshal(readTestFile(t, evidencePath), &report); err != nil {
			t.Fatalf("decode failure evidence: %v", err)
		}
		if report.Passed || report.PackageName != invalidPath || report.Summary == "" {
			t.Fatalf("unexpected failure report: %+v", report)
		}
	})
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
	path, err := filepath.Abs(filepath.Join("..", "..", "cloudring_core", "reference", "synthetic-service", "module-package.json"))
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

func testConformanceReport(name string) ocsv3.ConformanceReport {
	return ocsv3.ConformanceReport{
		APIVersion:  ocsv3.APIVersion,
		Kind:        "ConformanceReport",
		PackageName: name,
		Passed:      true,
		Summary:     "test conformance passed",
	}
}
