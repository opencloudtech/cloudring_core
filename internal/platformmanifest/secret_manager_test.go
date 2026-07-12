// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("2.5.5@sha256:6150c4a6b62067db6141c8da7a6a6b5763f4f47c315343d0c848b40fecdfd452"), []byte("2.5.5"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("mutable OpenBao image was accepted")
	}
}

func TestSecretManagerProfileRejectsDuplicateYAMLKeys(t *testing.T) {
	root := copyProfile(t)
	profile, err := os.OpenRoot(filepath.Join(root, profilePath))
	if err != nil {
		t.Fatal(err)
	}
	defer profile.Close()
	file, err := profile.OpenFile(filepath.Join("store", "platform-secrets.yaml"), os.O_APPEND|os.O_WRONLY, 0)
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
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("tls_disable        = 0"), []byte("tls_disable        = 1"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("disabled OpenBao listener TLS was accepted")
	}
}

func TestSecretManagerProfileRejectsCommentMaskedListenerTLS(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("tls_disable        = 0"), []byte("# tls_disable        = 0\n              tls_disable        = 1"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("comment-masked disabled OpenBao listener TLS was accepted")
	}
}

func TestSecretManagerProfileRejectsMissingListenerTLSSetting(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("              tls_disable        = 0\n"), nil)
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("missing OpenBao listener TLS setting was accepted")
	}
}

