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
	"encoding/json"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"sync"
	"time"
)

const (
	shortLoopInterval = 5 * time.Second
	longLoopInterval  = 5 * time.Minute
)

// Controller is designed to control the life cycle of the job
type Controller interface {
	// Run daemon process if needed
	Serve() error

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
	wg        *sync.WaitGroup
}

// NewController is the constructor of basic controller
func NewController(ctx *env.Context, ns string, pool *redis.Pool, callback job.HookCallback) Controller {
	return &basicController{
		context:   ctx.SystemContext,
		namespace: ns,
		pool:      pool,
		callback:  callback,
		wg:        ctx.WG,
	}
}

// Serve ...
func (bc *basicController) Serve() error {
	go bc.loopForRestoreDeadStatus()
	logger.Info("Status restoring loop is started")

	return nil
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

// loopForRestoreDeadStatus is a loop to restore the dead states of jobs
func (bc *basicController) loopForRestoreDeadStatus() {
	defer func() {
		logger.Info("Status restoring loop is stopped")
		bc.wg.Done()
	}()

	token := make(chan bool, 1)
	token <- true

	bc.wg.Add(1)
	for {
		<-token

		if err := bc.restoreDeadStatus(); err != nil {
			wait := shortLoopInterval
			if err == redis.ErrNil {
				// No elements
				wait = longLoopInterval
			}
			// wait for a while or be terminated
			select {
			case <-time.After(wait):
			case <-bc.context.Done():
				return
			}
		}

		// Return token
		token <- true
	}
}

// restoreDeadStatus try to restore the dead status
func (bc *basicController) restoreDeadStatus() error {
	// Get one
	deadOne, err := bc.popOneDead()
	if err != nil {
		return err
	}
	// Try to update status
	t, err := bc.Track(deadOne.JobID)
	if err != nil {
		return err
	}

	return t.UpdateStatusWithRetry(job.Status(deadOne.TargetStatus))
}

// popOneDead retrieves one dead status from the backend Q from lowest to highest
func (bc *basicController) popOneDead() (*job.SimpleStatusChange, error) {
	conn := bc.pool.Get()
	defer conn.Close()

	key := rds.KeyStatusUpdateRetryQueue(bc.namespace)
	v, err := rds.ZPopMin(conn, key)
	if err != nil {
		return nil, err
	}

	if bytes, ok := v.([]byte); ok {
		ssc := &job.SimpleStatusChange{}
		if err := json.Unmarshal(bytes, ssc); err == nil {
			return ssc, nil
		}
	}

	return nil, errors.New("pop one dead error: bad result reply")
}
