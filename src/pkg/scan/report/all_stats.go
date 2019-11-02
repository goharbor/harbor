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

package report

// AllStats provides the overall progress of the scan all process.
type AllStats struct {
	Total     int            `json:"total"`
	Completed int            `json:"completed"` // status is `Success`, `Error` or `Stopped`.
	Progress  *StatusMetrics `json:"progress"`
}

// StatusMetrics contains the metrics of each status.
type StatusMetrics struct {
	Pending int `json:"pending"`
	Running int `json:"running"`
	Success int `json:"success"`
	Error   int `json:"error"`
	Stopped int `json:"stopped"`
}
