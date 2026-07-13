// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

const validInput = `{
  "schemaVersion":"cloudring.openbao-kubernetes-auth-plan/v2",
  "authMount":"kubernetes-consumer-example",
  "authMountOwnership":"dedicated-create-owned",
  "kvV2Mount":"cloudring",
  "dataPrefix":"services/cloudring-consumer-example",
  "policyName":"cloudring-consumer-example-kv-read",
  "roleName":"cloudring-consumer-example",
  "workloadIdentity":{"namespace":"cloudring-consumer-example","serviceAccount":"cloudring-openbao-reader"},
  "audience":"openbao",
  "aliasNameSource":"serviceaccount_uid",
  "tokenTTL":"10m",
  "tokenMaxTTL":"30m",
  "tokenNoDefaultPolicy":true
}`

func TestRunPlansValidStdinContract(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"plan", "kubernetes-auth"}, strings.NewReader(validInput), &stdout, &stderr)
	if code != exitPlanned || !strings.Contains(stdout.String(), `"status":"planned"`) || stderr.Len() != 0 {
		t.Fatalf("run() code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "cloudring-consumer-example") {
		t.Fatalf("run() reflected contract identifiers: %q", stdout.String())
	}
}

func TestRunUsesStableExitSemantics(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		stdin  io.Reader
		stdout io.Writer
		want   int
	}{
		{"usage", nil, strings.NewReader(""), io.Discard, exitUsage},
		{"extra arg", []string{"plan", "kubernetes-auth", "file.json"}, strings.NewReader(validInput), io.Discard, exitUsage},
		{"blocked", []string{"plan", "kubernetes-auth"}, strings.NewReader(`{"unexpected":"do-not-reflect"}`), io.Discard, exitBlocked},
		{"input unavailable", []string{"plan", "kubernetes-auth"}, cliErrorReader{}, io.Discard, exitUsage},
		{"output unavailable", []string{"plan", "kubernetes-auth"}, strings.NewReader(validInput), cliErrorWriter{}, exitUsage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if got := run(test.args, test.stdin, test.stdout, &stderr); got != test.want {
				t.Fatalf("run() exit = %d, want %d, stderr=%q", got, test.want, stderr.String())
			}
			if strings.Contains(stderr.String(), "do-not-reflect") {
				t.Fatalf("stderr reflected input: %q", stderr.String())
			}
		})
	}
}

type cliErrorReader struct{}

func (cliErrorReader) Read([]byte) (int, error) { return 0, errors.New("sensitive reader detail") }

type cliErrorWriter struct{}

func (cliErrorWriter) Write([]byte) (int, error) { return 0, errors.New("sensitive writer detail") }
