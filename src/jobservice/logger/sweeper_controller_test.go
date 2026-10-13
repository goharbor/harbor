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

package logger

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/jobservice/logger/sweeper"
	libcfg "github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func init() {
	libcfg.DefaultCfgManager = common.InMemoryCfgManager
}

type mockSweeper struct {
	duration int
	sweepFn  func() (int, error)
}

func (m *mockSweeper) Sweep() (int, error) {
	if m.sweepFn != nil {
		return m.sweepFn()
	}
	return 0, nil
}

func (m *mockSweeper) Duration() int {
	return m.duration
}

func TestSweeperController_Duration(t *testing.T) {
	c := NewSweeperController(context.Background(), nil)
	assert.Equal(t, -1, c.Duration())
}

func TestSweeperController_DefaultSweepInterval(t *testing.T) {
	ctx := context.Background()
	c := NewSweeperController(ctx, nil)
	assert.Equal(t, oneDay, c.sweepInterval)
}

// TestSweeperController_PeriodicExecutionWithLargeRetentionDuration tests issue #24133:
// Even when a sweeper is configured with a large retention duration (e.g. 30 days),
// the sweeper loop must execute on startup AND periodically on its configured sweep interval,
// rather than setting the ticker to 30 days and delaying subsequent sweeps.
func TestSweeperController_PeriodicExecutionWithLargeRetentionDuration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sweepCount int32
	ms := &mockSweeper{
		duration: 30, // 30 days retention
		sweepFn: func() (int, error) {
			atomic.AddInt32(&sweepCount, 1)
			return 1, nil
		},
	}

	testInterval := 25 * time.Millisecond
	c := NewSweeperController(ctx, []sweeper.Interface{ms})
	c.sweepInterval = testInterval

	_, err := c.Sweep()
	require.NoError(t, err)

	// Wait for at least 3 sweeps: startup sweep + at least 2 subsequent ticker sweeps
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&sweepCount) >= 3
	}, 1*time.Second, 10*time.Millisecond, "expected at least 3 sweep runs (startup + ticker loops)")

	cancel()
	// Allow goroutines to exit cleanly
	time.Sleep(50 * time.Millisecond)
}

func TestSweeperController_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var sweepCount int32
	ms := &mockSweeper{
		duration: 14,
		sweepFn: func() (int, error) {
			atomic.AddInt32(&sweepCount, 1)
			return 0, nil
		},
	}

	testInterval := 20 * time.Millisecond
	c := NewSweeperController(ctx, []sweeper.Interface{ms})
	c.sweepInterval = testInterval

	_, err := c.Sweep()
	require.NoError(t, err)

	// Wait for startup sweep
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&sweepCount) >= 1
	}, 500*time.Millisecond, 10*time.Millisecond)

	// Cancel context to stop sweeper loop
	cancel()
	time.Sleep(50 * time.Millisecond)

	countAfterCancel := atomic.LoadInt32(&sweepCount)
	time.Sleep(60 * time.Millisecond)
	countLater := atomic.LoadInt32(&sweepCount)

	assert.Equal(t, countAfterCancel, countLater, "sweeper loop should not continue after context cancellation")
}

func TestSweeperController_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	expectedErr := errors.New("simulated disk error")
	var errorObserved int32

	ms := &mockSweeper{
		duration: 7,
		sweepFn: func() (int, error) {
			atomic.StoreInt32(&errorObserved, 1)
			return 0, expectedErr
		},
	}

	testInterval := 50 * time.Millisecond
	c := NewSweeperController(ctx, []sweeper.Interface{ms})
	c.sweepInterval = testInterval

	_, err := c.Sweep()
	require.NoError(t, err)

	// Wait for sweep to run and fail
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&errorObserved) == 1
	}, 500*time.Millisecond, 10*time.Millisecond)
}
