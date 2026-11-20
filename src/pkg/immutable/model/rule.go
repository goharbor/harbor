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

package model

import (
	"fmt"
	"sort"

	"github.com/beego/beego/v2/core/validation"

	"github.com/goharbor/harbor/src/lib/pattern"
	"github.com/goharbor/harbor/src/lib/selector/selectors/index"
)

// Metadata of the immutable rule
type Metadata struct {
	// UUID of rule
	ID int64 `json:"id"`

	// ProjectID of project
	ProjectID int64 `json:"project_id"`

	// Disabled rule
	Disabled bool `json:"disabled"`

	// Priority of rule when doing calculating
	Priority int `json:"priority"`

	// Action of the rule performs
	// "immutable"
	Action string `json:"action" valid:"Required"`

	// Template ID
	Template string `json:"template" valid:"Required"`

	// TagSelectors attached to the rule for filtering tags
	TagSelectors []*Selector `json:"tag_selectors" valid:"Required"`

	// Selector attached to the rule for filtering scope (e.g: repositories or namespaces)
	ScopeSelectors map[string][]*Selector `json:"scope_selectors" valid:"Required"`
}

// Valid Valid
func (m *Metadata) Valid(v *validation.Validation) {
	for _, ts := range m.TagSelectors {
		if pass, _ := v.Valid(ts); !pass {
			return
		}
	}
	for _, ss := range m.ScopeSelectors {
		for _, s := range ss {
			if pass, _ := v.Valid(s); !pass {
				return
			}
		}
	}
}

// Validate checks that every selector of the rule can be evaluated by the
// matcher. A selector with an unregistered kind or decoration, or with a
// malformed pattern, makes the matcher fail for every candidate of the project,
// so such a rule must not be stored.
func (m *Metadata) Validate() error {
	for _, s := range m.TagSelectors {
		if err := s.validate(); err != nil {
			return fmt.Errorf("invalid tag selector: %w", err)
		}
	}
	scopes := make([]string, 0, len(m.ScopeSelectors))
	for scope := range m.ScopeSelectors {
		scopes = append(scopes, scope)
	}
	sort.Strings(scopes)
	for _, scope := range scopes {
		for _, s := range m.ScopeSelectors[scope] {
			if err := s.validate(); err != nil {
				return fmt.Errorf("invalid %s scope selector: %w", scope, err)
			}
		}
	}
	return nil
}

// Selector to narrow down the list
type Selector struct {
	// Kind of the selector
	// "doublestar" or "label"
	Kind string `json:"kind" valid:"Required;Match(doublestar)"`

	// Decorated the selector
	// for "doublestar" : "matching" and "excluding"
	// for "label" : "with" and "without"
	Decoration string `json:"decoration" valid:"Required"`

	// Param for the selector
	Pattern string `json:"pattern" valid:"Required"`
}

// validate reports whether the selector can be built and its pattern parsed.
// A nil selector is ignored, the matcher skips it as well.
func (s *Selector) validate() error {
	if s == nil {
		return nil
	}
	if _, err := index.Get(s.Kind, s.Decoration, s.Pattern, ""); err != nil {
		return err
	}
	if err := pattern.ValidateDoublestar(s.Pattern); err != nil {
		return fmt.Errorf("pattern %q: %w", s.Pattern, err)
	}
	return nil
}
