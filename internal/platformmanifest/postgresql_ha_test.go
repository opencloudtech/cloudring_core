// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPostgreSQLHAProfileIsStructurallyReady(t *testing.T) {
	report, err := VerifyPostgreSQLHA(repositoryRoot(t))
	if err != nil {
		t.Fatalf("verify PostgreSQL HA profile: %v", err)
	}
	if report.Status != "ready" || report.Files != 4 || report.Documents != 8 || len(report.Checks) != 10 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestPostgreSQLHAProfileRejectsUnsafeChanges(t *testing.T) {
	tests := []struct {
		name        string
		stage       string
		old         string
		replacement string
	}{
		{"active source controller", "controllers", "  suspend: true\n", "  suspend: false\n"},
		{"mutable operator image", "controllers", "      tag: 1.30.0@sha256:a2701eb97cdd2a34b1fdb2cb51987f544b706e40bec72ae7146cd8580efefebb\n", "      tag: latest\n"},
		{"one controller", "controllers", "    replicaCount: 2\n", "    replicaCount: 1\n"},
		{"permissive webhook", "controllers", "        failurePolicy: Fail\n", "        failurePolicy: Ignore\n"},
		{"two database instances", "runtime", "  instances: 3\n", "  instances: 2\n"},
		{"mutable database image", "runtime", "  imageName: ghcr.io/cloudnative-pg/postgresql:18.4@sha256:17760b4508e1703f7bf3d5ee6c9335ff6d1bf62b034a3d5b32a43414a516789f\n", "  imageName: ghcr.io/cloudnative-pg/postgresql:latest\n"},
		{"superuser enabled", "runtime", "  enableSuperuserAccess: false\n", "  enableSuperuserAccess: true\n"},
		{"application owns database", "runtime", "      owner: cloudring_owner\n", "      owner: cloudring_app\n"},
		{"application is superuser", "runtime", "        superuser: false\n", "        superuser: true\n"},
		{"application can create database", "runtime", "        createdb: false\n", "        createdb: true\n"},
		{"PDB disabled", "runtime", "  enablePDB: true\n", "  enablePDB: false\n"},
		{"soft anti-affinity", "runtime", "    podAntiAffinityType: required\n", "    podAntiAffinityType: preferred\n"},
		{"async durability", "runtime", "      dataDurability: required\n", "      dataDurability: preferred\n"},
		{"failover quorum disabled", "runtime", "      failoverQuorum: true\n", "      failoverQuorum: false\n"},
		{"non-TLS minimum", "runtime", "      ssl_min_protocol_version: TLSv1.3\n", "      ssl_min_protocol_version: TLSv1.0\n"},
		{"non-replicated storage", "runtime", "    storageClass: longhorn-replicated\n", "    storageClass: local-path\n"},
		{"delete snapshots", "runtime", "      snapshotOwnerReference: none\n", "      snapshotOwnerReference: cluster\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := copyPostgreSQLHAProfile(t)
			path := filepath.Join(postgresqlHAProfilePath, test.stage, "resources.yaml")
			data := readPostgreSQLHAFile(t, root, path)
			data = replaceOnce(t, data, []byte(test.old), []byte(test.replacement))
			writePostgreSQLHAFile(t, root, path, data)
			if _, err := VerifyPostgreSQLHA(root); err == nil {
				t.Fatalf("unsafe change %q was accepted", test.name)
			}
		})
	}
}

func TestPostgreSQLHAProfileRejectsDuplicateYAMLKeys(t *testing.T) {
	root := copyPostgreSQLHAProfile(t)
	path := filepath.Join(postgresqlHAProfilePath, "runtime", "resources.yaml")
	data := readPostgreSQLHAFile(t, root, path)
	data = append(data, []byte("kind: Cluster\n")...)
	writePostgreSQLHAFile(t, root, path, data)
	if _, err := VerifyPostgreSQLHA(root); err == nil {
		t.Fatal("duplicate YAML key was accepted")
	}
}

func copyPostgreSQLHAProfile(t *testing.T) string {
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
	for _, stage := range []string{"controllers", "runtime"} {
		for _, name := range []string{"kustomization.yaml", "resources.yaml"} {
			relative := filepath.Join(postgresqlHAProfilePath, stage, name)
			data, err := source.ReadFile(relative)
			if err != nil {
				t.Fatal(err)
			}
			if err := destination.MkdirAll(filepath.Dir(relative), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := destination.WriteFile(relative, data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	return root
}

func readPostgreSQLHAFile(t *testing.T, root, relative string) []byte {
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

func writePostgreSQLHAFile(t *testing.T, root, relative string, data []byte) {
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
