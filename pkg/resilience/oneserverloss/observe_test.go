// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package oneserverloss

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type fakeClock struct {
	now      time.Time
	sleeps   int
	gapSleep int
}

func (clock *fakeClock) Now() time.Time { return clock.now }

func (clock *fakeClock) Sleep(ctx context.Context, duration time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	clock.sleeps++
	if clock.sleeps == clock.gapSleep {
		duration *= 3
	}
	clock.now = clock.now.Add(duration)
	return nil
}

type fakeReader struct {
	index           int
	replacement     bool
	lossQuorum      int
	workloadFailure bool
	vmNeverMigrates bool
}

func (reader *fakeReader) IdentitySHA256() string { return testDigest("kubectl") }

func (reader *fakeReader) ReadyZ(context.Context) error {
	reader.index++
	return nil
}

func (reader *fakeReader) ListPage(_ context.Context, resource Resource, namespace, selector, continuation string, _ int) ([]byte, error) {
	if continuation != "" {
		return nil, errors.New("unexpected continuation")
	}
	state := reader.index - 1
	var items []any
	switch resource.Resource {
	case "nodes":
		for node := 1; node <= 3; node++ {
			ready := state < 3 || state >= 5 || node != 1
			if state >= 3 && state < 5 && reader.lossQuorum > 0 && node > reader.lossQuorum {
				ready = false
			}
			uid := fmt.Sprintf("node-uid-%d", node)
			if node == 1 && state >= 5 && reader.replacement {
				uid = "replacement-node-uid"
			}
			items = append(items, nodeFixture(fmt.Sprintf("node-%d", node), uid, ready))
		}
	case "pods":
		if namespace == "kube-system" {
			for node := 1; node <= 3; node++ {
				ready := state < 3 || state >= 5 || node != 1
				if state >= 3 && state < 5 && reader.lossQuorum > 0 && node > reader.lossQuorum {
					ready = false
				}
				items = append(items, podFixture("kube-system", fmt.Sprintf("control-%d", node), fmt.Sprintf("pod-uid-%d", node), fmt.Sprintf("node-%d", node), ready,
					map[string]string{"component": strings.TrimPrefix(selector, "component=")}))
			}
		} else {
			ready := !(reader.workloadFailure && state >= 3 && state < 5)
			items = append(items, podFixture(namespace, "service-0", "service-pod-uid", "node-2", ready,
				map[string]string{"app.kubernetes.io/name": "service"}))
		}
	default:
		return nil, errors.New("unexpected list resource")
	}
	return listFixture(resource, items), nil
}

func (reader *fakeReader) Get(_ context.Context, resource Resource, namespace, name string) ([]byte, error) {
	state := reader.index - 1
	switch resource.Resource {
	case "virtualmachines":
		ready := state != 3
		return objectFixture(map[string]any{
			"apiVersion": "kubevirt.io/v1", "kind": "VirtualMachine",
			"metadata": metadataFixture(namespace, name, "vm-uid"), "status": map[string]any{"ready": ready},
		}), nil
	case "virtualmachineinstances":
		if state == 3 {
			return nil, ErrNotFound
		}
		node := "node-1"
		if state >= 4 && !reader.vmNeverMigrates {
			node = "node-2"
		}
		return objectFixture(map[string]any{
			"apiVersion": "kubevirt.io/v1", "kind": "VirtualMachineInstance",
			"metadata": metadataFixture(namespace, name, "vmi-uid"),
			"status":   map[string]any{"phase": "Running", "nodeName": node, "conditions": []any{map[string]any{"type": "Ready", "status": "True"}}},
		}), nil
	default:
		return nil, errors.New("unexpected get resource")
	}
}

type fakeProbe struct {
	clock   *fakeClock
	driftAt int64
}

func (probe fakeProbe) IdentitySHA256() string { return testDigest("probe-adapter") }

func (probe fakeProbe) Observe(_ context.Context, request ProbeRequest) (ProbeResponse, error) {
	digest := testDigest("business-state")
	if request.Sequence == probe.driftAt {
		digest = testDigest("changed-business-state")
	}
	now := probe.clock.Now().UTC().Format(time.RFC3339Nano)
	return ProbeResponse{
		SchemaVersion: ProbeResponseSchemaVersion, Implementation: "postgresql-probe", Version: "v1", RequestSHA256: digestJSON(request),
		AdapterExecutableSHA256: probe.IdentitySHA256(), HashAlgorithm: "sha256", DataSHA256: digest, ValidatedBytes: 4096,
		StartedAt: now, CompletedAt: now,
	}, nil
}

type recordingBarrier struct {
	marker ReadyMarker
	err    error
}

func (barrier *recordingBarrier) ReadyForFault(_ context.Context, marker ReadyMarker) error {
	barrier.marker = marker
	return barrier.err
}

