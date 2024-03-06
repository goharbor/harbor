package work

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// ErrNotDeleted is returned by functions that delete jobs to indicate that although the redis commands were successful,
// no object was actually deleted by those commmands.
var ErrNotDeleted = fmt.Errorf("nothing deleted")

// ErrNotRetried is returned by functions that retry jobs to indicate that although the redis commands were successful,
// no object was actually retried by those commmands.
var ErrNotRetried = fmt.Errorf("nothing retried")

// Client implements all of the functionality of the web UI. It can be used to inspect the status of a running cluster and retry dead jobs.
type Client struct {
	namespace string
	pool      *redis.Pool
}

// NewClient creates a new Client with the specified redis namespace and connection pool.
func NewClient(namespace string, pool *redis.Pool) *Client {
	return &Client{
		namespace: namespace,
		pool:      pool,
	}
}

// WorkerPoolHeartbeat represents the heartbeat from a worker pool. WorkerPool's write a heartbeat every 5 seconds so we know they're alive and includes config information.
type WorkerPoolHeartbeat struct {
	WorkerPoolID string   `json:"worker_pool_id"`
	StartedAt    int64    `json:"started_at"`
	HeartbeatAt  int64    `json:"heartbeat_at"`
	JobNames     []string `json:"job_names"`
	Concurrency  uint     `json:"concurrency"`
	Host         string   `json:"host"`
	Pid          int      `json:"pid"`
	WorkerIDs    []string `json:"worker_ids"`
}

// WorkerPoolHeartbeats queries Redis and returns all WorkerPoolHeartbeat's it finds (even for those worker pools which don't have a current heartbeat).
func (c *Client) WorkerPoolHeartbeats() ([]*WorkerPoolHeartbeat, error) {
	conn := c.pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(c.namespace)

	workerPoolIDs, err := redis.Strings(conn.Do("SMEMBERS", workerPoolsKey))
	if err != nil {
		return nil, err
	}
	sort.Strings(workerPoolIDs)

	for _, wpid := range workerPoolIDs {
		key := redisKeyHeartbeat(c.namespace, wpid)
		conn.Send("HGETALL", key)
	}

	if err := conn.Flush(); err != nil {
		logError("worker_pool_statuses.flush", err)
		return nil, err
	}

	heartbeats := make([]*WorkerPoolHeartbeat, 0, len(workerPoolIDs))

	for _, wpid := range workerPoolIDs {
		vals, err := redis.Strings(conn.Receive())
		if err != nil {
			logError("worker_pool_statuses.receive", err)
			return nil, err
		}

		heartbeat := &WorkerPoolHeartbeat{
			WorkerPoolID: wpid,
		}

		for i := 0; i < len(vals)-1; i += 2 {
			key := vals[i]
			value := vals[i+1]

			var err error
			if key == "heartbeat_at" {
				heartbeat.HeartbeatAt, err = strconv.ParseInt(value, 10, 64)
			} else if key == "started_at" {
				heartbeat.StartedAt, err = strconv.ParseInt(value, 10, 64)
			} else if key == "job_names" {
				heartbeat.JobNames = strings.Split(value, ",")
				sort.Strings(heartbeat.JobNames)
			} else if key == "concurrency" {
				var vv uint64
				vv, err = strconv.ParseUint(value, 10, 0)
				heartbeat.Concurrency = uint(vv)
			} else if key == "host" {
				heartbeat.Host = value
			} else if key == "pid" {
				var vv int64
				vv, err = strconv.ParseInt(value, 10, 0)
				heartbeat.Pid = int(vv)
			} else if key == "worker_ids" {
				heartbeat.WorkerIDs = strings.Split(value, ",")
				sort.Strings(heartbeat.WorkerIDs)
			}
			if err != nil {
				logError("worker_pool_statuses.parse", err)
				return nil, err
			}
		}

		heartbeats = append(heartbeats, heartbeat)
	}

	return heartbeats, nil
}

// WorkerObservation represents the latest observation taken from a worker. The observation indicates whether the worker is busy processing a job, and if so, information about that job.
type WorkerObservation struct {
	WorkerID string `json:"worker_id"`
	IsBusy   bool   `json:"is_busy"`

	// If IsBusy:
	JobName   string `json:"job_name"`
	JobID     string `json:"job_id"`
	StartedAt int64  `json:"started_at"`
	ArgsJSON  string `json:"args_json"`
	Checkin   string `json:"checkin"`
	CheckinAt int64  `json:"checkin_at"`
}

