// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package velero118

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
	"github.com/opencloudtech/CloudRING/pkg/backup/restoreproof"
	"github.com/opencloudtech/CloudRING/pkg/kubeconfigpipe"
)

const (
	maxAdapterResponseBytes  = 64 << 10
	maxPinnedExecutableBytes = 512 << 20
	defaultKubectlTimeout    = 2 * time.Minute
	defaultProbeTimeout      = 30 * time.Minute
	defaultBackendTimeout    = 2 * time.Minute
)

// KubectlReader uses kubectl's raw API path so discovery cannot silently select
// a different served version. It performs read-only GET requests.
type KubectlReader struct {
	executable *pinnedExecutable
	kubeconfig *kubeconfigpipe.Replay
}

func NewKubectlReader(binary string) (*KubectlReader, error) {
	return newKubectlReader(binary, nil)
}

// NewKubectlReaderFromKubeconfigFD consumes one pipe-backed kubeconfig from fd
// and replays it through a fresh anonymous pipe for every kubectl invocation.
// This supports multi-query collectors without ever persisting kubeconfig
// bytes or exposing them through argv or the child environment.
func NewKubectlReaderFromKubeconfigFD(binary string, fd int) (*KubectlReader, error) {
	replay, err := kubeconfigpipe.NewFromFD(fd)
	if err != nil {
		return nil, err
	}
	reader, err := newKubectlReader(binary, replay)
	if err != nil {
		_ = replay.Close()
		return nil, err
	}
	return reader, nil
}

func newKubectlReader(binary string, replay *kubeconfigpipe.Replay) (*KubectlReader, error) {
	resolved, err := resolveExecutable(binary)
	if err != nil {
		return nil, errors.New("resolve kubectl executable")
	}
	executable, err := pinExecutable(resolved, defaultKubectlTimeout)
	if err != nil {
		return nil, errors.New("pin kubectl executable")
	}
	return &KubectlReader{executable: executable, kubeconfig: replay}, nil
}

func (reader *KubectlReader) Close() error {
	if reader == nil {
		return nil
	}
	var executableErr error
	if reader.executable != nil {
		executableErr = reader.executable.Close()
	}
	return errors.Join(executableErr, reader.kubeconfig.Close())
}

func (reader *KubectlReader) run(ctx context.Context, arguments []string, maximum int64) ([]byte, []byte, error) {
	if reader == nil || reader.executable == nil {
		return nil, nil, errors.New("kubectl reader is closed")
	}
	if reader.kubeconfig == nil {
		return runCommand(ctx, reader.executable, arguments, nil, maximum)
	}
	return runCommandWithKubeconfig(ctx, reader.executable, arguments, maximum, reader.kubeconfig)
}

func (reader *KubectlReader) Get(ctx context.Context, gvr restoreproof.GVR, namespace, name string) ([]byte, error) {
	if reader == nil || reader.executable == nil || !safeGVR(gvr) || namespace != "" && !safeName(namespace) || !safeName(name) {
		return nil, errors.New("invalid Kubernetes object read")
	}
	path := rawResourcePath(gvr, namespace) + "/" + url.PathEscape(name)
	output, stderr, err := reader.run(ctx, []string{"get", "--raw", path}, strictjson.MaxDocumentBytes)
	defer zeroBytes(stderr)
	if err != nil {
		notFound := possibleNotFound(output, stderr)
		zeroBytes(output)
		if notFound {
			return nil, ErrNotFound
		}
		return nil, errors.New("Kubernetes raw object read failed")
	}
	if err := strictjson.Validate(output); err != nil {
		return nil, errors.New("Kubernetes raw object response is invalid")
	}
	return output, nil
}

func (reader *KubectlReader) ListPage(ctx context.Context, gvr restoreproof.GVR, namespace, selector, continueToken string, limit int) ([]byte, error) {
	if reader == nil || reader.executable == nil || !safeGVR(gvr) || namespace != "" && !safeName(namespace) || limit <= 0 || limit > 1000 || len(selector) > 2048 || len(continueToken) > 8192 {
		return nil, errors.New("invalid Kubernetes list read")
	}
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	if selector != "" {
		query.Set("labelSelector", selector)
	}
	if continueToken != "" {
		query.Set("continue", continueToken)
	}
	path := rawResourcePath(gvr, namespace) + "?" + query.Encode()
	output, stderr, err := reader.run(ctx, []string{"get", "--raw", path}, strictjson.MaxDocumentBytes)
	defer zeroBytes(stderr)
	if err != nil {
		zeroBytes(output)
		return nil, errors.New("Kubernetes raw list read failed")
	}
	if err := strictjson.Validate(output); err != nil {
		return nil, errors.New("Kubernetes raw list response is invalid")
	}
	return output, nil
}

