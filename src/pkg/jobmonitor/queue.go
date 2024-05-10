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
)

// QueueManager defines the operation related to job service queue
type QueueManager interface {
	List(ctx context.Context, monitClient JobServiceMonitorClient) ([]*Queue, error)
}

type queueManagerImpl struct{}

// NewQueueClient ...
func NewQueueClient() QueueManager {
	return &queueManagerImpl{}
}

func (w *queueManagerImpl) List(_ context.Context, monitClient JobServiceMonitorClient) ([]*Queue, error) {
	resultQueues := make([]*Queue, 0)
	queues, err := monitClient.Queues()
	if err != nil {
		return nil, err
	}
	for _, q := range queues {
		resultQueues = append(resultQueues, &Queue{
			JobType: q.JobName,
			Count:   q.Count,
			Latency: q.Latency,
		})
	}
	return resultQueues, nil
}
