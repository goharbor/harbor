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
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"

	comUtils "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

// Policy ...
type Policy struct {
	// Policy can be treated as job template of periodic job.
	// The info of policy will be copied into the scheduled job executions for the periodic job.
	ID            string         `json:"id"`
	JobName       string         `json:"job_name"`
	CronSpec      string         `json:"cron_spec"`
	JobParameters map[string]any `json:"job_params,omitempty"`
	WebHookURL    string         `json:"web_hook_url,omitempty"`
	NumericID     int64          `json:"numeric_id,omitempty"`
}

// Serialize the policy to raw data.
func (p *Policy) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// DeSerialize the raw json to policy.
func (p *Policy) DeSerialize(rawJSON []byte) error {
	return json.Unmarshal(rawJSON, p)
}

// Validate the policy
func (p *Policy) Validate() error {
	if utils.IsEmptyStr(p.ID) {
		return errors.New("missing ID in the periodic job policy object")
	}

	if utils.IsEmptyStr(p.JobName) {
		return errors.New("missing job name in the periodic job policy object")
	}

	if !utils.IsEmptyStr(p.WebHookURL) {
		if !utils.IsValidURL(p.WebHookURL) {
			return fmt.Errorf("bad web hook URL: %s", p.WebHookURL)
		}
	}

	if _, err := comUtils.CronParser().Parse(p.CronSpec); err != nil {
		return err
	}

	return nil
}

type policyWithScore struct {
	RawData []byte
	Score   int64
}

// Load all the policies from the backend storage.
func Load(namespace string, conn redis.Conn) ([]*Policy, error) {
	bytes, err := redis.Values(conn.Do("ZRANGE", rds.KeyPeriodicPolicy(namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return nil, err
	}

	policyWithScores := make([]policyWithScore, 0)
	if err := redis.ScanSlice(bytes, &policyWithScores); err != nil {
		return nil, err
	}

	policies := make([]*Policy, 0)
	for _, pw := range policyWithScores {
		p := &Policy{}

		if err := p.DeSerialize(pw.RawData); err != nil {
			// Ignore error which means the policy data is not valid
			// Only logged
			logger.Errorf("Malformed policy: %s; error: %s", pw.RawData, err)
			continue
		}

		// Validate the policy object
		if err := p.Validate(); err != nil {
			logger.Errorf("Policy validate error: %s", err)
			continue
		}

		p.NumericID = pw.Score
		policies = append(policies, p)

		logger.Debugf("Load periodic job policy: %s", string(pw.RawData))
	}

	logger.Debugf("Load %d periodic job policies", len(policies))

	return policies, nil
}