// WorkerObservations returns all of the WorkerObservation's it finds for all worker pools' workers.
func (c *Client) WorkerObservations() ([]*WorkerObservation, error) {
	conn := c.pool.Get()
	defer conn.Close()

	hbs, err := c.WorkerPoolHeartbeats()
	if err != nil {
		logError("worker_observations.worker_pool_heartbeats", err)
		return nil, err
	}

	var workerIDs []string
	for _, hb := range hbs {
		workerIDs = append(workerIDs, hb.WorkerIDs...)
	}

	for _, wid := range workerIDs {
		key := redisKeyWorkerObservation(c.namespace, wid)
		conn.Send("HGETALL", key)
	}

	if err := conn.Flush(); err != nil {
		logError("worker_observations.flush", err)
		return nil, err
	}

	observations := make([]*WorkerObservation, 0, len(workerIDs))

	for _, wid := range workerIDs {
		vals, err := redis.Strings(conn.Receive())
		if err != nil {
			logError("worker_observations.receive", err)
			return nil, err
		}

		ob := &WorkerObservation{
			WorkerID: wid,
		}

		for i := 0; i < len(vals)-1; i += 2 {
			key := vals[i]
			value := vals[i+1]

			ob.IsBusy = true

			var err error
			if key == "job_name" {
				ob.JobName = value
			} else if key == "job_id" {
				ob.JobID = value
			} else if key == "started_at" {
				ob.StartedAt, err = strconv.ParseInt(value, 10, 64)
			} else if key == "args" {
				ob.ArgsJSON = value
			} else if key == "checkin" {
				ob.Checkin = value
			} else if key == "checkin_at" {
				ob.CheckinAt, err = strconv.ParseInt(value, 10, 64)
			}
			if err != nil {
				logError("worker_observations.parse", err)
				return nil, err
			}
		}

		observations = append(observations, ob)
	}

	return observations, nil
}

// Queue represents a queue that holds jobs with the same name. It indicates their name, count, and latency (in seconds). Latency is a measurement of how long ago the next job to be processed was enqueued.
type Queue struct {
	JobName string `json:"job_name"`
	Count   int64  `json:"count"`
	Latency int64  `json:"latency"`
}

// Queues returns the Queue's it finds.
func (c *Client) Queues() ([]*Queue, error) {
	conn := c.pool.Get()
	defer conn.Close()

	key := redisKeyKnownJobs(c.namespace)
	jobNames, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		return nil, err
	}
	sort.Strings(jobNames)

	for _, jobName := range jobNames {
		conn.Send("LLEN", redisKeyJobs(c.namespace, jobName))
	}

	if err := conn.Flush(); err != nil {
		logError("client.queues.flush", err)
		return nil, err
	}

	queues := make([]*Queue, 0, len(jobNames))

	for _, jobName := range jobNames {
		count, err := redis.Int64(conn.Receive())
		if err != nil {
			logError("client.queues.receive", err)
			return nil, err
		}

		queue := &Queue{
			JobName: jobName,
			Count:   count,
		}

		queues = append(queues, queue)
	}

	for _, s := range queues {
		if s.Count > 0 {
			conn.Send("LINDEX", redisKeyJobs(c.namespace, s.JobName), -1)
		}
	}

	if err := conn.Flush(); err != nil {
		logError("client.queues.flush2", err)
		return nil, err
	}

	now := nowEpochSeconds()

	for _, s := range queues {
		if s.Count > 0 {
			b, err := redis.Bytes(conn.Receive())
			if err != nil {
				logError("client.queues.receive2", err)
				return nil, err
			}

			job, err := newJob(b, nil, nil)
			if err != nil {
				logError("client.queues.new_job", err)
			}
			s.Latency = now - job.EnqueuedAt
		}
	}

	return queues, nil
}

// RetryJob represents a job in the retry queue.
type RetryJob struct {
	RetryAt int64 `json:"retry_at"`
	*Job
}

// ScheduledJob represents a job in the scheduled queue.
type ScheduledJob struct {
	RunAt int64 `json:"run_at"`
	*Job
}

// DeadJob represents a job in the dead queue.
type DeadJob struct {
	DiedAt int64 `json:"died_at"`
	*Job
}

