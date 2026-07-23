// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package platformmanifest

import (
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"time"

	"github.com/opencloudtech/CloudRING/internal/strictjson"
)

const postgresqlRecoveryEvidenceSchemaVersion = "cloudring.postgresql-cnpg-offcell-recovery-evidence/v1"

var (
	postgresqlRecoveryDigestPattern   = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	postgresqlRecoveryRevisionPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	postgresqlRecoveryEvidenceInvalid = errors.New("PostgreSQL recovery evidence is invalid")
)

type postgresqlRecoveryEvidence struct {
	SchemaVersion  string                       `json:"schemaVersion"`
	SourceRevision string                       `json:"sourceRevision"`
	CollectedAt    string                       `json:"collectedAt"`
	ExpiresAt      string                       `json:"expiresAt"`
	OffCell        postgresqlRecoveryOffCell    `json:"offCell"`
	BaseBackup     postgresqlRecoveryBaseBackup `json:"baseBackup"`
	WALArchive     postgresqlRecoveryWALArchive `json:"walArchive"`
	Recovery       postgresqlRecoveryCluster    `json:"recovery"`
	Checksum       postgresqlRecoveryChecksum   `json:"checksum"`
	Cleanup        postgresqlRecoveryCleanup    `json:"cleanup"`
	Redaction      postgresqlRecoveryRedaction  `json:"redaction"`
	Verdict        string                       `json:"verdict"`
}

type postgresqlRecoveryOffCell struct {
	ObservedAt            string `json:"observedAt"`
	DestinationIdentity   string `json:"destinationIdentity"`
	FailureDomainDistinct bool   `json:"failureDomainDistinct"`
	RetentionDays         int    `json:"retentionDays"`
	ObjectLockMode        string `json:"objectLockMode"`
	ObjectLockMinimumDays int    `json:"objectLockMinimumDays"`
	ControlDeleteDenied   bool   `json:"controlDeleteDenied"`
}

type postgresqlRecoveryBaseBackup struct {
	Identity              string `json:"identity"`
	StartedAt             string `json:"startedAt"`
	CompletedAt           string `json:"completedAt"`
	Status                string `json:"status"`
	Bytes                 int64  `json:"bytes"`
	ObjectInventoryDigest string `json:"objectInventoryDigest"`
}

type postgresqlRecoveryWALArchive struct {
	FirstRecoverabilityPoint string  `json:"firstRecoverabilityPoint"`
	LastArchivedAt           string  `json:"lastArchivedAt"`
	LastFailedAt             *string `json:"lastFailedAt"`
	ReplayedThrough          string  `json:"replayedThrough"`
	Continuous               bool    `json:"continuous"`
}

type postgresqlRecoveryCluster struct {
	NamespaceIdentity    string `json:"namespaceIdentity"`
	ClusterIdentity      string `json:"clusterIdentity"`
	SourceIdentity       string `json:"sourceIdentity"`
	StartedAt            string `json:"startedAt"`
	ReadyAt              string `json:"readyAt"`
	ValidatedAt          string `json:"validatedAt"`
	ReadyInstances       int    `json:"readyInstances"`
	ExpectedInstances    int    `json:"expectedInstances"`
	ProductionRouteCount int    `json:"productionRouteCount"`
	WriteProbePassed     bool   `json:"writeProbePassed"`
}

type postgresqlRecoveryChecksum struct {
	Algorithm             string `json:"algorithm"`
	ProjectionVersion     string `json:"projectionVersion"`
	Source                string `json:"source"`
	Recovered             string `json:"recovered"`
	SourceCapturedAt      string `json:"sourceCapturedAt"`
	RecoveredCapturedAt   string `json:"recoveredCapturedAt"`
	SourceLogicalBytes    int64  `json:"sourceLogicalBytes"`
	RecoveredLogicalBytes int64  `json:"recoveredLogicalBytes"`
	SourceRowCount        int64  `json:"sourceRowCount"`
	RecoveredRowCount     int64  `json:"recoveredRowCount"`
	Matched               bool   `json:"matched"`
}

