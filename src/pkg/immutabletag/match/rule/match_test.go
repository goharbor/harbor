package rule

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
)

// MatchTestSuite ...
type MatchTestSuite struct {
	suite.Suite
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions
	ctr     immutabletag.APIController
	ruleID  int64
	ruleID2 int64
}

// SetupSuite ...
func (s *MatchTestSuite) SetupSuite() {
	s.t = s.T()
	s.assert = assert.New(s.t)
	s.require = require.New(s.t)
	s.ctr = immutabletag.ImmuCtr
}

func (s *MatchTestSuite) TestImmuMatch() {
	rule := &model.Metadata{
		ProjectID: 1,
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
	rule2 := &model.Metadata{
		ProjectID: 1,
		Priority:  1,
		Template:  "immutable_template",
		Action:    "immuablity",
		TagSelectors: []*model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "**",
			},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "mysql",
				},
			},
		},
	}

	id, err := s.ctr.CreateImmutableRule(rule)
	s.ruleID = id
	s.require.Nil(err)

	id, err = s.ctr.CreateImmutableRule(rule2)
	s.ruleID2 = id
	s.require.Nil(err)

	match := NewRuleMatcher(1)

	c1 := art.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "redis",
		Tag:         "release-1.10",
	}
	isMatch, err := match.Match(c1)
	s.require.Equal(isMatch, true)
	s.require.Nil(err)

	c2 := art.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "redis",
		Tag:         "1.10",
		Kind:        art.Image,
	}
	isMatch, err = match.Match(c2)
	s.require.Equal(isMatch, false)
	s.require.Nil(err)

	c3 := art.Candidate{
		NamespaceID: 1,
		Namespace:   "immutable",
		Repository:  "mysql",
		Tag:         "9.4.8",
		Kind:        art.Image,
	}
	isMatch, err = match.Match(c3)
	s.require.Equal(isMatch, true)
	s.require.Nil(err)

	c4 := art.Candidate{
		NamespaceID: 1,
		Namespace:   "immutable",
		Repository:  "hello",
		Tag:         "world",
		Kind:        art.Image,
	}
	isMatch, err = match.Match(c4)
	s.require.Equal(isMatch, false)
	s.require.Nil(err)
}

// TearDownSuite clears env for test suite
func (s *MatchTestSuite) TearDownSuite() {
	err := s.ctr.DeleteImmutableRule(s.ruleID)
	require.NoError(s.T(), err, "delete immutable")

	err = s.ctr.DeleteImmutableRule(s.ruleID2)
	require.NoError(s.T(), err, "delete immutable")
}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()

	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(MatchTestSuite))
}
