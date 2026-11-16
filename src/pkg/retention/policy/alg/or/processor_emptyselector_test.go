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

package or

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

// retainAllEvaluator keeps every candidate it receives.
type retainAllEvaluator struct{}

func (retainAllEvaluator) Process(artifacts []*selector.Candidate) ([]*selector.Candidate, error) {
	return artifacts, nil
}

func (retainAllEvaluator) Action() string {
	return action.Retain
}

// collectingPerformer records whatever candidates are handed to it and returns
// a successful result for each of them.
type collectingPerformer struct {
	performed []*selector.Candidate
}

func (p *collectingPerformer) Perform(_ context.Context, candidates []*selector.Candidate) ([]*selector.Result, error) {
	p.performed = append(p.performed, candidates...)
	results := make([]*selector.Result, 0, len(candidates))
	for _, c := range candidates {
		results = append(results, &selector.Result{Target: c})
	}
	return results, nil
}

// TestProcessWithoutSelectors verifies that a rule carrying no tag selectors
// (which means "match all") is still evaluated. The server only validates the
// cron string when a retention policy is saved, so a rule with an empty selector
// list is reachable through the API. Before the fix the OR processor dropped
// such evaluators, so their retain action never ran and the artifacts the rule
// was meant to keep were left to be deleted.
func TestProcessWithoutSelectors(t *testing.T) {
	artifacts := []*selector.Candidate{
		{Namespace: "library", Repository: "harbor", Kind: "image", Tags: []string{"latest"}, Digest: "sha256:a"},
		{Namespace: "library", Repository: "harbor", Kind: "image", Tags: []string{"dev"}, Digest: "sha256:b"},
	}

	perf := &collectingPerformer{}
	p := New([]*alg.Parameter{
		{
			Evaluator: retainAllEvaluator{},
			Selectors: nil, // no selectors -> match all
			Performer: perf,
		},
	})

	results, err := p.Process(context.Background(), artifacts)
	require.NoError(t, err)
	require.Equal(t, len(artifacts), len(results), "the match-all rule must retain every artifact")
	for _, r := range results {
		assert.NoError(t, r.Error)
	}
	assert.Equal(t, len(artifacts), len(perf.performed), "the retain action must be performed on every artifact")
}

var _ rule.Evaluator = retainAllEvaluator{}
var _ action.Performer = (*collectingPerformer)(nil)
