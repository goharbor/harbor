// Copyright 2018 The Harbor Authors. All rights reserved.

package period

import (
	"encoding/json"
	"sync"

	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

const (
	//periodicJobPolicyChangeEventSchedule : Schedule periodic job policy event
	periodicJobPolicyChangeEventSchedule = "Schedule"
	//periodicJobPolicyChangeEventUnSchedule : UnSchedule periodic job policy event
	periodicJobPolicyChangeEventUnSchedule = "UnSchedule"
)

//periodicJobPolicy ...
type periodicJobPolicy struct {
	//NOTES: The 'PolicyID' should not be set when serialize this policy struct to the zset
	//because each 'Policy ID' is different and it may cause issue of losing zset unique capability.
	PolicyID      string                 `json:"policy_id,omitempty"`
	JobName       string                 `json:"job_name"`
	JobParameters map[string]interface{} `json:"job_params"`
	CronSpec      string                 `json:"cron_spec"`
}

//periodicJobPolicyEvent is the event content of periodic job policy change.
type periodicJobPolicyEvent struct {
	Event             string             `json:"event"`
	PeriodicJobPolicy *periodicJobPolicy `json:"periodic_job_policy"`
}

//serialize the policy to raw data.
func (pjp *periodicJobPolicy) serialize() ([]byte, error) {
	return json.Marshal(pjp)
}

//deSerialize the raw json to policy.
func (pjp *periodicJobPolicy) deSerialize(rawJSON []byte) error {
	return json.Unmarshal(rawJSON, pjp)
}

//serialize the policy to raw data.
func (pjpe *periodicJobPolicyEvent) serialize() ([]byte, error) {
	return json.Marshal(pjpe)
}

//deSerialize the raw json to policy.
func (pjpe *periodicJobPolicyEvent) deSerialize(rawJSON []byte) error {
	return json.Unmarshal(rawJSON, pjpe)
}

//periodicJobPolicyStore is in-memory cache for the periodic job policies.
type periodicJobPolicyStore struct {
	lock     *sync.RWMutex
	policies map[string]*periodicJobPolicy //k-v pair and key is the policy ID
}

func (ps *periodicJobPolicyStore) addAll(items []*periodicJobPolicy) {
	if items == nil || len(items) == 0 {
		return
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	for _, item := range items {
		//Ignore the item with empty uuid
		if !utils.IsEmptyStr(item.PolicyID) {
			ps.policies[item.PolicyID] = item
		}
	}
}

func (ps *periodicJobPolicyStore) list() []*periodicJobPolicy {
	allItems := make([]*periodicJobPolicy, 0)

	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for _, v := range ps.policies {
		allItems = append(allItems, v)
	}

	return allItems
}

func (ps *periodicJobPolicyStore) add(jobPolicy *periodicJobPolicy) {
	if jobPolicy == nil || utils.IsEmptyStr(jobPolicy.PolicyID) {
		return
	}

	ps.lock.Lock()
	defer ps.lock.Unlock()

	ps.policies[jobPolicy.PolicyID] = jobPolicy
}

func (ps *periodicJobPolicyStore) remove(policyID string) *periodicJobPolicy {
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