// ScheduledJobs returns a list of ScheduledJob's. The page param is 1-based; each page is 20 items. The total number of items (not pages) in the list of scheduled jobs is also returned.
func (c *Client) ScheduledJobs(page uint) ([]*ScheduledJob, int64, error) {
	key := redisKeyScheduled(c.namespace)
	jobsWithScores, count, err := c.getZsetPage(key, page)
	if err != nil {
		logError("client.scheduled_jobs.get_zset_page", err)
		return nil, 0, err
	}

	jobs := make([]*ScheduledJob, 0, len(jobsWithScores))

	for _, jws := range jobsWithScores {
		jobs = append(jobs, &ScheduledJob{RunAt: jws.Score, Job: jws.job})
	}

	return jobs, count, nil
}

// RetryJobs returns a list of RetryJob's. The page param is 1-based; each page is 20 items. The total number of items (not pages) in the list of retry jobs is also returned.
func (c *Client) RetryJobs(page uint) ([]*RetryJob, int64, error) {
	key := redisKeyRetry(c.namespace)
	jobsWithScores, count, err := c.getZsetPage(key, page)
	if err != nil {
		logError("client.retry_jobs.get_zset_page", err)
		return nil, 0, err
	}

	jobs := make([]*RetryJob, 0, len(jobsWithScores))

	for _, jws := range jobsWithScores {
		jobs = append(jobs, &RetryJob{RetryAt: jws.Score, Job: jws.job})
	}

	return jobs, count, nil
}

// DeadJobs returns a list of DeadJob's. The page param is 1-based; each page is 20 items. The total number of items (not pages) in the list of dead jobs is also returned.
func (c *Client) DeadJobs(page uint) ([]*DeadJob, int64, error) {
	key := redisKeyDead(c.namespace)
	jobsWithScores, count, err := c.getZsetPage(key, page)
	if err != nil {
		logError("client.dead_jobs.get_zset_page", err)
		return nil, 0, err
	}

	jobs := make([]*DeadJob, 0, len(jobsWithScores))

	for _, jws := range jobsWithScores {
		jobs = append(jobs, &DeadJob{DiedAt: jws.Score, Job: jws.job})
	}

	return jobs, count, nil
}

// DeleteDeadJob deletes a dead job from Redis.
func (c *Client) DeleteDeadJob(diedAt int64, jobID string) error {
	ok, _, err := c.deleteZsetJob(redisKeyDead(c.namespace), diedAt, jobID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotDeleted
	}
	return nil
}

// RetryDeadJob retries a dead job. The job will be re-queued on the normal work queue for eventual processing by a worker.
func (c *Client) RetryDeadJob(diedAt int64, jobID string) error {
	// Get queues for job names
	queues, err := c.Queues()
	if err != nil {
		logError("client.retry_all_dead_jobs.queues", err)
		return err
	}

	// Extract job names
	var jobNames []string
	for _, q := range queues {
		jobNames = append(jobNames, q.JobName)
	}

	script := redis.NewScript(len(jobNames)+1, redisLuaRequeueSingleDeadCmd)

	args := make([]interface{}, 0, len(jobNames)+1+3)
	args = append(args, redisKeyDead(c.namespace)) // KEY[1]
	for _, jobName := range jobNames {
		args = append(args, redisKeyJobs(c.namespace, jobName)) // KEY[2, 3, ...]
	}
	args = append(args, redisKeyJobsPrefix(c.namespace)) // ARGV[1]
	args = append(args, nowEpochSeconds())
	args = append(args, diedAt)
	args = append(args, jobID)

	conn := c.pool.Get()
	defer conn.Close()

	cnt, err := redis.Int64(script.Do(conn, args...))
	if err != nil {
		logError("client.retry_dead_job.do", err)
		return err
	}

	if cnt == 0 {
		return ErrNotRetried
	}

	return nil
}

// RetryAllDeadJobs requeues all dead jobs. In other words, it puts them all back on the normal work queue for workers to pull from and process.
func (c *Client) RetryAllDeadJobs() error {
	// Get queues for job names
	queues, err := c.Queues()
	if err != nil {
		logError("client.retry_all_dead_jobs.queues", err)
		return err
	}

	// Extract job names
	var jobNames []string
	for _, q := range queues {
		jobNames = append(jobNames, q.JobName)
	}

	script := redis.NewScript(len(jobNames)+1, redisLuaRequeueAllDeadCmd)

	args := make([]interface{}, 0, len(jobNames)+1+3)
	args = append(args, redisKeyDead(c.namespace)) // KEY[1]
	for _, jobName := range jobNames {
		args = append(args, redisKeyJobs(c.namespace, jobName)) // KEY[2, 3, ...]
	}
	args = append(args, redisKeyJobsPrefix(c.namespace)) // ARGV[1]
	args = append(args, nowEpochSeconds())
	args = append(args, 1000)

	conn := c.pool.Get()
	defer conn.Close()

	// Cap iterations for safety (which could reprocess 1k*1k jobs).
	// This is conceptually an infinite loop but let's be careful.
	for i := 0; i < 1000; i++ {
		res, err := redis.Int64(script.Do(conn, args...))
		if err != nil {
			logError("client.retry_all_dead_jobs.do", err)
			return err
		}

		if res == 0 {
			break
		}
	}

	return nil
}

