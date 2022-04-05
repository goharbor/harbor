// Copyright 2018 Project Harbor Authors
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

package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLinkEntry(t *testing.T) {
	u1, err := SetLinkHeader("/v2/hello-wrold/tags/list", 10, "v10")
	assert.Nil(t, err)
	assert.Equal(t, u1, "</v2/hello-wrold/tags/list?last=v10&n=10>; rel=\"next\"")

	u2, err := SetLinkHeader("/v2/hello-wrold/tags/list", 5, "v5")
	assert.Nil(t, err)
	assert.Equal(t, u2, "</v2/hello-wrold/tags/list?last=v5&n=5>; rel=\"next\"")

}

func TestIndexString(t *testing.T) {
	a := []string{"B", "A", "C", "E"}

	assert.True(t, IndexString(a, "E") == 3)
	assert.True(t, IndexString(a, "B") == 1)
	assert.True(t, IndexString(a, "A") == 0)
	assert.True(t, IndexString(a, "C") == 2)

	assert.True(t, IndexString(a, "Z") == -1)
	assert.True(t, IndexString(a, "") == -1)
}

func Test_pickItems(t *testing.T) {
	p := func(n int) *int {
		return &n
	}

	type args struct {
		items []string
		n     *int
		last  string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 string
	}{
		{
			"no parameters",
			args{[]string{"a", "b", "c", "d"}, nil, ""},
			[]string{"a", "b", "c", "d"},
			"",
		},
		{
			"n=0",
			args{[]string{"a", "b", "c", "d"}, p(0), ""},
			[]string{},
			"",
		},
		{
			"n=1",
			args{[]string{"a", "b", "c", "d"}, p(1), ""},
			[]string{"a"},
			"a",
		},
		{
			"n=2",
			args{[]string{"a", "b", "c", "d"}, p(2), ""},
			[]string{"a", "b"},
			"b",
		},
		{
			"n=4", // n is the count of items
			args{[]string{"a", "b", "c", "d"}, p(4), ""},
			[]string{"a", "b", "c", "d"},
			"",
		},
		{
			"n=5", // n is bigger than the count of items
			args{[]string{"a", "b", "c", "d"}, p(5), ""},
			[]string{"a", "b", "c", "d"},
			"",
		},
		{
			"last=a",
			args{[]string{"a", "b", "c", "d"}, nil, "a"},
			[]string{"b", "c", "d"},
			"",
		},
		{
			"last=d",
			args{[]string{"a", "b", "c", "d"}, nil, "d"},
			[]string{},
			"",
		},
		{
			"n=1 last=a",
			args{[]string{"a", "b", "c", "d"}, p(1), "a"},
			[]string{"b"},
			"b",
		},
		{
			"n=2 last=a",
			args{[]string{"a", "b", "c", "d"}, p(2), "a"},
			[]string{"b", "c"},
			"c",
		},
		{
			"n=3 last=a", // just the left n items
			args{[]string{"a", "b", "c", "d"}, p(3), "a"},
			[]string{"b", "c", "d"},
			"",
		},
		{
			"n=4 last=a", // left items is less than n
			args{[]string{"a", "b", "c", "d"}, p(4), "a"},
			[]string{"b", "c", "d"},
			"",
		},
		{
			"n=1 last=d", // last is the last element of the items
			args{[]string{"a", "b", "c", "d"}, p(1), "d"},
			[]string{},
			"",
		},
		{
			"last=c", // last not found and there is an elem bigger than it
			args{[]string{"a", "b", "d", "e"}, nil, "c"},
			[]string{"d", "e"},
			"",
		},
		{
			"last=e", // last not found and there isn't an elem bigger than it
			args{[]string{"a", "b", "c", "d"}, nil, "e"},
			[]string{},
			"",
		},
		{
			"last=e n=10",
			args{[]string{"a", "b", "c", "d"}, p(10), "e"},
			[]string{},
			"",
		},
		{
			"one item",
			args{[]string{"a"}, nil, ""},
			[]string{"a"},
			"",
		},
		{
			"one item n=2",
			args{[]string{"a"}, p(2), ""},
			[]string{"a"},
			"",
		},
		{
			"two items",
			args{[]string{"a", "b"}, nil, ""},
			[]string{"a", "b"},
			"",
		},
		{
			"three items",
			args{[]string{"a", "b", "c"}, nil, ""},
			[]string{"a", "b", "c"},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := pickItems(tt.args.items, tt.args.n, tt.args.last)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pickItems() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("pickItems() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_sortedAndUniqueItems(t *testing.T) {
	type args struct {
		items []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"nil",
			args{nil},
			nil,
		},
		{
			"no item",
			args{[]string{}},
			[]string{},
		},
		{
			"one item",
			args{[]string{"a"}},
			[]string{"a"},
		},
		{
			"duplicate items",
			args{[]string{"a", "a", "a"}},
			[]string{"a"},
		},
		{
			"unordered and duplicate items",
			args{[]string{"a", "c", "a", "b", "a"}},
			[]string{"a", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortedAndUniqueItems(tt.args.items); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortedAndUniqueItems() = %v, want %v", got, tt.want)
			}
		})
	}
}