// ConfirmAbsent proves an exact-object 404 against a successful list of the
// same collection. A proxy or wrong-GVR 404 therefore cannot satisfy cleanup.
func (reader *KubectlReader) ConfirmAbsent(ctx context.Context, gvr restoreproof.GVR, namespace, name string) (bool, error) {
	apiVersion, kind, ok := exactListGVK(gvr)
	if reader == nil || reader.executable == nil || !ok || namespace != "" && !safeName(namespace) || !safeName(name) {
		return false, errors.New("invalid Kubernetes absence confirmation")
	}
	query := url.Values{}
	query.Set("fieldSelector", "metadata.name="+name)
	query.Set("limit", "2")
	path := rawResourcePath(gvr, namespace) + "?" + query.Encode()
	output, stderr, err := reader.run(ctx, []string{"get", "--raw", path}, strictjson.MaxDocumentBytes)
	defer zeroBytes(stderr)
	if err != nil {
		zeroBytes(output)
		return false, errors.New("Kubernetes exact collection read failed")
	}
	defer zeroBytes(output)
	page, err := DecodeListPage(output, apiVersion, kind)
	if err != nil || page.Continue != "" || len(page.Items) > 1 {
		return false, errors.New("Kubernetes exact collection response is invalid")
	}
	if len(page.Items) == 0 {
		return true, nil
	}
	var envelope objectEnvelope
	if strictjson.Decode(page.Items[0], &envelope) != nil || envelope.Metadata.Namespace != namespace || envelope.Metadata.Name != name {
		return false, errors.New("Kubernetes exact collection identity is invalid")
	}
	return false, nil
}

func exactListGVK(gvr restoreproof.GVR) (string, string, bool) {
	switch gvr {
	case restoreproof.CoreV1PVCGVR:
		return "v1", "PersistentVolumeClaimList", true
	case restoreproof.CoreV1PVGVR:
		return "v1", "PersistentVolumeList", true
	case restoreproof.CoreV1CMGVR:
		return "v1", "ConfigMapList", true
	case restoreproof.DataDownloadGVR:
		return "velero.io/v2alpha1", "DataDownloadList", true
	default:
		return "", "", false
	}
}

func possibleNotFound(output, stderr []byte) bool {
	for _, payload := range [][]byte{output, stderr} {
		var status struct {
			APIVersion string `json:"apiVersion"`
			Kind       string `json:"kind"`
			Reason     string `json:"reason"`
			Code       int    `json:"code"`
		}
		if strictjson.Decode(payload, &status) == nil && status.APIVersion == "v1" && status.Kind == "Status" && status.Reason == "NotFound" && status.Code == 404 {
			return true
		}
	}
	return bytes.Contains(stderr, []byte("(NotFound)"))
}

// ExecProbeObserver invokes a downstream read-only probe adapter. Object refs
// enter through stdin and adapter stderr is never reflected.
type ExecProbeObserver struct{ executable *pinnedExecutable }

func NewExecProbeObserver(path string) (*ExecProbeObserver, error) {
	resolved, err := resolveAbsoluteExecutable(path)
	if err != nil {
		return nil, errors.New("resolve data probe adapter")
	}
	executable, err := pinExecutable(resolved, defaultProbeTimeout)
	if err != nil {
		return nil, errors.New("pin data probe adapter")
	}
	return &ExecProbeObserver{executable: executable}, nil
}

func (observer *ExecProbeObserver) IdentitySHA256() string {
	if observer == nil || observer.executable == nil {
		return ""
	}
	return observer.executable.IdentitySHA256()
}

func (observer *ExecProbeObserver) Close() error {
	if observer == nil || observer.executable == nil {
		return nil
	}
	return observer.executable.Close()
}

