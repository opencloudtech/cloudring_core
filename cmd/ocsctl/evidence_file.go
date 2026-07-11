package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const evidenceTemporaryPattern = ".ocsctl-evidence-*"

type evidenceWriteHooks struct {
	afterCreateVerified     func(*os.File, string) error
	beforeReplaceValidation func(string, string) error
	remove                  func(string) error
}

// writePrivateFileSafely writes and syncs an owner-restricted same-directory
// temporary file before publishing it with the platform replacement primitive.
// It verifies the controlled namespace and the same file identity before and
// after replacement. No uninterrupted concurrent-reader guarantee is claimed
// for MoveFileEx on Windows.
func writePrivateFileSafely(path string, data []byte) error {
	return writePrivateFileSafelyWith(path, data, replaceEvidenceFile)
}

func writePrivateFileSafelyWith(path string, data []byte, replace func(string, string) error) error {
	return writePrivateFileSafelyWithHooks(path, data, replace, evidenceWriteHooks{})
}

func writePrivateFileSafelyWithHooks(path string, data []byte, replace func(string, string) error, hooks evidenceWriteHooks) (resultErr error) {
	path, namespace, err := prepareEvidenceNamespace(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := closeEvidenceNamespace(&namespace); err != nil {
			resultErr = errors.Join(resultErr, fmt.Errorf("close evidence namespace guard: %w", err))
		}
	}()
	if err := validateEvidenceTarget(path); err != nil {
		return err
	}

	dir := filepath.Dir(path)
	temporary, temporaryPath, err := createPrivateEvidenceTemporary(dir)
	if err != nil {
		return err
	}
	temporaryIdentity, err := evidenceIdentityFromOpenFile(temporary)
	if err != nil {
		return joinEvidenceCleanupError(
			fmt.Errorf("capture private evidence temporary identity: %w", err),
			temporary,
			temporaryPath,
			removeEvidenceFile,
		)
	}
	remove := hooks.remove
	if remove == nil {
		remove = removeEvidenceFile
	}
	closed := false
	defer func() {
		if !closed {
			if err := temporary.Close(); err != nil {
				resultErr = errors.Join(resultErr, fmt.Errorf("close unpublished evidence temporary file: %w", err))
			}
		}
		// #nosec G703 -- temporaryPath is returned by the platform's exclusive
		// creator, not accepted from the CLI; cleanup removes that exact unpublished file.
		if err := remove(temporaryPath); err != nil && !os.IsNotExist(err) {
			resultErr = errors.Join(resultErr, fmt.Errorf("remove unpublished evidence temporary file: %w", err))
		}
	}()

	if hooks.afterCreateVerified != nil {
		if err := hooks.afterCreateVerified(temporary, temporaryPath); err != nil {
			return fmt.Errorf("after private evidence creation: %w", err)
		}
	}
	if _, err := temporary.Write(data); err != nil {
		return fmt.Errorf("write evidence temporary file: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync evidence temporary file: %w", err)
	}
	if closeEvidenceTemporaryBeforeReplace() {
		if err := temporary.Close(); err != nil {
			closed = true
			return fmt.Errorf("close evidence temporary file: %w", err)
		}
		closed = true
	}

	if hooks.beforeReplaceValidation != nil {
		if err := hooks.beforeReplaceValidation(temporaryPath, path); err != nil {
			return fmt.Errorf("before evidence replacement validation: %w", err)
		}
	}

	// Recheck the controlled namespace, destination, and same temporary object
	// immediately before replacement. The platform verifier does not follow the
	// final path component and compares the parent and file identities captured
	// before any evidence bytes were written.
	if err := verifyEvidenceNamespace(path, namespace); err != nil {
		return fmt.Errorf("verify evidence namespace before replacement: %w", err)
	}
	if err := validateEvidenceTarget(path); err != nil {
		return err
	}
	if err := verifyEvidenceTemporary(path, temporaryPath, temporaryIdentity); err != nil {
		return fmt.Errorf("verify evidence temporary file before replacement: %w", err)
	}
	if err := replace(temporaryPath, path); err != nil {
		return fmt.Errorf("replace evidence file: %w", err)
	}
	if err := verifyPrivateEvidenceDestination(path, temporaryIdentity); err != nil {
		verificationErr := fmt.Errorf("verify published evidence file: %w", err)
		if cleanupErr := removeEvidenceDestinationIfIdentityMatches(path, temporaryIdentity); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
			return errors.Join(verificationErr, fmt.Errorf("remove unverified published evidence file: %w", cleanupErr))
		}
		return verificationErr
	}
	if !closed {
		closeErr := temporary.Close()
		closed = true
		if closeErr != nil {
			publicationErr := fmt.Errorf("close published evidence file: %w", closeErr)
			if cleanupErr := removeEvidenceDestinationIfIdentityMatches(path, temporaryIdentity); cleanupErr != nil && !os.IsNotExist(cleanupErr) {
				return errors.Join(publicationErr, fmt.Errorf("remove evidence after close failure: %w", cleanupErr))
			}
			return publicationErr
		}
	}
	return nil
}

func validateEvidenceTarget(path string) error {
	// #nosec G304 G703 -- Lstat examines the explicit --evidence destination
	// without following a symlink; this is the security check for that CLI path.
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect evidence destination: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("refuse to replace evidence destination that is a symbolic link")
	}
	if !info.Mode().IsRegular() {
		return errors.New("refuse to replace evidence destination that is a non-regular file")
	}
	return validateEvidenceTargetOwner(path)
}

func joinEvidenceCleanupError(primary error, file *os.File, path string, remove func(string) error) error {
	if closeErr := file.Close(); closeErr != nil {
		primary = errors.Join(primary, fmt.Errorf("close evidence temporary file during cleanup: %w", closeErr))
	}
	if removeErr := remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		primary = errors.Join(primary, fmt.Errorf("remove evidence temporary file during cleanup: %w", removeErr))
	}
	return primary
}
