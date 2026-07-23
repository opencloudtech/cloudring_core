package ocs

const APIVersion = "ocsv3.cloudring.io/v1alpha1"

const (
	ExecutionProfileLocal   = "local"
	ExecutionProfileRemote  = "remote"
	ExecutionProfileAPIOnly = "api-only"

	ApplicabilitySupported     = "supported"
	ApplicabilityNotApplicable = "not_applicable"
)

type Metadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Owner       string `json:"owner"`
	Version     string `json:"version"`
}

type ConnectorPackage struct {
	APIVersion    string              `json:"apiVersion"`
	Kind          string              `json:"kind"`
	Metadata      Metadata            `json:"metadata"`
	Service       ServiceConnector    `json:"service"`
	Billing       BillingConnector    `json:"billing"`
	Catalog       CatalogPublication  `json:"catalog"`
	Configuration ConfigurationSchema `json:"configuration"`
	Readiness     []ReadinessCheck    `json:"readiness"`
	TenantAccess  TenantAccessPolicy  `json:"tenantAccess"`
	Durability    DurabilityProfile   `json:"durability"`
	Distribution  DistributionProfile `json:"distribution"`
	Federation    FederationProfile   `json:"federation"`
	Commercial    CommercialProfile   `json:"commercial"`
}

type ServiceConnector struct {
	APIVersion string      `json:"apiVersion"`
	Kind       string      `json:"kind"`
	Metadata   Metadata    `json:"metadata"`
	Spec       ServiceSpec `json:"spec"`
}

type ServiceSpec struct {
	ExecutionProfile   string              `json:"executionProfile"`
	ProductAPI         ProductAPIContract  `json:"productAPI"`
	Capabilities       []Capability        `json:"capabilities"`
	Dependencies       []Dependency        `json:"dependencies"`
	Lifecycle          []LifecycleAction   `json:"lifecycle"`
	Automation         []AutomationTask    `json:"automation"`
	UsageMeters        []UsageMeter        `json:"usageMeters"`
	Billing            BillingProfile      `json:"billing"`
	PortalModules      []PortalModule      `json:"portalModules"`
	UI                 UIExtensionManifest `json:"ui"`
	AnalyticsEvents    []AnalyticsEvent    `json:"analyticsEvents"`
	KubernetesBindings []KubernetesBinding `json:"kubernetesBindings"`
	GatewayRoutes      []GatewayRoute      `json:"gatewayRoutes"`
	Secrets            SecretBoundary      `json:"secrets"`
	Policies           []PolicyRule        `json:"policies"`
	DataLifecycle      DataLifecycle       `json:"dataLifecycle"`
	States             []ServiceState      `json:"states"`
	Support            SupportProfile      `json:"support"`
	EvidenceBundles    []EvidenceBundle    `json:"evidenceBundles"`
}

type ProductAPIContract struct {
	Ref            string `json:"ref"`
	Version        string `json:"version"`
	Protocol       string `json:"protocol"`
	EndpointRef    string `json:"endpointRef,omitempty"`
	TrustPolicyRef string `json:"trustPolicyRef,omitempty"`
	HealthRef      string `json:"healthRef,omitempty"`
}

