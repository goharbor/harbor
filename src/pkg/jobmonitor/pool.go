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

package jobmonitor

import (
	"context"

	"github.com/gocraft/work"
)

var _ JobServiceMonitorClient = (*work.Client)(nil)

// JobServiceMonitorClient the interface to retrieve job service monitor metrics
type JobServiceMonitorClient interface {
	// WorkerPoolHeartbeats retrieves worker pool heartbeats
	WorkerPoolHeartbeats() ([]*work.WorkerPoolHeartbeat, error)
	// WorkerObservations retrieves worker observations
	WorkerObservations() ([]*work.WorkerObservation, error)
	// Queues retrieves the job queue information
	Queues() ([]*work.Queue, error)
}

// PoolManager the interface to retrieve job service monitor metrics
type PoolManager interface {
	// List retrieves pools information
	List(ctx context.Context, monitorClient JobServiceMonitorClient) ([]*WorkerPool, error)
}

type poolManager struct{}

// NewPoolManager create a PoolManager with namespace and redis Pool
func NewPoolManager() PoolManager {
	return &poolManager{}
}

func (p poolManager) List(_ context.Context, monitorClient JobServiceMonitorClient) ([]*WorkerPool, error) {
	workerPool := make([]*WorkerPool, 0)
	wh, err := monitorClient.WorkerPoolHeartbeats()
	if err != nil {
		return workerPool, err
	}
	for _, w := range wh {
		wp := &WorkerPool{
			ID:          w.WorkerPoolID,
			PID:         w.Pid,
			StartAt:     w.StartedAt,
			Concurrency: int(w.Concurrency),
			Host:        w.Host,
			HeartbeatAt: w.HeartbeatAt,
		}
		workerPool = append(workerPool, wp)
	}
	return workerPool, nil
}
