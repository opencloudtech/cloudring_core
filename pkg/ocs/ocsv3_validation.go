package ocs

import "fmt"

func validateOCSv3Package(missing *[]string, p ConnectorPackage) {
	requireOCSv3Distribution(missing, p.Distribution)
	requireOCSv3Federation(missing, p.Federation)
	requireOCSv3Commercial(missing, p.Commercial)
}

func validateOCSv3ServiceSpec(missing *[]string, spec ServiceSpec) {
	requireOCSv3ModuleHost(missing, spec.UI.ModuleHost)
	requireOCSv3AnalyticsEvents(missing, spec.AnalyticsEvents)
}

func requireOCSv3ModuleHost(missing *[]string, host MicrofrontendHostContract) {
	requireSurface(missing, host.Host != "", "module", "service", "ui.moduleHost.host")
	requireSurface(missing, host.Runtime != "", "module", "service", "ui.moduleHost.runtime")
	requireSurface(missing, host.MountRef != "", "module", "service", "ui.moduleHost.mountRef")
	requireSurface(missing, host.VersionRange != "", "module", "service", "ui.moduleHost.versionRange")
	requireSurface(missing, host.IntegrityRef != "", "module", "service", "ui.moduleHost.integrityRef")
	requireSurface(missing, host.Sandbox != "", "module", "service", "ui.moduleHost.sandbox")
	requireSurface(missing, len(host.AllowedEvents) > 0, "module", "service", "ui.moduleHost.allowedEvents")
	requireSurface(missing, len(host.RequiredContext) > 0, "module", "service", "ui.moduleHost.requiredContext")
}

func requireOCSv3AnalyticsEvents(missing *[]string, events []AnalyticsEvent) {
	requireSurface(missing, len(events) > 0, "analytics", "service", "analyticsEvents")
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

func requireOCSv3Federation(missing *[]string, profile FederationProfile) {
	requireSurface(missing, len(profile.Modes) > 0, "federation", "package", "federation.modes")
	requireSurface(missing, profile.MessageBusRef != "", "federation", "package", "federation.messageBusRef")
	requireSurface(missing, len(profile.CrossProviderScenarios) > 0, "federation", "package", "federation.crossProviderScenarios")
	requireSurface(missing, profile.PortabilityPolicyRef != "", "federation", "package", "federation.portabilityPolicyRef")
}

func requireOCSv3Commercial(missing *[]string, profile CommercialProfile) {
	requireSurface(missing, len(profile.Roles) > 0, "commercial", "package", "commercial.roles")
	requireSurface(missing, profile.RevenueModel != "", "commercial", "package", "commercial.revenueModel")
	requireSurface(missing, profile.LicenseRef != "", "commercial", "package", "commercial.licenseRef")
	requireSurface(missing, profile.ExpiryBehavior != "", "commercial", "package", "commercial.expiryBehavior")
	requireSurface(missing, profile.SupportRef != "", "commercial", "package", "commercial.supportRef")
	requireSurface(missing, profile.ServiceProvenance != "", "commercial", "package", "commercial.serviceProvenance")
	requireSurface(missing, profile.ResponsibilityMatrixRef != "", "commercial", "package", "commercial.responsibilityMatrixRef")
	requireSurface(missing, profile.ContinuityPlanRef != "", "commercial", "package", "commercial.continuityPlanRef")
}
