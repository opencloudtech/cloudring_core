// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package openbaoauth

import (
	"fmt"
)

var requiredLiveGates = []string{
	"openbao-active-unsealed-healthy",
	"local-reviewer-token-ca-and-tokenreview-proven",
	"auth-mount-type-and-config-prestate-captured",
	"static-reviewer-jwt-absent",
	"deprecated-issuer-config-prestate-captured",
	"kv-v2-mount-and-policy-cas-prestate-captured",
	"role-prestate-captured-and-drift-fails-closed",
	"exclusive-operator-lock-covers-all-non-cas-writes-and-readbacks",
	"shared-auth-mount-never-blindly-deleted-during-rollback",
	"positive-login-and-exact-policy-ttl-proven",
	"wrong-audience-serviceaccount-namespace-and-path-denied",
	"metadata-and-list-denied",
	"eso-tokenrequest-path-proven-without-legacy-token-secret",
	"secretstore-canary-sync-rotation-revocation-and-audit-proven",
}

var nonClaims = []string{
	"no OpenBao API request was performed",
	"no Kubernetes API request was performed",
	"no authentication mount, policy, or role was applied",
	"no SecretStore or ExternalSecret readiness was proven",
	"no production or live readiness is claimed",
	"the sanitized report does not bind a target and is not an approval receipt",
}

var fixedProfile = ProfileSummary{
	AuthType: "kubernetes", KubernetesHostMode: "in-cluster-service-dns",
	ReviewerSourceMode: "pod-local-rotating-service-account",
	Audience:           "openbao", AliasNameSource: "serviceaccount_uid",
	Capabilities: []string{"read"}, BoundIdentityCount: 1,
	TokenTTL: "10m", TokenMaxTTL: "30m", TokenNoDefaultPolicy: true,
}

// Build validates the complete contract and creates the deterministic
// source-only desired-state plan. Callers cannot bypass the profile validator.
func Build(contract Contract) (Plan, []Problem) {
	if problems := Validate(contract); len(problems) != 0 {
		return Plan{}, problems
	}
	policy := fmt.Sprintf("path %q {\n  capabilities = [\"read\"]\n}\n", contract.KVV2Mount+"/data/"+contract.DataPrefix+"/*")
	authMount := AuthMountDesired{Type: "kubernetes", Description: "CloudRING Kubernetes workload authentication"}
	config := KubernetesConfigDesired{
		KubernetesHost: "https://kubernetes.default.svc", KubernetesCACert: "", TokenReviewerJWT: "",
		PEMKeys: []string{}, Issuer: "", DisableISSValidation: true, DisableLocalCAJWT: false,
	}
	aclPolicy := ACLPolicyDesired{Policy: policy, CAS: -1, CASRequired: true}
	role := KubernetesRoleDesired{
		BoundServiceAccountNames:             []string{contract.WorkloadIdentity.ServiceAccount},
		BoundServiceAccountNamespaces:        []string{contract.WorkloadIdentity.Namespace},
		BoundServiceAccountNamespaceSelector: "", Audience: contract.Audience,
		AliasNameSource: contract.AliasNameSource, TokenPolicies: []string{contract.PolicyName},
		TokenNoDefaultPolicy: contract.TokenNoDefaultPolicy, TokenTTL: contract.TokenTTL,
		TokenMaxTTL: contract.TokenMaxTTL, TokenExplicitMaxTTL: contract.TokenMaxTTL,
		TokenType: "service", TokenNumUses: 0, TokenPeriod: 0, TokenBoundCIDRs: []string{},
		TokenStrictlyBindIP: false,
	}
	actions := []Action{
		{ID: "read-auth-mount", Method: "GET", EndpointClass: "auth-mount", Target: contract.AuthMount, CASMode: "none"},
		{ID: "create-auth-mount-if-absent", Method: "POST", EndpointClass: "auth-mount", Target: contract.AuthMount, Mutates: true, Conditional: true, PreStateRequired: true, RollbackRequired: true, CASMode: "api-cas-unavailable-exclusive-lock", DesiredState: authMount, FailClosedPrecondition: "existing-mount-must-be-kubernetes", ChangeRequiresApproval: true},
		{ID: "readback-auth-mount", Method: "GET", EndpointClass: "auth-mount", Target: contract.AuthMount, Conditional: true, PreStateRequired: true, CASMode: "exact-post-write-readback", FailClosedPrecondition: "mount-type-must-equal-kubernetes"},
		{ID: "read-kubernetes-auth-config", Method: "GET", EndpointClass: "kubernetes-auth-config", Target: contract.AuthMount, CASMode: "none"},
		{ID: "configure-local-reviewer-if-absent", Method: "POST", EndpointClass: "kubernetes-auth-config", Target: contract.AuthMount, Mutates: true, Conditional: true, PreStateRequired: true, RollbackRequired: true, CASMode: "api-cas-unavailable-exclusive-lock", DesiredState: config, FailClosedPrecondition: "config-must-be-absent-or-exact-complete-desired-state", ChangeRequiresApproval: true},
		{ID: "readback-kubernetes-auth-config", Method: "GET", EndpointClass: "kubernetes-auth-config", Target: contract.AuthMount, Conditional: true, PreStateRequired: true, CASMode: "exact-post-write-readback", FailClosedPrecondition: "config-must-match-complete-v2-5-5-desired-state"},
		{ID: "read-kv-v2-mount", Method: "GET", EndpointClass: "secret-mount", Target: contract.KVV2Mount, PreStateRequired: true, CASMode: "none", FailClosedPrecondition: "mount-type-must-be-kv-with-version-two"},
		{ID: "read-acl-policy", Method: "GET", EndpointClass: "acl-policy", Target: contract.PolicyName, CASMode: "none"},
		{ID: "create-acl-policy-if-absent", Method: "POST", EndpointClass: "acl-policy", Target: contract.PolicyName, Mutates: true, Conditional: true, PreStateRequired: true, RollbackRequired: true, CASMode: "create-only-cas-minus-one", DesiredState: aclPolicy, FailClosedPrecondition: "existing-policy-drift-must-block", ChangeRequiresApproval: true},
		{ID: "readback-acl-policy", Method: "GET", EndpointClass: "acl-policy", Target: contract.PolicyName, Conditional: true, PreStateRequired: true, CASMode: "exact-post-write-readback", FailClosedPrecondition: "policy-body-cas-mode-and-version-must-match"},
		{ID: "read-kubernetes-auth-role", Method: "GET", EndpointClass: "kubernetes-auth-role", Target: contract.AuthMount + "\x00" + contract.RoleName, CASMode: "none"},
		{ID: "create-role-if-absent", Method: "POST", EndpointClass: "kubernetes-auth-role", Target: contract.AuthMount + "\x00" + contract.RoleName, Mutates: true, Conditional: true, PreStateRequired: true, RollbackRequired: true, CASMode: "api-cas-unavailable-exclusive-lock", DesiredState: role, FailClosedPrecondition: "existing-role-drift-must-block", ChangeRequiresApproval: true},
		{ID: "readback-kubernetes-auth-role", Method: "GET", EndpointClass: "kubernetes-auth-role", Target: contract.AuthMount + "\x00" + contract.RoleName, Conditional: true, PreStateRequired: true, CASMode: "exact-post-write-readback", FailClosedPrecondition: "role-must-match-complete-desired-state"},
	}
	return Plan{AuthMount: authMount, AuthConfig: config, ACLPolicy: aclPolicy, Role: role, Actions: actions}, nil
}

