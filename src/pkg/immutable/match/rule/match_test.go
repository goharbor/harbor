package rule

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
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

// TestImmuMatchMultiSelector guards against a partial selector match: a rule that carries
// more than one repository (or tag) selector must protect every entry, not just
// the first one. Before the fix, Match only evaluated selectors[0], so a push to
// the repository named by the second selector escaped the rule.
func (s *MatchTestSuite) TestImmuMatchMultiSelector() {
	s.purgeRules(2)
	rule := &model.Metadata{
		ProjectID: 2,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{Kind: "doublestar", Decoration: "matches", Pattern: "release-**"},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "redis"},
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "mysql"},
			},
		},
	}
	id, err := s.ctr.CreateImmutableRule(orm.Context(), rule)
	s.require.Nil(err)
	defer func() {
		s.require.NoError(s.ctr.DeleteImmutableRule(orm.Context(), id))
	}()

	match := NewRuleMatcher()

	// first repository selector -> immutable (always worked)
	first := selector.Candidate{
		NamespaceID: 2, Namespace: "library", Repository: "redis",
		Tags: []string{"release-1.0"}, Kind: selector.Image,
	}
	isMatch, err := match.Match(orm.Context(), 2, first)
	s.require.Nil(err)
	s.require.True(isMatch, "tag covered by the first repository selector must be immutable")

	// second repository selector -> must also be immutable (the bug)
	second := selector.Candidate{
		NamespaceID: 2, Namespace: "library", Repository: "mysql",
		Tags: []string{"release-1.0"}, Kind: selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 2, second)
	s.require.Nil(err)
	s.require.True(isMatch, "tag covered by the second repository selector must be immutable")

	// a repository named by no selector stays mutable
	other := selector.Candidate{
		NamespaceID: 2, Namespace: "library", Repository: "postgres",
		Tags: []string{"release-1.0"}, Kind: selector.Image,
	}
	isMatch, err = match.Match(orm.Context(), 2, other)
	s.require.Nil(err)
	s.require.False(isMatch, "tag covered by no selector must stay mutable")
}

// TestImmuMatchMultiTagSelector guards the tag dimension of the same defect.
func (s *MatchTestSuite) TestImmuMatchMultiTagSelector() {
	s.purgeRules(3)
	rule := &model.Metadata{
		ProjectID: 3,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{Kind: "doublestar", Decoration: "matches", Pattern: "release-**"},
			{Kind: "doublestar", Decoration: "matches", Pattern: "stable-**"},
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "redis"},
			},
		},
	}
	id, err := s.ctr.CreateImmutableRule(orm.Context(), rule)
	s.require.Nil(err)
	defer func() {
		s.require.NoError(s.ctr.DeleteImmutableRule(orm.Context(), id))
	}()

	match := NewRuleMatcher()

	// tag covered by the second tag selector must be immutable (the bug)
	c := selector.Candidate{
		NamespaceID: 3, Namespace: "library", Repository: "redis",
		Tags: []string{"stable-1.0"}, Kind: selector.Image,
	}
	isMatch, err := match.Match(orm.Context(), 3, c)
	s.require.Nil(err)
	s.require.True(isMatch, "tag covered by the second tag selector must be immutable")
}