type Capability struct {
	Class       string `json:"class"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Dependency struct {
	ID                     string `json:"id"`
	CapabilityClass        string `json:"capabilityClass"`
	Role                   string `json:"role"`
	Portability            string `json:"portability"`
	ProductAPIRef          string `json:"productAPIRef"`
	VersionRange           string `json:"versionRange"`
	CompatibilityPolicyRef string `json:"compatibilityPolicyRef"`
	ImplementationRef      string `json:"implementationRef,omitempty"`
}

type LifecycleAction struct {
	Name           string `json:"name"`
	Applicability  string `json:"applicability"`
	Reason         string `json:"reason,omitempty"`
	Verb           string `json:"verb"`
	Idempotent     bool   `json:"idempotent"`
	IdempotencyKey string `json:"idempotencyKey"`
	RollbackRef    string `json:"rollbackRef,omitempty"`
}

type AutomationTask struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ActionRef   string `json:"actionRef"`
}

type UsageMeter struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Aggregation string `json:"aggregation,omitempty"`
}

type BillingProfile struct {
	Applicability string       `json:"applicability"`
	Reason        string       `json:"reason,omitempty"`
	ConnectorRef  string       `json:"connectorRef"`
	Meters        []UsageMeter `json:"meters"`
}

type KubernetesBinding struct {
	Group      string `json:"group"`
	Version    string `json:"version"`
	Kind       string `json:"kind"`
	Plural     string `json:"plural"`
	Scope      string `json:"scope"`
	CRDRef     string `json:"crdRef"`
	StatusPath string `json:"statusPath"`
	Condition  string `json:"condition"`
}

type GatewayRoute struct {
	Name        string   `json:"name"`
	ParentRef   string   `json:"parentRef"`
	Hostnames   []string `json:"hostnames"`
	Rules       []string `json:"rules"`
	EvidenceRef string   `json:"evidenceRef"`
}

type SecretBoundary struct {
	WorkloadIdentityRef string      `json:"workloadIdentityRef"`
	SecretRefs          []SecretRef `json:"secretRefs"`
}

type SecretRef struct {
	Name     string `json:"name"`
	Ref      string `json:"ref"`
	Purpose  string `json:"purpose"`
	Rotation string `json:"rotation"`
	RawValue string `json:"rawValue,omitempty"`
}

type PolicyRule struct {
	Name        string `json:"name"`
	Class       string `json:"class"`
	DecisionRef string `json:"decisionRef"`
	EvidenceRef string `json:"evidenceRef"`
}

type DataLifecycle struct {
	Export DataLifecycleAction `json:"export"`
	Delete DataLifecycleAction `json:"delete"`
}

type DataLifecycleAction struct {
	ActionRef   string `json:"actionRef"`
	Format      string `json:"format"`
	EvidenceRef string `json:"evidenceRef"`
}

type ServiceState struct {
	Name        string `json:"name"`
	Reason      string `json:"reason"`
	UserVisible bool   `json:"userVisible"`
	EvidenceRef string `json:"evidenceRef"`
	Remediation string `json:"remediation"`
}

type SupportProfile struct {
	Owner       string        `json:"owner"`
	Escalation  string        `json:"escalation"`
	Diagnostics []string      `json:"diagnostics"`
	DocsRef     string        `json:"docsRef"`
	Evidence    []EvidenceRef `json:"evidence"`
}

type CatalogPublication struct {
	ServiceClass string        `json:"serviceClass"`
	Visibility   string        `json:"visibility"`
	Portability  string        `json:"portability"`
	Plans        []CatalogPlan `json:"plans"`
}

type CatalogPlan struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	BillingMode string `json:"billingMode"`
	QuotaRef    string `json:"quotaRef,omitempty"`
}

type ConfigurationSchema struct {
	SchemaRef string `json:"schemaRef"`
	Version   string `json:"version"`
}

type ReadinessCheck struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	Condition   string `json:"condition"`
	EvidenceRef string `json:"evidenceRef"`
}

type TenantAccessPolicy struct {
	Scope          string              `json:"scope"`
	EntitlementRef string              `json:"entitlementRef"`
	Entitlements   []TenantEntitlement `json:"entitlements"`
	Permissions    []string            `json:"permissions"`
}

type TenantEntitlement struct {
	Name  string `json:"name"`
	Ref   string `json:"ref"`
	Scope string `json:"scope"`
}

type DurabilityProfile struct {
	StateClass        string        `json:"stateClass"`
	DataClasses       []string      `json:"dataClasses"`
	BackupPolicyRef   string        `json:"backupPolicyRef,omitempty"`
	RecoveryObjective string        `json:"recoveryObjective"`
	RecoveryEvidence  []EvidenceRef `json:"recoveryEvidence"`
}

type EvidenceBundle struct {
	Name       string        `json:"name"`
	Owner      string        `json:"owner"`
	Claim      string        `json:"claim"`
	Freshness  string        `json:"freshness"`
	Redaction  string        `json:"redaction"`
	Evidence   []EvidenceRef `json:"evidence"`
	NonClaims  []string      `json:"nonClaims"`
	ReviewPath string        `json:"reviewPath"`
}

type EvidenceRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Ref  string `json:"ref"`
}
