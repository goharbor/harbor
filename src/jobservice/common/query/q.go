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

package query

const (
	// DefaultPageSize defines the default page size
	DefaultPageSize uint = 25
	// ParamKeyPage defines query param key of page number
	ParamKeyPage = "page_number"
	// ParamKeyPageSize defines query param key of page size
	ParamKeyPageSize = "page_size"
	// ParamKeyNonStoppedOnly defines query param key of querying non stopped periodic executions
	ParamKeyNonStoppedOnly = "non_dead_only"
	// ExtraParamKeyNonStoppedOnly defines extra parameter key for querying non stopped periodic executions
	ExtraParamKeyNonStoppedOnly = "NonDeadOnly"
)

// ExtraParameters to keep non pagination query parameters
type ExtraParameters map[string]interface{}

// Set extra parameters
func (ep ExtraParameters) Set(key string, v interface{}) {
	if len(key) > 0 {
		ep[key] = v
	}
}

// Get the extra parameter by key
func (ep ExtraParameters) Get(key string) (interface{}, bool) {
	v, ok := ep[key]

	return v, ok
}

// Parameter for getting executions
type Parameter struct {
	PageNumber uint
	PageSize   uint
	Extras     ExtraParameters
}
