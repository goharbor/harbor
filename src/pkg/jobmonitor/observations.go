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

package jobmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	libRedis "github.com/goharbor/harbor/src/lib/redis"
)

type ObservationManager interface {
	// ObservationByJobNameAndPolicyID scans and filters active observations
	ObservationByJobNameAndPolicyID(ctx context.Context, jobName string, policyID int64) (observation *work.WorkerObservation, err error)
}

type ObservationManagerImpl struct {
}

func NewObservationManagerImpl() *ObservationManagerImpl {
	return &ObservationManagerImpl{}
}

func (m *ObservationManagerImpl) ObservationByJobNameAndPolicyID(_ context.Context, jobName string, policyID int64) (observation *work.WorkerObservation, err error) {
	monitorClient, err := GetJobServiceMonitorClient()
	if err != nil {
		return nil, errors.New(nil).WithCode(errors.PreconditionCode).WithMessagef("unable to get job monitor's client: %v", err)
	}
	observations, err := monitorClient.WorkerObservations()
	if err != nil {
		return nil, errors.New(nil).WithCode(errors.PreconditionCode).WithMessagef("unable to get jobs observations: %v", err)
	}
	for _, o := range observations {
		if observationMatch(o, jobName, policyID) {
			return o, nil
		}
	}
	return nil, nil
}

func observationMatch(o *work.WorkerObservation, jobName string, policyID int64) bool {
	if o.JobName != jobName {
		return false
	}
	args := map[string]interface{}{}
	if err := json.Unmarshal([]byte(o.ArgsJSON), &args); err != nil {
		return false
	}
	policyIDFromArgs, ok := args["policy_id"].(float64)
	return ok && int64(policyIDFromArgs) == policyID
}

func GetJobServiceMonitorClient() (JobServiceMonitorClient, error) {
	cfg, err := job.GlobalClient.GetJobServiceConfig()
	if err != nil {
		return nil, err
	}
	config := cfg.RedisPoolConfig
	pool, err := libRedis.GetRedisPool(JobServicePool, config.RedisURL, &libRedis.PoolParam{
		PoolMaxIdle:     0,
		PoolIdleTimeout: time.Duration(config.IdleTimeoutSecond) * time.Second,
	})
	if err != nil {
		log.Errorf("failed to get redis pool: %v", err)
		return nil, err
	}
	return work.NewClient(fmt.Sprintf("{%s}", config.Namespace), pool), nil
}
