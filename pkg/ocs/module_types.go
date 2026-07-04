package ocs

type PortalModule struct {
	Name        string   `json:"name"`
	Slot        string   `json:"slot"`
	Route       string   `json:"route"`
	APIRef      string   `json:"apiRef"`
	HostRef     string   `json:"hostRef"`
	MountRef    string   `json:"mountRef"`
	Permissions []string `json:"permissions"`
}

type UIExtensionManifest struct {
	EmbedRef         string                    `json:"embedRef"`
	ContextSchemaRef string                    `json:"contextSchemaRef"`
	HostAuthority    []string                  `json:"hostAuthority"`
	ExtensionActions []string                  `json:"extensionActions"`
	ModuleHost       MicrofrontendHostContract `json:"moduleHost"`
	Evidence         []EvidenceRef             `json:"evidence"`
}

type MicrofrontendHostContract struct {
	Host            string   `json:"host"`
	Runtime         string   `json:"runtime"`
	MountRef        string   `json:"mountRef"`
	VersionRange    string   `json:"versionRange"`
	IntegrityRef    string   `json:"integrityRef"`
	Sandbox         string   `json:"sandbox"`
	AllowedEvents   []string `json:"allowedEvents"`
	RequiredContext []string `json:"requiredContext"`
}

type AnalyticsEvent struct {
	Name        string   `json:"name"`
	Trigger     string   `json:"trigger"`
	Subject     string   `json:"subject"`
	Properties  []string `json:"properties"`
	EvidenceRef string   `json:"evidenceRef"`
}
