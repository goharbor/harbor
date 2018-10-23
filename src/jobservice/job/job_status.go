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
	// JobStatusPending   : job status pending
	JobStatusPending = "Pending"
	// JobStatusRunning   : job status running
	JobStatusRunning = "Running"
	// JobStatusStopped   : job status stopped
	JobStatusStopped = "Stopped"
	// JobStatusCancelled : job status cancelled
	JobStatusCancelled = "Cancelled"
	// JobStatusError     : job status error
	JobStatusError = "Error"
	// JobStatusSuccess   : job status success
	JobStatusSuccess = "Success"
	// JobStatusScheduled : job status scheduled
	JobStatusScheduled = "Scheduled"
)
