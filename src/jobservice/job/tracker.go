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

package job

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"
	"time"
)

const (
	// Try best to keep the job stats data but anyway clear it after a long time
	statDataExpireTime = 180 * 24 * 3600
	// Default page size for querying
	defaultPageSize = 25
)

// Tracker is designed to track the life cycle of the job described by the stats
// The status change is linear and then has strict preorder and successor
// Check should be enforced before switching
//
// Pending is default status when creating job, so no need to switch
type Tracker interface {
	// Save the job stats which tracked by this tracker to the backend
	//
	// Return:
	//   none nil error returned if any issues happened
	Save() error

	// Load the job stats which tracked by this tracker with the backend data
	//
	// Return:
	//   none nil error returned if any issues happened
	Load() error

	// Get the job stats which tracked by this tracker
	//
	// Returns:
	//  *models.Info : job stats data
	Job() *Stats

	// Update the properties of the job stats
	//
	// fieldAndValues ...interface{} : One or more properties being updated
	//
	// Returns:
	//  error if update failed
	Update(fieldAndValues ...interface{}) error

	// Executions returns the executions of the job tracked by this tracker.
	// Please pay attention, this only for periodic job.
	//
	// Returns:
	//   job execution IDs matched the query
	//   the total number
	//   error if any issues happened
	Executions(q *query.Parameter) ([]string, int64, error)

	// NumericID returns the numeric ID of periodic job.
	// Please pay attention, this only for periodic job.
	NumericID() (int64, error)

	// Mark the periodic job execution to done by update the score
	// of the relation between its periodic policy and execution to -1.
	PeriodicExecutionDone() error

	// Check in message
	CheckIn(message string) error

	// The current status of job
	Status() (Status, error)

	// Expire the job stats data
	Expire() error

	// Switch status to running
	Run() error

	// Switch status to scheduled
	Schedule() error

	// Switch status to stopped
	Stop() error

	// Switch the status to error
	Fail() error

	// Switch the status to success
	Succeed() error
}

// basicTracker implements Tracker interface based on redis
type basicTracker struct {
	namespace string
	context   context.Context
	pool      *redis.Pool
	jobID     string
	jobStats  *Stats
	callback  HookCallback
}

// NewBasicTrackerWithID builds a tracker with the provided job ID
func NewBasicTrackerWithID(
	jobID string,
	ctx context.Context,
	ns string,
	pool *redis.Pool,
	callback HookCallback,
) Tracker {
	return &basicTracker{
		namespace: ns,
		context:   ctx,
		pool:      pool,
		jobID:     jobID,
		callback:  callback,
	}
}

// NewBasicTrackerWithStats builds a tracker with the provided job stats
func NewBasicTrackerWithStats(
	stats *Stats,
	ctx context.Context,
	ns string,
	pool *redis.Pool,
	callback HookCallback,
) Tracker {
	return &basicTracker{
		namespace: ns,
		context:   ctx,
		pool:      pool,
		jobStats:  stats,
		jobID:     stats.Info.JobID,
		callback:  callback,
	}
}

// Refresh the job stats which tracked by this tracker
func (bt *basicTracker) Load() error {
	return bt.retrieve()
}

// Job returns the job stats which tracked by this tracker
func (bt *basicTracker) Job() *Stats {
	return bt.jobStats
}

// Update the properties of the job stats
func (bt *basicTracker) Update(fieldAndValues ...interface{}) error {
	if len(fieldAndValues) == 0 {
		errors.New("no properties specified to update")
	}

	conn := bt.pool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(bt.namespace, bt.jobID)
	args := []interface{}{"update_time", time.Now().Unix()} // update timestamp
	args = append(args, fieldAndValues...)

	return rds.HmSet(conn, key, args...)
}

// Status returns the current status of job tracked by this tracker
func (bt *basicTracker) Status() (Status, error) {
	// Retrieve the latest status again in case get the outdated one.
	conn := bt.pool.Get()
	defer conn.Close()

	rootKey := rds.KeyJobStats(bt.namespace, bt.jobID)
	return getStatus(conn, rootKey)
}

// NumericID returns the numeric ID of the periodic job
func (bt *basicTracker) NumericID() (int64, error) {
	if bt.jobStats.Info.NumericPID > 0 {
		return bt.jobStats.Info.NumericPID, nil
	}

	return -1, errors.Errorf("numeric ID not found for job: %s", bt.jobID)
}

// PeriodicExecutionDone mark the execution done
func (bt *basicTracker) PeriodicExecutionDone() error {
	if utils.IsEmptyStr(bt.jobStats.Info.UpstreamJobID) {
		return errors.Errorf("%s is not periodic job execution", bt.jobID)
	}

	key := rds.KeyUpstreamJobAndExecutions(bt.namespace, bt.jobStats.Info.UpstreamJobID)

	conn := bt.pool.Get()
	defer conn.Close()

	args := []interface{}{key, "XX", -1, bt.jobID}
	_, err := conn.Do("ZADD", args...)

	return err
}

