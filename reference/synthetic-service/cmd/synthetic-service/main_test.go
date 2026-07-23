package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRunCheckExercisesSyntheticReconciliation(t *testing.T) {
	var output bytes.Buffer
	if err := run(context.Background(), []string{"--mode=mock-provider", "--check"}, &output); err != nil {
		t.Fatalf("run check: %v", err)
	}
	if got := strings.TrimSpace(output.String()); got != "synthetic_service_check_ok mode=mock-provider phase=ready" {
		t.Fatalf("unexpected check output: %q", got)
	}
}

func TestRunRejectsUnknownMode(t *testing.T) {
	var output bytes.Buffer
	err := run(context.Background(), []string{"--mode=live-provider", "--check"}, &output)
	if err == nil || !strings.Contains(err.Error(), "unsupported mode") {
		t.Fatalf("expected unsupported mode error, got %v", err)
	}
}
