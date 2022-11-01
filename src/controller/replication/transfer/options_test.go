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

package transfer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	// test default options
	o := NewOptions()
	assert.Equal(t, int32(0), o.Speed)
	assert.Equal(t, false, o.CopyByChunk)

	// test with options
	// with speed
	withSpeed := WithSpeed(1024)
	// with copy by chunk
	withCopyByChunk := WithCopyByChunk(true)
	o = NewOptions(withSpeed, withCopyByChunk)
	assert.Equal(t, int32(1024), o.Speed)
	assert.Equal(t, true, o.CopyByChunk)
}
