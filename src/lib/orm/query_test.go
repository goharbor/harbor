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

package orm

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListQueriableCols(t *testing.T) {
	type model struct {
		Field1 string `orm:"column(field1)" json:"field1"`
		Field2 string `orm:"column(customized_field2)"`
		Field3 string
		Field4 string `orm:"column(field4)"`
	}
	// without ignoring columns
	cols := listQueriableCols(&model{})
	require.Len(t, cols, 7)
	_, exist := cols["Field1"]
	assert.True(t, exist)
	_, exist = cols["field1"]
	assert.True(t, exist)
	_, exist = cols["Field2"]
	assert.True(t, exist)
	_, exist = cols["customized_field2"]
	assert.True(t, exist)
	_, exist = cols["Field3"]
	assert.True(t, exist)
	_, exist = cols["Field4"]
	assert.True(t, exist)
	_, exist = cols["field4"]
	assert.True(t, exist)

	// with ignoring columns
	cols = listQueriableCols(&model{}, "Field4")
	require.Len(t, cols, 5)
	_, exist = cols["Field1"]
	assert.True(t, exist)
	_, exist = cols["field1"]
	assert.True(t, exist)
	_, exist = cols["Field2"]
	assert.True(t, exist)
	_, exist = cols["customized_field2"]
	assert.True(t, exist)
	_, exist = cols["Field3"]
	assert.True(t, exist)
}
