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

package opm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/logger"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

const (
	processBufferSize      = 1024
	opSaveStats            = "save_job_stats"
	opUpdateStatus         = "update_job_status"
	opCheckIn              = "check_in"
	opDieAt                = "mark_die_at"
	opReportStatus         = "report_status"
	opPersistExecutions    = "persist_executions"
	opUpdateStats          = "update_job_stats"
	maxFails               = 3
	jobStatsDataExpireTime = 60 * 60 * 24 * 7 // one week

	// CtlCommandStop : command stop
	CtlCommandStop = "stop"
	// CtlCommandCancel : command cancel
	CtlCommandCancel = "cancel"
	// CtlCommandRetry : command retry
	CtlCommandRetry = "retry"

	// EventRegisterStatusHook is event name of registering hook
	EventRegisterStatusHook = "register_hook"
)

type queueItem struct {
	Op    string
	Fails uint
	Data  interface{}
}

func (qi *queueItem) string() string {
	data, err := json.Marshal(qi)
	if err != nil {
		return fmt.Sprintf("%v", qi)
	}

	return string(data)
}

// RedisJobStatsManager implements JobStatsManager based on redis.
type RedisJobStatsManager struct {
	namespace   string
	redisPool   *redis.Pool
	context     context.Context
	stopChan    chan struct{}
	doneChan    chan struct{}
	processChan chan *queueItem
	isRunning   *atomic.Value
	hookStore   *HookStore  // cache the hook here to avoid requesting backend
	opCommands  *oPCommands // maintain the OP commands
}

// NewRedisJobStatsManager is constructor of RedisJobStatsManager
func NewRedisJobStatsManager(ctx context.Context, namespace string, redisPool *redis.Pool) JobStatsManager {
	isRunning := &atomic.Value{}
	isRunning.Store(false)

	return &RedisJobStatsManager{
		namespace:   namespace,
		context:     ctx,
		redisPool:   redisPool,
		stopChan:    make(chan struct{}, 1),
		doneChan:    make(chan struct{}, 1),
		processChan: make(chan *queueItem, processBufferSize),
		hookStore:   NewHookStore(),
		isRunning:   isRunning,
		opCommands:  newOPCommands(ctx, namespace, redisPool),
	}
}

// Start is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Start() {
	if rjs.isRunning.Load().(bool) {
		return
	}
	go rjs.loop()
	rjs.opCommands.Start()
	rjs.isRunning.Store(true)

	logger.Info("Redis job stats manager is started")
}

// Shutdown is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Shutdown() {
	defer func() {
		rjs.isRunning.Store(false)
	}()

	if !(rjs.isRunning.Load().(bool)) {
		return
	}

	rjs.opCommands.Stop()
	rjs.stopChan <- struct{}{}
	<-rjs.doneChan
}

// Save is implementation of same method in JobStatsManager interface.
// Async method
func (rjs *RedisJobStatsManager) Save(jobStats models.JobStats) {
	item := &queueItem{
		Op:   opSaveStats,
		Data: jobStats,
	}

	rjs.processChan <- item
}

// Retrieve is implementation of same method in JobStatsManager interface.
// Sync method
func (rjs *RedisJobStatsManager) Retrieve(jobID string) (models.JobStats, error) {
	if utils.IsEmptyStr(jobID) {
		return models.JobStats{}, errors.New("empty job ID")
	}

	res, err := rjs.getJobStats(jobID)
	if err != nil {
		return models.JobStats{}, err
	}

	if res.Stats.IsMultipleExecutions {
		executions, err := rjs.GetExecutions(jobID)
		if err != nil {
			return models.JobStats{}, err
		}

		res.Stats.Executions = executions
	}

	return res, nil
}

// SetJobStatus is implementation of same method in JobStatsManager interface.
// Async method
func (rjs *RedisJobStatsManager) SetJobStatus(jobID string, status string) {
	if utils.IsEmptyStr(jobID) || utils.IsEmptyStr(status) {
		return
	}

	item := &queueItem{
		Op:   opUpdateStatus,
		Data: []string{jobID, status},
	}

	rjs.processChan <- item

	// Report status at the same time
	rjs.submitStatusReportingItem(jobID, status, "")
}

