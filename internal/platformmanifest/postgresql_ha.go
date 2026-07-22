// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const postgresqlHAProfilePath = "deploy/kubernetes/postgresql-ha"

// VerifyPostgreSQLHA validates the reusable HA database source contract. It
// deliberately does not turn structural checks into a live durability claim.
func VerifyPostgreSQLHA(root string) (Report, error) {
	root, err := canonicalRoot(root)
	if err != nil {
		return Report{}, err
	}
	report := Report{Status: "blocked", Profile: "cloudring-postgresql-ha/v1"}
	repository, err := os.OpenRoot(root)
	if err != nil {
		return report, errors.New("open confined repository root")
	}
	defer repository.Close()

	var objects []object
	for _, stage := range []string{"controllers", "runtime"} {
		stageObjects, files, readErr := readPostgreSQLHAStage(repository, stage)
		if readErr != nil {
			return report, readErr
		}
		report.Files += files
		objects = append(objects, stageObjects...)
	}
	report.Documents = len(objects)
	if report.Files != 4 || report.Documents != 8 {
		return report, errors.New("PostgreSQL HA source inventory is incomplete")
	}
	index, err := exactPostgreSQLHAInventory(objects)
	if err != nil {
		return report, err
	}
	if err := validatePostgreSQLHAControllers(index); err != nil {
		return report, err
	}
	if err := validatePostgreSQLHARuntime(index); err != nil {
		return report, err
	}
	report.Status = "ready"
	report.Checks = []string{
		"source_controller_suspended",
		"controller_chart_and_image_pinned",
		"admission_webhooks_fail_closed",
		"controller_replicas_and_disruption_budget_ready",
		"three_postgresql_instances_hard_host_separated",
		"quorum_synchronous_durability_required",
		"tls_scram_and_least_privilege_roles_ready",
		"replicated_pgdata_and_wal_ready",
		"retained_online_snapshot_schedule_ready",
		"live_backup_restore_and_failover_non_claim_preserved",
	}
	return report, nil
}

func readPostgreSQLHAStage(root *os.Root, stage string) ([]object, int, error) {
	directory := filepath.Join(postgresqlHAProfilePath, stage)
	kustomization, err := readRegular(root, filepath.Join(directory, "kustomization.yaml"))
	if err != nil {
		return nil, 0, err
	}
	var manifest map[string]any
	if err := decodeOne(kustomization, &manifest); err != nil ||
		!exactMappingKeys(manifest, "apiVersion", "kind", "resources") ||
		nestedString(manifest, "apiVersion") != "kustomize.config.k8s.io/v1beta1" ||
		nestedString(manifest, "kind") != "Kustomization" ||
		!exactStringSequence(manifest["resources"], "resources.yaml") {
		return nil, 0, fmt.Errorf("PostgreSQL HA %s stage has an invalid kustomization", stage)
	}
	data, err := readRegular(root, filepath.Join(directory, "resources.yaml"))
	if err != nil {
		return nil, 0, err
	}
	for _, forbidden := range [][]byte{
		[]byte(":latest"), []byte("REPLACE_WITH"), []byte("example.invalid"),
		[]byte("kind: Secret"), []byte("postgres://"), []byte("sslmode=disable"),
	} {
		if bytes.Contains(data, forbidden) {
			return nil, 0, errors.New("PostgreSQL HA source contains mutable, unresolved, or secret material")
		}
	}
	objects, err := decodeObjects(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode PostgreSQL HA %s stage: %w", stage, err)
	}
	return objects, 2, nil
}

func exactPostgreSQLHAInventory(objects []object) (map[string]object, error) {
	index := make(map[string]object, len(objects))
	for _, item := range objects {
		key := item.Kind + "/" + item.Namespace + "/" + item.Name
		if _, duplicate := index[key]; duplicate {
			return nil, errors.New("duplicate PostgreSQL HA manifest object identity")
		}
		index[key] = item
	}
	expected := []string{
		"Cluster/cloudring-database/cloudring-postgres",
		"HelmRelease/cnpg-system/cnpg",
		"HelmRepository/cnpg-system/cnpg",
		"Namespace//cloudring-database",
		"Namespace//cnpg-system",
		"NetworkPolicy/cloudring-database/cloudring-postgres-ingress",
		"PodDisruptionBudget/cnpg-system/cnpg-controller-manager",
		"ScheduledBackup/cloudring-database/cloudring-postgres-volume-snapshot",
	}
	actual := make([]string, 0, len(index))
	for key := range index {
		actual = append(actual, key)
	}
	sort.Strings(actual)
	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		return nil, errors.New("PostgreSQL HA object inventory is not exact")
	}
	return index, nil
}

