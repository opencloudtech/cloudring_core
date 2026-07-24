// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"io"

	"github.com/opencloudtech/CloudRING/pkg/backup/cnpgrecovery"
)

// VerifyPostgreSQLRecoveryEvidence preserves the platform-manifest verifier
// API while delegating the portable evidence contract to its public package.
func VerifyPostgreSQLRecoveryEvidence(reader io.Reader) error {
	return cnpgrecovery.VerifyEvidence(reader)
}