func (rjs *RedisJobStatsManager) loop() {
	controlChan := make(chan struct{})

	defer func() {
		rjs.isRunning.Store(false)
		// Notify other sub goroutines
		close(controlChan)
		logger.Info("Redis job stats manager is stopped")
	}()

	for {
		select {
		case item := <-rjs.processChan:
			go func(item *queueItem) {
				clearHookCache := false
				if err := rjs.process(item); err != nil {
					item.Fails++
					if item.Fails < maxFails {
						logger.Warningf("Failed to process '%s' request with error: %s\n", item.Op, err)

						// Retry after a random interval
						go func() {
							timer := time.NewTimer(time.Duration(backoff(item.Fails)) * time.Second)
							defer timer.Stop()

							select {
							case <-timer.C:
								rjs.processChan <- item
								return
							case <-controlChan:
							}
						}()
					} else {
						logger.Errorf("Failed to process '%s' request with error: %s (%d times tried)\n", item.Op, err, maxFails)
						if item.Op == opReportStatus {
							clearHookCache = true
						}
					}
				} else {
					logger.Debugf("Operation is successfully processed: %s", item.string())

					if item.Op == opReportStatus {
						clearHookCache = true
					}
				}

				if clearHookCache {
					// Clear cache to save memory if job status is success or stopped.
					data := item.Data.([]string)
					status := data[2]
					if status == job.JobStatusSuccess || status == job.JobStatusStopped {
						rjs.hookStore.Remove(data[0])
					}
				}
			}(item)
			break
		case <-rjs.stopChan:
			rjs.doneChan <- struct{}{}
			return
		case <-rjs.context.Done():
			return
		}
	}
}

// SendCommand for the specified job
func (rjs *RedisJobStatsManager) SendCommand(jobID string, command string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if command != CtlCommandStop && command != CtlCommandCancel {
		return errors.New("unknown command")
	}

	if err := rjs.opCommands.Fire(jobID, command); err != nil {
		return err
	}

	// Directly add to op commands maintaining list
	return rjs.opCommands.Push(jobID, command)
}

// CheckIn mesage
func (rjs *RedisJobStatsManager) CheckIn(jobID string, message string) {
	if utils.IsEmptyStr(jobID) || utils.IsEmptyStr(message) {
		return
	}

	item := &queueItem{
		Op:   opCheckIn,
		Data: []string{jobID, message},
	}

	rjs.processChan <- item

	// Report checkin message at the same time
	rjs.submitStatusReportingItem(jobID, job.JobStatusRunning, message)
}

// CtlCommand checks if control command is fired for the specified job.
func (rjs *RedisJobStatsManager) CtlCommand(jobID string) (string, error) {
	if utils.IsEmptyStr(jobID) {
		return "", errors.New("empty job ID")
	}

	c, ok := rjs.opCommands.Pop(jobID)
	if !ok {
		return "", fmt.Errorf("no OP command fired to job %s", jobID)
	}

	return c, nil
}

// DieAt marks the failed jobs with the time they put into dead queue.
func (rjs *RedisJobStatsManager) DieAt(jobID string, dieAt int64) {
	if utils.IsEmptyStr(jobID) || dieAt == 0 {
		return
	}

	item := &queueItem{
		Op:   opDieAt,
		Data: []interface{}{jobID, dieAt},
	}

	rjs.processChan <- item
}

// RegisterHook is used to save the hook url or cache the url in memory.
func (rjs *RedisJobStatsManager) RegisterHook(jobID string, hookURL string, isCached bool) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if !utils.IsValidURL(hookURL) {
		return errors.New("invalid hook url")
	}

	if !isCached {
		return rjs.saveHook(jobID, hookURL)
	}

	rjs.hookStore.Add(jobID, hookURL)

	return nil
}

// GetHook returns the status web hook url for the specified job if existing
func (rjs *RedisJobStatsManager) GetHook(jobID string) (string, error) {
	if utils.IsEmptyStr(jobID) {
		return "", errors.New("empty job ID")
	}

	// First retrieve from the cache
	if hookURL, ok := rjs.hookStore.Get(jobID); ok {
		return hookURL, nil
	}

	return rjs.getHook(jobID)
}

