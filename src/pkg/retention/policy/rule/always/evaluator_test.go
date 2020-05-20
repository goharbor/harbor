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

package always

import (
	"github.com/goharbor/harbor/src/lib/selector"
	"testing"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EvaluatorTestSuite struct {
	suite.Suite
}

func (e *EvaluatorTestSuite) TestNew() {
	sut := New(rule.Parameters{})

	require.NotNil(e.T(), sut)
	require.IsType(e.T(), &evaluator{}, sut)
}

func (e *EvaluatorTestSuite) TestProcess() {
	sut := New(rule.Parameters{})
	input := []*selector.Candidate{{PushedTime: 0}, {PushedTime: 1}, {PushedTime: 2}, {PushedTime: 3}}

	result, err := sut.Process(input)

	require.NoError(e.T(), err)
	require.Len(e.T(), result, len(input))
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}
