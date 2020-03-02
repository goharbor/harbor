package rule

import (
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/art/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
)

// Matcher ...
type Matcher struct {
	rules []model.Metadata
}

// Match ...
func (rm *Matcher) Match(pid int64, c art.Candidate) (bool, error) {
	if err := rm.getImmutableRules(pid); err != nil {
		return false, err
	}

	cands := []*art.Candidate{&c}
	for _, r := range rm.rules {
		if r.Disabled {
			continue
		}

		// match repositories according to the repository selectors
		var repositoryCandidates []*art.Candidate
		repositorySelectors := r.ScopeSelectors["repository"]
		if len(repositorySelectors) < 1 {
			continue
		}
		repositorySelector := repositorySelectors[0]
		selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
			repositorySelector.Pattern)
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
		var tagCandidates []*art.Candidate
		tagSelectors := r.TagSelectors
		if len(tagSelectors) < 0 {
			continue
		}
		tagSelector := r.TagSelectors[0]
		selector, err = index.Get(tagSelector.Kind, tagSelector.Decoration,
			tagSelector.Pattern)
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

func (rm *Matcher) getImmutableRules(pid int64) error {
	rules, err := immutabletag.ImmuCtr.ListImmutableRules(pid)
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
