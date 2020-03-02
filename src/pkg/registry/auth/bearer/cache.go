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

package bearer

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"strings"
	"sync"
	"time"
)

func newCache(capacity int) *cache {
	return &cache{
		latency:  10,
		capacity: capacity,
		cache:    map[string]*token{},
	}
}

type cache struct {
	sync.RWMutex
	latency  int // second, the network latency in case that when the token is checked it doesn't expire but it does when used
	capacity int // the capacity of the cache map
	cache    map[string]*token
}

func (c *cache) get(scopes []*scope) *token {
	c.RLock()
	defer c.RUnlock()
	return c.cache[c.key(scopes)]
}

func (c *cache) set(scopes []*scope, token *token) {
	c.Lock()
	defer c.Unlock()
	// exceed the capacity, empty some elements: all expired token will be removed,
	// if no expired token, move the earliest one
	if len(c.cache) >= c.capacity {
		now := time.Now().UTC()
		var candidates []string
		var earliestKey string
		var earliestExpireTime time.Time
		for key, value := range c.cache {
			// parse error
			issueAt, err := time.Parse(time.RFC3339, value.IssuedAt)
			if err != nil {
				log.Errorf("failed to parse the issued at time of token %s: %v", token.IssuedAt, err)
				candidates = append(candidates, key)
				continue
			}

			expireAt := issueAt.Add(time.Duration(value.ExpiresIn-c.latency) * time.Second)
			// expired
			if expireAt.Before(now) {
				candidates = append(candidates, key)
				continue
			}
			// doesn't expired
			if len(earliestKey) == 0 || expireAt.Before(earliestExpireTime) {
				earliestKey = key
				earliestExpireTime = expireAt
				continue
			}
		}
		if len(candidates) == 0 {
			candidates = append(candidates, earliestKey)
		}
		for _, candidate := range candidates {
			delete(c.cache, candidate)
		}
	}
	c.cache[c.key(scopes)] = token
}

func (c *cache) key(scopes []*scope) string {
	var strs []string
	for _, scope := range scopes {
		strs = append(strs, scope.String())
	}
	return strings.Join(strs, "#")
}
