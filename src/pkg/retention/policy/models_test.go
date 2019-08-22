package policy

import (
	"fmt"
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAlgorithm(t *testing.T) {
	p := Metadata{
		Algorithm: "bad",
	}
	v := &validation.Validation{}
	ok, err := v.Valid(p)
	require.Nil(t, err)
	require.False(t, ok)
}

// beego 1.9.0 has bug to deal with nil interface
// func TestTrigger(t *testing.T) {
// 	p := Metadata{
// 		Algorithm: "or",
// 	}
// 	v := &validation.Validation{}
// 	ok, err := v.Valid(p)
//
// 	require.Nil(t, err)
// 	require.False(t, ok)
// 	for _, e := range v.Errors {
// 		fmt.Print(e.Field, e.Message)
// 	}
// }

func TestRule(t *testing.T) {
	p := &Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Action:   "retain",
				Template: "recentXdays",
				Parameters: rule.Parameters{
					"num": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "badkind", // validate doublestar
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &Scope{
			Level:     "project",
			Reference: 1,
		},
	}
	v := &validation.Validation{}
	ok, err := v.Valid(p)

	require.Nil(t, err)
	require.False(t, ok)
	for _, e := range v.Errors {
		fmt.Printf("%s %s\n", e.Field, e.Message)
	}
}
