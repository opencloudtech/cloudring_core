//go:build unix

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package drill

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockJournalFile(file *os.File) error   { return unix.Flock(int(file.Fd()), unix.LOCK_EX) }
func unlockJournalFile(file *os.File) error { return unix.Flock(int(file.Fd()), unix.LOCK_UN) }
