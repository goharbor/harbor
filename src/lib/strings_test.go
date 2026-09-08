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

package lib

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimLineBreaks(t *testing.T) {
	assert := assert.New(t)

	s := `
abc
def

`
	expect := "abcdef"
	actual := TrimLineBreaks(s)
	assert.Equal(expect, actual, "should trim line breaks")
}

func TestTitle(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"upper case", args{"MANUAL"}, "Manual"},
		{"lower case", args{"manual"}, "Manual"},
		{"empty", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, Title(tt.args.s), "Title(%v)", tt.args.s)
		})
	}
}

func TestSliceToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "Single element",
			input:    []string{"hello"},
			expected: []string{"HELLO"},
		},
		{
			name:     "Multiple elements",
			input:    []string{"hello", "world", "go"},
			expected: []string{"HELLO", "WORLD", "GO"},
		},
		{
			name:     "Already uppercase",
			input:    []string{"HELLO", "WORLD"},
			expected: []string{"HELLO", "WORLD"},
		},
		{
			name:     "Mixed case",
			input:    []string{"Hello", "World", "Go"},
			expected: []string{"HELLO", "WORLD", "GO"},
		},
		{
			name:     "With special characters",
			input:    []string{"hello!", "world?", "go#"},
			expected: []string{"HELLO!", "WORLD?", "GO#"},
		},
		{
			name:     "With numbers",
			input:    []string{"hello123", "world456"},
			expected: []string{"HELLO123", "WORLD456"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SliceToUpper(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("SliceToUpper() = %v, want %v", result, tt.expected)
			}
		})
	}
}
