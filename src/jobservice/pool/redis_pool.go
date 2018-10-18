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

package pool

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron"
)

var (
	workerPoolDeadTime = 10 * time.Second
)

const (
	workerPoolStatusHealthy = "Healthy"
	workerPoolStatusDead    = "Dead"

	// Copy from period.enqueuer
	periodicEnqueuerHorizon = 4 * time.Minute

	pingRedisMaxTimes = 10
)

// GoCraftWorkPool is the pool implementation based on gocraft/work powered by redis.
type GoCraftWorkPool struct {
	namespace     string
	redisPool     *redis.Pool
	pool          *work.WorkerPool
	enqueuer      *work.Enqueuer
	sweeper       *period.Sweeper
	client        *work.Client
	context       *env.Context
	scheduler     period.Interface
	statsManager  opm.JobStatsManager
	messageServer *MessageServer

	// no need to sync as write once and then only read
	// key is name of known job
	// value is the type of known job
	knownJobs map[string]interface{}
}

// RedisPoolContext ...
// We did not use this context to pass context info so far, just a placeholder.
type RedisPoolContext struct{}

// NewGoCraftWorkPool is constructor of goCraftWorkPool.
func NewGoCraftWorkPool(ctx *env.Context, namespace string, workerCount uint, redisPool *redis.Pool) *GoCraftWorkPool {
	pool := work.NewWorkerPool(RedisPoolContext{}, workerCount, namespace, redisPool)
	enqueuer := work.NewEnqueuer(namespace, redisPool)
	client := work.NewClient(namespace, redisPool)
	statsMgr := opm.NewRedisJobStatsManager(ctx.SystemContext, namespace, redisPool)
	scheduler := period.NewRedisPeriodicScheduler(ctx, namespace, redisPool, statsMgr)
	sweeper := period.NewSweeper(namespace, redisPool, client)
	msgServer := NewMessageServer(ctx.SystemContext, namespace, redisPool)
	return &GoCraftWorkPool{
		namespace:     namespace,
		redisPool:     redisPool,
		pool:          pool,
		enqueuer:      enqueuer,
		scheduler:     scheduler,
		sweeper:       sweeper,
		client:        client,
		context:       ctx,
		statsManager:  statsMgr,
		knownJobs:     make(map[string]interface{}),
		messageServer: msgServer,
	}
}

// Start to serve
// Unblock action
func (gcwp *GoCraftWorkPool) Start() error {
	if gcwp.redisPool == nil ||
		gcwp.pool == nil ||
		gcwp.context.SystemContext == nil {
		// report and exit
		return errors.New("Redis worker pool can not start as it's not correctly configured")
	}

	// Test the redis connection
	if err := gcwp.ping(); err != nil {
		return err
	}

	done := make(chan interface{}, 1)

	gcwp.context.WG.Add(1)
	go func() {
		var err error

		defer func() {
			gcwp.context.WG.Done()
			if err != nil {
				// report error
				gcwp.context.ErrorChan <- err
				done <- struct{}{} // exit immediately
			}
		}()

		// Register callbacks
		if err = gcwp.messageServer.Subscribe(period.EventSchedulePeriodicPolicy,
			func(data interface{}) error {
				return gcwp.handleSchedulePolicy(data)
			}); err != nil {
			return
		}
		if err = gcwp.messageServer.Subscribe(period.EventUnSchedulePeriodicPolicy,
			func(data interface{}) error {
				return gcwp.handleUnSchedulePolicy(data)
			}); err != nil {
			return
		}
		if err = gcwp.messageServer.Subscribe(opm.EventRegisterStatusHook,
			func(data interface{}) error {
				return gcwp.handleRegisterStatusHook(data)
			}); err != nil {
			return
		}
		if err = gcwp.messageServer.Subscribe(opm.EventFireCommand,
			func(data interface{}) error {
				return gcwp.handleOPCommandFiring(data)
			}); err != nil {
			return
		}

		startTimes := 0
	START_MSG_SERVER:
		// Start message server
		if err = gcwp.messageServer.Start(); err != nil {
			logger.Errorf("Message server exits with error: %s\n", err.Error())
			if startTimes < msgServerRetryTimes {
				startTimes++
				time.Sleep(time.Duration((int)(math.Pow(2, (float64)(startTimes)))+5) * time.Second)
				logger.Infof("Restart message server (%d times)\n", startTimes)
				goto START_MSG_SERVER
			}

			return
		}
	}()

	gcwp.context.WG.Add(1)
	go func() {
		defer func() {
			gcwp.context.WG.Done()
			gcwp.statsManager.Shutdown()
		}()
		// Start stats manager
		// None-blocking
		gcwp.statsManager.Start()

		// blocking call
		gcwp.scheduler.Start()
	}()

	gcwp.context.WG.Add(1)
	go func() {
		defer func() {
			gcwp.context.WG.Done()
			logger.Infof("Redis worker pool is stopped")
		}()

		// Clear dirty data before pool starting
		if err := gcwp.sweeper.ClearOutdatedScheduledJobs(); err != nil {
			// Only logged
			logger.Errorf("Clear outdated data before pool starting failed with error:%s\n", err)
		}

		// Append middlewares
		gcwp.pool.Middleware((*RedisPoolContext).logJob)

		gcwp.pool.Start()
		logger.Infof("Redis worker pool is started")

		// Block on listening context and done signal
		select {
		case <-gcwp.context.SystemContext.Done():
		case <-done:
		}

		gcwp.pool.Stop()
	}()

	return nil
}

