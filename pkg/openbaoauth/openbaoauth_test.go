// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package openbaoauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestEvaluateBuildsDeterministicSanitizedPlan(t *testing.T) {
	input := validJSON(t)
	first, err := Evaluate(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	second, err := Evaluate(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Evaluate() second error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("Evaluate() is not deterministic")
	}
	if first.Status != "planned" || first.Mode != "plan" || len(first.Actions) != 13 {
		t.Fatalf("unexpected report status/mode/actions: %#v", first)
	}
	if first.MutationPerformed || first.ApplyAuthorized || !first.ApplyApprovalNeeded || first.InputMaterialEchoed ||
		first.IdentifierFingerprintsEmitted {
		t.Fatalf("plan made an unsafe claim: %#v", first)
	}
	if first.Profile.AuthType != "kubernetes" || first.Profile.KubernetesHostMode != "in-cluster-service-dns" ||
		first.Profile.ReviewerCredential != "rotating-pod-local-service-account-token" ||
		first.Profile.Audience != "openbao" || first.Profile.AliasNameSource != "serviceaccount_uid" ||
		!reflect.DeepEqual(first.Profile.Capabilities, []string{"read"}) || first.Profile.BoundIdentityCount != 1 ||
		first.Profile.TokenTTL != "10m" || first.Profile.TokenMaxTTL != "30m" ||
		!first.Profile.TokenNoDefaultPolicy {
		t.Fatalf("report omitted or changed fixed profile facts: %#v", first.Profile)
	}
	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("Marshal(report) error = %v", err)
	}
	for _, forbidden := range []string{
		"cloudring-consumer-example", "cloudring-openbao-reader",
		"services/cloudring-consumer-example", "capabilities =", "kubernetes.default.svc",
	} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("sanitized report reflected forbidden input or desired state %q", forbidden)
		}
	}
	alternate := validContract()
	alternate.PolicyName = "different-policy"
	alternate.RoleName = "different-role"
	alternate.WorkloadIdentity.Namespace = "different-namespace"
	alternate.WorkloadIdentity.ServiceAccount = "different-service-account"
	alternate.DataPrefix = "different/service-prefix"
	alternateInput, err := json.Marshal(alternate)
	if err != nil {
		t.Fatalf("Marshal(alternate) error = %v", err)
	}
	alternateReport, err := Evaluate(bytes.NewReader(alternateInput))
	if err != nil {
		t.Fatalf("Evaluate(alternate) error = %v", err)
	}
	if !reflect.DeepEqual(first, alternateReport) {
		t.Fatal("sanitized report changes with private contract identifiers")
	}
}

