package ocs

import (
	"fmt"
	"strings"
)

func validateOCSv3Package(missing *[]string, invalid *[]string, p ConnectorPackage) {
	requireOCSv3Distribution(missing, p.Distribution)
	requireOCSv3Federation(missing, invalid, p.Federation)
	requireOCSv3Commercial(missing, invalid, p.Commercial)
}

func validateOCSv3ServiceSpec(missing *[]string, spec ServiceSpec) {
	if hasPortalExtension(spec) {
		requireOCSv3ModuleHost(missing, spec.UI.ModuleHost)
	}
	requireOCSv3AnalyticsEvents(missing, spec.AnalyticsEvents)
}

func hasPortalExtension(spec ServiceSpec) bool {
	ui := spec.UI
	host := ui.ModuleHost
	return len(spec.PortalModules) > 0 ||
		ui.EmbedRef != "" ||
		ui.ContextSchemaRef != "" ||
		len(ui.HostAuthority) > 0 ||
		len(ui.ExtensionActions) > 0 ||
		len(ui.Evidence) > 0 ||
		host.Host != "" ||
		host.Runtime != "" ||
		host.MountRef != "" ||
		host.VersionRange != "" ||
		host.SignatureRef != "" ||
		host.IntegrityRef != "" ||
		host.Sandbox != "" ||
		len(host.AllowedEvents) > 0 ||
		len(host.RequiredContext) > 0
}

func requireOCSv3ModuleHost(missing *[]string, host MicrofrontendHostContract) {
	requireSurface(missing, host.Host != "", "module", "service", "ui.moduleHost.host")
	requireSurface(missing, host.Runtime != "", "module", "service", "ui.moduleHost.runtime")
	requireSurface(missing, host.MountRef != "", "module", "service", "ui.moduleHost.mountRef")
	requireSurface(missing, host.VersionRange != "", "module", "service", "ui.moduleHost.versionRange")
	requireSurface(missing, host.SignatureRef != "", "module", "service", "ui.moduleHost.signatureRef")
	requireSurface(missing, host.IntegrityRef != "", "module", "service", "ui.moduleHost.integrityRef")
	requireSurface(missing, host.Sandbox != "", "module", "service", "ui.moduleHost.sandbox")
	requireSurface(missing, len(host.AllowedEvents) > 0, "module", "service", "ui.moduleHost.allowedEvents")
	requireSurface(missing, len(host.RequiredContext) > 0, "module", "service", "ui.moduleHost.requiredContext")
}

func requireOCSv3AnalyticsEvents(missing *[]string, events []AnalyticsEvent) {
	for i, event := range events {
		prefix := fmt.Sprintf("analyticsEvents[%d]", i)
		requireSurface(missing, event.Name != "", "analytics", "service", prefix+".name")
		requireSurface(missing, event.Trigger != "", "analytics", "service", prefix+".trigger")
		requireSurface(missing, event.Subject != "", "analytics", "service", prefix+".subject")
		requireSurface(missing, len(event.Properties) > 0, "analytics", "service", prefix+".properties")
		requireSurface(missing, event.EvidenceRef != "", "analytics", "service", prefix+".evidenceRef")
	}
}

func requireOCSv3Distribution(missing *[]string, profile DistributionProfile) {
	requireSurface(missing, len(profile.DeploymentProfiles) > 0, "distribution", "package", "distribution.deploymentProfiles")
	requireSurface(missing, len(profile.Channels) > 0, "distribution", "package", "distribution.channels")
	requireSurface(missing, len(profile.InfrastructureTargets) > 0, "distribution", "package", "distribution.infrastructureTargets")
	requireSurface(missing, profile.UpdatePolicyRef != "", "distribution", "package", "distribution.updatePolicyRef")
}

func requireOCSv3Federation(missing *[]string, invalid *[]string, profile FederationProfile) {
	switch profile.Applicability {
	case ApplicabilitySupported:
		requireSurface(missing, len(profile.Modes) > 0, "federation", "package", "federation.modes")
		requireSurface(missing, profile.MessageBusRef != "", "federation", "package", "federation.messageBusRef")
		requireSurface(missing, len(profile.CrossProviderScenarios) > 0, "federation", "package", "federation.crossProviderScenarios")
		requireSurface(missing, profile.PortabilityPolicyRef != "", "federation", "package", "federation.portabilityPolicyRef")
	case ApplicabilityNotApplicable:
		requireSurface(missing, strings.TrimSpace(profile.Reason) != "", "federation", "package", "federation.reason")
		if hasFederationMetadata(profile) {
			*invalid = append(*invalid, "federation metadata must be omitted when applicability is not_applicable")
		}
	default:
		*invalid = append(*invalid, "federation.applicability must be supported or not_applicable")
	}
}

func hasFederationMetadata(profile FederationProfile) bool {
	return len(profile.Modes) > 0 || profile.MessageBusRef != "" || len(profile.CrossProviderScenarios) > 0 || profile.PortabilityPolicyRef != ""
}

func requireOCSv3Commercial(missing *[]string, invalid *[]string, profile CommercialProfile) {
	switch profile.Applicability {
	case ApplicabilitySupported:
		requireSurface(missing, len(profile.Roles) > 0, "commercial", "package", "commercial.roles")
		requireSurface(missing, profile.RevenueModel != "", "commercial", "package", "commercial.revenueModel")
		requireSurface(missing, profile.LicenseRef != "", "commercial", "package", "commercial.licenseRef")
		requireSurface(missing, profile.ExpiryBehavior != "", "commercial", "package", "commercial.expiryBehavior")
		requireSurface(missing, profile.SupportRef != "", "commercial", "package", "commercial.supportRef")
		requireSurface(missing, profile.ServiceProvenance != "", "commercial", "package", "commercial.serviceProvenance")
		requireSurface(missing, profile.ResponsibilityMatrixRef != "", "commercial", "package", "commercial.responsibilityMatrixRef")
		requireSurface(missing, profile.ContinuityPlanRef != "", "commercial", "package", "commercial.continuityPlanRef")
	case ApplicabilityNotApplicable:
		requireSurface(missing, strings.TrimSpace(profile.Reason) != "", "commercial", "package", "commercial.reason")
		if hasCommercialMetadata(profile) {
			*invalid = append(*invalid, "commercial metadata must be omitted when applicability is not_applicable")
		}
	default:
		*invalid = append(*invalid, "commercial.applicability must be supported or not_applicable")
	}
}

func hasCommercialMetadata(profile CommercialProfile) bool {
	return len(profile.Roles) > 0 || profile.RevenueModel != "" ||
		profile.LicenseRef != "" || profile.ExpiryBehavior != "" || profile.SupportRef != "" ||
		profile.ServiceProvenance != "" || profile.ResponsibilityMatrixRef != "" || profile.ContinuityPlanRef != ""
}
