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

const longhornThreeNodeProfilePath = "deploy/kubernetes/storage/longhorn-three-node"

// VerifyLonghornThreeNode validates the reusable compact-cell profile without
// contacting a cluster or chart registry. Host and live durability checks stay
// downstream release gates.
func VerifyLonghornThreeNode(root string) (Report, error) {
	root, err := canonicalRoot(root)
	if err != nil {
		return Report{}, err
	}
	report := Report{Status: "blocked", Profile: "cloudring-longhorn-three-node/v1"}
	if _, err := VerifyCSISnapshotAPI(root); err != nil {
		return report, errors.New("canonical CSI snapshot API dependency is invalid")
	}
	repository, err := os.OpenRoot(root)
	if err != nil {
		return report, errors.New("open confined repository root")
	}
	defer repository.Close()

	objects, files, err := readLonghornThreeNodeStage(repository)
	if err != nil {
		return report, err
	}
	report.Files = files
	report.Documents = len(objects)
	if report.Files != 2 || report.Documents != 6 {
		return report, errors.New("Longhorn three-node source inventory is incomplete")
	}
	if err := validateLonghornThreeNodeObjects(objects); err != nil {
		return report, err
	}
	report.Status = "ready"
	report.Checks = []string{
		"source_release_suspended",
		"longhorn_version_pinned",
		"three_replica_anti_affinity_ready",
		"v1_engine_and_host_path_explicit",
		"degraded_creation_and_telemetry_disabled",
		"storage_class_non_default_and_delayed",
		"migratable_vm_storage_class_explicit",
		"single_retained_velero_snapshot_class",
		"canonical_csi_snapshot_api_stage_required",
		"ui_ingress_disabled",
	}
	return report, nil
}

func readLonghornThreeNodeStage(root *os.Root) ([]object, int, error) {
	directory := filepath.Join(longhornThreeNodeProfilePath, "runtime")
	kustomization, err := readRegular(root, filepath.Join(directory, "kustomization.yaml"))
	if err != nil {
		return nil, 0, err
	}
	var manifest map[string]any
	if err := decodeOne(kustomization, &manifest); err != nil ||
		!exactMappingKeys(manifest, "apiVersion", "kind", "resources") ||
		stringValue(manifest["apiVersion"]) != "kustomize.config.k8s.io/v1beta1" ||
		stringValue(manifest["kind"]) != "Kustomization" ||
		!exactStringSequence(manifest["resources"], "resources.yaml") {
		return nil, 0, errors.New("Longhorn runtime stage has an invalid kustomization")
	}
	data, err := readRegular(root, filepath.Join(directory, "resources.yaml"))
	if err != nil {
		return nil, 0, err
	}
	for _, forbidden := range [][]byte{[]byte("REPLACE_WITH"), []byte(":latest"), []byte("example.invalid")} {
		if bytes.Contains(data, forbidden) {
			return nil, 0, errors.New("Longhorn runtime contains an unresolved or mutable reference")
		}
	}
	objects, err := decodeObjects(data)
	if err != nil {
		return nil, 0, fmt.Errorf("decode Longhorn runtime: %w", err)
	}
	return objects, 2, nil
}

