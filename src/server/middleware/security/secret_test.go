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

package security

import (
	"net/http"
	"testing"

	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecret(t *testing.T) {
	secret := secret{}

	// contains no secret
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	ctx := secret.Generate(req)
	assert.Nil(t, ctx)

	// contains secret
	commonsecret.AddToRequest(req, "secret")
	ctx = secret.Generate(req)
	assert.NotNil(t, ctx)
}
