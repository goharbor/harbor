package latestk

import (
	"strconv"
	"testing"

	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_New(t *testing.T) {
	tests := []struct {
		Name      string
		args      rule.Parameters
		expectedK int
	}{
		{Name: "Valid", args: map[string]rule.Parameter{ParameterK: 5}, expectedK: 5},
		{Name: "Default If Negative", args: map[string]rule.Parameter{ParameterK: -1}, expectedK: DefaultK},
		{Name: "Default If Not Set", args: map[string]rule.Parameter{}, expectedK: DefaultK},
		{Name: "Default If Wrong Type", args: map[string]rule.Parameter{ParameterK: "foo"}, expectedK: DefaultK},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			e := New(tt.args).(*evaluator)

			require.Equal(t, tt.expectedK, e.k)
		})
	}
}

func TestEvaluator_Process(t *testing.T) {
	data := []*res.Candidate{{}, {}, {}, {}, {}}

	tests := []struct {
		k        int
		expected int
	}{
		{k: 0, expected: 0},
		{k: 1, expected: 1},
		{k: 3, expected: 3},
		{k: 5, expected: 5},
		{k: 6, expected: 5},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.k), func(t *testing.T) {
			e := New(map[string]rule.Parameter{ParameterK: tt.k})

			result, err := e.Process(data)

			require.NoError(t, err)
			require.Len(t, result, tt.expected)
		})
	}
}
