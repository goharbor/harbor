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

package opm

import (
	"sync"

	"github.com/goharbor/harbor/src/jobservice/utils"
)

// HookStore is used to cache the hooks in memory.
// Use job ID as key to index
type HookStore struct {
	lock *sync.RWMutex
	data map[string]string
}

// NewHookStore is to create a ptr of new HookStore.
func NewHookStore() *HookStore {
	return &HookStore{
		lock: new(sync.RWMutex),
		data: make(map[string]string),
	}
}

// Add new record
func (hs *HookStore) Add(jobID string, hookURL string) {
	if utils.IsEmptyStr(jobID) {
		return // do nothing
	}

	hs.lock.Lock()
	defer hs.lock.Unlock()

	hs.data[jobID] = hookURL
}

// Get one hook url by job ID
func (hs *HookStore) Get(jobID string) (string, bool) {
	hs.lock.RLock()
	defer hs.lock.RUnlock()

	hookURL, ok := hs.data[jobID]

	return hookURL, ok
}

// Remove the specified one
func (hs *HookStore) Remove(jobID string) (string, bool) {
	hs.lock.Lock()
	defer hs.lock.Unlock()

	hookURL, ok := hs.data[jobID]
	delete(hs.data, jobID)

	return hookURL, ok
}
