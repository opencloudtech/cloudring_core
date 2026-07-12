// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package ocs

import (
	"strings"
	"testing"
)

func Test_ConnectorPackageValidate_rejects_missing_microfrontend_host_contract(t *testing.T) {
	pkg := validConnectorPackage()
	pkg.Service.Spec.UI.ModuleHost = MicrofrontendHostContract{}

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected missing microfrontend host contract to fail")
	}
	for _, want := range []string{
		"ui.moduleHost.host",
		"ui.moduleHost.runtime",
		"ui.moduleHost.mountRef",
		"ui.moduleHost.integrityRef",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q in error %q", want, err.Error())
		}
	}
}

func Test_ConnectorPackageValidate_rejects_missing_module_analytics_events(t *testing.T) {
	pkg := validConnectorPackage()
	pkg.Service.Spec.AnalyticsEvents = nil

	err := pkg.Validate()
	if err == nil {
		t.Fatal("expected missing analytics events to fail")
	}
	if !strings.Contains(err.Error(), "analyticsEvents") {
		t.Fatalf("expected analyticsEvents in error %q", err.Error())
	}
}

func Test_ServiceConnectorValidate_rejects_portal_module_without_host_binding(t *testing.T) {
	connector := validServiceConnector()
	connector.Spec.PortalModules[0].HostRef = ""
	connector.Spec.PortalModules[0].MountRef = ""

	err := connector.Validate()
	if err == nil {
		t.Fatal("expected missing portal module host binding to fail")
	}
	for _, want := range []string{
		"portalModules[0].hostRef",
		"portalModules[0].mountRef",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q in error %q", want, err.Error())
		}
	}
}
