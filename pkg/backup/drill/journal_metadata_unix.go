//go:build unix

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package drill

import (
	"os"
	"syscall"
)

func protectedJournalFile(info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return info != nil && ok && info.Mode().IsRegular() && info.Mode().Perm() == 0o600 && stat.Uid == uint32(os.Geteuid()) && stat.Nlink == 1
}
