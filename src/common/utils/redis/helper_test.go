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
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const testingRedisHost = "REDIS_HOST"

func TestRedisLock(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	rm := New(con, "test-redis-lock", "test-value")

	successLock, err := rm.Require()
	assert.Nil(t, err)
	assert.True(t, successLock)

	time.Sleep(2 * time.Second)
	_, err = rm.Require()
	assert.NotNil(t, err)

	successUnLock, err := rm.Free()
	assert.Nil(t, err)
	assert.True(t, successUnLock)

}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return redisHost
}