// Check in message
func (bt *basicTracker) CheckIn(message string) error {
	if utils.IsEmptyStr(message) {
		return errors.New("check in error: empty message")
	}

	err := bt.fireHook(Status(bt.jobStats.Info.Status), message)
	err = bt.Update(
		"check_in", message,
		"check_in_at", time.Now().Unix(),
		"update_time", time.Now().Unix(),
	)

	return err
}

// Executions of the tracked job
func (bt *basicTracker) Executions(q *query.Parameter) ([]string, int64, error) {
	if bt.jobStats.Info.JobKind != KindPeriodic {
		return nil, 0, errors.New("only periodic job has executions")
	}

	conn := bt.pool.Get()
	defer conn.Close()

	key := rds.KeyUpstreamJobAndExecutions(bt.namespace, bt.jobID)

	// Pagination
	var pageNumber, pageSize uint = 1, defaultPageSize
	if q != nil {
		if q.PageNumber > 0 {
			pageNumber = q.PageNumber
		}
		if q.PageSize > 0 {
			pageSize = q.PageSize
		}
	}

	// Get total first
	total, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		return nil, 0, err
	}

	// No items
	result := make([]string, 0)
	if total == 0 || (int64)((pageNumber-1)*pageSize) >= total {
		return result, total, nil
	}

	min, max := (pageNumber-1)*pageSize, pageNumber*pageSize-1
	args := []interface{}{key, min, max}
	list, err := redis.Values(conn.Do("ZREVRANGE", args...))
	if err != nil {
		return nil, 0, err
	}

	for _, item := range list {
		if eID, ok := item.(string); ok {
			result = append(result, eID)
		}
	}

	return result, total, nil
}

// Expire job stats
func (bt *basicTracker) Expire() error {
	conn := bt.pool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(bt.namespace, bt.jobID)
	num, err := conn.Do("EXPIRE", key, statDataExpireTime)
	if err != nil {
		return err
	}

	if num == 0 {
		return errors.Errorf("job stats for expiring %s does not exist", bt.jobID)
	}

	return nil
}

// Run job
func (bt *basicTracker) Run() error {
	return bt.compareAndSet(RunningStatus)
}

// Schedule job
func (bt *basicTracker) Schedule() error {
	return bt.compareAndSet(ScheduledStatus)
}

// Stop job
// Stop is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Stop() error {
	err := bt.fireHook(StoppedStatus)
	err = bt.updateStatusWithRetry(StoppedStatus)

	return err
}

// Fail job
// Fail is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Fail() error {
	err := bt.fireHook(ErrorStatus)
	err = bt.updateStatusWithRetry(ErrorStatus)

	return err
}

// Succeed job
// Succeed is final status, if failed to do, retry should be enforced.
// Either one is failed, the final return will be marked as failed.
func (bt *basicTracker) Succeed() error {
	err := bt.fireHook(SuccessStatus)
	err = bt.updateStatusWithRetry(SuccessStatus)

	return err
}

// Save the stats of job tracked by this tracker
func (bt *basicTracker) Save() (err error) {
	if bt.jobStats == nil {
		errors.New("nil job stats to save")
	}

	conn := bt.pool.Get()
	defer conn.Close()

	// Alliance
	stats := bt.jobStats

	key := rds.KeyJobStats(bt.namespace, stats.Info.JobID)
	args := make([]interface{}, 0)
	args = append(args, key)
	args = append(args,
		"id", stats.Info.JobID,
		"name", stats.Info.JobName,
		"kind", stats.Info.JobKind,
		"unique", stats.Info.IsUnique,
		"status", stats.Info.Status,
		"ref_link", stats.Info.RefLink,
		"enqueue_time", stats.Info.EnqueueTime,
		"run_at", stats.Info.RunAt,
		"cron_spec", stats.Info.CronSpec,
		"web_hook_url", stats.Info.WebHookURL,
		"numeric_policy_id", stats.Info.NumericPID,
	)
	if stats.Info.CheckInAt > 0 && !utils.IsEmptyStr(stats.Info.CheckIn) {
		args = append(args,
			"check_in", stats.Info.CheckIn,
			"check_in_at", stats.Info.CheckInAt,
		)
	}
	if stats.Info.DieAt > 0 {
		args = append(args, "die_at", stats.Info.DieAt)
	}

	if !utils.IsEmptyStr(stats.Info.UpstreamJobID) {
		args = append(args, "upstream_job_id", stats.Info.UpstreamJobID)
	}
	// Set update timestamp
	args = append(args, "update_time", time.Now().Unix())

	// Do it in a transaction
	err = conn.Send("MULTI")
	err = conn.Send("HMSET", args...)

	// If job kind is periodic job, expire time should not be set
	// If job kind is scheduled job, expire time should be runAt+
	if stats.Info.JobKind != KindPeriodic {
		var expireTime int64 = statDataExpireTime
		if stats.Info.JobKind == KindScheduled {
			nowTime := time.Now().Unix()
			future := stats.Info.RunAt - nowTime
			if future > 0 {
				expireTime += future
			}
		}
		expireTime += rand.Int63n(15) // Avoid lots of keys being expired at the same time
		err = conn.Send("EXPIRE", key, expireTime)
	}

	// Link with its upstream job if upstream job ID exists for future querying
	if !utils.IsEmptyStr(stats.Info.UpstreamJobID) {
		k := rds.KeyUpstreamJobAndExecutions(bt.namespace, stats.Info.UpstreamJobID)
		zargs := []interface{}{k, "NX", stats.Info.RunAt, stats.Info.JobID}
		err = conn.Send("ZADD", zargs...)
	}

	// Check command send error only once here before executing
	if err != nil {
		return
	}

	_, err = conn.Do("EXEC")

	return
}