// ExpirePeriodicJobStats marks the periodic job stats expired
func (rjs *RedisJobStatsManager) ExpirePeriodicJobStats(jobID string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	// The periodic job (policy) is stopped/unscheduled and then
	// the stats of periodic job now can be expired
	key := utils.KeyJobStats(rjs.namespace, jobID)
	_, err := conn.Do("EXPIRE", key, jobStatsDataExpireTime)

	return err
}

// AttachExecution persist the links between upstream jobs and the related executions (jobs).
func (rjs *RedisJobStatsManager) AttachExecution(upstreamJobID string, executions ...string) error {
	if len(upstreamJobID) == 0 {
		return errors.New("empty upstream job ID is not allowed")
	}

	if len(executions) == 0 {
		return errors.New("no executions existing to persist")
	}

	// Send to process channel
	item := &queueItem{
		Op:   opPersistExecutions,
		Data: []interface{}{upstreamJobID, executions},
	}

	rjs.processChan <- item

	return nil
}

// GetExecutions returns the existing executions (IDs) for the specified job.
func (rjs *RedisJobStatsManager) GetExecutions(upstreamJobID string) ([]string, error) {
	if len(upstreamJobID) == 0 {
		return nil, errors.New("no upstream ID specified")
	}

	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyUpstreamJobAndExecutions(rjs.namespace, upstreamJobID)
	ids, err := redis.Strings(conn.Do("ZRANGE", key, 0, -1))
	if err != nil {
		if err == redis.ErrNil {
			return []string{}, nil
		}

		return nil, err
	}

	return ids, nil
}

// Update the properties of job stats
func (rjs *RedisJobStatsManager) Update(jobID string, fieldAndValues ...interface{}) error {
	if len(jobID) == 0 {
		return errors.New("no updating job")
	}

	if len(fieldAndValues) == 0 || len(fieldAndValues)%2 != 0 {
		return errors.New("filed and its value should be pair")
	}

	data := []interface{}{}
	data = append(data, jobID)
	data = append(data, fieldAndValues...)

	item := &queueItem{
		Op:   opUpdateStats,
		Data: data,
	}

	rjs.processChan <- item

	return nil
}

func (rjs *RedisJobStatsManager) submitStatusReportingItem(jobID string, status, checkIn string) {
	// Let it run in a separate goroutine to avoid waiting more time
	go func() {
		var (
			hookURL string
			ok      bool
			err     error
		)

		hookURL, ok = rjs.hookStore.Get(jobID)
		if !ok {
			// Retrieve from backend
			hookURL, err = rjs.getHook(jobID)
			if err != nil || !utils.IsValidURL(hookURL) {
				// logged and exit
				logger.Warningf("no status hook found for job %s\n, abandon status reporting", jobID)
				return
			}
		}

		item := &queueItem{
			Op:   opReportStatus,
			Data: []string{jobID, hookURL, status, checkIn},
		}

		rjs.processChan <- item
	}()
}

func (rjs *RedisJobStatsManager) reportStatus(jobID string, hookURL, status, checkIn string) error {
	reportingStatus := models.JobStatusChange{
		JobID:   jobID,
		Status:  status,
		CheckIn: checkIn,
	}
	// Return the whole metadata of the job.
	// To support forward compatibility, keep the original fields `Status` and `CheckIn`.
	// TODO: If querying job stats causes performance issues, a two-level cache should be enabled.
	jobStats, err := rjs.getJobStats(jobID)
	if err != nil {
		// Just logged
		logger.Errorf("Retrieving stats of job %s for hook reporting failed with error: %s", jobID, err)
	} else {
		// Override status/check in message
		// Just double confirmation
		jobStats.Stats.CheckIn = checkIn
		jobStats.Stats.Status = status
		reportingStatus.Metadata = jobStats.Stats
	}

	return DefaultHookClient.ReportStatus(hookURL, reportingStatus)
}

func (rjs *RedisJobStatsManager) updateJobStats(jobID string, fieldAndValues ...interface{}) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	args := make([]interface{}, 0, len(fieldAndValues)+1)

	args = append(args, key)
	args = append(args, fieldAndValues...)
	args = append(args, "update_time", time.Now().Unix())

	_, err := conn.Do("HMSET", args...)

	return err
}