func TestSecretManagerProfileRejectsMissingPersistentAudit(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte(persistentAuditHCL), nil)
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "persistent audit") {
		t.Fatalf("missing OpenBao persistent audit was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsUnsafePersistentAudit(t *testing.T) {
	tests := []struct {
		name        string
		old         string
		replacement string
	}{
		{name: "description", old: `description = "CloudRING persistent audit"`, replacement: `description = "Unreviewed audit"`},
		{name: "path", old: `file_path = "/openbao/audit/audit.log"`, replacement: `file_path = "/tmp/audit.log"`},
		{name: "mode", old: `mode      = "0600"`, replacement: `mode      = "0666"`},
		{name: "raw payloads", old: `log_raw   = "false"`, replacement: `log_raw   = "true"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := copyProfile(t)
			data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			data = replaceOnce(t, data, []byte(test.old), []byte(test.replacement))
			if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
				t.Fatal(err)
			}
			if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "persistent audit") {
				t.Fatalf("unsafe OpenBao persistent audit was accepted: %v", err)
			}
		})
	}
}

func TestSecretManagerProfileRejectsMissingServingCertificate(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "tls.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("kind: Certificate\nmetadata:\n  name: openbao-server"), []byte("kind: ConfigMap\nmetadata:\n  name: openbao-server"))
	if err := writeProfileFile(root, filepath.Join("runtime", "tls.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil {
		t.Fatal("missing OpenBao serving Certificate was accepted")
	}
}

func TestOpenBaoReadinessPostRendererProducesExactVerifiedTLSCommand(t *testing.T) {
	data, err := readProfileFile(repositoryRoot(t), filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	objects, err := decodeObjects(data)
	if err != nil {
		t.Fatal(err)
	}
	var release map[string]any
	for _, item := range objects {
		if item.Kind == "HelmRelease" && item.Name == "openbao" {
			release = item.Data
		}
	}
	command, err := openBaoReadinessPostRenderCommand(release)
	if err != nil {
		t.Fatalf("validate OpenBao readiness post-renderer: %v", err)
	}
	want := []string{"/bin/sh", "-ec", openBaoReadinessShellCommand}
	if strings.Join(command, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("unexpected rendered readiness command: %#v", command)
	}
	if strings.Contains(strings.Join(command, " "), "tls-skip-verify") {
		t.Fatalf("rendered readiness command disables TLS verification: %#v", command)
	}
}

func TestSecretManagerProfileRejectsClusterStoreOutsidePrivilegedNamespace(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("store", "platform-secrets.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("    - namespaces:\n        - external-secrets"), []byte("    - namespaces:\n        - tenant-workload"))
	if err := writeProfileFile(root, filepath.Join("store", "platform-secrets.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "privileged namespace boundary") {
		t.Fatalf("unsafe ClusterSecretStore namespace boundary was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsNonActiveClusterStoreService(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("store", "platform-secrets.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("server: https://openbao-active.openbao.svc:8200"), []byte("server: https://openbao.openbao.svc:8200"))
	if err := writeProfileFile(root, filepath.Join("store", "platform-secrets.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "workload identity contract") {
		t.Fatalf("non-active ClusterSecretStore service was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsDisabledAuthDelegator(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("      authDelegator:\n        enabled: true"), []byte("      authDelegator:\n        enabled: false"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "auth-delegator") {
		t.Fatalf("disabled OpenBao auth delegator was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsWidenedExternalSecretsIngress(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "network-policy.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte(`          podSelector:
            matchLabels:
              app.kubernetes.io/name: external-secrets
              app.kubernetes.io/instance: external-secrets
`), nil)
	if err := writeProfileFile(root, filepath.Join("runtime", "network-policy.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "NetworkPolicy") {
		t.Fatalf("namespace-wide External Secrets ingress was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsOrderedReadyOpenBaoBootstrap(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("      podManagementPolicy: Parallel"), []byte("      podManagementPolicy: OrderedReady"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "TLS HA Raft contract") {
		t.Fatalf("OrderedReady OpenBao bootstrap was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsPodIPHAAddress(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("        apiAddr: https://openbao-active.openbao.svc:8200"), []byte("        apiAddr: https://$(POD_IP):8200"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "TLS HA Raft contract") {
		t.Fatalf("Pod-IP OpenBao HA address was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsReadinessTLSSkipVerify(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte("                    exec bao status\n"), []byte("                    exec bao status -tls-skip-verify\n"))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "TLS verification") {
		t.Fatalf("insecure OpenBao readiness probe was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsReadinessWithoutCA(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte(`BAO_CACERT="/openbao/tls/client/ca.crt"`), []byte(`BAO_CACERT=""`))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "CA and pod DNS verification") {
		t.Fatalf("OpenBao readiness probe without CA was accepted: %v", err)
	}
}

func TestSecretManagerProfileRejectsReadinessWithoutPodSpecificServerName(t *testing.T) {
	root := copyProfile(t)
	data, err := readProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	data = replaceOnce(t, data, []byte(`export BAO_TLS_SERVER_NAME="${pod_dns}"`), []byte(`export BAO_TLS_SERVER_NAME="openbao.openbao.svc"`))
	if err := writeProfileFile(root, filepath.Join("runtime", "openbao-release.yaml"), data); err != nil {
		t.Fatal(err)
	}
	if _, err := VerifySecretManager(root); err == nil || !strings.Contains(err.Error(), "CA and pod DNS verification") {
		t.Fatalf("OpenBao readiness probe without pod-specific server name was accepted: %v", err)
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
	sourceRoot, err := os.OpenRoot(source)
	if err != nil {
		t.Fatal(err)
	}
	defer sourceRoot.Close()
	destination := filepath.Join(root, profilePath)
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	destinationRoot, err := os.OpenRoot(destination)
	if err != nil {
		t.Fatal(err)
	}
	defer destinationRoot.Close()
	for _, relative := range []string{
		"controllers/kustomization.yaml", "controllers/namespaces.yaml", "controllers/releases.yaml", "controllers/repositories.yaml",
		"runtime/kustomization.yaml", "runtime/network-policy.yaml", "runtime/openbao-release.yaml", "runtime/tls.yaml",
		"store/kustomization.yaml", "store/platform-secrets.yaml",
	} {
		data, err := sourceRoot.ReadFile(relative)
		if err != nil {
			t.Fatal(err)
		}
		if err := destinationRoot.MkdirAll(filepath.Dir(relative), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := destinationRoot.WriteFile(relative, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func readProfileFile(root, relative string) ([]byte, error) {
	profile, err := os.OpenRoot(filepath.Join(root, profilePath))
	if err != nil {
		return nil, err
	}
	defer profile.Close()
	return profile.ReadFile(relative)
}

func writeProfileFile(root, relative string, data []byte) error {
	profile, err := os.OpenRoot(filepath.Join(root, profilePath))
	if err != nil {
		return err
	}
	defer profile.Close()
	return profile.WriteFile(relative, data, 0o600)
}

func replaceOnce(t *testing.T, data, old, replacement []byte) []byte {
	t.Helper()
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
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

const persistentAuditHCL = `            audit "file" "persistent" {
              description = "CloudRING persistent audit"
              options {
                file_path = "/openbao/audit/audit.log"
                mode      = "0600"
                log_raw   = "false"
              }
            }

`
