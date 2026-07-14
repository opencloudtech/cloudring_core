//go:build unix

// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina ZZP 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package secureexec

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/opencloudtech/CloudRING/pkg/kubeconfigpipe"
)

func TestPinnedExecutableIgnoresLaterPATHReplacement(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "kubectl")
	writeExecutable(t, path, "#!/bin/sh\ncat \"$KUBECONFIG\" >/dev/null\nprintf first\n")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	executable, err := Pin("kubectl", 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	writeExecutable(t, path, "#!/bin/sh\ncat \"$KUBECONFIG\" >/dev/null\nprintf replacement\n")
	replay := newReplay(t, []byte("apiVersion: v1\n"))
	defer replay.Close()
	stdout, stderr, err := executable.Run(context.Background(), nil, nil, 1024, 1024, replay)
	defer zero(stdout)
	defer zero(stderr)
	if err != nil || string(stdout) != "first" || len(stderr) != 0 {
		t.Fatalf("pinned execution = %q, stderr=%q, err=%v", stdout, stderr, err)
	}
}

func TestPinnedExecutableBoundsCapturedOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adapter")
	writeExecutable(t, path, "#!/bin/sh\nprintf 0123456789abcdef\n")
	executable, err := PinAbsolute(path, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	stdout, stderr, err := executable.Run(context.Background(), nil, nil, 8, 8, nil)
	zero(stdout)
	zero(stderr)
	if err == nil || len(stdout) != 0 || len(stderr) != 0 {
		t.Fatalf("oversized output = stdout:%d stderr:%d err:%v", len(stdout), len(stderr), err)
	}
	if _, _, err := executable.Run(context.Background(), nil, nil, maxCapturedOutputBytes+1, 8, nil); err == nil {
		t.Fatal("unbounded capture limit unexpectedly passed")
	}
}

func TestPinnedExecutablePreservesBoundedCommandFailureDiagnostics(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adapter")
	writeExecutable(t, path, "#!/bin/sh\nprintf missing >&2\nexit 2\n")
	executable, err := PinAbsolute(path, 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	stdout, stderr, err := executable.Run(context.Background(), nil, nil, 64, 64, nil)
	defer zero(stdout)
	defer zero(stderr)
	if err == nil || len(stdout) != 0 || string(stderr) != "missing" {
		t.Fatalf("command failure = stdout:%q stderr:%q err:%v", stdout, stderr, err)
	}
}

func TestPinnedExecutableAllowsConcurrentReplayRuns(t *testing.T) {
	dir := t.TempDir()
	markerDir := filepath.Join(dir, "markers")
	if err := os.Mkdir(markerDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "kubectl")
	writeExecutable(t, path, `#!/bin/sh
set -eu
cat "$KUBECONFIG" >/dev/null
: > "$1/$2"
attempt=0
while [ "$(find "$1" -type f | wc -l | tr -d ' ')" -lt "$3" ]; do
  attempt=$((attempt + 1))
  [ "$attempt" -lt 100 ] || exit 94
  sleep 0.02
done
printf overlap-ok
`)
	executable, err := PinAbsolute(path, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	replay := newReplay(t, []byte("apiVersion: v1\n"))
	defer replay.Close()

	const invocations = 4
	var wait sync.WaitGroup
	errorsByInvocation := make([]error, invocations)
	for invocation := range invocations {
		wait.Add(1)
		go func() {
			defer wait.Done()
			stdout, stderr, runErr := executable.Run(
				context.Background(),
				[]string{markerDir, strconv.Itoa(invocation), strconv.Itoa(invocations)},
				nil,
				1024,
				1024,
				replay,
			)
			defer zero(stdout)
			defer zero(stderr)
			if runErr != nil || string(stdout) != "overlap-ok" || len(stderr) != 0 {
				errorsByInvocation[invocation] = errors.New("concurrent pinned replay invocation failed")
			}
		}()
	}
	wait.Wait()
	for invocation, err := range errorsByInvocation {
		if err != nil {
			t.Fatalf("invocation %d: %v", invocation, err)
		}
	}
}

func TestPinnedReplayKillsDescendantHoldingOutputAndKubeconfig(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "descendant-pid")
	scriptPath := filepath.Join(t.TempDir(), "kubectl")
	writeExecutable(t, scriptPath, "#!/bin/sh\ncat \"$KUBECONFIG\" >/dev/null\nsleep 300 &\nprintf '%s' \"$!\" > \"$1\"\nprintf ok\n")
	executable, err := PinAbsolute(scriptPath, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	replay := newReplay(t, bytes.Repeat([]byte{'a'}, 1<<20))
	defer replay.Close()
	started := time.Now()
	stdout, stderr, err := executable.Run(context.Background(), []string{pidPath}, nil, 1024, 1024, replay)
	if err == nil {
		t.Fatal("descendant retaining command descriptors unexpectedly passed")
	}
	if len(stdout) != 0 || len(stderr) != 0 {
		t.Fatal("lifecycle failure returned untrusted child output")
	}
	if time.Since(started) > 4*time.Second {
		t.Fatal("descendant cleanup exceeded the bounded wait")
	}
	assertProcessGone(t, pidPath)
}

func TestPinnedUnreadReplayCannotBlockDescendantCleanup(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "descendant-pid")
	scriptPath := filepath.Join(t.TempDir(), "kubectl")
	writeExecutable(t, scriptPath, "#!/bin/sh\nsleep 300 &\nprintf '%s' \"$!\" > \"$1\"\nprintf ok\n")
	executable, err := PinAbsolute(scriptPath, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer executable.Close()
	replay := newReplay(t, bytes.Repeat([]byte{'a'}, 1<<20))
	defer replay.Close()
	started := time.Now()
	stdout, stderr, err := executable.Run(context.Background(), []string{pidPath}, nil, 1024, 1024, replay)
	if err == nil {
		t.Fatal("command that ignored kubeconfig unexpectedly passed")
	}
	if len(stdout) != 0 || len(stderr) != 0 {
		t.Fatal("replay lifecycle failure returned untrusted child output")
	}
	if time.Since(started) > 4*time.Second {
		t.Fatal("unread replay blocked descendant cleanup")
	}
	assertProcessGone(t, pidPath)
}

func writeExecutable(t *testing.T, path, script string) {
	t.Helper()
	// #nosec G306 -- the test-owned file must be executable and is created inside t.TempDir().
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
}

func newReplay(t *testing.T, data []byte) *kubeconfigpipe.Replay {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	writeDone := make(chan error, 1)
	go func() {
		_, writeErr := writer.Write(data)
		writeDone <- errors.Join(writeErr, writer.Close())
	}()
	replay, err := kubeconfigpipe.NewFromFD(int(reader.Fd()))
	closeErr := reader.Close()
	writeErr := <-writeDone
	if err != nil || closeErr != nil || writeErr != nil {
		t.Fatalf("create replay: read=%v close=%v write=%v", err, closeErr, writeErr)
	}
	return replay
}

func assertProcessGone(t *testing.T, path string) {
	t.Helper()
	// #nosec G304 -- path is a test-owned pid file inside t.TempDir().
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(payload)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid descendant pid %q: %v", payload, err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		err = syscall.Kill(pid, 0)
		if errors.Is(err, syscall.ESRCH) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("descendant process %d survived cleanup", pid)
}
