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
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/goharbor/harbor/src/lib/cache"
	libredis "github.com/goharbor/harbor/src/lib/cache/redis"
)

var (
	// instance is a global redis client.
	_instance *redis.Client
	_once     sync.Once
)

// Instance returns the redis instance.
func Instance() *redis.Client {
	_once.Do(func() {
		url := os.Getenv("_REDIS_URL_REG")
		if url == "" {
			url = "redis://localhost:6379/1"
		}

		c, err := libredis.New(cache.Options{Address: url})
		if err != nil {
			panic(err)
		}

		_instance = c.(*libredis.Cache).Client
	})

	return _instance
}
