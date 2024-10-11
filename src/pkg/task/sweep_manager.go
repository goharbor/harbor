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
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// SweepMgr is a global sweep manager instance.
	SweepMgr = NewSweepManager()

	timeFormat      = "2006-01-02 15:04:05.999999999"
	defaultPageSize = 100000
	finalStatusCode = 3
)

type SweepManager interface {
	// ListCandidates lists the candidate execution ids which met the sweep criteria.
	ListCandidates(ctx context.Context, vendorType string, retainCnt int64) (execIDs []int64, err error)
	// Clean deletes the tasks belonging to the execution which in final status and deletes executions.
	Clean(ctx context.Context, execID []int64) (err error)
	// FixDanglingStateExecution fixes the dangling state execution.
	FixDanglingStateExecution(ctx context.Context) error
}

// sweepManager implements the interface SweepManager.
type sweepManager struct {
	execDAO dao.ExecutionDAO
}

// listVendorIDs lists distinct vendor ids by vendor type.
func (sm *sweepManager) listVendorIDs(ctx context.Context, vendorType string) ([]int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	var ids []int64
	if _, err = ormer.Raw(`SELECT DISTINCT vendor_id FROM execution WHERE vendor_type = ?`, vendorType).QueryRows(&ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// getCandidateMaxStartTime returns the max start time for candidate executions, obtain the start time of the xth recent one.
func (sm *sweepManager) getCandidateMaxStartTime(ctx context.Context, vendorType string, vendorID, retainCnt int64) (*time.Time, error) {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"VendorType": vendorType,
			"VendorID":   vendorID,
		},
		Sorts: []*q.Sort{
			{
				Key:  "StartTime",
				DESC: true,
			}},
		PageSize:   1,
		PageNumber: retainCnt,
	}
	executions, err := sm.execDAO.List(ctx, query)
	if err != nil {
		return nil, err
	}
	// list is null means that the execution count < retainCnt, return nil time
	if len(executions) == 0 {
		return nil, nil
	}

	return &executions[0].StartTime, nil
}

func (sm *sweepManager) ListCandidates(ctx context.Context, vendorType string, retainCnt int64) ([]int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	vendorIDs, err := sm.listVendorIDs(ctx, vendorType)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list vendor ids for vendor type %s", vendorType)
	}

	// execIDs stores the result
	var execIDs []int64
	for _, vendorID := range vendorIDs {
		maxStartTime, err := sm.getCandidateMaxStartTime(ctx, vendorType, vendorID, retainCnt)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get candidate max start time, vendor type: %s, vendor id: %d", vendorType, vendorID)
		}
		// continue if no max start time got that means no candidate executions
		if maxStartTime == nil {
			continue
		}
		// candidate criteria
		// 1. exact vendor type & vendor id
		// 2. start_time is before the max start time
		// 3. status is the final state
		// count the records for pagination
		sql := `SELECT COUNT(1) FROM execution WHERE vendor_type = ? AND vendor_id = ? AND start_time < ? AND status IN (?,?,?)`
		totalOfCandidate := 0
		params := []interface{}{
			vendorType,
			vendorID,
			maxStartTime.Format(timeFormat),
			// final status should in Error/Success/Stopped
			job.ErrorStatus.String(),
			job.SuccessStatus.String(),
			job.StoppedStatus.String(),
		}
		if err = ormer.Raw(sql, params...).QueryRow(&totalOfCandidate); err != nil {
			return nil, errors.Wrapf(err, "failed to count candidates, vendor type: %s, vendor id: %d", vendorType, vendorID)
		}
		// n is the page count of all candidates
		n := totalOfCandidate / defaultPageSize
		if totalOfCandidate%defaultPageSize > 0 {
			n = n + 1
		}

		sql = `SELECT id FROM execution WHERE vendor_type = ? AND vendor_id = ? AND start_time < ? AND status IN (?,?,?) ORDER BY id`
		// default page size is 100000
		q2 := &q.Query{PageSize: int64(defaultPageSize)}
		for i := n; i >= 1; i-- {
			q2.PageNumber = int64(i)
			// should copy params as pagination will append the slice
			paginationParams := make([]interface{}, len(params))
			copy(paginationParams, params)
			paginationSQL, paginationParams := orm.PaginationOnRawSQL(q2, sql, paginationParams)
			ids := make([]int64, 0, defaultPageSize)
			if _, err = ormer.Raw(paginationSQL, paginationParams...).QueryRows(&ids); err != nil {
				return nil, errors.Wrapf(err, "failed to list candidate execution ids, vendor type: %s, vendor id: %d", vendorType, vendorID)
			}
			execIDs = append(execIDs, ids...)
		}
	}

	return execIDs, nil
}

func (sm *sweepManager) Clean(ctx context.Context, execIDs []int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	// construct sql params
	params := make([]interface{}, 0, len(execIDs))
	for _, eid := range execIDs {
		params = append(params, eid)
	}
	// delete tasks
	sql := fmt.Sprintf("DELETE FROM task WHERE status_code = %d AND execution_id IN (%s)", finalStatusCode, orm.ParamPlaceholderForIn(len(params)))
	_, err = ormer.Raw(sql, params...).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to delete tasks")
	}
	// delete executions
	sql = fmt.Sprintf("DELETE FROM execution WHERE id IN (%s) AND id NOT IN (SELECT DISTINCT execution_id FROM task)", orm.ParamPlaceholderForIn(len(params)))
	_, err = ormer.Raw(sql, params...).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to delete executions")
	}

	return nil
}

// FixDanglingStateExecution update executions always running
func (sm *sweepManager) FixDanglingStateExecution(ctx context.Context) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	sql := `UPDATE execution
		SET status =
        CASE
            WHEN EXISTS (SELECT 1 FROM task WHERE execution_id = execution.id AND status = 'Error') THEN 'Error'
            WHEN EXISTS (SELECT 1 FROM task WHERE execution_id = execution.id AND status = 'Stopped') THEN 'Stopped'
            ELSE 'Success'
            END
WHERE status = 'Running'
  AND EXTRACT(epoch FROM NOW() - start_time)/3600 >  ?
  AND NOT EXISTS (SELECT 1 FROM task WHERE execution_id = execution.id AND status = 'Running')`

	_, err = ormer.Raw(sql, config.MaxDanglingHour()).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to fix dangling state execution")
	}
	return nil
}

func NewSweepManager() SweepManager {
	return &sweepManager{
		execDAO: dao.NewExecutionDAO(),
	}
}
