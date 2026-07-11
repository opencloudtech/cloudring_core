//go:build windows

package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	windowsEvidenceRandomBytes       = 16
	windowsEvidenceCreateAttempts    = 16
	windowsEvidenceSpecificRights    = 0x1ff
	windowsEvidenceFullAccess        = windows.ACCESS_MASK(windows.STANDARD_RIGHTS_REQUIRED | windows.SYNCHRONIZE | windowsEvidenceSpecificRights)
	windowsACLSizeInformationClass   = 2
	windowsACLHeaderBytes            = 8
	windowsMinimumSIDBytes           = 8
	windowsDirectoryDeleteChild      = 0x00000040
	windowsEvidenceNamespaceMutation = windows.ACCESS_MASK(
		windows.FILE_WRITE_DATA |
			windows.FILE_APPEND_DATA |
			windowsDirectoryDeleteChild |
			windows.FILE_WRITE_EA |
			windows.FILE_WRITE_ATTRIBUTES |
			windows.DELETE |
			windows.WRITE_DAC |
			windows.WRITE_OWNER |
			windows.GENERIC_WRITE |
			windows.GENERIC_ALL |
			windows.MAXIMUM_ALLOWED,
	)
)

var (
	windowsAdvapi32              = windows.NewLazySystemDLL("advapi32.dll")
	windowsProcIsValidACL        = windowsAdvapi32.NewProc("IsValidAcl")
	windowsProcGetACLInformation = windowsAdvapi32.NewProc("GetAclInformation")
)

type evidenceFileIdentity struct {
	volumeSerial  uint32
	fileIndexHigh uint32
	fileIndexLow  uint32
}

type evidenceNamespaceState struct {
	parentPath         string
	parentIdentity     evidenceFileIdentity
	guardHandles       []windows.Handle
	namespaceGuard     *os.File
	namespaceGuardPath string
}

// The Windows publication primitive cannot replace the temporary path while
// the os.File handle remains open with Go's default sharing mode. Windows
// therefore closes first and relies on the protected parent DACL, the open
// namespace sentinel, retained ancestor handles, and identity verification.
func closeEvidenceTemporaryBeforeReplace() bool {
	return true
}

func canonicalEvidenceDestinationForParentCreation(path string) (string, error) {
	return windowsExtendedPath(path)
}

type windowsEvidenceACLPolicy struct {
	owner      string
	principals map[string]struct{}
	sddl       string
}

type windowsEvidenceACE struct {
	sid      string
	mask     windows.ACCESS_MASK
	aceType  uint8
	aceFlags uint8
}

type windowsEvidenceDACL struct {
	control        windows.SECURITY_DESCRIPTOR_CONTROL
	defaulted      bool
	owner          string
	ownerDefaulted bool
	entries        []windowsEvidenceACE
}

type windowsACLSizeInformation struct {
	aceCount      uint32
	aclBytesInUse uint32
	aclBytesFree  uint32
}

func prepareEvidenceNamespace(path string) (string, evidenceNamespaceState, error) {
	extendedPath, err := windowsExtendedPath(path)
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("resolve evidence path: %w", err)
	}
	ancestors, err := windowsEvidenceAncestors(extendedPath)
	if err != nil {
		return "", evidenceNamespaceState{}, err
	}
	// Open a random, protected sentinel inside the selected parent before
	// resolving the ancestor chain. Windows documents that a directory cannot be
	// renamed while it contains an open file. Starting with the sentinel avoids a
	// capture-then-create gap in which the parent could be replaced and restored.
	namespaceGuard, namespaceGuardPath, err := createWindowsEvidenceNamespaceGuard(filepath.Dir(extendedPath))
	if err != nil {
		return "", evidenceNamespaceState{}, fmt.Errorf("create evidence namespace guard: %w", err)
	}
	state := evidenceNamespaceState{namespaceGuard: namespaceGuard, namespaceGuardPath: namespaceGuardPath}
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return "", evidenceNamespaceState{}, errors.Join(
			fmt.Errorf("build evidence namespace policy: %w", err),
			closeEvidenceNamespace(&state),
		)
	}
	parentIdentity, guardHandles, err := captureWindowsEvidenceAncestors(ancestors, policy)
	if err != nil {
		return "", evidenceNamespaceState{}, errors.Join(err, closeEvidenceNamespace(&state))
	}
	state.parentPath = ancestors[len(ancestors)-1]
	state.parentIdentity = parentIdentity
	state.guardHandles = guardHandles
	return extendedPath, state, nil
}

func closeEvidenceNamespace(state *evidenceNamespaceState) error {
	if state == nil {
		return nil
	}
	var result error
	if state.namespaceGuard != nil {
		if err := state.namespaceGuard.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("close evidence namespace file guard: %w", err))
		}
		state.namespaceGuard = nil
	}
	if state.namespaceGuardPath != "" {
		// The guard was created with FILE_FLAG_DELETE_ON_CLOSE, so CloseHandle
		// deletes the bound file object without a path-based cleanup race. This
		// path lookup is verification-only and never removes a replacement entry.
		if _, err := os.Lstat(state.namespaceGuardPath); err == nil {
			result = errors.Join(result, errors.New("evidence namespace file guard remains after delete-on-close"))
		} else if !os.IsNotExist(err) {
			result = errors.Join(result, fmt.Errorf("verify evidence namespace file guard deletion: %w", err))
		}
		state.namespaceGuardPath = ""
	}
	for index := len(state.guardHandles) - 1; index >= 0; index-- {
		if err := windows.CloseHandle(state.guardHandles[index]); err != nil {
			result = errors.Join(result, fmt.Errorf("close evidence ancestor guard %d: %w", index, err))
		}
	}
	state.guardHandles = nil
	return result
}

