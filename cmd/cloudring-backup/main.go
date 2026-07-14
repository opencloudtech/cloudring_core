// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
	"github.com/opencloudtech/CloudRING/pkg/backup/restoreproof"
	"github.com/opencloudtech/CloudRING/pkg/backup/velero118"
)

const maxPrivateArtifactBytes = 8 << 20

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "cloudring-backup: operation failed")
		os.Exit(1)
	}
}

func run(ctx context.Context, arguments []string, stdout io.Writer) error {
	if len(arguments) == 0 {
		return errors.New("command is required")
	}
	switch arguments[0] {
	case "baseline":
		return runBaseline(ctx, arguments[1:], stdout)
	case "collect":
		return runCollect(ctx, arguments[1:], stdout)
	case "verify":
		return runVerify(arguments[1:], stdout)
	default:
		return errors.New("unknown command")
	}
}

func runBaseline(ctx context.Context, arguments []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("baseline", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	requestPath := flags.String("request", "", "baseline request JSON")
	outputPath := flags.String("output", "", "new private baseline file")
	kubectl := flags.String("kubectl", "kubectl", "kubectl executable")
	kubeconfigFD := flags.Int("kubeconfig-fd", -1, "pipe descriptor containing kubeconfig; consumed once and replayed in memory")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 || *requestPath == "" || *outputPath == "" {
		return errors.New("invalid baseline arguments")
	}
	if err := validateNewArtifactDestinations(*outputPath); err != nil {
		return errors.New("baseline output destination is unavailable")
	}
	var request velero118.BaselineRequest
	if err := readStrictJSON(*requestPath, &request); err != nil {
		return errors.New("read baseline request")
	}
	reader, err := newCollectorKubectlReader(*kubectl, *kubeconfigFD)
	if err != nil {
		return err
	}
	defer reader.Close()
	baseline, err := velero118.BuildSourceBaseline(ctx, reader, request, velero118.SystemClock())
	if err != nil {
		return err
	}
	if err := writePrivateJSON(*outputPath, baseline); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, "status=baseline_written")
	return nil
}

func runCollect(ctx context.Context, arguments []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("collect", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	requestPath := flags.String("request", "", "collection request JSON")
	baselinePath := flags.String("baseline", "", "private source baseline JSON")
	archivePath := flags.String("archive", "", "Velero BackupContents tar.gz")
	probeAdapterPath := flags.String("data-probe-adapter", "", "absolute probe adapter executable")
	providerAdapterPath := flags.String("provider-adapter", "", "absolute provider observer executable")
	cleanupReadyPath := flags.String("cleanup-ready", "", "new atomic ready-for-cleanup marker")
	outputPath := flags.String("output", "", "new private receipt file")
	kubectl := flags.String("kubectl", "kubectl", "kubectl executable")
	kubeconfigFD := flags.Int("kubeconfig-fd", -1, "pipe descriptor containing kubeconfig; consumed once and replayed in memory")
	cleanupTimeout := flags.Duration("cleanup-timeout", 30*time.Minute, "maximum wait for downstream cleanup")
	pollInterval := flags.Duration("poll-interval", 2*time.Second, "cleanup observation interval")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 || *requestPath == "" || *baselinePath == "" || *archivePath == "" ||
		*probeAdapterPath == "" || *providerAdapterPath == "" || *cleanupReadyPath == "" || *outputPath == "" || *cleanupTimeout <= 0 || *pollInterval <= 0 || samePath(*cleanupReadyPath, *outputPath) {
		return errors.New("invalid collect arguments")
	}
	if err := validateNewArtifactDestinations(*cleanupReadyPath, *outputPath); err != nil {
		return errors.New("collect artifact destination is unavailable")
	}
	var request velero118.CollectionRequest
	if err := readStrictJSON(*requestPath, &request); err != nil {
		return errors.New("read collection request")
	}
	request.CleanupTimeout = *cleanupTimeout
	request.PollInterval = *pollInterval
	var baseline restoreproof.SourceBaseline
	if err := readStrictJSON(*baselinePath, &baseline); err != nil {
		return errors.New("read source baseline")
	}
	archive, err := os.Open(filepath.Clean(*archivePath))
	if err != nil {
		return errors.New("open Velero BackupContents")
	}
	archivedDataUpload, archiveErr := velero118.ReadArchivedDataUpload(archive, request.VeleroNamespace, request.DataUploadName)
	closeErr := archive.Close()
	if archiveErr != nil || closeErr != nil {
		zeroBytes(archivedDataUpload)
		return errors.New("read exact archived DataUpload")
	}
	defer zeroBytes(archivedDataUpload)
	reader, err := newCollectorKubectlReader(*kubectl, *kubeconfigFD)
	if err != nil {
		return err
	}
	defer reader.Close()
	probeObserver, err := velero118.NewExecProbeObserver(*probeAdapterPath)
	if err != nil {
		return err
	}
	defer probeObserver.Close()
	providerObserver, err := velero118.NewExecBackendObserver(*providerAdapterPath)
	if err != nil {
		return err
	}
	defer providerObserver.Close()
	cleanupBarrier := fileCleanupBarrier{path: *cleanupReadyPath}
	receipt, err := velero118.CollectCSIDataMoverVolumeLineage(ctx, reader, probeObserver, providerObserver, cleanupBarrier, request, baseline, archivedDataUpload, velero118.SystemClock())
	if err != nil {
		return err
	}
	if err := writePrivateJSON(*outputPath, receipt); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(stdout, "status=receipt_written")
	return nil
}

func newCollectorKubectlReader(binary string, kubeconfigFD int) (*velero118.KubectlReader, error) {
	if kubeconfigFD == -1 {
		return velero118.NewKubectlReader(binary)
	}
	if kubeconfigFD < 3 {
		return nil, errors.New("kubeconfig pipe descriptor is invalid")
	}
	return velero118.NewKubectlReaderFromKubeconfigFD(binary, kubeconfigFD)
}

func runVerify(arguments []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	receiptPath := flags.String("receipt", "", "private unsigned receipt JSON")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 || *receiptPath == "" {
		return errors.New("invalid verify arguments")
	}
	var receipt restoreproof.VolumeReceipt
	if err := readStrictJSON(*receiptPath, &receipt); err != nil {
		return errors.New("read restore proof receipt")
	}
	if err := restoreproof.ValidateCSIDataMoverVolumeReceipt(&receipt); err != nil {
		return errors.New("restore proof receipt is invalid")
	}
	_, _ = fmt.Fprintln(stdout, "status=verified")
	return nil
}

func readStrictJSON(path string, destination any) error {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := strictjson.Read(io.LimitReader(file, maxPrivateArtifactBytes+1))
	if err != nil {
		return err
	}
	defer zeroBytes(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("decode private JSON artifact")
	}
	return nil
}

func writePrivateJSON(path string, value any) error {
	clean := filepath.Clean(path)
	if clean == "." || clean == string(filepath.Separator) {
		return errors.New("private artifact path is invalid")
	}
	directory := filepath.Dir(clean)
	file, err := os.CreateTemp(directory, ".cloudring-private-artifact-*")
	if err != nil {
		return errors.New("create private artifact temporary file")
	}
	temporaryPath := file.Name()
	defer func() {
		_ = file.Close()
		_ = os.Remove(temporaryPath)
	}()
	if err := file.Chmod(0o600); err != nil {
		return errors.New("protect private artifact temporary file")
	}
	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return errors.New("encode private artifact")
	}
	if err := file.Sync(); err != nil {
		return errors.New("sync private artifact")
	}
	if err := file.Close(); err != nil {
		return errors.New("close private artifact")
	}
	// Link publishes a fully written inode atomically and fails if the target
	// already exists. The temporary file is created in the same directory.
	if err := os.Link(temporaryPath, clean); err != nil {
		return errors.New("publish private artifact without overwrite")
	}
	return nil
}

