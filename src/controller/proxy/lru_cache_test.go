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

package proxy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLocalLRUCache(t *testing.T) {
	ctx := context.TODO()

	var deletedDigests []string
	evictionCb := func(digest string) error {
		deletedDigests = append(deletedDigests, digest)
		return nil
	}

	// 100 bytes max size
	cache := NewLocalLRUCache(100, evictionCb)

	// Add 40 byte item
	cache.Add(ctx, "digest-1", 40)
	assert.Equal(t, int64(40), cache.GetCurrentSize())
	assert.Empty(t, deletedDigests)

	// Add 50 byte item (Total: 90/100)
	cache.Add(ctx, "digest-2", 50)
	assert.Equal(t, int64(90), cache.GetCurrentSize())

	// Access digest-1, making digest-2 the LRU
	time.Sleep(1 * time.Millisecond)
	cache.Access("digest-1")

	// Add 30 byte item (Total: 120/100 -> Evict digest-2 to bring total to 70/100)
	cache.Add(ctx, "digest-3", 30)
	assert.Equal(t, int64(70), cache.GetCurrentSize())

	assert.Contains(t, deletedDigests, "digest-2")
	assert.NotContains(t, deletedDigests, "digest-1")

	// Add 50 byte item (Total: 120/100 -> Evict digest-1 to bring total to 80/100)
	cache.Add(ctx, "digest-4", 50)
	assert.Contains(t, deletedDigests, "digest-1")
	assert.Equal(t, int64(80), cache.GetCurrentSize())

	// Test failing eviction handler
	cache = NewLocalLRUCache(50, func(d string) error {
		return errors.New("delete failed")
	})
	cache.Add(ctx, "fail-1", 40)
	cache.Add(ctx, "fail-2", 20)

	// Because eviction callback returned error, we still remove it from tracking to prevent infinite loops.
	assert.Equal(t, int64(20), cache.GetCurrentSize())
}
