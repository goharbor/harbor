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

// ValidateRepositoryFilter validates the repository filter kind and pattern.
// Empty pattern is valid and means all repositories are allowed.
func ValidateRepositoryFilter(filterPattern, kind string) error {
	_, err := Match("", filterPattern, kind)
	return err
}

// ValidateKind validates the repository filter kind. Empty kind is valid and
// means the doublestar default.
func ValidateKind(kind string) error {
	switch kind {
	case "", KindRegex, KindDoublestar:
		return nil
	default:
		return errors.Errorf("unsupported repository filter kind %q, must be %q or %q", kind, KindDoublestar, KindRegex)
	}
}

// Match returns true if the value matches the pattern according to the kind.
// Empty pattern matches all (returns true).
func Match(value, filterPattern, kind string) (bool, error) {
	filterPattern = strings.TrimSpace(filterPattern)
	if filterPattern == "" {
		return true, nil
	}
	switch kind {
	case KindRegex:
		// Compile the bare pattern first: a group-imbalanced pattern such as
		// `foo)|(bar` fails alone but compiles inside the ^(?:...)$ wrapper,
		// where the top-level alternation silently destroys the anchoring.
		if _, err := regexp.Compile(filterPattern); err != nil {
			return false, err
		}
		re, err := regexp.Compile("^(?:" + filterPattern + ")$")
		if err != nil {
			return false, err
		}
		return re.MatchString(value), nil
	case KindDoublestar, "":
		if !validateDoublestarPattern(filterPattern) {
			return false, doublestar.ErrBadPattern
		}
		return doublestar.Match(filterPattern, value)
	default:
		return false, errors.Errorf("unsupported repository filter kind %q", kind)
	}
}

// validateDoublestarPattern is a port of doValidatePattern from
// github.com/bmatcuk/doublestar/v4; the v1 dependency in go.mod has no
// ValidatePattern. Keep it in sync with the library's grammar if the
// dependency is ever bumped.
func validateDoublestarPattern(s string) bool {
	altDepth := 0
	l := len(s)
VALIDATE:
	for i := 0; i < l; i++ {
		switch s[i] {
		case '\\':
			// skip the next byte - return false if there is no next byte
			if i++; i >= l {
				return false
			}
			continue

		case '[':
			if i++; i >= l {
				// class didn't end
				return false
			}
			if s[i] == '^' || s[i] == '!' {
				i++
			}
			if i >= l || s[i] == ']' {
				// class didn't end or empty character class
				return false
			}

			for ; i < l; i++ {
				if s[i] == '\\' {
					i++
				} else if s[i] == ']' {
					// looks good
					continue VALIDATE
				}
			}

			// class didn't end
			return false

		case '{':
			altDepth++
			continue

		case '}':
			if altDepth == 0 {
				// alt end without a corresponding start
				return false
			}
			altDepth--
			continue
		}
	}

	// valid as long as all alts are closed
	return altDepth == 0
}
