// Copyright Project Harbor Authors. All rights reserved.

// Package tests provide test utilities
package tests

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	testingRedisHost      = "REDIS_HOST"
	testingNamespace      = "testing_job_service_v2"
)

// GiveMeRedisPool ...
func GiveMeRedisPool() *redis.Pool {
	redisHost := getRedisHost()
	redisPool := &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				fmt.Sprintf("%s:%d", redisHost, 6379),
				redis.DialConnectTimeout(dialConnectionTimeout),
				redis.DialReadTimeout(dialReadTimeout),
				redis.DialWriteTimeout(dialWriteTimeout),
			)
		},
	}

	return redisPool
}

// GiveMeTestNamespace ...
func GiveMeTestNamespace() string {
	return testingNamespace
}

// Clear ...
func Clear(key string, conn redis.Conn) error {
	if conn != nil {
		defer conn.Close()
		_, err := conn.Do("DEL", key)
		return err
	}

	return errors.New("failed to clear")
}

// ClearAll ...
func ClearAll(namespace string, conn redis.Conn) error {
	defer conn.Close()

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

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "10.160.178.186" // for local test
	}

	return redisHost
}
