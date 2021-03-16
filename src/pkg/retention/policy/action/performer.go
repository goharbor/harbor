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

package action

import (
	"context"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/immutable/match/rule"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
)

const (
	// Retain artifacts
	Retain = "retain"
)

// Performer performs the related actions targeting the candidates
type Performer interface {
	// Perform the action
	//
	//  Arguments:
	//    candidates []*art.Candidate : the targets to perform
	//
	//  Returns:
	//    []*art.Result : result infos
	//    error     : common error if any errors occurred
	Perform(ctx context.Context, candidates []*selector.Candidate) ([]*selector.Result, error)
}

// PerformerFactory is factory method for creating Performer
type PerformerFactory func(params interface{}, isDryRun bool) Performer

// retainAction make sure all the candidates will be retained and others will be cleared
type retainAction struct {
	all []*selector.Candidate
	// Indicate if it is a dry run
	isDryRun bool
}

// Perform the action
func (ra *retainAction) Perform(ctx context.Context, candidates []*selector.Candidate) (results []*selector.Result, err error) {
	retainedShare := make(map[string]bool)
	immutableShare := make(map[string]bool)
	for _, c := range candidates {
		retainedShare[c.Hash()] = true
	}

	for _, c := range ra.all {
		if _, ok := retainedShare[c.Hash()]; ok {
			continue
		}
		if isImmutable(ctx, c) {
			immutableShare[c.Hash()] = true
		}
	}

	// start to delete
	if len(ra.all) > 0 {
		for _, c := range ra.all {
			if _, ok := retainedShare[c.Hash()]; !ok {
				result := &selector.Result{
					Target: c,
				}
				if _, ok = immutableShare[c.Hash()]; ok {
					result.Error = &selector.ImmutableError{}
				} else {
					if !ra.isDryRun {
						if err := dep.DefaultClient.Delete(c); err != nil {
							result.Error = err
						}
					}
				}
				results = append(results, result)
			}
		}
	}

	return
}

func isImmutable(ctx context.Context, c *selector.Candidate) bool {
	projectID := c.NamespaceID
	repo := c.Repository
	_, repoName := utils.ParseRepository(repo)
	matched, err := rule.NewRuleMatcher().Match(ctx, projectID, selector.Candidate{
		Repository:  repoName,
		Tags:        c.Tags,
		NamespaceID: projectID,
	})
	if err != nil {
		log.Error(err)
		return false
	}
	return matched
}

// NewRetainAction is factory method for RetainAction
func NewRetainAction(params interface{}, isDryRun bool) Performer {
	if params != nil {
		if all, ok := params.([]*selector.Candidate); ok {
			return &retainAction{
				all:      all,
				isDryRun: isDryRun,
			}
		}
	}

	return &retainAction{
		all:      make([]*selector.Candidate, 0),
		isDryRun: isDryRun,
	}
}
