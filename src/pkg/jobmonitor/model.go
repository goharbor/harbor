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

// WorkerPool job service worker pool
type WorkerPool struct {
	ID          string `json:"pool_id"`
	PID         int    `json:"pid"`
	StartAt     int64  `json:"start_at"`
	HeartbeatAt int64  `json:"heartbeat_at"`
	Concurrency int    `json:"concurrency"`
	Host        string `json:"host"`
}

// Worker job service worker
type Worker struct {
	ID        string `json:"id"`
	PoolID    string `json:"pool_id"`
	IsBusy    bool   `json:"is_busy"`
	JobName   string `json:"job_name"`
	JobID     string `json:"job_id"`
	StartedAt int64  `json:"start_at"`
	CheckIn   string `json:"check_in"`
	CheckInAt int64  `json:"check_in_at"`
}

// Queue the job queue
type Queue struct {
	JobType string
	Count   int64
	Latency int64
	Paused  bool
}
