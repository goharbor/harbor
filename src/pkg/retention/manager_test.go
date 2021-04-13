package retention

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestPolicy(t *testing.T) {
	m := NewManager()
	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
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
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	ctx := orm.Context()
	id, err := m.CreatePolicy(ctx, p1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	p1, err = m.GetPolicy(ctx, id)
	assert.Nil(t, err)
	assert.EqualValues(t, "project", p1.Scope.Level)
	assert.True(t, p1.ID > 0)

	p1.Scope.Level = "test"
	err = m.UpdatePolicy(ctx, p1)
	assert.Nil(t, err)
	p1, err = m.GetPolicy(ctx, id)
	assert.Nil(t, err)
	assert.EqualValues(t, "test", p1.Scope.Level)

	err = m.DeletePolicy(ctx, id)
	assert.Nil(t, err)

	p1, err = m.GetPolicy(ctx, id)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such Retention policy"))
	assert.Nil(t, p1)
}

func TestExecution(t *testing.T) {
	m := NewManager()
	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
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
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	ctx := orm.Context()
	policyID, err := m.CreatePolicy(ctx, p1)
	assert.Nil(t, err)
	assert.True(t, policyID > 0)
}
