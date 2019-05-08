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

package mgt

import (
	"context"
	"fmt"
	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

// Manager defies the related operations to handle the management of job stats.
type Manager interface {
	// Get the stats data of all kinds of jobs.
	// Data returned by pagination.
	//
	// Arguments:
	//   q *query.Parameter : the query parameters
	//
	// Returns:
	//   The matched job stats list
	//   The total number of the jobs
	//   Non nil error if any issues meet
	GetJobs(q *query.Parameter) ([]*job.Stats, int64, error)

	// Get the executions of the specified periodic job by pagination
	//
	// Arguments:
	//   pID: ID of the periodic job
	//   q *query.Parameter: query parameters
	//
	// Returns:
	//   The matched job stats list,
	//   The total number of the executions,
	//   Non nil error if any issues meet.
	GetPeriodicExecution(pID string, q *query.Parameter) ([]*job.Stats, int64, error)

	// Get the scheduled jobs
	//
	// Arguments:
	//   q *query.Parameter: query parameters
	//
	// Returns:
	//   The matched job stats list,
	//   The total number of the executions,
	//   Non nil error if any issues meet.
	GetScheduledJobs(q *query.Parameter) ([]*job.Stats, int64, error)

	// Get the stats of the specified job
	//
	// Arguments:
	//   jobID string: ID of the job
	//
	// Returns:
	//   The job stats
	//   Non nil error if any issues meet
	GetJob(jobID string) (*job.Stats, error)

	// Save the job stats
	//
	// Arguments:
	//   job *job.Stats: the saving job stats
	//
	// Returns:
	//   Non nil error if any issues meet
	SaveJob(job *job.Stats) error
}

// basicManager is the default implementation of @manager,
// based on redis.
type basicManager struct {
	// system context
	ctx context.Context
	// db namespace
	namespace string
	// redis conn pool
	pool *redis.Pool
	// go work client
	client *work.Client
}

// NewManager news a basic manager
func NewManager(ctx context.Context, ns string, pool *redis.Pool) Manager {
	return &basicManager{
		ctx:       ctx,
		namespace: ns,
		pool:      pool,
		client:    work.NewClient(ns, pool),
	}
}

// GetJobs is implementation of Manager.GetJobs
// Because of the hash set used to keep the job stats, we can not support a standard pagination.
// A cursor is used to fetch the jobs with several batches.
func (bm *basicManager) GetJobs(q *query.Parameter) ([]*job.Stats, int64, error) {
	cursor, count := int64(0), query.DefaultPageSize
	if q != nil {
		if q.PageSize > 0 {
			count = q.PageSize
		}

		if cur, ok := q.Extras.Get(query.ExtraParamKeyCursor); ok {
			cursor = cur.(int64)
		}
	}

	pattern := rds.KeyJobStats(bm.namespace, "*")
	args := []interface{}{cursor, "MATCH", pattern, "COUNT", count}

	conn := bm.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	values, err := redis.Values(conn.Do("SCAN", args...))
	if err != nil {
		return nil, 0, err
	}
	if len(values) != 2 {
		return nil, 0, errors.New("malform scan results")
	}

	nextCur, err := strconv.ParseUint(string(values[0].([]byte)), 10, 8)
	if err != nil {
		return nil, 0, err
	}
	list := values[1].([]interface{})

	results := make([]*job.Stats, 0)
	for _, v := range list {
		if bytes, ok := v.([]byte); ok {
			statsKey := string(bytes)
			if i := strings.LastIndex(statsKey, ":"); i != -1 {
				jID := statsKey[i+1:]
				t := job.NewBasicTrackerWithID(bm.ctx, jID, bm.namespace, bm.pool, nil)
				if err := t.Load(); err != nil {
					logger.Errorf("retrieve stats data of job %s error: %s", jID, err)
					continue
				}

				results = append(results, t.Job())
			}
		}
	}

	return results, int64(nextCur), nil
}

// GetPeriodicExecution is implementation of Manager.GetPeriodicExecution
func (bm *basicManager) GetPeriodicExecution(pID string, q *query.Parameter) (results []*job.Stats, total int64, err error) {
	if utils.IsEmptyStr(pID) {
		return nil, 0, errors.New("nil periodic job ID")
	}

	tracker := job.NewBasicTrackerWithID(bm.ctx, pID, bm.namespace, bm.pool, nil)
	err = tracker.Load()
	if err != nil {
		return nil, 0, err
	}

	if tracker.Job().Info.JobKind != job.KindPeriodic {
		return nil, 0, errors.Errorf("only periodic job has executions: %s kind is received", tracker.Job().Info.JobKind)
	}

	conn := bm.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyUpstreamJobAndExecutions(bm.namespace, pID)

	executionIDs := make([]string, 0)
	// Query executions by "non stopped"
	if nonStoppedOnly, ok := q.Extras.Get(query.ExtraParamKeyNonStoppedOnly); ok {
		if v, yes := nonStoppedOnly.(bool); yes && v {
			executionIDs, total, err = queryExecutions(conn, key, q)
			if err != nil {
				return nil, 0, err
			}
		}
	} else {
		// Query all
		// Pagination
		var pageNumber, pageSize uint = 1, query.DefaultPageSize
		if q != nil {
			if q.PageNumber > 0 {
				pageNumber = q.PageNumber
			}
			if q.PageSize > 0 {
				pageSize = q.PageSize
			}
		}

		// Get total first
		total, err = redis.Int64(conn.Do("ZCARD", key))
		if err != nil {
			return nil, 0, err
		}

		// No items
		if total == 0 || (int64)((pageNumber-1)*pageSize) >= total {
			return results, total, nil
		}

		min, max := (pageNumber-1)*pageSize, pageNumber*pageSize-1
		args := []interface{}{key, min, max}
		list, err := redis.Values(conn.Do("ZREVRANGE", args...))
		if err != nil {
			return nil, 0, err
		}

		for _, item := range list {
			if eID, ok := item.([]byte); ok {
				executionIDs = append(executionIDs, string(eID))
			}
		}
	}

	for _, eID := range executionIDs {
		t := job.NewBasicTrackerWithID(bm.ctx, eID, bm.namespace, bm.pool, nil)
		if er := t.Load(); er != nil {
			logger.Errorf("track job %s error: %s", eID, err)
			continue
		}

		results = append(results, t.Job())
	}

	return
}

// GetScheduledJobs is implementation of Manager.GetScheduledJobs
func (bm *basicManager) GetScheduledJobs(q *query.Parameter) ([]*job.Stats, int64, error) {
	// PageSize is not supported here
	var page uint = 1
	if q != nil && q.PageNumber > 1 {
		page = q.PageNumber
	}

	sJobs, total, err := bm.client.ScheduledJobs(page)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*job.Stats, 0)
	for _, sJob := range sJobs {
		jID := sJob.ID
		if len(sJob.Args) > 0 {
			if _, ok := sJob.Args[period.PeriodicExecutionMark]; ok {
				// Periodic scheduled job
				jID = fmt.Sprintf("%s@%d", sJob.ID, sJob.RunAt)
			}
		}
		t := job.NewBasicTrackerWithID(bm.ctx, jID, bm.namespace, bm.pool, nil)
		err = t.Load()
		if err != nil {
			// Just log it
			logger.Errorf("mgt.basicManager: query scheduled jobs error: %s", err)
			continue
		}

		res = append(res, t.Job())
	}

	return res, total, nil
}

