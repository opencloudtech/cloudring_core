//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func privateEvidenceTestDirectory(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestUnixEvidenceRejectsWritableDestinationParent(t *testing.T) {
	directory := t.TempDir()
	// #nosec G302 -- this negative test deliberately creates a world-writable
	// t.TempDir-owned parent so the production namespace guard must reject it.
	if err := os.Chmod(directory, 0o777); err != nil {
		t.Fatalf("make evidence parent writable: %v", err)
	}
	t.Cleanup(func() {
		// #nosec G302 -- 0700 is the required private directory restoration mode.
		_ = os.Chmod(directory, 0o700)
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
	// #nosec G302 -- this negative fixture must expose a non-sticky writable
	// ancestor to prove the production ancestor check rejects it.
	if err := os.Chmod(mutableAncestor, 0o777); err != nil {
		t.Fatalf("make ancestor writable: %v", err)
	}
	t.Cleanup(func() {
		// #nosec G302 -- 0700 restores the t.TempDir descendant after the fixture.
		_ = os.Chmod(mutableAncestor, 0o700)
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
		// #nosec G302 -- the test intentionally changes its own t.TempDir to 0777
		// between validation and replacement to exercise fail-closed revalidation.
		return os.Chmod(directory, 0o777)
	}}
	err := writePrivateFileSafelyWithHooks(evidencePath, []byte("must not publish\n"), replaceEvidenceFile, hooks)
	if err == nil || !strings.Contains(err.Error(), "group/other write must be disabled") {
		t.Fatalf("write error = %v, want namespace-tamper rejection", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
	// #nosec G302 -- 0700 is the required private directory restoration mode.
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatalf("restore evidence parent mode: %v", err)
	}
}

func TestUnixEvidenceTemporaryIdentityReplacementFailsClosed(t *testing.T) {
	directory := t.TempDir()
	evidencePath := filepath.Join(directory, "temp-identity.json")
	hooks := evidenceWriteHooks{beforeReplaceValidation: func(temporaryPath, _ string) error {
		// #nosec G703 -- temporaryPath is the exact O_EXCL path supplied by the
		// internal production hook; removing it is the identity-substitution trigger.
		if err := os.Remove(temporaryPath); err != nil {
			return err
		}
		// #nosec G703 -- the same internal O_EXCL path is deliberately recreated
		// inside this t.TempDir to prove dev/inode substitution is rejected.
		return os.WriteFile(temporaryPath, []byte("replacement object\n"), 0o600)
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
	displacedPath := filepath.Join(directory, "displaced-sensitive.json")
	err := writePrivateFileSafelyWith(evidencePath, []byte("sensitive test object\n"), func(source, destination string) error {
		if err := replaceEvidenceFile(source, destination); err != nil {
			return err
		}
		// #nosec G703 -- destination is the canonical path supplied by the internal
		// replacement callback and displacedPath is fixed under this test's t.TempDir.
		if err := os.Rename(destination, displacedPath); err != nil {
			return err
		}
		// #nosec G703 -- destination is deliberately recreated under the controlled
		// test directory to prove post-publication dev/inode verification fails.
		return os.WriteFile(destination, []byte("different object\n"), 0o600)
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