func validateLonghornThreeNodeObjects(objects []object) error {
	index := map[string]object{}
	veleroSnapshotLabels := 0
	for _, item := range objects {
		key := item.Kind + "/" + item.Namespace + "/" + item.Name
		if _, duplicate := index[key]; duplicate {
			return errors.New("duplicate Longhorn manifest object identity")
		}
		index[key] = item
		if nestedString(item.Data, "metadata", "labels", "velero.io/csi-volumesnapshot-class") == "true" {
			veleroSnapshotLabels++
		}
	}
	expected := []string{
		"HelmRelease/longhorn-system/longhorn",
		"HelmRepository/longhorn-system/longhorn",
		"Namespace//longhorn-system",
		"StorageClass//longhorn-migratable",
		"StorageClass//longhorn-replicated",
		"VolumeSnapshotClass//longhorn-retain",
	}
	actual := make([]string, 0, len(index))
	for key := range index {
		actual = append(actual, key)
	}
	sort.Strings(actual)
	if strings.Join(actual, "\n") != strings.Join(expected, "\n") {
		return errors.New("Longhorn three-node object inventory is not exact")
	}
	require := func(kind, namespace, name string) object {
		return index[kind+"/"+namespace+"/"+name]
	}

	namespace := require("Namespace", "", "longhorn-system")
	if !exactStringMap(nested(namespace.Data, "metadata", "labels"), map[string]string{
		"app.kubernetes.io/part-of":          "cloudring-storage",
		"pod-security.kubernetes.io/enforce": "privileged",
		"pod-security.kubernetes.io/audit":   "baseline",
		"pod-security.kubernetes.io/warn":    "baseline",
	}) {
		return errors.New("Longhorn namespace boundary is invalid")
	}
	repository := require("HelmRepository", "longhorn-system", "longhorn")
	if !exactMappingKeys(nested(repository.Data, "spec"), "interval", "url") ||
		nestedString(repository.Data, "spec", "interval") != "1h" ||
		nestedString(repository.Data, "spec", "url") != "https://charts.longhorn.io" {
		return errors.New("Longhorn chart repository is invalid")
	}

	release := require("HelmRelease", "longhorn-system", "longhorn")
	if !exactMappingKeys(nested(release.Data, "spec"), "suspend", "interval", "timeout", "releaseName", "chart", "install", "upgrade", "values") ||
		!exactBool(release.Data, true, "spec", "suspend") ||
		nestedString(release.Data, "spec", "interval") != "15m" || nestedString(release.Data, "spec", "timeout") != "15m" ||
		nestedString(release.Data, "spec", "releaseName") != "longhorn" || !exactLonghornHelmChart(release.Data) ||
		nested(release.Data, "spec", "valuesFrom") != nil || nested(release.Data, "spec", "postRenderers") != nil ||
		nestedString(release.Data, "metadata", "annotations", "cloudring.org/non-claim") != "downstream-host-prerequisites-and-live-storage-evidence-required" ||
		nestedString(release.Data, "metadata", "annotations", "cloudring.org/requires-stage") != "deploy/kubernetes/storage/csi-snapshot-api/controller" {
		return errors.New("Longhorn Helm release activation or chart boundary is invalid")
	}
	if !validateLonghornRemediation(release.Data) || !validateLonghornValues(release.Data) {
		return errors.New("Longhorn Helm release values or remediation boundary is invalid")
	}

	storageClass := require("StorageClass", "", "longhorn-replicated")
	if !validateLonghornStorageClass(storageClass.Data, false) {
		return errors.New("Longhorn StorageClass is invalid")
	}
	migratableStorageClass := require("StorageClass", "", "longhorn-migratable")
	if !validateLonghornStorageClass(migratableStorageClass.Data, true) {
		return errors.New("Longhorn migratable StorageClass is invalid")
	}
	snapshotClass := require("VolumeSnapshotClass", "", "longhorn-retain")
	if veleroSnapshotLabels != 1 || !validateLonghornSnapshotClass(snapshotClass.Data) {
		return errors.New("Longhorn Velero snapshot class is invalid")
	}
	return nil
}

func exactLonghornHelmChart(release map[string]any) bool {
	chart, ok := nested(release, "spec", "chart", "spec").(map[string]any)
	return ok && exactMappingKeys(chart, "chart", "version", "sourceRef", "interval") &&
		nestedString(release, "spec", "chart", "spec", "chart") == "longhorn" &&
		nestedString(release, "spec", "chart", "spec", "version") == "1.12.0" &&
		nestedString(release, "spec", "chart", "spec", "sourceRef", "kind") == "HelmRepository" &&
		nestedString(release, "spec", "chart", "spec", "sourceRef", "name") == "longhorn" &&
		nestedString(release, "spec", "chart", "spec", "sourceRef", "namespace") == "longhorn-system" &&
		nestedString(release, "spec", "chart", "spec", "interval") == "1h"
}

func validateLonghornRemediation(release map[string]any) bool {
	install, installOK := nested(release, "spec", "install").(map[string]any)
	upgrade, upgradeOK := nested(release, "spec", "upgrade").(map[string]any)
	return installOK && upgradeOK && exactMappingKeys(install, "crds", "remediation") && exactMappingKeys(upgrade, "crds", "remediation") &&
		nestedString(release, "spec", "install", "crds") == "CreateReplace" && nestedNumber(release, "spec", "install", "remediation", "retries") == 3 &&
		nestedString(release, "spec", "upgrade", "crds") == "CreateReplace" && nestedNumber(release, "spec", "upgrade", "remediation", "retries") == 3 &&
		exactBool(release, true, "spec", "upgrade", "remediation", "remediateLastFailure")
}

