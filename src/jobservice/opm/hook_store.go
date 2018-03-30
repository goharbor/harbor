// Copyright 2018 The Harbor Authors. All rights reserved.

package opm

import (
	"sync"

	"github.com/vmware/harbor/src/jobservice/utils"
)

//HookStore is used to cache the hooks in memory.
//Use job ID as key to index
type HookStore struct {
	lock *sync.RWMutex
	data map[string]string
}

//NewHookStore is to create a ptr of new HookStore.
func NewHookStore() *HookStore {
	return &HookStore{
		lock: new(sync.RWMutex),
		data: make(map[string]string),
	}
}

//Add new record
func (hs *HookStore) Add(jobID string, hookURL string) {
	if utils.IsEmptyStr(jobID) {
		return //do nothing
	}

	hs.lock.Lock()
	defer hs.lock.Unlock()

	hs.data[jobID] = hookURL
}

//Get one hook url by job ID
func (hs *HookStore) Get(jobID string) (string, bool) {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	hookURL, ok := hs.data[jobID]

	return hookURL, ok
}

//Remove the specified one
func (hs *HookStore) Remove(jobID string) (string, bool) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	hookURL, ok := hs.data[jobID]
	delete(hs.data, jobID)

	return hookURL, ok
}
