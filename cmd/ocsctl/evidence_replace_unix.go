//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type evidenceFileIdentity struct {
	device string
	inode  string
}

type evidenceNamespaceState struct {
	parentPath     string
	parentIdentity evidenceFileIdentity
}

// POSIX rename permits an open source file. Retaining the descriptor until
// after publication prevents an unlinked temporary inode from being recycled
// into a same-dev/inode replacement between validation and rename.
func closeEvidenceTemporaryBeforeReplace() bool {
	return false
}

func closeEvidenceNamespace(_ *evidenceNamespaceState) error {
	return nil
}

func prepareEvidenceNamespace(path string) (string, evidenceNamespaceState, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("resolve evidence path: %w", err)
	}
	resolvedParent, err := filepath.EvalSymlinks(filepath.Dir(absolutePath))
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("resolve evidence parent without symlinks: %w", err)
	}
	resolvedParent, err = filepath.Abs(resolvedParent)
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("resolve absolute evidence parent: %w", err)
	}
	if err := validateUnixAncestorChain(resolvedParent); err != nil {
		return "", evidenceNamespaceState{}, err
	}
	parentIdentity, err := verifyControlledUnixDirectory(resolvedParent)
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("verify controlled evidence parent: %w", err)
	}
	return filepath.Join(resolvedParent, filepath.Base(absolutePath)), evidenceNamespaceState{
		parentPath:     resolvedParent,
		parentIdentity: parentIdentity,
	}, nil
}

func verifyEvidenceNamespace(path string, expected evidenceNamespaceState) error {
	if filepath.Dir(path) != expected.parentPath {
		return errors.New("evidence destination parent changed after validation")
	}
	if err := validateUnixAncestorChain(expected.parentPath); err != nil {
		return err
	}
	actual, err := verifyControlledUnixDirectory(expected.parentPath)
	if err != nil {
		return err
	}
	if actual != expected.parentIdentity {
		return fmt.Errorf("evidence parent identity changed: got dev=%s inode=%s, require dev=%s inode=%s", actual.device, actual.inode, expected.parentIdentity.device, expected.parentIdentity.inode)
	}
	return nil
}