type postgresqlRecoveryCleanup struct {
	StartedAt                  string                           `json:"startedAt"`
	CompletedAt                string                           `json:"completedAt"`
	Complete                   bool                             `json:"complete"`
	TwoSweepQuietWindowSeconds int                              `json:"twoSweepQuietWindowSeconds"`
	Sweeps                     []postgresqlRecoveryCleanupSweep `json:"sweeps"`
}

type postgresqlRecoveryCleanupSweep struct {
	ObservedAt                 string `json:"observedAt"`
	InventoryDigest            string `json:"inventoryDigest"`
	RecoveryNamespaceCount     int    `json:"recoveryNamespaceCount"`
	ClusterCount               int    `json:"clusterCount"`
	CredentialSecretCount      int    `json:"credentialSecretCount"`
	PersistentVolumeClaimCount int    `json:"persistentVolumeClaimCount"`
	ServiceCount               int    `json:"serviceCount"`
	RouteCount                 int    `json:"routeCount"`
}

type postgresqlRecoveryRedaction struct {
	ContainsCredentials bool   `json:"containsCredentials"`
	ContainsEndpoints   bool   `json:"containsEndpoints"`
	ContainsTenantData  bool   `json:"containsTenantData"`
	Verdict             string `json:"verdict"`
}

// VerifyPostgreSQLRecoveryEvidence validates one sanitized evidence instance.
// Structural source verification does not call this function and remains a
// live-readiness non-claim until an independently collected instance passes.
func VerifyPostgreSQLRecoveryEvidence(reader io.Reader) error {
	payload, err := strictjson.Read(reader)
	if err != nil {
		return postgresqlRecoveryEvidenceInvalid
	}
	var evidence postgresqlRecoveryEvidence
	if !exactPostgreSQLRecoveryEvidenceShape(payload) ||
		strictjson.DecodeExact(payload, &evidence) != nil ||
		validatePostgreSQLRecoveryEvidence(evidence) != nil {
		return postgresqlRecoveryEvidenceInvalid
	}
	return nil
}

func exactPostgreSQLRecoveryEvidenceShape(payload []byte) bool {
	root, ok := decodePostgreSQLRecoveryEvidenceObject(payload,
		"schemaVersion", "sourceRevision", "collectedAt", "expiresAt", "offCell", "baseBackup",
		"walArchive", "recovery", "checksum", "cleanup", "redaction", "verdict")
	if !ok {
		return false
	}
	if _, ok = decodePostgreSQLRecoveryEvidenceObject(root["offCell"],
		"observedAt", "destinationIdentity", "failureDomainDistinct", "retentionDays", "objectLockMode",
		"objectLockMinimumDays", "controlDeleteDenied"); !ok {
		return false
	}
	if _, ok = decodePostgreSQLRecoveryEvidenceObject(root["baseBackup"],
		"identity", "startedAt", "completedAt", "status", "bytes", "objectInventoryDigest"); !ok {
		return false
	}
	if _, ok = decodePostgreSQLRecoveryEvidenceObject(root["walArchive"],
		"firstRecoverabilityPoint", "lastArchivedAt", "lastFailedAt", "replayedThrough", "continuous"); !ok {
		return false
	}
	if _, ok = decodePostgreSQLRecoveryEvidenceObject(root["recovery"],
		"namespaceIdentity", "clusterIdentity", "sourceIdentity", "startedAt", "readyAt", "validatedAt",
		"readyInstances", "expectedInstances", "productionRouteCount", "writeProbePassed"); !ok {
		return false
	}
	if _, ok = decodePostgreSQLRecoveryEvidenceObject(root["checksum"],
		"algorithm", "projectionVersion", "source", "recovered", "sourceCapturedAt", "recoveredCapturedAt",
		"sourceLogicalBytes", "recoveredLogicalBytes", "sourceRowCount", "recoveredRowCount", "matched"); !ok {
		return false
	}
	cleanup, ok := decodePostgreSQLRecoveryEvidenceObject(root["cleanup"],
		"startedAt", "completedAt", "complete", "twoSweepQuietWindowSeconds", "sweeps")
	if !ok {
		return false
	}
	var sweeps []json.RawMessage
	if json.Unmarshal(cleanup["sweeps"], &sweeps) != nil || len(sweeps) != 2 {
		return false
	}
	for _, sweep := range sweeps {
		if _, sweepOK := decodePostgreSQLRecoveryEvidenceObject(sweep,
			"observedAt", "inventoryDigest", "recoveryNamespaceCount", "clusterCount", "credentialSecretCount",
			"persistentVolumeClaimCount", "serviceCount", "routeCount"); !sweepOK {
			return false
		}
	}
	_, ok = decodePostgreSQLRecoveryEvidenceObject(root["redaction"],
		"containsCredentials", "containsEndpoints", "containsTenantData", "verdict")
	return ok
}

