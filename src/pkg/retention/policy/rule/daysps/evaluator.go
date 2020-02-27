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

package daysps

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/pkg/artifactselector"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

const (
	// TemplateID of the rule
	TemplateID = "nDaysSinceLastPush"

	// ParameterN is the name of the metadata parameter for the N value
	ParameterN = TemplateID

	// DefaultN is the default number of days that an artifact must have
	// been pulled within to retain the tag or artifact.
	DefaultN = 30
)

type evaluator struct {
	n int
}

func (e *evaluator) Process(artifacts []*artifactselector.Candidate) (result []*artifactselector.Candidate, err error) {
	minPushTime := time.Now().UTC().Add(time.Duration(-1*24*e.n) * time.Hour).Unix()
	for _, a := range artifacts {
		if a.PushedTime >= minPushTime {
			result = append(result, a)
		}
	}

	return
}

func (e *evaluator) Action() string {
	return action.Retain
}

// New constructs a new 'Days Since Last Push' evaluator
func New(params rule.Parameters) rule.Evaluator {
	if params != nil {
		if p, ok := params[ParameterN]; ok {
			if v, ok := utils.ParseJSONInt(p); ok && v >= 0 {
				return &evaluator{n: int(v)}
			}
		}
	}

	log.Warningf("default parameter %d used for rule %s", DefaultN, TemplateID)

	return &evaluator{n: DefaultN}
}

// Valid ...
func Valid(params rule.Parameters) error {
	if params != nil {
		if p, ok := params[ParameterN]; ok {
			if v, ok := utils.ParseJSONInt(p); ok {
				if v < 0 {
					return fmt.Errorf("%s is less than zero", ParameterN)
				}
				if v > 20190904 {
					return fmt.Errorf("%s is too large", ParameterN)
				}
			} else {
				return fmt.Errorf("%s type error", ParameterN)
			}
		}
	}
	return nil
}
