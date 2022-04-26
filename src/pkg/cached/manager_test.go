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

package cached

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectKey(t *testing.T) {
	ok := NewObjectKey("artifact")
	// valid case
	s, err := ok.Format("id", 100, "digest", "9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0")
	assert.NoError(t, err, "format should not error")
	assert.Equal(t, "artifact:id:100:digest:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0", s)
	// invalid case
	_, err = ok.Format("id")
	assert.Error(t, err, "invalid length should error")
	_, err = ok.Format(1, 1)
	assert.Error(t, err, "invalid key type should error")
	_, err = ok.Format("id", struct{}{})
	assert.Error(t, err, "invalid value type should error")
}
