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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyMutex(t *testing.T) {
	km := keyMutex{m: &sync.Map{}}
	key := "key"

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			km.Lock(key)
			km.Unlock(key)
		}()
	}
	wg.Wait()

	assert.Panics(t, func() {
		km.Unlock(key)
	})
}