// buildReport produces the evidence-safe projection used by the CLI. It stays
// private so callers cannot inject arbitrary display fields into the report.
// It deliberately emits no hashes of low-entropy provider identifiers: such
// hashes are reversible by offline enumeration and are not anonymization.
func buildReport(plan Plan) Report {
	summaries := make([]ActionSummary, 0, len(plan.Actions))
	for _, action := range plan.Actions {
		summary := ActionSummary{
			ID: action.ID, Method: action.Method, EndpointClass: action.EndpointClass,
			Mutates: action.Mutates, Conditional: action.Conditional,
			PreStateRequired: action.PreStateRequired, RollbackRequired: action.RollbackRequired,
			CASMode: action.CASMode, FailClosedPrecondition: action.FailClosedPrecondition,
			ChangeApprovalRequired: action.ChangeRequiresApproval,
		}
		summaries = append(summaries, summary)
	}
	return Report{
		SchemaVersion: SchemaVersion, Mode: "plan", Status: "planned",
		MutationPerformed: false, ApplyAuthorized: false, ApplyApprovalNeeded: true, InputMaterialEchoed: false,
		IdentifierFingerprintsEmitted: false,
		Profile:                       cloneFixedProfile(),
		Actions:                       summaries, RequiredLiveGates: append([]string{}, requiredLiveGates...),
		NonClaims: append([]string{}, nonClaims...),
	}
}

func blockedReport(problems []Problem) Report {
	return Report{
		SchemaVersion: SchemaVersion, Mode: "plan", Status: "blocked",
		MutationPerformed: false, ApplyAuthorized: false, ApplyApprovalNeeded: true, InputMaterialEchoed: false,
		IdentifierFingerprintsEmitted: false,
		Profile:                       cloneFixedProfile(),
		Blockers:                      append([]Problem{}, problems...), RequiredLiveGates: append([]string{}, requiredLiveGates...),
		NonClaims: append([]string{}, nonClaims...),
	}
}

func cloneFixedProfile() ProfileSummary {
	profile := fixedProfile
	profile.Capabilities = append([]string{}, fixedProfile.Capabilities...)
	return profile
}
