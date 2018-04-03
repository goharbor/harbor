package work

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/garyburd/redigo/redis"
)

type worker struct {
	workerID    string
	poolID      string
	namespace   string
	pool        *redis.Pool
	jobTypes    map[string]*jobType
	middleware  []*middlewareHandler
	contextType reflect.Type

	redisFetchScript *redis.Script
	sampler          prioritySampler
	*observer

	stopChan         chan struct{}
	doneStoppingChan chan struct{}

	drainChan        chan struct{}
	doneDrainingChan chan struct{}
}

func newWorker(namespace string, poolID string, pool *redis.Pool, contextType reflect.Type, middleware []*middlewareHandler, jobTypes map[string]*jobType) *worker {
	workerID := makeIdentifier()
	ob := newObserver(namespace, pool, workerID)

	w := &worker{
		workerID:    workerID,
		poolID:      poolID,
		namespace:   namespace,
		pool:        pool,
		contextType: contextType,

		observer: ob,

		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),

		drainChan:        make(chan struct{}),
		doneDrainingChan: make(chan struct{}),
	}

	w.updateMiddlewareAndJobTypes(middleware, jobTypes)

	return w
}

// note: can't be called while the thing is started
func (w *worker) updateMiddlewareAndJobTypes(middleware []*middlewareHandler, jobTypes map[string]*jobType) {
	w.middleware = middleware
	sampler := prioritySampler{}
	for _, jt := range jobTypes {
		sampler.add(jt.Priority, redisKeyJobs(w.namespace, jt.Name), redisKeyJobsInProgress(w.namespace, w.poolID, jt.Name))
	}
	w.sampler = sampler
	w.jobTypes = jobTypes
	w.redisFetchScript = redis.NewScript(len(jobTypes)*2, redisLuaRpoplpushMultiCmd)
}

func (w *worker) start() {
	go w.loop()
	go w.observer.start()
}

func (w *worker) stop() {
	w.stopChan <- struct{}{}
	<-w.doneStoppingChan
	w.observer.drain()
	w.observer.stop()
}

func (w *worker) drain() {
	w.drainChan <- struct{}{}
	<-w.doneDrainingChan
	w.observer.drain()
}

var sleepBackoffsInMilliseconds = []int64{0, 10, 100, 1000, 5000}

func (w *worker) loop() {
	var drained bool
	var consequtiveNoJobs int64

	// Begin immediately. We'll change the duration on each tick with a timer.Reset()
	timer := time.NewTimer(0)
	defer timer.Stop()

	for {
		select {
		case <-w.stopChan:
			w.doneStoppingChan <- struct{}{}
			return
		case <-w.drainChan:
			drained = true
			timer.Reset(0)
		case <-timer.C:
			gotJob := true
			for gotJob {
				job, err := w.fetchJob()
				if err != nil {
					logError("worker.fetch", err)
					gotJob = false
					timer.Reset(10 * time.Millisecond)
				} else if job != nil {
					w.processJob(job)
					consequtiveNoJobs = 0
				} else {
					gotJob = false
					if drained {
						w.doneDrainingChan <- struct{}{}
						drained = false
					}
					consequtiveNoJobs++
					idx := consequtiveNoJobs
					if idx >= int64(len(sleepBackoffsInMilliseconds)) {
						idx = int64(len(sleepBackoffsInMilliseconds)) - 1
					}
					timer.Reset(time.Duration(sleepBackoffsInMilliseconds[idx]) * time.Millisecond)
				}
			}
		}
	}
}

func (w *worker) fetchJob() (*Job, error) {
	// resort queues
	// NOTE: we could optimize this to only resort every second, or something.
	w.sampler.sample()

	var scriptArgs = make([]interface{}, 0, len(w.sampler.samples)*2)
	for _, s := range w.sampler.samples {
		scriptArgs = append(scriptArgs, s.redisJobs, s.redisJobsInProg)
	}

	conn := w.pool.Get()
	defer conn.Close()
	values, err := redis.Values(w.redisFetchScript.Do(conn, scriptArgs...))
	if err == redis.ErrNil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if len(values) != 3 {
		return nil, fmt.Errorf("need 3 elements back")
	}

	rawJSON, ok := values[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("response msg not bytes")
	}

	dequeuedFrom, ok := values[1].([]byte)
	if !ok {
		return nil, fmt.Errorf("response queue not bytes")
	}

	inProgQueue, ok := values[2].([]byte)
	if !ok {
		return nil, fmt.Errorf("response in prog not bytes")
	}

	job, err := newJob(rawJSON, dequeuedFrom, inProgQueue)
	if err != nil {
		return nil, err
	}

	return job, nil
}

