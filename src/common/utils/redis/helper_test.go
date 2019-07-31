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
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/stretchr/testify/assert"
)

const testingRedisHost = "REDIS_HOST"

func init() {
	os.Setenv("REDIS_LOCK_MAX_RETRY", "5")
}

func TestRedisLock(t *testing.T) {
	con, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", getRedisHost(), 6379))
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

func TestRequireLock(t *testing.T) {
	assert := assert.New(t)

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", getRedisHost(), 6379))
	assert.Nil(err)
	defer conn.Close()

	if l, err := RequireLock(utils.GenerateRandomString(), conn); assert.Nil(err) {
		l.Free()
	}

	if l, err := RequireLock(utils.GenerateRandomString()); assert.Nil(err) {
		FreeLock(l)
	}

	key := utils.GenerateRandomString()
	if l, err := RequireLock(key); assert.Nil(err) {
		defer FreeLock(l)

		_, err = RequireLock(key)
		assert.Error(err)
	}
}

func TestFreeLock(t *testing.T) {
	assert := assert.New(t)

	if l, err := RequireLock(utils.GenerateRandomString()); assert.Nil(err) {
		assert.Nil(FreeLock(l))
	}

	conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", getRedisHost(), 6379))
	assert.Nil(err)

	if l, err := RequireLock(utils.GenerateRandomString(), conn); assert.Nil(err) {
		conn.Close()
		assert.Error(FreeLock(l))
	}
}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return redisHost
}
