package lastx

import (
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_New(t *testing.T) {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedX int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterX: 3}, expectedX: 3},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterX: -3}, expectedX: DefaultX},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedX: DefaultX},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{}, expectedX: DefaultX},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedX, e.x)
		})
	}
}

func TestEvaluator_Process(t *testing.T) {
	now := time.Now().UTC()
	data := []*res.Candidate{
		{PushedTime: now.Add(time.Duration(1*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(2*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(3*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(4*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(5*-24) * time.Hour).Unix()},
		{PushedTime: now.Add(time.Duration(99*-24) * time.Hour).Unix()},
	}

	tests := []struct {
		days     int
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
		t.Run(fmt.Sprintf("%d days - should keep %d", tt.days, tt.expected), func(t *testing.T) {
			e := New(rule.Parameters{ParameterX: tt.days})

			result, err := e.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)
		})
	}
}
