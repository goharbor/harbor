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

package latestk

import (
	"strconv"
	"testing"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/suite"
)

type EvaluatorTestSuite struct {
	suite.Suite

	artifacts []*res.Candidate
}

func (e *EvaluatorTestSuite) SetupSuite() {
	e.artifacts = []*res.Candidate{
		{PulledTime: 1, PushedTime: 2},
		{PulledTime: 3, PushedTime: 4},
		{PulledTime: 6, PushedTime: 5},
		{PulledTime: 8, PushedTime: 7},
		{PulledTime: 9, PushedTime: 9},
		{PulledTime: 10, PushedTime: 10},
		{PulledTime: 0, PushedTime: 11},
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	tests := []struct {
		k             int
		expected      int
		minActiveTime int64
	}{
		{k: 0, expected: 0},
		{k: 1, expected: 1, minActiveTime: 11},
		{k: 2, expected: 2, minActiveTime: 10},
		{k: 5, expected: 5, minActiveTime: 6},
		{k: 6, expected: 6, minActiveTime: 3},
		{k: 99, expected: len(e.artifacts)},
	}
	for _, tt := range tests {
		e.T().Run(strconv.Itoa(tt.k), func(t *testing.T) {
			sut := &evaluator{k: tt.k}

			result, err := sut.Process(e.artifacts)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)

			for _, v := range result {
				assert.True(t, activeTime(v) >= tt.minActiveTime)
			}
		})
	}
}

func (e *EvaluatorTestSuite) TestNew() {
	tests := []struct {
		name      string
		params    rule.Parameters
		expectedK int
	}{
		{name: "Valid", params: rule.Parameters{ParameterK: 5}, expectedK: 5},
		{name: "Default If Negative", params: rule.Parameters{ParameterK: -5}, expectedK: DefaultK},
		{name: "Default If Wrong Type", params: rule.Parameters{ParameterK: "5"}, expectedK: DefaultK},
		{name: "Default If Wrong Key", params: rule.Parameters{"n": 5}, expectedK: DefaultK},
		{name: "Default If Empty", params: rule.Parameters{}, expectedK: DefaultK},
	}
	for _, tt := range tests {
		e.T().Run(tt.name, func(t *testing.T) {
			sut := New(tt.params).(*evaluator)

			require.Equal(t, tt.expectedK, sut.k)
		})
	}
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}
