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
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils"
	"time"
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
	defaultDelay    = 5 * time.Second
	defaultMaxRetry = 5
	defaultExpiry   = 600 * time.Second
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

// DefaultOptions ...
func DefaultOptions() *Options {
	opt := &Options{
		retryDelay: defaultDelay,
		expiry:     defaultExpiry,
		maxRetry:   defaultMaxRetry,
	}
	return opt
}