func (observer *ExecProbeObserver) Observe(ctx context.Context, request ProbeRequest) (ProbeObservation, error) {
	if observer == nil || observer.executable == nil || request.SchemaVersion != AdapterRequestSchemaVersion || !validSHA256(request.Challenge) ||
		request.AdapterExecutableSHA256 != observer.IdentitySHA256() {
		return ProbeObservation{}, errors.New("invalid data probe adapter invocation")
	}
	input, err := json.Marshal(request)
	if err != nil {
		return ProbeObservation{}, errors.New("encode data probe request")
	}
	requestSHA256 := restoreproof.SHA256(string(input))
	output, childError, err := runCommand(ctx, observer.executable, nil, input, maxAdapterResponseBytes)
	zeroBytes(input)
	zeroBytes(childError)
	if err != nil {
		zeroBytes(output)
		return ProbeObservation{}, errors.New("data probe adapter failed")
	}
	defer zeroBytes(output)
	var observation ProbeObservation
	if err := strictjson.DecodeExact(output, &observation); err != nil || !validProbeObservation(observation) ||
		observation.RequestSHA256 != requestSHA256 || observation.AdapterExecutableSHA256 != observer.IdentitySHA256() {
		return ProbeObservation{}, errors.New("data probe adapter response is invalid")
	}
	return observation, nil
}

// ExecBackendObserver invokes a provider adapter with the raw handle only on
// stdin. The handle is zeroed from the request buffer after execution.
type ExecBackendObserver struct{ executable *pinnedExecutable }

func NewExecBackendObserver(path string) (*ExecBackendObserver, error) {
	resolved, err := resolveAbsoluteExecutable(path)
	if err != nil {
		return nil, errors.New("resolve provider observer adapter")
	}
	executable, err := pinExecutable(resolved, defaultBackendTimeout)
	if err != nil {
		return nil, errors.New("pin provider observer adapter")
	}
	return &ExecBackendObserver{executable: executable}, nil
}

func (observer *ExecBackendObserver) IdentitySHA256() string {
	if observer == nil || observer.executable == nil {
		return ""
	}
	return observer.executable.IdentitySHA256()
}

func (observer *ExecBackendObserver) Close() error {
	if observer == nil || observer.executable == nil {
		return nil
	}
	return observer.executable.Close()
}

func (observer *ExecBackendObserver) Observe(ctx context.Context, request BackendRequest) (BackendObservation, error) {
	if observer == nil || observer.executable == nil || request.SchemaVersion != AdapterRequestSchemaVersion || !validSHA256(request.Challenge) ||
		request.AdapterExecutableSHA256 != observer.IdentitySHA256() || request.Operation != "observe" || request.SourceKind != "persistent-volume" ||
		request.ArtifactHandle == "" || restoreproof.SHA256(request.ArtifactHandle) != request.ArtifactHandleSHA256 {
		return BackendObservation{}, errors.New("invalid provider observer invocation")
	}
	input, err := json.Marshal(request)
	if err != nil {
		return BackendObservation{}, errors.New("encode provider observation request")
	}
	requestSHA256 := restoreproof.SHA256(string(input))
	output, childError, err := runCommand(ctx, observer.executable, nil, input, maxAdapterResponseBytes)
	zeroBytes(input)
	zeroBytes(childError)
	if err != nil {
		zeroBytes(output)
		return BackendObservation{}, errors.New("provider observer adapter failed")
	}
	defer zeroBytes(output)
	var observation BackendObservation
	if err := strictjson.DecodeExact(output, &observation); err != nil || !validBackendObservation(observation) ||
		observation.RequestSHA256 != requestSHA256 || observation.AdapterExecutableSHA256 != observer.IdentitySHA256() ||
		observation.ArtifactHandleSHA256 != request.ArtifactHandleSHA256 {
		return BackendObservation{}, errors.New("provider observer adapter response is invalid")
	}
	return observation, nil
}

func rawResourcePath(gvr restoreproof.GVR, namespace string) string {
	var base string
	if gvr.Group == "" {
		base = "/api/" + url.PathEscape(gvr.Version)
	} else {
		base = "/apis/" + url.PathEscape(gvr.Group) + "/" + url.PathEscape(gvr.Version)
	}
	if namespace != "" {
		base += "/namespaces/" + url.PathEscape(namespace)
	}
	return base + "/" + url.PathEscape(gvr.Resource)
}

func safeGVR(gvr restoreproof.GVR) bool {
	return (gvr.Group == "" || safeAPISegment(gvr.Group)) && safeAPISegment(gvr.Version) && safeAPISegment(gvr.Resource)
}

