// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package drill

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
	"github.com/opencloudtech/CloudRING/pkg/kubeconfigpipe"
	"github.com/opencloudtech/CloudRING/pkg/secureexec"
)

const (
	maxAdapterStdout = 1 << 20
	maxAdapterStderr = 32 << 10
)

type adapterRunner interface {
	IdentitySHA256() string
	RunWithEnvironment(context.Context, []string, []byte, int64, int64, []string, *kubeconfigpipe.Replay) ([]byte, []byte, error)
	Close() error
}

type Adapter struct {
	runner adapterRunner
	replay *kubeconfigpipe.Replay
}

func PinAdapter(path string, timeout time.Duration, replay *kubeconfigpipe.Replay) (*Adapter, error) {
	if replay == nil {
		return nil, errors.New("backup drill adapter requires pipe-backed kubeconfig replay")
	}
	runner, err := secureexec.PinAbsolute(path, timeout)
	if err != nil {
		return nil, errors.New("pin backup drill adapter")
	}
	return &Adapter{runner: runner, replay: replay}, nil
}

func (adapter *Adapter) IdentitySHA256() string {
	if adapter == nil || adapter.runner == nil {
		return ""
	}
	return adapter.runner.IdentitySHA256()
}

func (adapter *Adapter) Close() error {
	if adapter == nil || adapter.runner == nil {
		return nil
	}
	return adapter.runner.Close()
}

func (adapter *Adapter) invoke(ctx context.Context, request AdapterRequest) (AdapterResponse, error) {
	if adapter == nil || adapter.runner == nil || adapter.IdentitySHA256() != request.AdapterExecutableSHA256 || request.RequestSHA256 != AdapterRequestSHA256(request) {
		return AdapterResponse{}, errors.New("backup drill adapter request identity is invalid")
	}
	input, err := json.Marshal(request)
	if err != nil {
		return AdapterResponse{}, errors.New("encode backup drill adapter request")
	}
	defer zero(input)
	stdout, stderr, err := adapter.runner.RunWithEnvironment(ctx, []string{"drill"}, input, maxAdapterStdout, maxAdapterStderr, []string{"LANG=C", "LC_ALL=C"}, adapter.replay)
	defer zero(stdout)
	defer zero(stderr)
	if err != nil {
		return AdapterResponse{}, errors.New("backup drill adapter execution failed")
	}
	if containsSecretLookingMaterial(stdout) || containsSecretLookingMaterial(stderr) {
		return AdapterResponse{}, errors.New("backup drill adapter emitted forbidden credential-like material")
	}
	var response AdapterResponse
	if strictjson.DecodeExact(stdout, &response) != nil {
		return AdapterResponse{}, errors.New("backup drill adapter response is invalid JSON")
	}
	if err := ValidateAdapterResponse(request, response); err != nil {
		return response, err
	}
	return response, nil
}

func containsSecretLookingMaterial(value []byte) bool {
	lower := bytes.ToLower(value)
	for _, forbidden := range [][]byte{
		[]byte(`"password"`), []byte(`"token"`), []byte(`"secret"`), []byte(`"authorization"`),
		[]byte(`"cookie"`), []byte("-----begin private key-----"), []byte("-----begin rsa private key-----"),
	} {
		if bytes.Contains(lower, forbidden) {
			return true
		}
	}
	return false
}
