package exporter

import (
	"fmt"
	"time"

	"github.com/gocraft/work"
	redislib "github.com/goharbor/harbor/src/lib/redis"
	"github.com/gomodule/redigo/redis"
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = 10 * time.Second
	dialWriteTimeout      = 10 * time.Second
)

var (
	redisPool   *redis.Pool
	jsClient    *work.Client
	jsNamespace string
)

// RedisPoolConfig ...
type RedisPoolConfig struct {
	URL               string
	Namespace         string
	IdleTimeoutSecond int
}

// InitBackendWorker initiate backend worker
func InitBackendWorker(redisPoolConfig *RedisPoolConfig) {
	pool, err := redislib.GetRedisPool("JobService", redisPoolConfig.URL, &redislib.PoolParam{
		PoolMaxIdle:           6,
		PoolIdleTimeout:       time.Duration(redisPoolConfig.IdleTimeoutSecond) * time.Second,
		DialConnectionTimeout: dialConnectionTimeout,
		DialReadTimeout:       dialReadTimeout,
		DialWriteTimeout:      dialWriteTimeout,
	})
	if err != nil {
		panic(err)
	}
	redisPool = pool
	jsNamespace = fmt.Sprintf("{%s}", redisPoolConfig.Namespace)
	// Start the backend worker
	jsClient = work.NewClient(jsNamespace, pool)

}

// GetBackendWorker ...
func GetBackendWorker() *work.Client {
	return jsClient
}

// GetRedisPool ...
func GetRedisPool() *redis.Pool {
	return redisPool
}

func redisNamespacePrefix(namespace string) string {
	l := len(namespace)
	if (l > 0) && (namespace[l-1] != ':') {
		namespace = namespace + ":"
	}
	return namespace
}

func redisKeyJobsLockInfo(namespace, jobName string) string {
	return redisNamespacePrefix(namespace) + "jobs:" + jobName + ":lock_info"
}

func redisKeyKnownJobs(namespace string) string {
	return redisNamespacePrefix(namespace) + "known_jobs"
}