func (w *worker) processJob(job *Job) {
	if job.Unique {
		w.deleteUniqueJob(job)
	}
	if jt, ok := w.jobTypes[job.Name]; ok {
		w.observeStarted(job.Name, job.ID, job.Args)
		job.observer = w.observer // for Checkin
		_, runErr := runJob(job, w.contextType, w.middleware, jt)
		w.observeDone(job.Name, job.ID, runErr)
		if runErr != nil {
			job.failed(runErr)
			w.addToRetryOrDead(jt, job, runErr)
		} else {
			w.removeJobFromInProgress(job)
		}
	} else {
		// NOTE: since we don't have a jobType, we don't know max retries
		runErr := fmt.Errorf("stray job: no handler")
		logError("process_job.stray", runErr)
		job.failed(runErr)
		w.addToDead(job, runErr)
	}
}

func (w *worker) deleteUniqueJob(job *Job) {
	uniqueKey, err := redisKeyUniqueJob(w.namespace, job.Name, job.Args)
	if err != nil {
		logError("worker.delete_unique_job.key", err)
	}
	conn := w.pool.Get()
	defer conn.Close()

	_, err = conn.Do("DEL", uniqueKey)
	if err != nil {
		logError("worker.delete_unique_job.del", err)
	}
}

func (w *worker) removeJobFromInProgress(job *Job) {
	conn := w.pool.Get()
	defer conn.Close()

	_, err := conn.Do("LREM", job.inProgQueue, 1, job.rawJSON)
	if err != nil {
		logError("worker.remove_job_from_in_progress.lrem", err)
	}
}

func (w *worker) addToRetryOrDead(jt *jobType, job *Job, runErr error) {
	failsRemaining := int64(jt.MaxFails) - job.Fails
	if failsRemaining > 0 {
		w.addToRetry(job, runErr)
	} else {
		if !jt.SkipDead {
			w.addToDead(job, runErr)
		}
	}
}

func (w *worker) addToRetry(job *Job, runErr error) {
	rawJSON, err := job.serialize()
	if err != nil {
		logError("worker.add_to_retry", err)
		return
	}

	conn := w.pool.Get()
	defer conn.Close()

	conn.Send("MULTI")
	conn.Send("LREM", job.inProgQueue, 1, job.rawJSON)
	conn.Send("ZADD", redisKeyRetry(w.namespace), nowEpochSeconds()+backoff(job.Fails), rawJSON)
	_, err = conn.Do("EXEC")
	if err != nil {
		logError("worker.add_to_retry.exec", err)
	}

}

func (w *worker) addToDead(job *Job, runErr error) {
	rawJSON, err := job.serialize()

	if err != nil {
		logError("worker.add_to_dead.serialize", err)
		return
	}

	conn := w.pool.Get()
	defer conn.Close()

	// NOTE: sidekiq limits the # of jobs: only keep jobs for 6 months, and only keep a max # of jobs
	// The max # of jobs seems really horrible. Seems like operations should be on top of it.
	// conn.Send("ZREMRANGEBYSCORE", redisKeyDead(w.namespace), "-inf", now - keepInterval)
	// conn.Send("ZREMRANGEBYRANK", redisKeyDead(w.namespace), 0, -maxJobs)

	conn.Send("MULTI")
	conn.Send("LREM", job.inProgQueue, 1, job.rawJSON)
	conn.Send("ZADD", redisKeyDead(w.namespace), nowEpochSeconds(), rawJSON)
	_, err = conn.Do("EXEC")
	if err != nil {
		logError("worker.add_to_dead.exec", err)
	}
}

// backoff returns number of seconds t
func backoff(fails int64) int64 {
	return (fails * fails * fails * fails) + 15 + (rand.Int63n(30) * (fails + 1))
}