// RegisterJob is used to register the job to the pool.
// j is the type of job
func (gcwp *GoCraftWorkPool) RegisterJob(name string, j interface{}) error {
	if utils.IsEmptyStr(name) || j == nil {
		return errors.New("job can not be registered with empty name or nil interface")
	}

	// j must be job.Interface
	if _, ok := j.(job.Interface); !ok {
		return errors.New("job must implement the job.Interface")
	}

	// 1:1 constraint
	if jInList, ok := gcwp.knownJobs[name]; ok {
		return fmt.Errorf("Job name %s has been already registered with %s", name, reflect.TypeOf(jInList).String())
	}

	// Same job implementation can be only registered with one name
	for jName, jInList := range gcwp.knownJobs {
		jobImpl := reflect.TypeOf(j).String()
		if reflect.TypeOf(jInList).String() == jobImpl {
			return fmt.Errorf("Job %s has been already registered with name %s", jobImpl, jName)
		}
	}

	redisJob := NewRedisJob(j, gcwp.context, gcwp.statsManager)

	// Get more info from j
	theJ := Wrap(j)

	gcwp.pool.JobWithOptions(name,
		work.JobOptions{MaxFails: theJ.MaxFails()},
		func(job *work.Job) error {
			return redisJob.Run(job)
		}, // Use generic handler to handle as we do not accept context with this way.
	)
	gcwp.knownJobs[name] = j // keep the name of registered jobs as known jobs for future validation

	logger.Infof("Register job %s with name %s", reflect.TypeOf(j).String(), name)

	return nil
}

// RegisterJobs is used to register multiple jobs to pool.
func (gcwp *GoCraftWorkPool) RegisterJobs(jobs map[string]interface{}) error {
	if jobs == nil || len(jobs) == 0 {
		return nil
	}

	for name, j := range jobs {
		if err := gcwp.RegisterJob(name, j); err != nil {
			return err
		}
	}

	return nil
}

