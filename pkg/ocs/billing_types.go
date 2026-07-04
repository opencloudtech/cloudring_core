package ocs

type BillingConnector struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   Metadata       `json:"metadata"`
	Meters     []UsageMeter   `json:"meters"`
	CostMeters []CostMeter    `json:"costMeters"`
	Events     []BillingEvent `json:"events"`
}

type CostMeter struct {
	Name        string `json:"name"`
	Currency    string `json:"currency"`
	UnitPrice   string `json:"unitPrice"`
	MeterRef    string `json:"meterRef"`
	EvidenceRef string `json:"evidenceRef"`
}

type BillingEvent struct {
	Name           string `json:"name"`
	Meter          string `json:"meter"`
	Idempotent     bool   `json:"idempotent"`
	IdempotencyKey string `json:"idempotencyKey"`
	EntitlementRef string `json:"entitlementRef"`
	Attribution    string `json:"attribution"`
	ReplayPolicy   string `json:"replayPolicy"`
}
