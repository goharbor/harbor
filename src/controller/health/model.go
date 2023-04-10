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

package health

// OverallHealthStatus defines the overall health status of the system
type OverallHealthStatus struct {
	Status     string                   `json:"status"`
	Components []*ComponentHealthStatus `json:"components"`
}

// ComponentHealthStatus defines the specific component health status
type ComponentHealthStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthy bool

func (h healthy) String() string {
	if h {
		return "healthy"
	}
	return "unhealthy"
}
