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

package redis

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
)

var (
	// ErrUnLock ...
	ErrUnLock = errors.New("error to release the redis lock")
)

const (
	unlockScript = `
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end
`
)

// Mutex ...
type Mutex struct {
	Conn  redis.Conn
	key   string
	value string
	opts  Options
}

// New ...
func New(conn redis.Conn, key, value string) *Mutex {
	o := *DefaultOptions()
	if value == "" {
		value = utils.GenerateRandomString()
	}
	return &Mutex{conn, key, value, o}
}

// Require retry to require the lock
func (rm *Mutex) Require() (bool, error) {
	var isRequired bool
	var err error

	for i := 0; i < rm.opts.maxRetry; i++ {
		isRequired, err = rm.require()
		if isRequired {
			break
		}
		if err != nil || !isRequired {
			time.Sleep(rm.opts.retryDelay)
		}
	}

	return isRequired, err
}

// require get the redis lock, for details, just refer to https://redis.io/topics/distlock
func (rm *Mutex) require() (bool, error) {
	reply, err := redis.String(rm.Conn.Do("SET", rm.key, rm.value, "NX", "PX", int(rm.opts.expiry/time.Millisecond)))
	if err != nil {
		return false, err
	}
	return reply == "OK", nil
}

// Free releases the lock, for details, just refer to https://redis.io/topics/distlock
func (rm *Mutex) Free() (bool, error) {
	script := redis.NewScript(1, unlockScript)
	resp, err := redis.Int(script.Do(rm.Conn, rm.key, rm.value))
	if err != nil {
		return false, err
	}
	if resp == 0 {
		return false, ErrUnLock
	}
	return true, nil
}

// Options ...
type Options struct {
	retryDelay time.Duration
	expiry     time.Duration
	maxRetry   int
}

var (
	opt     *Options
	optOnce sync.Once

	defaultDelay    = int64(1) // 1 second
	defaultMaxRetry = 600
	defaultExpire   = int64(2 * time.Hour / time.Second) // 2 hours
)

// DefaultOptions ...
func DefaultOptions() *Options {
	optOnce.Do(func() {
		retryDelay, err := strconv.ParseInt(os.Getenv("REDIS_LOCK_RETRY_DELAY"), 10, 64)
		if err != nil || retryDelay < 0 {
			retryDelay = defaultDelay
		}

		maxRetry, err := strconv.Atoi(os.Getenv("REDIS_LOCK_MAX_RETRY"))
		if err != nil || maxRetry < 0 {
			maxRetry = defaultMaxRetry
		}

		expire, err := strconv.ParseInt(os.Getenv("REDIS_LOCK_EXPIRE"), 10, 64)
		if err != nil || expire < 0 {
			expire = defaultExpire
		}

		opt = &Options{
			retryDelay: time.Duration(retryDelay) * time.Second,
			expiry:     time.Duration(expire) * time.Second,
			maxRetry:   maxRetry,
		}
	})

	return opt
}

var (
	pool     *redis.Pool
	poolOnce sync.Once

	poolMaxIdle           = 200
	poolMaxActive         = 1000
	poolIdleTimeout int64 = 180
)

// DefaultPool return default redis pool
func DefaultPool() *redis.Pool {
	poolOnce.Do(func() {
		maxIdle, err := strconv.Atoi(os.Getenv("REDIS_POOL_MAX_IDLE"))
		if err != nil || maxIdle < 0 {
			maxIdle = poolMaxIdle
		}

		maxActive, err := strconv.Atoi(os.Getenv("REDIS_POOL_MAX_ACTIVE"))
		if err != nil || maxActive < 0 {
			maxActive = poolMaxActive
		}

		idleTimeout, err := strconv.ParseInt(os.Getenv("REDIS_POOL_IDLE_TIMEOUT"), 10, 64)
		if err != nil || idleTimeout < 0 {
			idleTimeout = poolIdleTimeout
		}

		pool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				url := config.GetRedisOfRegURL()
				if url == "" {
					url = "redis://localhost:6379/1"
				}

				return redis.DialURL(url)
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
			MaxIdle:     maxIdle,
			MaxActive:   maxActive,
			IdleTimeout: time.Duration(idleTimeout) * time.Second,
			Wait:        true,
		}
	})

	return pool
}

// RequireLock returns lock by key
func RequireLock(key string, conns ...redis.Conn) (*Mutex, error) {
	var conn redis.Conn
	if len(conns) > 0 {
		conn = conns[0]
	} else {
		conn = DefaultPool().Get()
	}

	m := New(conn, key, utils.GenerateRandomString())
	ok, err := m.Require()
	if err != nil {
		return nil, fmt.Errorf("require redis lock failed: %v", err)
	}

	if !ok {
		return nil, fmt.Errorf("unable to require lock for %s", key)
	}

	return m, nil
}

// FreeLock free lock
func FreeLock(m *Mutex) error {
	if _, err := m.Free(); err != nil {
		log.Warningf("failed to free lock %s, error: %v", m.key, err)
		return err
	}

	if err := m.Conn.Close(); err != nil {
		log.Warningf("failed to close the redis con for lock %s, error: %v", m.key, err)
		return err
	}

	return nil
}
