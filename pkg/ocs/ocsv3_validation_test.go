// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package ocs

import (
	"strings"
	"testing"
)

func Test_ConnectorPackageValidate_rejects_missing_distribution_and_applicability_profiles(t *testing.T) {
	// Given
	pkg := validConnectorPackage()
	pkg.Distribution = DistributionProfile{}
	pkg.Federation = FederationProfile{}
	pkg.Commercial = CommercialProfile{}

	// When
	err := pkg.Validate()

	// Then
	if err == nil {
		t.Fatal("expected missing OCSv3 package profiles to fail")
	}
	for _, want := range []string{
		"distribution.deploymentProfiles",
		"distribution.channels",
		"federation.applicability",
		"commercial.applicability",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("expected %q in error %q", want, err.Error())
		}
	}
}

func Test_ConnectorPackageValidate_accepts_ocsv3_package_profiles(t *testing.T) {
	// Given
	pkg := validConnectorPackage()

	// When
	err := pkg.Validate()

	// Then
	if err != nil {
		t.Fatalf("expected OCSv3 package profiles to validate: %v", err)
	}
}