// DeleteAllDeadJobs deletes all dead jobs.
func (c *Client) DeleteAllDeadJobs() error {
	conn := c.pool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", redisKeyDead(c.namespace))
	if err != nil {
		logError("client.delete_all_dead_jobs", err)
		return err
	}

	return nil
}

// DeleteScheduledJob deletes a job in the scheduled queue.
func (c *Client) DeleteScheduledJob(scheduledFor int64, jobID string) error {
	ok, jobBytes, err := c.deleteZsetJob(redisKeyScheduled(c.namespace), scheduledFor, jobID)
	if err != nil {
		return err
	}

	// If we get a job back, parse it and see if it's a unique job. If it is, we need to delete the unique key.
	if len(jobBytes) > 0 {
		job, err := newJob(jobBytes, nil, nil)
		if err != nil {
			logError("client.delete_scheduled_job.new_job", err)
			return err
		}

		if job.Unique {
			uniqueKey, err := redisKeyUniqueJob(c.namespace, job.Name, job.Args)
			if err != nil {
				logError("client.delete_scheduled_job.redis_key_unique_job", err)
				return err
			}
			conn := c.pool.Get()
			defer conn.Close()

			_, err = conn.Do("DEL", uniqueKey)
			if err != nil {
				logError("worker.delete_unique_job.del", err)
				return err
			}
		}
	}

	if !ok {
		return ErrNotDeleted
	}
	return nil
}

// DeleteRetryJob deletes a job in the retry queue.
func (c *Client) DeleteRetryJob(retryAt int64, jobID string) error {
	ok, _, err := c.deleteZsetJob(redisKeyRetry(c.namespace), retryAt, jobID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotDeleted
	}
	return nil
}

// deleteZsetJob deletes the job in the specified zset (dead, retry, or scheduled queue). zsetKey is like "work:dead" or "work:scheduled". The function deletes all jobs with the given jobID with the specified zscore (there should only be one, but in theory there could be bad data). It will return if at least one job is deleted and if
func (c *Client) deleteZsetJob(zsetKey string, zscore int64, jobID string) (bool, []byte, error) {
	script := redis.NewScript(1, redisLuaDeleteSingleCmd)

	args := make([]interface{}, 0, 1+2)
	args = append(args, zsetKey) // KEY[1]
	args = append(args, zscore)  // ARGV[1]
	args = append(args, jobID)   // ARGV[2]

	conn := c.pool.Get()
	defer conn.Close()
	values, err := redis.Values(script.Do(conn, args...))
	if len(values) != 2 {
		return false, nil, fmt.Errorf("need 2 elements back from redis command")
	}

	cnt, err := redis.Int64(values[0], err)
	jobBytes, err := redis.Bytes(values[1], err)
	if err != nil {
		logError("client.delete_zset_job.do", err)
		return false, nil, err
	}

	return cnt > 0, jobBytes, nil
}

type jobScore struct {
	JobBytes []byte
	Score    int64
	job      *Job
}

func (c *Client) getZsetPage(key string, page uint) ([]jobScore, int64, error) {
	conn := c.pool.Get()
	defer conn.Close()

	if page == 0 {
		page = 1
	}

	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, "-inf", "+inf", "WITHSCORES", "LIMIT", (page-1)*20, 20))
	if err != nil {
		logError("client.get_zset_page.values", err)
		return nil, 0, err
	}

	var jobsWithScores []jobScore

	if err := redis.ScanSlice(values, &jobsWithScores); err != nil {
		logError("client.get_zset_page.scan_slice", err)
		return nil, 0, err
	}

	for i, jws := range jobsWithScores {
		job, err := newJob(jws.JobBytes, nil, nil)
		if err != nil {
			logError("client.get_zset_page.new_job", err)
			return nil, 0, err
		}

		jobsWithScores[i].job = job
	}

	count, err := redis.Int64(conn.Do("ZCARD", key))
	if err != nil {
		logError("client.get_zset_page.int64", err)
		return nil, 0, err
	}

	return jobsWithScores, count, nil
}
