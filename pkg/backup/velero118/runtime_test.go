//go:build !windows

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package velero118

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/opencloudtech/CloudRING/pkg/backup/restoreproof"
)

func TestKubectlReaderUsesExactRawAPIPath(t *testing.T) {
	argumentsPath := filepath.Join(t.TempDir(), "arguments")
	object := string(kubeObject(t, "v1", "PersistentVolumeClaim", metadata("volume", "tenant", "uid", "1", nil, map[string]string{}), map[string]any{}, map[string]any{}, nil))
	script := writeExecutable(t, "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$CLOUDRING_ARGUMENTS\"\nprintf '%s' \"$CLOUDRING_RESPONSE\"\n")
	t.Setenv("CLOUDRING_ARGUMENTS", argumentsPath)
	t.Setenv("CLOUDRING_RESPONSE", object)
	reader, err := NewKubectlReader(script)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if _, err := reader.Get(t.Context(), restoreproof.CoreV1PVCGVR, "tenant", "volume"); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	// #nosec G304 -- argumentsPath is a test-owned path inside t.TempDir().
	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(arguments))
	if got != "get --raw /api/v1/namespaces/tenant/persistentvolumeclaims/volume" || strings.Contains(got, "-o") {
		t.Fatalf("kubectl arguments = %q", got)
	}
}

func TestExecBackendObserverKeepsRawHandleOnStdin(t *testing.T) {
	inputPath := filepath.Join(t.TempDir(), "input")
	script := writeExecutable(t, "#!/bin/sh\ncat > \"$CLOUDRING_INPUT\"\nprintf '%s' \"$CLOUDRING_RESPONSE\"\n")
	t.Setenv("CLOUDRING_INPUT", inputPath)
	observer, err := NewExecBackendObserver(script)
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	handle := "synthetic-provider-handle"
	request := BackendRequest{
		SchemaVersion: AdapterRequestSchemaVersion, Challenge: digest("challenge"), AdapterExecutableSHA256: observer.IdentitySHA256(),
		Operation: "observe", SourceKind: "persistent-volume", ArtifactHandle: handle, ArtifactHandleSHA256: restoreproof.SHA256(handle),
	}
	present := false
	response := BackendObservation{
		SchemaVersion: AdapterResponseSchemaVersion, Implementation: "provider-adapter", Version: "v1", Present: &present,
		RequestSHA256: adapterRequestSHA256(request), AdapterExecutableSHA256: observer.IdentitySHA256(), ArtifactHandleSHA256: request.ArtifactHandleSHA256,
		ObservedAt: "2026-07-14T12:00:00Z", EvidenceRef: "runtime/provider", EvidenceSHA256: digest("evidence"),
	}
	t.Setenv("CLOUDRING_RESPONSE", string(mustJSON(t, response)))
	// The original path changes after construction; execution must continue from
	// the private pinned snapshot and retain the advertised identity.
	// #nosec G306 -- test-only executable in a per-test private directory.
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 99\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.Observe(t.Context(), request)
	if err != nil || observation.Present == nil || *observation.Present {
		t.Fatalf("Observe() = %#v, %v", observation, err)
	}
	// #nosec G304 -- inputPath is a test-owned path inside t.TempDir().
	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(input), handle) {
		t.Fatal("provider handle was not delivered on stdin")
	}
}

