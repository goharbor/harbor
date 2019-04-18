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

package lcm

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// Controller is designed to control the life cycle of the job
type Controller interface {
	// New tracker from the new provided stats
	New(stats *job.Stats) (job.Tracker, error)

	// Track the life cycle of the specified existing job
	Track(jobID string) (job.Tracker, error)
}

// basicController is default implementation of Controller based on redis
type basicController struct {
	context   context.Context
	namespace string
	pool      *redis.Pool
	callback  job.HookCallback
}

// NewController is the constructor of basic controller
func NewController(ctx context.Context, ns string, pool *redis.Pool, callback job.HookCallback) Controller {
	return &basicController{
		context:   ctx,
		namespace: ns,
		pool:      pool,
		callback:  callback,
	}
}

// New tracker
func (bc *basicController) New(stats *job.Stats) (job.Tracker, error) {
	if stats == nil {
		return nil, errors.New("nil stats when creating job tracker")
	}

	if err := stats.Validate(); err != nil {
		return nil, errors.Errorf("error occurred when creating job tracker: %s", err)
	}

	bt := job.NewBasicTrackerWithStats(stats, bc.context, bc.namespace, bc.pool, bc.callback)
	if err := bt.Save(); err != nil {
		return nil, err
	}

	return bt, nil
}

// Track and attache with the job
func (bc *basicController) Track(jobID string) (job.Tracker, error) {
	bt := job.NewBasicTrackerWithID(jobID, bc.context, bc.namespace, bc.pool, bc.callback)
	if err := bt.Load(); err != nil {
		return nil, err
	}

	return bt, nil
}
