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

package dao

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/gtask"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"

	// init the db config
	_ "github.com/goharbor/harbor/src/pkg/config/db"
)

func init() {
	// register the execution status refresh task if enable the async update
	if interval := config.GetExecutionStatusRefreshIntervalSeconds(); interval > 0 {
		gtask.DefaultPool().AddTask(scanAndRefreshOutdateStatus, time.Duration(interval)*time.Second)
	}
}

func RegisterExecutionStatusChangePostFunc(vendor string, fc ExecutionStatusChangePostFunc) {
	executionStatusChangePostFuncRegistry[vendor] = fc
}

var (
	// ExecDAO is the global execution dao
	ExecDAO                               = NewExecutionDAO()
	executionStatusChangePostFuncRegistry = map[string]ExecutionStatusChangePostFunc{}
	// execStatusOutdateKeyRegex is the regex for the execution status outdate key,
	// the regex used to parse exec id and vendor type from the key.
	// e.g. execution:id:100:vendor:REPLICATION:status_outdate
	execStatusOutdateKeyRegex = regexp.MustCompile(`execution:id:(\d+):vendor:([A-Z0-9_]+):status_outdate`)
)

// ExecutionStatusChangePostFunc is the function called after the execution status changed
type ExecutionStatusChangePostFunc func(ctx context.Context, executionID int64, status string) (err error)

// ExecutionDAO is the data access object interface for execution
type ExecutionDAO interface {
	// Count returns the total count of executions according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List the executions according to the query
	// Query the "ExtraAttrs" by setting 'query.Keywords["ExtraAttrs.key"]="value"'
	List(ctx context.Context, query *q.Query) (executions []*Execution, err error)
	// Get the specified execution
	Get(ctx context.Context, id int64) (execution *Execution, err error)
	// Create an execution
	Create(ctx context.Context, execution *Execution) (id int64, err error)
	// Update the specified execution. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, execution *Execution, props ...string) (err error)
	// Delete the specified execution
	Delete(ctx context.Context, id int64) (err error)
	// GetMetrics returns the task metrics for the specified execution
	GetMetrics(ctx context.Context, id int64) (metrics *Metrics, err error)
	// RefreshStatus refreshes the status of the specified execution according to it's tasks. If it's status
	// is final, update the end time as well
	// If the status is changed, the returning "statusChanged" is set as "true" and the current status indicates
	// the changed status
	RefreshStatus(ctx context.Context, id int64) (statusChanged bool, currentStatus string, err error)
	// AsyncRefreshStatus refreshes the status of the specified execution in the async mode, which will register
	// a update flag in the redis and then wait for global periodic job to scan and update the status to db finally.
	AsyncRefreshStatus(ctx context.Context, id int64, vendor string) (err error)
}

// NewExecutionDAO returns an instance of ExecutionDAO
func NewExecutionDAO() ExecutionDAO {
	return &executionDAO{
		taskDAO: NewTaskDAO(),
	}
}

type executionDAO struct {
	taskDAO TaskDAO
}

