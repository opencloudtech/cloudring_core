//go:build windows

package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestWindowsEvidenceRejectsPermissiveParent(t *testing.T) {
	directory := t.TempDir()
	if err := applyWindowsTestDACL(directory, "D:P(A;OICI;FA;;;WD)(A;OICI;FA;;;AU)"); err != nil {
		t.Fatalf("make parent DACL permissive: %v", err)
	}
	evidencePath := filepath.Join(directory, "evidence.json")
	err := writePrivateFileSafely(evidencePath, []byte("must not be written\n"))
	if err == nil || !strings.Contains(err.Error(), "namespace-mutation rights") {
		t.Fatalf("write error = %v, want unsafe-parent rejection", err)
	}
	if _, statErr := os.Lstat(evidencePath); !os.IsNotExist(statErr) {
		t.Fatalf("evidence unexpectedly exists in permissive namespace: %v", statErr)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsEvidencePrivateParentSuccessAndExactSecurity(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "evidence.json")
	if err := os.WriteFile(evidencePath, []byte("old evidence\n"), 0o600); err != nil {
		t.Fatalf("create old evidence: %v", err)
	}
	if err := applyWindowsTestDACL(evidencePath, "D:P(A;;FA;;;WD)(A;;FA;;;AU)"); err != nil {
		t.Fatalf("make old evidence DACL permissive: %v", err)
	}

	want := []byte("new protected evidence\n")
	if err := writePrivateFileSafely(evidencePath, want); err != nil {
		t.Fatalf("write protected evidence: %v", err)
	}
	if got, err := os.ReadFile(evidencePath); err != nil {
		t.Fatalf("current user read protected evidence: %v", err)
	} else if !bytes.Equal(got, want) {
		t.Fatalf("protected evidence = %q, want %q", got, want)
	}
	writable, err := os.OpenFile(evidencePath, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("current user open protected evidence for write: %v", err)
	}
	if err := writable.Close(); err != nil {
		t.Fatalf("close protected evidence write handle: %v", err)
	}

	protected := inspectWindowsEvidencePathForTest(t, evidencePath)
	policy := windowsEvidencePolicyForTest(t)
	if err := verifyWindowsEvidenceSnapshot(protected, policy); err != nil {
		t.Fatalf("published owner/DACL policy: %v", err)
	}
	assertWindowsDACLDoesNotContainPrincipal(t, protected, wellKnownSIDString(t, windows.WinWorldSid))
	assertWindowsDACLDoesNotContainPrincipal(t, protected, wellKnownSIDString(t, windows.WinAuthenticatedUserSid))
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsEvidenceCreationHookRunsAfterOwnerAndDACLVerificationBeforeWrite(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "hook.json")
	hookCalled := false
	hooks := evidenceWriteHooks{afterCreateVerified: func(file *os.File, path string) error {
		hookCalled = true
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("stat create-time evidence: %w", err)
		}
		if info.Size() != 0 {
			return fmt.Errorf("create-time evidence already contains %d bytes", info.Size())
		}
		policy, err := newWindowsEvidenceACLPolicy()
		if err != nil {
			return err
		}
		if err := verifyWindowsEvidenceHandle(windows.Handle(file.Fd()), policy); err != nil {
			return fmt.Errorf("hook owner/DACL verification: %w", err)
		}
		if path == evidencePath {
			return errors.New("temporary path unexpectedly equals destination")
		}
		return nil
	}}
	want := []byte("written only after verified hook\n")
	if err := writePrivateFileSafelyWithHooks(evidencePath, want, replaceEvidenceFile, hooks); err != nil {
		t.Fatalf("write evidence with create hook: %v", err)
	}
	if !hookCalled {
		t.Fatal("create-time verification hook was not called")
	}
	if got, err := os.ReadFile(evidencePath); err != nil {
		t.Fatalf("read hook evidence: %v", err)
	} else if !bytes.Equal(got, want) {
		t.Fatalf("hook evidence = %q, want %q", got, want)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsEvidenceTemporaryHandleDeniesDeleteSharing(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "share-mode.json")
	hooks := evidenceWriteHooks{afterCreateVerified: func(_ *os.File, path string) error {
		err := removeEvidenceFile(path)
		if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
			return fmt.Errorf("delete while temporary handle open = %v, want sharing violation", err)
		}
		return nil
	}}
	if err := writePrivateFileSafelyWithHooks(evidencePath, []byte("share protected\n"), replaceEvidenceFile, hooks); err != nil {
		t.Fatalf("write share-protected evidence: %v", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsEvidencePostReplaceDACLVerificationFailsClosed(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "tampered.json")
	err := writePrivateFileSafelyWith(evidencePath, []byte("must not remain published\n"), func(source string, destination string) error {
		if err := replaceEvidenceFile(source, destination); err != nil {
			return err
		}
		return applyWindowsTestDACL(destination, "D:P(A;;FA;;;WD)(A;;FA;;;AU)")
	})
	if err == nil || !strings.Contains(err.Error(), "verify published evidence file") {
		t.Fatalf("write error = %v, want post-replacement DACL verification failure", err)
	}
	if _, statErr := os.Stat(evidencePath); !os.IsNotExist(statErr) {
		t.Fatalf("unverified evidence remains published: %v", statErr)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsEvidenceRejectsAttackerOwnerWithExactDACL(t *testing.T) {
	policy := windowsEvidencePolicyForTest(t)
	snapshot := exactWindowsEvidenceSnapshotForTest(policy)
	snapshot.owner = wellKnownSIDString(t, windows.WinWorldSid)
	err := verifyWindowsEvidenceSnapshot(snapshot, policy)
	if err == nil || !strings.Contains(err.Error(), "require effective user") {
		t.Fatalf("attacker-owned exact DACL error = %v, want exact-owner rejection", err)
	}
}

func TestWindowsExistingEvidenceOwnerTrustPolicy(t *testing.T) {
	policy := windowsEvidencePolicyForTest(t)
	for owner := range policy.principals {
		if err := validateExistingWindowsEvidenceOwner(owner, false, policy); err != nil {
			t.Fatalf("trusted existing owner %s rejected: %v", owner, err)
		}
	}
	attacker := wellKnownSIDString(t, windows.WinWorldSid)
	if err := validateExistingWindowsEvidenceOwner(attacker, false, policy); err == nil || !strings.Contains(err.Error(), "untrusted principal") {
		t.Fatalf("attacker owner error = %v, want untrusted-principal rejection", err)
	}
	if err := validateExistingWindowsEvidenceOwner(policy.owner, true, policy); err == nil || !strings.Contains(err.Error(), "defaulted") {
		t.Fatalf("defaulted owner error = %v, want rejection", err)
	}
}

func TestWindowsEvidenceDACLInspectorRejectsUnsupportedACEBeforeSIDParsing(t *testing.T) {
	descriptor, err := windows.SecurityDescriptorFromString("O:SYD:P(D;;FR;;;AN)(A;;FA;;;WD)")
	if err != nil {
		t.Fatalf("build unsupported ACE descriptor: %v", err)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatalf("read unsupported ACE DACL: %v", err)
	}
	entries, err := inspectWindowsACL(dacl)
	if err != nil {
		t.Fatalf("inspect supported basic deny ACE layout: %v", err)
	}
	if len(entries) != 2 || entries[0].aceType != windows.ACCESS_DENIED_ACE_TYPE {
		t.Fatalf("basic deny ACE parsing = %+v, want deny then allow", entries)
	}

	objectDescriptor, err := windows.SecurityDescriptorFromString("O:SYD:P(OA;;FR;00000000-0000-0000-0000-000000000000;;WD)")
	if err != nil {
		t.Skipf("object ACE fixture unsupported by this Windows version: %v", err)
	}
	objectDACL, _, err := objectDescriptor.DACL()
	if err != nil {
		t.Fatalf("read object ACE DACL: %v", err)
	}
	_, err = inspectWindowsACL(objectDACL)
	if err == nil || !strings.Contains(err.Error(), "unsupported type") {
		t.Fatalf("object ACE inspection error = %v, want unsupported-type rejection before SID parsing", err)
	}
	runtime.KeepAlive(descriptor)
	runtime.KeepAlive(objectDescriptor)
}

func TestWindowsEvidenceDACLInspectorRejectsMalformedACL(t *testing.T) {
	policy := windowsEvidencePolicyForTest(t)
	descriptor, err := windows.SecurityDescriptorFromString(policy.sddl)
	if err != nil {
		t.Fatalf("build valid descriptor: %v", err)
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		t.Fatalf("read valid DACL: %v", err)
	}
	var ace *windows.ACCESS_ALLOWED_ACE
	if err := windows.GetAce(dacl, 0, &ace); err != nil {
		t.Fatalf("read valid ACE: %v", err)
	}
	originalSize := ace.Header.AceSize
	ace.Header.AceSize = ^uint16(0)
	t.Cleanup(func() { ace.Header.AceSize = originalSize })
	_, err = inspectWindowsACL(dacl)
	if err == nil {
		t.Fatal("malformed ACL was accepted")
	}
	runtime.KeepAlive(descriptor)
}

func TestWindowsEvidenceSupportsExtendedLengthPath(t *testing.T) {
	directory := t.TempDir()
	for len(filepath.Join(directory, "evidence.json")) <= windows.MAX_PATH+80 {
		directory = filepath.Join(directory, strings.Repeat("long-path-component-", 2))
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatalf("create long evidence directory: %v", err)
	}
	applyWindowsPrivateEvidencePolicy(t, directory)
	evidencePath := filepath.Join(directory, "evidence.json")
	if len(evidencePath) <= windows.MAX_PATH {
		t.Fatalf("test path length = %d, want > %d", len(evidencePath), windows.MAX_PATH)
	}
	extended, err := windowsExtendedPath(evidencePath)
	if err != nil {
		t.Fatalf("resolve extended evidence path: %v", err)
	}
	if !strings.HasPrefix(extended, `\\?\`) {
		t.Fatalf("extended evidence path = %q, want \\\\?\\ prefix", extended)
	}

	want := []byte("long path evidence\n")
	if err := writePrivateFileSafely(evidencePath, want); err != nil {
		t.Fatalf("write long-path evidence: %v", err)
	}
	if got, err := os.ReadFile(extended); err != nil {
		t.Fatalf("read long-path evidence: %v", err)
	} else if !bytes.Equal(got, want) {
		t.Fatalf("long-path evidence = %q, want %q", got, want)
	}
	_ = inspectWindowsEvidencePathForTest(t, extended)
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsExtendedPathRejectsAlternateNamespacesAndDriveRelativePaths(t *testing.T) {
	for _, path := range []string{
		`\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1\evidence.json`,
		`\\?\Volume{00000000-0000-0000-0000-000000000000}\evidence.json`,
		`\\.\C:\evidence.json`,
		`C:relative\evidence.json`,
		`\\?\C:relative\evidence.json`,
		`\\?\UNC\server`,
		`\\?\UNC\server\share\..\evidence.json`,
		`\\?\C:\safe\evidence.json:stream`,
	} {
		t.Run(strings.NewReplacer(`\`, "_", `:`, "_").Replace(path), func(t *testing.T) {
			if got, err := windowsExtendedPath(path); err == nil {
				t.Fatalf("windowsExtendedPath(%q) = %q, want rejection", path, got)
			}
		})
	}
	for _, path := range []string{`\\?\C:\safe\evidence.json`, `\\?\UNC\server\share\evidence.json`} {
		if got, err := windowsExtendedPath(path); err != nil || got == "" {
			t.Fatalf("windowsExtendedPath(%q) = %q, %v", path, got, err)
		}
	}
}

func TestWindowsEvidenceNamespaceIdentityDetectsParentReplacement(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "identity.json")
	safePath, state, err := prepareEvidenceNamespace(evidencePath)
	if err != nil {
		t.Fatalf("prepare evidence namespace: %v", err)
	}
	backup := directory + "-original"
	if err := os.Rename(directory, backup); err == nil {
		if restoreErr := os.Rename(backup, directory); restoreErr != nil {
			t.Fatalf("ancestor guard allowed rename and original could not be restored: %v", restoreErr)
		}
		t.Fatal("ancestor guard allowed parent rename while publication was active")
	} else if !errors.Is(err, windows.ERROR_SHARING_VIOLATION) && !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		t.Fatalf("rename guarded parent error = %v, want sharing/access denial", err)
	}
	if err := closeEvidenceNamespace(&state); err != nil {
		t.Fatalf("close evidence namespace guard for identity test: %v", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
	if err := os.Rename(directory, backup); err != nil {
		t.Skipf("cannot exercise post-guard directory identity replacement: %v", err)
	}
	restored := false
	t.Cleanup(func() {
		if restored {
			return
		}
		_ = os.RemoveAll(directory)
		_ = os.Rename(backup, directory)
	})
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatalf("create replacement parent: %v", err)
	}
	applyWindowsPrivateEvidencePolicy(t, directory)
	err = verifyEvidenceNamespace(safePath, state)
	if err == nil || !strings.Contains(err.Error(), "parent identity changed") {
		t.Fatalf("namespace verification error = %v, want parent identity change", err)
	}
	if err := os.Remove(directory); err != nil {
		t.Fatalf("remove replacement parent: %v", err)
	}
	if err := os.Rename(backup, directory); err != nil {
		t.Fatalf("restore original parent: %v", err)
	}
	restored = true
}

func TestWindowsEvidenceNamespaceTamperBeforeReplaceFailsClosed(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "namespace-tamper.json")
	hooks := evidenceWriteHooks{beforeReplaceValidation: func(_, _ string) error {
		return applyWindowsTestDACL(directory, "D:P(A;OICI;FA;;;WD)")
	}}
	err := writePrivateFileSafelyWithHooks(evidencePath, []byte("must not publish\n"), replaceEvidenceFile, hooks)
	if err == nil || !strings.Contains(err.Error(), "namespace-mutation rights") {
		t.Fatalf("write error = %v, want namespace-tamper rejection", err)
	}
	assertNoEvidenceTemporaryFiles(t, directory)
	applyWindowsPrivateEvidencePolicy(t, directory)
}

func TestWindowsEvidenceTemporaryIdentityReplacementFailsClosed(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "temp-identity.json")
	hooks := evidenceWriteHooks{beforeReplaceValidation: func(temporaryPath, _ string) error {
		if err := removeEvidenceFile(temporaryPath); err != nil {
			return err
		}
		if err := os.WriteFile(temporaryPath, []byte("replacement object\n"), 0o600); err != nil {
			return err
		}
		return applyWindowsExactEvidencePolicy(temporaryPath)
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

func TestWindowsEvidencePublishedIdentityReplacementFailsClosed(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "published-identity.json")
	displacedPath := filepath.Join(directory, "displaced-sensitive.json")
	err := writePrivateFileSafelyWith(evidencePath, []byte("sensitive test object\n"), func(source, destination string) error {
		if err := replaceEvidenceFile(source, destination); err != nil {
			return err
		}
		if err := os.Rename(destination, displacedPath); err != nil {
			return err
		}
		if err := os.WriteFile(destination, []byte("different object\n"), 0o600); err != nil {
			return err
		}
		return applyWindowsExactEvidencePolicy(destination)
	})
	if err == nil || !strings.Contains(err.Error(), "file identity changed") {
		t.Fatalf("write error = %v, want published identity rejection", err)
	}
	for _, path := range []string{evidencePath, displacedPath} {
		if err := removeEvidenceFile(path); err != nil {
			t.Fatalf("remove identity-test artifact %s: %v", path, err)
		}
	}
	assertNoEvidenceTemporaryFiles(t, directory)
}

func TestWindowsConcurrentReadersObserveCompleteOldAndNewPayloads(t *testing.T) {
	directory := newWindowsPrivateEvidenceDirectory(t)
	evidencePath := filepath.Join(directory, "concurrent.json")
	oldPayload := bytes.Repeat([]byte("old-evidence-block\n"), 2048)
	newPayload := bytes.Repeat([]byte("new-evidence-block\n"), 2048)
	if err := writePrivateFileSafely(evidencePath, oldPayload); err != nil {
		t.Fatalf("write initial concurrent evidence: %v", err)
	}

	start := make(chan struct{})
	stop := make(chan struct{})
	ready := make(chan struct{}, 2)
	readerErrors := make(chan error, 2)
	var readers sync.WaitGroup
	var successfulReads atomic.Int64
	var oldReads atomic.Int64
	var newReads atomic.Int64
	var transientReadErrors atomic.Int64
	for reader := 0; reader < 2; reader++ {
		readers.Add(1)
		go func(readerID int) {
			defer readers.Done()
			ready <- struct{}{}
			<-start
			for {
				select {
				case <-stop:
					return
				default:
				}
				payload, err := os.ReadFile(evidencePath)
				if err != nil {
					if isTransientWindowsReplacementError(err) || os.IsNotExist(err) {
						transientReadErrors.Add(1)
						continue
					}
					readerErrors <- fmt.Errorf("reader %d: %w", readerID, err)
					return
				}
				successfulReads.Add(1)
				switch {
				case bytes.Equal(payload, oldPayload):
					oldReads.Add(1)
				case bytes.Equal(payload, newPayload):
					newReads.Add(1)
				default:
					readerErrors <- fmt.Errorf("reader %d observed partial/mixed payload of %d bytes", readerID, len(payload))
					return
				}
				runtime.Gosched()
			}
		}(reader)
	}
	<-ready
	<-ready
	close(start)
	var stopOnce sync.Once
	stopReaders := func() {
		stopOnce.Do(func() { close(stop) })
		readers.Wait()
	}
	defer stopReaders()
	waitForWindowsReadObservation(t, &oldReads, "old payload")

	deadline := time.Now().Add(15 * time.Second)
	replacementRetries := 0
	for replacement := 0; replacement < 24; {
		payload := oldPayload
		if replacement%2 == 0 {
			payload = newPayload
		}
		err := writePrivateFileSafely(evidencePath, payload)
		if err == nil {
			replacement++
			continue
		}
		if !isTransientWindowsReplacementError(err) {
			stopReaders()
			t.Fatalf("replace evidence with concurrent readers: %v", err)
		}
		replacementRetries++
		if time.Now().After(deadline) {
			stopReaders()
			t.Fatalf("concurrent readers prevented replacement for 15s; retries=%d", replacementRetries)
		}
		runtime.Gosched()
	}
	if err := writePrivateFileSafely(evidencePath, newPayload); err != nil {
		stopReaders()
		t.Fatalf("publish final new payload: %v", err)
	}
	waitForWindowsReadObservation(t, &newReads, "new payload")
	stopReaders()
	close(readerErrors)
	for err := range readerErrors {
		t.Fatal(err)
	}
	if successfulReads.Load() == 0 || oldReads.Load() == 0 || newReads.Load() == 0 {
		t.Fatalf("read observations successful=%d old=%d new=%d, require each > 0", successfulReads.Load(), oldReads.Load(), newReads.Load())
	}
	assertNoEvidenceTemporaryFiles(t, directory)

	// MoveFileEx documents replace/write-through flags, not uninterrupted reader
	// availability. Successful reads are proven complete old/new payloads; any
	// transient sharing/not-found errors and writer retries remain explicit.
	t.Logf("Windows replacement observations: successful=%d old=%d new=%d reader_transient=%d writer_retries=%d", successfulReads.Load(), oldReads.Load(), newReads.Load(), transientReadErrors.Load(), replacementRetries)
}

func waitForWindowsReadObservation(t *testing.T, counter *atomic.Int64, description string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for counter.Load() == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("no successful reader observed %s within 5s", description)
		}
		time.Sleep(time.Millisecond)
	}
}

func newWindowsPrivateEvidenceDirectory(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	applyWindowsPrivateEvidencePolicy(t, directory)
	return directory
}

func privateEvidenceTestDirectory(t *testing.T) string {
	t.Helper()
	return newWindowsPrivateEvidenceDirectory(t)
}

func applyWindowsPrivateEvidencePolicy(t *testing.T, path string) {
	t.Helper()
	policy := windowsEvidencePolicyForTest(t)
	if err := applyWindowsTestSecurity(path, windowsPrivateDirectorySDDLForTest(policy)); err != nil {
		t.Fatalf("apply private evidence namespace policy: %v", err)
	}
}

func windowsPrivateDirectorySDDLForTest(policy windowsEvidenceACLPolicy) string {
	principals := make([]string, 0, len(policy.principals))
	for principal := range policy.principals {
		principals = append(principals, principal)
	}
	sort.Strings(principals)
	var descriptor strings.Builder
	fmt.Fprintf(&descriptor, "O:%sD:P", policy.owner)
	for _, principal := range principals {
		fmt.Fprintf(&descriptor, "(A;OICI;FA;;;%s)", principal)
	}
	return descriptor.String()
}

func applyWindowsExactEvidencePolicy(path string) error {
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return err
	}
	return applyWindowsTestSecurity(path, policy.sddl)
}

func windowsEvidencePolicyForTest(t *testing.T) windowsEvidenceACLPolicy {
	t.Helper()
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		t.Fatalf("build expected evidence policy: %v", err)
	}
	return policy
}

func exactWindowsEvidenceSnapshotForTest(policy windowsEvidenceACLPolicy) windowsEvidenceDACL {
	entries := make([]windowsEvidenceACE, 0, len(policy.principals))
	for principal := range policy.principals {
		entries = append(entries, windowsEvidenceACE{sid: principal, mask: windowsEvidenceFullAccess, aceType: windows.ACCESS_ALLOWED_ACE_TYPE})
	}
	return windowsEvidenceDACL{
		control: windows.SE_DACL_PRESENT | windows.SE_DACL_PROTECTED,
		owner:   policy.owner,
		entries: entries,
	}
}

func inspectWindowsEvidencePathForTest(t *testing.T, path string) windowsEvidenceDACL {
	t.Helper()
	handle, _, err := openWindowsEvidenceObject(path, false, true, "test evidence")
	if err != nil {
		t.Fatalf("open evidence security for test: %v", err)
	}
	snapshot, inspectErr := inspectWindowsEvidenceDACL(handle)
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		inspectErr = errors.Join(inspectErr, closeErr)
	}
	if inspectErr != nil {
		t.Fatalf("inspect evidence security for test: %v", inspectErr)
	}
	return snapshot
}

func applyWindowsTestDACL(path string, sddl string) error {
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		return err
	}
	extendedPath, err := windowsExtendedPath(path)
	if err != nil {
		return err
	}
	err = windows.SetNamedSecurityInfo(
		extendedPath,
		windows.SE_FILE_OBJECT,
		windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		nil,
		nil,
		dacl,
		nil,
	)
	runtime.KeepAlive(descriptor)
	return err
}

func applyWindowsTestSecurity(path string, sddl string) error {
	descriptor, err := windows.SecurityDescriptorFromString(sddl)
	if err != nil {
		return err
	}
	owner, _, err := descriptor.Owner()
	if err != nil {
		return err
	}
	dacl, _, err := descriptor.DACL()
	if err != nil {
		return err
	}
	extendedPath, err := windowsExtendedPath(path)
	if err != nil {
		return err
	}
	err = windows.SetNamedSecurityInfo(
		extendedPath,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION,
		owner,
		nil,
		dacl,
		nil,
	)
	runtime.KeepAlive(descriptor)
	return err
}

func wellKnownSIDString(t *testing.T, sidType windows.WELL_KNOWN_SID_TYPE) string {
	t.Helper()
	sid, err := windows.CreateWellKnownSid(sidType)
	if err != nil {
		t.Fatalf("create well-known SID %d: %v", sidType, err)
	}
	return sid.String()
}

func assertWindowsDACLDoesNotContainPrincipal(t *testing.T, snapshot windowsEvidenceDACL, principal string) {
	t.Helper()
	for _, entry := range snapshot.entries {
		if entry.sid == principal {
			t.Fatalf("DACL unexpectedly grants principal %s mask %#x", principal, entry.mask)
		}
	}
}

func isTransientWindowsReplacementError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_ACCESS_DENIED) ||
		errors.Is(err, windows.ERROR_FILE_NOT_FOUND)
}
