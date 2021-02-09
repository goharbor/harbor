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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryUntil(t *testing.T) {
	assert := assert.New(t)

	i := 0
	f1 := func() error {
		i++
		return fmt.Errorf("failed")
	}
	assert.Error(RetryUntil(f1, RetryInitialInterval(time.Second), RetryMaxInterval(time.Second), RetryTimeout(time.Second*5)))
	// f1 called time     0s - sleep - 1s - sleep - 2s - sleep - 3s - sleep - 4s - sleep - 5s
	// i after f1 called  1            2            3            4            5            6
	// the i may be 5 or 6 depend on timeout or default which is seleted by the select statement
	assert.LessOrEqual(i, 6)

	f2 := func() error {
		return nil
	}
	assert.Nil(RetryUntil(f2))

	i = 0
	f3 := func() error {
		defer func() {
			i++
		}()

		if i < 2 {
			return fmt.Errorf("failed")
		}
		return nil
	}
	assert.Nil(RetryUntil(f3))

	RetryUntil(
		f1,
		RetryTimeout(time.Second*5),
		RetryCallback(func(err error, sleep time.Duration) {
			fmt.Printf("failed to exec f1 retry after %s : %v\n", sleep, err)
		}),
	)
}
