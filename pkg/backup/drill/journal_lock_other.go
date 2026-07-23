//go:build !unix

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package drill

import (
	"errors"
	"os"
)

func lockJournalFile(_ *os.File) error   { return errors.New("journal locking is unsupported") }
func unlockJournalFile(_ *os.File) error { return nil }