// GetJob is implementation of Manager.GetJob
func (bm *basicManager) GetJob(jobID string) (*job.Stats, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError("empty job ID")
	}

	t := job.NewBasicTrackerWithID(bm.ctx, jobID, bm.namespace, bm.pool, nil)
	if err := t.Load(); err != nil {
		return nil, err
	}

	return t.Job(), nil
}

// SaveJob is implementation of Manager.SaveJob
func (bm *basicManager) SaveJob(j *job.Stats) error {
	if j == nil {
		return errs.BadRequestError("nil saving job stats")
	}

	t := job.NewBasicTrackerWithStats(bm.ctx, j, bm.namespace, bm.pool, nil)
	return t.Save()
}

// queryExecutions queries periodic executions by status
func queryExecutions(conn redis.Conn, dataKey string, q *query.Parameter) ([]string, int64, error) {
	total, err := redis.Int64(conn.Do("ZCOUNT", dataKey, 0, "+inf"))
	if err != nil {
		return nil, 0, err
	}

	var pageNumber, pageSize uint = 1, query.DefaultPageSize
	if q.PageNumber > 0 {
		pageNumber = q.PageNumber
	}
	if q.PageSize > 0 {
		pageSize = q.PageSize
	}

	results := make([]string, 0)
	if total == 0 || (int64)((pageNumber-1)*pageSize) >= total {
		return results, total, nil
	}

	offset := (pageNumber - 1) * pageSize
	args := []interface{}{dataKey, "+inf", 0, "LIMIT", offset, pageSize}

	eIDs, err := redis.Values(conn.Do("ZREVRANGEBYSCORE", args...))
	if err != nil {
		return nil, 0, err
	}

	for _, eID := range eIDs {
		if eIDBytes, ok := eID.([]byte); ok {
			results = append(results, string(eIDBytes))
		}
	}

	return results, total, nil
}
