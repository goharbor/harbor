package work

import (
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

func TestHeartbeater(t *testing.T) {
	pool := newTestPool(":6379")
	ns := "work"

	tMock := int64(1425263409)
	setNowEpochSecondsMock(tMock)
	defer resetNowEpochSecondsMock()

	jobTypes := map[string]*jobType{
		"foo": nil,
		"bar": nil,
	}

	heart := newWorkerPoolHeartbeater(ns, pool, "abcd", jobTypes, 10, []string{"ccc", "bbb"})
	heart.start()

	time.Sleep(20 * time.Millisecond)

	assert.True(t, redisInSet(pool, redisKeyWorkerPools(ns), "abcd"))

	h := readHash(pool, redisKeyHeartbeat(ns, "abcd"))
	assert.Equal(t, "1425263409", h["heartbeat_at"])
	assert.Equal(t, "1425263409", h["started_at"])
	assert.Equal(t, "bar,foo", h["job_names"])
	assert.Equal(t, "bbb,ccc", h["worker_ids"])
	assert.Equal(t, "10", h["concurrency"])

	assert.True(t, h["pid"] != "")
	assert.True(t, h["host"] != "")

	heart.stop()

	assert.False(t, redisInSet(pool, redisKeyWorkerPools(ns), "abcd"))
}

func redisInSet(pool *redis.Pool, key, member string) bool {
	conn := pool.Get()
	defer conn.Close()

	v, err := redis.Bool(conn.Do("SISMEMBER", key, member))
	if err != nil {
		panic("could not delete retry/dead queue: " + err.Error())
	}
	return v
}
