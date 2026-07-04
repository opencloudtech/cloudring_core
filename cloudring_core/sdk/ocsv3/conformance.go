package ocsv3

import (
	"fmt"
	"strings"
	"time"
)

type ConformanceProblem struct {
	Surface     string `json:"surface"`
	Field       string `json:"field"`
	Message     string `json:"message"`
	Remediation string `json:"remediation"`
}

type ConformanceReport struct {
	APIVersion       string               `json:"apiVersion"`
	Kind             string               `json:"kind"`
	GeneratedAt      string               `json:"generatedAt"`
	PackageName      string               `json:"packageName"`
	PackageVersion   string               `json:"packageVersion"`
	Passed           bool                 `json:"passed"`
	Summary          string               `json:"summary"`
	CheckedSurfaces  []string             `json:"checkedSurfaces"`
	Problems         []ConformanceProblem `json:"problems,omitempty"`
	NonClaims        []string             `json:"nonClaims"`
	RecommendedNext  []string             `json:"recommendedNext"`
	ProviderNeutral  bool                 `json:"providerNeutral"`
	ProductionMutate bool                 `json:"productionMutate"`
}

type EvidenceReceipt struct {
	APIVersion      string   `json:"apiVersion"`
	Kind            string   `json:"kind"`
	GeneratedAt     string   `json:"generatedAt"`
	Status          string   `json:"status"`
	Subject         string   `json:"subject"`
	Summary         string   `json:"summary"`
	Commands        []string `json:"commands"`
	CheckedSurfaces []string `json:"checkedSurfaces"`
	NonClaims       []string `json:"nonClaims"`
}

func CheckConformance(pkg ConnectorPackage) ConformanceReport {
	report := ConformanceReport{
		APIVersion:       APIVersion,
		Kind:             "ConformanceReport",
		GeneratedAt:      time.Now().UTC().Format(time.RFC3339),
		PackageName:      pkg.Metadata.Name,
		PackageVersion:   pkg.Metadata.Version,
		CheckedSurfaces:  conformanceSurfaces(),
		NonClaims:        conformanceNonClaims(),
		RecommendedNext:  []string{"publish package only after this report passes in CI", "keep evidence receipts fresh before production enablement"},
		ProviderNeutral:  true,
		ProductionMutate: false,
	}
	var problems []ConformanceProblem
	if err := pkg.Validate(); err != nil {
		problems = append(problems, problem("schema", "package", err.Error(), "fix the connector package fields reported by the canonical OCSv3 validator"))
	}
	checkServiceSpec(&problems, pkg.Service.Spec)
	checkPackageProfiles(&problems, pkg)
	report.Problems = problems
	report.Passed = len(problems) == 0
	if report.Passed {
		report.Summary = fmt.Sprintf("OCSv3 conformance passed for %s", displayName(pkg.Metadata.Name))
		return report
	}
	report.Summary = fmt.Sprintf("OCSv3 conformance failed for %s with %d problem(s)", displayName(pkg.Metadata.Name), len(problems))
	return report
}

func BuildEvidenceReceipt(report ConformanceReport) EvidenceReceipt {
	status := "failed"
	if report.Passed {
		status = "ok"
	}
	return EvidenceReceipt{
		APIVersion:      APIVersion,
		Kind:            "EvidenceReceipt",
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		Status:          status,
		Subject:         report.PackageName,
		Summary:         report.Summary,
		Commands:        []string{"ocsctl conformance <module-package.json>"},
		CheckedSurfaces: append([]string{}, report.CheckedSurfaces...),
		NonClaims:       append([]string{}, report.NonClaims...),
	}
}

