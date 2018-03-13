//Refer github.com/gocraft/work

package period

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/robfig/cron"
	"github.com/vmware/harbor/src/common/utils/log"
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
	log.Info("Periodic enqueuer is started")
}

func (pe *periodicEnqueuer) stop() {
	pe.stopChan <- struct{}{}
	<-pe.doneStoppingChan
}

func (pe *periodicEnqueuer) loop() {
	defer func() {
		log.Info("Periodic enqueuer is stopped")
	}()
	// Begin reaping periodically
	timer := time.NewTimer(periodicEnqueuerSleep + time.Duration(rand.Intn(30))*time.Second)
	defer timer.Stop()

	if pe.shouldEnqueue() {
		err := pe.enqueue()
		if err != nil {
			log.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
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
					log.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
				}
			}
		}
	}
}

func (pe *periodicEnqueuer) enqueue() error {
	now := nowEpochSeconds()
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
			id := makeUniquePeriodicID(pj.jobName, pl.PolicyID, epoch) //Use policy ID to track the jobs related with it

			job := &work.Job{
				Name: pj.jobName,
				ID:   id,

				// This is technically wrong, but this lets the bytes be identical for the same periodic job instance. If we don't do this, we'd need to use a different approach -- probably giving each periodic job its own history of the past 100 periodic jobs, and only scheduling a job if it's not in the history.
				EnqueuedAt: epoch,
				Args:       pl.JobParameters, //Pass parameters to scheduled job here
			}

			rawJSON, err := serializeJob(job)
			if err != nil {
				return err
			}

			_, err = conn.Do("ZADD", redisKeyScheduled(pe.namespace), epoch, rawJSON)
			if err != nil {
				return err
			}

			log.Infof("Schedule job %s for policy %s\n", pj.jobName, pl.PolicyID)
		}
	}

	_, err := conn.Do("SET", redisKeyLastPeriodicEnqueue(pe.namespace), now)

	return err
}

func (pe *periodicEnqueuer) shouldEnqueue() bool {
	conn := pe.pool.Get()
	defer conn.Close()

	lastEnqueue, err := redis.Int64(conn.Do("GET", redisKeyLastPeriodicEnqueue(pe.namespace)))
	if err == redis.ErrNil {
		return true
	} else if err != nil {
		log.Errorf("periodic_enqueuer.should_enqueue:%s\n", err)
		return true
	}

	return lastEnqueue < (nowEpochSeconds() - int64(periodicEnqueuerSleep/time.Minute))
}

var nowMock int64

func nowEpochSeconds() int64 {
	if nowMock != 0 {
		return nowMock
	}
	return time.Now().Unix()
}

func makeUniquePeriodicID(name, spec string, epoch int64) string {
	return fmt.Sprintf("periodic:job:%s:%s:%d", name, spec, epoch)
}

func serializeJob(job *work.Job) ([]byte, error) {
	return json.Marshal(job)
}

func redisNamespacePrefix(namespace string) string {
	l := len(namespace)
	if (l > 0) && (namespace[l-1] != ':') {
		namespace = namespace + ":"
	}
	return namespace
}

func redisKeyScheduled(namespace string) string {
	return redisNamespacePrefix(namespace) + "scheduled"
}

func redisKeyLastPeriodicEnqueue(namespace string) string {
	return redisNamespacePrefix(namespace) + "last_periodic_enqueue"
}
