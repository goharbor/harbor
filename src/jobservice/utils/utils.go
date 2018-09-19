// Copyright Project Harbor Authors. All rights reserved.

// Package utils provides reusable and sharable utilities for other packages and components.
package utils

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

// IsEmptyStr check if the specified str is empty (len ==0) after triming prefix and suffix spaces.
func IsEmptyStr(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

// ReadEnv return the value of env variable.
func ReadEnv(key string) string {
	return os.Getenv(key)
}

// FileExists check if the specified exists.
func FileExists(file string) bool {
	if !IsEmptyStr(file) {
		_, err := os.Stat(file)
		if err == nil {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}

		return true
	}

	return false
}

// DirExists check if the specified dir exists
func DirExists(path string) bool {
	if IsEmptyStr(path) {
		return false
	}

	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

// IsValidPort check if port is valid.
func IsValidPort(port uint) bool {
	return port != 0 && port < 65536
}

// IsValidURL validates if the url is well-formted
func IsValidURL(address string) bool {
	if IsEmptyStr(address) {
		return false
	}

	if _, err := url.Parse(address); err != nil {
		return false
	}

	return true
}

// TranslateRedisAddress translates the comma format to redis URL
func TranslateRedisAddress(commaFormat string) (string, bool) {
	if IsEmptyStr(commaFormat) {
		return "", false
	}

	sections := strings.Split(commaFormat, ",")
	totalSections := len(sections)
	if totalSections == 0 {
		return "", false
	}

	urlParts := []string{}
	// section[0] should be host:port
	redisURL := fmt.Sprintf("redis://%s", sections[0])
	if _, err := url.Parse(redisURL); err != nil {
		return "", false
	}
	urlParts = append(urlParts, "redis://", sections[0])
	// Ignore weight
	// Check password
	if totalSections >= 3 && !IsEmptyStr(sections[2]) {
		urlParts = []string{urlParts[0], fmt.Sprintf("%s:%s@", "arbitrary_username", sections[2]), urlParts[1]}
	}

	if totalSections >= 4 && !IsEmptyStr(sections[3]) {
		if _, err := strconv.Atoi(sections[3]); err == nil {
			urlParts = append(urlParts, "/", sections[3])
		}
	}

	return strings.Join(urlParts, ""), true
}

// JobScore represents the data item with score in the redis db.
type JobScore struct {
	JobBytes []byte
	Score    int64
}

// GetZsetByScore get the items from the zset filtered by the specified score scope.
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
