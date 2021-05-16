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

package selector

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/stretchr/testify/suite"
)

type foo struct {
	Field1 string `orm:"-"`
	Field2 string `orm:"column(customized_field2)" filter:"false"`
	Field3 string `sort:"false"`
	Field4 string `sort:"default:desc"`
}

type SelectorTestSuite struct {
	suite.Suite
}

func (suite *SelectorTestSuite) TestSelectOption() {
	s := New(context.TODO())

	sql, args, err := s.Build(
		From(&foo{}),
		Select(&foo{}),
	)

	suite.Nil(err)
	suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0", sql)
	suite.Empty(args)
}

func (suite *SelectorTestSuite) TestCountOption() {
	s := New(context.TODO())

	sql, args, err := s.Build(
		From(&foo{}),
		Count(),
	)

	suite.Nil(err)
	suite.Equal("SELECT COUNT(*) FROM foo T0", sql)
	suite.Empty(args)
}

func (suite *SelectorTestSuite) TestFilterOption() {
	{
		s := New(context.TODO())

		sql, args, err := s.Build(
			From(&foo{}),
			Select(&foo{}),
			Filter(&foo{}, map[string]interface{}{
				"field4": q.NewOrList([]interface{}{"a", "b"}),
				"Field3": "f",
				"unknow": "f",
			}),
		)

		suite.Nil(err)
		suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 WHERE T0.field3 = $1 AND T0.field4 IN ($2, $3)", sql)
		suite.Len(args, 3)
	}

	{
		s := New(context.TODO())

		sql, args, err := s.Build(
			From(&foo{}),
			Select(&foo{}),
			Filter(&foo{}, map[string]interface{}{
				"field4": q.NewFuzzyMatchValue("a"),
			}),
		)

		suite.Nil(err)
		suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 WHERE UPPER(T0.field4::text) LIKE UPPER($1)", sql)
		suite.Len(args, 1)
		suite.Equal("%a%", args[0])
	}

	{
		s := New(context.TODO())

		sql, args, err := s.Build(
			From(&foo{}),
			Select(&foo{}),
			Filter(&foo{}, map[string]interface{}{
				"field4": q.NewRange("a", "b"),
			}),
		)

		suite.Nil(err)
		suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 WHERE T0.field4 >= $1 AND T0.field4 <= $2", sql)
		suite.Len(args, 2)
		suite.Equal("a", args[0])
		suite.Equal("b", args[1])
	}
}

func (suite *SelectorTestSuite) TestSortsOption() {
	{
		s := New(context.TODO())

		sql, args, err := s.Build(
			From(&foo{}),
			Select(&foo{}),
			Sorts(&foo{}, []*q.Sort{
				{Key: "Field2", DESC: true},
				{Key: "Field3"},
			}),
		)

		suite.Nil(err)
		suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 ORDER BY T0.customized_field2 DESC", sql)
		suite.Empty(args)
	}
	{
		s := New(context.TODO())

		sql, args, err := s.Build(
			From(&foo{}),
			Select(&foo{}),
			Sorts(&foo{}, nil),
		)

		suite.Nil(err)
		suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 ORDER BY T0.field4 DESC", sql)
		suite.Empty(args)
	}
}

func (suite *SelectorTestSuite) TestPaginationOption() {
	s := New(context.TODO())

	sql, args, err := s.Build(
		From(&foo{}),
		Select(&foo{}),
		Pagination(3, 10),
	)

	suite.Nil(err)
	suite.Equal("SELECT T0.customized_field2, T0.field3, T0.field4 FROM foo T0 LIMIT 10 OFFSET 20", sql)
	suite.Empty(args)
}

func TestSelectorTestSuite(t *testing.T) {
	suite.Run(t, &SelectorTestSuite{})
}
