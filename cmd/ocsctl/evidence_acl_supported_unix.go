//go:build darwin || linux

package main

import (
	"bytes"
	"fmt"

	"golang.org/x/sys/unix"
)

const maximumEvidenceXattrListBytes = 1 << 20

func verifyUnixNoExtendedACL(path string) error {
	size, err := unix.Listxattr(path, nil)
	if err != nil {
		return fmt.Errorf("list extended attributes for ACL verification: %w", err)
	}
	if size == 0 {
		return nil
	}
	if size < 0 || size > maximumEvidenceXattrListBytes {
		return fmt.Errorf("extended attribute list size %d is outside the supported ACL verification bound", size)
	}
	buffer := make([]byte, size)
	written, err := unix.Listxattr(path, buffer)
	if err != nil {
		return fmt.Errorf("read extended attributes for ACL verification: %w", err)
	}
	if written < 0 || written > len(buffer) {
		return fmt.Errorf("extended attribute list returned invalid size %d", written)
	}
	for _, name := range bytes.Split(buffer[:written], []byte{0}) {
		switch string(name) {
		case "system.posix_acl_access", "system.nfs4_acl", "trusted.SGI_ACL_FILE", "com.apple.system.Security", "com.apple.acl.text":
			return fmt.Errorf("filesystem ACL %q is outside the supported UID/mode namespace contract", name)
		}
	}
	return nil
}
