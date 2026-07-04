package ocs

import (
	"fmt"
	"strings"
)

func validateAlpha2Package(missing *[]string, invalid *[]string, p ConnectorPackage) {
	requireSurface(missing, len(p.Durability.RecoveryEvidence) > 0, "durability", "service", "durability.recoveryEvidence")
	for i, dependency := range p.Service.Spec.Dependencies {
		prefix := fmt.Sprintf("service.spec.dependencies[%d]", i)
		if isNonPortable(dependency.Portability) || isProviderLockIn(dependency.Portability) {
			*invalid = append(*invalid, problem("portability", "service", prefix+".portability", "must describe a portable OCSv3 dependency contract"))
		}
		if isPlatformCoupled(dependency.ImplementationRef) {
			*invalid = append(*invalid, problem("coupling", "service", prefix+".implementationRef", "must not reference platform-core or internal implementation paths"))
		}
	}
	for i, route := range p.Service.Spec.GatewayRoutes {
		prefix := fmt.Sprintf("service.spec.gatewayRoutes[%d]", i)
		if isProviderLockIn(route.ParentRef) {
			*invalid = append(*invalid, problem("provider-lock-in", "service", prefix+".parentRef", "must reference a portable Gateway API parent"))
		}
	}
}

func validateAlpha2ServiceSpec(missing *[]string, invalid *[]string, spec ServiceSpec) {
	validateUI(missing, spec.UI)
	validateGatewayRoutes(missing, spec.GatewayRoutes)
	validateSecrets(missing, invalid, spec.Secrets)
	validatePolicies(missing, spec.Policies)
	validateDataLifecycle(missing, spec.DataLifecycle)
	validateDataLifecycleRefs(invalid, spec.Lifecycle, spec.DataLifecycle)
	validateStates(missing, spec.States)
	validateSupportEvidence(missing, spec.Support)
	validateEvidenceBundles(missing, spec.EvidenceBundles)
}

func validateUI(missing *[]string, ui UIExtensionManifest) {
	requireSurface(missing, ui.EmbedRef != "", "ui", "service", "ui.embedRef")
	requireSurface(missing, ui.ContextSchemaRef != "", "ui", "service", "ui.contextSchemaRef")
	requireSurface(missing, len(ui.HostAuthority) > 0, "ui", "service", "ui.hostAuthority")
	requireSurface(missing, len(ui.ExtensionActions) > 0, "ui", "service", "ui.extensionActions")
	requireSurface(missing, len(ui.Evidence) > 0, "ui", "service", "ui.evidence")
	validateEvidenceRefs(missing, "ui.evidence", ui.Evidence)
}

func validateGatewayRoutes(missing *[]string, routes []GatewayRoute) {
	requireSurface(missing, len(routes) > 0, "gateway", "service", "gatewayRoutes")
	for i, route := range routes {
		prefix := fmt.Sprintf("gatewayRoutes[%d]", i)
		requireSurface(missing, route.Name != "", "gateway", "service", prefix+".name")
		requireSurface(missing, route.ParentRef != "", "gateway", "service", prefix+".parentRef")
		requireSurface(missing, len(route.Hostnames) > 0, "gateway", "service", prefix+".hostnames")
		requireSurface(missing, len(route.Rules) > 0, "gateway", "service", prefix+".rules")
		requireSurface(missing, route.EvidenceRef != "", "gateway", "service", prefix+".evidenceRef")
	}
}

func validateSecrets(missing *[]string, invalid *[]string, secrets SecretBoundary) {
	requireSurface(missing, secrets.WorkloadIdentityRef != "", "secrets", "service", "secrets.workloadIdentityRef")
	requireSurface(missing, len(secrets.SecretRefs) > 0, "secrets", "service", "secrets.secretRefs")
	for i, ref := range secrets.SecretRefs {
		prefix := fmt.Sprintf("secrets.secretRefs[%d]", i)
		requireSurface(missing, ref.Name != "", "secrets", "service", prefix+".name")
		requireSurface(missing, ref.Ref != "", "secrets", "service", prefix+".ref")
		requireSurface(missing, ref.Purpose != "", "secrets", "service", prefix+".purpose")
		requireSurface(missing, ref.Rotation != "", "secrets", "service", prefix+".rotation")
		if strings.TrimSpace(ref.RawValue) != "" {
			*invalid = append(*invalid, problem("secrets", "service", prefix+".rawValue", "must use references only, not raw secret material"))
		}
	}
}

func validatePolicies(missing *[]string, policies []PolicyRule) {
	requireSurface(missing, len(policies) > 0, "policy", "service", "policies")
	for i, policy := range policies {
		prefix := fmt.Sprintf("policies[%d]", i)
		requireSurface(missing, policy.Name != "", "policy", "service", prefix+".name")
		requireSurface(missing, policy.Class != "", "policy", "service", prefix+".class")
		requireSurface(missing, policy.DecisionRef != "", "policy", "service", prefix+".decisionRef")
		requireSurface(missing, policy.EvidenceRef != "", "policy", "service", prefix+".evidenceRef")
	}
}

