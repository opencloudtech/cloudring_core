// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"bytes"
	"testing"
)

func TestVerifyPostgreSQLRecoveryEvidenceDelegatesToPublicContract(t *testing.T) {
	if err := VerifyPostgreSQLRecoveryEvidence(bytes.NewBufferString(`{}`)); err == nil {
		t.Fatal("invalid recovery evidence was accepted")
	}
}
