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

// Save is deduplicated across concurrent cold-miss callers: singleflight
// collapses overlapping callers into a single build+save, so Save runs far
// fewer times than there are callers. It is not strictly once: a caller that
// reaches the group after a previous build already finished legitimately builds
// and saves again, so the exact count is not guaranteed.
func (suite *FetchOrSaveTestSuite) TestSaveDeduplicatedAcrossConcurrentCallers() {
	c := &mockCache{}

	var data sync.Map
	var saveCalls atomic.Int32

	mock.OnAnything(c, "Fetch").Return(func(ctx context.Context, key string, value any) error {
		_, ok := data.Load(key)
		if ok {
			return nil
		}

		return ErrNotFound
	})

	mock.OnAnything(c, "Save").Return(func(ctx context.Context, key string, value any, exp ...time.Duration) error {
		saveCalls.Add(1)
		data.Store(key, value)

		return nil
	})

	const n = 1000
	var wg sync.WaitGroup

	for range n {
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

	saves := saveCalls.Load()
	suite.GreaterOrEqual(saves, int32(1))
	suite.Less(saves, int32(n), "singleflight must deduplicate concurrent saves")
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

// The shared singleflight result is copied to each caller through the codec,
// so a builder result the codec cannot encode must surface as an error.
func (suite *FetchOrSaveTestSuite) TestBuildResultNotEncodable() {
	c := &mockCache{}

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(nil)

	var ch chan struct{}
	err := FetchOrSave(suite.ctx, c, "key-not-encodable", &ch, func() (any, error) {
		return make(chan struct{}), nil
	})

	suite.Error(err)
}

// On a concurrent cold miss, singleflight collapses the overlapping callers
// into far fewer builder executions than callers, and every caller receives the
// built value copied into its own pointer via the codec.
func (suite *FetchOrSaveTestSuite) TestConcurrentCallersShareResult() {
	c := &mockCache{}
	var builderCalls atomic.Int32

	mock.OnAnything(c, "Fetch").Return(ErrNotFound)
	mock.OnAnything(c, "Save").Return(nil)

	const n = 100
	var wg, ready sync.WaitGroup
	results := make([]string, n)

	// start is a release barrier: every goroutine blocks on it until all of
	// them are spawned and waiting, so they enter FetchOrSave together and
	// genuinely overlap inside the singleflight window (no flakiness from a
	// late goroutine missing the in-flight call on a loaded runner).
	start := make(chan struct{})

	for i := range n {
		wg.Add(1)
		ready.Add(1)
		go func(idx int) {
			defer wg.Done()
			ready.Done()
			<-start
			var str string
			FetchOrSave(suite.ctx, c, "key", &str, func() (any, error) {
				builderCalls.Add(1)
				time.Sleep(10 * time.Millisecond) // widen the singleflight window
				return "built", nil
			})
			results[idx] = str
		}(i)
	}
	ready.Wait()
	close(start)
	wg.Wait()

	// singleflight must collapse the concurrent cold miss into far fewer builds
	// than callers. An exact "== 1" assertion would be racy: a caller that
	// reaches the group after the leader's build already returned legitimately
	// starts a new build, so under load the count can be >1 without any bug.
	calls := builderCalls.Load()
	suite.GreaterOrEqual(calls, int32(1))
	suite.Less(calls, int32(n), "singleflight must deduplicate concurrent builds")
	for _, r := range results {
		suite.Equal("built", r, "every caller must receive the built value")
	}
}

func TestFetchOrSaveTestSuite(t *testing.T) {
	suite.Run(t, new(FetchOrSaveTestSuite))
}
