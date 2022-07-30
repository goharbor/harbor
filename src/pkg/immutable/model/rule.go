package model

import (
	"github.com/beego/beego/validation"
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
