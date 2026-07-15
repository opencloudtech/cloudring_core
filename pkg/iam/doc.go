// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

// Package iam implements provider-neutral, fail-closed authorization decisions
// for CloudRING organizations, tenants, projects, support grants, API-token
// references, break-glass controls, and audit sinks.
//
// A subject or API-token reference never authenticates itself. Policy denies
// unless its configured AuthenticationVerifier returns matching trusted
// authentication, MFA, session, and credential-class evidence.
//
// A Policy is configured by populating its exported maps before it starts
// serving requests. Once serving begins, callers must treat that configuration
// as immutable. Authorize and AuditEvents may then be called concurrently when
// the configured Clock, AuthenticationVerifier, and AuditSink are themselves
// safe for concurrent use. The built-in clocks and MemoryAuditSink satisfy
// their parts of that contract. Authorize samples its clock once and uses that
// instant for authentication, evaluation, and audit.
package iam
