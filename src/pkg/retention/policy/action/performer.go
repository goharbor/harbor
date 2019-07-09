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

import "github.com/goharbor/harbor/src/pkg/retention/res"

const (
	// Retain artifacts
	Retain = "retain"
)

// Performer performs the related actions targeting the candidates
type Performer interface {
	// Perform the action
	//
	//  Arguments:
	//    candidates []*res.Candidate : the targets to perform
	//
	//  Returns:
	//    []*res.Result : result infos
	//    error     : common error if any errors occurred
	Perform(candidates []*res.Candidate) ([]*res.Result, error)
}

// PerformerFactory is factory method for creating Performer
type PerformerFactory func(params interface{}) Performer

// retainAction make sure all the candidates will be retained and others will be cleared
type retainAction struct {
	all []*res.Candidate
}

// Perform the action
func (ra *retainAction) Perform(candidates []*res.Candidate) ([]*res.Result, error) {
	// TODO: REPLACE SAMPLE CODE WITH REAL IMPLEMENTATION
	results := make([]*res.Result, 0)

	for _, c := range candidates {
		results = append(results, &res.Result{
			Target: c,
		})
	}

	return results, nil
}

// NewRetainAction is factory method for RetainAction
func NewRetainAction(params interface{}) Performer {
	if params != nil {
		if all, ok := params.([]*res.Candidate); ok {
			return &retainAction{
				all: all,
			}
		}
	}

	return &retainAction{
		all: make([]*res.Candidate, 0),
	}
}

func init() {
	// Register itself
	Register(Retain, NewRetainAction)
}
