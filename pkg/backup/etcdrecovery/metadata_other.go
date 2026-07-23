// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

//go:build !aix && !darwin && !dragonfly && !freebsd && !linux && !netbsd && !openbsd && !solaris

package etcdrecovery

import "os"

func ownedByCurrentUser(os.FileInfo) bool   { return false }
func ownedByCurrentOrRoot(os.FileInfo) bool { return false }
func workspaceRegularSafe(os.FileInfo) bool { return false }
