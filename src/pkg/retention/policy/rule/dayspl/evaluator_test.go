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

package dayspl

import (
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/assert"
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
		expectedN int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterN: float64(5)}, expectedN: 5},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterN: float64(-1)}, expectedN: DefaultN},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedN: DefaultN},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{ParameterN: "foo"}, expectedN: DefaultN},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedN, e.n)
		})
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	now := time.Now().UTC()
	data := []*res.Candidate{
		{PulledTime: daysAgo(now, 1)},
		{PulledTime: daysAgo(now, 2)},
		{PulledTime: daysAgo(now, 3)},
		{PulledTime: daysAgo(now, 4)},
		{PulledTime: daysAgo(now, 5)},
		{PulledTime: daysAgo(now, 10)},
		{PulledTime: daysAgo(now, 20)},
		{PulledTime: daysAgo(now, 30)},
	}

	tests := []struct {
		n           float64
		expected    int
		minPullTime int64
	}{
		{n: 0, expected: 0, minPullTime: 0},
		{n: 1, expected: 1, minPullTime: daysAgo(now, 1)},
		{n: 2, expected: 2, minPullTime: daysAgo(now, 2)},
		{n: 3, expected: 3, minPullTime: daysAgo(now, 3)},
		{n: 4, expected: 4, minPullTime: daysAgo(now, 4)},
		{n: 5, expected: 5, minPullTime: daysAgo(now, 5)},
		{n: 15, expected: 6, minPullTime: daysAgo(now, 10)},
		{n: 90, expected: 8, minPullTime: daysAgo(now, 30)},
	}

	for _, tt := range tests {
		e.T().Run(fmt.Sprintf("%v", tt.n), func(t *testing.T) {
			sut := New(map[string]rule.Parameter{ParameterN: tt.n})

			result, err := sut.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)

			for _, v := range result {
				assert.False(t, v.PulledTime < tt.minPullTime)
			}
		})
	}
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}

func daysAgo(from time.Time, n int) int64 {
	return from.Add(time.Duration(-1*24*n) * time.Hour).Unix()
}
