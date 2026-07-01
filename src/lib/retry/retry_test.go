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

package retry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAbort(t *testing.T) {
	assert := assert.New(t)

	e1 := Abort(nil)
	assert.Equal("retry abort", e1.Error())

	e2 := Abort(fmt.Errorf("failed to call func"))
	assert.Equal("retry abort, error: failed to call func", e2.Error())
}

func TestRetry(t *testing.T) {
	assert := assert.New(t)

	i := 0
	f1 := func() error {
		i++
		return fmt.Errorf("failed")
	}
	assert.Error(Retry(f1, InitialInterval(time.Second), MaxInterval(time.Second), Timeout(time.Second*5)))
	// f1 called time     0s - sleep - 1s - sleep - 2s - sleep - 3s - sleep - 4s - sleep - 5s
	// i after f1 called  1            2            3            4            5            6
	// the i may be 5 or 6 depend on timeout or default which is seleted by the select statement
	assert.LessOrEqual(i, 6)

	f2 := func() error {
		return nil
	}
	assert.Nil(Retry(f2))

	i = 0
	f3 := func() error {
		defer func() {
			i++
		}()

		if i < 2 {
			return fmt.Errorf("failed")
		}
		return nil
	}
	assert.Nil(Retry(f3))

	Retry(
		f1,
		Timeout(time.Second*5),
		Callback(func(err error, sleep time.Duration) {
			fmt.Printf("failed to exec f1 retry after %s : %v\n", sleep, err)
		}),
	)

	err := Retry(func() error {
		return fmt.Errorf("always failed")
	})

	assert.Error(err)
	assert.Equal("retry timeout: always failed", err.Error())

	i = 0
	f4 := func() error {
		if i == 3 {
			return Abort(fmt.Errorf("abort"))
		}

		i++
		return fmt.Errorf("error")
	}
	assert.Error(Retry(f4, InitialInterval(time.Second), MaxInterval(time.Second), Timeout(time.Second*5)))
	assert.LessOrEqual(i, 3)
}

func TestRetryContextAlreadyCanceled(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	start := time.Now()
	err := Retry(func() error {
		calls++
		return fmt.Errorf("should not be called")
	}, Context(ctx), Timeout(5*time.Second))

	assert.Error(err)
	assert.ErrorIs(err, context.Canceled)
	assert.Equal(0, calls, "f should not be called when ctx is already canceled")
	assert.Less(time.Since(start), 100*time.Millisecond, "Retry must return immediately on pre-canceled ctx")
}

func TestRetryContextCancelMidLoop(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	start := time.Now()

	// Cancel after 200ms so we leave retry mid-backoff, well before its 5s Timeout.
	time.AfterFunc(200*time.Millisecond, cancel)

	err := Retry(func() error {
		calls++
		return fmt.Errorf("transient")
	}, Context(ctx),
		InitialInterval(50*time.Millisecond),
		MaxInterval(100*time.Millisecond),
		Timeout(5*time.Second),
	)

	elapsed := time.Since(start)
	assert.Error(err)
	assert.ErrorIs(err, context.Canceled)
	assert.Less(elapsed, time.Second, "Retry must observe mid-loop cancellation, not wait for Timeout: elapsed=%s", elapsed)
	assert.Greater(calls, 0)
}

func TestRetryTerminatesWhenFReturnsContextCanceled(t *testing.T) {
	assert := assert.New(t)

	calls := 0
	start := time.Now()
	err := Retry(func() error {
		calls++
		return context.Canceled
	}, Timeout(5*time.Second))

	elapsed := time.Since(start)
	assert.ErrorIs(err, context.Canceled)
	assert.Equal(1, calls, "f should be called exactly once — context errors are terminal")
	assert.Less(elapsed, 100*time.Millisecond, "Retry must not sleep after a ctx error: elapsed=%s", elapsed)
}

func TestRetryTerminatesWhenFReturnsDeadlineExceeded(t *testing.T) {
	assert := assert.New(t)

	calls := 0
	err := Retry(func() error {
		calls++
		return fmt.Errorf("wrapped: %w", context.DeadlineExceeded)
	}, Timeout(5*time.Second))

	assert.ErrorIs(err, context.DeadlineExceeded)
	assert.Equal(1, calls, "wrapped ctx errors should also be terminal (via errors.Is)")
}

func TestRetryCancellableSleep(t *testing.T) {
	assert := assert.New(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// First call returns a normal error so we enter the inter-attempt sleep.
	// We configure a 2s min/max interval, then cancel the ctx 100ms in.
	time.AfterFunc(100*time.Millisecond, cancel)

	start := time.Now()
	err := Retry(func() error {
		return fmt.Errorf("transient")
	}, Context(ctx),
		InitialInterval(2*time.Second),
		MaxInterval(2*time.Second),
		Timeout(10*time.Second),
	)
	elapsed := time.Since(start)

	assert.ErrorIs(err, context.Canceled)
	assert.Less(elapsed, time.Second, "the 2s sleep must be interrupted by ctx cancellation: elapsed=%s", elapsed)
}
