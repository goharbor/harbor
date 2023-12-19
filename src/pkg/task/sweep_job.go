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

package task

import (
	"encoding/json"
	"os"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

func init() {
	// the default batch size is 65535
	sweepBatchSize = 65535
	envBatchSize := os.Getenv("EXECUTION_SWEEP_BATCH_SIZE")
	if len(envBatchSize) > 0 {
		batchSize, err := strconv.Atoi(envBatchSize)
		if err != nil {
			logger.Errorf("failed to parse the batch size from env, value: %s, error: %v", envBatchSize, err)
		} else {
			if batchSize <= 0 || batchSize > 65535 {
				logger.Errorf("invalid batch size %d for sweep job, should be positive and not over than 65535", batchSize)
			} else {
				// override the batch size if provided is valid
				sweepBatchSize = batchSize
			}
		}
	}
}

var (
	// notice that the batch size should not over than 65535 as length limitation of postgres parameters
	sweepBatchSize int
	errStop        = errors.New("stopped")
)

const (
	// ExecRetainCounts is the params key of execution retain count
	ExecRetainCounts = "execution_retain_counts"
)

// SweepJob used to cleanup the executions and tasks for different vendors.
type SweepJob struct {
	execRetainCountsMap map[string]int64
	logger              logger.Interface
	mgr                 SweepManager
}

// MaxFails of sweep job. Don't need to retry.
func (sj *SweepJob) MaxFails() uint {
	return 1
}

// MaxCurrency limit 1 concurrency of sweep job.
func (sj *SweepJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry indicates no need to retry sweep job.
func (sj *SweepJob) ShouldRetry() bool {
	return false
}

// Validate the parameters of preheat job.
func (sj *SweepJob) Validate(_ job.Parameters) error {
	return nil
}

// Run the sweep process.
func (sj *SweepJob) Run(ctx job.Context, params job.Parameters) error {
	if err := sj.init(ctx, params); err != nil {
		return err
	}

	sj.logger.Info("start to run sweep job")

	if err := sj.mgr.FixDanglingStateExecution(ctx.SystemContext()); err != nil {
		sj.logger.Errorf("failed to fix dangling state executions, error: %v", err)
	}

	var errs errors.Errors
	for vendor, cnt := range sj.execRetainCountsMap {
		if sj.shouldStop(ctx) {
			sj.logger.Info("received the stop signal, quit sweep job")
			return nil
		}

		if err := sj.sweep(ctx, vendor, cnt); err != nil {
			if err == errStop {
				sj.logger.Info("received the stop signal, quit sweep job")
				return nil
			}

			sj.logger.Errorf("[%s] failed to run sweep, error: %v", vendor, err)
			errs = append(errs, err)
		}
	}

	sj.logger.Info("end to run sweep job")

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (sj *SweepJob) init(ctx job.Context, params job.Parameters) error {
	if sj.mgr == nil {
		// use global manager if no sweep manager found
		sj.mgr = SweepMgr
	}
	sj.logger = ctx.GetLogger()
	sj.parseParams(params)
	return nil
}

func (sj *SweepJob) parseParams(params job.Parameters) {
	sj.execRetainCountsMap = make(map[string]int64)
	execRetainCounts, err := json.Marshal(params[ExecRetainCounts])
	if err != nil {
		sj.logger.Errorf("failed to marshal params %+v, error: %v", params[ExecRetainCounts], err)
		return
	}

	if err = json.Unmarshal(execRetainCounts, &sj.execRetainCountsMap); err != nil {
		sj.logger.Errorf("failed to unmarshal params %s, error: %v", string(execRetainCounts), err)
		return
	}
}

// sweep cleanup the executions/tasks by vendor type and retain count.
func (sj *SweepJob) sweep(ctx job.Context, vendorType string, retainCount int64) error {
	sj.logger.Infof("[%s] start to sweep, retain latest %d executions", vendorType, retainCount)

	start := time.Now()
	candidates, err := sj.mgr.ListCandidates(ctx.SystemContext(), vendorType, retainCount)
	if err != nil {
		sj.logger.Errorf("[%s] failed to list candidates, error: %v", vendorType, err)
		return err
	}

	total := len(candidates)
	sj.logger.Infof("[%s] listed %d candidate executions for sweep", vendorType, total)
	// batch clean the executions and tasks
	for i := 0; i < total; i += sweepBatchSize {
		// checkpoint
		if sj.shouldStop(ctx) {
			return errStop
		}
		// calculate the batch position
		j := i + sweepBatchSize
		// avoid overflow
		if j > total {
			j = total
		}

		if err = sj.mgr.Clean(ctx.SystemContext(), candidates[i:j]); err != nil {
			sj.logger.Errorf("[%s] failed to batch clean candidates, error: %v", vendorType, err)
			return err
		}
	}

	sj.logger.Infof("[%s] end to sweep, %d executions were deleted in total, elapsed time: %v", vendorType, total, time.Since(start))

	return nil
}

func (sj *SweepJob) shouldStop(ctx job.Context) bool {
	opCmd, exit := ctx.OPCommand()
	if exit && opCmd.IsStop() {
		return true
	}
	return false
}