func (rjs *RedisJobStatsManager) updateJobStatus(jobID string, status string) error {
	args := make([]interface{}, 0, 4)
	args = append(args, "status", status)
	if status == job.JobStatusSuccess {
		// make sure the 'die_at' is reset in case it's a retrying job
		args = append(args, "die_at", 0)
	}

	return rjs.updateJobStats(jobID, args...)
}

func (rjs *RedisJobStatsManager) checkIn(jobID string, message string) error {

	now := time.Now().Unix()
	args := make([]interface{}, 0, 4)
	args = append(args, "check_in", message, "check_in_at", now)

	return rjs.updateJobStats(jobID, args...)
}

func (rjs *RedisJobStatsManager) dieAt(jobID string, baseTime int64) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	// Query the dead job in the time scope of [baseTime,baseTime+5]
	key := utils.RedisKeyDead(rjs.namespace)
	jobWithScores, err := utils.GetZsetByScore(rjs.redisPool, key, []int64{baseTime, baseTime + 5})
	if err != nil {
		return err
	}

	for _, jws := range jobWithScores {
		if j, err := utils.DeSerializeJob(jws.JobBytes); err == nil {
			if j.ID == jobID {
				// Found
				args := make([]interface{}, 0, 6)
				args = append(args, "die_at", jws.Score)
				return rjs.updateJobStats(jobID, args...)
			}
		}
	}

	return fmt.Errorf("seems %s is not a dead job", jobID)
}

func (rjs *RedisJobStatsManager) getJobStats(jobID string) (models.JobStats, error) {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	vals, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return models.JobStats{}, err
	}

	if vals == nil || len(vals) == 0 {
		return models.JobStats{}, errs.NoObjectFoundError(fmt.Sprintf("job '%s'", jobID))
	}

	res := models.JobStats{
		Stats: &models.JobStatData{},
	}
	for i, l := 0, len(vals); i < l; i = i + 2 {
		prop := vals[i]
		value := vals[i+1]
		switch prop {
		case "id":
			res.Stats.JobID = value
			break
		case "name":
			res.Stats.JobName = value
			break
		case "kind":
			res.Stats.JobKind = value
		case "unique":
			v, err := strconv.ParseBool(value)
			if err != nil {
				v = false
			}
			res.Stats.IsUnique = v
		case "status":
			res.Stats.Status = value
			break
		case "ref_link":
			res.Stats.RefLink = value
			break
		case "enqueue_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.EnqueueTime = v
			break
		case "update_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.UpdateTime = v
			break
		case "run_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.RunAt = v
			break
		case "check_in_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.CheckInAt = v
			break
		case "check_in":
			res.Stats.CheckIn = value
			break
		case "cron_spec":
			res.Stats.CronSpec = value
			break
		case "die_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.DieAt = v
		case "upstream_job_id":
			res.Stats.UpstreamJobID = value
			break
		case "multiple_executions":
			v, err := strconv.ParseBool(value)
			if err != nil {
				v = false
			}
			res.Stats.IsMultipleExecutions = v
			break
		default:
			break
		}
	}

	return res, nil
}

func (rjs *RedisJobStatsManager) saveJobStats(jobStats models.JobStats) error {
	if jobStats.Stats == nil {
		return errors.New("malformed job stats object")
	}

	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobStats.Stats.JobID)
	args := make([]interface{}, 0)
	args = append(args, key)
	args = append(args,
		"id", jobStats.Stats.JobID,
		"name", jobStats.Stats.JobName,
		"kind", jobStats.Stats.JobKind,
		"unique", jobStats.Stats.IsUnique,
		"status", jobStats.Stats.Status,
		"ref_link", jobStats.Stats.RefLink,
		"enqueue_time", jobStats.Stats.EnqueueTime,
		"update_time", jobStats.Stats.UpdateTime,
		"run_at", jobStats.Stats.RunAt,
		"cron_spec", jobStats.Stats.CronSpec,
		"multiple_executions", jobStats.Stats.IsMultipleExecutions,
	)
	if jobStats.Stats.CheckInAt > 0 && !utils.IsEmptyStr(jobStats.Stats.CheckIn) {
		args = append(args,
			"check_in", jobStats.Stats.CheckIn,
			"check_in_at", jobStats.Stats.CheckInAt,
		)
	}
	if jobStats.Stats.DieAt > 0 {
		args = append(args, "die_at", jobStats.Stats.DieAt)
	}

	if len(jobStats.Stats.UpstreamJobID) > 0 {
		args = append(args, "upstream_job_id", jobStats.Stats.UpstreamJobID)
	}

	conn.Send("HMSET", args...)
	// If job kind is periodic job, expire time should not be set
	// If job kind is scheduled job, expire time should be runAt+1day
	if jobStats.Stats.JobKind != job.JobKindPeriodic {
		var expireTime int64 = jobStatsDataExpireTime
		if jobStats.Stats.JobKind == job.JobKindScheduled {
			nowTime := time.Now().Unix()
			future := jobStats.Stats.RunAt - nowTime
			if future > 0 {
				expireTime += future
			}
		}
		expireTime += rand.Int63n(30) // Avoid lots of keys being expired at the same time
		conn.Send("EXPIRE", key, expireTime)
	}

	return conn.Flush()
}

