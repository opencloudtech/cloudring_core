// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLonghornThreeNodeProfileIsStructurallyReady(t *testing.T) {
	report, err := VerifyLonghornThreeNode(repositoryRoot(t))
	if err != nil {
		t.Fatalf("verify Longhorn three-node profile: %v", err)
	}
	if report.Status != "ready" || report.Files != 2 || report.Documents != 5 || len(report.Checks) != 8 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestLonghornThreeNodeProfileRejectsUnsafeChanges(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		replacement string
	}{
		{"mutable chart", "      version: 1.12.0\n", "      version: latest\n"},
		{"source activation", "  suspend: true\n", "  suspend: false\n"},
		{"chart default class", "      defaultClass: false\n", "      defaultClass: true\n"},
		{"two replicas", "      defaultReplicaCount: '{\"v1\":\"3\",\"v2\":\"3\"}'\n", "      defaultReplicaCount: '{\"v1\":\"2\",\"v2\":\"2\"}'\n"},
		{"soft node anti-affinity", "      replicaSoftAntiAffinity: false\n", "      replicaSoftAntiAffinity: true\n"},
		{"hard zone anti-affinity", "      replicaZoneSoftAntiAffinity: true\n", "      replicaZoneSoftAntiAffinity: false\n"},
		{"degraded creation", "      allowVolumeCreationWithDegradedAvailability: false\n", "      allowVolumeCreationWithDegradedAvailability: true\n"},
		{"usage metrics", "      allowCollectingLonghornUsageMetrics: false\n", "      allowCollectingLonghornUsageMetrics: true\n"},
		{"revision counter disabled", "      disableRevisionCounter: '{\"v1\":\"false\"}'\n", "      disableRevisionCounter: '{\"v1\":\"true\"}'\n"},
		{"v2 engine", "      v2DataEngine: false\n", "      v2DataEngine: true\n"},
		{"wrong provisioner", "provisioner: driver.longhorn.io\n", "provisioner: forbidden.example.csi\n"},
		{"immediate binding", "volumeBindingMode: WaitForFirstConsumer\n", "volumeBindingMode: Immediate\n"},
		{"delete snapshots", "deletionPolicy: Retain\n", "deletionPolicy: Delete\n"},
		{"ui ingress", "    ingress:\n      enabled: false\n", "    ingress:\n      enabled: true\n"},
		{"second Velero selector", "    app.kubernetes.io/part-of: cloudring-storage\n  annotations:\n    storageclass.kubernetes.io/is-default-class: \"false\"\n", "    app.kubernetes.io/part-of: cloudring-storage\n    velero.io/csi-volumesnapshot-class: \"true\"\n  annotations:\n    storageclass.kubernetes.io/is-default-class: \"false\"\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := copyLonghornThreeNodeProfile(t)
			data := readLonghornThreeNodeProfileFile(t, root, "runtime/resources.yaml")
			data = replaceOnce(t, data, []byte(test.old), []byte(test.replacement))
			writeLonghornThreeNodeProfileFile(t, root, "runtime/resources.yaml", data)
			if _, err := VerifyLonghornThreeNode(root); err == nil {
				t.Fatalf("unsafe change %q was accepted", test.name)
			}
		})
	}
}

func TestLonghornThreeNodeProfileRejectsDuplicateYAMLKeys(t *testing.T) {
	root := copyLonghornThreeNodeProfile(t)
	data := readLonghornThreeNodeProfileFile(t, root, "runtime/resources.yaml")
	data = append(data, []byte("driver: driver.longhorn.io\n")...)
	writeLonghornThreeNodeProfileFile(t, root, "runtime/resources.yaml", data)
	if _, err := VerifyLonghornThreeNode(root); err == nil {
		t.Fatal("duplicate YAML key was accepted")
	}
}

func copyLonghornThreeNodeProfile(t *testing.T) string {
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
	for _, relative := range []string{"runtime/kustomization.yaml", "runtime/resources.yaml"} {
		sourcePath := filepath.Join(longhornThreeNodeProfilePath, relative)
		data, err := source.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		destinationPath := filepath.Join(longhornThreeNodeProfilePath, relative)
		if err := destination.MkdirAll(filepath.Dir(destinationPath), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := destination.WriteFile(destinationPath, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readLonghornThreeNodeProfileFile(t *testing.T, root, relative string) []byte {
	t.Helper()
	repository, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	data, err := repository.ReadFile(filepath.Join(longhornThreeNodeProfilePath, relative))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeLonghornThreeNodeProfileFile(t *testing.T, root, relative string, data []byte) {
	t.Helper()
	repository, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if err := repository.WriteFile(filepath.Join(longhornThreeNodeProfilePath, relative), data, 0o600); err != nil {
		t.Fatal(err)
	}
}
