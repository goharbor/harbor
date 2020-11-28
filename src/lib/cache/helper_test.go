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

package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	cachetesting "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type Foobar struct {
	Foo string
	Bar int
}

type FetchOrSaveTestSuite struct {
	suite.Suite
}

func (suite *FetchOrSaveTestSuite) TestFetchInternalError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(c, "key", &str, func() (interface{}, error) {
		return "str", nil
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestBuildError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)

	var str string
	err := FetchOrSave(c, "key", &str, func() (interface{}, error) {
		return nil, fmt.Errorf("oops")
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestSaveError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(c, "key", &str, func() (interface{}, error) {
		return "str", nil
	})

	suite.Nil(err)
	suite.Equal("str", str)
}

func (suite *FetchOrSaveTestSuite) TestSaveCalledOnlyOneTime() {
	c := &cachetesting.Cache{}

	var data sync.Map

	mock.OnAnything(c, "Fetch").Return(func(key string, value interface{}) error {
		_, ok := data.Load(key)
		if ok {
			return nil
		}

		return ErrNotFound
	})

	mock.OnAnything(c, "Save").Return(func(key string, value interface{}, exp ...time.Duration) error {
		data.Store(key, value)

		return nil
	})

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			var str string
			FetchOrSave(c, "key", &str, func() (interface{}, error) {
				return "str", nil
			})
		}()
	}

	wg.Wait()

	c.AssertNumberOfCalls(suite.T(), "Save", 1)
}

func (suite *FetchOrSaveTestSuite) TestWithCache() {
	c := &cachetesting.Cache{}

	var data sync.Map

	mock.OnAnything(c, "Fetch").Return(func(key string, value interface{}) error {
		_, ok := data.Load(key)
		if ok {
			return nil
		}

		return ErrNotFound
	}).Run(func(args mock.Arguments) {
		key := args.Get(0).(string)
		value := args.Get(1)

		val, ok := data.Load(key)
		if ok {
			simpleCopy(value, val)
		}
	})

	mock.OnAnything(c, "Save").Return(func(key string, value interface{}, exp ...time.Duration) error {
		data.Store(key, value)

		return nil
	})

	var str string
	FetchOrSaveWithContext(NewContext(context.TODO(), c), "string", &str, func() (interface{}, error) {
		return "str", nil
	})
	suite.Equal("str", str)

	var i int
	FetchOrSaveWithContext(NewContext(nil, c), "int", &i, func() (interface{}, error) {
		return 10, nil
	})
	suite.Equal(10, i)

	{
		var foo Foobar
		bar := Foobar{Foo: "foo", Bar: 1}
		FetchOrSaveWithContext(context.TODO(), "struct-1", &foo, func() (interface{}, error) {
			return bar, nil
		})

		suite.Equal(bar, foo)
	}

	{
		var foo Foobar
		bar := Foobar{Foo: "foo", Bar: 1}
		FetchOrSaveWithContext(NewContext(nil, c), "struct-2", &foo, func() (interface{}, error) {
			return &bar, nil
		})

		suite.Equal(bar, foo)
	}
}

func (suite *FetchOrSaveTestSuite) TestWithouthCache() {
	{
		var str string
		FetchOrSaveWithContext(context.TODO(), "string", &str, func() (interface{}, error) {
			return "str", nil
		})
		suite.Equal("str", str)
	}

	{
		var str string
		FetchOrSaveWithContext(context.TODO(), "string", &str, func() (interface{}, error) {
			v := "str"
			return &v, nil
		})
		suite.Equal("str", str)
	}

	{
		var i int
		FetchOrSaveWithContext(context.TODO(), "int", &i, func() (interface{}, error) {
			return 1, nil
		})
		suite.Equal(1, i)
	}

	{
		var foo Foobar
		bar := Foobar{Foo: "foo", Bar: 1}
		FetchOrSaveWithContext(context.TODO(), "struct-1", &foo, func() (interface{}, error) {
			return bar, nil
		})

		suite.Equal(bar, foo)
	}

	{
		var foo Foobar
		bar := Foobar{Foo: "foo", Bar: 1}
		FetchOrSaveWithContext(context.TODO(), "struct-2", &foo, func() (interface{}, error) {
			return &bar, nil
		})

		suite.Equal(bar, foo)
	}

	{
		var foo Foobar
		err := FetchOrSaveWithContext(context.TODO(), "struct-3", &foo, func() (interface{}, error) {
			return nil, fmt.Errorf("oops")
		})

		suite.Error(err)
	}
}

func TestFetchOrSaveTestSuite(t *testing.T) {
	suite.Run(t, new(FetchOrSaveTestSuite))
}

func BenchmarkFetchOrSaveWithContextNoCache(b *testing.B) {
	var foo Foobar
	bar := Foobar{Foo: "foo", Bar: 1}

	ctx := context.TODO()
	for i := 0; i < b.N; i++ {
		FetchOrSaveWithContext(ctx, "struct", &foo, func() (interface{}, error) {
			return bar, nil
		})
	}
}

func BenchmarkAssign(b *testing.B) {
	var foo Foobar
	bar := Foobar{Foo: "foo", Bar: 1}

	for i := 0; i < b.N; i++ {
		foo = bar
	}

	_ = foo
}
