// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIProducesReadyPreflightAndDeterministicPlan(t *testing.T) {
	root := filepath.Join("..", "..")
	profile := filepath.Join(root, "examples", "provider-site-profile.yaml")
	for _, command := range []string{"preflight", "plan"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{command, "--profile", profile}, nil, &stdout, &stderr); code != 0 {
			t.Fatalf("%s failed with code %d: %s", command, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), `"nonClaim": "preflight-and-plan-only"`) {
			t.Fatalf("%s output lacks non-claim: %s", command, stdout.String())
		}
	}
}

func TestCLIBlocksInvalidProfileWithoutEchoingInput(t *testing.T) {
	secretLikeCanary := "value-that-must-not-be-echoed"
	var stdout, stderr bytes.Buffer
	code := run([]string{"preflight", "--profile", "-"}, strings.NewReader("unknown: "+secretLikeCanary), &stdout, &stderr)
	if code != 1 || strings.Contains(stdout.String()+stderr.String(), secretLikeCanary) {
		t.Fatalf("invalid input was accepted or echoed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestCLIBlocksSemanticInvalidNameWithoutEchoingInput(t *testing.T) {
	secretLikeCanary := "invalid_name"
	profile := strings.Replace(validCLIProfile(t), "name: synthetic-provider-site", "name: "+secretLikeCanary, 1)
	var stdout, stderr bytes.Buffer
	code := run([]string{"preflight", "--profile", "-"}, strings.NewReader(profile), &stdout, &stderr)
	if code != 1 || strings.Contains(stdout.String()+stderr.String(), secretLikeCanary) {
		t.Fatalf("semantic invalid input was accepted or echoed: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestCLIBlocksMissingProfile(t *testing.T) {
	var stdout, stderr bytes.Buffer
	path := filepath.Join(t.TempDir(), "missing.yaml")
	if code := run([]string{"plan", "--profile", path}, nil, &stdout, &stderr); code != 1 {
		t.Fatalf("missing profile returned %d", code)
	}
}

func TestMainExampleExists(t *testing.T) {
	if _, err := os.Stat(filepath.Join("..", "..", "examples", "provider-site-profile.yaml")); err != nil {
		t.Fatal(err)
	}
}

func validCLIProfile(t *testing.T) string {
	t.Helper()
	// #nosec G304 -- the test reads the repository-owned example at a fixed relative path.
	payload, err := os.ReadFile(filepath.Join("..", "..", "examples", "provider-site-profile.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
