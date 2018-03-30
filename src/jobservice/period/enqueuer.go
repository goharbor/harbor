//Refer github.com/gocraft/work

package period

import (
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/robfig/cron"
	"github.com/vmware/harbor/src/jobservice/job"
	"github.com/vmware/harbor/src/jobservice/logger"
	"github.com/vmware/harbor/src/jobservice/utils"
)

const (
	periodicEnqueuerSleep   = 2 * time.Minute
	periodicEnqueuerHorizon = 4 * time.Minute
)

type periodicEnqueuer struct {
	namespace             string
	pool                  *redis.Pool
	policyStore           *periodicJobPolicyStore
	scheduledPeriodicJobs []*scheduledPeriodicJob
	stopChan              chan struct{}
	doneStoppingChan      chan struct{}
}

type periodicJob struct {
	jobName  string
	spec     string
	schedule cron.Schedule
}

type scheduledPeriodicJob struct {
	scheduledAt      time.Time
	scheduledAtEpoch int64
	*periodicJob
}

func newPeriodicEnqueuer(namespace string, pool *redis.Pool, policyStore *periodicJobPolicyStore) *periodicEnqueuer {
	return &periodicEnqueuer{
		namespace:        namespace,
		pool:             pool,
		policyStore:      policyStore,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}
}

func (pe *periodicEnqueuer) start() {
	go pe.loop()
	logger.Info("Periodic enqueuer is started")
}

func (pe *periodicEnqueuer) stop() {
	pe.stopChan <- struct{}{}
	<-pe.doneStoppingChan
}

func (pe *periodicEnqueuer) loop() {
	defer func() {
		logger.Info("Periodic enqueuer is stopped")
	}()
	// Begin reaping periodically
	timer := time.NewTimer(periodicEnqueuerSleep + time.Duration(rand.Intn(30))*time.Second)
	defer timer.Stop()

	if pe.shouldEnqueue() {
		err := pe.enqueue()
		if err != nil {
			logger.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
		}
	}

	for {
		select {
		case <-pe.stopChan:
			pe.doneStoppingChan <- struct{}{}
			return
		case <-timer.C:
			timer.Reset(periodicEnqueuerSleep + time.Duration(rand.Intn(30))*time.Second)
			if pe.shouldEnqueue() {
				err := pe.enqueue()
				if err != nil {
					logger.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
				}
			}
		}
	}
}

func (pe *periodicEnqueuer) enqueue() error {
	now := utils.NowEpochSeconds()
	nowTime := time.Unix(now, 0)
	horizon := nowTime.Add(periodicEnqueuerHorizon)

	conn := pe.pool.Get()
	defer conn.Close()

	for _, pl := range pe.policyStore.list() {
		schedule, err := cron.Parse(pl.CronSpec)
		if err != nil {
			//The cron spec should be already checked at top components.
			//Just in cases, if error occurred, ignore it
			continue
		}
		pj := &periodicJob{
			jobName:  pl.JobName,
			spec:     pl.CronSpec,
			schedule: schedule,
		}
		for t := pj.schedule.Next(nowTime); t.Before(horizon); t = pj.schedule.Next(t) {
			epoch := t.Unix()
			job := &work.Job{
				Name: pj.jobName,
				ID:   pl.PolicyID, //Same with the id of the policy it's being scheduled for

				// This is technically wrong, but this lets the bytes be identical for the same periodic job instance. If we don't do this, we'd need to use a different approach -- probably giving each periodic job its own history of the past 100 periodic jobs, and only scheduling a job if it's not in the history.
				EnqueuedAt: epoch,
				Args:       pl.JobParameters, //Pass parameters to scheduled job here
			}

			rawJSON, err := utils.SerializeJob(job)
			if err != nil {
				return err
			}

			_, err = conn.Do("ZADD", utils.RedisKeyScheduled(pe.namespace), epoch, rawJSON)
			if err != nil {
				return err
			}

			logger.Infof("Schedule job %s for policy %s at %d\n", pj.jobName, pl.PolicyID, epoch)
		}
		//Directly use redis conn to update the periodic job (policy) status
		//Do not care the result
		conn.Do("HMSET", utils.KeyJobStats(pe.namespace, pl.PolicyID), "status", job.JobStatusScheduled, "update_time", time.Now().Unix())
	}

	_, err := conn.Do("SET", utils.RedisKeyLastPeriodicEnqueue(pe.namespace), now)

	return err
}

func (pe *periodicEnqueuer) shouldEnqueue() bool {
	conn := pe.pool.Get()
	defer conn.Close()

	lastEnqueue, err := redis.Int64(conn.Do("GET", utils.RedisKeyLastPeriodicEnqueue(pe.namespace)))
	if err == redis.ErrNil {
		return true
	} else if err != nil {
		logger.Errorf("periodic_enqueuer.should_enqueue:%s\n", err)
		return true
	}

	return lastEnqueue < (utils.NowEpochSeconds() - int64(periodicEnqueuerSleep/time.Minute))
}