func safeAPISegment(value string) bool {
	if value == "" || len(value) > 253 || strings.ContainsAny(value, "/\\?&#%\x00\t\r\n ") {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z' || character >= '0' && character <= '9' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

func resolveExecutable(binary string) (string, error) {
	if binary == "" {
		binary = "kubectl"
	}
	if filepath.IsAbs(binary) {
		return resolveAbsoluteExecutable(binary)
	}
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return resolveAbsoluteExecutable(absolute)
}

func resolveAbsoluteExecutable(path string) (string, error) {
	if !filepath.IsAbs(path) || strings.ContainsRune(path, '\x00') {
		return "", errors.New("executable path must be absolute")
	}
	clean := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(clean)
	if err != nil || !filepath.IsAbs(resolved) {
		return "", errors.New("resolve executable identity")
	}
	return resolved, nil
}

type pinnedExecutable struct {
	mu             sync.Mutex
	file           *os.File
	invocationPath string
	snapshotDir    string
	useDescriptor  bool
	identitySHA256 string
	timeout        time.Duration
	closed         bool
}

func pinExecutable(path string, timeout time.Duration) (*pinnedExecutable, error) {
	if timeout <= 0 || runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil, errors.New("pinned executable runtime is unsupported")
	}
	// #nosec G304 -- path is an absolute, symlink-resolved executable identity and is validated below as a bounded regular executable.
	source, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0o111 == 0 || info.Size() <= 0 || info.Size() > maxPinnedExecutableBytes {
		return nil, errors.New("executable identity is invalid")
	}
	snapshotDir, err := os.MkdirTemp("", ".cloudring-pinned-exec-")
	if err != nil {
		return nil, errors.New("create pinned executable directory")
	}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(snapshotDir)
		}
	}()
	// #nosec G302 -- the directory must be searchable only by its owner so its private executable can run.
	if err := os.Chmod(snapshotDir, 0o700); err != nil {
		return nil, errors.New("protect pinned executable directory")
	}
	snapshotPath := filepath.Join(snapshotDir, "executable")
	// #nosec G304 G302 -- snapshotPath is inside the fresh private directory and an executable copy requires owner execute permission.
	snapshot, err := os.OpenFile(snapshotPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o500)
	if err != nil {
		return nil, errors.New("create pinned executable snapshot")
	}
	hasher := sha256.New()
	copied, copyErr := io.Copy(io.MultiWriter(snapshot, hasher), io.LimitReader(source, maxPinnedExecutableBytes+1))
	chmodErr := snapshot.Chmod(0o500)
	syncErr := snapshot.Sync()
	closeErr := snapshot.Close()
	if copyErr != nil || chmodErr != nil || syncErr != nil || closeErr != nil || copied != info.Size() {
		return nil, errors.New("write pinned executable snapshot")
	}
	// #nosec G304 -- snapshotPath is the exact private file created above and is revalidated immediately after opening.
	pinned, err := os.Open(snapshotPath)
	if err != nil {
		return nil, errors.New("open pinned executable snapshot")
	}
	pinnedInfo, statErr := pinned.Stat()
	if statErr != nil || !pinnedInfo.Mode().IsRegular() || pinnedInfo.Mode().Perm()&0o111 == 0 || pinnedInfo.Size() != info.Size() {
		_ = pinned.Close()
		return nil, errors.New("pinned executable snapshot is invalid")
	}
	invocationPath := snapshotPath
	useDescriptor := runtime.GOOS == "linux"
	retainedDir := snapshotDir
	if useDescriptor {
		invocationPath = "/proc/self/fd/3"
		if err := os.Remove(snapshotPath); err != nil {
			_ = pinned.Close()
			return nil, errors.New("unlink pinned executable snapshot")
		}
		if err := os.Remove(snapshotDir); err != nil {
			_ = pinned.Close()
			return nil, errors.New("remove pinned executable directory")
		}
		retainedDir = ""
	}
	cleanup = false
	return &pinnedExecutable{file: pinned, invocationPath: invocationPath, snapshotDir: retainedDir, useDescriptor: useDescriptor, identitySHA256: hex.EncodeToString(hasher.Sum(nil)), timeout: timeout}, nil
}

func (executable *pinnedExecutable) IdentitySHA256() string {
	if executable == nil {
		return ""
	}
	executable.mu.Lock()
	defer executable.mu.Unlock()
	if executable.closed {
		return ""
	}
	return executable.identitySHA256
}

func (executable *pinnedExecutable) Close() error {
	if executable == nil {
		return nil
	}
	executable.mu.Lock()
	defer executable.mu.Unlock()
	if executable.closed {
		return nil
	}
	executable.closed = true
	closeErr := executable.file.Close()
	removeErr := error(nil)
	if executable.snapshotDir != "" {
		removeErr = os.RemoveAll(executable.snapshotDir)
	}
	if closeErr != nil {
		return closeErr
	}
	return removeErr
}

func runCommand(ctx context.Context, executable *pinnedExecutable, arguments []string, input []byte, maximum int64) ([]byte, []byte, error) {
	return runCommandConfigured(ctx, executable, arguments, input, maximum, nil)
}

