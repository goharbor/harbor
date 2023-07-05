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

const (
	// ImageScanJob is name of scan job it will be used as key to register to job service.
	ImageScanJob = "IMAGE_SCAN"
	// ImageScanAllJob is the name of "scanall" job in job service
	ImageScanAllJob = "IMAGE_SCAN_ALL"
	// ImageGC the name of image garbage collection job in job service
	ImageGC = "IMAGE_GC"
	// ImageGCReadOnly the name of image garbage collection read only job in job service
	ImageGCReadOnly = "IMAGE_GC_READ_ONLY"
	// JobKindGeneric : Kind of generic job
	JobKindGeneric = "Generic"
	// JobKindScheduled : Kind of scheduled job
	JobKindScheduled = "Scheduled"
	// JobKindPeriodic : Kind of periodic job
	JobKindPeriodic = "Periodic"

	// JobServiceStatusPending   : job status pending
	JobServiceStatusPending = "Pending"
	// JobServiceStatusRunning   : job status running
	JobServiceStatusRunning = "Running"
	// JobServiceStatusStopped   : job status stopped
	JobServiceStatusStopped = "Stopped"
	// JobServiceStatusCancelled : job status cancelled
	JobServiceStatusCancelled = "Cancelled"
	// JobServiceStatusError     : job status error
	JobServiceStatusError = "Error"
	// JobServiceStatusSuccess   : job status success
	JobServiceStatusSuccess = "Success"
	// JobServiceStatusScheduled : job status scheduled
	JobServiceStatusScheduled = "Scheduled"

	// JobActionStop : the action to stop the job
	JobActionStop = "stop"
)
