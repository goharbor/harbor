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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSentinelURL(t *testing.T) {
	url := "redis+sentinel://anonymous:password@host1:26379,host2:26379/mymaster/1?idle_timeout_seconds=30&max_retries=10&min_retry_backoff=1&max_retry_backoff=10&dial_timeout=30&read_timeout=5&write_timeout=5&pool_fifo=true&pool_size=1000&min_idle_conns=100&max_conn_age=10&pool_timeout=10"
	o, err := ParseSentinelURL(url)
	assert.NoError(t, err)
	assert.Equal(t, "anonymous", o.Username)
	assert.Equal(t, "password", o.Password)
	assert.Equal(t, []string{"host1:26379", "host2:26379"}, o.SentinelAddrs)
	assert.Equal(t, "mymaster", o.MasterName)
	assert.Equal(t, 1, o.DB)
	assert.Equal(t, 30*time.Second, o.IdleTimeout)
	assert.Equal(t, 10, o.MaxRetries)
	assert.Equal(t, 10*time.Second, o.MaxRetryBackoff)
	assert.Equal(t, 1*time.Second, o.MinRetryBackoff)
	assert.Equal(t, 30*time.Second, o.DialTimeout)
	assert.Equal(t, 5*time.Second, o.ReadTimeout)
	assert.Equal(t, 5*time.Second, o.WriteTimeout)
	assert.Equal(t, true, o.PoolFIFO)
	assert.Equal(t, 1000, o.PoolSize)
	assert.Equal(t, 100, o.MinIdleConns)
	assert.Equal(t, 10*time.Second, o.MaxConnAge)
	assert.Equal(t, 10*time.Second, o.PoolTimeout)

	// invalid url should return err
	url = "invalid"
	_, err = ParseSentinelURL(url)
	assert.Error(t, err, "invalid url should return err")

}
