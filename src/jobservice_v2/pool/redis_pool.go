// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job"
	"github.com/vmware/harbor/src/jobservice_v2/models"
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
	namespace string
	redisPool *redis.Pool
	pool      *work.WorkerPool
	enqueuer  *work.Enqueuer
	sweeper   *period.Sweeper
	client    *work.Client
	context   *env.Context
	scheduler period.Interface

	//no need to sync as write once and then only read
	//key is name of known job
	//value is the flag indicating if the job requires parameters
	knownJobs map[string]bool
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
	scheduler := period.NewRedisPeriodicScheduler(ctx.SystemContext, cfg.Namespace, redisPool)
	sweeper := period.NewSweeper(cfg.Namespace, redisPool, client)
	return &GoCraftWorkPool{
		namespace: cfg.Namespace,
		redisPool: redisPool,
		pool:      pool,
		enqueuer:  enqueuer,
		scheduler: scheduler,
		sweeper:   sweeper,
		client:    client,
		context:   ctx,
		knownJobs: make(map[string]bool),
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
		defer func() {
			gcwp.context.WG.Done()
		}()
		//blocking call
		if err := gcwp.scheduler.Start(); err != nil {
			//Scheduler exits with error
			gcwp.context.ErrorChan <- err
			done <- struct{}{}
			return
		}
	}()

	gcwp.context.WG.Add(1)
	go func() {
		defer func() {
			gcwp.context.WG.Done()
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
		log.Infof("Redis worker pool is stopped")
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

	//Use redis job wrapper pointer to keep the data required by the job.Interface.
	redisJob := job.NewRedisJob(j, gcwp.context)

	//Get more info from j
	theJ := redisJob.Wrap()

	gcwp.pool.JobWithOptions(name,
		work.JobOptions{MaxFails: theJ.MaxFails()},
		func(job *work.Job) error {
			return redisJob.Run(job)
		}, //Use generic handler to handle as we do not accept context with this way.
	)
	gcwp.knownJobs[name] = theJ.ParamsRequired() //keep the name of registered jobs as known jobs for future validation

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

	res := generateResult(j, job.JobKindGeneric, isUnique)
	if err := gcwp.saveStats(res); err != nil {
		//Once running job, let it fly away
		//The client method may help if the job is still in progress when get stats of this job
		log.Errorf("Failed to save stats of job %s with error: %s\n", res.Stats.JobID, err)
	}
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

	res := generateResult(j.Job, job.JobKindScheduled, isUnique)
	res.Stats.RunAt = j.RunAt

	if err := gcwp.saveStats(res); err != nil {
		//As job is already scheduled, we should not block this call
		//Use client method to help get the status of this fly-away job
		log.Errorf("Failed to save stats of job %s with error: %s\n", res.Stats.JobID, err)
	}

	return res, nil
}

//PeriodicallyEnqueue job
func (gcwp *GoCraftWorkPool) PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error) {
	id, err := gcwp.scheduler.Schedule(jobName, params, cronSetting)
	if err != nil {
		return models.JobStats{}, err
	}

	//TODO: Need more data
	//TODO: EnqueueTime should be got from cron spec
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:       id,
			JobName:     jobName,
			Status:      job.JobStatusPending,
			EnqueueTime: time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", id),
		},
	}, nil
}

//Stats of pool
func (gcwp *GoCraftWorkPool) Stats() (models.JobPoolStats, error) {
	//Get the status of workerpool via client
	hbs, err := gcwp.client.WorkerPoolHeartbeats()
	if err != nil {
		return models.JobPoolStats{}, err
	}

	//Find the heartbeat of this pool via pid
	pid := os.Getpid()
	for _, hb := range hbs {
		if hb.Pid == pid {
			wPoolStatus := workerPoolStatusHealthy
			if time.Unix(hb.HeartbeatAt, 0).Add(workerPoolDeadTime).Before(time.Now()) {
				wPoolStatus = workerPoolStatusDead
			}
			stats := models.JobPoolStats{
				WorkerPoolID: hb.WorkerPoolID,
				StartedAt:    hb.StartedAt,
				HeartbeatAt:  hb.HeartbeatAt,
				JobNames:     hb.JobNames,
				Concurrency:  hb.Concurrency,
				Status:       wPoolStatus,
			}

			return stats, nil
		}
	}

	return models.JobPoolStats{}, errors.New("Failed to get stats of worker pool")
}

//IsKnownJob ...
func (gcwp *GoCraftWorkPool) IsKnownJob(name string) (bool, bool) {
	v, ok := gcwp.knownJobs[name]
	return ok, v
}

func (gcwp *GoCraftWorkPool) saveStats(stats models.JobStats) error {
	conn := gcwp.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(gcwp.namespace, stats.Stats.JobID)
	args := make([]interface{}, 0)
	args = append(args, key)
	args = append(args,
		"id", stats.Stats.JobID,
		"name", stats.Stats.JobName,
		"kind", stats.Stats.JobKind,
		"unique", stats.Stats.IsUnique,
		"status", stats.Stats.Status,
		"ref_link", stats.Stats.RefLink,
		"enqueue_time", stats.Stats.EnqueueTime,
		"update_time", stats.Stats.UpdateTime,
		"run_at", stats.Stats.RunAt,
	)
	if stats.Stats.CheckInAt > 0 && !utils.IsEmptyStr(stats.Stats.CheckIn) {
		args = append(args,
			"check_in", stats.Stats.CheckIn,
			"check_in_at", stats.Stats.CheckInAt,
		)
	}

	conn.Send("HMSET", args...)
	//If job kind is periodic job, expire time should be set
	//If job kind is scheduled job, expire time should be runAt+1day
	if stats.Stats.JobKind != job.JobKindPeriodic {
		var expireTime int64 = 60 * 60 * 24
		if stats.Stats.JobKind == job.JobKindScheduled {
			nowTime := time.Now().Unix()
			future := stats.Stats.RunAt - nowTime
			if future > 0 {
				expireTime += future
			}
		}
		conn.Send("EXPIRE", key, expireTime)
	}

	return conn.Flush()
}

//log the job
func (rpc *RedisPoolContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	//TODO: Also update the job status to 'pending'
	log.Infof("Job incoming: %s:%s", job.ID, job.Name)
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
