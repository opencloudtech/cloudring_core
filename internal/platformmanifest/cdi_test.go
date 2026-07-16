// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCDIProfileIsStructurallyReady(t *testing.T) {
	report, err := VerifyCDI(repositoryRoot(t))
	if err != nil {
		t.Fatalf("verify CDI profile: %v", err)
	}
	if report.Status != "ready" || report.Files != 5 || report.Documents != 9 || len(report.Checks) != 7 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestCDIProfileRejectsUnsafeChanges(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		old         string
		replacement string
	}{
		{"active base operator", "controllers/upstream-cdi-operator-v1.65.0.yaml", "  replicas: 0\n", "  replicas: 1\n"},
		{"mutable operator image", "controllers/upstream-cdi-operator-v1.65.0.yaml", "quay.io/kubevirt/cdi-operator:v1.65.0@sha256:42ce149c020523b466cd8cb5e413bad9800d93f502d82ced69a2d98a01944ce5", "quay.io/kubevirt/cdi-operator:latest"},
		{"disabled activation", "activation/kustomization.yaml", "        value: 1\n", "        value: 0\n"},
		{"wrong activation target", "activation/kustomization.yaml", "      name: cdi-operator\n", "      name: other-operator\n"},
		{"unsafe removal", "runtime/resources.yaml", "  uninstallStrategy: BlockUninstallIfWorkloadsExist\n", "  uninstallStrategy: RemoveWorkloads\n"},
		{"delayed binding disabled", "runtime/resources.yaml", "      - HonorWaitForFirstConsumer\n", "      - OtherGate\n"},
		{"non-linux workload", "runtime/resources.yaml", "  workload:\n    nodeSelector:\n      kubernetes.io/os: linux\n", "  workload:\n    nodeSelector:\n      kubernetes.io/os: windows\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := copyCDIProfile(t)
			path := filepath.Join(cdiProfilePath, test.path)
			data := readCDIFile(t, root, path)
			data = replaceOnce(t, data, []byte(test.old), []byte(test.replacement))
			writeCDIFile(t, root, path, data)
			if _, err := VerifyCDI(root); err == nil {
				t.Fatalf("unsafe change %q was accepted", test.name)
			}
		})
	}
}

func TestCDIProfileRejectsDuplicateYAMLKeys(t *testing.T) {
	root := copyCDIProfile(t)
	path := filepath.Join(cdiProfilePath, "runtime", "resources.yaml")
	data := readCDIFile(t, root, path)
	data = append(data, []byte("spec: {}\n")...)
	writeCDIFile(t, root, path, data)
	if _, err := VerifyCDI(root); err == nil {
		t.Fatal("duplicate YAML key was accepted")
	}
}

func copyCDIProfile(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	source, err := os.OpenRoot(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	destination, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	for _, relative := range []string{
		"controllers/kustomization.yaml",
		"controllers/upstream-cdi-operator-v1.65.0.yaml",
		"activation/kustomization.yaml",
		"runtime/kustomization.yaml",
		"runtime/resources.yaml",
	} {
		path := filepath.Join(cdiProfilePath, relative)
		data, err := source.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := destination.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := destination.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readCDIFile(t *testing.T, root, relative string) []byte {
	t.Helper()
	repository, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	data, err := repository.ReadFile(relative)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeCDIFile(t *testing.T, root, relative string, data []byte) {
	t.Helper()
	repository, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if err := repository.WriteFile(relative, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
