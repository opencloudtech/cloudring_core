package controller

import (
	"fmt"
	"strings"
	"time"
)

type Phase string

const (
	PhasePending      Phase = "pending"
	PhaseProvisioning Phase = "provisioning"
	PhaseReady        Phase = "ready"
	PhaseDenied       Phase = "denied"
	PhaseDegraded     Phase = "degraded"
	PhaseRetryable    Phase = "retryable"
	PhaseDeleting     Phase = "deleting"
	PhaseFailed       Phase = "failed"
)

type Claim struct {
	Name       string
	Namespace  string
	ProjectRef string
	Plan       string
}

type Status struct {
	Phase       Phase
	EvidenceRef string
	Reason      string
	NextAction  string
}

type Receipt struct {
	ID             string    `json:"id"`
	Action         string    `json:"action"`
	ClaimRef       string    `json:"claimRef"`
	IdempotencyKey string    `json:"idempotencyKey"`
	GeneratedAt    time.Time `json:"generatedAt"`
}

type BillingEvent struct {
	Name           string `json:"name"`
	Meter          string `json:"meter"`
	IdempotencyKey string `json:"idempotencyKey"`
	Subject        string `json:"subject"`
}

func Reconcile(claim Claim, action string, now time.Time) (Status, Receipt, BillingEvent, error) {
	if err := validateClaim(claim); err != nil {
		return Status{Phase: PhaseDenied, Reason: err.Error(), NextAction: "fix claim and retry"}, Receipt{}, BillingEvent{}, err
	}
	status := statusForAction(action)
	ref := claim.Namespace + "/" + claim.Name
	receipt := Receipt{
		ID:             action + ":" + ref,
		Action:         action,
		ClaimRef:       ref,
		IdempotencyKey: action + ":" + claim.ProjectRef + ":" + claim.Name,
		GeneratedAt:    now.UTC(),
	}
	event := BillingEvent{
		Name:           "synthetic-service-usage-recorded",
		Meter:          "synthetic_units",
		IdempotencyKey: receipt.IdempotencyKey,
		Subject:        ref,
	}
	return status, receipt, event, nil
}

func validateClaim(claim Claim) error {
	if strings.TrimSpace(claim.Name) == "" {
		return fmt.Errorf("claim name is required")
	}
	if strings.TrimSpace(claim.Namespace) == "" {
		return fmt.Errorf("claim namespace is required")
	}
	if strings.TrimSpace(claim.ProjectRef) == "" {
		return fmt.Errorf("projectRef is required")
	}
	if strings.TrimSpace(claim.Plan) == "" {
		return fmt.Errorf("plan is required")
	}
	return nil
}

func statusForAction(action string) Status {
	switch action {
	case "provision", "restore", "retry", "rollback":
		return Status{Phase: PhaseReady, EvidenceRef: "evidence.synthetic-service.ready", NextAction: "continue normal operation"}
	case "backup":
		return Status{Phase: PhaseReady, EvidenceRef: "evidence.synthetic-service.backup", NextAction: "retain backup receipt"}
	case "export":
		return Status{Phase: PhaseReady, EvidenceRef: "evidence.synthetic-service.export", NextAction: "download portable archive"}
	case "delete":
		return Status{Phase: PhaseDeleting, EvidenceRef: "evidence.synthetic-service.delete", NextAction: "wait for finalizer cleanup"}
	default:
		return Status{Phase: PhaseRetryable, EvidenceRef: "evidence.synthetic-service.retryable", Reason: "unknown action", NextAction: "retry with supported lifecycle action"}
	}
}