func decodePostgreSQLRecoveryEvidenceObject(payload []byte, required ...string) (map[string]json.RawMessage, bool) {
	var object map[string]json.RawMessage
	if json.Unmarshal(payload, &object) != nil || len(object) != len(required) {
		return nil, false
	}
	for _, key := range required {
		if _, present := object[key]; !present {
			return nil, false
		}
	}
	return object, true
}

func validatePostgreSQLRecoveryEvidence(evidence postgresqlRecoveryEvidence) error {
	if evidence.SchemaVersion != postgresqlRecoveryEvidenceSchemaVersion ||
		!postgresqlRecoveryRevisionPattern.MatchString(evidence.SourceRevision) ||
		evidence.Verdict != "pass" ||
		!validPostgreSQLRecoveryDigest(evidence.OffCell.DestinationIdentity) ||
		!evidence.OffCell.FailureDomainDistinct || evidence.OffCell.RetentionDays < 30 ||
		(evidence.OffCell.ObjectLockMode != "governance" && evidence.OffCell.ObjectLockMode != "compliance") ||
		evidence.OffCell.ObjectLockMinimumDays < 30 || !evidence.OffCell.ControlDeleteDenied ||
		!validPostgreSQLRecoveryDigest(evidence.BaseBackup.Identity) ||
		!validPostgreSQLRecoveryDigest(evidence.BaseBackup.ObjectInventoryDigest) ||
		evidence.BaseBackup.Status != "completed" || evidence.BaseBackup.Bytes <= 0 ||
		evidence.WALArchive.LastFailedAt != nil || !evidence.WALArchive.Continuous ||
		!validPostgreSQLRecoveryDigest(evidence.Recovery.NamespaceIdentity) ||
		!validPostgreSQLRecoveryDigest(evidence.Recovery.ClusterIdentity) ||
		!validPostgreSQLRecoveryDigest(evidence.Recovery.SourceIdentity) ||
		evidence.Recovery.ReadyInstances != 1 || evidence.Recovery.ExpectedInstances != 1 ||
		evidence.Recovery.ProductionRouteCount != 0 || !evidence.Recovery.WriteProbePassed ||
		evidence.Checksum.Algorithm != "sha256" ||
		evidence.Checksum.ProjectionVersion != "cloudring-postgresql-logical-state/v1" ||
		!validPostgreSQLRecoveryDigest(evidence.Checksum.Source) ||
		evidence.Checksum.Source != evidence.Checksum.Recovered || !evidence.Checksum.Matched ||
		evidence.Checksum.SourceLogicalBytes <= 0 ||
		evidence.Checksum.SourceLogicalBytes != evidence.Checksum.RecoveredLogicalBytes ||
		evidence.Checksum.SourceRowCount <= 0 ||
		evidence.Checksum.SourceRowCount != evidence.Checksum.RecoveredRowCount ||
		!evidence.Cleanup.Complete || evidence.Cleanup.TwoSweepQuietWindowSeconds < 30 ||
		len(evidence.Cleanup.Sweeps) != 2 ||
		evidence.Redaction.ContainsCredentials || evidence.Redaction.ContainsEndpoints ||
		evidence.Redaction.ContainsTenantData || evidence.Redaction.Verdict != "pass" {
		return postgresqlRecoveryEvidenceInvalid
	}
	for _, sweep := range evidence.Cleanup.Sweeps {
		if !validPostgreSQLRecoveryDigest(sweep.InventoryDigest) ||
			sweep.RecoveryNamespaceCount != 0 || sweep.ClusterCount != 0 ||
			sweep.CredentialSecretCount != 0 || sweep.PersistentVolumeClaimCount != 0 ||
			sweep.ServiceCount != 0 || sweep.RouteCount != 0 {
			return postgresqlRecoveryEvidenceInvalid
		}
	}
	return validatePostgreSQLRecoveryChronology(evidence)
}

