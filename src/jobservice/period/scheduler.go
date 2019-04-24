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

package period

// Scheduler defines operations the periodic scheduler should have.
type Scheduler interface {
	// Start to serve periodic job scheduling process
	//
	// Returns:
	//  error if any problems happened
	Start() error

	// Stop the working periodic job scheduling process
	//
	// Returns;
	//  error if any problems happened
	Stop() error

	// Schedule the specified cron job policy.
	//
	// policy *Policy           : The job template of the scheduling periodical jobs
	//
	// Returns:
	//  int64 the numeric id of policy
	//  error if failed to schedule
	Schedule(policy *Policy) (int64, error)

	// Unschedule the specified cron job policy.
	//
	// policyID string: The ID of cron job policy.
	//
	// Return:
	//  error if failed to unschedule
	UnSchedule(policyID string) error
}
