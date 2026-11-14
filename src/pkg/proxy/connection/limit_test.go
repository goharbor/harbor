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

package connection

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnLimiter_Acquire_Release(t *testing.T) {
	// Given
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_max_connection_key"
	defer rdb.Del(ctx, key)
	maxConn := 10

	// When
	tokens := make([]string, 0, maxConn)
	for range maxConn {
		token, ok := Limiter.Acquire(ctx, rdb, key, maxConn)
		require.True(t, ok)
		tokens = append(tokens, token)
	}

	// Then no further connection fits until one is released
	_, ok := Limiter.Acquire(ctx, rdb, key, maxConn)
	assert.False(t, ok)
	for _, token := range tokens {
		Limiter.Release(ctx, rdb, key, token)
	}
	n, err := rdb.ZCard(ctx, key).Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), n)
}

func TestConnLimiter_Acquire_ExpiresAfterSlotTTL(t *testing.T) {
	// Given
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_acquire_ttl_key"
	defer rdb.Del(ctx, key)

	// When
	_, ok := Limiter.Acquire(ctx, rdb, key, 1)

	// Then
	require.True(t, ok)
	ttl, err := rdb.TTL(ctx, key).Result()
	assert.Nil(t, err)
	assert.True(t, ttl > 0 && ttl <= SlotTTL, "a held connection expires after SlotTTL unless refreshed, got %v", ttl)
}

func TestConnLimiter_Acquire_IgnoresConnectionsWhoseHolderStoppedRefreshing(t *testing.T) {
	// Given two held connections, one of them expired while the other keeps
	// refreshing
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_expired_holder_key"
	defer rdb.Del(ctx, key)
	dead, ok := Limiter.Acquire(ctx, rdb, key, 2)
	require.True(t, ok)
	alive, ok := Limiter.Acquire(ctx, rdb, key, 2)
	require.True(t, ok)
	expire(ctx, rdb, key, dead)
	Limiter.Refresh(ctx, rdb, key, alive)

	// When
	_, ok = Limiter.Acquire(ctx, rdb, key, 2)

	// Then the expired connection's slot is free again
	assert.True(t, ok)
	n, err := rdb.ZCard(ctx, key).Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(2), n)
}

func TestConnLimiter_Refresh_KeepsTheConnectionFromExpiring(t *testing.T) {
	// Given a held connection about to expire
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_refresh_key"
	defer rdb.Del(ctx, key)
	token, ok := Limiter.Acquire(ctx, rdb, key, 1)
	require.True(t, ok)
	expire(ctx, rdb, key, token)

	// When
	Limiter.Refresh(ctx, rdb, key, token)

	// Then it counts again
	_, ok = Limiter.Acquire(ctx, rdb, key, 1)
	assert.False(t, ok)
}

func TestConnLimiter_Refresh_DoesNotResurrectAnExpiredConnection(t *testing.T) {
	// Given
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_refresh_expired_key"
	defer rdb.Del(ctx, key)

	// When
	Limiter.Refresh(ctx, rdb, key, "token-of-a-holder-that-expired")

	// Then
	exists, err := rdb.Exists(ctx, key).Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), exists)
}

// expire backdates a held connection so that it counts as expired.
func expire(ctx context.Context, rdb *redis.Client, key, token string) {
	rdb.ZAdd(ctx, key, redis.Z{Score: float64(time.Now().Add(-time.Second).UnixMilli()), Member: token})
}

func testRedisClient() *redis.Client {
	redisHost := "localhost"
	if redisAddress := os.Getenv("REDIS_HOST"); len(redisAddress) > 0 {
		redisHost = redisAddress
	}
	return redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", redisHost)})
}
