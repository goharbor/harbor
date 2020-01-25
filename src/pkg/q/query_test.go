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

package q

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCopy(t *testing.T) {
	// nil
	q := Copy(nil)
	assert.Nil(t, q)

	// not nil
	query := &Query{
		PageNumber: 1,
		PageSize:   10,
		Keywords: map[string]interface{}{
			"key": "value",
		},
	}
	q = Copy(query)
	require.NotNil(t, q)
	assert.Equal(t, int64(1), q.PageNumber)
	assert.Equal(t, int64(10), q.PageSize)
	assert.Equal(t, "value", q.Keywords["key"].(string))
	// changes for the copy doesn't effect the original one
	q.PageSize = 20
	q.Keywords["key"] = "value2"
	assert.Equal(t, int64(10), query.PageSize)
	assert.Equal(t, "value", query.Keywords["key"].(string))
}
