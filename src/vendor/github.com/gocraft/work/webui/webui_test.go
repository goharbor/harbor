package webui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestWebUIStartStop(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	s := NewServer(ns, pool, ":6666")
	s.Start()
	s.Stop()
}

type TestContext struct{}

func TestWebUIQueues(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	// Get some stuff to to show up in the jobs:
	enqueuer := work.NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue("wat", nil)
	assert.NoError(t, err)
	enqueuer.Enqueue("foo", nil)
	enqueuer.Enqueue("zaz", nil)

	// Start a pool to work on it. It's going to work on the queues
	// side effect of that is knowing which jobs are avail
	wp := work.NewWorkerPool(TestContext{}, 10, ns, pool)
	wp.Job("wat", func(job *work.Job) error {
		return nil
	})
	wp.Job("foo", func(job *work.Job) error {
		return nil
	})
	wp.Job("zaz", func(job *work.Job) error {
		return nil
	})
	wp.Start()
	time.Sleep(20 * time.Millisecond)
	wp.Stop()

	// Now that we have the jobs, populate some queues
	enqueuer.Enqueue("wat", nil)
	enqueuer.Enqueue("wat", nil)
	enqueuer.Enqueue("wat", nil)
	enqueuer.Enqueue("foo", nil)
	enqueuer.Enqueue("foo", nil)
	enqueuer.Enqueue("zaz", nil)

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/queues", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	var res []interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.Equal(t, 3, len(res))

	foomap, ok := res[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "foo", foomap["job_name"])
	assert.EqualValues(t, 2, foomap["count"])
	assert.EqualValues(t, 0, foomap["latency"])
}

func TestWebUIWorkerPools(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	wp := work.NewWorkerPool(TestContext{}, 10, ns, pool)
	wp.Job("wat", func(job *work.Job) error { return nil })
	wp.Job("bob", func(job *work.Job) error { return nil })
	wp.Start()
	defer wp.Stop()

	wp2 := work.NewWorkerPool(TestContext{}, 11, ns, pool)
	wp2.Job("foo", func(job *work.Job) error { return nil })
	wp2.Job("bar", func(job *work.Job) error { return nil })
	wp2.Start()
	defer wp2.Stop()

	time.Sleep(20 * time.Millisecond)

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/worker_pools", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	var res []interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(res))

	w1stat, ok := res[0].(map[string]interface{})
	assert.True(t, ok)
	assert.True(t, w1stat["worker_pool_id"] != "")
	// NOTE: WorkerPoolStatus is tested elsewhere.
}

func TestWebUIBusyWorkers(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	// Keep a job in the in-progress state without using sleeps
	wgroup := sync.WaitGroup{}
	wgroup2 := sync.WaitGroup{}
	wgroup2.Add(1)

	wp := work.NewWorkerPool(TestContext{}, 10, ns, pool)
	wp.Job("wat", func(job *work.Job) error {
		wgroup2.Done()
		wgroup.Wait()
		return nil
	})
	wp.Start()
	defer wp.Stop()

	wp2 := work.NewWorkerPool(TestContext{}, 11, ns, pool)
	wp2.Start()
	defer wp2.Stop()

	time.Sleep(10 * time.Millisecond)

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/busy_workers", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	var res []interface{}
	err := json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))

	wgroup.Add(1)

	// Ok, now let's make a busy worker
	enqueuer := work.NewEnqueuer(ns, pool)
	enqueuer.Enqueue("wat", nil)
	wgroup2.Wait()
	time.Sleep(5 * time.Millisecond) // need to let obsever process

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/busy_workers", nil)
	s.router.ServeHTTP(recorder, request)
	wgroup.Done()
	assert.Equal(t, 200, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	if len(res) == 1 {
		hash, ok := res[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "wat", hash["job_name"])
		assert.Equal(t, true, hash["is_busy"])
	}
}

