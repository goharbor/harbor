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

package doublestar

import (
	"github.com/bmatcuk/doublestar"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
)

const (
	// Kind ...
	Kind = "doublestar"
	// Matches [pattern] for tag (default)
	Matches = "matches"
	// Excludes [pattern] for tag (default)
	Excludes = "excludes"
	// RepoMatches represents repository matches [pattern]
	RepoMatches = "repoMatches"
	// RepoExcludes represents repository excludes [pattern]
	RepoExcludes = "repoExcludes"
)

// selector for regular expression
type selector struct {
	// Pre defined pattern declarator
	// "matches", "excludes", "repoMatches" or "repoExcludes"
	decoration string
	// The pattern expression
	pattern string
}

// Select candidates by regular expressions
func (s *selector) Select(artifacts []*res.Candidate) (selected []*res.Candidate, err error) {
	value := ""
	excludes := false

	for _, art := range artifacts {
		switch s.decoration {
		case Matches:
			value = art.Tag
		case Excludes:
			value = art.Tag
			excludes = true
		case RepoMatches:
			value = art.Repository
		case RepoExcludes:
			value = art.Repository
			excludes = true
		}

		if len(value) > 0 {
			matched, err := match(s.pattern, value)
			if err != nil {
				// if error occurred, directly throw it out
				return nil, err
			}

			if (matched && !excludes) || (!matched && excludes) {
				selected = append(selected, art)
			}
		}
	}

	return selected, nil
}

// New is factory method for doublestar selector
func New(decoration string, pattern string) res.Selector {
	return &selector{
		decoration: decoration,
		pattern:    pattern,
	}
}

// match returns whether the str matches the pattern
func match(pattern, str string) (bool, error) {
	if len(pattern) == 0 {
		return true, nil
	}
	return doublestar.Match(pattern, str)
}

func init() {
	// Register doublestar selector
	selectors.Register(Kind, []string{Matches, Excludes}, New)
}
