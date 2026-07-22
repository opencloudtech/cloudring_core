// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package velero118

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/opencloudtech/CloudRING/pkg/backup/restoreproof"
)

func TestObserveDataUploadResultEstablishesWatchBeforeRestore(t *testing.T) {
	clock := &fakeClock{now: mustTime(t, "2026-07-14T12:00:00Z")}
	request, restoreObject, resultConfigMap := observationFixture(t)
	barrier := &observationReadyBarrier{}
	reader := &observationReader{
		initialList: listObject(t, "v1", "ConfigMapList", "10", "", nil),
		restore:     restoreObject,
		events: []WatchEvent{
			{Type: "ADDED", Object: resultConfigMap},
			{Type: "MODIFIED", Object: resultConfigMap},
			{Type: "DELETED", Object: withDeletionTimestamp(resultConfigMap)},
		},
		nextResourceVersion: "13",
		onWatch: func() {
			if barrier.calls != 1 {
				t.Fatalf("watch started before readiness publication: calls=%d", barrier.calls)
			}
			clock.now = mustTime(t, "2026-07-14T12:00:03Z")
		},
	}

	observation, err := ObserveDataUploadResult(t.Context(), reader, barrier, request, time.Minute, time.Second, clock)
	if err != nil {
		t.Fatalf("ObserveDataUploadResult() error = %v", err)
	}
	if barrier.calls != 1 || reader.watchCalls != 1 || observation.EventType != "ADDED" || observation.ObservedAt != "2026-07-14T12:00:02Z" ||
		observation.RequestSHA256 != requestJSONSHA256(request) || observation.ObjectSHA256 != canonicalJSONSHA256(resultConfigMap) ||
		observation.EvidenceSHA256 != dataUploadResultObservationEvidenceSHA256(observation) {
		t.Fatalf("observation binding is incomplete: barrier=%d watch=%d observation=%#v", barrier.calls, reader.watchCalls, observation)
	}
	if err := validateDataUploadResultObservation(observation); err != nil {
		t.Fatalf("generated observation is invalid: %v", err)
	}
}

func TestObserveDataUploadResultFailsClosedOnWatchGaps(t *testing.T) {
	request, restoreObject, resultConfigMap := observationFixture(t)
	cases := []struct {
		name        string
		preexisting bool
		initial     []byte
		events      []WatchEvent
		watchErr    error
		barrierErr  error
	}{
		{name: "restore already exists", preexisting: true},
		{name: "initial result already exists", initial: listObject(t, "v1", "ConfigMapList", "10", "", []json.RawMessage{resultConfigMap})},
		{name: "expired resource version", watchErr: errors.New("resourceVersion expired")},
		{name: "modified without added", events: []WatchEvent{{Type: "MODIFIED", Object: resultConfigMap}}},
		{name: "duplicate added", events: []WatchEvent{{Type: "ADDED", Object: resultConfigMap}, {Type: "ADDED", Object: resultConfigMap}}},
		{name: "readiness publication fails", barrierErr: errors.New("synthetic readiness failure")},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			clock := &fakeClock{now: mustTime(t, "2026-07-14T12:00:00Z")}
			initial := test.initial
			if initial == nil {
				initial = listObject(t, "v1", "ConfigMapList", "10", "", nil)
			}
			barrier := &observationReadyBarrier{err: test.barrierErr}
			reader := &observationReader{
				preexistingRestore:  test.preexisting,
				initialList:         initial,
				restore:             restoreObject,
				events:              test.events,
				watchErr:            test.watchErr,
				nextResourceVersion: "11",
				onWatch:             func() { clock.now = mustTime(t, "2026-07-14T12:00:03Z") },
			}
			if _, err := ObserveDataUploadResult(t.Context(), reader, barrier, request, time.Minute, time.Second, clock); err == nil {
				t.Fatal("unsafe DataUploadResult observation unexpectedly passed")
			}
			if test.preexisting && barrier.calls != 0 || test.barrierErr != nil && reader.watchCalls != 0 {
				t.Fatalf("fail-closed ordering was violated: barrier=%d watch=%d", barrier.calls, reader.watchCalls)
			}
		})
	}
}

func TestValidateDataUploadResultObservationRejectsInvalidCaptureTimeline(t *testing.T) {
	request, restoreObject, resultConfigMap := observationFixture(t)
	clock := &fakeClock{now: mustTime(t, "2026-07-14T12:00:00Z")}
	barrier := &observationReadyBarrier{}
	reader := &observationReader{
		initialList:         listObject(t, "v1", "ConfigMapList", "10", "", nil),
		restore:             restoreObject,
		events:              []WatchEvent{{Type: "ADDED", Object: resultConfigMap}},
		nextResourceVersion: "11",
		onWatch:             func() { clock.now = mustTime(t, "2026-07-14T12:00:03Z") },
	}
	valid, err := ObserveDataUploadResult(t.Context(), reader, barrier, request, time.Minute, time.Second, clock)
	if err != nil {
		t.Fatalf("ObserveDataUploadResult() error = %v", err)
	}
	for _, test := range []struct {
		name       string
		capturedAt string
	}{
		{name: "capture before observation", capturedAt: "2026-07-14T12:00:01Z"},
		{name: "capture too late", capturedAt: "2026-07-14T12:00:33.000000001Z"},
	} {
		t.Run(test.name, func(t *testing.T) {
			observation := valid
			observation.CapturedAt = test.capturedAt
			observation.EvidenceSHA256 = dataUploadResultObservationEvidenceSHA256(observation)
			if err := validateDataUploadResultObservation(observation); err == nil {
				t.Fatal("invalid DataUploadResult capture timeline unexpectedly passed")
			}
		})
	}
}