func checkServiceSpec(problems *[]ConformanceProblem, spec ServiceSpec) {
	requireSurface(problems, len(spec.Capabilities) > 0, "service", "service.spec.capabilities", "declare at least one service capability")
	requireSurface(problems, len(spec.Dependencies) > 0, "service", "service.spec.dependencies", "declare portable service dependencies")
	requireLifecycle(problems, spec.Lifecycle)
	requireSurface(problems, len(spec.Automation) > 0, "automation", "service.spec.automation", "declare at least one lifecycle automation task")
	requireSurface(problems, len(spec.UsageMeters) > 0, "billing", "service.spec.usageMeters", "declare usage meters")
	requireSurface(problems, spec.Billing.ConnectorRef != "", "billing", "service.spec.billing.connectorRef", "link service billing profile to billing connector")
	requireSurface(problems, len(spec.Billing.Meters) > 0, "billing", "service.spec.billing.meters", "link billing meters to usage meters")
	requirePortal(problems, spec)
	requireSurface(problems, len(spec.AnalyticsEvents) > 0, "analytics", "service.spec.analyticsEvents", "declare product analytics events")
	requireSurface(problems, len(spec.KubernetesBindings) > 0, "kubernetes", "service.spec.kubernetesBindings", "declare upstream Kubernetes API bindings")
	requireSurface(problems, spec.Secrets.WorkloadIdentityRef != "", "secrets", "service.spec.secrets.workloadIdentityRef", "use workload identity instead of raw secrets")
	requireSurface(problems, len(spec.Policies) > 0, "iam", "service.spec.policies", "declare IAM/policy decisions")
	requireDataLifecycle(problems, spec.DataLifecycle)
	requireStates(problems, spec.States)
	requireSupport(problems, spec.Support)
	requireSurface(problems, len(spec.EvidenceBundles) > 0, "evidence", "service.spec.evidenceBundles", "declare evidence bundles and freshness policy")
}

func requireLifecycle(problems *[]ConformanceProblem, actions []LifecycleAction) {
	required := map[string]bool{"provision": false, "backup": false, "restore": false, "export": false, "delete": false, "retry": false, "rollback": false}
	for _, action := range actions {
		name := strings.ToLower(strings.TrimSpace(action.Name))
		if _, ok := required[name]; ok {
			required[name] = true
		}
		if name == "delete" || name == "retry" || name == "repair" {
			requireSurface(problems, action.RollbackRef != "", "lifecycle", "service.spec.lifecycle."+name+".rollbackRef", "link mutating lifecycle action to rollback evidence")
		}
		requireSurface(problems, action.Idempotent, "lifecycle", "service.spec.lifecycle."+name+".idempotent", "make lifecycle action idempotent")
		requireSurface(problems, action.IdempotencyKey != "", "lifecycle", "service.spec.lifecycle."+name+".idempotencyKey", "declare idempotency key")
	}
	for name, seen := range required {
		requireSurface(problems, seen, "lifecycle", "service.spec.lifecycle."+name, "cover provision, backup, restore, export, delete, retry, and rollback")
	}
}

func requirePortal(problems *[]ConformanceProblem, spec ServiceSpec) {
	requireSurface(problems, len(spec.PortalModules) > 0, "portal", "service.spec.portalModules", "declare at least one portal/microfrontend extension")
	host := spec.UI.ModuleHost
	requireSurface(problems, host.Host != "", "portal", "service.spec.ui.moduleHost.host", "declare microfrontend host")
	requireSurface(problems, host.Runtime != "", "portal", "service.spec.ui.moduleHost.runtime", "declare microfrontend runtime")
	requireSurface(problems, host.MountRef != "", "portal", "service.spec.ui.moduleHost.mountRef", "declare microfrontend mount reference")
	requireSurface(problems, host.IntegrityRef != "", "portal", "service.spec.ui.moduleHost.integrityRef", "declare integrity evidence")
	requireSurface(problems, host.Sandbox != "", "portal", "service.spec.ui.moduleHost.sandbox", "declare sandbox policy")
}

