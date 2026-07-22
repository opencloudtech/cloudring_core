// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

// Package gitopsownership defines a provider-neutral, fail-closed contract for
// proving that an accepted Flux source artifact and its exact Kustomization
// roots uniquely own the declared critical resource families. It validates
// evidence only; collectors, provider bindings, and live mutations remain the
// responsibility of downstream installations.
package gitopsownership
