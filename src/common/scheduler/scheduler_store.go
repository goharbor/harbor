package scheduler

import "sync"
import "strings"

const defaultSize = 10

//Store define the basic operations for storing and managing policy watcher.
//The concrete implementation should consider concurrent supporting scenario.
//
type Store interface {
	//Put a new policy in.
	Put(key string, value *Watcher)

	//Get the corresponding policy with the key.
	Get(key string) *Watcher

	//Exists is to check if the key existing in the store.
	Exists(key string) bool

	//Remove the specified policy and return its reference.
	Remove(key string) *Watcher

	//Size return the total count of items in store.
	Size() uint32

	//GetAll is to get all the items in the store.
	GetAll() []*Watcher

	//Clear store.
	Clear()
}

//ConcurrentStore implements Store interface and supports concurrent operations.
type ConcurrentStore struct {
	//Read-write mutex to synchronize the data map.
	mutex *sync.RWMutex

	//Map used to keep the policy list.
	data map[string]*Watcher
}

//NewConcurrentStore is used to create a new store and return the pointer reference.
func NewConcurrentStore(initialSize uint32) *ConcurrentStore {
	var initSize uint32 = defaultSize
	if initialSize > 0 {
		initSize = initialSize
	}
	mutex := new(sync.RWMutex)
	data := make(map[string]*Watcher, initSize)

	return &ConcurrentStore{mutex, data}
}

//Put a policy into store.
func (cs *ConcurrentStore) Put(key string, value *Watcher) {
	if strings.TrimSpace(key) == "" || value == nil {
		return
	}

	defer cs.mutex.Unlock()

	cs.mutex.Lock()
	cs.data[key] = value
}

//Get policy via key.
func (cs *ConcurrentStore) Get(key string) *Watcher {
	if strings.TrimSpace(key) == "" {
		return nil
	}

	defer cs.mutex.RUnlock()

	cs.mutex.RLock()
	return cs.data[key]
}

//Exists is used to check whether or not the key exists in store.
func (cs *ConcurrentStore) Exists(key string) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}

	defer cs.mutex.RUnlock()

	cs.mutex.RLock()
	_, ok := cs.data[key]

	return ok
}

//Remove is to delete the specified policy.
func (cs *ConcurrentStore) Remove(key string) *Watcher {
	if !cs.Exists(key) {
		return nil
	}

	defer cs.mutex.Unlock()

	cs.mutex.Lock()
	if wt, ok := cs.data[key]; ok {
		delete(cs.data, key)
		return wt
	}

	return nil
}

//Size return the total count of items in store.
func (cs *ConcurrentStore) Size() uint32 {
	return (uint32)(len(cs.data))
}

//GetAll to get all the items of store.
func (cs *ConcurrentStore) GetAll() []*Watcher {
	all := []*Watcher{}

	defer cs.mutex.RUnlock()
	cs.mutex.RLock()
	for _, v := range cs.data {
		all = append(all, v)
	}

	return all
}

//Clear all the items in store.
func (cs *ConcurrentStore) Clear() {
	if cs.Size() == 0 {
		return
	}

	defer cs.mutex.Unlock()
	cs.mutex.Lock()

	for k := range cs.data {
		delete(cs.data, k)
	}
}
