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

package lib

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// ErrRetryTimeout timeout error for retrying
	ErrRetryTimeout = errors.New("retry timeout")
)

// RetryOptions options for the retry functions
type RetryOptions struct {
	InitialInterval time.Duration                        // the initial interval for retring after failure, default 100 milliseconds
	MaxInterval     time.Duration                        // the max interval for retring after failure, default 1 second
	Timeout         time.Duration                        // the total time before returning if something is wrong, default 1 minute
	Callback        func(err error, sleep time.Duration) // the callback function for Retry when the f called failed
}

// RetryOption ...
type RetryOption func(*RetryOptions)

// RetryInitialInterval set initial interval
func RetryInitialInterval(initial time.Duration) RetryOption {
	return func(opts *RetryOptions) {
		opts.InitialInterval = initial
	}
}

// RetryMaxInterval set max interval
func RetryMaxInterval(max time.Duration) RetryOption {
	return func(opts *RetryOptions) {
		opts.MaxInterval = max
	}
}

// RetryTimeout set timeout interval
func RetryTimeout(timeout time.Duration) RetryOption {
	return func(opts *RetryOptions) {
		opts.Timeout = timeout
	}
}

// RetryCallback set callback
func RetryCallback(callback func(err error, sleep time.Duration)) RetryOption {
	return func(opts *RetryOptions) {
		opts.Callback = callback
	}
}

// RetryUntil retry until f run successfully or timeout
//
// NOTE: This function will use exponential backoff and jitter for retrying, see
// https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/ for more information
func RetryUntil(f func() error, options ...RetryOption) error {
	opts := &RetryOptions{}

	for _, o := range options {
		o(opts)
	}

	if opts.InitialInterval <= 0 {
		opts.InitialInterval = time.Millisecond * 100
	}

	if opts.MaxInterval <= 0 {
		opts.MaxInterval = time.Second
	}

	if opts.Timeout <= 0 {
		opts.Timeout = time.Minute
	}

	timeout := time.After(opts.Timeout)
	for attempt := 1; ; attempt++ {
		select {
		case <-timeout:
			return ErrRetryTimeout
		default:
			if err := f(); err != nil {
				sleep := getBackoff(attempt, opts.InitialInterval, opts.MaxInterval, true)
				if opts.Callback != nil {
					opts.Callback(err, sleep)
				}

				time.Sleep(sleep)
			} else {
				return nil
			}
		}
	}
}

func getBackoff(attempt int, initialInterval, maxInterval time.Duration, equalJitter bool) time.Duration {
	max := float64(maxInterval)
	base := float64(initialInterval)

	dur := base * math.Pow(2, float64(attempt))
	if equalJitter {
		dur = dur/2 + float64(rand.Int63n(int64(dur))/2)
	}

	if dur < base {
		dur = base
	}

	if dur > max {
		dur = max
	}

	return time.Duration(dur)
}
