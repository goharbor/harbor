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

type void = struct{}

// Set a simple set
type Set map[interface{}]void

// Add add item to set
func (s Set) Add(item interface{}) {
	s[item] = void{}
}

// Exists returns true when item in the set
func (s Set) Exists(item interface{}) bool {
	_, ok := s[item]

	return ok
}

// Items returns the items in the set
func (s Set) Items() []interface{} {
	var items []interface{}
	for item := range s {
		items = append(items, item)
	}

	return items
}
