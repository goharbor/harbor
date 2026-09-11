//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestKeepSlotAlive_RefreshesUntilStopped(t *testing.T) {
	// Given
	var refreshes atomic.Int32
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		keepSlotAlive(stop, time.Millisecond, func() { refreshes.Add(1) })
		close(done)
	}()
	require.Eventually(t, func() bool { return refreshes.Load() >= 3 }, time.Second, time.Millisecond)

	// When
	close(stop)

	// Then
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("keepSlotAlive did not return after stop")
	}
	settled := refreshes.Load()
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, settled, refreshes.Load(), "no refresh may happen after stop")
}
