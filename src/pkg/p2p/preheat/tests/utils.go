package tests

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Pool used to create a redis pool
func Pool() *redis.Pool {
	redisPool := &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				"redis://localhost:6379",
				redis.DialConnectTimeout(30*time.Second),
				redis.DialReadTimeout(15*time.Second),
				redis.DialWriteTimeout(15*time.Second),
			)
		},
	}

	return redisPool
}

// Clear the specified redis key
func Clear(pool *redis.Pool, key string) error {
	conn := pool.Get()
	defer conn.Close()

	args := []interface{}{
		fmt.Sprintf("%s:store", key),
		fmt.Sprintf("%s:index", key),
	}

	_, err := conn.Do("DEL", args...)

	return err
}
