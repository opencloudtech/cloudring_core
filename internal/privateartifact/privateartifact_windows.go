// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package privateartifact

import (
	"errors"
	"os"
)

func privateDirectoryMode(os.FileMode) bool {
	return true
}

func syncDirectory(*os.Root) error {
	return nil
}

func readOwnerOnly(string, int64, func()) ([]byte, error) {
	return nil, errors.New("protected private artifact reads are unsupported on Windows")
}
