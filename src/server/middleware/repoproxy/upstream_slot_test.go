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
	"context"
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

const testWaitInterval = time.Millisecond

// trueAfter returns a func that is false for its first n calls.
func trueAfter(n int) func() bool {
	calls := 0
	return func() bool {
		calls++
		return calls > n
	}
}

func never() bool { return false }

func TestWaitForUpstreamSlot_AcquiresImmediately(t *testing.T) {
	// Given
	localChecked := false
	existsLocally := func() bool {
		localChecked = true
		return false
	}

	// When
	access, err := waitForUpstreamSlot(context.Background(), trueAfter(0), existsLocally, testWaitInterval)

	// Then
	require.NoError(t, err)
	require.Equal(t, fetchUpstream, access)
	require.False(t, localChecked, "a free slot must not cost a local lookup")
}

func TestWaitForUpstreamSlot_ServesLocalOnceAnotherRequestStoredTheBlob(t *testing.T) {
	// Given
	existsLocally := trueAfter(2)

	// When
	access, err := waitForUpstreamSlot(context.Background(), never, existsLocally, testWaitInterval)

	// Then
	require.NoError(t, err)
	require.Equal(t, serveLocal, access)
}

func TestWaitForUpstreamSlot_FetchesUpstreamWhenASlotFreesUp(t *testing.T) {
	// Given
	acquire := trueAfter(50)

	// When
	access, err := waitForUpstreamSlot(context.Background(), acquire, never, testWaitInterval)

	// Then
	require.NoError(t, err)
	require.Equal(t, fetchUpstream, access)
}

func TestWaitForUpstreamSlot_ChecksLocalBeforeReacquiring(t *testing.T) {
	// Given the slot frees and the blob lands locally while this request
	// sleeps
	freed := false
	acquire := func() bool {
		if freed {
			return true
		}
		freed = true
		return false
	}
	existsLocally := func() bool { return freed }

	// When
	access, err := waitForUpstreamSlot(context.Background(), acquire, existsLocally, testWaitInterval)

	// Then the local copy wins, so the slot is not spent on a fetch nobody
	// needs
	require.NoError(t, err)
	require.Equal(t, serveLocal, access)
}

func TestWaitForUpstreamSlot_StopsWhenTheClientGoesAway(t *testing.T) {
	// Given
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// When
	_, err := waitForUpstreamSlot(ctx, never, never, time.Hour)

	// Then
	require.ErrorIs(t, err, context.Canceled)
}
