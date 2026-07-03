package ocs

type DistributionProfile struct {
	DeploymentProfiles    []string `json:"deploymentProfiles"`
	Channels              []string `json:"channels"`
	InfrastructureTargets []string `json:"infrastructureTargets"`
	UpdatePolicyRef       string   `json:"updatePolicyRef"`
}

type FederationProfile struct {
	Modes                  []string `json:"modes"`
	MessageBusRef          string   `json:"messageBusRef"`
	CrossProviderScenarios []string `json:"crossProviderScenarios"`
	PortabilityPolicyRef   string   `json:"portabilityPolicyRef"`
}

type CommercialProfile struct {
	Roles                   []string `json:"roles"`
	RevenueModel            string   `json:"revenueModel"`
	LicenseRef              string   `json:"licenseRef"`
	ExpiryBehavior          string   `json:"expiryBehavior"`
	SupportRef              string   `json:"supportRef"`
	ServiceProvenance       string   `json:"serviceProvenance"`
	ResponsibilityMatrixRef string   `json:"responsibilityMatrixRef"`
	ContinuityPlanRef       string   `json:"continuityPlanRef"`
}
