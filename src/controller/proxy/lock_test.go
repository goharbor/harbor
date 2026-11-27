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
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if len(host) == 0 {
		host = "localhost"
	}
	return redis.NewClient(&redis.Options{Addr: host + ":6379"})
}

func TestRedisLocker(t *testing.T) {
	ctx := context.Background()
	rdb := testRedisClient()
	defer rdb.Close()
	key := manifestListLockKey("library/" + uuid.NewString())
	defer rdb.Del(ctx, key)
	newLocker := func(timeout time.Duration) *redisLocker {
		return &redisLocker{client: func() (*redis.Client, error) { return rdb, nil }, expiration: time.Minute, timeout: timeout}
	}

	_, unlock, err := newLocker(time.Second).lock(ctx, key)
	require.NoError(t, err)

	// another core instance times out while the key is locked
	_, _, err = newLocker(200*time.Millisecond).lock(ctx, key)
	assert.Error(t, err)

	// it locks the key once it is unlocked
	locked := make(chan func(), 1)
	go func() {
		if _, u, err := newLocker(5*time.Second).lock(ctx, key); err == nil {
			locked <- u
		}
	}()
	time.Sleep(300 * time.Millisecond)
	assert.Empty(t, locked)
	unlock()
	select {
	case u := <-locked:
		u()
	case <-time.After(5 * time.Second):
		t.Fatal("the key isn't locked after it is unlocked")
	}

	// unlocking doesn't release the lock taken by another instance after it expired
	_, unlock, err = newLocker(time.Second).lock(ctx, key)
	require.NoError(t, err)
	require.NoError(t, rdb.Set(ctx, key, "another instance", time.Minute).Err())
	unlock()
	val, err := rdb.Get(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, "another instance", val)
}

func TestRedisLockerRenew(t *testing.T) {
	ctx := context.Background()
	rdb := testRedisClient()
	defer rdb.Close()
	key := manifestListLockKey("library/" + uuid.NewString())
	defer rdb.Del(ctx, key)
	l := &redisLocker{client: func() (*redis.Client, error) { return rdb, nil }, expiration: 300 * time.Millisecond, timeout: time.Second}

	// the lock is renewed while it is held
	lockCtx, unlock, err := l.lock(ctx, key)
	require.NoError(t, err)
	time.Sleep(time.Second)
	assert.NoError(t, lockCtx.Err())
	assert.Equal(t, int64(1), rdb.Exists(ctx, key).Val())

	// its context is canceled once it is lost
	require.NoError(t, rdb.Set(ctx, key, "another instance", time.Minute).Err())
	select {
	case <-lockCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("the context isn't canceled when the lock is lost")
	}
	unlock()
	assert.Equal(t, "another instance", rdb.Get(ctx, key).Val())

	// unlocking cancels its context and deletes it
	lockCtx, unlock, err = l.lock(ctx, key+"-2")
	require.NoError(t, err)
	defer rdb.Del(ctx, key+"-2")
	unlock()
	assert.ErrorIs(t, lockCtx.Err(), context.Canceled)
	assert.Equal(t, int64(0), rdb.Exists(ctx, key+"-2").Val())
}
