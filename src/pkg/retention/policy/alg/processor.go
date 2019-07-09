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
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
)

// Processor processing the whole policy targeting a repository.
// Methods are defined to reflect the standard structure of the policy:
// list of rules with corresponding selectors plus an action performer.
type Processor interface {
	// Process the artifact candidates
	//
	//  Arguments:
	//    artifacts []*res.Candidate : process the retention candidates
	//
	//  Returns:
	//    []*res.Result : the processed results
	//    error         : common error object if any errors occurred
	Process(artifacts []*res.Candidate) ([]*res.Result, error)

	// Add a rule evaluator for the processor
	//
	//  Arguments:
	//    evaluator rule.Evaluator : a rule evaluator
	//    selector res.Selector    : selector to narrow down the scope, optional
	AddEvaluator(evaluator rule.Evaluator, selector res.Selector)

	// Set performer for the processor
	//
	//  Arguments:
	//    performer action.Performer : a performer implementation
	SetPerformer(performer action.Performer)
}
