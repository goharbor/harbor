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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/testing/mock"
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
	c := &mockCache{}

	mock.OnAnything(c, "Fetch").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
		return "str", nil
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestBuildError() {
	c := &mockCache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
		return nil, fmt.Errorf("oops")
	})

	suite.Equal(fmt.Errorf("oops"), err)
}

func (suite *FetchOrSaveTestSuite) TestSaveError() {
	c := &mockCache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(fmt.Errorf("oops"))

	var str string
	err := FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
		return "str", nil
	})

	suite.Nil(err)
	suite.Equal("str", str)
}

func (suite *FetchOrSaveTestSuite) TestSaveCalledOnlyOneTime() {
	c := &mockCache{}

	var data sync.Map

	mock.OnAnything(c, "Fetch").Return(func(ctx context.Context, key string, value any) error {
		_, ok := data.Load(key)
		if ok {
			return nil
		}

		return ErrNotFound
	})

	mock.OnAnything(c, "Save").Return(func(ctx context.Context, key string, value any, exp ...time.Duration) error {
		data.Store(key, value)

		return nil
	})

	var wg sync.WaitGroup

	for range 1000 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			var str string
			FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
				return "str", nil
			})
		}()
	}

	wg.Wait()

	c.AssertNumberOfCalls(suite.T(), "Save", 1)
}

// Save must be called even if the HTTP request context is already canceled,
// because helper.go uses context.WithoutCancel before calling Save.
func (suite *FetchOrSaveTestSuite) TestSaveCalledEvenWhenContextCanceled() {
	c := &mockCache{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // simulate a canceled HTTP request context

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(nil)

	var str string
	err := FetchOrSave(ctx, c, "key-canceled", &str, func() (any, error) {
		return "str", nil
	})

	suite.Nil(err)
	suite.Equal("str", str)
	c.AssertNumberOfCalls(suite.T(), "Save", 1)
}

// On a concurrent cold miss, builder runs exactly once (singleflight dedup)
// and every concurrent caller receives the built value in its own pointer.
func (suite *FetchOrSaveTestSuite) TestConcurrentCallersShareResult() {
	c := &mockCache{}
	var builderCalls atomic.Int32

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(nil)

	const n = 100
	var wg sync.WaitGroup
	results := make([]string, n)

	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var str string
			FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
				builderCalls.Add(1)
				time.Sleep(10 * time.Millisecond) // widen the singleflight window
				return "built", nil
			})
			results[idx] = str
		}(i)
	}
	wg.Wait()

	suite.Equal(int32(1), builderCalls.Load(), "builder must run exactly once for concurrent callers")
	for _, r := range results {
		suite.Equal("built", r, "every caller must receive the built value")
	}
}

func TestFetchOrSaveTestSuite(t *testing.T) {
	suite.Run(t, new(FetchOrSaveTestSuite))
}
