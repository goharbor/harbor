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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFetchOrSaveConcurrency(t *testing.T) {
	c := newMockCache(t)
	key := "test-key"
	var callCount int32

	builder := func() (any, error) {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(100 * time.Millisecond)
		return "test-value", nil
	}

	c.On("Fetch", mock.Anything, key, mock.Anything).Return(ErrNotFound)
	c.On("Save", mock.Anything, key, "test-value").Return(nil).Once()

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var value string
			err := FetchOrSave(context.Background(), c, key, &value, builder)
			assert.NoError(t, err)
			assert.Equal(t, "test-value", value)
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), callCount, "builder should be called only once")
	c.AssertNumberOfCalls(t, "Save", 1)
}
