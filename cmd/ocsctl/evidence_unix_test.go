//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"
)

func privateEvidenceTestDirectory(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func makeEvidenceTestPredecessorPermissive(path string) error {
	return unix.Chmod(path, 0o644)
}

func TestUnixEvidenceRejectsWritableDestinationParent(t *testing.T) {
	directory := t.TempDir()
	if err := unix.Chmod(directory, 0o777); err != nil {
		t.Fatalf("make evidence parent writable: %v", err)
	}
	t.Cleanup(func() {
		_ = unix.Chmod(directory, 0o700)
	})
	evidencePath := filepath.Join(directory, "evidence.json")
	err := writePrivateFileSafely(evidencePath, []byte("must not publish\n"))
	if err == nil || !strings.Contains(err.Error(), "group/other write must be disabled") {
		t.Fatalf("write error = %v, want writable-parent rejection", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestUnixEvidenceRejectsWritableNonStickyAncestor(t *testing.T) {
	root := t.TempDir()
	mutableAncestor := filepath.Join(root, "mutable")
	controlledParent := filepath.Join(mutableAncestor, "controlled")
	if err := os.MkdirAll(controlledParent, 0o700); err != nil {
		t.Fatalf("create evidence namespace: %v", err)
	}
	if err := unix.Chmod(mutableAncestor, 0o777); err != nil {
		t.Fatalf("make ancestor writable: %v", err)
	}
	t.Cleanup(func() {
		_ = unix.Chmod(mutableAncestor, 0o700)
	})
	err := writePrivateFileSafely(filepath.Join(controlledParent, "evidence.json"), []byte("must not publish\n"))
	if err == nil || !strings.Contains(err.Error(), "namespace mutation") {
		t.Fatalf("write error = %v, want mutable-ancestor rejection", err)
	}
	assertNoEvidenceTemporaryFiles(t, controlledParent)
}

func TestUnixEvidenceNamespaceTamperBeforeReplaceFailsClosed(t *testing.T) {
	directory := t.TempDir()
	evidencePath := filepath.Join(directory, "namespace-tamper.json")
	hooks := evidenceWriteHooks{beforeReplaceValidation: func(_, _ string) error {
		return unix.Chmod(directory, 0o777)
	}}
	err := writePrivateFileSafelyWithHooks(evidencePath, []byte("must not publish\n"), replaceEvidenceFile, hooks)
	if err == nil || !strings.Contains(err.Error(), "group/other write must be disabled") {
		t.Fatalf("write error = %v, want namespace-tamper rejection", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
	if err := unix.Chmod(directory, 0o700); err != nil {
		t.Fatalf("restore evidence parent mode: %v", err)
	}
}

func TestUnixEvidenceTemporaryIdentityReplacementFailsClosed(t *testing.T) {
	directory := t.TempDir()
	evidencePath := filepath.Join(directory, "temp-identity.json")
	hooks := evidenceWriteHooks{beforeReplaceValidation: func(temporaryPath, _ string) error {
		if err := removeWithinParent(temporaryPath); err != nil {
			return err
		}
		return writeTestFileWithinParent(temporaryPath, []byte("replacement object\n"), 0o600)
	}}
	err := writePrivateFileSafelyWithHooks(evidencePath, []byte("original object\n"), replaceEvidenceFile, hooks)
	if err == nil || !strings.Contains(err.Error(), "file identity changed") {
		t.Fatalf("write error = %v, want temporary identity rejection", err)
	}
	if _, statErr := os.Lstat(evidencePath); !os.IsNotExist(statErr) {
		t.Fatalf("destination unexpectedly published: %v", statErr)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestUnixEvidencePublishedIdentityReplacementFailsClosed(t *testing.T) {
	directory := t.TempDir()
	evidencePath := filepath.Join(directory, "published-identity.json")
	var displacedPath string
	err := writePrivateFileSafelyWith(evidencePath, []byte("sensitive test object\n"), func(source, destination string) error {
		if err := replaceEvidenceFile(source, destination); err != nil {
			return err
		}
		displacedPath = filepath.Join(filepath.Dir(destination), "displaced-sensitive.json")
		if err := renameWithinParent(destination, displacedPath); err != nil {
			return err
		}
		return writeTestFileWithinParent(destination, []byte("different object\n"), 0o600)
	})
	if err == nil || !strings.Contains(err.Error(), "file identity changed") {
		t.Fatalf("write error = %v, want published identity rejection", err)
	}
	for _, path := range []string{evidencePath, displacedPath} {
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove identity-test artifact %s: %v", path, err)
		}
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}
