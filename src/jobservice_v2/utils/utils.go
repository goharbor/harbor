// Copyright 2018 The Harbor Authors. All rights reserved.

//Package utils provides reusable and sharable utilities for other packages and components.
package utils

import (
	"errors"
	"os"
	"strings"

	"github.com/garyburd/redigo/redis"
)

//IsEmptyStr check if the specified str is empty (len ==0) after triming prefix and suffix spaces.
func IsEmptyStr(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

//ReadEnv return the value of env variable.
func ReadEnv(key string) string {
	return os.Getenv(key)
}

//FileExists check if the specified exists.
func FileExists(file string) bool {
	if !IsEmptyStr(file) {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	return false
}

//IsValidPort check if port is valid.
func IsValidPort(port uint) bool {
	return port != 0 && port < 65536
}

//JobScore represents the data item with score in the redis db.
type JobScore struct {
	JobBytes []byte
	Score    int64
}

//GetZsetByScore get the items from the zset filtered by the specified score scope.
func GetZsetByScore(pool *redis.Pool, key string, scores []int64) ([]JobScore, error) {
	if pool == nil || IsEmptyStr(key) || len(scores) < 2 {
		return nil, errors.New("bad arguments")
	}

	conn := pool.Get()
	defer conn.Close()

	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, scores[0], scores[1], "WITHSCORES"))
	if err != nil {
		return nil, err
	}

	var jobsWithScores []JobScore

	if err := redis.ScanSlice(values, &jobsWithScores); err != nil {
		return nil, err
	}

	return jobsWithScores, nil
}
