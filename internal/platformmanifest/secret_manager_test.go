// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretManagerProfileIsStructurallyReady(t *testing.T) {
	root := repositoryRoot(t)
	report, err := VerifySecretManager(root)
	if err != nil {
		t.Fatalf("verify secret-manager profile: %v", err)
	}
	if report.Status != "ready" || report.Documents != 15 || len(report.Checks) != 6 {
		t.Fatalf("unexpected report: %#v", report)
	}
}

func TestSecretManagerProfileRejectsMutableImage(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "runtime", "openbao-release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("2.5.5@sha256:6150c4a6b62067db6141c8da7a6a6b5763f4f47c315343d0c848b40fecdfd452"), []byte("2.5.5"))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("mutable OpenBao image was accepted")
	}
}

func TestSecretManagerProfileRejectsDuplicateYAMLKeys(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "store", "platform-secrets.yaml")
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("kind: ClusterSecretStore\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("duplicate YAML key was accepted")
	}
}

func TestSecretManagerProfileRejectsDisabledListenerTLS(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "runtime", "openbao-release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("tls_disable        = 0"), []byte("tls_disable        = 1"))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("disabled OpenBao listener TLS was accepted")
	}
}

func TestSecretManagerProfileRejectsCommentMaskedListenerTLS(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "runtime", "openbao-release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("tls_disable        = 0"), []byte("# tls_disable        = 0\n              tls_disable        = 1"))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("comment-masked disabled OpenBao listener TLS was accepted")
	}
}

func TestSecretManagerProfileRejectsMissingListenerTLSSetting(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "runtime", "openbao-release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("              tls_disable        = 0\n"), nil)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("missing OpenBao listener TLS setting was accepted")
	}
}

func TestSecretManagerProfileRejectsMissingServingCertificate(t *testing.T) {
	root := copyProfile(t)
	path := filepath.Join(root, profilePath, "runtime", "tls.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("kind: Certificate\nmetadata:\n  name: openbao-server"), []byte("kind: ConfigMap\nmetadata:\n  name: openbao-server"))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("missing OpenBao serving Certificate was accepted")
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func copyProfile(t *testing.T) string {
	t.Helper()
	source := filepath.Join(repositoryRoot(t), profilePath)
	root := t.TempDir()
	destination := filepath.Join(root, profilePath)
	if err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	}); err != nil {
		t.Fatal(err)
	}
	return root
}

func replaceOnce(t *testing.T, data, old, replacement []byte) []byte {
	t.Helper()
	position := -1
	for index := 0; index+len(old) <= len(data); index++ {
		match := true
		for offset := range old {
			if data[index+offset] != old[offset] {
				match = false
				break
			}
		}
		if match {
			position = index
			break
		}
	}
	if position < 0 {
		t.Fatal("fixture token not found")
	}
	result := make([]byte, 0, len(data)-len(old)+len(replacement))
	result = append(result, data[:position]...)
	result = append(result, replacement...)
	result = append(result, data[position+len(old):]...)
	return result
}
