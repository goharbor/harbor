package dao

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/lib/orm"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestPolicy(t *testing.T) {
	p := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Action:   "retain",
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
	p1 := &models.RetentionPolicy{
		ScopeLevel:  p.Scope.Level,
		TriggerKind: p.Trigger.Kind,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	data, _ := json.Marshal(p)
	p1.Data = string(data)

	ctx := orm.Context()
	id, err := CreatePolicy(ctx, p1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	p1, err = GetPolicy(ctx, id)
	assert.Nil(t, err)
	assert.EqualValues(t, "project", p1.ScopeLevel)
	assert.True(t, p1.ID > 0)

	p1.ScopeLevel = "test"
	err = UpdatePolicy(ctx, p1)
	assert.Nil(t, err)
	p1, err = GetPolicy(ctx, id)
	assert.Nil(t, err)
	assert.EqualValues(t, "test", p1.ScopeLevel)

	err = DeletePolicy(ctx, id)
	assert.Nil(t, err)

	p1, err = GetPolicy(ctx, id)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "no row found"))
}