type fileCleanupBarrier struct{ path string }

func (barrier fileCleanupBarrier) ReadyForCleanup(ctx context.Context, notice velero118.CleanupReady) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	parsed, timeErr := time.Parse(time.RFC3339Nano, notice.ReadyAt)
	_, digestErr := hex.DecodeString(notice.CleanupRunNonceSHA256)
	if notice.SchemaVersion != velero118.CleanupReadySchemaVersion || notice.Status != velero118.CleanupReadyStatus ||
		timeErr != nil || parsed.UTC().Format(time.RFC3339Nano) != notice.ReadyAt || len(notice.CleanupRunNonceSHA256) != 64 || digestErr != nil {
		return errors.New("cleanup readiness notice is invalid")
	}
	return writePrivateJSON(barrier.path, notice)
}

func samePath(left, right string) bool {
	leftDestination, leftErr := canonicalDestination(left)
	rightDestination, rightErr := canonicalDestination(right)
	if leftErr != nil || rightErr != nil {
		return true
	}
	if leftDestination == rightDestination {
		return true
	}
	return filepath.Dir(leftDestination) == filepath.Dir(rightDestination) && strings.EqualFold(filepath.Base(leftDestination), filepath.Base(rightDestination))
}

func canonicalDestination(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	parent, err := filepath.EvalSymlinks(filepath.Dir(absolute))
	if err != nil {
		return "", err
	}
	return filepath.Join(parent, filepath.Base(absolute)), nil
}

func validateNewArtifactDestinations(paths ...string) error {
	destinations := make([]string, 0, len(paths))
	for _, path := range paths {
		destination, err := canonicalDestination(path)
		if err != nil {
			return err
		}
		// #nosec G703 -- this is an intentional fail-closed lstat of a caller-selected local output destination.
		if _, err := os.Lstat(destination); err == nil || !os.IsNotExist(err) {
			return errors.New("artifact destination already exists or cannot be inspected")
		}
		for _, previous := range destinations {
			if destination == previous || filepath.Dir(destination) == filepath.Dir(previous) && strings.EqualFold(filepath.Base(destination), filepath.Base(previous)) {
				return errors.New("artifact destinations overlap")
			}
		}
		destinations = append(destinations, destination)
	}
	return nil
}

func zeroBytes(value []byte) {
	for index := range value {
		value[index] = 0
	}
}
