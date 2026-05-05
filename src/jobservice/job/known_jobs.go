// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package job

import "github.com/goharbor/harbor/src/lib"

// Define the register name constants of known jobs

const (
	// SampleJob is name of demo job
	SampleJob = "DEMO"

	// ImageScanJobVendorType is name of scan job it will be used as key to register to job service.
	ImageScanJobVendorType = "IMAGE_SCAN"
	// SBOMJobVendorType key to create sbom generate execution.
	SBOMJobVendorType = "SBOM"
	// GarbageCollectionVendorType job name
	GarbageCollectionVendorType = "GARBAGE_COLLECTION"
	// ReplicationVendorType : the name of the replication job in job service
	ReplicationVendorType = "REPLICATION"
	// WebhookJobVendorType : the name of the webhook job in job service
	WebhookJobVendorType = "WEBHOOK"
	// SlackJobVendorType : the name of the slack job in job service
	SlackJobVendorType = "SLACK"
	// RetentionVendorType : the name of the retention job
	RetentionVendorType = "RETENTION"
	// P2PPreheatVendorType : the name of the P2P preheat job
	P2PPreheatVendorType = "P2P_PREHEAT"
	// PurgeAuditVendorType : the name of purge audit job
	PurgeAuditVendorType = "PURGE_AUDIT_LOG"
	// SystemArtifactCleanupVendorType : the name of the SystemArtifact cleanup job
	SystemArtifactCleanupVendorType = "SYSTEM_ARTIFACT_CLEANUP"
	// ScanDataExportVendorType : the name of the scan data export job
	ScanDataExportVendorType = "SCAN_DATA_EXPORT"
	// ExecSweepVendorType: the name of the execution sweep job
	ExecSweepVendorType = "EXECUTION_SWEEP"
	// ScanAllVendorType: the name of the scan all job
	ScanAllVendorType = "SCAN_ALL"
	// AuditLogsGDPRCompliantVendorType : the name of the job which makes audit logs table GDPR-compliant
	AuditLogsGDPRCompliantVendorType = "AUDIT_LOGS_GDPR_COMPLIANT"
)

var (
	// executionSweeperCount stores the count for execution retained
	executionSweeperCount = map[string]int64{
		ImageScanJobVendorType:          lib.GetEnvInt64("IMAGE_SCAN_EXECUTION_RETENTION_COUNT", 1),
		SBOMJobVendorType:               lib.GetEnvInt64("SBOM_EXECUTION_RETENTION_COUNT", 1),
		ScanAllVendorType:               lib.GetEnvInt64("SCAN_ALL_EXECUTION_RETENTION_COUNT", 1),
		PurgeAuditVendorType:            lib.GetEnvInt64("PURGE_AUDIT_EXECUTION_RETENTION_COUNT", 10),
		ExecSweepVendorType:             lib.GetEnvInt64("EXECUTION_SWEEP_EXECUTION_RETENTION_COUNT", 10),
		GarbageCollectionVendorType:     lib.GetEnvInt64("GARBAGE_COLLECTION_EXECUTION_RETENTION_COUNT", 50),
		SlackJobVendorType:              lib.GetEnvInt64("SLACK_EXECUTION_RETENTION_COUNT", 50),
		WebhookJobVendorType:            lib.GetEnvInt64("WEBHOOK_EXECUTION_RETENTION_COUNT", 50),
		ReplicationVendorType:           lib.GetEnvInt64("REPLICATION_EXECUTION_RETENTION_COUNT", 50),
		ScanDataExportVendorType:        lib.GetEnvInt64("SCAN_DATA_EXPORT_EXECUTION_RETENTION_COUNT", 50),
		SystemArtifactCleanupVendorType: lib.GetEnvInt64("SYSTEM_ARTIFACT_CLEANUP_EXECUTION_RETENTION_COUNT", 50),
		P2PPreheatVendorType:            lib.GetEnvInt64("P2P_PREHEAT_EXECUTION_RETENTION_COUNT", 50),
		RetentionVendorType:             lib.GetEnvInt64("RETENTION_EXECUTION_RETENTION_COUNT", 50),
	}
)

// GetExecutionSweeperCount gets the count of execution records retained by the sweeper
func GetExecutionSweeperCount() map[string]int64 {
	return executionSweeperCount
}
