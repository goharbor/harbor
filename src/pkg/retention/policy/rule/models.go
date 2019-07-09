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

package rule

// Metadata of the retention rule
type Metadata struct {
	// UUID of rule
	ID int `json:"id"`

	// Priority of rule when doing calculating
	Priority int `json:"priority"`

	// Action of the rule performs
	// "retain"
	Action string `json:"action"`

	// Template ID
	Template string `json:"template"`

	// The parameters of this rule
	Parameters Parameters `json:"params"`

	// Selector attached to the rule for filtering tags
	TagSelectors []*Selector `json:"tag_selectors"`

	// Selector attached to the rule for filtering scope (e.g: repositories or namespaces)
	ScopeSelectors map[string][]*Selector `json:"scope_selectors"`
}

// Selector to narrow down the list
type Selector struct {
	// Kind of the selector
	// "regularExpression" or "label"
	Kind string `json:"kind"`

	// Decorated the selector
	// for "regularExpression" : "matches" and "excludes"
	// for "label" : "with" and "without"
	Decoration string `json:"decoration"`

	// Param for the selector
	Pattern string `json:"pattern"`
}

// Parameters of rule, indexed by the key
type Parameters map[string]Parameter

// Parameter of rule
type Parameter interface{}
