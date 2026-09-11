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
)

func TestConnLimiter_Acquire_Release(t *testing.T) {
	redisAddress := os.Getenv("REDIS_HOST")
	redisHost := "localhost"
	if len(redisAddress) > 0 {
		redisHost = redisAddress
	}

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", redisHost), // Redis server address
		Password: "",                                // No password set
		DB:       0,                                 // Use default DB
	})
	key := "test_max_connection_key"
	maxConn := 10
	for range 10 {
		result := Limiter.Acquire(ctx, rdb, key, maxConn)
		assert.True(t, result)
	}
	// after max connection reached, it should be false
	result2 := Limiter.Acquire(ctx, rdb, key, maxConn)
	assert.False(t, result2)

	for range 10 {
		Limiter.Release(ctx, rdb, key)
	}

	// connection in redis should be 0 finally
	n, err := rdb.Get(ctx, key).Int()
	assert.Nil(t, err)
	assert.Equal(t, 0, n)

}

func TestConnLimiter_Acquire_ExpiresAfterSlotTTL(t *testing.T) {
	// Given
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_acquire_ttl_key"
	defer rdb.Del(ctx, key)

	// When
	assert.True(t, Limiter.Acquire(ctx, rdb, key, 1))

	// Then
	ttl, err := rdb.TTL(ctx, key).Result()
	assert.Nil(t, err)
	assert.True(t, ttl > 0 && ttl <= SlotTTL, "a held slot expires after SlotTTL unless refreshed, got %v", ttl)
}

func TestConnLimiter_Refresh_KeepsTheSlotFromExpiring(t *testing.T) {
	// Given a held slot about to expire
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_refresh_key"
	defer rdb.Del(ctx, key)
	assert.True(t, Limiter.Acquire(ctx, rdb, key, 1))
	rdb.Expire(ctx, key, time.Second)

	// When
	Limiter.Refresh(ctx, rdb, key)

	// Then
	ttl, err := rdb.TTL(ctx, key).Result()
	assert.Nil(t, err)
	assert.True(t, ttl > time.Second, "refresh must extend the slot back to SlotTTL, got %v", ttl)
}

func TestConnLimiter_Refresh_DoesNotResurrectAnExpiredSlot(t *testing.T) {
	// Given
	ctx := context.Background()
	rdb := testRedisClient()
	key := "test_refresh_expired_key"
	defer rdb.Del(ctx, key)

	// When
	Limiter.Refresh(ctx, rdb, key)

	// Then
	exists, err := rdb.Exists(ctx, key).Result()
	assert.Nil(t, err)
	assert.Equal(t, int64(0), exists)
}

func testRedisClient() *redis.Client {
	redisHost := "localhost"
	if redisAddress := os.Getenv("REDIS_HOST"); len(redisAddress) > 0 {
		redisHost = redisAddress
	}
	return redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", redisHost)})
}
