// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package siteprofilecontract

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

func TestExampleConformsToMachineReadableSchema(t *testing.T) {
	schema, err := jsonschema.NewCompiler().Compile("provider-site-profile.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded := exampleJSON(t)
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if err := schema.Validate(instance); err != nil {
		t.Fatalf("example does not conform: %v", err)
	}
}

func TestSchemaRejectsReadinessOverclaim(t *testing.T) {
	var document map[string]any
	if err := json.Unmarshal(exampleJSON(t), &document); err != nil {
		t.Fatal(err)
	}
	document["spec"].(map[string]any)["nonClaim"] = "production-ready"
	if schemaAccepts(t, document) {
		t.Fatal("readiness overclaim passed schema validation")
	}
}

func TestSchemaRejectsInventoryWithoutBaselineControlPlaneAndWorkerRoles(t *testing.T) {
	var document map[string]any
	if err := json.Unmarshal(exampleJSON(t), &document); err != nil {
		t.Fatal(err)
	}
	nodes := document["spec"].(map[string]any)["inventory"].(map[string]any)["nodes"].([]any)
	for _, rawNode := range nodes {
		rawNode.(map[string]any)["roles"] = []any{"gateway"}
	}
	if schemaAccepts(t, document) {
		t.Fatal("gateway-only inventory passed baseline role validation")
	}
}

func schemaAccepts(t *testing.T, document any) bool {
	t.Helper()
	schema, err := jsonschema.NewCompiler().Compile("provider-site-profile.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	return schema.Validate(instance) == nil
}

func exampleJSON(t *testing.T) []byte {
	t.Helper()
	path := filepath.Join("..", "..", "examples", "provider-site-profile.yaml")
	// #nosec G304 -- the test reads the repository-owned example at a fixed relative path.
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	decoder := yaml.NewDecoder(bytes.NewReader(payload))
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
