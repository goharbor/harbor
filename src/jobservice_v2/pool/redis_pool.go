// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

import (
	"errors"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job"
	"github.com/vmware/harbor/src/jobservice_v2/models"
	"github.com/vmware/harbor/src/jobservice_v2/opm"
	"github.com/vmware/harbor/src/jobservice_v2/period"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

var (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	workerPoolDeadTime    = 10 * time.Second
)

const (
	workerPoolStatusHealthy = "Healthy"
	workerPoolStatusDead    = "Dead"
)

//GoCraftWorkPool is the pool implementation based on gocraft/work powered by redis.
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

	//no need to sync as write once and then only read
	//key is name of known job
	//value is the type of known job
	knownJobs map[string]interface{}
}

//RedisPoolConfig defines configurations for GoCraftWorkPool.
type RedisPoolConfig struct {
	RedisHost   string
	RedisPort   uint
	Namespace   string
	WorkerCount uint
}

//RedisPoolContext ...
//We did not use this context to pass context info so far, just a placeholder.
type RedisPoolContext struct{}

//NewGoCraftWorkPool is constructor of goCraftWorkPool.
func NewGoCraftWorkPool(ctx *env.Context, cfg RedisPoolConfig) *GoCraftWorkPool {
	redisPool := &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort),
				redis.DialConnectTimeout(dialConnectionTimeout),
				redis.DialReadTimeout(dialReadTimeout),
				redis.DialWriteTimeout(dialWriteTimeout),
			)
		},
	}
	pool := work.NewWorkerPool(RedisPoolContext{}, cfg.WorkerCount, cfg.Namespace, redisPool)
	enqueuer := work.NewEnqueuer(cfg.Namespace, redisPool)
	client := work.NewClient(cfg.Namespace, redisPool)
	scheduler := period.NewRedisPeriodicScheduler(ctx, cfg.Namespace, redisPool)
	sweeper := period.NewSweeper(cfg.Namespace, redisPool, client)
	statsMgr := opm.NewRedisJobStatsManager(ctx.SystemContext, cfg.Namespace, redisPool, client, scheduler)
	msgServer := NewMessageServer(ctx.SystemContext, cfg.Namespace, redisPool)
	return &GoCraftWorkPool{
		namespace:     cfg.Namespace,
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

//Start to serve
//Unblock action
func (gcwp *GoCraftWorkPool) Start() {
	if gcwp.redisPool == nil ||
		gcwp.pool == nil ||
		gcwp.context.SystemContext == nil {
		//report and exit
		gcwp.context.ErrorChan <- errors.New("Redis worker pool can not start as it's not correctly configured")
		return
	}

	done := make(chan interface{}, 1)

	gcwp.context.WG.Add(1)
	go func() {
		var err error

		defer func() {
			gcwp.context.WG.Done()
			if err != nil {
				//report error
				gcwp.context.ErrorChan <- err
				done <- struct{}{} //exit immediately
			}
		}()

		//Register callbacks
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

		//Start message server
		if err = gcwp.messageServer.Start(); err != nil {
			return
		}
	}()

	gcwp.context.WG.Add(1)
	go func() {
		defer func() {
			gcwp.context.WG.Done()
			gcwp.statsManager.Shutdown()
		}()
		//Start stats manager
		//None-blocking
		gcwp.statsManager.Start()

		//blocking call
		gcwp.scheduler.Start()
	}()

	gcwp.context.WG.Add(1)
	go func() {
		defer func() {
			gcwp.context.WG.Done()
			log.Infof("Redis worker pool is stopped")
		}()

		//Clear dirty data before pool starting
		if err := gcwp.sweeper.ClearOutdatedScheduledJobs(); err != nil {
			//Only logged
			log.Errorf("Clear outdated data before pool starting failed with error:%s\n", err)
		}

		//Append middlewares
		gcwp.pool.Middleware((*RedisPoolContext).logJob)

		gcwp.pool.Start()
		log.Infof("Redis worker pool is started")

		//Block on listening context and done signal
		select {
		case <-gcwp.context.SystemContext.Done():
		case <-done:
		}

		gcwp.pool.Stop()
	}()
}

//RegisterJob is used to register the job to the pool.
//j is the type of job
func (gcwp *GoCraftWorkPool) RegisterJob(name string, j interface{}) error {
	if utils.IsEmptyStr(name) || j == nil {
		return errors.New("job can not be registered with empty name or nil interface")
	}

	//j must be job.Interface
	if _, ok := j.(job.Interface); !ok {
		return errors.New("job must implement the job.Interface")
	}

	redisJob := NewRedisJob(j, gcwp.context, gcwp.statsManager)

	//Get more info from j
	theJ := Wrap(j)

	gcwp.pool.JobWithOptions(name,
		work.JobOptions{MaxFails: theJ.MaxFails()},
		func(job *work.Job) error {
			return redisJob.Run(job)
		}, //Use generic handler to handle as we do not accept context with this way.
	)
	gcwp.knownJobs[name] = j //keep the name of registered jobs as known jobs for future validation

	return nil
}

//RegisterJobs is used to register multiple jobs to pool.
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

//Enqueue job
func (gcwp *GoCraftWorkPool) Enqueue(jobName string, params models.Parameters, isUnique bool) (models.JobStats, error) {
	var (
		j   *work.Job
		err error
	)

	//Enqueue job
	if isUnique {
		j, err = gcwp.enqueuer.EnqueueUnique(jobName, params)
	} else {
		j, err = gcwp.enqueuer.Enqueue(jobName, params)
	}

	if err != nil {
		return models.JobStats{}, err
	}

	//avoid backend pool bug
	if j == nil {
		return models.JobStats{}, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	res := generateResult(j, job.JobKindGeneric, isUnique)
	//Save data with async way. Once it fails to do, let it escape
	//The client method may help if the job is still in progress when get stats of this job
	gcwp.statsManager.Save(res)

	return res, nil
}

//Schedule job
func (gcwp *GoCraftWorkPool) Schedule(jobName string, params models.Parameters, runAfterSeconds uint64, isUnique bool) (models.JobStats, error) {
	var (
		j   *work.ScheduledJob
		err error
	)

	//Enqueue job in
	if isUnique {
		j, err = gcwp.enqueuer.EnqueueUniqueIn(jobName, int64(runAfterSeconds), params)
	} else {
		j, err = gcwp.enqueuer.EnqueueIn(jobName, int64(runAfterSeconds), params)
	}

	if err != nil {
		return models.JobStats{}, err
	}

	//avoid backend pool bug
	if j == nil {
		return models.JobStats{}, fmt.Errorf("job '%s' can not be enqueued, please check the job metatdata", jobName)
	}

	res := generateResult(j.Job, job.JobKindScheduled, isUnique)
	res.Stats.RunAt = j.RunAt

	//As job is already scheduled, we should not block this call
	//Once it fails to do, use client method to help get the status of the escape job
	gcwp.statsManager.Save(res)

	return res, nil
}

//PeriodicallyEnqueue job
func (gcwp *GoCraftWorkPool) PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error) {
	id, nextRun, err := gcwp.scheduler.Schedule(jobName, params, cronSetting)
	if err != nil {
		return models.JobStats{}, err
	}

	res := models.JobStats{
		Stats: &models.JobStatData{
			JobID:       id,
			JobName:     jobName,
			Status:      job.JobStatusPending,
			JobKind:     job.JobKindPeriodic,
			CronSpec:    cronSetting,
			EnqueueTime: time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", id),
			RunAt:       nextRun,
		},
	}

	gcwp.statsManager.Save(res)

	return res, nil
}

//GetJobStats return the job stats of the specified enqueued job.
func (gcwp *GoCraftWorkPool) GetJobStats(jobID string) (models.JobStats, error) {
	if utils.IsEmptyStr(jobID) {
		return models.JobStats{}, errors.New("empty job ID")
	}

	return gcwp.statsManager.Retrieve(jobID)
}

//Stats of pool
func (gcwp *GoCraftWorkPool) Stats() (models.JobPoolStats, error) {
	//Get the status of workerpool via client
	hbs, err := gcwp.client.WorkerPoolHeartbeats()
	if err != nil {
		return models.JobPoolStats{}, err
	}

	//Find the heartbeat of this pool via pid
	stats := make([]*models.JobPoolStatsData, 0)
	for _, hb := range hbs {
		if hb.HeartbeatAt == 0 {
			continue //invalid ones
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

//StopJob will stop the job
func (gcwp *GoCraftWorkPool) StopJob(jobID string) error {
	return gcwp.statsManager.Stop(jobID)
}

//CancelJob will cancel the job
func (gcwp *GoCraftWorkPool) CancelJob(jobID string) error {
	return gcwp.statsManager.Cancel(jobID)
}

//RetryJob retry the job
func (gcwp *GoCraftWorkPool) RetryJob(jobID string) error {
	return gcwp.statsManager.Retry(jobID)
}

//IsKnownJob ...
func (gcwp *GoCraftWorkPool) IsKnownJob(name string) (interface{}, bool) {
	v, ok := gcwp.knownJobs[name]
	return v, ok
}

//ValidateJobParameters ...
func (gcwp *GoCraftWorkPool) ValidateJobParameters(jobType interface{}, params map[string]interface{}) error {
	if jobType == nil {
		return errors.New("nil job type")
	}

	theJ := Wrap(jobType)
	return theJ.Validate(params)
}

//RegisterHook registers status hook url
//sync method
func (gcwp *GoCraftWorkPool) RegisterHook(jobID string, hookURL string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	if utils.IsEmptyStr(hookURL) {
		return errors.New("empty hook url")
	}

	return gcwp.statsManager.RegisterHook(jobID, hookURL, false)
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

//log the job
func (rpc *RedisPoolContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Infof("Job incoming: %s:%s", job.Name, job.ID)
	return next()
}

//generate the job stats data
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