func TestObserveBuildsContinuousOfflineVerifiableReceipt(t *testing.T) {
	receipt, marker := runHappyObserver(t, &fakeReader{})
	if err := ValidateReadyMarker(marker); err != nil {
		t.Fatalf("ValidateReadyMarker: %v", err)
	}
	if err := ValidateReceipt(&receipt); err != nil {
		t.Fatalf("ValidateReceipt: %v", err)
	}
	schema, err := jsonschema.NewCompiler().Compile("../../../contracts/one-server-loss/receipt.schema.json")
	if err != nil {
		t.Fatalf("compile receipt schema: %v", err)
	}
	payload, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(payload))
	if err != nil || schema.Validate(instance) != nil {
		t.Fatalf("runtime receipt does not match public schema: %v", err)
	}
	if len(receipt.PreLoss.Samples) != 3 || len(receipt.Loss.Samples) != 2 || len(receipt.Recovered.Samples) != 3 {
		t.Fatalf("phase sample counts = %d/%d/%d, want 3/2/3", len(receipt.PreLoss.Samples), len(receipt.Loss.Samples), len(receipt.Recovered.Samples))
	}
	if receipt.Loss.Samples[len(receipt.Loss.Samples)-1].VM.VMIOnTarget {
		t.Fatal("VM did not migrate away from target in loss evidence")
	}
}

func TestObserveFailsClosedForContinuityViolations(t *testing.T) {
	tests := []struct {
		name   string
		reader *fakeReader
		probe  func(*fakeClock) fakeProbe
		gap    int
	}{
		{name: "target replacement", reader: &fakeReader{replacement: true}},
		{name: "quorum loss", reader: &fakeReader{lossQuorum: 1}},
		{name: "workload loss", reader: &fakeReader{workloadFailure: true}},
		{name: "VM misses SLO", reader: &fakeReader{vmNeverMigrates: true}},
		{name: "data drift", reader: &fakeReader{}, probe: func(clock *fakeClock) fakeProbe { return fakeProbe{clock: clock, driftAt: 5} }},
		{name: "sample gap", reader: &fakeReader{}, gap: 3},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			clock := newFakeClock()
			clock.gapSleep = test.gap
			probe := fakeProbe{clock: clock}
			if test.probe != nil {
				probe = test.probe(clock)
			}
			barrier := &recordingBarrier{}
			if _, err := Observe(context.Background(), test.reader, probe, barrier, validRequest(), clock); err == nil {
				t.Fatal("Observe accepted a continuity violation")
			}
		})
	}
}

func TestObserveStopsWhenReadyMarkerCannotBePublished(t *testing.T) {
	clock := newFakeClock()
	reader := &fakeReader{}
	barrier := &recordingBarrier{err: errors.New("synthetic marker failure")}
	if _, err := Observe(context.Background(), reader, fakeProbe{clock: clock}, barrier, validRequest(), clock); err == nil {
		t.Fatal("Observe continued after marker publication failure")
	}
	if reader.index != 3 {
		t.Fatalf("reader sample count = %d, want 3 pre-fault samples", reader.index)
	}
}

func TestValidateReceiptRejectsRehashedIdentityReplacement(t *testing.T) {
	receipt, _ := runHappyObserver(t, &fakeReader{})
	replacement := testDigest("replacement")
	for index := range receipt.Recovered.Samples {
		receipt.Recovered.Samples[index].TargetNodeUIDSHA256 = replacement
		receipt.Recovered.Samples[index].SampleSHA256 = sampleDigest(receipt.Recovered.Samples[index])
	}
	receipt.Recovered.SamplesSHA256 = digestJSON(receipt.Recovered.Samples)
	receipt.ReceiptSHA256 = receiptDigest(receipt)
	if err := ValidateReceipt(&receipt); err == nil {
		t.Fatal("ValidateReceipt accepted a rehashed target-node replacement")
	}
}

func TestValidateReceiptRejectsRehashedTimelineAndProbeReplay(t *testing.T) {
	receipt, _ := runHappyObserver(t, &fakeReader{})
	ready := mustTimestamp(receipt.ReadyMarkerAt).Add(500 * time.Millisecond)
	receipt.ReadyMarkerAt = ready.Format(time.RFC3339Nano)
	receipt.ReceiptSHA256 = receiptDigest(receipt)
	if err := ValidateReceipt(&receipt); err == nil {
		t.Fatal("ValidateReceipt accepted a marker time not backed by a sample")
	}

	receipt, _ = runHappyObserver(t, &fakeReader{})
	receipt.PreLoss.Samples[1].DataProbe.RequestSHA256 = receipt.PreLoss.Samples[0].DataProbe.RequestSHA256
	receipt.PreLoss.Samples[1].SampleSHA256 = sampleDigest(receipt.PreLoss.Samples[1])
	receipt.PreLoss.SamplesSHA256 = digestJSON(receipt.PreLoss.Samples)
	receipt.ReceiptSHA256 = receiptDigest(receipt)
	if err := ValidateReceipt(&receipt); err == nil {
		t.Fatal("ValidateReceipt accepted a replayed data-probe response")
	}
}