func validateLonghornValues(release map[string]any) bool {
	values, ok := nested(release, "spec", "values").(map[string]any)
	if !ok || !exactMappingKeys(values, "global", "persistence", "preUpgradeChecker", "defaultSettings", "longhornUI", "ingress", "metrics") {
		return false
	}
	tolerations, ok := nested(values, "global", "tolerations").([]any)
	if !ok || len(tolerations) != 1 {
		return false
	}
	toleration, ok := tolerations[0].(map[string]any)
	if !ok || !exactMappingKeys(toleration, "key", "operator", "effect") || stringValue(toleration["key"]) != "node-role.kubernetes.io/control-plane" || stringValue(toleration["operator"]) != "Exists" || stringValue(toleration["effect"]) != "NoSchedule" {
		return false
	}
	settings, ok := nested(values, "defaultSettings").(map[string]any)
	return ok && exactMappingKeys(settings,
		"createDefaultDiskLabeledNodes", "defaultDataPath", "defaultReplicaCount", "replicaSoftAntiAffinity", "replicaZoneSoftAntiAffinity", "replicaDiskSoftAntiAffinity",
		"replicaAutoBalance", "storageOverProvisioningPercentage", "storageMinimalAvailablePercentage", "storageReservedPercentageForDefaultDisk", "taintToleration",
		"upgradeChecker", "disableRevisionCounter", "allowVolumeCreationWithDegradedAvailability", "allowCollectingLonghornUsageMetrics", "v1DataEngine", "v2DataEngine") &&
		exactBool(values, false, "persistence", "createStorageClass") && exactBool(values, false, "persistence", "defaultClass") &&
		exactBool(values, false, "preUpgradeChecker", "jobEnabled") && exactBool(values, false, "preUpgradeChecker", "upgradeVersionCheck") &&
		exactBool(values, false, "defaultSettings", "createDefaultDiskLabeledNodes") && nestedString(values, "defaultSettings", "defaultDataPath") == "/var/lib/longhorn" &&
		nestedString(values, "defaultSettings", "defaultReplicaCount") == `{"v1":"3","v2":"3"}` &&
		exactBool(values, false, "defaultSettings", "replicaSoftAntiAffinity") && exactBool(values, true, "defaultSettings", "replicaZoneSoftAntiAffinity") && exactBool(values, false, "defaultSettings", "replicaDiskSoftAntiAffinity") &&
		nestedString(values, "defaultSettings", "replicaAutoBalance") == "best-effort" && nestedNumber(values, "defaultSettings", "storageOverProvisioningPercentage") == 100 &&
		nestedNumber(values, "defaultSettings", "storageMinimalAvailablePercentage") == 25 && nestedNumber(values, "defaultSettings", "storageReservedPercentageForDefaultDisk") == 25 &&
		nestedString(values, "defaultSettings", "taintToleration") == "node-role.kubernetes.io/control-plane:NoSchedule" &&
		exactBool(values, false, "defaultSettings", "upgradeChecker") && nestedString(values, "defaultSettings", "disableRevisionCounter") == `{"v1":"false"}` &&
		exactBool(values, false, "defaultSettings", "allowVolumeCreationWithDegradedAvailability") &&
		exactBool(values, false, "defaultSettings", "allowCollectingLonghornUsageMetrics") && exactBool(values, true, "defaultSettings", "v1DataEngine") && exactBool(values, false, "defaultSettings", "v2DataEngine") &&
		nestedNumber(values, "longhornUI", "replicas") == 2 && exactBool(values, false, "ingress", "enabled") && exactBool(values, false, "metrics", "serviceMonitor", "enabled")
}

func validateLonghornStorageClass(class map[string]any, migratable bool) bool {
	parameters, ok := nested(class, "parameters").(map[string]any)
	if !ok {
		return false
	}
	parameterKeys := []string{"numberOfReplicas", "staleReplicaTimeout", "fsType", "dataLocality", "unmapMarkSnapChainRemoved", "disableRevisionCounter", "dataEngine"}
	if migratable {
		parameterKeys = append(parameterKeys, "migratable")
	}
	return exactMappingKeys(class, "apiVersion", "kind", "metadata", "provisioner", "parameters", "allowVolumeExpansion", "reclaimPolicy", "volumeBindingMode") &&
		nestedString(class, "apiVersion") == "storage.k8s.io/v1" && nestedString(class, "provisioner") == "driver.longhorn.io" &&
		nestedString(class, "metadata", "annotations", "storageclass.kubernetes.io/is-default-class") == "false" &&
		exactMappingKeys(parameters, parameterKeys...) &&
		nestedString(class, "parameters", "numberOfReplicas") == "3" && nestedString(class, "parameters", "staleReplicaTimeout") == "30" &&
		nestedString(class, "parameters", "fsType") == "ext4" && nestedString(class, "parameters", "dataLocality") == "disabled" &&
		nestedString(class, "parameters", "unmapMarkSnapChainRemoved") == "ignored" && nestedString(class, "parameters", "disableRevisionCounter") == "false" &&
		nestedString(class, "parameters", "dataEngine") == "v1" && (!migratable || nestedString(class, "parameters", "migratable") == "true") && exactBool(class, true, "allowVolumeExpansion") &&
		nestedString(class, "reclaimPolicy") == "Delete" && nestedString(class, "volumeBindingMode") == "WaitForFirstConsumer"
}

func validateLonghornSnapshotClass(class map[string]any) bool {
	parameters, ok := nested(class, "parameters").(map[string]any)
	return ok && exactMappingKeys(class, "apiVersion", "kind", "metadata", "driver", "deletionPolicy", "parameters") &&
		nestedString(class, "apiVersion") == "snapshot.storage.k8s.io/v1" && nestedString(class, "driver") == "driver.longhorn.io" &&
		nestedString(class, "deletionPolicy") == "Retain" && nestedString(class, "metadata", "labels", "velero.io/csi-volumesnapshot-class") == "true" &&
		nestedString(class, "metadata", "annotations", "snapshot.storage.kubernetes.io/is-default-class") == "false" &&
		exactMappingKeys(parameters, "type") && nestedString(class, "parameters", "type") == "snap"
}
