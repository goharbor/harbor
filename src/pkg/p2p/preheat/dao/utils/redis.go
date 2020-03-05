package utils

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RedisPool used to create a redis pool
func RedisPool(addr string) *redis.Pool {
	redisPool := &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				addr,
				redis.DialConnectTimeout(30*time.Second),
				redis.DialReadTimeout(15*time.Second),
				redis.DialWriteTimeout(15*time.Second),
			)
		},
	}

	return redisPool
}

// get redis address
func RedisAddr(rawAddr string) (string, bool) {
	if len(rawAddr) == 0 {
		return "", false
	}

	segments := strings.SplitN(rawAddr, ",", 3)
	if len(segments) <= 1 {
		return "", false
	}

	addrParts := []string{}
	addrParts = append(addrParts, "redis://")
	if len(segments) >= 3 && len(segments[2]) > 0 {
		addrParts = append(addrParts, fmt.Sprintf("%s:%s@", "arbitrary_username", segments[2]))
	}
	addrParts = append(addrParts, segments[0], "/0") // use default db index 0

	//verify
	redisAddr := strings.Join(addrParts, "")
	_, err := url.Parse(redisAddr)
	if err != nil {
		return "", false
	}

	return redisAddr, true
}
