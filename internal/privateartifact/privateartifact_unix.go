// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package privateartifact

import (
	"os"
)

func privateDirectoryMode(mode os.FileMode) bool {
	return mode.Perm()&0o022 == 0
}

func syncDirectory(root *os.Root) error {
	directory, err := root.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
