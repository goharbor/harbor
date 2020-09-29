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

package secret

import (
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestFromRequest(t *testing.T) {
	assert := assert.New(t)
	secret := "mysecret"
	req, _ := http.NewRequest("GET", "http://test.com", nil)
	req.Header.Add("Authorization", "Harbor-Secret "+secret)
	assert.Equal(secret, FromRequest(req))
	assert.Equal("", FromRequest(nil))
}

func TestAddToRequest(t *testing.T) {
	assert := assert.New(t)
	secret := "mysecret"
	req, _ := http.NewRequest("GET", "http://test.com", nil)
	err := AddToRequest(req, secret)
	assert.Nil(err)
	assert.Equal(secret, FromRequest(req))
	assert.NotNil(AddToRequest(nil, secret))
}
