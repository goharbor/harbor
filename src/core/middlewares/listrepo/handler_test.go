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

package listrepo

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMatchListRepos(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/_catalog", nil)
	res1 := matchListRepos(req1)
	assert.False(res1, "%s %v is not a request to list repos", req1.Method, req1.URL)

	req2, _ := http.NewRequest("GET", "http://127.0.0.1:5000/v2/_catalog", nil)
	res2 := matchListRepos(req2)
	assert.True(res2, "%s %v is a request to list repos", req2.Method, req2.URL)

	req3, _ := http.NewRequest("GET", "https://192.168.0.5:443/v1/_catalog", nil)
	res3 := matchListRepos(req3)
	assert.False(res3, "%s %v is not a request to pull manifest", req3.Method, req3.URL)

}
