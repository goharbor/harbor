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

	"github.com/redis/go-redis/v9"

	"github.com/goharbor/harbor/src/lib/log"
)

// ConLimiter is used to limit the number of connections to the upstream service
type ConnLimiter struct {
}

// Limiter is a global connection limiter instance
var Limiter = &ConnLimiter{}

// Used to compare and increase connection number in redis
//
// KEYS[1]: key of max_conn_upstream
// ARGV[1]: max connection limit
var increaseWithLimitText = `
local current = tonumber(redis.call('GET', KEYS[1]) or '0')
local max = tonumber(ARGV[1])

if current + 1 <= max then
    redis.call('INCRBY', KEYS[1], 1)
	redis.call('EXPIRE', KEYS[1], 3600) -- set expire to avoid always lock
    return 1
else
    return 0
end
`

var acquireScript = redis.NewScript(increaseWithLimitText)

// Acquire tries to acquire a connection, returns true if successful
func (c *ConnLimiter) Acquire(ctx context.Context, rdb *redis.Client, key string, limit int) bool {
	result, err := acquireScript.Run(ctx, rdb, []string{key}, fmt.Sprintf("%v", limit)).Int()
	if err != nil {
		log.Errorf("failed to get the connection lock in redis, error %v", err)
		return false
	}
	log.Debugf("Acquire script result is %d", result)
	return result == 1
}

var decreaseText = `
local val = tonumber(redis.call("GET", KEYS[1]) or "0")
if val > 0 then
    redis.call("DECR", KEYS[1])
end
return 0
`

var decreaseScript = redis.NewScript(decreaseText)

// Release releases a connection in redis
func (c *ConnLimiter) Release(ctx context.Context, rdb *redis.Client, key string) {
	_, err := decreaseScript.Run(ctx, rdb, []string{key}).Int()
	if err != nil {
		log.Infof("release connection failed:%v", err)
	}
}
