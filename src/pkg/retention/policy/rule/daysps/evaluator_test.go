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
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/internal/selector"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
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
	data := []*selector.Candidate{
		{PushedTime: daysAgo(now, 1, time.Hour)},
		{PushedTime: daysAgo(now, 2, time.Hour)},
		{PushedTime: daysAgo(now, 3, time.Hour)},
		{PushedTime: daysAgo(now, 4, time.Hour)},
		{PushedTime: daysAgo(now, 5, time.Hour)},
		{PushedTime: daysAgo(now, 10, time.Hour)},
		{PushedTime: daysAgo(now, 20, time.Hour)},
		{PushedTime: daysAgo(now, 30, time.Hour)},
	}

	tests := []struct {
		n           float64
		expected    int
		minPushTime int64
	}{
		{n: 0, expected: 0, minPushTime: 0},
		{n: 1, expected: 1, minPushTime: daysAgo(now, 1, 0)},
		{n: 2, expected: 2, minPushTime: daysAgo(now, 2, 0)},
		{n: 3, expected: 3, minPushTime: daysAgo(now, 3, 0)},
		{n: 4, expected: 4, minPushTime: daysAgo(now, 4, 0)},
		{n: 5, expected: 5, minPushTime: daysAgo(now, 5, 0)},
		{n: 15, expected: 6, minPushTime: daysAgo(now, 10, 0)},
		{n: 90, expected: 8, minPushTime: daysAgo(now, 30, 0)},
	}

	for _, tt := range tests {
		e.T().Run(fmt.Sprintf("%v", tt.n), func(t *testing.T) {
			sut := New(map[string]rule.Parameter{ParameterN: tt.n})

			result, err := sut.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)

			for _, v := range result {
				assert.False(t, v.PushedTime < tt.minPushTime)
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
		{Name: "Negative", args: map[string]rule.Parameter{ParameterN: -1}, expectedK: errors.New("nDaysSinceLastPush is less than zero")},
		{Name: "Big", args: map[string]rule.Parameter{ParameterN: 21000000}, expectedK: errors.New("nDaysSinceLastPush is too large")},
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

func daysAgo(from time.Time, n int, offset time.Duration) int64 {
	return from.Add(time.Duration(-1*24*n)*time.Hour + offset).Unix()
}
