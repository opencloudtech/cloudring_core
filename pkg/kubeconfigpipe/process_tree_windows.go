//go:build windows

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeconfigpipe

import "os/exec"

// Pipe-backed kubeconfig replay is already rejected on native Windows. Keep
// direct command execution buildable without making a containment claim.
func configureProcessTree(_ *exec.Cmd)     {}
func cleanupProcessTree(_ *exec.Cmd) error { return nil }
