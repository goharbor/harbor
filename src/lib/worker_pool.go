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

// NewWorkerPool creates a new worker pool with specified size
func NewWorkerPool(size int32) *WorkerPool {
	wp := &WorkerPool{}
	wp.queue = make(chan struct{}, size)
	return wp
}

// WorkerPool controls the concurrency limit of task/process
type WorkerPool struct {
	queue chan struct{}
}

// GetWorker hangs until a free worker available
func (w *WorkerPool) GetWorker() {
	w.queue <- struct{}{}
}

// ReleaseWorker hangs until the worker return back into the pool
func (w *WorkerPool) ReleaseWorker() {
	<-w.queue
}
