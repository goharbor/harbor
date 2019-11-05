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

package all

// Stats provides the overall progress of the scan all process.
type Stats struct {
	Total uint `json:"total"`
	// Status including `Success`, `Error` or `Stopped` will be counted as completed.
	// This data may be influenced by job retrying
	Completed uint          `json:"completed"`
	Metrics   StatusMetrics `json:"metrics"`
	Requester string        `json:"requester"`
}

// StatusMetrics contains the metrics of each status.
// The key should be the following valid status texts:
//   - "pending"
//   - "running"
//   - "success"
//   - "error"
//   - "stopped"
type StatusMetrics map[string]uint
