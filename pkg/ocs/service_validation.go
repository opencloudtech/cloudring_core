package ocs

import (
	"fmt"
	"strings"
)

func (c ServiceConnector) Validate() error {
	var missing []string
	var invalid []string

	require(&missing, c.APIVersion == APIVersion, "apiVersion")
	require(&missing, c.Kind == "ServiceConnector", "kind")
	requireMetadata(&missing, c.Metadata)
	require(&missing, len(c.Spec.Capabilities) > 0, "capabilities")
	require(&missing, len(c.Spec.Lifecycle) > 0, "lifecycle")
	require(&missing, len(c.Spec.Automation) > 0, "automation")
	require(&missing, len(c.Spec.UsageMeters) > 0, "usageMeters")
	require(&missing, c.Spec.Billing.ConnectorRef != "", "billing.connectorRef")
	require(&missing, len(c.Spec.Billing.Meters) > 0, "billing.meters")
	require(&missing, len(c.Spec.PortalModules) > 0, "portalModules")
	require(&missing, len(c.Spec.KubernetesBindings) > 0, "kubernetesBindings")
	require(&missing, c.Spec.Support.Owner != "", "support.owner")
	require(&missing, len(c.Spec.Support.Diagnostics) > 0, "support.diagnostics")

	validateCapabilities(&missing, c.Spec.Capabilities)
	validateDependencies(&missing, c.Spec.Dependencies)
	validateKubernetesBindings(&missing, c.Spec.KubernetesBindings)
	validateLifecycle(&missing, c.Spec.Lifecycle)
	validateAutomationTasks(&missing, c.Spec.Automation)
	validateUsageMeters(&missing, "usageMeters", c.Spec.UsageMeters)
	validateUsageMeters(&missing, "billing.meters", c.Spec.Billing.Meters)
	validatePortalModules(&missing, c.Spec.PortalModules)
	validateAlpha2ServiceSpec(&missing, &invalid, c.Spec)
	validateOCSv3ServiceSpec(&missing, c.Spec)

	if len(missing) > 0 || len(invalid) > 0 {
		problems := append([]string{}, missing...)
		problems = append(problems, invalid...)
		return fmt.Errorf("service connector invalid: %s", strings.Join(problems, ", "))
	}
	return nil
}

func validateCapabilities(missing *[]string, capabilities []Capability) {
	for i, capability := range capabilities {
		prefix := fmt.Sprintf("capabilities[%d]", i)
		require(missing, capability.Class != "", prefix+".class")
		require(missing, capability.Name != "", prefix+".name")
	}
}

func validateDependencies(missing *[]string, dependencies []Dependency) {
	for i, dependency := range dependencies {
		prefix := fmt.Sprintf("dependencies[%d]", i)
		require(missing, dependency.ID != "", prefix+".id")
		require(missing, dependency.CapabilityClass != "", prefix+".capabilityClass")
		require(missing, dependency.Role != "", prefix+".role")
		require(missing, dependency.Portability != "", prefix+".portability")
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

func validateLifecycle(missing *[]string, lifecycle []LifecycleAction) {
	for i, action := range lifecycle {
		prefix := fmt.Sprintf("lifecycle[%d]", i)
		require(missing, action.Name != "", prefix+".name")
		require(missing, action.Verb != "", prefix+".verb")
		require(missing, action.Idempotent, prefix+".idempotent")
		require(missing, action.IdempotencyKey != "", prefix+".idempotencyKey")
	}
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
