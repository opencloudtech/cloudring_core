// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package iam

import (
	"sync"
	"time"
)

type AuditEvent struct {
	Actor              string
	RepresentedSubject string
	Action             Action
	Target             TargetRef
	Reason             string
	TicketRef          string
	CorrelationID      string
	Result             AuditResult
	PolicyRule         string
	Timestamp          time.Time
	Error              string
	BreakGlass         bool
	SupportGrantRef    string
	APITokenRef        string
	TargetPrincipal    string
	TargetLifecycle    TenantState
	TargetNamespace    string
	TargetResource     string
	TargetResourceName string
	TargetOrganization string
	TargetTenant       string
	TargetProject      string
	CredentialClass    CredentialClass
	MFA                MFAAssurance
	Session            SessionAssurance
}

type AuditSink interface {
	Append(AuditEvent) error
	Events() []AuditEvent
}

type MemoryAuditSink struct {
	mu     sync.RWMutex
	events []AuditEvent
}

func NewMemoryAuditSink() *MemoryAuditSink {
	return &MemoryAuditSink{events: []AuditEvent{}}
}

func (sink *MemoryAuditSink) Append(event AuditEvent) error {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	sink.events = append(sink.events, event)
	return nil
}

func (sink *MemoryAuditSink) Events() []AuditEvent {
	sink.mu.RLock()
	defer sink.mu.RUnlock()
	return append([]AuditEvent{}, sink.events...)
}

type FailingAuditSink struct{}

func (FailingAuditSink) Append(AuditEvent) error {
	return ErrAuditRequired
}

func (FailingAuditSink) Events() []AuditEvent {
	return nil
}
