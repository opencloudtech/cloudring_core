// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package backupproof_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
	"github.com/opencloudtech/CloudRING/pkg/backup/velero118"
)

func TestBackupProofContractsAreStrictJSONAndFixtureMatchesRuntime(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"baseline-request.schema.json", "collection-request.schema.json", "adapter-protocol.schema.json", "cleanup-ready.schema.json", "restore-proof.schema.json"} {
		// #nosec G304 -- paths are a closed repository-owned fixture list.
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := strictjson.Decode(data, &schema); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
			t.Fatalf("%s: unexpected JSON Schema dialect", path)
		}
	}
	for fixture, schemaPath := range map[string]string{
		"fixtures/synthetic-baseline-request.json":   "baseline-request.schema.json",
		"fixtures/synthetic-collection-request.json": "collection-request.schema.json",
		"fixtures/synthetic-cleanup-ready.json":      "cleanup-ready.schema.json",
	} {
		schema, err := jsonschema.NewCompiler().Compile(schemaPath)
		if err != nil {
			t.Fatalf("compile %s: %v", schemaPath, err)
		}
		// #nosec G304 -- fixture and schema paths are a closed repository-owned map.
		payload, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatal(err)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(instance); err != nil {
			t.Fatalf("%s does not match %s: %v", fixture, schemaPath, err)
		}
	}
	fixture, err := os.ReadFile("fixtures/synthetic-collection-request.json")
	if err != nil {
		t.Fatal(err)
	}
	var request velero118.CollectionRequest
	if err := strictjson.Decode(fixture, &request); err != nil {
		t.Fatal(err)
	}
	if request.SchemaVersion != velero118.CollectionRequestSchemaVersion || request.SourceNamespace == request.TargetNamespace || request.SourcePVC != request.TargetPVC ||
		request.ServerStatusRequestName == "" || request.ServerStatusRequestUIDSHA256 == "" || request.CleanupRunNonceSHA256 == "" {
		t.Fatal("synthetic collection request does not match the runtime contract")
	}
	baselineFixture, err := os.ReadFile("fixtures/synthetic-baseline-request.json") // #nosec G304 -- repository-owned test fixture path.
	if err != nil {
		t.Fatal(err)
	}
	var baselineRequest velero118.BaselineRequest
	if err := strictjson.DecodeExact(baselineFixture, &baselineRequest); err != nil || baselineRequest.SchemaVersion != velero118.BaselineRequestSchemaVersion || baselineRequest.SourceNamespace == "" || baselineRequest.SourcePVC == "" {
		t.Fatal("synthetic baseline request does not match the pre-restore runtime contract")
	}
	var fullRequestAsBaseline velero118.BaselineRequest
	if err := strictjson.DecodeExact(fixture, &fullRequestAsBaseline); err == nil {
		t.Fatal("full collection request unexpectedly passed the baseline runtime contract")
	}
	// #nosec G304 -- repository-owned test fixture path.
	readyFixture, err := os.ReadFile("fixtures/synthetic-cleanup-ready.json")
	if err != nil {
		t.Fatal(err)
	}
	var ready velero118.CleanupReady
	if err := strictjson.DecodeExact(readyFixture, &ready); err != nil || ready.SchemaVersion != velero118.CleanupReadySchemaVersion ||
		ready.Status != velero118.CleanupReadyStatus || ready.CleanupRunNonceSHA256 != request.CleanupRunNonceSHA256 {
		t.Fatal("synthetic cleanup marker does not match the runtime contract")
	}
}

func TestRequestEvidencePrefixBoundaryMatchesSchema(t *testing.T) {
	t.Parallel()
	for fixturePath, schemaPath := range map[string]string{
		"fixtures/synthetic-baseline-request.json":   "baseline-request.schema.json",
		"fixtures/synthetic-collection-request.json": "collection-request.schema.json",
	} {
		schema, err := jsonschema.NewCompiler().Compile(schemaPath)
		if err != nil {
			t.Fatal(err)
		}
		// #nosec G304 -- paths are a closed repository-owned fixture map.
		payload, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
		var request map[string]any
		if err := json.Unmarshal(payload, &request); err != nil {
			t.Fatal(err)
		}
		request["evidencePrefix"] = strings.Repeat("a", 229)
		boundary, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(boundary))
		if err != nil || schema.Validate(instance) != nil {
			t.Fatalf("%s rejected 229-byte evidence prefix", schemaPath)
		}
		request["evidencePrefix"] = strings.Repeat("a", 230)
		overflow, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		instance, err = jsonschema.UnmarshalJSON(bytes.NewReader(overflow))
		if err != nil {
			t.Fatal(err)
		}
		if schema.Validate(instance) == nil {
			t.Fatalf("%s accepted 230-byte evidence prefix", schemaPath)
		}
		request["evidencePrefix"] = "a..b"
		dotTraversal, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		instance, err = jsonschema.UnmarshalJSON(bytes.NewReader(dotTraversal))
		if err != nil {
			t.Fatal(err)
		}
		if schema.Validate(instance) == nil {
			t.Fatalf("%s accepted evidence prefix with dot traversal", schemaPath)
		}
	}
}
