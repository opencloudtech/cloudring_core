// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package identity

type ManagementDecision struct {
	Authenticated bool
	TokenValid    bool
	IAMAllow      bool
	IAMErr        string
}

func ManagementPanelAllowed(decision ManagementDecision) bool {
	return decision.Authenticated && decision.TokenValid && decision.IAMAllow && decision.IAMErr == ""
}
