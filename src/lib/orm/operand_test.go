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

package orm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
)

type typed struct {
	ID           int64     `orm:"column(id)"`
	Name         string    `orm:"column(name)"`
	Score        float64   `orm:"column(score)"`
	Enabled      bool      `orm:"column(enabled)"`
	CreationTime time.Time `orm:"column(creation_time)"`
	Annotations  string    `orm:"column(annotations);type(jsonb)"`
}

func TestParseOperandKind(t *testing.T) {
	meta := parseModel(&typed{})
	cases := map[string]operandKind{
		"id":            kindInt,
		"name":          kindAny,
		"score":         kindFloat,
		"enabled":       kindBool,
		"creation_time": kindTime,
		"annotations":   kindJSON,
	}
	for column, kind := range cases {
		assert.Equal(t, kind, meta.Keys[column].Operand, column)
	}
}

func TestValidateOperands(t *testing.T) {
	parsed := func(query string) any {
		built, err := q.Build(query, "", 0, 0)
		require.NoError(t, err)
		for _, v := range built.Keywords {
			return v
		}
		return nil
	}

	cases := []struct {
		name  string
		kind  operandKind
		query string
		valid bool
	}{
		{"time range", kindTime, "k=[2020-01-01~2021-01-01]", true},
		{"time range with time of day", kindTime, "k=[2020-01-01T10:00:00~2021-01-01T10:00:00]", true},
		{"time range open ended", kindTime, "k=[2020-01-01~]", true},
		{"time range with garbage", kindTime, "k=[a~b]", false},
		{"time exact with garbage", kindTime, "k=abc", false},
		{"time or list with garbage", kindTime, "k={a b}", false},
		{"time exact", kindTime, "k=2020-01-01 10:00:00", true},
		{"time fuzzy", kindTime, "k=~abc", true},
		{"integer exact", kindInt, "k=1", true},
		{"integer exact with garbage", kindInt, "k=abc", false},
		{"integer range with garbage", kindInt, "k=[a~b]", false},
		{"integer or list", kindInt, "k={1 2 3}", true},
		{"integer or list with garbage", kindInt, "k={1 x}", false},
		{"integer out of range", kindInt, "k=99999999999999999999", false},
		{"integer given a time", kindInt, "k=2020-01-01T10:00:00", false},
		{"float exact", kindFloat, "k=1.5", true},
		{"float range", kindFloat, "k=[1~5]", true},
		{"float range with garbage", kindFloat, "k=[a~b]", false},
		{"bool exact", kindBool, "k=true", true},
		{"bool with garbage", kindBool, "k=maybe", false},
		{"json number", kindJSON, "k=1", true},
		{"json with garbage", kindJSON, "k=abc", false},
		{"string accepts anything", kindAny, "k=[a~b]", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateOperands(c.kind, parsed(c.query))
			if c.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSetFiltersRejectsInvalidOperand(t *testing.T) {
	meta := parseModel(&typed{})
	for _, keywords := range []map[string]any{
		{"creation_time": "abc"},
		{"creation_time": &q.Range{Min: "a", Max: "b"}},
		{"id": "abc"},
		{"id": &q.OrList{Values: []any{int64(1), "x"}}},
	} {
		_, err := setFilters(context.Background(), nil, &q.Query{Keywords: keywords}, meta)
		require.Error(t, err)
		assert.True(t, errors.IsErr(err, errors.BadRequestCode), "%v", keywords)
	}
}
