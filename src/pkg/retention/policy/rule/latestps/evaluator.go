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

package latestps

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/selector"
	"math"
	"sort"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/retention/policy/action"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

const (
	// TemplateID of latest k rule
	TemplateID = "latestPushedK"
	// ParameterK ...
	ParameterK = TemplateID
	// DefaultK defines the default K
	DefaultK = 10
)

// evaluator for evaluating latest k tags
type evaluator struct {
	// latest k
	k int
}

// Process the candidates based on the rule definition
func (e *evaluator) Process(artifacts []*selector.Candidate) ([]*selector.Candidate, error) {
	// The updated proposal does not guarantee the order artifacts are provided, so we have to sort them first
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].PushedTime > artifacts[j].PushedTime
	})

	i := e.k
	if i > len(artifacts) {
		i = len(artifacts)
	}

	return artifacts[:i], nil
}

// Specify what action is performed to the candidates processed by this evaluator
func (e *evaluator) Action() string {
	return action.Retain
}

// New a Evaluator
func New(params rule.Parameters) rule.Evaluator {
	if params != nil {
		if p, ok := params[ParameterK]; ok {
			if v, ok := utils.ParseJSONInt(p); ok && v >= 0 {
				return &evaluator{
					k: int(v),
				}
			}
		}
	}

	log.Warningf("default parameter %d used for rule %s", DefaultK, TemplateID)

	return &evaluator{
		k: DefaultK,
	}
}

// Valid ...
func Valid(params rule.Parameters) error {
	if params != nil {
		if p, ok := params[ParameterK]; ok {
			if v, ok := utils.ParseJSONInt(p); ok {
				if v < 0 {
					return fmt.Errorf("%s is less than zero", ParameterK)
				}
				if v >= math.MaxInt16 {
					return fmt.Errorf("%s is too large", ParameterK)
				}
			} else {
				return fmt.Errorf("%s type error", ParameterK)
			}
		}
	}
	return nil
}