func captureWindowsEvidenceAncestors(ancestors []string, policy windowsEvidenceACLPolicy) (evidenceFileIdentity, []windows.Handle, error) {
	if len(ancestors) == 0 {
		return evidenceFileIdentity{}, nil, errors.New("evidence destination has no parent namespace")
	}
	handles := make([]windows.Handle, 0, len(ancestors))
	cleanup := func(primary error) error {
		for index := len(handles) - 1; index >= 0; index-- {
			if err := windows.CloseHandle(handles[index]); err != nil {
				primary = errors.Join(primary, fmt.Errorf("close evidence ancestor guard after validation failure: %w", err))
			}
		}
		return primary
	}
	var parentIdentity evidenceFileIdentity
	for index, ancestor := range ancestors {
		// With FILE_SHARE_DELETE deliberately omitted, Windows keeps every opened
		// ancestor from being renamed or deleted until publication and verification
		// finish. This closes the path-component race without assuming permissive
		// ancestor ACLs are private.
		handle, identity, err := openWindowsEvidenceObject(ancestor, true, false, "evidence ancestor guard")
		if err != nil {
			return evidenceFileIdentity{}, nil, cleanup(err)
		}
		handles = append(handles, handle)
		if index != len(ancestors)-1 {
			continue
		}
		snapshot, err := inspectWindowsEvidenceDACL(handle)
		if err == nil {
			err = verifyTrustedWindowsEvidenceParent(snapshot, policy)
		}
		if err != nil {
			return evidenceFileIdentity{}, nil, cleanup(fmt.Errorf("verify controlled evidence parent %q: %w", ancestor, err))
		}
		parentIdentity = identity
	}
	return parentIdentity, handles, nil
}

func verifyEvidenceNamespace(path string, expected evidenceNamespaceState) error {
	ancestors, err := windowsEvidenceAncestors(path)
	if err != nil {
		return err
	}
	if ancestors[len(ancestors)-1] != expected.parentPath {
		return errors.New("evidence destination parent changed after validation")
	}
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return err
	}
	actual, err := verifyWindowsEvidenceAncestors(ancestors, policy)
	if err != nil {
		return err
	}
	if actual != expected.parentIdentity {
		return fmt.Errorf("evidence parent identity changed: got %s, require %s", formatWindowsEvidenceIdentity(actual), formatWindowsEvidenceIdentity(expected.parentIdentity))
	}
	return nil
}