// Enqueue job
func (gcwp *GoCraftWorkPool) Enqueue(jobName string, params models.Parameters, isUnique bool) (models.JobStats, error) {
	var (
		j   *work.Job
		err error
	)

	// Enqueue job
	if isUnique {
		j, err = gcwp.enqueuer.EnqueueUnique(jobName, params)
	} else {
		j, err = gcwp.enqueuer.Enqueue(jobName, params)
	}

	if err != nil {
		return models.JobStats{}, err
	}

	// avoid backend pool bug
	if j == nil {
		return models.JobStats{}, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	res := generateResult(j, job.JobKindGeneric, isUnique)
	// Save data with async way. Once it fails to do, let it escape
	// The client method may help if the job is still in progress when get stats of this job
	gcwp.statsManager.Save(res)

	return res, nil
}

// Schedule job
func (gcwp *GoCraftWorkPool) Schedule(jobName string, params models.Parameters, runAfterSeconds uint64, isUnique bool) (models.JobStats, error) {
	var (
		j   *work.ScheduledJob
		err error
	)

	// Enqueue job in
	if isUnique {
		j, err = gcwp.enqueuer.EnqueueUniqueIn(jobName, int64(runAfterSeconds), params)
	} else {
		j, err = gcwp.enqueuer.EnqueueIn(jobName, int64(runAfterSeconds), params)
	}

	if err != nil {
		return models.JobStats{}, err
	}

	// avoid backend pool bug
	if j == nil {
		return models.JobStats{}, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	res := generateResult(j.Job, job.JobKindScheduled, isUnique)
	res.Stats.RunAt = j.RunAt

	// As job is already scheduled, we should not block this call
	// Once it fails to do, use client method to help get the status of the escape job
	gcwp.statsManager.Save(res)

	return res, nil
}

// PeriodicallyEnqueue job
func (gcwp *GoCraftWorkPool) PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error) {
	id, nextRun, err := gcwp.scheduler.Schedule(jobName, params, cronSetting)
	if err != nil {
		return models.JobStats{}, err
	}

	res := models.JobStats{
		Stats: &models.JobStatData{
			JobID:                id,
			JobName:              jobName,
			Status:               job.JobStatusPending,
			JobKind:              job.JobKindPeriodic,
			CronSpec:             cronSetting,
			EnqueueTime:          time.Now().Unix(),
			UpdateTime:           time.Now().Unix(),
			RefLink:              fmt.Sprintf("/api/v1/jobs/%s", id),
			RunAt:                nextRun,
			IsMultipleExecutions: true, // True for periodic job
		},
	}

	gcwp.statsManager.Save(res)

	return res, nil
}

// GetJobStats return the job stats of the specified enqueued job.
func (gcwp *GoCraftWorkPool) GetJobStats(jobID string) (models.JobStats, error) {
	if utils.IsEmptyStr(jobID) {
		return models.JobStats{}, errors.New("empty job ID")
	}

	return gcwp.statsManager.Retrieve(jobID)
}

// Stats of pool
func (gcwp *GoCraftWorkPool) Stats() (models.JobPoolStats, error) {
	// Get the status of workerpool via client
	hbs, err := gcwp.client.WorkerPoolHeartbeats()
	if err != nil {
		return models.JobPoolStats{}, err
	}

	// Find the heartbeat of this pool via pid
	stats := make([]*models.JobPoolStatsData, 0)
	for _, hb := range hbs {
		if hb.HeartbeatAt == 0 {
			continue // invalid ones
		}

		wPoolStatus := workerPoolStatusHealthy
		if time.Unix(hb.HeartbeatAt, 0).Add(workerPoolDeadTime).Before(time.Now()) {
			wPoolStatus = workerPoolStatusDead
		}
		stat := &models.JobPoolStatsData{
			WorkerPoolID: hb.WorkerPoolID,
			StartedAt:    hb.StartedAt,
			HeartbeatAt:  hb.HeartbeatAt,
			JobNames:     hb.JobNames,
			Concurrency:  hb.Concurrency,
			Status:       wPoolStatus,
		}
		stats = append(stats, stat)
	}

	if len(stats) == 0 {
		return models.JobPoolStats{}, errors.New("Failed to get stats of worker pools")
	}

	return models.JobPoolStats{
		Pools: stats,
	}, nil
}

// StopJob will stop the job
func (gcwp *GoCraftWorkPool) StopJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	theJob, err := gcwp.statsManager.Retrieve(jobID)
	if err != nil {
		return err
	}

	needSetStopStatus := false

	switch theJob.Stats.JobKind {
	case job.JobKindGeneric:
		// Only running job can be stopped
		if theJob.Stats.Status != job.JobStatusRunning {
			return fmt.Errorf("job '%s' is not a running job", jobID)
		}
	case job.JobKindScheduled:
		// we need to delete the scheduled job in the queue if it is not running yet
		// otherwise, nothing need to do
		if theJob.Stats.Status == job.JobStatusScheduled {
			if err := gcwp.client.DeleteScheduledJob(theJob.Stats.RunAt, jobID); err != nil {
				return err
			}
			needSetStopStatus = true
		}
	case job.JobKindPeriodic:
		// firstly delete the periodic job policy
		if err := gcwp.scheduler.UnSchedule(jobID); err != nil {
			return err
		}

		logger.Infof("Periodic job policy %s is removed", jobID)

		// secondly we need try to delete the job instances scheduled for this periodic job, a try best action
		gcwp.deleteScheduledJobsOfPeriodicPolicy(theJob.Stats.JobID, theJob.Stats.CronSpec) // ignore error as we have logged
		// thirdly expire the job stats of this periodic job if exists
		if err := gcwp.statsManager.ExpirePeriodicJobStats(theJob.Stats.JobID); err != nil {
			// only logged
			logger.Errorf("Expire the stats of job %s failed with error: %s\n", theJob.Stats.JobID, err)
		}

		needSetStopStatus = true
	default:
		break
	}

	// Check if the job has 'running' instance
	if theJob.Stats.Status == job.JobStatusRunning {
		// Send 'stop' ctl command to the running instance
		if err := gcwp.statsManager.SendCommand(jobID, opm.CtlCommandStop); err != nil {
			return err
		}
		// The job running instance will set the status to 'stopped'
		needSetStopStatus = false
	}

	// If needed, update the job status to 'stopped'
	if needSetStopStatus {
		gcwp.statsManager.SetJobStatus(jobID, job.JobStatusStopped)
	}

	return nil
}