func validateDataLifecycle(missing *[]string, lifecycle DataLifecycle) {
	validateDataLifecycleAction(missing, "dataLifecycle.export", lifecycle.Export)
	validateDataLifecycleAction(missing, "dataLifecycle.delete", lifecycle.Delete)
}

func validateDataLifecycleAction(missing *[]string, path string, action DataLifecycleAction) {
	requireSurface(missing, action.ActionRef != "", "data", "service", path+".actionRef")
	requireSurface(missing, action.Format != "", "data", "service", path+".format")
	requireSurface(missing, action.EvidenceRef != "", "data", "service", path+".evidenceRef")
}

func validateDataLifecycleRefs(invalid *[]string, lifecycle []LifecycleAction, data DataLifecycle) {
	refs := map[string]bool{}
	for _, action := range lifecycle {
		if action.Name == "" {
			continue
		}
		refs[action.Name] = true
		refs["lifecycle."+action.Name] = true
	}
	for _, item := range []struct {
		path string
		ref  string
	}{
		{path: "dataLifecycle.export.actionRef", ref: data.Export.ActionRef},
		{path: "dataLifecycle.delete.actionRef", ref: data.Delete.ActionRef},
	} {
		if item.ref != "" && !refs[item.ref] {
			*invalid = append(*invalid, problem("data", "service", item.path, "must reference a lifecycle action"))
		}
	}
}

func validateStates(missing *[]string, states []ServiceState) {
	requireSurface(missing, len(states) > 0, "state", "service", "states")
	seen := map[string]bool{}
	for i, state := range states {
		prefix := fmt.Sprintf("states[%d]", i)
		name := strings.ToLower(strings.TrimSpace(state.Name))
		seen[name] = true
		requireSurface(missing, state.Name != "", "state", "service", prefix+".name")
		requireSurface(missing, state.Reason != "", "state", "service", prefix+".reason")
		requireSurface(missing, state.UserVisible, "state", "service", prefix+".userVisible")
		requireSurface(missing, state.EvidenceRef != "", "state", "service", prefix+".evidenceRef")
		requireSurface(missing, state.Remediation != "", "state", "service", prefix+".remediation")
	}
	requireSurface(missing, seen["degraded"], "state", "service", "states.degraded")
	requireSurface(missing, seen["blocked"], "state", "service", "states.blocked")
}

func validateSupportEvidence(missing *[]string, support SupportProfile) {
	requireSurface(missing, len(support.Evidence) > 0, "support", "service", "support.evidence")
	validateEvidenceRefs(missing, "support.evidence", support.Evidence)
}

func validateEvidenceBundles(missing *[]string, bundles []EvidenceBundle) {
	requireSurface(missing, len(bundles) > 0, "evidence", "service", "evidenceBundles")
	for i, bundle := range bundles {
		prefix := fmt.Sprintf("evidenceBundles[%d]", i)
		requireSurface(missing, bundle.Name != "", "evidence", "service", prefix+".name")
		requireSurface(missing, bundle.Owner != "", "evidence", "service", prefix+".owner")
		requireSurface(missing, bundle.Claim != "", "evidence", "service", prefix+".claim")
		requireSurface(missing, bundle.Freshness != "", "evidence", "service", prefix+".freshness")
		requireSurface(missing, bundle.Redaction != "", "evidence", "service", prefix+".redaction")
		requireSurface(missing, len(bundle.Evidence) > 0, "evidence", "service", prefix+".evidence")
		requireSurface(missing, len(bundle.NonClaims) > 0, "evidence", "service", prefix+".nonClaims")
		requireSurface(missing, bundle.ReviewPath != "", "evidence", "service", prefix+".reviewPath")
		validateEvidenceRefs(missing, prefix+".evidence", bundle.Evidence)
	}
}

func requireSurface(missing *[]string, ok bool, class string, owner string, path string) {
	if !ok {
		*missing = append(*missing, problem(class, owner, path, "missing required OCSv3 surface"))
	}
}

func problem(class string, owner string, path string, detail string) string {
	return fmt.Sprintf("class=%s owner=%s path=%s detail=%s", class, owner, path, detail)
}

func isPlatformCoupled(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(normalized, "internal/") ||
		strings.Contains(normalized, "internal\\") ||
		strings.Contains(normalized, "platform-internal") ||
		strings.Contains(normalized, "github.com/opencloudtech/cloudring_core/internal")
}

func isProviderLockIn(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "provider-lock-in", "provider-specific", "ovh", "aws", "azure", "gcp":
		return true
	default:
		return strings.Contains(strings.ToLower(value), "ovh/")
	}
}