func windowsEvidenceAncestors(destination string) ([]string, error) {
	if err := validateWindowsExtendedPath(destination); err != nil {
		return nil, fmt.Errorf("validate evidence destination namespace: %w", err)
	}
	lower := strings.ToLower(destination)
	var root string
	var components []string
	if strings.HasPrefix(lower, `\\?\unc\`) {
		remainder := destination[len(`\\?\UNC\`):]
		parts := strings.Split(remainder, `\`)
		if len(parts) < 3 || parts[len(parts)-1] == "" {
			return nil, errors.New("UNC evidence destination must include a filename below a server share")
		}
		root = `\\?\UNC\` + parts[0] + `\` + parts[1] + `\`
		components = parts[2 : len(parts)-1]
	} else {
		if len(destination) < len(`\\?\C:\`) {
			return nil, errors.New("drive evidence destination must include an absolute filename")
		}
		root = destination[:len(`\\?\C:\`)]
		remainder := destination[len(root):]
		parts := strings.Split(remainder, `\`)
		if len(parts) == 0 || parts[len(parts)-1] == "" {
			return nil, errors.New("drive evidence destination must include a filename")
		}
		components = parts[:len(parts)-1]
	}
	ancestors := []string{root}
	current := strings.TrimSuffix(root, `\`)
	for _, component := range components {
		current += `\` + component
		ancestors = append(ancestors, current)
	}
	return ancestors, nil
}

func verifyWindowsEvidenceAncestors(ancestors []string, policy windowsEvidenceACLPolicy) (evidenceFileIdentity, error) {
	if len(ancestors) == 0 {
		return evidenceFileIdentity{}, errors.New("evidence destination has no parent namespace")
	}
	var parentIdentity evidenceFileIdentity
	for index, ancestor := range ancestors {
		handle, identity, err := openWindowsEvidenceObject(ancestor, true, true, "evidence ancestor")
		if err != nil {
			return evidenceFileIdentity{}, err
		}
		if index == len(ancestors)-1 {
			snapshot, inspectErr := inspectWindowsEvidenceDACL(handle)
			if inspectErr == nil {
				inspectErr = verifyTrustedWindowsEvidenceParent(snapshot, policy)
			}
			if closeErr := windows.CloseHandle(handle); closeErr != nil {
				inspectErr = errors.Join(inspectErr, fmt.Errorf("close evidence parent verification handle: %w", closeErr))
			}
			if inspectErr != nil {
				return evidenceFileIdentity{}, fmt.Errorf("verify controlled evidence parent %q: %w", ancestor, inspectErr)
			}
			parentIdentity = identity
			continue
		}
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			return evidenceFileIdentity{}, fmt.Errorf("close evidence ancestor verification handle: %w", closeErr)
		}
	}
	return parentIdentity, nil
}

func verifyTrustedWindowsEvidenceParent(snapshot windowsEvidenceDACL, policy windowsEvidenceACLPolicy) error {
	if snapshot.ownerDefaulted {
		return errors.New("evidence parent owner is defaulted")
	}
	if _, trusted := policy.principals[snapshot.owner]; !trusted {
		return fmt.Errorf("evidence parent owner %s is not a trusted principal", snapshot.owner)
	}
	if snapshot.control&windows.SE_DACL_PRESENT == 0 {
		return errors.New("evidence parent DACL is not present")
	}
	if snapshot.defaulted {
		return errors.New("evidence parent DACL is defaulted")
	}
	for _, entry := range snapshot.entries {
		if entry.aceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			continue
		}
		if _, trusted := policy.principals[entry.sid]; trusted {
			continue
		}
		if entry.aceFlags&windows.INHERIT_ONLY_ACE != 0 {
			continue
		}
		if dangerous := entry.mask & windowsEvidenceNamespaceMutation; dangerous != 0 {
			return fmt.Errorf("evidence parent grants unexpected principal %s namespace-mutation rights %#x", entry.sid, dangerous)
		}
	}
	return nil
}

func createPrivateEvidenceTemporary(dir string) (*os.File, string, error) {
	return createPrivateWindowsEvidenceFile(dir, false)
}

func createWindowsEvidenceNamespaceGuard(dir string) (*os.File, string, error) {
	return createPrivateWindowsEvidenceFile(dir, true)
}

func createPrivateWindowsEvidenceFile(dir string, deleteOnClose bool) (*os.File, string, error) {
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return nil, "", fmt.Errorf("build evidence DACL policy: %w", err)
	}
	descriptor, err := windows.SecurityDescriptorFromString(policy.sddl)
	if err != nil {
		return nil, "", fmt.Errorf("build protected evidence security descriptor: %w", err)
	}
	securityAttributes := windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})), // #nosec G103 -- Win32 requires the ABI structure size.
		SecurityDescriptor: descriptor,
		InheritHandle:      0,
	}

	for attempt := 0; attempt < windowsEvidenceCreateAttempts; attempt++ {
		randomName := make([]byte, windowsEvidenceRandomBytes)
		if _, err := rand.Read(randomName); err != nil {
			return nil, "", fmt.Errorf("generate evidence temporary filename: %w", err)
		}
		temporaryPath := filepath.Join(dir, strings.TrimSuffix(evidenceTemporaryPattern, "*")+hex.EncodeToString(randomName))
		extendedPath, err := windowsExtendedPath(temporaryPath)
		if err != nil {
			return nil, "", fmt.Errorf("resolve evidence temporary path: %w", err)
		}
		pathUTF16, err := windows.UTF16PtrFromString(extendedPath)
		if err != nil {
			return nil, "", fmt.Errorf("encode evidence temporary path: %w", err)
		}

		desiredAccess := uint32(windows.GENERIC_READ | windows.GENERIC_WRITE | windows.READ_CONTROL)
		flags := uint32(windows.FILE_ATTRIBUTE_NORMAL)
		if deleteOnClose {
			desiredAccess |= windows.DELETE
			flags |= windows.FILE_FLAG_DELETE_ON_CLOSE
		}
		handle, createErr := windows.CreateFile(
			pathUTF16,
			desiredAccess,
			windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
			&securityAttributes,
			windows.CREATE_NEW,
			flags,
			0,
		)
		runtime.KeepAlive(descriptor)
		runtime.KeepAlive(securityAttributes)
		if errors.Is(createErr, windows.ERROR_FILE_EXISTS) || errors.Is(createErr, windows.ERROR_ALREADY_EXISTS) {
			continue
		}
		if createErr != nil {
			return nil, "", fmt.Errorf("CreateFileW private evidence temporary file: %w", createErr)
		}

		if err := verifyWindowsEvidenceHandle(handle, policy); err != nil {
			return nil, "", cleanupCreatedWindowsEvidence(handle, temporaryPath, fmt.Errorf("verify create-time evidence owner and DACL: %w", err))
		}
		file := os.NewFile(uintptr(handle), temporaryPath)
		if file == nil {
			return nil, "", cleanupCreatedWindowsEvidence(handle, temporaryPath, errors.New("wrap evidence temporary file handle"))
		}
		return file, temporaryPath, nil
	}
	return nil, "", fmt.Errorf("create private evidence temporary file: %d cryptographic name collisions", windowsEvidenceCreateAttempts)
}

func cleanupCreatedWindowsEvidence(handle windows.Handle, path string, primary error) error {
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		primary = errors.Join(primary, fmt.Errorf("close created evidence handle during cleanup: %w", closeErr))
	}
	if removeErr := removeEvidenceFile(path); removeErr != nil && !os.IsNotExist(removeErr) {
		primary = errors.Join(primary, fmt.Errorf("remove created evidence during cleanup: %w", removeErr))
	}
	return primary
}

func evidenceIdentityFromOpenFile(file *os.File) (evidenceFileIdentity, error) {
	identity, attributes, err := windowsEvidenceIdentityFromHandle(windows.Handle(file.Fd()))
	if err != nil {
		return evidenceFileIdentity{}, err
	}
	if attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 || attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		return evidenceFileIdentity{}, errors.New("evidence temporary handle is not a non-reparse regular file")
	}
	return identity, nil
}

func verifyEvidenceTemporary(_ string, temporaryPath string, expected evidenceFileIdentity) error {
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return err
	}
	handle, actual, err := openWindowsEvidenceObject(temporaryPath, false, false, "evidence temporary file")
	if err != nil {
		return err
	}
	verificationErr := compareWindowsEvidenceIdentity(actual, expected)
	if verificationErr == nil {
		verificationErr = verifyWindowsEvidenceHandle(handle, policy)
	}
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		verificationErr = errors.Join(verificationErr, fmt.Errorf("close evidence temporary verification handle: %w", closeErr))
	}
	return verificationErr
}

func verifyPrivateEvidenceDestination(path string, expected evidenceFileIdentity) error {
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return fmt.Errorf("build evidence DACL policy: %w", err)
	}
	handle, actual, err := openWindowsEvidenceObject(path, false, true, "published evidence")
	if err != nil {
		return err
	}
	verificationErr := compareWindowsEvidenceIdentity(actual, expected)
	if verificationErr == nil {
		verificationErr = verifyWindowsEvidenceHandle(handle, policy)
	}
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		verificationErr = errors.Join(verificationErr, fmt.Errorf("close published evidence verification handle: %w", closeErr))
	}
	return verificationErr
}

func openWindowsEvidenceObject(path string, directory bool, shareDelete bool, label string) (windows.Handle, evidenceFileIdentity, error) {
	extendedPath, err := windowsExtendedPath(path)
	if err != nil {
		return windows.InvalidHandle, evidenceFileIdentity{}, fmt.Errorf("resolve %s path: %w", label, err)
	}
	pathUTF16, err := windows.UTF16PtrFromString(extendedPath)
	if err != nil {
		return windows.InvalidHandle, evidenceFileIdentity{}, fmt.Errorf("encode %s path: %w", label, err)
	}
	shareMode := uint32(windows.FILE_SHARE_READ | windows.FILE_SHARE_WRITE)
	if shareDelete {
		shareMode |= windows.FILE_SHARE_DELETE
	}
	flags := uint32(windows.FILE_ATTRIBUTE_NORMAL | windows.FILE_FLAG_OPEN_REPARSE_POINT)
	if directory {
		flags |= windows.FILE_FLAG_BACKUP_SEMANTICS
	}
	handle, err := windows.CreateFile(
		pathUTF16,
		windows.READ_CONTROL|windows.FILE_READ_ATTRIBUTES,
		shareMode,
		nil,
		windows.OPEN_EXISTING,
		flags,
		0,
	)
	if err != nil {
		return windows.InvalidHandle, evidenceFileIdentity{}, fmt.Errorf("open %s without following reparse points: %w", label, err)
	}
	identity, attributes, inspectErr := windowsEvidenceIdentityFromHandle(handle)
	if inspectErr == nil && attributes&windows.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		inspectErr = fmt.Errorf("%s is a reparse point", label)
	}
	if inspectErr == nil && directory && attributes&windows.FILE_ATTRIBUTE_DIRECTORY == 0 {
		inspectErr = fmt.Errorf("%s is not a directory", label)
	}
	if inspectErr == nil && !directory && attributes&windows.FILE_ATTRIBUTE_DIRECTORY != 0 {
		inspectErr = fmt.Errorf("%s is a directory", label)
	}
	if inspectErr != nil {
		if closeErr := windows.CloseHandle(handle); closeErr != nil {
			inspectErr = errors.Join(inspectErr, fmt.Errorf("close invalid %s handle: %w", label, closeErr))
		}
		return windows.InvalidHandle, evidenceFileIdentity{}, inspectErr
	}
	return handle, identity, nil
}

func windowsEvidenceIdentityFromHandle(handle windows.Handle) (evidenceFileIdentity, uint32, error) {
	var info windows.ByHandleFileInformation
	if err := windows.GetFileInformationByHandle(handle, &info); err != nil {
		return evidenceFileIdentity{}, 0, fmt.Errorf("GetFileInformationByHandle evidence object: %w", err)
	}
	identity := evidenceFileIdentity{
		volumeSerial:  info.VolumeSerialNumber,
		fileIndexHigh: info.FileIndexHigh,
		fileIndexLow:  info.FileIndexLow,
	}
	if identity.fileIndexHigh == 0 && identity.fileIndexLow == 0 {
		return evidenceFileIdentity{}, 0, errors.New("filesystem did not expose a stable Windows file ID")
	}
	return identity, info.FileAttributes, nil
}

func compareWindowsEvidenceIdentity(actual evidenceFileIdentity, expected evidenceFileIdentity) error {
	if actual != expected {
		return fmt.Errorf("evidence file identity changed: got %s, require %s", formatWindowsEvidenceIdentity(actual), formatWindowsEvidenceIdentity(expected))
	}
	return nil
}

func formatWindowsEvidenceIdentity(identity evidenceFileIdentity) string {
	return fmt.Sprintf("volume=%08x file=%08x%08x", identity.volumeSerial, identity.fileIndexHigh, identity.fileIndexLow)
}

func verifyWindowsEvidenceHandle(handle windows.Handle, policy windowsEvidenceACLPolicy) error {
	snapshot, err := inspectWindowsEvidenceDACL(handle)
	if err != nil {
		return err
	}
	return verifyWindowsEvidenceSnapshot(snapshot, policy)
}

func verifyWindowsEvidenceSnapshot(snapshot windowsEvidenceDACL, policy windowsEvidenceACLPolicy) error {
	if snapshot.ownerDefaulted {
		return errors.New("evidence owner is defaulted")
	}
	if snapshot.owner != policy.owner {
		return fmt.Errorf("evidence owner is %s, require effective user %s", snapshot.owner, policy.owner)
	}
	if snapshot.control&windows.SE_DACL_PRESENT == 0 {
		return errors.New("evidence DACL is not present")
	}
	if snapshot.control&windows.SE_DACL_PROTECTED == 0 {
		return errors.New("evidence DACL is not protected from inheritance")
	}
	if snapshot.defaulted {
		return errors.New("evidence DACL is defaulted")
	}
	if len(snapshot.entries) != len(policy.principals) {
		return fmt.Errorf("evidence DACL has %d ACEs, require %d", len(snapshot.entries), len(policy.principals))
	}

	missing := make(map[string]struct{}, len(policy.principals))
	for principal := range policy.principals {
		missing[principal] = struct{}{}
	}
	for _, entry := range snapshot.entries {
		if entry.aceType != windows.ACCESS_ALLOWED_ACE_TYPE {
			return fmt.Errorf("evidence DACL contains non-allow ACE type %d", entry.aceType)
		}
		if entry.aceFlags != 0 {
			return fmt.Errorf("evidence DACL ACE for %s has inherited/object flags %#x", entry.sid, entry.aceFlags)
		}
		if entry.mask != windowsEvidenceFullAccess {
			return fmt.Errorf("evidence DACL ACE for %s has mask %#x, require %#x", entry.sid, entry.mask, windowsEvidenceFullAccess)
		}
		if _, ok := missing[entry.sid]; !ok {
			return fmt.Errorf("evidence DACL grants unexpected or duplicate principal %s", entry.sid)
		}
		delete(missing, entry.sid)
	}
	if len(missing) != 0 {
		return fmt.Errorf("evidence DACL is missing %d required principal(s)", len(missing))
	}
	return nil
}

func inspectWindowsEvidenceDACL(handle windows.Handle) (windowsEvidenceDACL, error) {
	descriptor, err := windows.GetSecurityInfo(
		handle,
		windows.SE_FILE_OBJECT,
		windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION,
	)
	if err != nil {
		return windowsEvidenceDACL{}, fmt.Errorf("GetSecurityInfo evidence owner and DACL: %w", err)
	}
	if descriptor == nil || !descriptor.IsValid() {
		return windowsEvidenceDACL{}, errors.New("evidence security descriptor is invalid")
	}
	control, _, err := descriptor.Control()
	if err != nil {
		return windowsEvidenceDACL{}, fmt.Errorf("read evidence security descriptor control: %w", err)
	}
	owner, ownerDefaulted, err := descriptor.Owner()
	if err != nil {
		return windowsEvidenceDACL{}, fmt.Errorf("read evidence owner: %w", err)
	}
	if owner == nil || !owner.IsValid() {
		return windowsEvidenceDACL{}, errors.New("evidence owner SID is invalid")
	}
	ownerString := owner.String()
	if ownerString == "" {
		return windowsEvidenceDACL{}, errors.New("evidence owner SID could not be rendered")
	}
	dacl, defaulted, err := descriptor.DACL()
	if err != nil {
		return windowsEvidenceDACL{}, fmt.Errorf("read evidence DACL: %w", err)
	}
	if dacl == nil {
		return windowsEvidenceDACL{}, errors.New("evidence has a null DACL")
	}
	entries, err := inspectWindowsACL(dacl)
	if err != nil {
		return windowsEvidenceDACL{}, err
	}
	runtime.KeepAlive(descriptor)
	return windowsEvidenceDACL{
		control:        control,
		defaulted:      defaulted,
		owner:          ownerString,
		ownerDefaulted: ownerDefaulted,
		entries:        entries,
	}, nil
}

func inspectWindowsACL(dacl *windows.ACL) ([]windowsEvidenceACE, error) {
	information, err := getWindowsACLSizeInformation(dacl)
	if err != nil {
		return nil, err
	}
	aclBase := uintptr(unsafe.Pointer(dacl)) // #nosec G103 -- bounds originate from validated GetAclInformation output.
	aclEnd := aclBase + uintptr(information.aclBytesInUse)
	aclDataStart := aclBase + windowsACLHeaderBytes
	if aclEnd < aclBase || aclDataStart < aclBase || information.aclBytesInUse < windowsACLHeaderBytes {
		return nil, errors.New("evidence DACL byte range is invalid")
	}
	entries := make([]windowsEvidenceACE, 0, information.aceCount)
	for index := uint32(0); index < information.aceCount; index++ {
		var rawACE *windows.ACCESS_ALLOWED_ACE
		if err := windows.GetAce(dacl, index, &rawACE); err != nil {
			return nil, fmt.Errorf("read evidence DACL ACE %d: %w", index, err)
		}
		if rawACE == nil {
			return nil, fmt.Errorf("evidence DACL ACE %d is nil", index)
		}
		rawACEPointer := unsafe.Pointer(rawACE) // #nosec G103 -- GetAce returns the documented ACE pointer.
		aceAddress := uintptr(rawACEPointer)    // #nosec G103 -- pointer is checked against the validated ACL range before dereference.
		if aceAddress < aclDataStart || aceAddress > aclEnd-unsafe.Sizeof(windows.ACE_HEADER{}) {
			return nil, fmt.Errorf("evidence DACL ACE %d header is outside the ACL", index)
		}
		header := (*windows.ACE_HEADER)(rawACEPointer) // #nosec G103 -- the complete header was range-checked above.
		if header.AceType != windows.ACCESS_ALLOWED_ACE_TYPE && header.AceType != windows.ACCESS_DENIED_ACE_TYPE {
			return nil, fmt.Errorf("evidence DACL ACE %d has unsupported type %d", index, header.AceType)
		}
		aceSize := uintptr(header.AceSize)
		aceEnd := aceAddress + aceSize
		if aceEnd < aceAddress || aceSize < unsafe.Offsetof(windows.ACCESS_ALLOWED_ACE{}.SidStart)+windowsMinimumSIDBytes || aceEnd > aclEnd {
			return nil, fmt.Errorf("evidence DACL ACE %d has an invalid bounded size", index)
		}
		sidOffset := unsafe.Offsetof(windows.ACCESS_ALLOWED_ACE{}.SidStart)
		sidAddress := aceAddress + sidOffset
		if sidAddress < aceAddress {
			return nil, fmt.Errorf("evidence DACL ACE %d SID address overflowed", index)
		}
		sidPointer := unsafe.Add(rawACEPointer, sidOffset) // #nosec G103 -- offset and complete SID are bounded against the ACE below.
		// A SID header is eight bytes. Bound it before reading SubAuthorityCount or
		// invoking any SID API, then bound the calculated complete SID as well.
		minimumSIDEnd := sidAddress + windowsMinimumSIDBytes
		if minimumSIDEnd < sidAddress || minimumSIDEnd > aceEnd {
			return nil, fmt.Errorf("evidence DACL ACE %d has a truncated SID header", index)
		}
		subAuthorityCount := *(*uint8)(unsafe.Add(sidPointer, 1)) // #nosec G103 -- the SID header byte was bounded above.
		sidSize := uintptr(windowsMinimumSIDBytes) + uintptr(subAuthorityCount)*4
		if sidAddress+sidSize < sidAddress || sidAddress+sidSize > aceEnd {
			return nil, fmt.Errorf("evidence DACL ACE %d SID extends beyond the ACE", index)
		}
		sid := (*windows.SID)(sidPointer) // #nosec G103 -- the full variable-length SID was bounded before this conversion.
		if !sid.IsValid() {
			return nil, fmt.Errorf("evidence DACL ACE %d has an invalid SID", index)
		}
		if uintptr(sid.Len()) != sidSize {
			return nil, fmt.Errorf("evidence DACL ACE %d SID length changed during validation", index)
		}
		sidString := sid.String()
		if sidString == "" {
			return nil, fmt.Errorf("evidence DACL ACE %d SID could not be rendered", index)
		}
		mask := *(*windows.ACCESS_MASK)(unsafe.Add(rawACEPointer, unsafe.Offsetof(windows.ACCESS_ALLOWED_ACE{}.Mask))) // #nosec G103 -- the fixed mask lies before the already-bounded SID.
		entries = append(entries, windowsEvidenceACE{
			sid:      sidString,
			mask:     mask,
			aceType:  header.AceType,
			aceFlags: header.AceFlags,
		})
	}
	runtime.KeepAlive(dacl)
	return entries, nil
}

func getWindowsACLSizeInformation(dacl *windows.ACL) (windowsACLSizeInformation, error) {
	if dacl == nil {
		return windowsACLSizeInformation{}, errors.New("evidence DACL is nil")
	}
	valid, _, validErr := windowsProcIsValidACL.Call(uintptr(unsafe.Pointer(dacl))) // #nosec G103 -- IsValidAcl is the documented Win32 validator for PACL.
	runtime.KeepAlive(dacl)
	if valid == 0 {
		return windowsACLSizeInformation{}, fmt.Errorf("IsValidAcl rejected evidence DACL: %v", normalizeWindowsCallError(validErr))
	}
	var information windowsACLSizeInformation
	ok, _, infoErr := windowsProcGetACLInformation.Call(
		uintptr(unsafe.Pointer(dacl)),
		uintptr(unsafe.Pointer(&information)),
		unsafe.Sizeof(information),
		windowsACLSizeInformationClass,
	) // #nosec G103 -- GetAclInformation writes the documented ACL_SIZE_INFORMATION structure.
	runtime.KeepAlive(dacl)
	if ok == 0 {
		return windowsACLSizeInformation{}, fmt.Errorf("GetAclInformation evidence DACL: %v", normalizeWindowsCallError(infoErr))
	}
	return information, nil
}

func normalizeWindowsCallError(err error) error {
	if err == nil || errors.Is(err, windows.ERROR_SUCCESS) {
		return errors.New("Win32 call returned false without an error code")
	}
	return err
}

func newWindowsEvidenceACLPolicy() (windowsEvidenceACLPolicy, error) {
	tokenUser, err := windows.GetCurrentThreadEffectiveToken().GetTokenUser()
	if err != nil {
		return windowsEvidenceACLPolicy{}, fmt.Errorf("get current effective user SID: %w", err)
	}
	systemSID, err := windows.CreateWellKnownSid(windows.WinLocalSystemSid)
	if err != nil {
		return windowsEvidenceACLPolicy{}, fmt.Errorf("create LocalSystem SID: %w", err)
	}
	administratorsSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		return windowsEvidenceACLPolicy{}, fmt.Errorf("create Administrators SID: %w", err)
	}
	if tokenUser == nil || tokenUser.User.Sid == nil || !tokenUser.User.Sid.IsValid() {
		return windowsEvidenceACLPolicy{}, errors.New("current effective user SID is invalid")
	}
	owner := tokenUser.User.Sid.String()
	if owner == "" {
		return windowsEvidenceACLPolicy{}, errors.New("current effective user SID could not be rendered")
	}

	principals := make(map[string]struct{}, 3)
	for _, sid := range []*windows.SID{tokenUser.User.Sid, systemSID, administratorsSID} {
		if sid == nil || !sid.IsValid() {
			return windowsEvidenceACLPolicy{}, errors.New("evidence DACL principal SID is invalid")
		}
		sidString := sid.String()
		if sidString == "" {
			return windowsEvidenceACLPolicy{}, errors.New("evidence DACL principal SID could not be rendered")
		}
		principals[sidString] = struct{}{}
	}

	principalNames := make([]string, 0, len(principals))
	for principal := range principals {
		principalNames = append(principalNames, principal)
	}
	sort.Strings(principalNames)
	// O:<effective-user> sets the owner explicitly. D:P protects the DACL from
	// inheritance; the verifier requires the exact owner and exact ACE set.
	var descriptor strings.Builder
	fmt.Fprintf(&descriptor, "O:%sD:P", owner)
	for _, principal := range principalNames {
		fmt.Fprintf(&descriptor, "(A;;FA;;;%s)", principal)
	}
	return windowsEvidenceACLPolicy{owner: owner, principals: principals, sddl: descriptor.String()}, nil
}

func validateWindowsExtendedPath(path string) error {
	if strings.ContainsRune(path, '\x00') {
		return errors.New("Windows path contains NUL")
	}
	if strings.Contains(path, `/`) {
		return errors.New("extended Windows path must use backslash separators")
	}
	lower := strings.ToLower(path)
	var components []string
	switch {
	case strings.HasPrefix(lower, `\\?\unc\`):
		remainder := path[len(`\\?\UNC\`):]
		components = strings.Split(remainder, `\`)
		if len(components) < 2 || components[0] == "" || components[1] == "" {
			return errors.New("extended UNC path must identify a server and share")
		}
	case strings.HasPrefix(lower, `\\?\`):
		if len(path) < len(`\\?\C:\`) || !isASCIIWindowsDriveLetter(path[4]) || path[5] != ':' || path[6] != '\\' {
			return errors.New("extended Windows path must be an absolute drive path or UNC path")
		}
		components = strings.Split(path[len(`\\?\C:\`):], `\`)
	default:
		return errors.New("Windows path is outside the allowed extended drive/UNC namespace")
	}
	for index, component := range components {
		if component == "" {
			if index != len(components)-1 {
				return errors.New("Windows path contains an empty namespace component")
			}
			continue
		}
		if component == "." || component == ".." || strings.Contains(component, ":") {
			return fmt.Errorf("Windows path contains unsupported component %q", component)
		}
	}
	return nil
}

func windowsExtendedPath(path string) (string, error) {
	if strings.ContainsRune(path, '\x00') {
		return "", errors.New("Windows path contains NUL")
	}
	if strings.HasPrefix(strings.ToLower(path), `\\?\`) {
		if err := validateWindowsExtendedPath(path); err != nil {
			return "", err
		}
		if strings.HasPrefix(strings.ToLower(path), `\\?\unc\`) {
			return `\\?\UNC\` + path[len(`\\?\UNC\`):], nil
		}
		return `\\?\` + path[len(`\\?\`):], nil
	}
	if strings.HasPrefix(path, `\\.\`) {
		return "", errors.New("Windows device paths are not valid evidence destinations")
	}
	if len(path) >= 2 && isASCIIWindowsDriveLetter(path[0]) && path[1] == ':' && (len(path) == 2 || (path[2] != '\\' && path[2] != '/')) {
		return "", errors.New("drive-relative Windows paths are not valid evidence destinations")
	}
	fullPath, err := windows.FullPath(path)
	if err != nil {
		return "", err
	}
	var extended string
	if strings.HasPrefix(fullPath, `\\`) {
		extended = `\\?\UNC\` + strings.TrimPrefix(fullPath, `\\`)
	} else {
		extended = `\\?\` + fullPath
	}
	if err := validateWindowsExtendedPath(extended); err != nil {
		return "", err
	}
	return extended, nil
}

func isASCIIWindowsDriveLetter(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func validateEvidenceTargetOwner(path string) error {
	policy, err := newWindowsEvidenceACLPolicy()
	if err != nil {
		return err
	}
	handle, _, err := openWindowsEvidenceObject(path, false, true, "existing evidence destination")
	if err != nil {
		return err
	}
	owner, defaulted, ownerErr := inspectWindowsEvidenceOwner(handle)
	if ownerErr == nil {
		ownerErr = validateExistingWindowsEvidenceOwner(owner, defaulted, policy)
	}
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		ownerErr = errors.Join(ownerErr, fmt.Errorf("close existing evidence owner verification handle: %w", closeErr))
	}
	return ownerErr
}

func validateExistingWindowsEvidenceOwner(owner string, defaulted bool, policy windowsEvidenceACLPolicy) error {
	if defaulted {
		return errors.New("existing evidence destination owner is defaulted")
	}
	if _, trusted := policy.principals[owner]; !trusted {
		return fmt.Errorf("refuse evidence destination owned by untrusted principal %s", owner)
	}
	return nil
}

func inspectWindowsEvidenceOwner(handle windows.Handle) (string, bool, error) {
	descriptor, err := windows.GetSecurityInfo(handle, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return "", false, fmt.Errorf("GetSecurityInfo evidence owner: %w", err)
	}
	if descriptor == nil || !descriptor.IsValid() {
		return "", false, errors.New("evidence owner security descriptor is invalid")
	}
	owner, defaulted, err := descriptor.Owner()
	if err != nil {
		return "", false, err
	}
	if owner == nil || !owner.IsValid() {
		return "", false, errors.New("evidence owner SID is invalid")
	}
	ownerString := owner.String()
	if ownerString == "" {
		return "", false, errors.New("evidence owner SID could not be rendered")
	}
	runtime.KeepAlive(descriptor)
	return ownerString, defaulted, nil
}

func removeEvidenceDestinationIfIdentityMatches(path string, expected evidenceFileIdentity) error {
	handle, actual, err := openWindowsEvidenceObject(path, false, true, "unverified published evidence")
	if err != nil {
		return err
	}
	identityErr := compareWindowsEvidenceIdentity(actual, expected)
	if closeErr := windows.CloseHandle(handle); closeErr != nil {
		identityErr = errors.Join(identityErr, fmt.Errorf("close unverified evidence identity handle: %w", closeErr))
	}
	if identityErr != nil {
		return fmt.Errorf("refuse to remove unverified published evidence: %w", identityErr)
	}
	return removeEvidenceFile(path)
}

func removeEvidenceFile(path string) error {
	extendedPath, err := windowsExtendedPath(path)
	if err != nil {
		return &os.PathError{Op: "remove", Path: path, Err: err}
	}
	pathUTF16, err := windows.UTF16PtrFromString(extendedPath)
	if err != nil {
		return &os.PathError{Op: "remove", Path: path, Err: err}
	}
	if err := windows.DeleteFile(pathUTF16); err != nil {
		return &os.PathError{Op: "remove", Path: path, Err: err}
	}
	return nil
}

func replaceEvidenceFile(source string, destination string) error {
	extendedSource, err := windowsExtendedPath(source)
	if err != nil {
		return fmt.Errorf("resolve evidence source path: %w", err)
	}
	extendedDestination, err := windowsExtendedPath(destination)
	if err != nil {
		return fmt.Errorf("resolve evidence destination path: %w", err)
	}
	sourceUTF16, err := windows.UTF16PtrFromString(extendedSource)
	if err != nil {
		return err
	}
	destinationUTF16, err := windows.UTF16PtrFromString(extendedDestination)
	if err != nil {
		return err
	}

	// MOVEFILE_REPLACE_EXISTING requests replacement and MOVEFILE_WRITE_THROUGH
	// requests completion before return. Microsoft does not document this API as
	// an uninterrupted or reader-atomic publication contract.
	return windows.MoveFileEx(sourceUTF16, destinationUTF16, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
