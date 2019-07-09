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

package list

import "github.com/goharbor/harbor/src/pkg/retention/res"

const (
	// InRepos for in [repositories]
	InRepos = "in repositories"
	// NotInRepos for not in [repositories]
	NotInRepos = "not in repositories"
	// InTags for in [tags]
	InTags = "in tags"
	// NotInTags for not in [tags]
	NotInTags = "not in tags"
)

// selector for regular expression
type selector struct {
	// Pre defined pattern declarator
	// "InRepo", "NotInRepo", "InTag" and "NotInTags"
	decoration string
	// The item list
	values []string
}

// Select candidates by regular expressions
func (s *selector) Select(artifacts []*res.Candidate) ([]*res.Candidate, error) {
	return nil, nil
}

// New is factory method for list selector
func New(decoration string, pattern interface{}) res.Selector {
	return &selector{
		decoration: decoration,
		values:     pattern.([]string),
	}
}
