// Copyright 2018 The Harbor Authors. All rights reserved.

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

	"github.com/vmware/harbor/src/jobservice/errs"
	"github.com/vmware/harbor/src/jobservice/logger"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/jobservice/job"
	"github.com/vmware/harbor/src/jobservice/models"
	"github.com/vmware/harbor/src/jobservice/utils"
)

const (
	processBufferSize = 1024
	opSaveStats       = "save_job_stats"
	opUpdateStatus    = "update_job_status"
	opCheckIn         = "check_in"
	opDieAt           = "mark_die_at"
	opReportStatus    = "report_status"
	maxFails          = 3

	//CtlCommandStop : command stop
	CtlCommandStop = "stop"
	//CtlCommandCancel : command cancel
	CtlCommandCancel = "cancel"
	//CtlCommandRetry : command retry
	CtlCommandRetry = "retry"

	//EventRegisterStatusHook is event name of registering hook
	EventRegisterStatusHook = "register_hook"
)

type queueItem struct {
	op    string
	fails uint
	data  interface{}
}

//RedisJobStatsManager implements JobStatsManager based on redis.
type RedisJobStatsManager struct {
	namespace   string
	redisPool   *redis.Pool
	context     context.Context
	stopChan    chan struct{}
	doneChan    chan struct{}
	processChan chan *queueItem
	isRunning   *atomic.Value
	hookStore   *HookStore  //cache the hook here to avoid requesting backend
	opCommands  *oPCommands //maintain the OP commands
}

//NewRedisJobStatsManager is constructor of RedisJobStatsManager
func NewRedisJobStatsManager(ctx context.Context, namespace string, redisPool *redis.Pool) *RedisJobStatsManager {
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

//Start is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Start() {
	if rjs.isRunning.Load().(bool) {
		return
	}
	go rjs.loop()
	rjs.opCommands.Start()
	rjs.isRunning.Store(true)

	logger.Info("Redis job stats manager is started")
}

//Shutdown is implementation of same method in JobStatsManager interface.
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

//Save is implementation of same method in JobStatsManager interface.
//Async method
func (rjs *RedisJobStatsManager) Save(jobStats models.JobStats) {
	item := &queueItem{
		op:   opSaveStats,
		data: jobStats,
	}

	rjs.processChan <- item
}

//Retrieve is implementation of same method in JobStatsManager interface.
//Sync method
func (rjs *RedisJobStatsManager) Retrieve(jobID string) (models.JobStats, error) {
	if utils.IsEmptyStr(jobID) {
		return models.JobStats{}, errors.New("empty job ID")
	}

	return rjs.getJobStats(jobID)
}

//SetJobStatus is implementation of same method in JobStatsManager interface.
//Async method
func (rjs *RedisJobStatsManager) SetJobStatus(jobID string, status string) {
	if utils.IsEmptyStr(jobID) || utils.IsEmptyStr(status) {
		return
	}

	item := &queueItem{
		op:   opUpdateStatus,
		data: []string{jobID, status},
	}

	rjs.processChan <- item

	//Report status at the same time
	rjs.submitStatusReportingItem(jobID, status, "")
}

