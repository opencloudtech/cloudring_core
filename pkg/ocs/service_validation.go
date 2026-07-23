package ocs

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	structuredReferencePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[.-][a-z][a-z0-9]*)*$`)
	boundedVersionRangePattern = regexp.MustCompile(`^>=[0-9]+\.[0-9]+\.[0-9]+,<[0-9]+\.[0-9]+\.[0-9]+$`)
)

func (c ServiceConnector) Validate() error {
	var missing []string
	var invalid []string

	require(&missing, c.APIVersion == APIVersion, "apiVersion")
	require(&missing, c.Kind == "ServiceConnector", "kind")
	requireMetadata(&missing, c.Metadata)
	require(&missing, c.Spec.ProductAPI.Ref != "", "productAPI.ref")
	require(&missing, c.Spec.ProductAPI.Version != "", "productAPI.version")
	require(&missing, c.Spec.ProductAPI.Protocol != "", "productAPI.protocol")
	require(&missing, len(c.Spec.Capabilities) > 0, "capabilities")
	require(&missing, len(c.Spec.Lifecycle) > 0, "lifecycle")
	require(&missing, len(c.Spec.Automation) > 0, "automation")
	require(&missing, c.Spec.Support.Owner != "", "support.owner")
	require(&missing, len(c.Spec.Support.Diagnostics) > 0, "support.diagnostics")

	validateCapabilities(&missing, c.Spec.Capabilities)
	validateDependencies(&missing, &invalid, c.Spec.Dependencies)
	validateProductAPIRefs(&invalid, c.Spec.ProductAPI)
	validateExecutionProfile(&missing, &invalid, c.Spec)
	validateKubernetesBindings(&missing, c.Spec.KubernetesBindings)
	validateLifecycle(&missing, &invalid, c.Spec.Lifecycle)
	validateAutomationTasks(&missing, c.Spec.Automation)
	validateServiceBilling(&missing, &invalid, c.Spec)
	if hasPortalExtension(c.Spec) {
		validatePortalModules(&missing, c.Spec.PortalModules)
	}
	validateAlpha2ServiceSpec(&missing, &invalid, c.Spec)
	validateOCSv3ServiceSpec(&missing, c.Spec)

	if len(missing) > 0 || len(invalid) > 0 {
		problems := append([]string{}, missing...)
		problems = append(problems, invalid...)
		return fmt.Errorf("service connector invalid: %s", strings.Join(problems, ", "))
	}
	return nil
}

func validateProductAPIRefs(invalid *[]string, api ProductAPIContract) {
	for _, item := range []struct {
		path  string
		value string
	}{
		{path: "productAPI.ref", value: api.Ref},
		{path: "productAPI.endpointRef", value: api.EndpointRef},
		{path: "productAPI.trustPolicyRef", value: api.TrustPolicyRef},
		{path: "productAPI.healthRef", value: api.HealthRef},
	} {
		if item.value == "" {
			continue
		}
		if !structuredReferencePattern.MatchString(item.value) {
			*invalid = append(*invalid, item.path+" must be a source-safe reference, not a raw endpoint or credential")
		}
	}
}

func validateServiceBilling(missing *[]string, invalid *[]string, spec ServiceSpec) {
	switch spec.Billing.Applicability {
	case ApplicabilitySupported:
		require(missing, len(spec.UsageMeters) > 0, "usageMeters")
		require(missing, spec.Billing.ConnectorRef != "", "billing.connectorRef")
		require(missing, len(spec.Billing.Meters) > 0, "billing.meters")
		validateUsageMeters(missing, "usageMeters", spec.UsageMeters)
		validateUsageMeters(missing, "billing.meters", spec.Billing.Meters)
	case ApplicabilityNotApplicable:
		require(missing, strings.TrimSpace(spec.Billing.Reason) != "", "billing.reason")
		if len(spec.UsageMeters) > 0 || spec.Billing.ConnectorRef != "" || len(spec.Billing.Meters) > 0 {
			*invalid = append(*invalid, "non-billable service must omit usageMeters, billing.connectorRef, and billing.meters")
		}
	default:
		*invalid = append(*invalid, "billing.applicability must be supported or not_applicable")
	}
}

func validateExecutionProfile(missing *[]string, invalid *[]string, spec ServiceSpec) {
	switch spec.ExecutionProfile {
	case ExecutionProfileLocal:
		// Local products may use Kubernetes bindings, but their implementation
		// technology remains outside the OCS product contract.
	case ExecutionProfileRemote, ExecutionProfileAPIOnly:
		require(missing, spec.ProductAPI.EndpointRef != "", "productAPI.endpointRef")
		require(missing, spec.ProductAPI.TrustPolicyRef != "", "productAPI.trustPolicyRef")
		require(missing, spec.ProductAPI.HealthRef != "", "productAPI.healthRef")
		if len(spec.KubernetesBindings) > 0 {
			*invalid = append(*invalid, fmt.Sprintf("kubernetesBindings must be omitted for %s execution profile", spec.ExecutionProfile))
		}
	case "":
		require(missing, false, "executionProfile")
	default:
		*invalid = append(*invalid, "executionProfile must be one of local, remote, api-only")
	}
}

func validateCapabilities(missing *[]string, capabilities []Capability) {
	for i, capability := range capabilities {
		prefix := fmt.Sprintf("capabilities[%d]", i)
		require(missing, capability.Class != "", prefix+".class")
		require(missing, capability.Name != "", prefix+".name")
	}
}

func validateDependencies(missing *[]string, invalid *[]string, dependencies []Dependency) {
	for i, dependency := range dependencies {
		prefix := fmt.Sprintf("dependencies[%d]", i)
		require(missing, dependency.ID != "", prefix+".id")
		require(missing, dependency.CapabilityClass != "", prefix+".capabilityClass")
		require(missing, dependency.Role != "", prefix+".role")
		require(missing, dependency.Portability != "", prefix+".portability")
		require(missing, dependency.ProductAPIRef != "", prefix+".productAPIRef")
		require(missing, dependency.VersionRange != "", prefix+".versionRange")
		require(missing, dependency.CompatibilityPolicyRef != "", prefix+".compatibilityPolicyRef")
		for _, item := range []struct {
			path  string
			value string
		}{
			{path: prefix + ".productAPIRef", value: dependency.ProductAPIRef},
			{path: prefix + ".compatibilityPolicyRef", value: dependency.CompatibilityPolicyRef},
		} {
			if item.value != "" && !structuredReferencePattern.MatchString(item.value) {
				*invalid = append(*invalid, item.path+" must be a structured reference")
			}
		}
		if dependency.VersionRange != "" && !boundedVersionRangePattern.MatchString(dependency.VersionRange) {
			*invalid = append(*invalid, prefix+".versionRange must be a bounded range such as >=1.0.0,<2.0.0")
		}
	}
}

func validateKubernetesBindings(missing *[]string, bindings []KubernetesBinding) {
	for i, binding := range bindings {
		prefix := fmt.Sprintf("kubernetesBindings[%d]", i)
		require(missing, binding.Group != "", prefix+".group")
		require(missing, binding.Version != "", prefix+".version")
		require(missing, binding.Kind != "", prefix+".kind")
		require(missing, binding.Plural != "", prefix+".plural")
		require(missing, binding.Scope != "", prefix+".scope")
		require(missing, binding.CRDRef != "", prefix+".crdRef")
		require(missing, binding.StatusPath != "", prefix+".statusPath")
		require(missing, binding.Condition != "", prefix+".condition")
	}
}

func validateLifecycle(missing *[]string, invalid *[]string, lifecycle []LifecycleAction) {
	seen := map[string]bool{}
	for i, action := range lifecycle {
		prefix := fmt.Sprintf("lifecycle[%d]", i)
		require(missing, action.Name != "", prefix+".name")
		name := strings.ToLower(strings.TrimSpace(action.Name))
		if name != "" {
			if seen[name] {
				*invalid = append(*invalid, prefix+".name must be unique")
			}
			seen[name] = true
		}
		switch action.Applicability {
		case ApplicabilitySupported:
			require(missing, action.Verb != "", prefix+".verb")
			require(missing, action.Idempotent, prefix+".idempotent")
			require(missing, action.IdempotencyKey != "", prefix+".idempotencyKey")
		case ApplicabilityNotApplicable:
			require(missing, strings.TrimSpace(action.Reason) != "", prefix+".reason")
			if action.Verb != "" || action.Idempotent || action.IdempotencyKey != "" || action.RollbackRef != "" {
				*invalid = append(*invalid, prefix+" must not declare executable fields when applicability is not_applicable")
			}
		default:
			*invalid = append(*invalid, prefix+".applicability must be supported or not_applicable")
		}
	}

	for _, name := range []string{"provision", "resume", "resize", "deprovision"} {
		require(missing, seen[name], "lifecycle."+name)
	}
	require(missing, seen["hold"] || seen["suspend"], "lifecycle.holdOrSuspend")
}

func validateAutomationTasks(missing *[]string, tasks []AutomationTask) {
	for i, task := range tasks {
		prefix := fmt.Sprintf("automation[%d]", i)
		require(missing, task.Name != "", prefix+".name")
		require(missing, task.ActionRef != "", prefix+".actionRef")
	}
}

func validateUsageMeters(missing *[]string, field string, meters []UsageMeter) {
	for i, meter := range meters {
		prefix := fmt.Sprintf("%s[%d]", field, i)
		require(missing, meter.Name != "", prefix+".name")
		require(missing, meter.Unit != "", prefix+".unit")
	}
}

func validatePortalModules(missing *[]string, modules []PortalModule) {
	require(missing, len(modules) > 0, "portalModules")
	for i, module := range modules {
		prefix := fmt.Sprintf("portalModules[%d]", i)
		require(missing, module.Name != "", prefix+".name")
		require(missing, module.Slot != "", prefix+".slot")
		require(missing, module.Route != "", prefix+".route")
		require(missing, module.APIRef != "", prefix+".apiRef")
		require(missing, module.HostRef != "", prefix+".hostRef")
		require(missing, module.MountRef != "", prefix+".mountRef")
		require(missing, len(module.Permissions) > 0, prefix+".permissions")
	}
}
