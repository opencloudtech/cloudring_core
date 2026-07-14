//go:build windows

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package velero118

import "os/exec"

// Native Windows adapter execution is outside this collector version's Linux
// operational scope. pinExecutable rejects it before this function is reached.
func configureProcessTree(_ *exec.Cmd)     {}
func cleanupProcessTree(_ *exec.Cmd) error { return nil }