func runCommandWithKubeconfig(ctx context.Context, executable *pinnedExecutable, arguments []string, maximum int64, kubeconfig *kubeconfigpipe.Replay) ([]byte, []byte, error) {
	if kubeconfig == nil {
		return nil, nil, errors.New("pipe-backed kubeconfig is missing")
	}
	return runCommandConfigured(ctx, executable, arguments, nil, maximum, kubeconfig)
}

func runCommandConfigured(ctx context.Context, executable *pinnedExecutable, arguments []string, input []byte, maximum int64, kubeconfig *kubeconfigpipe.Replay) ([]byte, []byte, error) {
	if executable == nil {
		return nil, nil, errors.New("command executable is missing")
	}
	executable.mu.Lock()
	defer executable.mu.Unlock()
	if executable.closed || executable.file == nil || executable.timeout <= 0 {
		return nil, nil, errors.New("command executable is closed")
	}
	if executable.useDescriptor {
		if _, err := executable.file.Seek(0, io.SeekStart); err != nil {
			return nil, nil, errors.New("rewind command executable")
		}
	} else if err := executable.verifySnapshot(); err != nil {
		return nil, nil, errors.New("rewind command executable")
	}
	invocationContext, cancel := context.WithTimeout(ctx, executable.timeout)
	defer cancel()
	// #nosec G204 -- execution uses an already-open, content-hashed file
	// descriptor; validated arguments contain no adapter data or credentials.
	command := exec.CommandContext(invocationContext, executable.invocationPath, arguments...)
	if executable.useDescriptor {
		command.ExtraFiles = []*os.File{executable.file}
	}
	var completeKubeconfig func() error
	if kubeconfig != nil {
		var err error
		completeKubeconfig, err = kubeconfig.Attach(command)
		if err != nil {
			return nil, nil, err
		}
	}
	configureProcessTree(command)
	if input != nil {
		command.Stdin = bytes.NewReader(input)
	}
	stdout := boundedBuffer{maximum: int(maximum)}
	stderr := boundedBuffer{maximum: 64 << 10}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	cleanupErr := cleanupProcessTree(command)
	var kubeconfigReplayErr error
	if completeKubeconfig != nil {
		kubeconfigReplayErr = completeKubeconfig()
	}
	if stdout.exceeded || stderr.exceeded {
		zeroBytes(stdout.buffer.Bytes())
		zeroBytes(stderr.buffer.Bytes())
		return nil, nil, errors.New("command output exceeded limit")
	}
	output := append([]byte(nil), stdout.buffer.Bytes()...)
	errorOutput := append([]byte(nil), stderr.buffer.Bytes()...)
	zeroBytes(stdout.buffer.Bytes())
	zeroBytes(stderr.buffer.Bytes())
	if cleanupErr != nil {
		zeroBytes(output)
		zeroBytes(errorOutput)
		return nil, nil, errors.New("command process-tree cleanup failed")
	}
	if kubeconfigReplayErr != nil {
		zeroBytes(output)
		zeroBytes(errorOutput)
		return nil, nil, kubeconfigReplayErr
	}
	if err != nil {
		return output, errorOutput, errors.New("command failed")
	}
	zeroBytes(errorOutput)
	return output, nil, nil
}

func (executable *pinnedExecutable) verifySnapshot() error {
	if executable.snapshotDir == "" || executable.invocationPath == "" {
		return errors.New("pinned executable snapshot is missing")
	}
	file, err := os.Open(executable.invocationPath)
	if err != nil {
		return err
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, io.LimitReader(file, maxPinnedExecutableBytes+1)); err != nil {
		return err
	}
	if hex.EncodeToString(hasher.Sum(nil)) != executable.identitySHA256 {
		return errors.New("pinned executable snapshot changed")
	}
	return nil
}

type boundedBuffer struct {
	buffer   bytes.Buffer
	maximum  int
	exceeded bool
}

func (writer *boundedBuffer) Write(value []byte) (int, error) {
	if writer.maximum <= 0 {
		writer.exceeded = true
		return len(value), nil
	}
	remaining := writer.maximum + 1 - writer.buffer.Len()
	if remaining > 0 {
		if remaining > len(value) {
			remaining = len(value)
		}
		_, _ = writer.buffer.Write(value[:remaining])
	}
	if writer.buffer.Len() > writer.maximum || remaining < len(value) {
		writer.exceeded = true
	}
	return len(value), nil
}

func zeroBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
