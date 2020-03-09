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

package latestpl

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/internal/selector"
	"math"
	"sort"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

const (
	// TemplateID of the rule
	TemplateID = "latestPulledN"

	// ParameterN is the name of the metadata parameter for the N value
	ParameterN = TemplateID

	// DefaultN is the default number of tags to retain
	DefaultN = 10
)

type evaluator struct {
	n int
}

func (e *evaluator) Process(artifacts []*selector.Candidate) ([]*selector.Candidate, error) {
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].PulledTime > artifacts[j].PulledTime
	})

	i := e.n
	if i > len(artifacts) {
		i = len(artifacts)
	}

	return artifacts[:i], nil
}

func (e *evaluator) Action() string {
	return action.Retain
}

// New constructs an evaluator with the given parameters
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
				if v >= math.MaxInt16 {
					return fmt.Errorf("%s is too large", ParameterN)
				}
			} else {
				return fmt.Errorf("%s type error", ParameterN)
			}
		}
	}
	return nil
}
