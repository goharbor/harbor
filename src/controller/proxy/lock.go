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

package proxy

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	libredis "github.com/goharbor/harbor/src/lib/redis"
)

const (
	// lockExpiration frees the lock of a core instance that stops before unlocking it
	lockExpiration = time.Minute
	// lockTimeout is how long to wait for the lock held by another core instance
	lockTimeout       = 2 * time.Minute
	lockRetryInterval = 100 * time.Millisecond
)

// locker locks keys across core instances
type locker interface {
	// lock waits until key is locked. It returns the function that unlocks it, and a context
	// that is canceled when the lock is lost, e.g. when it can't be renewed
	lock(ctx context.Context, key string) (lockCtx context.Context, unlock func(), err error)
}

// renewScript extends the lock only while it is held by the caller
var renewScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0
`)

// unlockScript deletes the lock only while it is held by the caller, not by another instance after it expired
var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

// redisLocker locks keys in the Redis shared by core instances
type redisLocker struct {
	client     func() (*redis.Client, error)
	expiration time.Duration
	timeout    time.Duration
}

func newRedisLocker() *redisLocker {
	return &redisLocker{client: libredis.GetHarborClient, expiration: lockExpiration, timeout: lockTimeout}
}

func (l *redisLocker) lock(ctx context.Context, key string) (context.Context, func(), error) {
	rdb, err := l.client()
	if err != nil {
		return nil, nil, err
	}
	token := uuid.NewString()
	timeout := time.NewTimer(l.timeout)
	defer timeout.Stop()
	for {
		ok, err := rdb.SetNX(ctx, key, token, l.expiration).Result()
		if err != nil {
			return nil, nil, err
		}
		if ok {
			lockCtx, cancel := context.WithCancel(ctx)
			go l.renew(lockCtx, cancel, rdb, key, token)
			return lockCtx, func() {
				cancel()
				if err := unlockScript.Run(context.WithoutCancel(ctx), rdb, []string{key}, token).Err(); err != nil {
					log.Warningf("failed to unlock %v, error: %v", key, err)
				}
			}, nil
		}
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-timeout.C:
			return nil, nil, errors.Errorf("timeout waiting for the lock %v", key)
		case <-time.After(lockRetryInterval):
		}
	}
}

// renew extends the lock held with token until ctx is done, and cancels ctx when the lock is lost
func (l *redisLocker) renew(ctx context.Context, cancel context.CancelFunc, rdb *redis.Client, key, token string) {
	ticker := time.NewTicker(l.expiration / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			renewed, err := renewScript.Run(ctx, rdb, []string{key}, token, l.expiration.Milliseconds()).Int()
			if err != nil || renewed == 0 {
				if ctx.Err() == nil {
					log.Warningf("lost the lock %v, error: %v", key, err)
				}
				cancel()
				return
			}
		}
	}
}
