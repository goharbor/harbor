package immutabletag

import (
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	ctr     Controller
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions

	ruleID int64
}

// SetupSuite ...
func (s *ControllerTestSuite) SetupSuite() {
	test.InitDatabaseFromEnv()
	s.t = s.T()
	s.assert = assert.New(s.t)
	s.require = require.New(s.t)
	s.ctr = ImmuCtr
}

func (s *ControllerTestSuite) TestImmutableRule() {

	var err error

	projectID, err := dao.AddProject(models.Project{
		Name:    "TestImmutableRule",
		OwnerID: 1,
	})

	rule := &model.Metadata{
		ProjectID: projectID,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-**",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "redis",
				},
			},
		},
	}
	s.ruleID, err = s.ctr.CreateImmutableRule(rule)
	s.require.Nil(err)

	update := &model.Metadata{
		ID:        s.ruleID,
		ProjectID: projectID,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-**",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "postgres",
				},
			},
		},
		Disabled: false,
	}
	err = s.ctr.UpdateImmutableRule(projectID, update)
	s.require.Nil(err)

	getRule, err := s.ctr.GetImmutableRule(s.ruleID)
	s.require.Nil(err)
	s.require.Equal("postgres", getRule.ScopeSelectors["repository"][0].Pattern)

	update2 := &model.Metadata{
		ID:        s.ruleID,
		ProjectID: projectID,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-**",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "postgres",
				},
			},
		},
		Disabled: true,
	}
	err = s.ctr.UpdateImmutableRule(projectID, update2)
	s.require.Nil(err)
	getRule, err = s.ctr.GetImmutableRule(s.ruleID)
	s.require.Nil(err)
	s.require.True(getRule.Disabled)

	rule2 := &model.Metadata{
		ProjectID: projectID,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "latest",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "redis",
				},
			},
		},
	}
	s.ruleID, err = s.ctr.CreateImmutableRule(rule2)
	s.require.Nil(err)

	rules, err := s.ctr.ListImmutableRules(projectID)
	s.require.Nil(err)
	s.require.Equal(2, len(rules))

}

// TearDownSuite clears env for test suite
func (s *ControllerTestSuite) TearDownSuite() {
	err := s.ctr.DeleteImmutableRule(s.ruleID)
	require.NoError(s.T(), err, "delete immutable rule")
}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
