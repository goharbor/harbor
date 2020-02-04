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
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
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
	Perform(candidates []*art.Candidate) ([]*art.Result, error)
}

// PerformerFactory is factory method for creating Performer
type PerformerFactory func(params interface{}, isDryRun bool) Performer

// retainAction make sure all the candidates will be retained and others will be cleared
type retainAction struct {
	all []*art.Candidate
	// Indicate if it is a dry run
	isDryRun bool
}

// Perform the action
func (ra *retainAction) Perform(candidates []*art.Candidate) (results []*art.Result, err error) {
	retained := make(map[string]bool)
	immutable := make(map[string]bool)
	retainedShare := make(map[string]bool)
	immutableShare := make(map[string]bool)
	for _, c := range candidates {
		retained[c.NameHash()] = true
		retainedShare[c.Hash()] = true
	}

	for _, c := range ra.all {
		if _, ok := retainedShare[c.Hash()]; ok {
			continue
		}
		if isImmutable(c) {
			immutable[c.NameHash()] = true
			immutableShare[c.Hash()] = true
		}
	}

	// start to delete
	if len(ra.all) > 0 {
		for _, c := range ra.all {
			if _, ok := retained[c.NameHash()]; !ok {
				if _, ok = retainedShare[c.Hash()]; !ok {
					result := &art.Result{
						Target: c,
					}
					if _, ok = immutable[c.NameHash()]; ok {
						result.Error = &art.ImmutableError{}
					} else if _, ok = immutableShare[c.Hash()]; ok {
						result.Error = &art.ImmutableError{IsShareDigest: true}
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
	}

	return
}

func isImmutable(c *art.Candidate) bool {
	projectID := c.NamespaceID
	repo := c.Repository
	tag := c.Tag
	_, repoName := utils.ParseRepository(repo)
	matched, err := rule.NewRuleMatcher().Match(projectID, art.Candidate{
		Repository:  repoName,
		Tag:         tag,
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
		if all, ok := params.([]*art.Candidate); ok {
			return &retainAction{
				all:      all,
				isDryRun: isDryRun,
			}
		}
	}

	return &retainAction{
		all:      make([]*art.Candidate, 0),
		isDryRun: isDryRun,
	}
}
