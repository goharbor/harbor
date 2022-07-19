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
	ctx context.Context
}

func (suite *FetchOrSaveTestSuite) SetupSuite() {
	suite.ctx = context.TODO()
}

func (suite *FetchOrSaveTestSuite) TestFetchInternalError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (interface{}, error) {
		return "str", nil
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestBuildError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (interface{}, error) {
		return nil, fmt.Errorf("oops")
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestSaveError() {
	c := &cachetesting.Cache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (interface{}, error) {
		return "str", nil
	})

	suite.Nil(err)
	suite.Equal("str", str)
}

func (suite *FetchOrSaveTestSuite) TestSaveCalledOnlyOneTime() {
	c := &cachetesting.Cache{}

	var data sync.Map

	mock.OnAnything(c, "Fetch").Return(func(ctx context.Context, key string, value interface{}) error {
		_, ok := data.Load(key)
		if ok {
			return nil
		}

		return ErrNotFound
	})

	mock.OnAnything(c, "Save").Return(func(ctx context.Context, key string, value interface{}, exp ...time.Duration) error {
		data.Store(key, value)

		return nil
	})

	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			var str string
			FetchOrSave(suite.ctx, c, "key", &str, func() (interface{}, error) {
				return "str", nil
			})
		}()
	}

	wg.Wait()

	c.AssertNumberOfCalls(suite.T(), "Save", 1)
}

func TestFetchOrSaveTestSuite(t *testing.T) {
	suite.Run(t, new(FetchOrSaveTestSuite))
}
