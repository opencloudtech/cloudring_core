package ocs

import (
	"fmt"
	"strings"
)

func (p ConnectorPackage) Validate() error {
	var missing []string
	var invalid []string

	require(&missing, p.APIVersion == APIVersion, "apiVersion")
	require(&missing, p.Kind == "ConnectorPackage", "kind")
	requireMetadata(&missing, p.Metadata)

	if err := p.Service.Validate(); err != nil {
		invalid = append(invalid, "service: "+err.Error())
	}
	if err := p.Billing.Validate(); err != nil {
		invalid = append(invalid, "billing: "+err.Error())
	}

	require(&missing, p.Catalog.ServiceClass != "", "catalog.serviceClass")
	require(&missing, p.Catalog.Visibility != "", "catalog.visibility")
	require(&missing, p.Catalog.Portability != "", "catalog.portability")
	require(&missing, len(p.Catalog.Plans) > 0, "catalog.plans")
	for i, plan := range p.Catalog.Plans {
		prefix := fmt.Sprintf("catalog.plans[%d]", i)
		require(&missing, plan.Name != "", prefix+".name")
		require(&missing, plan.DisplayName != "", prefix+".displayName")
		require(&missing, plan.BillingMode != "", prefix+".billingMode")
	}

	require(&missing, p.Configuration.SchemaRef != "", "configuration.schemaRef")
	require(&missing, p.Configuration.Version != "", "configuration.version")
	require(&missing, len(p.Readiness) > 0, "readiness")
	for i, check := range p.Readiness {
		prefix := fmt.Sprintf("readiness[%d]", i)
		require(&missing, check.Name != "", prefix+".name")
		require(&missing, check.Type != "", prefix+".type")
		require(&missing, check.Target != "", prefix+".target")
		require(&missing, check.Condition != "", prefix+".condition")
		require(&missing, check.EvidenceRef != "", prefix+".evidenceRef")
	}

	require(&missing, p.TenantAccess.Scope != "", "tenantAccess.scope")
	require(&missing, p.TenantAccess.EntitlementRef != "", "tenantAccess.entitlementRef")
	require(&missing, len(p.TenantAccess.Entitlements) > 0, "tenantAccess.entitlements")
	require(&missing, len(p.TenantAccess.Permissions) > 0, "tenantAccess.permissions")
	for i, entitlement := range p.TenantAccess.Entitlements {
		prefix := fmt.Sprintf("tenantAccess.entitlements[%d]", i)
		require(&missing, entitlement.Name != "", prefix+".name")
		require(&missing, entitlement.Ref != "", prefix+".ref")
		require(&missing, entitlement.Scope != "", prefix+".scope")
	}
	require(&missing, p.Durability.StateClass != "", "durability.stateClass")
	require(&missing, len(p.Durability.DataClasses) > 0, "durability.dataClasses")
	require(&missing, p.Durability.RecoveryObjective != "", "durability.recoveryObjective")
	require(&missing, len(p.Durability.RecoveryEvidence) > 0, "durability.recoveryEvidence")
	validateEvidenceRefs(&missing, "durability.recoveryEvidence", p.Durability.RecoveryEvidence)

	if p.Service.Spec.Billing.ConnectorRef != "" && p.Billing.Metadata.Name != "" && p.Service.Spec.Billing.ConnectorRef != p.Billing.Metadata.Name {
		invalid = append(invalid, "service.billing.connectorRef must match billing.metadata.name")
	}

	if len(p.Service.Spec.Capabilities) > 0 && p.Catalog.ServiceClass != "" && !capabilityClassExists(p.Service.Spec.Capabilities, p.Catalog.ServiceClass) {
		invalid = append(invalid, "catalog.serviceClass must reference a declared capability class")
	}

	validateAutomationRefs(&missing, &invalid, p.Service.Spec.Lifecycle, p.Service.Spec.Automation)
	validateBillingMeters(&missing, &invalid, p.Service.Spec.UsageMeters, p.Service.Spec.Billing.Meters, p.Billing.Meters)
	validateAlpha2Package(&missing, &invalid, p)
	validateOCSv3Package(&missing, p)

	for i, dependency := range p.Service.Spec.Dependencies {
		if isNonPortable(dependency.Portability) {
			invalid = append(invalid, fmt.Sprintf("service.spec.dependencies[%d].portability must describe a portable contract", i))
		}
	}

	if len(missing) > 0 || len(invalid) > 0 {
		problems := append([]string{}, missing...)
		problems = append(problems, invalid...)
		return fmt.Errorf("connector package invalid: %s", strings.Join(problems, ", "))
	}
	return nil
}

func requireMetadata(missing *[]string, metadata Metadata) {
	require(missing, metadata.Name != "", "metadata.name")
	require(missing, metadata.DisplayName != "", "metadata.displayName")
	require(missing, metadata.Owner != "", "metadata.owner")
	require(missing, metadata.Version != "", "metadata.version")
}

func require(missing *[]string, ok bool, name string) {
	if !ok {
		*missing = append(*missing, name)
	}
}

func validateEvidenceRefs(missing *[]string, field string, refs []EvidenceRef) {
	for i, ref := range refs {
		prefix := fmt.Sprintf("%s[%d]", field, i)
		require(missing, ref.Name != "", prefix+".name")
		require(missing, ref.Type != "", prefix+".type")
		require(missing, ref.Ref != "", prefix+".ref")
	}
}

func capabilityClassExists(capabilities []Capability, class string) bool {
	for _, capability := range capabilities {
		if capability.Class == class {
			return true
		}
	}
	return false
}

func validateAutomationRefs(missing *[]string, invalid *[]string, lifecycle []LifecycleAction, tasks []AutomationTask) {
	refs := map[string]bool{}
	for _, action := range lifecycle {
		if action.Name == "" {
			continue
		}
		refs["lifecycle."+action.Name] = true
		refs[action.Name] = true
	}

	for i, task := range tasks {
		if task.ActionRef == "" {
			continue
		}
		if !refs[task.ActionRef] {
			*invalid = append(*invalid, fmt.Sprintf("service.spec.automation[%d].actionRef must reference a lifecycle action", i))
		}
	}
}

func isNonPortable(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "none", "implementation-specific", "vendor-specific", "platform-internal":
		return true
	default:
		return false
	}
}
