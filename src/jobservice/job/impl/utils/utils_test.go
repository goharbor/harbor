// Copyright 2018 The Harbor Authors
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
package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	assert := assert.New(t)
	os.Setenv("", "")
	_, err := GetClient()
	assert.NotNil(err, "Error should be thrown if secret is not set")
	os.Setenv("JOBSERVICE_SECRET", "thesecret")
	c, err := GetClient()
	assert.Nil(err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Header.Get("Authorization")
		assert.Equal(secret.HeaderPrefix+"thesecret", v)
	}))
	defer ts.Close()
	c.Get(ts.URL)

	os.Setenv("", "")
	_, err = GetClient()
	assert.Nil(err, "Error should be nil once client is initialized")

}
