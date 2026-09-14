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
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/goharbor/harbor/src/lib/log"
)

// ConLimiter is used to limit the number of connections to the upstream service
type ConnLimiter struct {
}

// Limiter is a global connection limiter instance
var Limiter = &ConnLimiter{}

// SlotTTL is how long a held connection outlives its last Acquire or Refresh.
// A holder that neither releases nor refreshes within this time, because the
// process died, stops counting against the limit.
const SlotTTL = 60 * time.Second

// The held connections under a key are a sorted set of tokens scored by
// their expiry, so that each holder's connection expires on its own: one
// holder refreshing its token does not keep a dead holder's counted.
//
// KEYS[1]: key of max_conn_upstream
// ARGV[1]: max connection limit
// ARGV[2]: token of the connection to acquire
// ARGV[3]: now, unix milliseconds
// ARGV[4]: expiry of the connection, unix milliseconds
// ARGV[5]: milliseconds until the whole key expires unless refreshed
var acquireText = `
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', ARGV[3])
if redis.call('ZCARD', KEYS[1]) < tonumber(ARGV[1]) then
    redis.call('ZADD', KEYS[1], ARGV[4], ARGV[2])
    redis.call('PEXPIRE', KEYS[1], ARGV[5])
    return 1
end
return 0
`

var acquireScript = redis.NewScript(acquireText)

// Acquire tries to acquire a connection. On success it returns the token that
// identifies the connection to Refresh and Release.
func (c *ConnLimiter) Acquire(ctx context.Context, rdb *redis.Client, key string, limit int) (string, bool) {
	token := uuid.NewString()
	now := time.Now()
	result, err := acquireScript.Run(ctx, rdb, []string{key},
		limit, token, now.UnixMilli(), now.Add(SlotTTL).UnixMilli(), SlotTTL.Milliseconds()).Int()
	if err != nil {
		log.Errorf("failed to get the connection lock in redis, error %v", err)
		return "", false
	}
	log.Debugf("Acquire script result is %d", result)
	return token, result == 1
}

// Refresh extends the connection's expiry by SlotTTL. It is a no-op on a
// connection that has already expired, so a late refresh cannot bring it back.
func (c *ConnLimiter) Refresh(ctx context.Context, rdb *redis.Client, key, token string) {
	expiry := float64(time.Now().Add(SlotTTL).UnixMilli())
	if err := rdb.ZAddXX(ctx, key, redis.Z{Score: expiry, Member: token}).Err(); err != nil {
		log.Warningf("failed to refresh the connection lock in redis, key: %s, error: %v", key, err)
		return
	}
	// Every live token expires within SlotTTL, so the key may go with them.
	if err := rdb.PExpire(ctx, key, SlotTTL).Err(); err != nil {
		log.Warningf("failed to refresh the connection lock's expiry in redis, key: %s, error: %v", key, err)
	}
}

// Release releases the connection in redis
func (c *ConnLimiter) Release(ctx context.Context, rdb *redis.Client, key, token string) {
	if err := rdb.ZRem(ctx, key, token).Err(); err != nil {
		log.Infof("release connection failed:%v", err)
	}
}
