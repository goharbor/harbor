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
	"github.com/goharbor/harbor/src/pkg/artifactselector"
)

const (
	// Kind ...
	Kind = "doublestar"
	// Matches [pattern] for tag (default)
	Matches = "matches"
	// Excludes [pattern] for tag (default)
	Excludes = "excludes"
	// UNAGGED [pattern] for tag (default)
	UNAGGED = "untagged"
	// RepoMatches represents repository matches [pattern]
	RepoMatches = "repoMatches"
	// RepoExcludes represents repository excludes [pattern]
	RepoExcludes = "repoExcludes"
	// NSMatches represents namespace matches [pattern]
	NSMatches = "nsMatches"
	// NSExcludes represents namespace excludes [pattern]
	NSExcludes = "nsExcludes"
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
func (s *selector) Select(artifacts []*artifactselector.Candidate) (selected []*artifactselector.Candidate, err error) {
	value := ""
	excludes := false

	for _, art := range artifacts {
		switch s.decoration {
		case Matches:
			s, err := s.tagSelectMatch(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case Excludes:
			s, err := s.tagSelectExclude(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case UNAGGED:
			s, err := s.tagSelectUntagged(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case RepoMatches:
			value = art.Repository
		case RepoExcludes:
			value = art.Repository
			excludes = true
		case NSMatches:
			value = art.Namespace
		case NSExcludes:
			value = art.Namespace
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

func (s *selector) tagSelectMatch(artifact *artifactselector.Candidate) (selected bool, err error) {
	for _, t := range artifact.Tags {
		matched, err := match(s.pattern, t)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

func (s *selector) tagSelectExclude(artifact *artifactselector.Candidate) (selected bool, err error) {
	for _, t := range artifact.Tags {
		matched, err := match(s.pattern, t)
		if err != nil {
			return false, err
		}
		if !matched {
			return true, nil
		}
	}
	return false, nil
}

func (s *selector) tagSelectUntagged(artifact *artifactselector.Candidate) (selected bool, err error) {
	return len(artifact.Tags) == 0, nil
}

// New is factory method for doublestar selector
func New(decoration string, pattern string) artifactselector.Selector {
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
