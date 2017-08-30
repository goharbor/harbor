// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package utils

import (
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	scanAllMarker      *TimeMarker
	scanOverviewMarker = &TimeMarker{
		interval: 15 * time.Second,
	}
	once sync.Once
)

//TimeMarker is used to control an action not to be taken frequently within the interval
type TimeMarker struct {
	sync.RWMutex
	next     time.Time
	interval time.Duration
}

//Mark tries to mark a future time, which is after the duration of interval from the time it's called.
func (t *TimeMarker) Mark() {
	t.Lock()
	defer t.Unlock()
	t.next = time.Now().Add(t.interval)
}

//Check returns true if the current time is after the mark by this marker, and the caction the mark guards and be taken.
func (t *TimeMarker) Check() bool {
	t.RLock()
	defer t.RUnlock()
	return time.Now().After(t.next)
}

//Next returns the time of the next mark.
func (t *TimeMarker) Next() time.Time {
	t.RLock()
	defer t.RUnlock()
	return t.next
}

//ScanAllMarker ...
func ScanAllMarker() *TimeMarker {
	once.Do(func() {
		a := os.Getenv("HARBOR_SCAN_ALL_INTERVAL")
		if m, err := strconv.Atoi(a); err == nil {
			scanAllMarker = &TimeMarker{
				interval: time.Duration(m) * time.Minute,
			}
		} else {
			scanAllMarker = &TimeMarker{
				interval: 2 * time.Hour,
			}
		}
	})
	return scanAllMarker
}

//ScanOverviewMarker ...
func ScanOverviewMarker() *TimeMarker {
	return scanOverviewMarker
}
