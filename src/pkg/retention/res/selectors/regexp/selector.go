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

package regexp

import (
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
)

const (
	// Kind ...
	Kind = "regularExpression"
	// Matches [pattern]
	Matches = "matches"
	// Excludes [pattern]
	Excludes = "excludes"
)

// selector for regular expression
type selector struct {
	// Pre defined pattern declarator
	// "matches" and "excludes"
	decoration string
	// The pattern expression
	pattern string
}

// Select candidates by regular expressions
func (s *selector) Select(artifacts []*res.Candidate) ([]*res.Candidate, error) {
	return nil, nil
}

// New is factory method for regexp selector
func New(decoration string, pattern string) res.Selector {
	return &selector{
		decoration: decoration,
		pattern:    pattern,
	}
}

func init() {
	// Register regexp selector
	selectors.Register(Kind, []string{Matches, Excludes}, New)
}
