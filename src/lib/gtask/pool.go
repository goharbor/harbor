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
	"sync"
	"time"
)

func DefaultPool() *Pool {
	return pool
}

var (
	// pool is the global task pool.
	pool = NewPool()
)

type taskFunc func(ctx context.Context)

// Pool is the task pool for managing some async jobs.
type Pool struct {
	stopCh chan struct{}
	wg     sync.WaitGroup
	lock   sync.Mutex
	tasks  []*task
}

func NewPool() *Pool {
	return &Pool{
		stopCh: make(chan struct{}),
	}
}

type task struct {
	fn       taskFunc
	interval time.Duration
}

func (p *Pool) AddTask(fn taskFunc, interval time.Duration) {
	t := &task{
		fn:       fn,
		interval: interval,
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	p.tasks = append(p.tasks, t)
}

func (p *Pool) Start(ctx context.Context) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, task := range p.tasks {
		p.wg.Add(1)
		go p.doTask(ctx, task)
	}
}

func (p *Pool) doTask(ctx context.Context, task *task) {
	defer p.wg.Done()
	for {
		select {
		// wait for stop signal
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
			task.fn(ctx)
			// interval is 0 means it's a one time job, return directly
			if task.interval == 0 {
				return
			}
			time.Sleep(task.interval)
		}
	}
}

func (p *Pool) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}
