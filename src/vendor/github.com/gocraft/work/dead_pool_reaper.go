package work

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	deadTime   = 5 * time.Minute
	reapPeriod = 10 * time.Minute
)

type deadPoolReaper struct {
	namespace        string
	pool             *redis.Pool
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
}

func newDeadPoolReaper(namespace string, pool *redis.Pool) *deadPoolReaper {
	return &deadPoolReaper{
		namespace:        namespace,
		pool:             pool,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}
}

func (r *deadPoolReaper) start() {
	go r.loop()
}

func (r *deadPoolReaper) stop() {
	r.stopChan <- struct{}{}
	<-r.doneStoppingChan
}

func (r *deadPoolReaper) loop() {
	// Reap
	if err := r.reap(); err != nil {
		logError("dead_pool_reaper.reap", err)
	}

	// Begin reaping periodically
	timer := time.NewTimer(reapPeriod)
	defer timer.Stop()

	for {
		select {
		case <-r.stopChan:
			r.doneStoppingChan <- struct{}{}
			return
		case <-timer.C:
			// Schedule next occurrence with jitter
			timer.Reset(reapPeriod + time.Duration(rand.Intn(30))*time.Second)

			// Reap
			if err := r.reap(); err != nil {
				logError("dead_pool_reaper.reap", err)
			}
		}
	}
}

func (r *deadPoolReaper) reap() error {
	// Get dead pools
	deadPoolIDs, err := r.findDeadPools()
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(r.namespace)

	// Cleanup all dead pools
	for deadPoolID, jobTypes := range deadPoolIDs {
		// Requeue all dangling jobs
		r.requeueInProgressJobs(deadPoolID, jobTypes)

		// Remove hearbeat
		_, err = conn.Do("DEL", redisKeyHeartbeat(r.namespace, deadPoolID))
		if err != nil {
			return err
		}

		// Remove from set
		_, err = conn.Do("SREM", workerPoolsKey, deadPoolID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *deadPoolReaper) requeueInProgressJobs(poolID string, jobTypes []string) error {
	redisRequeueScript := redis.NewScript(len(jobTypes)*2, redisLuaRpoplpushMultiCmd)

	var scriptArgs = make([]interface{}, 0, len(jobTypes)*2)
	for _, jobType := range jobTypes {
		scriptArgs = append(scriptArgs, redisKeyJobsInProgress(r.namespace, poolID, jobType), redisKeyJobs(r.namespace, jobType))
	}

	conn := r.pool.Get()
	defer conn.Close()

	// Keep moving jobs until all queues are empty
	for {
		values, err := redis.Values(redisRequeueScript.Do(conn, scriptArgs...))
		if err == redis.ErrNil {
			return nil
		} else if err != nil {
			return err
		}

		if len(values) != 3 {
			return fmt.Errorf("need 3 elements back")
		}
	}
}

func (r *deadPoolReaper) findDeadPools() (map[string][]string, error) {
	conn := r.pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(r.namespace)

	workerPoolIDs, err := redis.Strings(conn.Do("SMEMBERS", workerPoolsKey))
	if err != nil {
		return nil, err
	}

	deadPools := map[string][]string{}
	for _, workerPoolID := range workerPoolIDs {
		heartbeatKey := redisKeyHeartbeat(r.namespace, workerPoolID)

		// Check that last heartbeat was long enough ago to consider the pool dead
		heartbeatAt, err := redis.Int64(conn.Do("HGET", heartbeatKey, "heartbeat_at"))
		if err == redis.ErrNil {
			continue
		}
		if err != nil {
			return nil, err
		}

		if time.Unix(heartbeatAt, 0).Add(deadTime).After(time.Now()) {
			continue
		}

		jobTypesList, err := redis.String(conn.Do("HGET", heartbeatKey, "job_names"))
		if err == redis.ErrNil {
			continue
		}
		if err != nil {
			return nil, err
		}

		deadPools[workerPoolID] = strings.Split(jobTypesList, ",")
	}

	return deadPools, nil
}