func requireDataLifecycle(problems *[]ConformanceProblem, lifecycle DataLifecycle) {
	requireSurface(problems, lifecycle.Export.ActionRef != "", "data", "service.spec.dataLifecycle.export.actionRef", "declare export action")
	requireSurface(problems, lifecycle.Export.EvidenceRef != "", "data", "service.spec.dataLifecycle.export.evidenceRef", "link export evidence")
	requireSurface(problems, lifecycle.Delete.ActionRef != "", "data", "service.spec.dataLifecycle.delete.actionRef", "declare delete action")
	requireSurface(problems, lifecycle.Delete.EvidenceRef != "", "data", "service.spec.dataLifecycle.delete.evidenceRef", "link delete evidence")
}

func requireStates(problems *[]ConformanceProblem, states []ServiceState) {
	required := map[string]bool{"ready": false, "denied": false, "degraded": false, "blocked": false, "retryable": false}
	for _, state := range states {
		name := strings.ToLower(strings.TrimSpace(state.Name))
		if _, ok := required[name]; ok {
			required[name] = true
		}
		requireSurface(problems, state.EvidenceRef != "", "states", "service.spec.states."+name+".evidenceRef", "link state evidence")
		requireSurface(problems, state.Remediation != "", "states", "service.spec.states."+name+".remediation", "explain operator/user remediation")
	}
	for name, seen := range required {
		requireSurface(problems, seen, "states", "service.spec.states."+name, "cover ready, denied, degraded, blocked, and retryable states")
	}
}

func requireSupport(problems *[]ConformanceProblem, support SupportProfile) {
	requireSurface(problems, support.Owner != "", "support", "service.spec.support.owner", "declare support owner")
	requireSurface(problems, len(support.Diagnostics) > 0, "support", "service.spec.support.diagnostics", "declare support diagnostics")
	requireSurface(problems, support.DocsRef != "", "support", "service.spec.support.docsRef", "link support documentation")
	requireSurface(problems, len(support.Evidence) > 0, "support", "service.spec.support.evidence", "link support evidence")
}

func checkPackageProfiles(problems *[]ConformanceProblem, pkg ConnectorPackage) {
	requireSurface(problems, len(pkg.Readiness) > 0, "readiness", "readiness", "declare readiness checks")
	requireSurface(problems, len(pkg.TenantAccess.Entitlements) > 0, "tenant", "tenantAccess.entitlements", "declare tenant entitlements")
	requireSurface(problems, len(pkg.TenantAccess.Permissions) > 0, "iam", "tenantAccess.permissions", "declare IAM permissions")
	requireSurface(problems, pkg.Durability.BackupPolicyRef != "", "durability", "durability.backupPolicyRef", "declare backup policy")
	requireSurface(problems, len(pkg.Durability.RecoveryEvidence) > 0, "durability", "durability.recoveryEvidence", "link recovery evidence")
	requireSurface(problems, len(pkg.Distribution.DeploymentProfiles) > 0, "distribution", "distribution.deploymentProfiles", "declare distribution profiles")
	requireSurface(problems, len(pkg.Federation.Modes) > 0, "federation", "federation.modes", "declare federation modes")
	requireSurface(problems, len(pkg.Commercial.Roles) > 0, "commercial", "commercial.roles", "declare commercial roles")
}

func requireSurface(problems *[]ConformanceProblem, ok bool, surface string, field string, remediation string) {
	if ok {
		return
	}
	*problems = append(*problems, problem(surface, field, "missing or incomplete OCSv3 "+field, remediation))
}

func problem(surface string, field string, message string, remediation string) ConformanceProblem {
	return ConformanceProblem{Surface: surface, Field: field, Message: message, Remediation: remediation}
}

func conformanceSurfaces() []string {
	return []string{
		"schema", "service", "billing", "portal", "iam", "tenant", "support", "evidence", "readiness", "lifecycle", "durability", "states", "analytics", "distribution", "federation", "commercial",
	}
}

func conformanceNonClaims() []string {
	return []string{
		"conformance does not claim live production readiness",
		"conformance does not perform provider, DNS, Kubernetes, or billing mutation",
		"conformance requires sanitized evidence before production enablement",
	}
}

func displayName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "unnamed package"
	}
	return name
}
