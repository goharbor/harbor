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
package session

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	User string
	Pass string
}

func TestCodec(t *testing.T) {
	u := &User{User: "admin", Pass: "123456"}
	m := make(map[interface{}]interface{})
	m["user"] = u
	c, err := codec.Encode(m)
	assert.NoError(t, err, "encode should not error")

	v := make(map[interface{}]interface{})
	err = codec.Decode(c, &v)
	assert.NoError(t, err, "decode should not error")

	user, exist := v["user"]
	assert.True(t, exist, "user should exist")
	assert.True(t, assert.ObjectsAreEqualValues(u, user), "user should equal")
}