func (e *executionDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := e.querySetter(ctx, query, orm.WithSortDisabled(true))
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (e *executionDAO) List(ctx context.Context, query *q.Query) ([]*Execution, error) {
	executions := []*Execution{}
	qs, err := e.querySetter(ctx, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&executions); err != nil {
		return nil, err
	}
	return executions, nil
}

func (e *executionDAO) Get(ctx context.Context, id int64) (*Execution, error) {
	execution := &Execution{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(execution); err != nil {
		if e := orm.AsNotFoundError(err, "execution %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return execution, nil
}

func (e *executionDAO) Create(ctx context.Context, execution *Execution) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return ormer.Insert(execution)
}

func (e *executionDAO) Update(ctx context.Context, execution *Execution, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(execution, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("execution %d not found", execution.ID)
	}
	return nil
}

func (e *executionDAO) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Execution{
		ID: id,
	})
	if err != nil {
		if e := orm.AsForeignKeyError(err,
			"the execution %d is referenced by other resources", id); e != nil {
			err = e
		}
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("execution %d not found", id)
	}
	return nil
}

func (e *executionDAO) GetMetrics(ctx context.Context, id int64) (*Metrics, error) {
	scs, err := e.taskDAO.ListStatusCount(ctx, id)
	if err != nil {
		return nil, err
	}
	metrics := &Metrics{}
	if len(scs) == 0 {
		return metrics, nil
	}

	for _, sc := range scs {
		switch sc.Status {
		case job.SuccessStatus.String():
			metrics.SuccessTaskCount = sc.Count
		case job.ErrorStatus.String():
			metrics.ErrorTaskCount = sc.Count
		case job.PendingStatus.String():
			metrics.PendingTaskCount = sc.Count
		case job.RunningStatus.String():
			metrics.RunningTaskCount = sc.Count
		case job.ScheduledStatus.String():
			metrics.ScheduledTaskCount = sc.Count
		case job.StoppedStatus.String():
			metrics.StoppedTaskCount = sc.Count
		default:
			log.Errorf("unknown task status: %s", sc.Status)
		}
	}
	metrics.TaskCount = metrics.SuccessTaskCount + metrics.ErrorTaskCount +
		metrics.PendingTaskCount + metrics.RunningTaskCount +
		metrics.ScheduledTaskCount + metrics.StoppedTaskCount
	return metrics, nil
}

func (e *executionDAO) RefreshStatus(ctx context.Context, id int64) (bool, string, error) {
	// as the status of the execution can be refreshed by multiple operators concurrently
	// we use the optimistic locking to avoid the conflict and retry 5 times at most
	for i := 0; i < 5; i++ {
		statusChanged, currentStatus, retry, err := e.refreshStatus(ctx, id)
		if err != nil {
			return false, "", err
		}
		if !retry {
			return statusChanged, currentStatus, nil
		}
	}
	return false, "", fmt.Errorf("failed to refresh the status of the execution %d after %d retries", id, 5)
}

// the returning values:
// 1. bool: is the status changed
// 2. string: the current status if changed
// 3. bool: whether a retry is needed
// 4. error: the error
func (e *executionDAO) refreshStatus(ctx context.Context, id int64) (bool, string, bool, error) {
	execution, err := e.Get(ctx, id)
	if err != nil {
		return false, "", false, err
	}
	metrics, err := e.GetMetrics(ctx, id)
	if err != nil {
		return false, "", false, err
	}
	// no task, return directly
	if metrics.TaskCount == 0 {
		return false, "", false, nil
	}

	var status string
	if metrics.PendingTaskCount > 0 || metrics.RunningTaskCount > 0 || metrics.ScheduledTaskCount > 0 {
		status = job.RunningStatus.String()
	} else if metrics.ErrorTaskCount > 0 {
		status = job.ErrorStatus.String()
	} else if metrics.StoppedTaskCount > 0 {
		status = job.StoppedStatus.String()
	} else if metrics.SuccessTaskCount > 0 {
		status = job.SuccessStatus.String()
	}

	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return false, "", false, err
	}

	sql := `update execution set status = ?, revision = revision+1, update_time = ? where id = ? and revision = ?`
	result, err := ormer.Raw(sql, status, time.Now(), id, execution.Revision).Exec()
	if err != nil {
		return false, "", false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, "", false, err
	}

	// if the count of affected rows is 0, that means the execution is updating by others, retry
	if n == 0 {
		return false, "", true, nil
	}

	/* this is another solution to solve the concurrency issue for refreshing the execution status
	// set a score for each status:
	// 		pending, running, scheduled - 4
	// 		error - 3
	//		stopped - 2
	//		success - 1
	// and set the status of record with highest score as the status of execution
	sql := `with status_score as (
				select status,
					case
						when status='%s' or status='%s' or status='%s' then 4
						when status='%s' then 3
						when status='%s' then 2
						when status='%s' then 1
						else 0
					end as score
				from task
				where execution_id=?
				group by status
			)
			update execution
			set status=(
				select
					case
						when max(score)=4 then '%s'
						when max(score)=3 then '%s'
						when max(score)=2 then '%s'
						when max(score)=1 then '%s'
						when max(score)=0 then ''
					end as status
				from status_score)
			where id = ?`
	sql = fmt.Sprintf(sql, job.PendingStatus.String(), job.RunningStatus.String(), job.ScheduledStatus.String(),
		job.ErrorStatus.String(), job.StoppedStatus.String(), job.SuccessStatus.String(),
		job.RunningStatus.String(), job.ErrorStatus.String(), job.StoppedStatus.String(), job.SuccessStatus.String())
	if _, err = ormer.Raw(sql, id, id).Exec(); err != nil {
		return err
	}
	*/

	// update the end time if the status is final, otherwise set the end time as NULL, this is useful
	// for retrying jobs
	sql = `update execution
			set end_time = (
				case
					when status='%s' or status='%s' or status='%s' then  (
						select max(end_time)
						from task
						where execution_id=?)
					else NULL
				end)
			where id=?`
	sql = fmt.Sprintf(sql, job.ErrorStatus.String(), job.StoppedStatus.String(), job.SuccessStatus.String())
	_, err = ormer.Raw(sql, id, id).Exec()
	return status != execution.Status, status, false, err
}

