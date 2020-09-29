package work

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

type requeuer struct {
	namespace string
	pool      *redis.Pool

	redisRequeueScript *redis.Script
	redisRequeueArgs   []interface{}

	stopChan         chan struct{}
	doneStoppingChan chan struct{}

	drainChan        chan struct{}
	doneDrainingChan chan struct{}
}

func newRequeuer(namespace string, pool *redis.Pool, requeueKey string, jobNames []string) *requeuer {
	args := make([]interface{}, 0, len(jobNames)+2+2)
	args = append(args, requeueKey)              // KEY[1]
	args = append(args, redisKeyDead(namespace)) // KEY[2]
	for _, jobName := range jobNames {
		args = append(args, redisKeyJobs(namespace, jobName)) // KEY[3, 4, ...]
	}
	args = append(args, redisKeyJobsPrefix(namespace)) // ARGV[1]
	args = append(args, 0)                             // ARGV[2] -- NOTE: We're going to change this one on every call

	return &requeuer{
		namespace: namespace,
		pool:      pool,

		redisRequeueScript: redis.NewScript(len(jobNames)+2, redisLuaZremLpushCmd),
		redisRequeueArgs:   args,

		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),

		drainChan:        make(chan struct{}),
		doneDrainingChan: make(chan struct{}),
	}
}

func (r *requeuer) start() {
	go r.loop()
}

func (r *requeuer) stop() {
	r.stopChan <- struct{}{}
	<-r.doneStoppingChan
}

func (r *requeuer) drain() {
	r.drainChan <- struct{}{}
	<-r.doneDrainingChan
}

func (r *requeuer) loop() {
	// Just do this simple thing for now.
	// If we have 100 processes all running requeuers,
	// there's probably too much hitting redis.
	// So later on we'l have to implement exponential backoff
	ticker := time.Tick(1000 * time.Millisecond)

	for {
		select {
		case <-r.stopChan:
			r.doneStoppingChan <- struct{}{}
			return
		case <-r.drainChan:
			for r.process() {
			}
			r.doneDrainingChan <- struct{}{}
		case <-ticker:
			for r.process() {
			}
		}
	}
}

func (r *requeuer) process() bool {
	conn := r.pool.Get()
	defer conn.Close()

	r.redisRequeueArgs[len(r.redisRequeueArgs)-1] = nowEpochSeconds()

	res, err := redis.String(r.redisRequeueScript.Do(conn, r.redisRequeueArgs...))
	if err == redis.ErrNil {
		return false
	} else if err != nil {
		logError("requeuer.process", err)
		return false
	}

	if res == "" {
		return false
	} else if res == "dead" {
		logError("requeuer.process.dead", fmt.Errorf("no job name"))
		return true
	} else if res == "ok" {
		return true
	}

	return false
}