func TestListDecoderZeroesRawResponseAfterCopy(t *testing.T) {
	payload := []byte(`{"apiVersion":"v1","kind":"NodeList","metadata":{"resourceVersion":"1"},"items":[{"apiVersion":"v1","kind":"Node","metadata":{"name":"node-1","uid":"uid-1","resourceVersion":"1","labels":{"node-role.kubernetes.io/control-plane":""}},"status":{"conditions":[{"type":"Ready","status":"True"}]}}]}`)
	items, err := listAll(context.Background(), fuzzListReader{payload: payload}, nodeResource, "", "")
	if err != nil {
		t.Fatalf("listAll: %v", err)
	}
	for index, value := range payload {
		if value != 0 {
			t.Fatalf("raw payload byte %d was not zeroed", index)
		}
	}
	if len(items) != 1 {
		t.Fatalf("copied item count = %d, want 1", len(items))
	}
	if _, err := decodeNode(items[0]); err != nil {
		t.Fatalf("decode copied Node: %v", err)
	}
	zeroPayloads(items)
}

func TestValidateRequestRejectsUnsafeOrUnboundedBindings(t *testing.T) {
	request := validRequest()
	request.Workloads[0].ID = "Private Namespace"
	if _, err := validateRequest(request); err == nil {
		t.Fatal("validateRequest accepted unsafe public evidence ID")
	}
	request = validRequest()
	request.FaultArrivalTimeout = "60m0s"
	if _, err := validateRequest(request); err == nil {
		t.Fatal("validateRequest accepted non-canonical duration")
	}
}

func runHappyObserver(t *testing.T, reader *fakeReader) (Receipt, ReadyMarker) {
	t.Helper()
	clock := newFakeClock()
	barrier := &recordingBarrier{}
	receipt, err := Observe(context.Background(), reader, fakeProbe{clock: clock}, barrier, validRequest(), clock)
	if err != nil {
		t.Fatalf("Observe: %v", err)
	}
	return receipt, barrier.marker
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)}
}

func validRequest() Request {
	return Request{
		SchemaVersion: RequestSchemaVersion, RunNonceSHA256: testDigest("run-nonce"), TargetNodeName: "node-1",
		PollInterval: "1s", FaultArrivalTimeout: "5s", MinimumLossWindow: "2s", RecoveryTimeout: "10s", RecoveryStabilityWindow: "2s", MinimumControlPlaneMembers: 3,
		Workloads: []WorkloadTarget{{ID: "control-workload", Namespace: "service-system", MatchLabels: map[string]string{"app.kubernetes.io/name": "service"}, MinimumReadyPods: 1, MinimumDistinctReadyNodes: 1}},
		VM:        VMTarget{ID: "continuity-vm", Namespace: "virtualization-system", Name: "continuity-vm", RequirePreLossOnTarget: true, MaximumUnavailableDuration: "2s"},
		DataProbe: DataProbeTarget{ID: "business-state", QueryRef: "canonical-business-state", MinimumValidatedBytes: 1024},
	}
}

func listFixture(resource Resource, items []any) []byte {
	return objectFixture(map[string]any{
		"apiVersion": apiVersion(resource), "kind": resource.ListKind,
		"metadata": map[string]any{"resourceVersion": "100", "continue": ""}, "items": items,
	})
}

func nodeFixture(name, uid string, ready bool) map[string]any {
	status := "False"
	if ready {
		status = "True"
	}
	return map[string]any{
		"apiVersion": "v1", "kind": "Node", "metadata": map[string]any{
			"name": name, "uid": uid, "resourceVersion": "100", "labels": map[string]string{"node-role.kubernetes.io/control-plane": ""},
		}, "status": map[string]any{"conditions": []any{map[string]any{"type": "Ready", "status": status}}},
	}
}

func podFixture(namespace, name, uid, node string, ready bool, labels map[string]string) map[string]any {
	status, phase := "False", "Pending"
	if ready {
		status, phase = "True", "Running"
	}
	return map[string]any{
		"apiVersion": "v1", "kind": "Pod", "metadata": metadataWithLabelsFixture(namespace, name, uid, labels), "spec": map[string]any{"nodeName": node},
		"status": map[string]any{"phase": phase, "conditions": []any{map[string]any{"type": "Ready", "status": status}}},
	}
}

func metadataFixture(namespace, name, uid string) map[string]any {
	return map[string]any{"namespace": namespace, "name": name, "uid": uid, "resourceVersion": "100"}
}

func metadataWithLabelsFixture(namespace, name, uid string, labels map[string]string) map[string]any {
	metadata := metadataFixture(namespace, name, uid)
	metadata["labels"] = labels
	return metadata
}

func objectFixture(value any) []byte {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return payload
}

func testDigest(value string) string { return digestJSON(value) }