type jsonbStru struct {
	keyPrefix string
	key       string
	value     interface{}
}

func (e *executionDAO) querySetter(ctx context.Context, query *q.Query, options ...orm.Option) (orm.QuerySeter, error) {
	qs, err := orm.QuerySetter(ctx, &Execution{}, query, options...)
	if err != nil {
		return nil, err
	}

	// append the filter for "extra attrs"
	if query != nil && len(query.Keywords) > 0 {
		var (
			jsonbStrus []jsonbStru
			args       []interface{}
		)

		for key, value := range query.Keywords {
			if strings.HasPrefix(key, "ExtraAttrs.") && key != "ExtraAttrs." {
				jsonbStrus = append(jsonbStrus, jsonbStru{
					keyPrefix: "ExtraAttrs.",
					key:       key,
					value:     value,
				})
			}
			if strings.HasPrefix(key, "extra_attrs.") && key != "extra_attrs." {
				jsonbStrus = append(jsonbStrus, jsonbStru{
					keyPrefix: "extra_attrs.",
					key:       key,
					value:     value,
				})
			}
		}
		if len(jsonbStrus) == 0 {
			return qs, nil
		}

		idSQL, args := buildInClauseSQLForExtraAttrs(jsonbStrus)
		inClause, err := orm.CreateInClause(ctx, idSQL, args...)
		if err != nil {
			return nil, err
		}
		qs = qs.FilterRaw("id", inClause)
	}

	return qs, nil
}

// Param keys is strings.Split() after trim "extra_attrs."/"ExtraAttrs." prefix
// key with keyPrefix supports multi-level query operator on PostgreSQL JSON data
// examples:
// key = extra_attrs.id,
//
//	==> sql = "select id from execution where extra_attrs->>?=?", args = {id, value}
//
// key = extra_attrs.artifact.digest
//
//	==> sql = "select id from execution where extra_attrs->?->>?=?", args = {artifact, id, value}
//
// key = extra_attrs.a.b.c
//
//	==> sql = "select id from execution where extra_attrs->?->?->>?=?", args = {a, b, c, value}
func buildInClauseSQLForExtraAttrs(jsonbStrus []jsonbStru) (string, []interface{}) {
	if len(jsonbStrus) == 0 {
		return "", nil
	}

	var cond string
	var args []interface{}
	sql := "select id from execution where"

	for i, jsonbStr := range jsonbStrus {
		if jsonbStr.key == "" || jsonbStr.value == "" {
			return "", nil
		}
		keys := strings.Split(strings.TrimPrefix(jsonbStr.key, jsonbStr.keyPrefix), ".")
		if len(keys) == 1 {
			if i == 0 {
				cond += "extra_attrs->>?=?"
			} else {
				cond += " and extra_attrs->>?=?"
			}
		}
		if len(keys) >= 2 {
			elements := make([]string, len(keys)-1)
			for i := range elements {
				elements[i] = "?"
			}
			s := strings.Join(elements, "->")
			if i == 0 {
				cond += fmt.Sprintf("extra_attrs->%s->>?=?", s)
			} else {
				cond += fmt.Sprintf(" and extra_attrs->%s->>?=?", s)
			}
		}

		for _, item := range keys {
			args = append(args, item)
		}
		args = append(args, jsonbStr.value)
	}

	return fmt.Sprintf("%s %s", sql, cond), args
}

