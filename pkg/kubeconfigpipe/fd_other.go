//go:build !linux && !darwin

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package kubeconfigpipe

import (
	"errors"
	"os"
)

func duplicateFD(_ int) (*os.File, error) {
	return nil, errors.New("pipe-backed kubeconfig runtime is unsupported")
}
