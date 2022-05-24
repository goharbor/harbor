package immutable

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	htesting.Suite
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
	s.ctr = Ctr
}

func (s *ControllerTestSuite) TestImmutableRule() {

	var err error
	ctx := s.Context()

	projectID, err := pkg.ProjectMgr.Create(ctx, &proModels.Project{
		Name:    "testimmutablerule",
		OwnerID: 1,
	})
	if s.Nil(err) {
		defer pkg.ProjectMgr.Delete(ctx, projectID)
	}

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
	s.ruleID, err = s.ctr.CreateImmutableRule(orm.Context(), rule)
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
	err = s.ctr.UpdateImmutableRule(orm.Context(), projectID, update)
	s.require.Nil(err)

	getRule, err := s.ctr.GetImmutableRule(orm.Context(), s.ruleID)
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
	err = s.ctr.UpdateImmutableRule(orm.Context(), projectID, update2)
	s.require.Nil(err)
	getRule, err = s.ctr.GetImmutableRule(orm.Context(), s.ruleID)
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
	s.ruleID, err = s.ctr.CreateImmutableRule(orm.Context(), rule2)
	s.require.Nil(err)

	rules, err := s.ctr.ListImmutableRules(orm.Context(), q.New(q.KeyWords{"ProjectID": projectID}))
	s.require.Nil(err)
	s.require.Equal(2, len(rules))

}

// TearDownSuite clears env for test suite
func (s *ControllerTestSuite) TearDownSuite() {
	err := s.ctr.DeleteImmutableRule(orm.Context(), s.ruleID)
	require.NoError(s.T(), err, "delete immutable rule")
}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}
