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
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EvaluatorTestSuite struct {
	suite.Suite
}

func (e *EvaluatorTestSuite) TestNew() {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedK int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterN: float64(5)}, expectedK: 5},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterN: float64(-1)}, expectedK: DefaultN},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedK: DefaultN},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{ParameterN: "foo"}, expectedK: DefaultN},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedK, e.n)
		})
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	data := []*res.Candidate{{PulledTime: 0}, {PulledTime: 1}, {PulledTime: 2}, {PulledTime: 3}, {PulledTime: 4}}
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	tests := []struct {
		n           float64
		expected    int
		minPullTime int64
	}{
		{n: 0, expected: 0, minPullTime: 0},
		{n: 1, expected: 1, minPullTime: 4},
		{n: 3, expected: 3, minPullTime: 2},
		{n: 5, expected: 5, minPullTime: 0},
		{n: 6, expected: 5, minPullTime: 0},
	}

	for _, tt := range tests {
		e.T().Run(fmt.Sprintf("%v", tt.n), func(t *testing.T) {
			ev := New(map[string]rule.Parameter{ParameterN: tt.n})

			result, err := ev.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)

			for _, v := range result {
				require.False(e.T(), v.PulledTime < tt.minPullTime)
			}
		})
	}
}

func (e *EvaluatorTestSuite) TestValid() {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedK error
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterN: 5}, expectedK: nil},
		{Name: "Negative", args: map[string]rule.Parameter{ParameterN: -1}, expectedK: errors.New("latestPulledN is less than zero")},
		{Name: "Big", args: map[string]rule.Parameter{ParameterN: 40000}, expectedK: errors.New("latestPulledN is too large")},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			err := Valid(tt.args)

			require.Equal(t, tt.expectedK, err)
		})
	}
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}