func (rjs *RedisJobStatsManager) saveExecutions(upstreamJobID string, executions []string) error {
	key := utils.KeyUpstreamJobAndExecutions(rjs.namespace, upstreamJobID)

	conn := rjs.redisPool.Get()
	defer conn.Close()

	err := conn.Send("MULTI")
	if err != nil {
		return err
	}

	args := []interface{}{key}
	baseScore := time.Now().Unix()
	for index, execution := range executions {
		args = append(args, baseScore+int64(index), execution)
	}

	if err := conn.Send("ZADD", args...); err != nil {
		return err
	}

	// add expire time
	if err := conn.Send("EXPIRE", key, jobStatsDataExpireTime); err != nil {
		return err
	}

	_, err = conn.Do("EXEC")

	return err
}

func (rjs *RedisJobStatsManager) process(item *queueItem) error {
	switch item.Op {
	case opSaveStats:
		jobStats := item.Data.(models.JobStats)
		return rjs.saveJobStats(jobStats)
	case opUpdateStatus:
		data := item.Data.([]string)
		return rjs.updateJobStatus(data[0], data[1])
	case opCheckIn:
		data := item.Data.([]string)
		return rjs.checkIn(data[0], data[1])
	case opDieAt:
		data := item.Data.([]interface{})
		return rjs.dieAt(data[0].(string), data[1].(int64))
	case opReportStatus:
		data := item.Data.([]string)
		return rjs.reportStatus(data[0], data[1], data[2], data[3])
	case opPersistExecutions:
		data := item.Data.([]interface{})
		return rjs.saveExecutions(data[0].(string), data[1].([]string))
	case opUpdateStats:
		data := item.Data.([]interface{})
		return rjs.updateJobStats(data[0].(string), data[1:]...)
	default:
		break
	}

	return nil
}

// HookData keeps the hook url info
type HookData struct {
	JobID   string `json:"job_id"`
	HookURL string `json:"hook_url"`
}

func (rjs *RedisJobStatsManager) saveHook(jobID string, hookURL string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	args := make([]interface{}, 0, 3)
	args = append(args, key, "status_hook", hookURL)
	msg := &models.Message{
		Event: EventRegisterStatusHook,
		Data: &HookData{
			JobID:   jobID,
			HookURL: hookURL,
		},
	}
	rawJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// hook is saved into the job stats
	// We'll not set expire time here, the expire time of the key will be set when saving job stats
	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("HMSET", args...); err != nil {
		return err
	}
	if err := conn.Send("PUBLISH", utils.KeyPeriodicNotification(rjs.namespace), rawJSON); err != nil {
		return err
	}

	_, err = conn.Do("EXEC")
	return err
}

func (rjs *RedisJobStatsManager) getHook(jobID string) (string, error) {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	hookURL, err := redis.String(conn.Do("HMGET", key, "status_hook"))
	if err != nil {
		if err == redis.ErrNil {
			return "", fmt.Errorf("no registered web hook found for job '%s'", jobID)
		}
		return "", err
	}

	return hookURL, nil
}

func backoff(seed uint) int {
	if seed < 1 {
		seed = 1
	}

	return int(math.Pow(float64(seed+1), float64(seed))) + rand.Intn(5)
}
