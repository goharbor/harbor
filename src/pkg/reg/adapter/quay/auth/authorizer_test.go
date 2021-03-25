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

package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthorizer(t *testing.T) {
	req1, err := http.NewRequest("GET", "http://1.1.1.1/v2/_catalog", nil)
	assert.NoError(t, err)
	assert.True(t, isCatalog(req1))

	req2, err := http.NewRequest("GET", "http://1.1.1.1/v2/library/nginx/tags/list", nil)
	assert.NoError(t, err)
	assert.False(t, isCatalog(req2))
}
