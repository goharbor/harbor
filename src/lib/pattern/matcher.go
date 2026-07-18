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

package pattern

import (
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// KindRegex indicates regular expression pattern matching
	KindRegex = "regex"
	// KindDoublestar indicates doublestar (glob) pattern matching
	KindDoublestar = "doublestar"
)

// RepositoryFilter represents a repository filter configuration with pattern and kind
type RepositoryFilter struct {
	// Filter is the pattern expression to match against
	Filter string
	// Kind is the type of pattern matching: "regex" or "doublestar"
	Kind string
}

// NewRepositoryFilter creates a RepositoryFilter from separate pattern and kind strings.
// If kind is empty, it defaults to KindDoublestar.
func NewRepositoryFilter(filterPattern, kind string) *RepositoryFilter {
	return &RepositoryFilter{
		Filter: strings.TrimSpace(filterPattern),
		Kind:   strings.TrimSpace(kind),
	}
}

// Match returns true if the value matches the filter pattern.
// Returns true if the filter is empty.
func (rf *RepositoryFilter) Match(value string) (bool, error) {
	if rf == nil || rf.Filter == "" {
		return true, nil
	}
	matcher := NewMatcher(rf.Kind)
	return matcher.Match(value, rf.Filter)
}

// ValidateRepositoryFilter validates the repository filter kind and pattern.
// Empty pattern is valid and means all repositories are allowed.
func ValidateRepositoryFilter(filterPattern, kind string) error {
	rf := NewRepositoryFilter(filterPattern, kind)
	if rf.Kind != "" && rf.Kind != KindRegex && rf.Kind != KindDoublestar {
		return errors.Errorf("unsupported repository filter kind %q", kind)
	}
	_, err := rf.Match("")
	return err
}

// Matcher is an interface for matching strings against patterns
type Matcher interface {
	// Match returns true if the value matches the pattern
	Match(value, pattern string) (bool, error)
}

// RegexMatcher implements Matcher using regular expressions
type RegexMatcher struct{}

// Match matches value against a regular expression pattern
func (r *RegexMatcher) Match(value, pattern string) (bool, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	match := re.FindStringIndex(value)
	return match != nil && match[0] == 0 && match[1] == len(value), nil
}

// DoublestarMatcher implements Matcher using doublestar (glob) patterns
type DoublestarMatcher struct{}

// Match matches value against a doublestar (glob) pattern
func (d *DoublestarMatcher) Match(value, pattern string) (bool, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true, nil
	}
	return doublestar.Match(pattern, value)
}

// NewMatcher creates a new Matcher based on the specified kind.
// Defaults to KindDoublestar when kind is empty or unrecognised.
func NewMatcher(kind string) Matcher {
	switch kind {
	case KindRegex:
		return &RegexMatcher{}
	case KindDoublestar:
		fallthrough
	default:
		return &DoublestarMatcher{}
	}
}