func buildExecStatusOutdateKey(id int64, vendor string) string {
	return fmt.Sprintf("execution:id:%d:vendor:%s:status_outdate", id, vendor)
}

func extractExecIDVendorFromKey(key string) (int64, string, error) {
	// input: execution:id:100:vendor:GARBAGE_COLLECTION:status_outdate
	// output: [execution:id:100:vendor:GARBAGE_COLLECTION:status_outdate 100 GARBAGE_COLLECTION]
	matches := execStatusOutdateKeyRegex.FindStringSubmatch(key)
	if len(matches) < 3 {
		return 0, "", errors.Errorf("invalid format: %s", key)
	}

	id, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, matches[2], err
	}

	return id, matches[2], nil
}

func (e *executionDAO) AsyncRefreshStatus(ctx context.Context, id int64, vendor string) (err error) {
	key := buildExecStatusOutdateKey(id, vendor)
	if cache.Default().Contains(ctx, key) {
		// return earlier if already have the key
		return nil
	}
	// save the key to redis, the value is useless so set it to empty
	return cache.Default().Save(ctx, key, "")
}

// scanAndRefreshOutdateStatus scans the outdate execution status from redis and then refresh the status to db,
// do not want to expose to external use so keep it as private.
func scanAndRefreshOutdateStatus(ctx context.Context) {
	iter, err := cache.Default().Scan(ctx, "execution:id:*vendor:*status_outdate")
	if err != nil {
		log.Errorf("failed to scan the outdate executions, error: %v", err)
		return
	}

	var keys []string
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	// return earlier if no keys found which represents no outdate execution
	if len(keys) == 0 {
		log.Debug("skip to refresh, no outdate execution status found")
		return
	}
	// TODO: refactor
	// shuffle the keys to avoid the conflict and improve efficiency when multiple core instance existed,
	// but currently if multiple instances get the same set of keys at the same time, then eventually everyone
	// will still need to repeat the same work(refresh same execution), which needs to be optimized later.
	lib.ShuffleStringSlice(keys)

	log.Infof("scanned out %d executions with outdate status, refresh status to db", len(keys))
	var succeed, failed int64
	// update the execution status execution to db
	for _, key := range keys {
		execID, vendor, err := extractExecIDVendorFromKey(key)
		if err != nil {
			log.Errorf("failed to extract execution id from key %s, error: %v", key, err)
			failed++
			continue
		}

		statusChanged, currentStatus, err := ExecDAO.RefreshStatus(ctx, execID)
		if err != nil {
			// no need to refresh and should clean cache if the execution is not found
			if errors.IsNotFoundErr(err) {
				if err = cache.Default().Delete(ctx, key); err != nil {
					log.Errorf("failed to delete the key %s in cache, error: %v", key, err)
				}
				succeed++
				continue
			}
			log.Errorf("failed to refresh the status of execution %d, error: %v", execID, err)
			failed++
			continue
		}

		succeed++
		log.Debugf("refresh the status of execution %d successfully, new status: %s", execID, currentStatus)
		// run the status change post function
		// just print error log, not return error for post action
		if fc, exist := executionStatusChangePostFuncRegistry[vendor]; exist && statusChanged {
			if err = fc(ctx, execID, currentStatus); err != nil {
				logger.Errorf("failed to run the execution status change post function for execution %d, error: %v", execID, err)
			}
		}
		// delete the key from redis, it does not matter if the deletion fails, wait for the next round.
		if err = cache.Default().Delete(ctx, key); err != nil {
			log.Errorf("failed to delete the key %s in cache, error: %v", key, err)
		}
	}

	log.Infof("refresh outdate execution status done, %d succeed, %d failed", succeed, failed)
}