func TestWebUIRetryJobs(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"
	cleanKeyspace(ns, pool)

	enqueuer := work.NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue("wat", nil)
	assert.Nil(t, err)

	wp := work.NewWorkerPool(TestContext{}, 2, ns, pool)
	wp.Job("wat", func(job *work.Job) error {
		return fmt.Errorf("ohno")
	})
	wp.Start()
	wp.Drain()
	wp.Stop()

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/retry_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var res struct {
		Count int64 `json:"count"`
		Jobs  []struct {
			RetryAt int64  `json:"retry_at"`
			Name    string `json:"name"`
			Fails   int64  `json:"fails"`
		} `json:"jobs"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, res.Count)
	assert.Equal(t, 1, len(res.Jobs))
	if len(res.Jobs) == 1 {
		assert.True(t, res.Jobs[0].RetryAt > 0)
		assert.Equal(t, "wat", res.Jobs[0].Name)
		assert.EqualValues(t, 1, res.Jobs[0].Fails)
	}
}

func TestWebUIScheduledJobs(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "testwork"
	cleanKeyspace(ns, pool)

	enqueuer := work.NewEnqueuer(ns, pool)
	_, err := enqueuer.EnqueueIn("watter", 1, nil)
	assert.Nil(t, err)

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/scheduled_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var res struct {
		Count int64 `json:"count"`
		Jobs  []struct {
			RunAt int64  `json:"run_at"`
			Name  string `json:"name"`
		} `json:"jobs"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.EqualValues(t, 1, res.Count)
	assert.Equal(t, 1, len(res.Jobs))
	if len(res.Jobs) == 1 {
		assert.True(t, res.Jobs[0].RunAt > 0)
		assert.Equal(t, "watter", res.Jobs[0].Name)
	}
}

func TestWebUIDeadJobs(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "testwork"
	cleanKeyspace(ns, pool)

	enqueuer := work.NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue("wat", nil)
	_, err = enqueuer.Enqueue("wat", nil)
	assert.Nil(t, err)

	wp := work.NewWorkerPool(TestContext{}, 2, ns, pool)
	wp.JobWithOptions("wat", work.JobOptions{Priority: 1, MaxFails: 1}, func(job *work.Job) error {
		return fmt.Errorf("ohno")
	})
	wp.Start()
	wp.Drain()
	wp.Stop()

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var res struct {
		Count int64 `json:"count"`
		Jobs  []struct {
			DiedAt int64  `json:"died_at"`
			Name   string `json:"name"`
			ID     string `json:"id"`
			Fails  int64  `json:"fails"`
		} `json:"jobs"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.EqualValues(t, 2, res.Count)
	assert.Equal(t, 2, len(res.Jobs))
	var diedAt0, diedAt1 int64
	var id0, id1 string
	if len(res.Jobs) == 2 {
		assert.True(t, res.Jobs[0].DiedAt > 0)
		assert.Equal(t, "wat", res.Jobs[0].Name)
		assert.EqualValues(t, 1, res.Jobs[0].Fails)

		diedAt0, diedAt1 = res.Jobs[0].DiedAt, res.Jobs[1].DiedAt
		id0, id1 = res.Jobs[0].ID, res.Jobs[1].ID
	} else {
		return
	}

	// Ok, now let's retry one and delete one.
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("POST", fmt.Sprintf("/delete_dead_job/%d/%s", diedAt0, id0), nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("POST", fmt.Sprintf("/retry_dead_job/%d/%s", diedAt1, id1), nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	// Make sure dead queue is empty
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, res.Count)

	// Make sure the "wat" queue has 1 item in it
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/queues", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var queueRes []struct {
		JobName string `json:"job_name"`
		Count   int64  `json:"count"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &queueRes)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(queueRes))
	if len(queueRes) == 1 {
		assert.Equal(t, "wat", queueRes[0].JobName)
	}
}

func TestWebUIDeadJobsDeleteRetryAll(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "testwork"
	cleanKeyspace(ns, pool)

	enqueuer := work.NewEnqueuer(ns, pool)
	_, err := enqueuer.Enqueue("wat", nil)
	_, err = enqueuer.Enqueue("wat", nil)
	assert.Nil(t, err)

	wp := work.NewWorkerPool(TestContext{}, 2, ns, pool)
	wp.JobWithOptions("wat", work.JobOptions{Priority: 1, MaxFails: 1}, func(job *work.Job) error {
		return fmt.Errorf("ohno")
	})
	wp.Start()
	wp.Drain()
	wp.Stop()

	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var res struct {
		Count int64 `json:"count"`
		Jobs  []struct {
			DiedAt int64  `json:"died_at"`
			Name   string `json:"name"`
			ID     string `json:"id"`
			Fails  int64  `json:"fails"`
		} `json:"jobs"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)

	assert.EqualValues(t, 2, res.Count)
	assert.Equal(t, 2, len(res.Jobs))

	// Ok, now let's retry all
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("POST", "/retry_all_dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	// Make sure dead queue is empty
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, res.Count)

	// Make sure the "wat" queue has 2 items in it
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/queues", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	var queueRes []struct {
		JobName string `json:"job_name"`
		Count   int64  `json:"count"`
	}
	err = json.Unmarshal(recorder.Body.Bytes(), &queueRes)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(queueRes))
	if len(queueRes) == 1 {
		assert.Equal(t, "wat", queueRes[0].JobName)
		assert.EqualValues(t, 2, queueRes[0].Count)
	}

	// Make them dead again:
	wp.Start()
	wp.Drain()
	wp.Stop()

	// Make sure we have 2 dead things again:
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.EqualValues(t, 2, res.Count)

	// Now delete them:
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("POST", "/delete_all_dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)

	// Make sure dead queue is empty
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/dead_jobs", nil)
	s.router.ServeHTTP(recorder, request)
	assert.Equal(t, 200, recorder.Code)
	err = json.Unmarshal(recorder.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.EqualValues(t, 0, res.Count)
}

func TestWebUIAssets(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "testwork"
	s := NewServer(ns, pool, ":6666")

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/", nil)
	s.router.ServeHTTP(recorder, request)
	body := string(recorder.Body.Bytes())
	assert.Regexp(t, "html", body)

	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("GET", "/work.js", nil)
	s.router.ServeHTTP(recorder, request)
}

func newTestPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   3,
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		},
		Wait: true,
	}
}

func cleanKeyspace(namespace string, pool *redis.Pool) {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", namespace+"*"))
	if err != nil {
		panic("could not get keys: " + err.Error())
	}
	for _, k := range keys {
		if _, err := conn.Do("DEL", k); err != nil {
			panic("could not del: " + err.Error())
		}
	}
}