func validatePostgreSQLHAControllers(index map[string]object) error {
	repository := index["HelmRepository/cnpg-system/cnpg"].Data
	if nestedString(repository, "spec", "url") != "https://cloudnative-pg.github.io/charts" ||
		nestedString(repository, "spec", "interval") != "1h" {
		return errors.New("CloudNativePG chart repository is invalid")
	}
	release := index["HelmRelease/cnpg-system/cnpg"].Data
	if !exactBool(release, true, "spec", "suspend") ||
		nestedString(release, "spec", "releaseName") != "cnpg" ||
		nestedString(release, "spec", "chart", "spec", "chart") != "cloudnative-pg" ||
		nestedString(release, "spec", "chart", "spec", "version") != "0.29.0" ||
		nestedString(release, "spec", "chart", "spec", "sourceRef", "name") != "cnpg" ||
		nestedString(release, "spec", "chart", "spec", "sourceRef", "namespace") != "cnpg-system" ||
		nestedString(release, "metadata", "annotations", "cloudring.org/non-claim") == "" {
		return errors.New("CloudNativePG release boundary is invalid")
	}
	if nestedNumber(release, "spec", "values", "replicaCount") != 2 ||
		nestedString(release, "spec", "values", "image", "repository") != "ghcr.io/cloudnative-pg/cloudnative-pg" ||
		nestedString(release, "spec", "values", "image", "tag") != "1.30.0@sha256:a2701eb97cdd2a34b1fdb2cb51987f544b706e40bec72ae7146cd8580efefebb" ||
		!exactBool(release, true, "spec", "values", "crds", "create") ||
		nestedString(release, "spec", "values", "webhook", "mutating", "failurePolicy") != "Fail" ||
		nestedString(release, "spec", "values", "webhook", "validating", "failurePolicy") != "Fail" ||
		!exactBool(release, true, "spec", "values", "config", "clusterWide") ||
		!exactBool(release, false, "spec", "values", "rbac", "aggregateClusterRoles") ||
		!exactBool(release, false, "spec", "values", "containerSecurityContext", "allowPrivilegeEscalation") ||
		!exactBool(release, true, "spec", "values", "containerSecurityContext", "readOnlyRootFilesystem") {
		return errors.New("CloudNativePG controller values are invalid")
	}
	spreads, ok := nested(release, "spec", "values", "topologySpreadConstraints").([]any)
	if !ok || len(spreads) != 1 {
		return errors.New("CloudNativePG controller topology is invalid")
	}
	spread, ok := spreads[0].(map[string]any)
	if !ok || nestedNumber(spread, "maxSkew") != 1 || nestedString(spread, "topologyKey") != "kubernetes.io/hostname" || nestedString(spread, "whenUnsatisfiable") != "DoNotSchedule" {
		return errors.New("CloudNativePG controller topology is invalid")
	}
	budget := index["PodDisruptionBudget/cnpg-system/cnpg-controller-manager"].Data
	if nestedNumber(budget, "spec", "minAvailable") != 1 ||
		nestedString(budget, "spec", "selector", "matchLabels", "app.kubernetes.io/instance") != "cnpg" ||
		nestedString(budget, "spec", "selector", "matchLabels", "app.kubernetes.io/name") != "cloudnative-pg" {
		return errors.New("CloudNativePG controller disruption budget is invalid")
	}
	return nil
}