type observationReadyBarrier struct {
	calls  int
	notice DataUploadResultObservationReady
	err    error
}

func (barrier *observationReadyBarrier) ReadyForRestore(_ context.Context, notice DataUploadResultObservationReady) error {
	barrier.calls++
	barrier.notice = notice
	return barrier.err
}

type observationReader struct {
	preexistingRestore  bool
	initialList         []byte
	restore             []byte
	events              []WatchEvent
	nextResourceVersion string
	watchErr            error
	onWatch             func()
	watchCalls          int
	watched             bool
}

func (reader *observationReader) Get(_ context.Context, gvr restoreproof.GVR, namespace, name string) ([]byte, error) {
	if gvr != restoreproof.VeleroV1RestoreGVR || namespace != "velero" || name != "restore-copy" {
		return nil, errors.New("unexpected observation read")
	}
	if !reader.preexistingRestore && !reader.watched {
		return nil, ErrNotFound
	}
	return append([]byte(nil), reader.restore...), nil
}

func (reader *observationReader) ListPage(_ context.Context, gvr restoreproof.GVR, namespace, selector, continuation string, limit int) ([]byte, error) {
	if gvr != restoreproof.CoreV1CMGVR || namespace != "velero" || selector != "velero.io/pvc-namespace-name=source.volume,velero.io/resource-usage=DataUpload" || continuation != "" || limit != 100 {
		return nil, errors.New("unexpected observation list")
	}
	return append([]byte(nil), reader.initialList...), nil
}

func (reader *observationReader) WatchPage(_ context.Context, gvr restoreproof.GVR, namespace, selector, resourceVersion string, timeoutSeconds int) ([]WatchEvent, string, error) {
	reader.watchCalls++
	reader.watched = true
	if reader.onWatch != nil {
		reader.onWatch()
	}
	if gvr != restoreproof.CoreV1CMGVR || namespace != "velero" || selector != "velero.io/pvc-namespace-name=source.volume,velero.io/resource-usage=DataUpload" || resourceVersion != "10" || timeoutSeconds != 5 {
		return nil, "", errors.New("unexpected observation watch")
	}
	return append([]WatchEvent(nil), reader.events...), reader.nextResourceVersion, reader.watchErr
}

func (reader *observationReader) ConfirmAbsent(_ context.Context, gvr restoreproof.GVR, namespace, name string) (bool, error) {
	if gvr != restoreproof.VeleroV1RestoreGVR || namespace != "velero" || name != "restore-copy" {
		return false, errors.New("unexpected observation absence confirmation")
	}
	return !reader.preexistingRestore, nil
}

func observationFixture(t *testing.T) (DataUploadResultObservationRequest, []byte, []byte) {
	t.Helper()
	restoreUID := "restore-uid"
	request := DataUploadResultObservationRequest{
		SchemaVersion:   DataUploadResultObservationRequestSchemaVersion,
		VeleroNamespace: "velero", RestoreName: "restore-copy", SourceNamespace: "source", SourcePVC: "volume",
		DataUploadName: "backup-volume-1", EvidencePrefix: "runtime/task22-observation",
	}
	restoreObject := kubeObject(t, "velero.io/v1", "Restore", metadata("restore-copy", "velero", restoreUID, "20", nil, nil), map[string]any{
		"backupName": "backup-direct", "scheduleName": "", "namespaceMapping": map[string]string{"source": "target"},
	}, map[string]any{"phase": "InProgress", "startTimestamp": "2026-07-14T12:00:01Z", "errors": 0, "warnings": 0}, nil)
	resultMetadata := metadata("backup-volume-1-result", "velero", "result-cm-uid", "11", nil, map[string]string{
		"velero.io/restore-uid": restoreUID, "velero.io/pvc-namespace-name": "source.volume", "velero.io/resource-usage": "DataUpload",
	})
	resultMetadata["creationTimestamp"] = "2026-07-14T12:00:02Z"
	resultPayload := string(mustJSON(t, map[string]any{
		"backupStorageLocation": "offcell", "datamover": "", "snapshotID": "snapshot-id", "sourceNamespace": "source",
		"dataMoverResult": map[string]string{}, "nodeOS": "linux", "snapshotSize": 4096,
	}))
	resultConfigMap := kubeObject(t, "v1", "ConfigMap", resultMetadata, nil, nil, map[string]string{restoreUID: resultPayload})
	return request, restoreObject, resultConfigMap
}
