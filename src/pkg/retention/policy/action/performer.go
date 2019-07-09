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
type PerformerFactory func() Performer

// retainAction make sure all the candidates will be retained and others will be cleared
type retainAction struct{}

// Perform the action
func (ra *retainAction) Perform(candidates []*res.Candidate) ([]*res.Result, error) {
	return nil, nil
}

// NewRetainAction is factory method for RetainAction
func NewRetainAction() Performer {
	return &retainAction{}
}