func TestExecBackendObserverRejectsUnknownMissingReplayAndInvalidVersion(t *testing.T) {
	script := writeExecutable(t, "#!/bin/sh\ncat >/dev/null\nprintf '%s' \"$CLOUDRING_RESPONSE\"\n")
	observer, err := NewExecBackendObserver(script)
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	request := BackendRequest{
		SchemaVersion: AdapterRequestSchemaVersion, Challenge: digest("challenge"), AdapterExecutableSHA256: observer.IdentitySHA256(),
		Operation: "observe", SourceKind: "persistent-volume", ArtifactHandle: "synthetic-handle", ArtifactHandleSHA256: digest("synthetic-handle"),
	}
	present := false
	base := BackendObservation{
		SchemaVersion: AdapterResponseSchemaVersion, Implementation: "provider-adapter", Version: "v1", Present: &present,
		RequestSHA256: adapterRequestSHA256(request), AdapterExecutableSHA256: observer.IdentitySHA256(), ArtifactHandleSHA256: request.ArtifactHandleSHA256,
		ObservedAt: "2026-07-14T12:00:00Z", EvidenceRef: "runtime/provider", EvidenceSHA256: digest("evidence"),
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "unknown field", mutate: func(value map[string]any) { value["unknown"] = true }},
		{name: "missing present", mutate: func(value map[string]any) { delete(value, "present") }},
		{name: "replayed request", mutate: func(value map[string]any) { value["requestSha256"] = digest("old-request") }},
		{name: "invalid version", mutate: func(value map[string]any) { value["version"] = "1.0.0" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			payload := mustJSON(t, base)
			var value map[string]any
			if err := json.Unmarshal(payload, &value); err != nil {
				t.Fatal(err)
			}
			test.mutate(value)
			t.Setenv("CLOUDRING_RESPONSE", string(mustJSON(t, value)))
			if _, err := observer.Observe(t.Context(), request); err == nil {
				t.Fatal("invalid adapter response unexpectedly passed")
			}
		})
	}
}

func TestPinnedCommandTimeoutKillsProcessGroup(t *testing.T) {
	script := writeExecutable(t, "#!/bin/sh\nsleep 30 &\nwait\n")
	executable, err := pinExecutable(script, 50*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	started := time.Now()
	if _, _, err := runCommand(context.Background(), executable, nil, nil, 1024); err == nil {
		t.Fatal("timed out process tree unexpectedly succeeded")
	}
	if time.Since(started) > 2*time.Second {
		t.Fatal("process-tree timeout was not enforced")
	}
}

func TestSuccessfulPinnedCommandKillsBackgroundProcessGroup(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "background-pid")
	script := writeExecutable(t, "#!/bin/sh\nsleep 300 </dev/null >/dev/null 2>&1 &\nprintf '%s' \"$!\" > \"$CLOUDRING_BACKGROUND_PID\"\nprintf 'ok'\n")
	t.Setenv("CLOUDRING_BACKGROUND_PID", pidPath)
	executable, err := pinExecutable(script, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	output, _, err := runCommand(context.Background(), executable, nil, nil, 1024)
	if err != nil || string(output) != "ok" {
		t.Fatalf("successful command = %q, %v", output, err)
	}
	// #nosec G304 -- pidPath is a test-owned path inside t.TempDir().
	payload, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(payload)))
	if err != nil || pid <= 0 {
		t.Fatalf("background pid = %q, %v", payload, err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		err = syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("background process %d survived successful adapter exit", pid)
}

func TestAdapterErrorsDoNotReflectChildOutput(t *testing.T) {
	script := writeExecutable(t, "#!/bin/sh\nprintf 'private-child-value' >&2\nexit 1\n")
	observer, err := NewExecBackendObserver(script)
	if err != nil {
		t.Fatal(err)
	}
	defer observer.Close()
	_, err = observer.Observe(t.Context(), BackendRequest{
		SchemaVersion: AdapterRequestSchemaVersion, Challenge: digest("challenge"), AdapterExecutableSHA256: observer.IdentitySHA256(), Operation: "observe", SourceKind: "persistent-volume",
		ArtifactHandle: "synthetic-handle", ArtifactHandleSHA256: restoreproof.SHA256("synthetic-handle"),
	})
	if err == nil || strings.Contains(err.Error(), "private-child-value") {
		t.Fatalf("sanitized error = %v", err)
	}
}

func writeExecutable(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "adapter")
	// #nosec G306 -- test-only executable in a per-test private directory.
	if err := os.WriteFile(path, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}
