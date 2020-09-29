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

package list

import (
	"container/list"
	"sync"
)

// SyncList is a sync list based on the container/list
type SyncList struct {
	// For synchronization
	lock *sync.RWMutex
	// Use interface slice as the backend data struct
	l *list.List
}

// New a sync list
func New() *SyncList {
	return &SyncList{
		lock: &sync.RWMutex{},
		l:    list.New(),
	}
}

// Iterate the list
func (l *SyncList) Iterate(f func(ele interface{}) bool) {
	l.lock.RLock()
	defer l.lock.RUnlock()

	// Get the front pointer
	for e := l.l.Front(); e != nil; {
		// Keep the next one
		next := e.Next()

		if f(e.Value) {
			l.l.Remove(e)
		}

		e = next
	}
}

// Push the element to the back of the list
func (l *SyncList) Push(ele interface{}) {
	if ele != nil {
		l.lock.Lock()
		defer l.lock.Unlock()

		l.l.PushBack(ele)
	}
}
