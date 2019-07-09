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

package lastx

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
)

const (
	// TemplateID of last x days rule
	TemplateID = "lastXDays"
	// ParameterX ...
	ParameterX = TemplateID
	// DefaultX defines the default X
	DefaultX = 10
)

// evaluator for evaluating last x days
type evaluator struct {
	// last x days
	x int
}

// Process the candidates based on the rule definition
func (e *evaluator) Process(artifacts []*res.Candidate) ([]*res.Candidate, error) {
	return nil, nil
}

// New a Evaluator
func New(params rule.Parameters) rule.Evaluator {
	if params != nil {
		if param, ok := params[ParameterX]; ok {
			if v, ok := param.(int); ok {
				return &evaluator{
					x: v,
				}
			}
		}
	}

	log.Debugf("default parameter %d used for rule %s", DefaultX, TemplateID)

	return &evaluator{
		x: DefaultX,
	}
}

func init() {
	// Register itself
	rule.Register(&rule.IndexMeta{
		TemplateID: TemplateID,
		Action:     action.Retain,
		Parameters: []*rule.IndexedParam{
			{
				Name:     ParameterX,
				Type:     "int",
				Unit:     "days",
				Required: true,
			},
		},
	}, New)
}