// TestImmuMatchMultiExcludeSelector checks that several exclusion selectors in
// one dimension behave like the portal's single `{a,b}` exclusion pattern: a
// candidate is excluded when it matches any of them.
func (s *MatchTestSuite) TestImmuMatchMultiExcludeSelector() {
	s.purgeRules(4)
	cases := []struct {
		name     string
		repoSels []*model.Selector
		tagSels  []*model.Selector
		repo     string
		tags     []string
		want     bool
	}{
		{
			name: "repoExcludes: first excluded repository stays mutable",
			repoSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "redis"},
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "mysql"},
			},
			tagSels: []*model.Selector{{Kind: "doublestar", Decoration: "matches", Pattern: "**"}},
			repo:    "redis", tags: []string{"1.0"}, want: false,
		},
		{
			name: "repoExcludes: second excluded repository stays mutable",
			repoSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "redis"},
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "mysql"},
			},
			tagSels: []*model.Selector{{Kind: "doublestar", Decoration: "matches", Pattern: "**"}},
			repo:    "mysql", tags: []string{"1.0"}, want: false,
		},
		{
			name: "repoExcludes: repository named by no exclusion is immutable",
			repoSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "redis"},
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "mysql"},
			},
			tagSels: []*model.Selector{{Kind: "doublestar", Decoration: "matches", Pattern: "**"}},
			repo:    "postgres", tags: []string{"1.0"}, want: true,
		},
		{
			name:     "excludes: tag matching the second exclusion stays mutable",
			repoSels: []*model.Selector{{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"}},
			tagSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "excludes", Pattern: "dev-**"},
				{Kind: "doublestar", Decoration: "excludes", Pattern: "test-**"},
			},
			repo: "redis", tags: []string{"test-1"}, want: false,
		},
		{
			name:     "excludes: tag matching no exclusion is immutable",
			repoSels: []*model.Selector{{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"}},
			tagSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "excludes", Pattern: "dev-**"},
				{Kind: "doublestar", Decoration: "excludes", Pattern: "test-**"},
			},
			repo: "redis", tags: []string{"release-1"}, want: true,
		},
		{
			name:     "excludes: artifact whose tags are each excluded by a different selector stays mutable",
			repoSels: []*model.Selector{{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"}},
			tagSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "excludes", Pattern: "dev-**"},
				{Kind: "doublestar", Decoration: "excludes", Pattern: "test-**"},
			},
			repo: "redis", tags: []string{"dev-1", "test-1"}, want: false,
		},
		{
			name:     "excludes: artifact with one non-excluded tag is immutable",
			repoSels: []*model.Selector{{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"}},
			tagSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "excludes", Pattern: "dev-**"},
				{Kind: "doublestar", Decoration: "excludes", Pattern: "test-**"},
			},
			repo: "redis", tags: []string{"dev-1", "release-1"}, want: true,
		},
		{
			name: "mixed: exclusion carves out of a matches selector",
			repoSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"},
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "sandbox/**"},
			},
			tagSels: []*model.Selector{{Kind: "doublestar", Decoration: "matches", Pattern: "**"}},
			repo:    "sandbox/app", tags: []string{"1.0"}, want: false,
		},
		{
			name: "mixed: repository outside the exclusion is immutable",
			repoSels: []*model.Selector{
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"},
				{Kind: "doublestar", Decoration: "repoExcludes", Pattern: "sandbox/**"},
			},
			tagSels: []*model.Selector{{Kind: "doublestar", Decoration: "matches", Pattern: "**"}},
			repo:    "prod/app", tags: []string{"1.0"}, want: true,
		},
	}

	const pid = int64(4)
	for _, tc := range cases {
		s.Run(tc.name, func() {
			id, err := s.ctr.CreateImmutableRule(orm.Context(), &model.Metadata{
				ProjectID:      pid,
				Priority:       1,
				Action:         "immutable",
				Template:       "immutable_template",
				TagSelectors:   tc.tagSels,
				ScopeSelectors: map[string][]*model.Selector{"repository": tc.repoSels},
			})
			s.require.NoError(err)
			defer func() {
				s.require.NoError(s.ctr.DeleteImmutableRule(orm.Context(), id))
			}()

			isMatch, err := NewRuleMatcher().Match(orm.Context(), pid, selector.Candidate{
				NamespaceID: pid, Namespace: "library", Repository: tc.repo,
				Tags: tc.tags, Kind: selector.Image,
			})
			s.require.NoError(err)
			s.Equal(tc.want, isMatch)
		})
	}
}

// TestImmuMatchNilSelector checks that null selector entries stored through the
// API are ignored instead of panicking.
func (s *MatchTestSuite) TestImmuMatchNilSelector() {
	const pid = int64(5)
	s.purgeRules(pid)
	id, err := s.ctr.CreateImmutableRule(orm.Context(), &model.Metadata{
		ProjectID: pid,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*model.Selector{
			{Kind: "doublestar", Decoration: "matches", Pattern: "release-**"},
			nil,
		},
		ScopeSelectors: map[string][]*model.Selector{
			"repository": {
				nil,
				{Kind: "doublestar", Decoration: "repoMatches", Pattern: "redis"},
			},
		},
	})
	s.require.NoError(err)
	defer func() {
		s.require.NoError(s.ctr.DeleteImmutableRule(orm.Context(), id))
	}()

	match := NewRuleMatcher()
	isMatch, err := match.Match(orm.Context(), pid, selector.Candidate{
		NamespaceID: pid, Namespace: "library", Repository: "redis",
		Tags: []string{"release-1.0"}, Kind: selector.Image,
	})
	s.require.NoError(err)
	s.True(isMatch)

	isMatch, err = match.Match(orm.Context(), pid, selector.Candidate{
		NamespaceID: pid, Namespace: "library", Repository: "mysql",
		Tags: []string{"release-1.0"}, Kind: selector.Image,
	})
	s.require.NoError(err)
	s.False(isMatch)
}

// purgeRules removes rules left in the project by an interrupted run, since
// Match evaluates every rule of the project.
func (s *MatchTestSuite) purgeRules(pid int64) {
	rules, err := s.ctr.ListImmutableRules(orm.Context(), q.New(q.KeyWords{"ProjectID": pid}))
	s.require.NoError(err)
	for _, r := range rules {
		s.require.NoError(s.ctr.DeleteImmutableRule(orm.Context(), r.ID))
	}
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
