// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package tests provide test utilities
package tests

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"

	redislib "github.com/goharbor/harbor/src/lib/redis"
)

const (
	poolMaxIdle           = 6
	PoolMaxActive         = 6
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	testingRedisHostEnv   = "REDIS_HOST"
	testingRedisPort      = 6379
	testingNamespace      = "testing_job_service_v2"
)

// GiveMeRedisPool ...
func GiveMeRedisPool() *redis.Pool {
	pool, _ := redislib.GetRedisPool("test", GetRedisURL(), &redislib.PoolParam{
		PoolMaxIdle:           poolMaxIdle,
		PoolMaxActive:         PoolMaxActive,
		DialConnectionTimeout: dialConnectionTimeout,
		DialReadTimeout:       dialReadTimeout,
		DialWriteTimeout:      dialWriteTimeout,
	})
	return pool
}

// GiveMeTestNamespace ...
func GiveMeTestNamespace() string {
	return testingNamespace
}

// GetRedisURL ...
func GetRedisURL() string {
	redisHost := os.Getenv(testingRedisHostEnv)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return fmt.Sprintf("redis://%s:%d", redisHost, testingRedisPort)
}

// Clear ...
func Clear(key string, conn redis.Conn) error {
	if conn != nil {
		_, err := conn.Do("DEL", key)
		return err
	}

	return errors.New("failed to clear")
}

// ClearAll ...
func ClearAll(namespace string, conn redis.Conn) error {
	keys, err := redis.Strings(conn.Do("KEYS", fmt.Sprintf("%s:*", namespace)))
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		if err := conn.Send("DEL", key); err != nil {
			return err
		}
	}

	return conn.Flush()
}
