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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTruncate(t *testing.T) {
	assert := assert.New(t)

	// n > length
	str := "abc"
	suffix := "#123"
	n := 10
	assert.Equal("abc#123", Truncate(str, suffix, n))

	// n == length
	str = "abc"
	suffix = "#123"
	n = 7
	assert.Equal("abc#123", Truncate(str, suffix, n))

	// n < length
	str = "abc"
	suffix = "#123"
	n = 5
	assert.Equal("a#123", Truncate(str, suffix, n))
}
