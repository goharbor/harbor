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

package alg

import (
	"context"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

// Processor processing the whole policy targeting a repository.
// Methods are defined to reflect the standard structure of the policy:
// list of rules with corresponding selectors plus an action performer.
type Processor interface {
	// Process the artifact candidates
	//
	//  Arguments:
	//    artifacts []*art.Candidate : process the retention candidates
	//
	//  Returns:
	//    []*art.Result : the processed results
	//    error         : common error object if any errors occurred
	Process(ctx context.Context, artifacts []*selector.Candidate) ([]*selector.Result, error)
}

// Parameter for constructing a processor
// Represents one rule
type Parameter struct {
	// Evaluator for the rule
	Evaluator rule.Evaluator

	// Selectors for the rule
	Selectors []selector.Selector

	// Performer for the rule evaluator
	Performer action.Performer
}

// Factory for creating processor
type Factory func([]*Parameter) Processor
