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
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
)

// LRUEntry represents an item in the LRU Cache tracking system.
// We only track the digest and size; the actual blob data remains on the filesystem (local registry).
type LRUEntry struct {
	Digest       string
	Size         int64
	LastAccessed time.Time
}

// LocalLRUCache tracks image blobs pulled through the proxy cache
// and handles eviction when the overall size exceeds the limit.
type LocalLRUCache struct {
	mu           sync.RWMutex
	maxSizeBytes int64
	currentSize  int64

	// evictList holds *LRUEntry. Front is most recently used.
	evictList *list.List
	// items maps digest -> list element pointer
	items map[string]*list.Element

	// evictionCallback is triggered when an item must be deleted from local disk
	evictionCallback func(digest string) error
}

// NewLocalLRUCache creates a new bounded LRU cache for tracking proxy blobs.
// maxSizeBytes configures the hard limit of the cache. If 0, no eviction happens.
// evictionCallback handles the physical deletion of the blob from the local registry.
func NewLocalLRUCache(maxSizeBytes int64, cb func(digest string) error) *LocalLRUCache {
	return &LocalLRUCache{
		maxSizeBytes:     maxSizeBytes,
		evictList:        list.New(),
		items:            make(map[string]*list.Element),
		evictionCallback: cb,
	}
}

// Add tracks a new blob or updates an existing one's access time.
// Triggers eviction if the new size exceeds maxSizeBytes.
func (c *LocalLRUCache) Add(ctx context.Context, digest string, size int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If tracking is disabled (maxSize == 0), don't bother
	if c.maxSizeBytes <= 0 {
		return
	}

	if ent, ok := c.items[digest]; ok {
		// Existing item: Update access time and move to front
		c.evictList.MoveToFront(ent)
		entry := ent.Value.(*LRUEntry)

		// Adjust size differential if somehow the size changed
		c.currentSize += (size - entry.Size)
		entry.Size = size
		entry.LastAccessed = time.Now()
	} else {
		// New item: Add to tracker
		ent := &LRUEntry{
			Digest:       digest,
			Size:         size,
			LastAccessed: time.Now(),
		}
		element := c.evictList.PushFront(ent)
		c.items[digest] = element
		c.currentSize += size
	}

	// Validate cache limits
	c.evictUntilLimit(ctx)
}

// Access marks a blob as recently used, moving it to the front of the LRU queue.
func (c *LocalLRUCache) Access(digest string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[digest]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*LRUEntry).LastAccessed = time.Now()
	}
}

// Remove manually drops a blob from the tracker (e.g. if deleted manually by user/registry).
func (c *LocalLRUCache) Remove(digest string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ent, ok := c.items[digest]; ok {
		c.removeElement(ent)
	}
}

// evictUntilLimit forces the cache size to drop under maxSizeBytes
// Ensure caller holds the lock before executing this.
func (c *LocalLRUCache) evictUntilLimit(ctx context.Context) {
	for c.currentSize > c.maxSizeBytes && c.evictList.Len() > 0 {
		// Get Oldest (Back of list)
		element := c.evictList.Back()
		if element != nil {
			entry := element.Value.(*LRUEntry)

			// Callback to actually delete the item from the registry disk/DB
			if c.evictionCallback != nil {
				err := c.evictionCallback(entry.Digest)
				if err != nil {
					log.Errorf("LRU Cache: Failed to evict blob digest %s: %v", entry.Digest, err)
					// If it fails, remove it from tracker anyway to avoid infinite loops,
					// or move to front. We choose removal to stop tracking dead ends.
					c.removeElement(element)
				} else {
					log.Infof("LRU Cache: Successfully evicted proxy blob %s (size: %d) to maintain cache limits", entry.Digest, entry.Size)
					c.removeElement(element)
				}
			} else {
				c.removeElement(element)
			}
		}
	}
}

func (c *LocalLRUCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	entry := e.Value.(*LRUEntry)
	delete(c.items, entry.Digest)
	c.currentSize -= entry.Size
}

// GetCurrentSize returns the tracked size in bytes.
func (c *LocalLRUCache) GetCurrentSize() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize
}
