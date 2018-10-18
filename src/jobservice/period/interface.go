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

import "github.com/goharbor/harbor/src/jobservice/models"

// Interface defines operations the periodic scheduler should have.
type Interface interface {
	// Schedule the specified cron job policy.
	//
	// jobName string           : The name of periodical job
	// params models.Parameters : The parameters required by the periodical job
	// cronSpec string          : The periodical settings with cron format
	//
	// Returns:
	//  The uuid of the cron job policy
	//  The latest next trigger time
	//  error if failed to schedule
	Schedule(jobName string, params models.Parameters, cronSpec string) (string, int64, error)

	// Unschedule the specified cron job policy.
	//
	// cronJobPolicyID string: The ID of cron job policy.
	//
	// Return:
	//  error if failed to unschedule
	UnSchedule(cronJobPolicyID string) error

	// Load and cache data if needed
	//
	// Return:
	//  error if failed to do
	Load() error

	// Clear all the cron job policies.
	//
	// Return:
	//  error if failed to do
	Clear() error

	// Start to serve
	Start()

	// Accept the pushed policy and cache it
	//
	// policy *PeriodicJobPolicy : the periodic policy being accept
	//
	// Return:
	//  error if failed to do
	AcceptPeriodicPolicy(policy *PeriodicJobPolicy) error

	// Remove the specified policy from the cache if it is existing
	//
	// policyID string : ID of the policy being removed
	//
	// Return:
	//  the ptr of the being deletd policy
	RemovePeriodicPolicy(policyID string) *PeriodicJobPolicy
}