// CancelJob will cancel the job
func (gcwp *GoCraftWorkPool) CancelJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	theJob, err := gcwp.statsManager.Retrieve(jobID)
	if err != nil {
		return err
	}

	switch theJob.Stats.JobKind {
	case job.JobKindGeneric:
		if theJob.Stats.Status != job.JobStatusRunning {
			return fmt.Errorf("only running job can be cancelled, job '%s' seems not running now", theJob.Stats.JobID)
		}

		// Send 'cancel' ctl command to the running instance
		if err := gcwp.statsManager.SendCommand(jobID, opm.CtlCommandCancel); err != nil {
			return err
		}
		break
	default:
		return fmt.Errorf("job kind '%s' does not support 'cancel' operation", theJob.Stats.JobKind)
	}

	return nil
}

// RetryJob retry the job
func (gcwp *GoCraftWorkPool) RetryJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	theJob, err := gcwp.statsManager.Retrieve(jobID)
	if err != nil {
		return err
	}

	if theJob.Stats.DieAt == 0 {
		return fmt.Errorf("job '%s' is not a retryable job", jobID)
	}

	return gcwp.client.RetryDeadJob(theJob.Stats.DieAt, jobID)
}

// IsKnownJob ...
func (gcwp *GoCraftWorkPool) IsKnownJob(name string) (interface{}, bool) {
	v, ok := gcwp.knownJobs[name]
	return v, ok
}

// ValidateJobParameters ...
func (gcwp *GoCraftWorkPool) ValidateJobParameters(jobType interface{}, params map[string]interface{}) error {
	if jobType == nil {
		return errors.New("nil job type")
	}

	theJ := Wrap(jobType)
	return theJ.Validate(params)
}

