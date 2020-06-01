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

package base

// variable definitions
var (
	// API version v2.0
	APIVersionV20 = "v2.0"
	// API version v2.1
	APIVersionV21 = "v2.1"
	// The commended API version in current deployment
	RecommendedAPIVersion = APIVersionV21
	// All available API versions maintained by the current deployment
	AvailableAPIVersions = []string{APIVersionV20, APIVersionV21}
)
