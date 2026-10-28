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
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestTruncateUTF8(t *testing.T) {
	assert := assert.New(t)

	// empty string
	assert.Equal("", TruncateUTF8("", 10))

	// string within maxLen
	assert.Equal("hello", TruncateUTF8("hello", 10))
	assert.Equal("hello", TruncateUTF8("hello", 5))

	// ASCII truncation
	assert.Equal("hello...", TruncateUTF8("hello world", 8))

	// UTF-8 multibyte characters (3 bytes each)
	// "你好世界" is 12 bytes. maxLen 8 -> target 5 bytes, which cuts midway through '好' (bytes 3..5).
	// Backtracks to '好' start at byte 3, returning "你..." (6 bytes <= 8).
	res := TruncateUTF8("你好世界", 8)
	assert.Equal("你...", res)
	assert.True(len(res) <= 8)

	// 4-byte runes (emoji: 🚀 is 4 bytes)
	// "🚀🚀" is 8 bytes. maxLen 6 -> target 3 bytes, cuts inside first emoji.
	// Backtracks to 0, returning "...".
	resEmoji := TruncateUTF8("🚀🚀", 6)
	assert.Equal("...", resEmoji)
	assert.True(len(resEmoji) <= 6)

	// Invalid UTF-8 bytes are stripped
	resInvalid := TruncateUTF8("hello\xff\xfe world", 8)
	assert.Equal("hello...", resInvalid)

	// Very small maxLen
	assert.Equal("..", TruncateUTF8("hello", 2))
	assert.Equal("", TruncateUTF8("hello", 0))
	assert.Equal("", TruncateUTF8("hello", -1))
}