// Fire the hook event
func (bt *basicTracker) fireHook(status Status, checkIn ...string) error {
	// Check if hook URL is registered
	if utils.IsEmptyStr(bt.jobStats.Info.WebHookURL) {
		// Do nothing
		return nil
	}

	change := &StatusChange{
		JobID:    bt.jobID,
		Status:   status.String(),
		Metadata: bt.jobStats.Info,
	}

	if len(checkIn) > 0 {
		change.CheckIn = checkIn[0]
	}

	// If callback is registered, then trigger now
	if bt.callback != nil {
		return bt.callback(bt.jobStats.Info.WebHookURL, change)
	}

	return nil
}

// If update status failed, then retry if permitted.
// Try best to do
func (bt *basicTracker) updateStatusWithRetry(targetStatus Status) error {
	err := bt.compareAndSet(targetStatus)
	if err != nil {
		// If still need to retry
		// Check the update timestamp
		if time.Now().Unix()-bt.jobStats.Info.UpdateTime < 2*24*3600 {
			// Keep on retrying
			go func() {
				select {
				case <-time.After(time.Duration(5)*time.Minute + time.Duration(rand.Int31n(13))*time.Second):
					if err := bt.updateStatusWithRetry(targetStatus); err != nil {
						logger.Errorf("Retry of updating status of job %s error: %s", bt.jobID, err)
					}
				case <-bt.context.Done():
					return // terminated
				}
			}()
		}
	}

	return err
}

func (bt *basicTracker) compareAndSet(targetStatus Status) error {
	conn := bt.pool.Get()
	defer conn.Close()

	rootKey := rds.KeyJobStats(bt.namespace, bt.jobID)

	st, err := getStatus(conn, rootKey)
	if err != nil {
		return err
	}

	if st.Compare(targetStatus) >= 0 {
		return fmt.Errorf("mismatch job status: current %s, setting to %s", st, targetStatus)
	}

	return setStatus(conn, rootKey, targetStatus)
}

// retrieve the stats of job tracked by this tracker from the backend data
func (bt *basicTracker) retrieve() error {
	conn := bt.pool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(bt.namespace, bt.jobID)
	vals, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return err
	}

	if vals == nil || len(vals) == 0 {
		return errs.NoObjectFoundError(bt.jobID)
	}

	res := &Stats{
		Info: &StatsInfo{},
	}

	for i, l := 0, len(vals); i < l; i = i + 2 {
		prop := vals[i]
		value := vals[i+1]
		switch prop {
		case "id":
			res.Info.JobID = value
			break
		case "name":
			res.Info.JobName = value
			break
		case "kind":
			res.Info.JobKind = value
		case "unique":
			v, err := strconv.ParseBool(value)
			if err != nil {
				v = false
			}
			res.Info.IsUnique = v
		case "status":
			res.Info.Status = value
			break
		case "ref_link":
			res.Info.RefLink = value
			break
		case "enqueue_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.EnqueueTime = v
			break
		case "update_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.UpdateTime = v
			break
		case "run_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.RunAt = v
			break
		case "check_in_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.CheckInAt = v
			break
		case "check_in":
			res.Info.CheckIn = value
			break
		case "cron_spec":
			res.Info.CronSpec = value
			break
		case "web_hook_url":
			res.Info.WebHookURL = value
			break
		case "die_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.DieAt = v
		case "upstream_job_id":
			res.Info.UpstreamJobID = value
			break
		case "numeric_policy_id":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Info.NumericPID = v
			break
		default:
			break
		}
	}

	bt.jobStats = res

	return nil
}

func getStatus(conn redis.Conn, key string) (Status, error) {
	values, err := rds.HmGet(conn, key, "status")
	if err != nil {
		return "", err
	}

	if len(values) == 1 {
		st := Status(values[0].([]byte))
		if st.Validate() == nil {
			return st, nil
		}
	}

	return "", errors.New("malformed status data returned")
}

func setStatus(conn redis.Conn, key string, status Status) error {
	return rds.HmSet(conn, key, "status", status.String(), "update_time", time.Now().Unix())
}
