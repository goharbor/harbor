package policy

import (
	"fmt"
	"testing"

	"github.com/beego/beego/v2/core/validation"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
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

type PolicyTestSuite struct {
	suite.Suite

	policy *Metadata
}

// TestRetentionPolicy is the entry method of running PolicyTestSuite.
func TestRetentionPolicy(t *testing.T) {
	suite.Run(t, &PolicyTestSuite{})
}

// SetupSuite prepares the env for PolicyTestSuite.
func (p *PolicyTestSuite) SetupSuite() {
	p.policy = &Metadata{}
	p.policy.Trigger = &Trigger{}
}

// TearDownSuite clears the env for PolicyTestSuite.
func (p *PolicyTestSuite) TearDownSuite() {
	p.policy = nil
}

func (p *PolicyTestSuite) TestValidateRetentionPolicy() {
	p.policy.Trigger.Kind = TriggerKindSchedule

	// cron is not in the map of trigger setting
	p.NoError(p.policy.ValidateRetentionPolicy())

	// cron value is an empty string
	p.policy.Trigger.Settings = map[string]any{"cron": ""}
	p.NoError(p.policy.ValidateRetentionPolicy())

	//  the 1st field of cron value is not 0
	p.policy.Trigger.Settings = map[string]any{"cron": "1 0 0 1 1 *"}
	p.Error(p.policy.ValidateRetentionPolicy())

	// valid cron value
	p.policy.Trigger.Settings = map[string]any{"cron": "0 0 0 1 1 *"}
	p.NoError(p.policy.ValidateRetentionPolicy())
}

func TestValidateRetentionPolicyRules(t *testing.T) {
	cases := []struct {
		name     string
		rule     rule.Metadata
		errorMsg string
	}{
		{
			name: "rule without parameters",
			rule: rule.Metadata{Template: "always", Parameters: rule.Parameters{}},
		},
		{
			name: "rule with valid parameter",
			rule: rule.Metadata{Template: "nDaysSinceLastPull", Parameters: rule.Parameters{"nDaysSinceLastPull": 30}},
		},
		{
			name:     "empty template",
			rule:     rule.Metadata{Parameters: rule.Parameters{}},
			errorMsg: "empty rule template ID",
		},
		{
			name:     "unknown template",
			rule:     rule.Metadata{Template: "unknown", Parameters: rule.Parameters{}},
			errorMsg: "rule evaluator unknown is not registered",
		},
		{
			name:     "missing required parameter",
			rule:     rule.Metadata{Template: "nDaysSinceLastPull", Parameters: rule.Parameters{"n_days_since_last_pull": 30}},
			errorMsg: "missing required parameter nDaysSinceLastPull for rule nDaysSinceLastPull",
		},
		{
			name:     "nil parameters",
			rule:     rule.Metadata{Template: "latestPushedK"},
			errorMsg: "missing required parameter latestPushedK for rule latestPushedK",
		},
		{
			name:     "invalid parameter value",
			rule:     rule.Metadata{Template: "latestPushedK", Parameters: rule.Parameters{"latestPushedK": -10}},
			errorMsg: "latestPushedK is less than zero",
		},
		{
			name:     "invalid parameter type",
			rule:     rule.Metadata{Template: "nDaysSinceLastPush", Parameters: rule.Parameters{"nDaysSinceLastPush": "30"}},
			errorMsg: "nDaysSinceLastPush type error",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := &Metadata{
				Rules: []rule.Metadata{c.rule},
			}
			err := p.ValidateRetentionPolicy()
			if c.errorMsg == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.True(t, errors.IsErr(err, errors.BadRequestCode))
			require.Contains(t, err.Error(), c.errorMsg)
		})
	}
}

func TestRule(t *testing.T) {
	p := &Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Action:   "retain",
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
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
			Settings: map[string]any{
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
	require.True(t, v.HasErrors())
	require.EqualValues(t, "Kind", v.Errors[0].Field)
	for _, e := range v.Errors {
		fmt.Printf("%s %s\n", e.Field, e.Message)
	}
}

func TestParamValid(t *testing.T) {
	p := &Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Action:   "retain",
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": -10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
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
			Settings: map[string]any{
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
	require.True(t, v.HasErrors())
	require.EqualValues(t, "Parameters", v.Errors[0].Field)

	p = &Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Action:   "retain",
				Template: "nDaysSinceLastPull",
				Parameters: rule.Parameters{
					"nDaysSinceLastPull": 20201010,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
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
			Settings: map[string]any{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &Scope{
			Level:     "project",
			Reference: 1,
		},
	}
	v = &validation.Validation{}
	ok, err = v.Valid(p)
	require.Nil(t, err)
	require.False(t, ok)
	require.True(t, v.HasErrors())
	require.EqualValues(t, "Parameters", v.Errors[0].Field)
}
