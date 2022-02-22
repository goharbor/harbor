package rule

import (
	"context"

	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/q"
	iselector "github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable/match"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

// Matcher ...
type Matcher struct {
	rules []*model.Metadata
}

// Match ...
func (rm *Matcher) Match(ctx context.Context, pid int64, c iselector.Candidate) (bool, error) {
	if err := rm.getImmutableRules(ctx, pid); err != nil {
		return false, err
	}

	cands := []*iselector.Candidate{&c}
	for _, r := range rm.rules {
		if r.Disabled {
			continue
		}

		// match repositories according to the repository selectors
		var repositoryCandidates []*iselector.Candidate
		repositorySelectors := r.ScopeSelectors["repository"]
		if len(repositorySelectors) < 1 {
			continue
		}
		repositorySelector := repositorySelectors[0]
		selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
			repositorySelector.Pattern, "")
		if err != nil {
			return false, err
		}
		repositoryCandidates, err = selector.Select(cands)
		if err != nil {
			return false, err
		}
		if len(repositoryCandidates) == 0 {
			continue
		}

		// match tag according to the tag selectors
		var tagCandidates []*iselector.Candidate
		tagSelectors := r.TagSelectors
		if len(tagSelectors) < 1 {
			continue
		}
		tagSelector := r.TagSelectors[0]
		// for immutable policy, should not keep untagged artifacts by default.
		selector, err = index.Get(tagSelector.Kind, tagSelector.Decoration,
			tagSelector.Pattern, "{\"untagged\": false}")
		if err != nil {
			return false, err
		}
		tagCandidates, err = selector.Select(cands)
		if err != nil {
			return false, err
		}
		if len(tagCandidates) == 0 {
			continue
		}

		return true, nil
	}
	return false, nil
}

func (rm *Matcher) getImmutableRules(ctx context.Context, pid int64) error {
	rules, err := immutable.Ctr.ListImmutableRules(ctx, q.New(q.KeyWords{"ProjectID": pid}))
	if err != nil {
		return err
	}
	rm.rules = rules
	return nil
}

// NewRuleMatcher ...
func NewRuleMatcher() match.ImmutableTagMatcher {
	return &Matcher{}
}
