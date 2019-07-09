package dao

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestPolicy(t *testing.T) {
	p := &policy.Metadata{
		ID:        1,
		Algorithm: "OR",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "recentXdays",
				Parameters: rule.Parameters{
					"num": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "label",
						Decoration: "with",
						Pattern:    "latest",
					},
					{
						Kind:       "regularExpression",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: []*rule.Selector{
					{
						Kind:       "regularExpression",
						Decoration: "matches",
						Pattern:    ".+",
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

	id, err := CreatePolicy(p1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	p1, err = GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "project", p1.ScopeLevel)

	p1.ScopeLevel = "test"
	err = UpdatePolicy(p1)
	assert.Nil(t, err)
	p1, err = GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "test", p1.ScopeLevel)

	err = DeletePolicy(id)
	assert.Nil(t, err)

	p1, err = GetPolicy(id)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "no row found"))
}