func validatePostgreSQLHARuntime(index map[string]object) error {
	cluster := index["Cluster/cloudring-database/cloudring-postgres"].Data
	if nestedNumber(cluster, "spec", "instances") != 3 ||
		nestedString(cluster, "spec", "imageName") != "ghcr.io/cloudnative-pg/postgresql:18.4@sha256:17760b4508e1703f7bf3d5ee6c9335ff6d1bf62b034a3d5b32a43414a516789f" ||
		!exactBool(cluster, false, "spec", "enableSuperuserAccess") ||
		!exactBool(cluster, true, "spec", "enablePDB") ||
		nestedString(cluster, "spec", "primaryUpdateMethod") != "switchover" ||
		nestedString(cluster, "spec", "primaryUpdateStrategy") != "unsupervised" ||
		nestedString(cluster, "metadata", "annotations", "cloudring.org/non-claim") == "" {
		return errors.New("PostgreSQL cluster HA boundary is invalid")
	}
	if nestedString(cluster, "spec", "bootstrap", "initdb", "database") != "cloudring" ||
		nestedString(cluster, "spec", "bootstrap", "initdb", "owner") != "cloudring_owner" ||
		nestedString(cluster, "spec", "bootstrap", "initdb", "secret", "name") != "cloudring-postgres-owner" ||
		!exactBool(cluster, true, "spec", "bootstrap", "initdb", "dataChecksums") {
		return errors.New("PostgreSQL bootstrap identity is invalid")
	}
	roles, ok := nested(cluster, "spec", "managed", "roles").([]any)
	if !ok || len(roles) != 1 {
		return errors.New("PostgreSQL managed application role is invalid")
	}
	applicationRole, ok := roles[0].(map[string]any)
	if !ok || !exactMappingKeys(applicationRole,
		"name", "ensure", "comment", "login", "superuser", "createdb", "createrole",
		"inherit", "replication", "bypassrls", "connectionLimit", "passwordSecret") ||
		nestedString(applicationRole, "name") != "cloudring_app" ||
		nestedString(applicationRole, "ensure") != "present" ||
		nestedString(applicationRole, "comment") == "" ||
		!exactBool(applicationRole, true, "login") ||
		!exactBool(applicationRole, false, "superuser") ||
		!exactBool(applicationRole, false, "createdb") ||
		!exactBool(applicationRole, false, "createrole") ||
		!exactBool(applicationRole, true, "inherit") ||
		!exactBool(applicationRole, false, "replication") ||
		!exactBool(applicationRole, false, "bypassrls") ||
		nestedNumber(applicationRole, "connectionLimit") != 50 ||
		nestedString(applicationRole, "passwordSecret", "name") != "cloudring-postgres-app" {
		return errors.New("PostgreSQL managed application role is invalid")
	}
	if nestedString(cluster, "spec", "postgresql", "parameters", "password_encryption") != "scram-sha-256" ||
		nestedString(cluster, "spec", "postgresql", "parameters", "ssl_min_protocol_version") != "TLSv1.3" ||
		nestedString(cluster, "spec", "postgresql", "parameters", "synchronous_commit") != "remote_apply" ||
		nested(cluster, "spec", "postgresql", "pg_hba") != nil {
		return errors.New("PostgreSQL authentication or TLS policy is invalid")
	}
	if nestedString(cluster, "spec", "postgresql", "synchronous", "method") != "any" ||
		nestedNumber(cluster, "spec", "postgresql", "synchronous", "number") != 1 ||
		nestedString(cluster, "spec", "postgresql", "synchronous", "dataDurability") != "required" ||
		!exactBool(cluster, true, "spec", "postgresql", "synchronous", "failoverQuorum") {
		return errors.New("PostgreSQL synchronous durability policy is invalid")
	}
	if !exactBool(cluster, true, "spec", "affinity", "enablePodAntiAffinity") ||
		nestedString(cluster, "spec", "affinity", "podAntiAffinityType") != "required" ||
		nestedString(cluster, "spec", "affinity", "topologyKey") != "kubernetes.io/hostname" ||
		nestedString(cluster, "spec", "affinity", "nodeSelector", "kubernetes.io/os") != "linux" {
		return errors.New("PostgreSQL failure-domain placement is invalid")
	}
	for _, path := range []string{"storage", "walStorage"} {
		if nestedString(cluster, "spec", path, "storageClass") != "longhorn-replicated" ||
			nestedString(cluster, "spec", path, "size") == "" ||
			!exactBool(cluster, true, "spec", path, "resizeInUseVolumes") {
			return errors.New("PostgreSQL replicated storage is invalid")
		}
	}
	if nestedString(cluster, "spec", "backup", "target") != "prefer-standby" ||
		nestedString(cluster, "spec", "backup", "volumeSnapshot", "className") != "longhorn-retain" ||
		nestedString(cluster, "spec", "backup", "volumeSnapshot", "walClassName") != "longhorn-retain" ||
		nestedString(cluster, "spec", "backup", "volumeSnapshot", "snapshotOwnerReference") != "none" ||
		!exactBool(cluster, true, "spec", "backup", "volumeSnapshot", "online") {
		return errors.New("PostgreSQL retained snapshot policy is invalid")
	}
	backup := index["ScheduledBackup/cloudring-database/cloudring-postgres-volume-snapshot"].Data
	if nestedString(backup, "spec", "method") != "volumeSnapshot" ||
		nestedString(backup, "spec", "backupOwnerReference") != "self" ||
		nestedString(backup, "spec", "cluster", "name") != "cloudring-postgres" ||
		nestedString(backup, "spec", "schedule") != "0 0 */6 * * *" ||
		!exactBool(backup, false, "spec", "immediate") {
		return errors.New("PostgreSQL scheduled backup is invalid")
	}
	policy := index["NetworkPolicy/cloudring-database/cloudring-postgres-ingress"].Data
	if nestedString(policy, "spec", "podSelector", "matchLabels", "cnpg.io/cluster") != "cloudring-postgres" ||
		!exactStringSequence(nested(policy, "spec", "policyTypes"), "Ingress") ||
		lenSequence(nested(policy, "spec", "ingress")) != 4 || nested(policy, "spec", "egress") != nil ||
		!postgresqlCrossNamespaceClientPolicy(policy) {
		return errors.New("PostgreSQL network boundary is invalid")
	}
	return nil
}

func postgresqlCrossNamespaceClientPolicy(policy map[string]any) bool {
	ingress, _ := nested(policy, "spec", "ingress").([]any)
	for _, rawRule := range ingress {
		rule, _ := rawRule.(map[string]any)
		from, _ := rule["from"].([]any)
		if len(from) != 1 {
			continue
		}
		peer, _ := from[0].(map[string]any)
		if nestedString(peer, "namespaceSelector", "matchLabels", "cloudring.org/postgresql-client-namespace") == "true" &&
			nestedString(peer, "podSelector", "matchLabels", "cloudring.org/postgresql-client") == "true" {
			return true
		}
	}
	return false
}

func lenSequence(value any) int {
	items, _ := value.([]any)
	return len(items)
}
