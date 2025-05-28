// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rule

import (
	"context"
	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/q"
	iselector "github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable/match"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	policyindex "github.com/goharbor/harbor/src/pkg/retention/policy/rule/index"
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
		params := rule.Parameters{}
		for k, v := range r.Parameters {
			params[k] = v
		}

		evaluator, err := policyindex.Get(r.Template, params)
		if err != nil {
			return false, err
		}

		ruleCandidates, err := evaluator.Process(cands)
		if err != nil {
			return false, err
		}
		if len(ruleCandidates) == 0 {
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
