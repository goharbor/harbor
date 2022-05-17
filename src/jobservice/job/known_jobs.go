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

// Define the register name constants of known jobs

const (
	// SampleJob is name of demo job
	SampleJob = "DEMO"

	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// GarbageCollection job name
	GarbageCollection = "GARBAGE_COLLECTION"
	// Replication : the name of the replication job in job service
	Replication = "REPLICATION"
	// WebhookJob : the name of the webhook job in job service
	WebhookJob = "WEBHOOK"
	// SlackJob : the name of the slack job in job service
	SlackJob = "SLACK"
	// Retention : the name of the retention job
	Retention = "RETENTION"
	// P2PPreheat : the name of the P2P preheat job
	P2PPreheat = "P2P_PREHEAT"
	// PurgeAudit : the name of purge audit job
	PurgeAudit = "PURGE_AUDIT"
)
