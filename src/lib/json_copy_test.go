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

type jsonCopyFoo struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONCopy(t *testing.T) {
	assert := assert.New(t)

	{
		var m map[string]interface{}
		foo := &jsonCopyFoo{
			Name: "foo",
			Age:  1,
		}

		assert.Nil(m)
		assert.Nil(JSONCopy(&m, foo))
		assert.NotNil(m)
		assert.Len(m, 2)
	}

	{
		var m map[string]interface{}
		var foo *jsonCopyFoo

		assert.Nil(m)
		assert.Nil(JSONCopy(&m, foo))
		assert.Nil(m)
	}

	{
		m := map[string]interface{}{
			"name": "foo",
			"age":  1,
		}
		var foo *jsonCopyFoo
		assert.Nil(JSONCopy(&foo, &m))
		assert.NotNil(foo)
		assert.Equal("foo", foo.Name)
		assert.Equal(1, foo.Age)
	}

	{
		assert.Error(JSONCopy(nil, JSONCopy))
	}
}
