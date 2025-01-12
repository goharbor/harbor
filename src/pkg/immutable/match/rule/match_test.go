package rule

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	policyindex "github.com/goharbor/harbor/src/pkg/retention/policy/rule/index"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"os"
	"testing"
	"time"
)

// MatchTestSuite ...
type MatchTestSuite struct {
	suite.Suite
	t       *testing.T
	assert  *assert.Assertions
	require *require.Assertions
	ctr     immutable.Controller
	ruleIDs []int64
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

	// any tag in postgres repo pushed within last 5 days should be immutable
	rule3 := &model.Metadata{
		ProjectID: 1,
		Priority:  1,
		Template:  "nDaysSinceLastPush",
		Action:    "immutability",
		Parameters: map[string]model.Parameter{
			"nDaysSinceLastPush": 4,
		},
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
					Pattern:    "postgres",
				},
			},
		},
	}

	// nginx repo pilled within last 2 days
	rule4 := &model.Metadata{
		ProjectID: 1,
		Priority:  1,
		Template:  "nDaysSinceLastPull",
		Parameters: map[string]model.Parameter{
			"nDaysSinceLastPull": 2,
		},
		Action: "immutability",
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
					Pattern:    "nginx",
				},
			},
		},
	}

	id, err := s.ctr.CreateImmutableRule(orm.Context(), rule)
	s.ruleIDs = append(s.ruleIDs, id)
	s.require.Nil(err)

	id, err = s.ctr.CreateImmutableRule(orm.Context(), rule2)
	s.ruleIDs = append(s.ruleIDs, id)
	s.require.Nil(err)

	id, err = s.ctr.CreateImmutableRule(orm.Context(), rule3)
	s.ruleIDs = append(s.ruleIDs, id)
	s.require.Nil(err)

	id, err = s.ctr.CreateImmutableRule(orm.Context(), rule4)
	s.ruleIDs = append(s.ruleIDs, id)
	s.require.Nil(err)

	match := NewRuleMatcher(policyindex.Get)

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

	//

	// conditional cases
	// postgres pushed 10d ago
	c6 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "postgres",
		// no tags
		Tags: []string{
			"latest",
		},
		PushedTime: time.Now().Add(-24 * time.Hour * 10).Unix(),
		Kind:       selector.Image,
	}

	isMatch, err = match.Match(orm.Context(), 1, c6)
	s.require.Equal(isMatch, false) // no longer immutable
	s.require.Nil(err)

	// postgres pushed 2d ago
	c7 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "postgres",
		// no tags
		Tags: []string{
			"latest",
		},
		PushedTime: time.Now().Add(-24 * time.Hour * 2).Unix(),
		Kind:       selector.Image,
	}

	isMatch, err = match.Match(orm.Context(), 1, c7)
	s.require.Equal(isMatch, true) // it is still immutable
	s.require.Nil(err)

	// nginx pulled 49h ago
	c8 := selector.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "nginx",
		// no tags
		Tags: []string{
			"latest",
		},
		PulledTime: time.Now().Add(-49 * time.Hour).Unix(),
		PushedTime: time.Now().Unix(), // <-we don't cate
		Kind:       selector.Image,
	}

	isMatch, err = match.Match(orm.Context(), 1, c8)
	s.require.Equal(isMatch, false) // no longer immutable (rule protects only within 2d)
	s.require.Nil(err)
}

// TearDownSuite clears env for test suite
func (s *MatchTestSuite) TearDownSuite() {
	for _, id := range s.ruleIDs {
		err := s.ctr.DeleteImmutableRule(orm.Context(), id)
		require.NoError(s.T(), err, "delete immutable")
	}
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