// RegisterHook registers status hook url
// sync method
func (gcwp *GoCraftWorkPool) RegisterHook(jobID string, hookURL string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if !utils.IsValidURL(hookURL) {
		return errors.New("invalid hook url")
	}

	return gcwp.statsManager.RegisterHook(jobID, hookURL, false)
}

func (gcwp *GoCraftWorkPool) deleteScheduledJobsOfPeriodicPolicy(policyID string, cronSpec string) error {
	schedule, err := cron.Parse(cronSpec)
	if err != nil {
		logger.Errorf("cron spec '%s' is not valid", cronSpec)
		return err
	}

	now := time.Now().Unix()
	nowTime := time.Unix(now, 0)
	horizon := nowTime.Add(periodicEnqueuerHorizon)

	// try to delete more
	// return the last error if occurred
	for t := schedule.Next(nowTime); t.Before(horizon); t = schedule.Next(t) {
		epoch := t.Unix()
		if err = gcwp.client.DeleteScheduledJob(epoch, policyID); err != nil {
			// only logged
			logger.Warningf("Delete scheduled instance for periodic job %s failed with error: %s\n", policyID, err)
		} else {
			logger.Infof("Delete scheduled job for periodic job policy %s: runat = %d", policyID, epoch)
		}
	}

	return err
}

func (gcwp *GoCraftWorkPool) handleSchedulePolicy(data interface{}) error {
	if data == nil {
		return errors.New("nil data interface")
	}

	pl, ok := data.(*period.PeriodicJobPolicy)
	if !ok {
		return errors.New("malformed policy object")
	}

	return gcwp.scheduler.AcceptPeriodicPolicy(pl)
}

func (gcwp *GoCraftWorkPool) handleUnSchedulePolicy(data interface{}) error {
	if data == nil {
		return errors.New("nil data interface")
	}

	pl, ok := data.(*period.PeriodicJobPolicy)
	if !ok {
		return errors.New("malformed policy object")
	}

	removed := gcwp.scheduler.RemovePeriodicPolicy(pl.PolicyID)
	if removed == nil {
		return errors.New("nothing removed")
	}

	return nil
}

func (gcwp *GoCraftWorkPool) handleRegisterStatusHook(data interface{}) error {
	if data == nil {
		return errors.New("nil data interface")
	}

	hook, ok := data.(*opm.HookData)
	if !ok {
		return errors.New("malformed hook object")
	}

	return gcwp.statsManager.RegisterHook(hook.JobID, hook.HookURL, true)
}

func (gcwp *GoCraftWorkPool) handleOPCommandFiring(data interface{}) error {
	if data == nil {
		return errors.New("nil data interface")
	}

	commands, ok := data.([]interface{})
	if !ok || len(commands) != 2 {
		return errors.New("malformed op commands object")
	}
	jobID, ok := commands[0].(string)
	command, ok := commands[1].(string)
	if !ok {
		return errors.New("malformed op command info")
	}

	return gcwp.statsManager.SendCommand(jobID, command)
}

// log the job
func (rpc *RedisPoolContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	logger.Infof("Job incoming: %s:%s", job.Name, job.ID)
	return next()
}

// Ping the redis server
func (gcwp *GoCraftWorkPool) ping() error {
	conn := gcwp.redisPool.Get()
	defer conn.Close()

	var err error
	for count := 1; count <= pingRedisMaxTimes; count++ {
		if _, err = conn.Do("ping"); err == nil {
			return nil
		}

		time.Sleep(time.Duration(count+4) * time.Second)
	}

	return fmt.Errorf("connect to redis server timeout: %s", err.Error())
}

// generate the job stats data
func generateResult(j *work.Job, jobKind string, isUnique bool) models.JobStats {
	if j == nil {
		return models.JobStats{}
	}

	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:       j.ID,
			JobName:     j.Name,
			JobKind:     jobKind,
			IsUnique:    isUnique,
			Status:      job.JobStatusPending,
			EnqueueTime: j.EnqueuedAt,
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", j.ID),
		},
	}
}
