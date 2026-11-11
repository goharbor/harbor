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
	"errors"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"

	"github.com/goharbor/harbor/src/lib/cache"
	libredis "github.com/goharbor/harbor/src/lib/cache/redis"
	"github.com/goharbor/harbor/src/lib/log"
)

var (
	// registry is a global redis client for registry db
	registry     *redis.Client
	registryOnce = &sync.Once{}

	// harbor is a global redis client for harbor db
	harbor     *redis.Client
	harborOnce = &sync.Once{}
)

// GetRegistryClient returns the registry redis client.
func GetRegistryClient() (*redis.Client, error) {
	registryOnce.Do(func() {
		url := os.Getenv("_REDIS_URL_REG")
		c, err := libredis.New(cache.Options{Address: url})
		if err != nil {
			log.Errorf("failed to initialize redis client for registry, error: %v", err)
			// reset the once to support retry if error occurred
			registryOnce = &sync.Once{}
			return
		}

		if c != nil {
			registry = c.(*libredis.Cache).Client
		}
	})

	if registry == nil {
		return nil, errors.New("no registry redis client initialized")
	}

	return registry, nil
}

// GetHarborClient returns the harbor redis client.
func GetHarborClient() (*redis.Client, error) {
	harborOnce.Do(func() {
		// parse redis url for harbor business, use core url by default
		url := os.Getenv("_REDIS_URL_HARBOR")
		if url == "" {
			url = os.Getenv("_REDIS_URL_CORE")
		}
		c, err := libredis.New(cache.Options{Address: url})
		if err != nil {
			log.Errorf("failed to initialize redis client for harbor, error: %v", err)
			// reset the once to support retry if error occurred
			harborOnce = &sync.Once{}
			return
		}

		if c != nil {
			harbor = c.(*libredis.Cache).Client
		}
	})

	if harbor == nil {
		return nil, errors.New("no harbor redis client initialized")
	}

	return harbor, nil
}