func (rjs *RedisJobStatsManager) loop() {
	controlChan := make(chan struct{})

	defer func() {
		rjs.isRunning.Store(false)
		//Notify other sub goroutines
		close(controlChan)
		logger.Info("Redis job stats manager is stopped")
	}()

	for {
		select {
		case item := <-rjs.processChan:
			go func(item *queueItem) {
				clearHookCache := false
				if err := rjs.process(item); err != nil {
					item.fails++
					if item.fails < maxFails {
						logger.Warningf("Failed to process '%s' request with error: %s\n", item.op, err)

						//Retry after a random interval
						go func() {
							timer := time.NewTimer(time.Duration(backoff(item.fails)) * time.Second)
							defer timer.Stop()

							select {
							case <-timer.C:
								rjs.processChan <- item
								return
							case <-controlChan:
							}
						}()
					} else {
						logger.Errorf("Failed to process '%s' request with error: %s (%d times tried)\n", item.op, err, maxFails)
						if item.op == opReportStatus {
							clearHookCache = true
						}
					}
				} else {
					if item.op == opReportStatus {
						clearHookCache = true
					}
				}

				if clearHookCache {
					//Clear cache to save memory if job status is success or stopped.
					data := item.data.([]string)
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

//SendCommand for the specified job
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

	//Directly add to op commands maintaining list
	return rjs.opCommands.Push(jobID, command)
}

//CheckIn mesage
func (rjs *RedisJobStatsManager) CheckIn(jobID string, message string) {
	if utils.IsEmptyStr(jobID) || utils.IsEmptyStr(message) {
		return
	}

	item := &queueItem{
		op:   opCheckIn,
		data: []string{jobID, message},
	}

	rjs.processChan <- item

	//Report checkin message at the same time
	rjs.submitStatusReportingItem(jobID, job.JobStatusRunning, message)
}

//CtlCommand checks if control command is fired for the specified job.
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

//DieAt marks the failed jobs with the time they put into dead queue.
func (rjs *RedisJobStatsManager) DieAt(jobID string, dieAt int64) {
	if utils.IsEmptyStr(jobID) || dieAt == 0 {
		return
	}

	item := &queueItem{
		op:   opDieAt,
		data: []interface{}{jobID, dieAt},
	}

	rjs.processChan <- item
}

//RegisterHook is used to save the hook url or cache the url in memory.
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

//ExpirePeriodicJobStats marks the periodic job stats expired
func (rjs *RedisJobStatsManager) ExpirePeriodicJobStats(jobID string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	//The periodic job (policy) is stopped/unscheduled and then
	//the stats of periodic job now can be expired
	key := utils.KeyJobStats(rjs.namespace, jobID)
	expireTime := 24 * 60 * 60 //1 day
	_, err := conn.Do("EXPIRE", key, expireTime)

	return err
}

func (rjs *RedisJobStatsManager) submitStatusReportingItem(jobID string, status, checkIn string) {
	//Let it run in a separate goroutine to avoid waiting more time
	go func() {
		var (
			hookURL string
			ok      bool
			err     error
		)

		hookURL, ok = rjs.hookStore.Get(jobID)
		if !ok {
			//Retrieve from backend
			hookURL, err = rjs.getHook(jobID)
			if err != nil || !utils.IsValidURL(hookURL) {
				//logged and exit
				logger.Warningf("no status hook found for job %s\n, abandon status reporting", jobID)
				return
			}
		}

		item := &queueItem{
			op:   opReportStatus,
			data: []string{jobID, hookURL, status, checkIn},
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

	return DefaultHookClient.ReportStatus(hookURL, reportingStatus)
}

func (rjs *RedisJobStatsManager) updateJobStatus(jobID string, status string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	args := make([]interface{}, 0, 6)
	args = append(args, key, "status", status, "update_time", time.Now().Unix())
	if status == job.JobStatusSuccess {
		//make sure the 'die_at' is reset in case it's a retrying job
		args = append(args, "die_at", 0)
	}
	_, err := conn.Do("HMSET", args...)

	return err
}

func (rjs *RedisJobStatsManager) checkIn(jobID string, message string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	now := time.Now().Unix()
	key := utils.KeyJobStats(rjs.namespace, jobID)
	args := make([]interface{}, 0, 7)
	args = append(args, key, "check_in", message, "check_in_at", now, "update_time", now)
	_, err := conn.Do("HMSET", args...)

	return err
}

func (rjs *RedisJobStatsManager) dieAt(jobID string, baseTime int64) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	//Query the dead job in the time scope of [baseTime,baseTime+5]
	key := utils.RedisKeyDead(rjs.namespace)
	jobWithScores, err := utils.GetZsetByScore(rjs.redisPool, key, []int64{baseTime, baseTime + 5})
	if err != nil {
		return err
	}

	for _, jws := range jobWithScores {
		if j, err := utils.DeSerializeJob(jws.JobBytes); err == nil {
			if j.ID == jobID {
				//Found
				statsKey := utils.KeyJobStats(rjs.namespace, jobID)
				args := make([]interface{}, 0, 7)
				args = append(args, statsKey, "die_at", jws.Score, "update_time", time.Now().Unix())
				_, err := conn.Do("HMSET", args...)
				return err
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

	conn.Send("HMSET", args...)
	//If job kind is periodic job, expire time should not be set
	//If job kind is scheduled job, expire time should be runAt+1day
	if jobStats.Stats.JobKind != job.JobKindPeriodic {
		var expireTime int64 = 60 * 60 * 24
		if jobStats.Stats.JobKind == job.JobKindScheduled {
			nowTime := time.Now().Unix()
			future := jobStats.Stats.RunAt - nowTime
			if future > 0 {
				expireTime += future
			}
		}
		expireTime += rand.Int63n(30) //Avoid lots of keys being expired at the same time
		conn.Send("EXPIRE", key, expireTime)
	}

	return conn.Flush()
}

func (rjs *RedisJobStatsManager) process(item *queueItem) error {
	switch item.op {
	case opSaveStats:
		jobStats := item.data.(models.JobStats)
		return rjs.saveJobStats(jobStats)
	case opUpdateStatus:
		data := item.data.([]string)
		return rjs.updateJobStatus(data[0], data[1])
	case opCheckIn:
		data := item.data.([]string)
		return rjs.checkIn(data[0], data[1])
	case opDieAt:
		data := item.data.([]interface{})
		return rjs.dieAt(data[0].(string), data[1].(int64))
	case opReportStatus:
		data := item.data.([]string)
		return rjs.reportStatus(data[0], data[1], data[2], data[3])
	default:
		break
	}

	return nil
}

//HookData keeps the hook url info
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

	//hook is saved into the job stats
	//We'll not set expire time here, the expire time of the key will be set when saving job stats
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
	vals, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return "", err
	}

	for i, l := 0, len(vals); i < l; i = i + 2 {
		prop := vals[i]
		value := vals[i+1]
		switch prop {
		case "status_hook":
			return value, nil
		default:
			break
		}
	}

	return "", fmt.Errorf("no hook found for job '%s'", jobID)
}

func backoff(seed uint) int {
	if seed < 1 {
		seed = 1
	}

	return int(math.Pow(float64(seed+1), float64(seed))) + rand.Intn(5)
}
