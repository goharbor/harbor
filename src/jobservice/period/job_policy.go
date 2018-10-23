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

import (
	"encoding/json"
	"sync"

	"github.com/goharbor/harbor/src/jobservice/utils"
)

const (
	// periodicJobPolicyChangeEventSchedule : Schedule periodic job policy event
	periodicJobPolicyChangeEventSchedule = "Schedule"
	// periodicJobPolicyChangeEventUnSchedule : UnSchedule periodic job policy event
	periodicJobPolicyChangeEventUnSchedule = "UnSchedule"
)

// PeriodicJobPolicy ...
type PeriodicJobPolicy struct {
	// NOTES: The 'PolicyID' should not be set when serialize this policy struct to the zset
	// because each 'Policy ID' is different and it may cause issue of losing zset unique capability.
	PolicyID      string                 `json:"policy_id,omitempty"`
	JobName       string                 `json:"job_name"`
	JobParameters map[string]interface{} `json:"job_params"`
	CronSpec      string                 `json:"cron_spec"`
}

// Serialize the policy to raw data.
func (pjp *PeriodicJobPolicy) Serialize() ([]byte, error) {
	return json.Marshal(pjp)
}

// DeSerialize the raw json to policy.
func (pjp *PeriodicJobPolicy) DeSerialize(rawJSON []byte) error {
	return json.Unmarshal(rawJSON, pjp)
}

// periodicJobPolicyStore is in-memory cache for the periodic job policies.
type periodicJobPolicyStore struct {
	lock     *sync.RWMutex
	policies map[string]*PeriodicJobPolicy // k-v pair and key is the policy ID
}

func (ps *periodicJobPolicyStore) addAll(items []*PeriodicJobPolicy) {
	if items == nil || len(items) == 0 {
		return
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, item := range items {
		// Ignore the item with empty uuid
		if !utils.IsEmptyStr(item.PolicyID) {
			ps.policies[item.PolicyID] = item
		}
	}
}

func (ps *periodicJobPolicyStore) list() []*PeriodicJobPolicy {
	allItems := make([]*PeriodicJobPolicy, 0)

	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for _, v := range ps.policies {
		allItems = append(allItems, v)
	}

	return allItems
}

func (ps *periodicJobPolicyStore) add(jobPolicy *PeriodicJobPolicy) {
	if jobPolicy == nil || utils.IsEmptyStr(jobPolicy.PolicyID) {
		return
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.policies[jobPolicy.PolicyID] = jobPolicy
}

func (ps *periodicJobPolicyStore) remove(policyID string) *PeriodicJobPolicy {
	if utils.IsEmptyStr(policyID) {
		return nil
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	if item, ok := ps.policies[policyID]; ok {
		delete(ps.policies, policyID)
		return item
	}

	return nil
}

func (ps *periodicJobPolicyStore) size() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.policies)
}
