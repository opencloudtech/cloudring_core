package ocsv3

import (
	"encoding/json"
	"fmt"

	"github.com/opencloudtech/CloudRING/pkg/ocs"
)

const APIVersion = ocs.APIVersion

const (
	ExecutionProfileLocal      = ocs.ExecutionProfileLocal
	ExecutionProfileRemote     = ocs.ExecutionProfileRemote
	ExecutionProfileAPIOnly    = ocs.ExecutionProfileAPIOnly
	ApplicabilitySupported     = ocs.ApplicabilitySupported
	ApplicabilityNotApplicable = ocs.ApplicabilityNotApplicable
)

type Metadata = ocs.Metadata
type ConnectorPackage = ocs.ConnectorPackage
type ServiceConnector = ocs.ServiceConnector
type ServiceSpec = ocs.ServiceSpec
type ProductAPIContract = ocs.ProductAPIContract
type Capability = ocs.Capability
type Dependency = ocs.Dependency
type LifecycleAction = ocs.LifecycleAction
type AutomationTask = ocs.AutomationTask
type UsageMeter = ocs.UsageMeter
type BillingProfile = ocs.BillingProfile
type KubernetesBinding = ocs.KubernetesBinding
type GatewayRoute = ocs.GatewayRoute
type SecretBoundary = ocs.SecretBoundary
type SecretRef = ocs.SecretRef
type PolicyRule = ocs.PolicyRule
type DataLifecycle = ocs.DataLifecycle
type DataLifecycleAction = ocs.DataLifecycleAction
type ServiceState = ocs.ServiceState
type SupportProfile = ocs.SupportProfile
type CatalogPublication = ocs.CatalogPublication
type CatalogPlan = ocs.CatalogPlan
type ConfigurationSchema = ocs.ConfigurationSchema
type ReadinessCheck = ocs.ReadinessCheck
type TenantAccessPolicy = ocs.TenantAccessPolicy
type TenantEntitlement = ocs.TenantEntitlement
type DurabilityProfile = ocs.DurabilityProfile
type EvidenceBundle = ocs.EvidenceBundle
type EvidenceRef = ocs.EvidenceRef
type DistributionProfile = ocs.DistributionProfile
type FederationProfile = ocs.FederationProfile
type CommercialProfile = ocs.CommercialProfile
type BillingConnector = ocs.BillingConnector
type CostMeter = ocs.CostMeter
type BillingEvent = ocs.BillingEvent
type PortalModule = ocs.PortalModule
type UIExtensionManifest = ocs.UIExtensionManifest
type MicrofrontendHostContract = ocs.MicrofrontendHostContract
type AnalyticsEvent = ocs.AnalyticsEvent

func ParseConnectorPackage(data []byte) (ConnectorPackage, error) {
	var pkg ConnectorPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ConnectorPackage{}, fmt.Errorf("parse connector package: %w", err)
	}
	return pkg, nil
}

func ValidateConnectorPackageBytes(data []byte) error {
	pkg, err := ParseConnectorPackage(data)
	if err != nil {
		return err
	}
	if err := pkg.Validate(); err != nil {
		return fmt.Errorf("validate connector package: %w", err)
	}
	return nil
}
