package work

import (
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDeadPoolReaper(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"

	conn := pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(ns)

	// Create redis data
	var err error
	cleanKeyspace(ns, pool)
	err = conn.Send("SADD", workerPoolsKey, "1")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "2")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "3")
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "1"),
		"heartbeat_at", time.Now().Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "2"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "3"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)
	err = conn.Flush()
	assert.NoError(t, err)

	// Test getting dead pool
	reaper := newDeadPoolReaper(ns, pool)
	deadPools, err := reaper.findDeadPools()
	assert.NoError(t, err)
	assert.Equal(t, deadPools, map[string][]string{"2": {"type1", "type2"}, "3": {"type1", "type2"}})

	// Test requeueing jobs
	_, err = conn.Do("lpush", redisKeyJobsInProgress(ns, "2", "type1"), "foo")
	assert.NoError(t, err)

	// Ensure 0 jobs in jobs queue
	jobsCount, err := redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Ensure 1 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Reap
	err = reaper.reap()
	assert.NoError(t, err)

	// Ensure 1 jobs in jobs queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Ensure 0 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)
}

func TestDeadPoolReaperNoHeartbeat(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"

	conn := pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(ns)

	// Create redis data
	var err error
	cleanKeyspace(ns, pool)
	err = conn.Send("SADD", workerPoolsKey, "1")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "2")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "3")
	assert.NoError(t, err)
	err = conn.Flush()
	assert.NoError(t, err)

	// Test getting dead pool
	reaper := newDeadPoolReaper(ns, pool)
	deadPools, err := reaper.findDeadPools()
	assert.NoError(t, err)
	assert.Equal(t, deadPools, map[string][]string{})

	// Test requeueing jobs
	_, err = conn.Do("lpush", redisKeyJobsInProgress(ns, "2", "type1"), "foo")
	assert.NoError(t, err)

	// Ensure 0 jobs in jobs queue
	jobsCount, err := redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Ensure 1 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Reap
	err = reaper.reap()
	assert.NoError(t, err)

	// Ensure 0 jobs in jobs queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Ensure 1 job in inprogress queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)
}

func TestDeadPoolReaperNoJobTypes(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"

	conn := pool.Get()
	defer conn.Close()

	workerPoolsKey := redisKeyWorkerPools(ns)

	// Create redis data
	var err error
	cleanKeyspace(ns, pool)
	err = conn.Send("SADD", workerPoolsKey, "1")
	assert.NoError(t, err)
	err = conn.Send("SADD", workerPoolsKey, "2")
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "1"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
	)
	assert.NoError(t, err)

	err = conn.Send("HMSET", redisKeyHeartbeat(ns, "2"),
		"heartbeat_at", time.Now().Add(-1*time.Hour).Unix(),
		"job_names", "type1,type2",
	)
	assert.NoError(t, err)

	err = conn.Flush()
	assert.NoError(t, err)

	// Test getting dead pool
	reaper := newDeadPoolReaper(ns, pool)
	deadPools, err := reaper.findDeadPools()
	assert.NoError(t, err)
	assert.Equal(t, deadPools, map[string][]string{"2": {"type1", "type2"}})

	// Test requeueing jobs
	_, err = conn.Do("lpush", redisKeyJobsInProgress(ns, "1", "type1"), "foo")
	assert.NoError(t, err)
	_, err = conn.Do("lpush", redisKeyJobsInProgress(ns, "2", "type1"), "foo")
	assert.NoError(t, err)

	// Ensure 0 jobs in jobs queue
	jobsCount, err := redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)

	// Ensure 1 job in inprogress queue for each job
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "1", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Reap. Ensure job 2 is requeued but not job 1
	err = reaper.reap()
	assert.NoError(t, err)

	// Ensure 1 jobs in jobs queue
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobs(ns, "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Ensure 1 job in inprogress queue for 1
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "1", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 1, jobsCount)

	// Ensure 0 jobs in inprogress queue for 2
	jobsCount, err = redis.Int(conn.Do("llen", redisKeyJobsInProgress(ns, "2", "type1")))
	assert.NoError(t, err)
	assert.Equal(t, 0, jobsCount)
}