func validateUnixAncestorChain(path string) error {
	clean := filepath.Clean(path)
	if !filepath.IsAbs(clean) {
		return errors.New("evidence parent is not absolute")
	}
	current := string(os.PathSeparator)
	previousInfo, err := os.Lstat(current)
	if err != nil {
		return fmt.Errorf("inspect evidence namespace root: %w", err)
	}
	if err := verifyUnixNoExtendedACL(current); err != nil {
		return fmt.Errorf("verify evidence namespace root ACL: %w", err)
	}
	remaining := strings.TrimPrefix(clean, current)
	if remaining == "" {
		remaining = "."
	}
	for _, component := range strings.Split(remaining, string(os.PathSeparator)) {
		if component == "." || component == "" {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current) // #nosec G304 G703 -- canonical parent components are checked without following links.
		if err != nil {
			return fmt.Errorf("inspect evidence ancestor %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("evidence ancestor %q is a symbolic link", current)
		}
		if !info.IsDir() {
			return fmt.Errorf("evidence ancestor %q is not a directory", current)
		}
		if err := verifyUnixNoExtendedACL(current); err != nil {
			return fmt.Errorf("verify evidence ancestor %q ACL: %w", current, err)
		}
		if err := verifyUnixNamespaceEdge(filepath.Dir(current), previousInfo, current, info); err != nil {
			return err
		}
		previousInfo = info
	}
	return nil
}

func verifyUnixNamespaceEdge(parentPath string, parentInfo os.FileInfo, childPath string, childInfo os.FileInfo) error {
	if parentInfo.Mode().Perm()&0o022 == 0 {
		return nil
	}
	if parentInfo.Mode()&os.ModeSticky == 0 {
		return fmt.Errorf("evidence ancestor %q grants group/other namespace mutation", parentPath)
	}
	_, childOwner, err := unixIdentityAndOwner(childInfo)
	if err != nil {
		return fmt.Errorf("inspect sticky-directory child %q owner: %w", childPath, err)
	}
	// Sticky-directory rename/delete protection applies to entries owned by the
	// effective user (and to root-owned system entries). Other owners remain an
	// untrusted namespace edge and are rejected.
	if childOwner != unixEffectiveUID() && childOwner != "0" {
		return fmt.Errorf("sticky evidence ancestor %q contains child %q owned by untrusted UID %s", parentPath, childPath, childOwner)
	}
	return nil
}

func verifyControlledUnixDirectory(path string) (evidenceFileIdentity, error) {
	info, err := os.Lstat(path) // #nosec G304 G703 -- the canonical destination parent is checked without following links.
	if err != nil {
		return evidenceFileIdentity{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return evidenceFileIdentity{}, errors.New("evidence parent is not a non-symlink directory")
	}
	if err := verifyUnixNoExtendedACL(path); err != nil {
		return evidenceFileIdentity{}, err
	}
	identity, owner, err := unixIdentityAndOwner(info)
	if err != nil {
		return evidenceFileIdentity{}, err
	}
	if owner != unixEffectiveUID() {
		return evidenceFileIdentity{}, fmt.Errorf("evidence parent owner UID is %s, require effective UID %s", owner, unixEffectiveUID())
	}
	if permissions := info.Mode().Perm(); permissions&0o022 != 0 {
		return evidenceFileIdentity{}, fmt.Errorf("evidence parent permissions are %04o; group/other write must be disabled", permissions)
	}
	return identity, nil
}

func createPrivateEvidenceTemporary(dir string) (*os.File, string, error) {
	// #nosec G703 -- dir is a canonical, owner-controlled destination parent.
	// CreateTemp randomizes the filename, uses O_EXCL, and creates it with 0600.
	temporary, err := os.CreateTemp(dir, evidenceTemporaryPattern)
	if err != nil {
		return nil, "", fmt.Errorf("create private evidence temporary file: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		return nil, "", joinEvidenceCleanupError(
			fmt.Errorf("set evidence temporary file permissions: %w", err),
			temporary,
			temporary.Name(),
			removeEvidenceFile,
		)
	}
	if err := verifyPrivateUnixEvidenceFile(temporary.Name(), nil); err != nil {
		return nil, "", joinEvidenceCleanupError(
			fmt.Errorf("verify private evidence temporary file: %w", err),
			temporary,
			temporary.Name(),
			removeEvidenceFile,
		)
	}
	return temporary, temporary.Name(), nil
}

func evidenceIdentityFromOpenFile(file *os.File) (evidenceFileIdentity, error) {
	info, err := file.Stat()
	if err != nil {
		return evidenceFileIdentity{}, err
	}
	identity, owner, err := unixIdentityAndOwner(info)
	if err != nil {
		return evidenceFileIdentity{}, err
	}
	if !info.Mode().IsRegular() {
		return evidenceFileIdentity{}, errors.New("evidence temporary handle is not a regular file")
	}
	if owner != unixEffectiveUID() {
		return evidenceFileIdentity{}, fmt.Errorf("evidence temporary owner UID is %s, require effective UID %s", owner, unixEffectiveUID())
	}
	return identity, nil
}

func verifyEvidenceTemporary(_ string, temporaryPath string, expected evidenceFileIdentity) error {
	return verifyPrivateUnixEvidenceFile(temporaryPath, &expected)
}

func verifyPrivateEvidenceDestination(path string, expected evidenceFileIdentity) error {
	return verifyPrivateUnixEvidenceFile(path, &expected)
}

func verifyPrivateUnixEvidenceFile(path string, expected *evidenceFileIdentity) error {
	info, err := os.Lstat(path) // #nosec G304 G703 -- verifies the canonical evidence path without following symlinks.
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("evidence path is not a non-symlink regular file")
	}
	if permissions := info.Mode().Perm(); permissions != 0o600 {
		return fmt.Errorf("evidence permissions are %04o, require 0600", permissions)
	}
	identity, owner, err := unixIdentityAndOwner(info)
	if err != nil {
		return err
	}
	if owner != unixEffectiveUID() {
		return fmt.Errorf("evidence owner UID is %s, require effective UID %s", owner, unixEffectiveUID())
	}
	if expected != nil && identity != *expected {
		return fmt.Errorf("evidence file identity changed: got dev=%s inode=%s, require dev=%s inode=%s", identity.device, identity.inode, expected.device, expected.inode)
	}
	return nil
}

func validateEvidenceTargetOwner(path string) error {
	info, err := os.Lstat(path) // #nosec G304 G703 -- checks ownership without following the selected destination.
	if err != nil {
		return err
	}
	_, owner, err := unixIdentityAndOwner(info)
	if err != nil {
		return err
	}
	if owner != unixEffectiveUID() {
		return fmt.Errorf("refuse evidence destination owned by UID %s; require effective UID %s", owner, unixEffectiveUID())
	}
	return nil
}

func unixIdentityAndOwner(info os.FileInfo) (evidenceFileIdentity, string, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat == nil {
		return evidenceFileIdentity{}, "", errors.New("filesystem did not expose Unix file identity and owner")
	}
	return evidenceFileIdentity{device: fmt.Sprint(stat.Dev), inode: fmt.Sprint(stat.Ino)}, fmt.Sprint(stat.Uid), nil
}

func unixEffectiveUID() string {
	return fmt.Sprint(os.Geteuid())
}

func removeEvidenceDestinationIfIdentityMatches(path string, expected evidenceFileIdentity) error {
	info, err := os.Lstat(path) // #nosec G304 G703 -- validates identity before removing an unverified published file.
	if err != nil {
		return err
	}
	actual, _, err := unixIdentityAndOwner(info)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("refuse to remove unverified evidence with changed identity dev=%s inode=%s", actual.device, actual.inode)
	}
	return removeEvidenceFile(path)
}

func removeEvidenceFile(path string) error {
	return os.Remove(path) // #nosec G304 G703 -- removes only a bound temporary or identity-verified published path.
}

func replaceEvidenceFile(source string, destination string) error {
	// #nosec G703 -- source is an O_EXCL file whose dev/inode was rechecked in
	// an owner-controlled parent; destination was Lstat-checked in that parent.
	return os.Rename(source, destination)
}
