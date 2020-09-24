package lastx

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/selector"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
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
		expectedX int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterX: float64(3)}, expectedX: 3},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterX: float64(-3)}, expectedX: DefaultX},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedX: DefaultX},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{}, expectedX: DefaultX},
	}

	for _, tt := range tests {
		e.T().Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedX, e.x)
		})
	}
}

func (e *EvaluatorTestSuite) TestProcess() {
	now := time.Now().UTC()
	data := []*selector.Candidate{
		{PushedTime: now.Add(time.Duration(1*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(2*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(3*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(4*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(5*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(99*-24) * time.Hour).Unix()},
	}

	tests := []struct {
		days     float64
		expected int
	}{
		{days: 0, expected: 0},
		{days: 1, expected: 0},
		{days: 2, expected: 1},
		{days: 3, expected: 2},
		{days: 4, expected: 3},
		{days: 5, expected: 4},
		{days: 6, expected: 5},
		{days: 7, expected: 5},
	}

	for _, tt := range tests {
		e.T().Run(fmt.Sprintf("%v days - should keep %d", tt.days, tt.expected), func(t *testing.T) {
			e := New(rule.Parameters{ParameterX: tt.days})

			result, err := e.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)
		})
	}
}

func TestEvaluatorSuite(t *testing.T) {
	suite.Run(t, &EvaluatorTestSuite{})
}
