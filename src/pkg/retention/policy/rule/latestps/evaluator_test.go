package latestps

import (
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/require"
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
		{Name: "Valid", args: map[string]rule.Parameter{ParameterK: float64(5)}, expectedK: 5},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterK: float64(-1)}, expectedK: DefaultK},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedK: DefaultK},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{ParameterK: "foo"}, expectedK: DefaultK},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedK, e.k)
		})
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	data := []*res.Candidate{{PushedTime: 0}, {PushedTime: 1}, {PushedTime: 2}, {PushedTime: 3}, {PushedTime: 4}}
	rand.Shuffle(len(data), func(i, j int) {
		data[i], data[j] = data[j], data[i]
	})

	tests := []struct {
		k        float64
		expected int
	}{
		{k: 0, expected: 0},
		{k: 1, expected: 1},
		{k: 3, expected: 3},
		{k: 5, expected: 5},
		{k: 6, expected: 5},
	}

	for _, tt := range tests {
		e.T().Run(fmt.Sprintf("%v", tt.k), func(t *testing.T) {
			e := New(map[string]rule.Parameter{ParameterK: tt.k})

			result, err := e.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)
		})
	}
}

func (e *EvaluatorTestSuite) TestValid() {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedK error
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterK: 5}, expectedK: nil},
		{Name: "Negative", args: map[string]rule.Parameter{ParameterK: -1}, expectedK: errors.New("latestPushedK is less than zero")},
		{Name: "Big", args: map[string]rule.Parameter{ParameterK: 40000}, expectedK: errors.New("latestPushedK is too large")},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			err := Valid(tt.args)

			require.Equal(t, tt.expectedK, err)
		})
	}
}

func TestEvaluator(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}
