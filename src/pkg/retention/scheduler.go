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

package retention

// Scheduler of launching retention jobs
type Scheduler interface {
	// Schedule the job to periodically run the retentions
	//
	//  Arguments:
	//    policyID string : uuid of the retention policy
	//    cron string     : cron pattern like `0-59/5 12 * * * *`
	//  Returns:
	//    the returned job ID
	//    common error object if any errors occurred
	Schedule(policyID string, cron string) (string, error)

	// Unschedule the specified retention policy
	//
	//  Arguments:
	//    policyID string : uuid of the retention policy
	//
	//  Returns:
	//    common error object if any errors occurred
	UnSchedule(policyID string) error
}
