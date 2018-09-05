package scheduler

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Store define the basic operations for storing and managing policy watcher.
type Store interface {
	// Put a new policy in.
	Put(key string, value *Watcher) error

	// Get the corresponding policy with the key.
	Get(key string) *Watcher

	// Exists is to check if the key existing in the store.
	Exists(key string) bool

	// Remove the specified policy and return its reference.
	Remove(key string) *Watcher

	// Size return the total count of items in store.
	Size() uint32

	// GetAll is to get all the items in the store.
	GetAll() []*Watcher

	// Clear store.
	Clear()
}

// DefaultStore implements Store interface to keep the scheduled policies.
// Not support concurrent sync.
type DefaultStore struct {
	// Support sync locking
	*sync.RWMutex

	// Map used to keep the policy list.
	data map[string]*Watcher
}

// NewDefaultStore is used to create a new store and return the pointer reference.
func NewDefaultStore() *DefaultStore {
	return &DefaultStore{new(sync.RWMutex), make(map[string]*Watcher)}
}

// Put a policy into store.
func (cs *DefaultStore) Put(key string, value *Watcher) error {
	if strings.TrimSpace(key) == "" || value == nil {
		return errors.New("Bad arguments")
	}

	cs.Lock()
	defer cs.Unlock()

	if _, ok := cs.data[key]; ok {
		return fmt.Errorf("Duplicayed policy with name %s", key)
	}

	cs.data[key] = value

	return nil
}

// Get policy via key.
func (cs *DefaultStore) Get(key string) *Watcher {
	if strings.TrimSpace(key) == "" {
		return nil
	}

	cs.RLock()
	defer cs.RUnlock()

	return cs.data[key]
}

// Exists is used to check whether or not the key exists in store.
func (cs *DefaultStore) Exists(key string) bool {
	if strings.TrimSpace(key) == "" {
		return false
	}

	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.data[key]

	return ok
}

// Remove is to delete the specified policy.
func (cs *DefaultStore) Remove(key string) *Watcher {
	if strings.TrimSpace(key) == "" {
		return nil
	}

	cs.Lock()
	defer cs.Unlock()

	if wt, ok := cs.data[key]; ok {
		delete(cs.data, key)
		return wt
	}

	return nil
}

// Size return the total count of items in store.
func (cs *DefaultStore) Size() uint32 {
	cs.RLock()
	defer cs.RUnlock()

	return (uint32)(len(cs.data))
}

// GetAll to get all the items of store.
func (cs *DefaultStore) GetAll() []*Watcher {
	cs.RLock()
	defer cs.RUnlock()

	all := []*Watcher{}

	for _, v := range cs.data {
		all = append(all, v)
	}

	return all
}

// Clear all the items in store.
func (cs *DefaultStore) Clear() {
	cs.Lock()
	defer cs.Unlock()

	if (uint32)(len(cs.data)) == 0 {
		return
	}

	for k, v := range cs.data {
		delete(cs.data, k)
		v.Stop()
	}
}
