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
	"github.com/vmware/harbor/src/jobservice_v2/period"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

var (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

//GoCraftWorkPool is the pool implementation based on gocraft/work powered by redis.
type GoCraftWorkPool struct {
	redisPool *redis.Pool
	pool      *work.WorkerPool
	enqueuer  *work.Enqueuer
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
	return &GoCraftWorkPool{
		redisPool: redisPool,
		pool:      pool,
		enqueuer:  enqueuer,
		scheduler: scheduler,
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

	return generateResult(j), nil
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

	res := generateResult(j.Job)
	res.Stats.RunAt = time.Unix(j.RunAt, 0)

	return res, nil
}

//PeriodicallyEnqueue job
func (gcwp *GoCraftWorkPool) PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error) {
	id, err := gcwp.scheduler.Schedule(jobName, params, cronSetting)
	if err != nil {
		return models.JobStats{}, err
	}

	//TODO: Need more data
	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:       id,
			JobName:     jobName,
			Status:      job.JobStatusPending,
			EnqueueTime: time.Unix(time.Now().Unix(), 0),
			UpdateTime:  time.Unix(time.Now().Unix(), 0),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", id),
		},
	}, nil
}

//Stats of pool
func (gcwp *GoCraftWorkPool) Stats() (models.JobPoolStats, error) {
	return models.JobPoolStats{}, nil
}

//IsKnownJob ...
func (gcwp *GoCraftWorkPool) IsKnownJob(name string) (bool, bool) {
	v, ok := gcwp.knownJobs[name]
	return ok, v
}

//log the job
func (rpc *RedisPoolContext) logJob(job *work.Job, next work.NextMiddlewareFunc) error {
	//TODO: Also update the job status to 'pending'
	log.Infof("Job incoming: %s:%s", job.ID, job.Name)
	return next()
}

//generate the job stats data
func generateResult(j *work.Job) models.JobStats {
	if j == nil {
		return models.JobStats{}
	}

	return models.JobStats{
		Stats: &models.JobStatData{
			JobID:       j.ID,
			JobName:     j.Name,
			Status:      job.JobStatusPending,
			EnqueueTime: time.Unix(j.EnqueuedAt, 0),
			UpdateTime:  time.Unix(time.Now().Unix(), 0),
			RefLink:     fmt.Sprintf("/api/v1/jobs/%s", j.ID),
		},
	}
}
