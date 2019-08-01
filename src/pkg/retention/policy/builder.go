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

package policy

import (
	"fmt"

	index4 "github.com/goharbor/harbor/src/pkg/retention/policy/action/index"

	index3 "github.com/goharbor/harbor/src/pkg/retention/policy/alg/index"

	index2 "github.com/goharbor/harbor/src/pkg/retention/res/selectors/index"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule/index"

	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/pkg/errors"
)

// Builder builds the runnable processor from the raw policy
type Builder interface {
	// Builds runnable processor
	//
	//  Arguments:
	//    policy *Metadata : the simple metadata of retention policy
	//    isDryRun bool    : indicate if we need to build a processor for dry run
	//
	//  Returns:
	//    Processor : a processor implementation to process the candidates
	//    error     : common error object if any errors occurred
	Build(policy *lwp.Metadata, isDryRun bool) (alg.Processor, error)
}

// NewBuilder news a basic builder
func NewBuilder(all []*res.Candidate) Builder {
	return &basicBuilder{
		allCandidates: all,
	}
}

// basicBuilder is default implementation of Builder interface
type basicBuilder struct {
	allCandidates []*res.Candidate
}

// Build policy processor from the raw policy
func (bb *basicBuilder) Build(policy *lwp.Metadata, isDryRun bool) (alg.Processor, error) {
	if policy == nil {
		return nil, errors.New("nil policy to build processor")
	}

	params := make([]*alg.Parameter, 0)

	for _, r := range policy.Rules {
		evaluator, err := index.Get(r.Template, r.Parameters)
		if err != nil {
			return nil, err
		}

		perf, err := index4.Get(r.Action, bb.allCandidates, isDryRun)
		if err != nil {
			return nil, errors.Wrap(err, "get action performer by metadata")
		}

		sl := make([]res.Selector, 0)
		for _, s := range r.TagSelectors {
			sel, err := index2.Get(s.Kind, s.Decoration, s.Pattern)
			if err != nil {
				return nil, errors.Wrap(err, "get selector by metadata")
			}

			sl = append(sl, sel)
		}

		params = append(params, &alg.Parameter{
			Evaluator: evaluator,
			Selectors: sl,
			Performer: perf,
		})
	}

	p, err := index3.Get(policy.Algorithm, params)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("get processor for algorithm: %s", policy.Algorithm))
	}

	return p, nil
}
