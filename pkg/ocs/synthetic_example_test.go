// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package ocs

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func Test_SyntheticConnectorPackageExample_validates(t *testing.T) {
	raw, err := os.ReadFile("../../examples/synthetic-service-module/connector-package.json")
	if err != nil {
		t.Fatalf("read synthetic connector package example: %v", err)
	}
	var pkg ConnectorPackage
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatalf("unmarshal synthetic connector package example: %v", err)
	}
	if err := pkg.Validate(); err != nil {
		t.Fatalf("expected synthetic connector package to validate: %v", err)
	}
}

func Test_SyntheticInvalidImplementationReference_is_rejected(t *testing.T) {
	raw, err := os.ReadFile("testdata/invalid-implementation-reference.json")
	if err != nil {
		t.Fatalf("read invalid implementation-reference fixture: %v", err)
	}
	var pkg ConnectorPackage
	if err := json.Unmarshal(raw, &pkg); err != nil {
		t.Fatalf("unmarshal invalid implementation-reference fixture: %v", err)
	}

	err = pkg.Validate()
	if err == nil {
		t.Fatal("expected implementation-specific dependency reference to fail")
	}
	want := "class=coupling owner=service path=service.spec.dependencies[0].implementationRef"
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("expected %q in validation error %q", want, err.Error())
	}
}