func validatePostgreSQLRecoveryChronology(evidence postgresqlRecoveryEvidence) error {
	values := []string{
		evidence.OffCell.ObservedAt,
		evidence.BaseBackup.StartedAt,
		evidence.BaseBackup.CompletedAt,
		evidence.WALArchive.FirstRecoverabilityPoint,
		evidence.WALArchive.LastArchivedAt,
		evidence.WALArchive.ReplayedThrough,
		evidence.Recovery.StartedAt,
		evidence.Recovery.ReadyAt,
		evidence.Recovery.ValidatedAt,
		evidence.Checksum.SourceCapturedAt,
		evidence.Checksum.RecoveredCapturedAt,
		evidence.Cleanup.StartedAt,
		evidence.Cleanup.Sweeps[0].ObservedAt,
		evidence.Cleanup.Sweeps[1].ObservedAt,
		evidence.Cleanup.CompletedAt,
		evidence.CollectedAt,
		evidence.ExpiresAt,
	}
	parsed := make([]time.Time, len(values))
	for index, value := range values {
		instant, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			return postgresqlRecoveryEvidenceInvalid
		}
		parsed[index] = instant
	}
	offCellObserved, backupStarted, backupCompleted := parsed[0], parsed[1], parsed[2]
	firstRecoverable, lastArchived, replayedThrough := parsed[3], parsed[4], parsed[5]
	recoveryStarted, recoveryReady, recoveryValidated := parsed[6], parsed[7], parsed[8]
	sourceCaptured, recoveredCaptured := parsed[9], parsed[10]
	cleanupStarted, firstSweep, secondSweep, cleanupCompleted := parsed[11], parsed[12], parsed[13], parsed[14]
	collected, expires := parsed[15], parsed[16]
	quietWindow := time.Duration(evidence.Cleanup.TwoSweepQuietWindowSeconds) * time.Second
	if offCellObserved.After(backupStarted) || !backupStarted.Before(backupCompleted) ||
		firstRecoverable.After(backupCompleted) || sourceCaptured.Before(backupCompleted) ||
		lastArchived.Before(sourceCaptured) || replayedThrough.Before(sourceCaptured) ||
		!lastArchived.Before(recoveryStarted) || !recoveryStarted.Before(recoveryReady) ||
		recoveredCaptured.Before(recoveryReady) || recoveredCaptured.After(recoveryValidated) ||
		replayedThrough.After(recoveredCaptured) ||
		recoveryValidated.Before(recoveryReady) || !recoveryValidated.Before(cleanupStarted) ||
		firstSweep.Before(cleanupStarted) || secondSweep.Sub(firstSweep) < quietWindow ||
		!secondSweep.Before(cleanupCompleted) || !cleanupCompleted.Before(collected) ||
		!collected.Before(expires) {
		return postgresqlRecoveryEvidenceInvalid
	}
	return nil
}

func validPostgreSQLRecoveryDigest(value string) bool {
	return postgresqlRecoveryDigestPattern.MatchString(value)
}
