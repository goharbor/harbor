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
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/action"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
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
func (e *evaluator) Process(artifacts []*selector.Candidate) (retain []*selector.Candidate, err error) {
	cutoff := time.Now().Add(time.Duration(e.x*-24) * time.Hour)
	for _, a := range artifacts {
		if time.Unix(a.PushedTime, 0).UTC().After(cutoff) {
			retain = append(retain, a)
		}
	}

	return
}

// Specify what action is performed to the candidates processed by this evaluator
func (e *evaluator) Action() string {
	return action.Retain
}

// New a Evaluator
func New(params rule.Parameters) rule.Evaluator {
	if params != nil {
		if p, ok := params[ParameterX]; ok {
			if v, ok := utils.ParseJSONInt(p); ok && v >= 0 {
				return &evaluator{
					x: int(v),
				}
			}
		}
	}

	log.Warningf("default parameter %d used for rule %s", DefaultX, TemplateID)

	return &evaluator{
		x: DefaultX,
	}
}
