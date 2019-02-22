package trigger

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

const (
	// The max count of items the cache can keep
	defaultCapacity = 1000
)

// Item keeps more metadata of the triggers which are stored in the heap.
type Item struct {
	// Which policy the trigger belong to
	policyID int64

	// Frequency of cache querying
	// First compration factor
	frequency int

	// The timestamp of being put into heap
	// Second compration factor
	timestamp int64

	// The index in the heap
	index int
}

// MetaQueue implements heap.Interface and holds items which are metadata of trigger
type MetaQueue []*Item

// Len return the size of the queue
func (mq MetaQueue) Len() int {
	return len(mq)
}

// Less is a comparator of heap
func (mq MetaQueue) Less(i, j int) bool {
	return mq[i].frequency < mq[j].frequency ||
		(mq[i].frequency == mq[j].frequency &&
			mq[i].timestamp < mq[j].timestamp)
}

// Swap the items to rebuild heap
func (mq MetaQueue) Swap(i, j int) {
	mq[i], mq[j] = mq[j], mq[i]
	mq[i].index = i
	mq[j].index = j
}

// Push item into heap
func (mq *MetaQueue) Push(x interface{}) {
	item := x.(*Item)
	n := len(*mq)
	item.index = n
	item.timestamp = time.Now().UTC().UnixNano()
	*mq = append(*mq, item)
}

// Pop smallest item from heap
func (mq *MetaQueue) Pop() interface{} {
	old := *mq
	n := len(old)
	item := old[n-1] // Smallest item
	item.index = -1  // For safety
	*mq = old[:n-1]
	return item
}

// Update the frequency of item
func (mq *MetaQueue) Update(item *Item) {
	item.frequency++
	heap.Fix(mq, item.index)
}

// CacheItem is the data stored in the cache.
// It contains trigger and heap item references.
type CacheItem struct {
	// The trigger reference
	trigger Interface

	// The heap item reference
	item *Item
}

// Cache is used to cache the enabled triggers with specified capacity.
// If exceed the capacity, cached items will be adjusted with the following rules:
// The item with least usage frequency will be replaced;
// If multiple items with same usage frequency, the oldest one will be replaced.
type Cache struct {
	// The max count of items this cache can keep
	capacity int

	// Lock to handle concurrent case
	lock *sync.RWMutex

	// Hash map for quick locating cached item
	hash map[string]CacheItem

	// Heap for quick locating the trigger with least usage
	queue *MetaQueue
}

// NewCache is constructor of cache
func NewCache(capacity int) *Cache {
	capa := capacity
	if capa <= 0 {
		capa = defaultCapacity
	}

	// Initialize heap
	mq := make(MetaQueue, 0)
	heap.Init(&mq)

	return &Cache{
		capacity: capa,
		lock:     new(sync.RWMutex),
		hash:     make(map[string]CacheItem),
		queue:    &mq,
	}
}

// Get the trigger interface with the specified policy ID
func (c *Cache) Get(policyID int64) Interface {
	if policyID <= 0 {
		return nil
	}

	c.lock.RLock()
	defer c.lock.RUnlock()

	k := c.key(policyID)

	if cacheItem, ok := c.hash[k]; ok {
		// Update frequency
		c.queue.Update(cacheItem.item)
		return cacheItem.trigger
	}

	return nil
}

// Put the item into cache with ID of ploicy as key
func (c *Cache) Put(policyID int64, trigger Interface) {
	if policyID <= 0 || trigger == nil {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// Exceed the capacity?
	if c.Size() >= c.capacity {
		// Pop one for the new one
		v := heap.Pop(c.queue)
		item := v.(*Item)
		// Remove from hash
		delete(c.hash, c.key(item.policyID))
	}

	// Add to meta queue
	item := &Item{
		policyID:  policyID,
		frequency: 1,
	}
	heap.Push(c.queue, item)

	// Cache
	cacheItem := CacheItem{
		trigger: trigger,
		item:    item,
	}

	k := c.key(policyID)
	c.hash[k] = cacheItem
}

// Remove the trigger attached to the specified policy
func (c *Cache) Remove(policyID int64) Interface {
	if policyID > 0 {
		c.lock.Lock()
		defer c.lock.Unlock()

		// If existing
		k := c.key(policyID)
		if cacheItem, ok := c.hash[k]; ok {
			// Remove from heap
			heap.Remove(c.queue, cacheItem.item.index)

			// Remove from hash
			delete(c.hash, k)

			return cacheItem.trigger
		}

	}

	return nil
}

// Size return the count of triggers in the cache
func (c *Cache) Size() int {
	return len(c.hash)
}

// Generate a hash key with the policy ID
func (c *Cache) key(policyID int64) string {
	return fmt.Sprintf("trigger-%d", policyID)
}
