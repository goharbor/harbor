package rule

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MatchTestSuite ...
type MatchTestSuite struct {
	suite.Suite
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions
	ctr     immutable.Controller
	ruleID  int64
	ruleID2 int64
}

// SetupSuite ...
func (s *MatchTestSuite) SetupSuite() {
	s.t = s.T()
	s.assert = assert.New(s.t)
	s.require = require.New(s.t)
	s.ctr = immutable.Ctr
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

	id, err := s.ctr.CreateImmutableRule(orm.Context(), rule)
	s.ruleID = id
	s.require.Nil(err)

	id, err = s.ctr.CreateImmutableRule(orm.Context(), rule2)
	s.ruleID2 = id
	s.require.Nil(err)

	match := NewRuleMatcher()

	c1 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "redis",
		Tags:        []string{"release-1.10"},
	}
	isMatch, err := match.Match(orm.Context(), 1, c1)
	s.require.Equal(isMatch, true)
	s.require.Nil(err)

	c2 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "redis",
		Tags:        []string{"1.10"},
		Kind:        selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 1, c2)
	s.require.Equal(isMatch, false)
	s.require.Nil(err)

	c3 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "immutable",
		Repository:  "mysql",
		Tags:        []string{"9.4.8"},
		Kind:        selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 1, c3)
	s.require.Equal(isMatch, true)
	s.require.Nil(err)

	c4 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "immutable",
		Repository:  "hello",
		Tags:        []string{"world"},
		Kind:        selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 1, c4)
	s.require.Equal(isMatch, false)
	s.require.Nil(err)

	// untagged case
	c5 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "redis",
		// no tags
		Tags: []string{},
		Kind: selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 1, c5)
	s.require.Equal(isMatch, false)
	s.require.Nil(err)
}

// TearDownSuite clears env for test suite
func (s *MatchTestSuite) TearDownSuite() {
	err := s.ctr.DeleteImmutableRule(orm.Context(), s.ruleID)
	require.NoError(s.T(), err, "delete immutable")

	err = s.ctr.DeleteImmutableRule(orm.Context(), s.ruleID2)
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