func TestBuildEnforcesExactLeastPrivilegeDesiredState(t *testing.T) {
	contract := validContract()
	plan, problems := Build(contract)
	if len(problems) != 0 {
		t.Fatalf("Build() problems = %#v", problems)
	}
	if plan.AuthMount.Type != "kubernetes" {
		t.Fatalf("auth mount type = %q", plan.AuthMount.Type)
	}
	if plan.AuthConfig.KubernetesHost != "https://kubernetes.default.svc" ||
		plan.AuthConfig.KubernetesCACert != "" || plan.AuthConfig.TokenReviewerJWT != "" ||
		plan.AuthConfig.Issuer != "" || !plan.AuthConfig.DisableISSValidation ||
		plan.AuthConfig.DisableLocalCAJWT || len(plan.AuthConfig.PEMKeys) != 0 {
		t.Fatalf("unsafe local reviewer config: %#v", plan.AuthConfig)
	}
	wantPolicy := "path \"cloudring/data/services/cloudring-consumer-example/*\" {\n  capabilities = [\"read\"]\n}\n"
	if plan.ACLPolicy.Policy != wantPolicy || plan.ACLPolicy.CAS != -1 || !plan.ACLPolicy.CASRequired {
		t.Fatalf("ACL policy = %#v", plan.ACLPolicy)
	}
	role := plan.Role
	if !reflect.DeepEqual(role.BoundServiceAccountNames, []string{"cloudring-openbao-reader"}) ||
		!reflect.DeepEqual(role.BoundServiceAccountNamespaces, []string{"cloudring-consumer-example"}) ||
		role.BoundServiceAccountNamespaceSelector != "" || role.Audience != "openbao" ||
		role.AliasNameSource != "serviceaccount_uid" ||
		!reflect.DeepEqual(role.TokenPolicies, []string{"cloudring-consumer-example-kv-read"}) ||
		!role.TokenNoDefaultPolicy || role.TokenTTL != "10m" || role.TokenMaxTTL != "30m" ||
		role.TokenExplicitMaxTTL != "30m" || role.TokenType != "service" ||
		role.TokenNumUses != 0 || role.TokenPeriod != 0 || len(role.TokenBoundCIDRs) != 0 ||
		role.TokenStrictlyBindIP {
		t.Fatalf("role violates exact workload profile: %#v", role)
	}
	if len(plan.Actions) != 13 || plan.Actions[8].CASMode != "create-only-cas-minus-one" ||
		plan.Actions[11].CASMode != "api-cas-unavailable-exclusive-lock" ||
		plan.Actions[12].CASMode != "exact-post-write-readback" {
		t.Fatalf("unexpected action ordering or concurrency contract: %#v", plan.Actions)
	}
}

func TestBuildCannotBypassValidation(t *testing.T) {
	contract := validContract()
	contract.DataPrefix = "services/*"
	plan, problems := Build(contract)
	if len(problems) == 0 || len(plan.Actions) != 0 {
		t.Fatalf("Build() accepted invalid contract: plan=%#v problems=%#v", plan, problems)
	}
}

func TestParseRejectsInvalidProfiles(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Contract)
		path string
		code string
	}{
		{"schema", func(value *Contract) { value.SchemaVersion = "v2" }, "$.schemaVersion", "unsupported_value"},
		{"reserved auth mount", func(value *Contract) { value.AuthMount = "token" }, "$.authMount", "reserved_name"},
		{"invalid kv mount", func(value *Contract) { value.KVV2Mount = "CloudRING" }, "$.kvV2Mount", "invalid_dns_label"},
		{"wildcard prefix", func(value *Contract) { value.DataPrefix = "services/*" }, "$.dataPrefix", "invalid_safe_prefix"},
		{"traversal prefix", func(value *Contract) { value.DataPrefix = "services/../private" }, "$.dataPrefix", "invalid_safe_prefix"},
		{"leading slash", func(value *Contract) { value.DataPrefix = "/services/example" }, "$.dataPrefix", "invalid_safe_prefix"},
		{"policy wildcard", func(value *Contract) { value.PolicyName = "policy-*" }, "$.policyName", "invalid_dns_label"},
		{"default policy", func(value *Contract) { value.PolicyName = "default" }, "$.policyName", "reserved_policy_name"},
		{"root policy", func(value *Contract) { value.PolicyName = "root" }, "$.policyName", "reserved_policy_name"},
		{"response wrapping policy", func(value *Contract) { value.PolicyName = "response-wrapping" }, "$.policyName", "reserved_policy_name"},
		{"role path", func(value *Contract) { value.RoleName = "role/name" }, "$.roleName", "invalid_dns_label"},
		{"namespace wildcard", func(value *Contract) { value.WorkloadIdentity.Namespace = "*" }, "$.workloadIdentity.namespace", "invalid_dns_label"},
		{"service account empty", func(value *Contract) { value.WorkloadIdentity.ServiceAccount = "" }, "$.workloadIdentity.serviceAccount", "invalid_dns_label"},
		{"audience", func(value *Contract) { value.Audience = "vault" }, "$.audience", "must_equal_openbao"},
		{"alias", func(value *Contract) { value.AliasNameSource = "serviceaccount_name" }, "$.aliasNameSource", "must_use_serviceaccount_uid"},
		{"ttl", func(value *Contract) { value.TokenTTL = "1h" }, "$.tokenTTL", "must_equal_cloudring_10m_profile"},
		{"max ttl", func(value *Contract) { value.TokenMaxTTL = "1h" }, "$.tokenMaxTTL", "must_equal_cloudring_30m_profile"},
		{"default policy", func(value *Contract) { value.TokenNoDefaultPolicy = false }, "$.tokenNoDefaultPolicy", "must_be_true"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := validContract()
			test.edit(&value)
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			_, problems := Parse(data)
			if !containsProblem(problems, test.path, test.code) {
				t.Fatalf("Parse() problems = %#v, want %s/%s", problems, test.path, test.code)
			}
		})
	}
}

func TestParseRejectsNonCanonicalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		code  string
	}{
		{"empty", "", "invalid_json"},
		{"malformed", "{", "invalid_json"},
		{"trailing", "{} {}", "invalid_json"},
		{"array", "[]", "invalid_json_contract"},
		{"unknown", `{"credentialMarker":"do-not-reflect"}`, "unknown_field"},
		{"duplicate", `{"schemaVersion":"one","schemaVersion":"two"}`, "duplicate_field"},
		{"nested duplicate", `{"workloadIdentity":{"namespace":"one","namespace":"two"}}`, "duplicate_field"},
		{"nested unknown", `{"workloadIdentity":{"credentialMarker":"do-not-reflect"}}`, "unknown_field"},
		{"wrong field type", `{"tokenNoDefaultPolicy":"true"}`, "invalid_json_contract"},
		{"excess nesting", strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34), "invalid_json_contract"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, problems := Parse([]byte(test.input))
			if len(problems) == 0 || problems[0].Code != test.code {
				t.Fatalf("Parse() problems = %#v, want code %q", problems, test.code)
			}
		})
	}
}

func TestEvaluateDoesNotReflectRejectedMaterial(t *testing.T) {
	marker := "credential-marker-do-not-reflect"
	contract := validContract()
	contract.Audience = marker
	input, err := json.Marshal(contract)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	report, err := Evaluate(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal(report) error = %v", err)
	}
	if report.Status != "blocked" || bytes.Contains(encoded, []byte(marker)) {
		t.Fatalf("blocked report reflected input: %s", encoded)
	}
}

func TestEvaluateBoundsInputAndSanitizesReadErrors(t *testing.T) {
	report, err := Evaluate(strings.NewReader(strings.Repeat("x", MaxInputBytes+1)))
	if err != nil || report.Status != "blocked" ||
		!containsProblem(report.Blockers, "$", "input_too_large") {
		t.Fatalf("oversized Evaluate() report=%#v err=%v", report, err)
	}
	if _, err := Evaluate(errorReader{}); !errors.Is(err, errInputUnavailable) {
		t.Fatalf("Evaluate(errorReader) error = %v", err)
	}
}

func validContract() Contract {
	return Contract{
		SchemaVersion: SchemaVersion, AuthMount: "kubernetes", KVV2Mount: "cloudring",
		DataPrefix: "services/cloudring-consumer-example", PolicyName: "cloudring-consumer-example-kv-read",
		RoleName:         "cloudring-consumer-example",
		WorkloadIdentity: WorkloadIdentity{Namespace: "cloudring-consumer-example", ServiceAccount: "cloudring-openbao-reader"},
		Audience:         "openbao", AliasNameSource: "serviceaccount_uid", TokenTTL: "10m", TokenMaxTTL: "30m",
		TokenNoDefaultPolicy: true,
	}
}

func validJSON(t *testing.T) []byte {
	t.Helper()
	data, err := json.Marshal(validContract())
	if err != nil {
		t.Fatalf("Marshal(valid contract) error = %v", err)
	}
	return data
}

func containsProblem(problems []Problem, path, code string) bool {
	for _, problem := range problems {
		if problem.Path == path && problem.Code == code {
			return true
		}
	}
	return false
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
