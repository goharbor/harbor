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
package auth

import (
	"testing"
	"time"
)

var l = NewUserLock(2 * time.Second)

func TestLock(t *testing.T) {
	t.Log("Locking john")
	l.Lock("john")
	if !l.IsLocked("john") {
		t.Errorf("John should be locked")
	}
	t.Log("Locking jack")
	l.Lock("jack")
	t.Log("Sleep for 2 seconds and check...")
	time.Sleep(2 * time.Second)
	if l.IsLocked("jack") {
		t.Errorf("After 2 seconds, jack shouldn't be locked")
	}
	if l.IsLocked("daniel") {
		t.Errorf("daniel has never been locked, he should not be locked")
	}
}
