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

package gtask

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddTask(t *testing.T) {
	pool := NewPool()

	taskNum := 3
	taskInterval := time.Duration(0)
	for i := 0; i < taskNum; i++ {
		fn := func(ctx context.Context) {
			t.Logf("Task %d is running...", i)
		}

		pool.AddTask(fn, taskInterval)
	}

	if len(pool.tasks) != taskNum {
		t.Errorf("Expected %d tasks but found %d", taskNum, len(pool.tasks))
	}
}

func TestStartAndStop(t *testing.T) {
	// test normal case
	{
		pool := NewPool()
		// create channel with buffer
		ch1 := make(chan struct{}, 5)
		ch2 := make(chan struct{}, 5)
		// test one-time job
		t1 := &task{
			interval: 0,
			fn: func(ctx context.Context) {
				ch1 <- struct{}{}
			},
		}
		// test interval job
		t2 := &task{
			interval: 100 * time.Millisecond,
			fn: func(ctx context.Context) {
				ch2 <- struct{}{}
			},
		}

		pool.tasks = []*task{t1, t2}

		ctx1, cancel1 := context.WithCancel(context.Background())
		defer cancel1()
		pool.Start(ctx1)

		// Let it run for a bit
		time.Sleep(300 * time.Millisecond)
		// ch1 should only have one element as it's a one time job
		assert.Equal(t, 1, len(ch1))
		// ch2 should have elements over 2 as sleep 300ms and interval is 100ms
		assert.Greater(t, len(ch2), 2)
		pool.Stop()
		close(ch1)
		close(ch2)
	}

	// test context timeout case
	{
		pool := NewPool()
		ch1 := make(chan struct{}, 2)
		t1 := &task{
			interval: 100 * time.Millisecond,
			fn: func(ctx context.Context) {
				ch1 <- struct{}{}
			},
		}

		pool.tasks = []*task{t1}
		ctx1, cancel1 := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel1()
		pool.Start(ctx1)
		// Let it run for a bit
		time.Sleep(200 * time.Millisecond)
		assert.Equal(t, 1, len(ch1))
		pool.Stop()
		close(ch1)
	}
}
