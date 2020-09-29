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
	"fmt"
	"github.com/goharbor/harbor/src/lib"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	ps "github.com/goharbor/harbor/src/pkg/proxy/secret"
)

func TestProxyCacheSecret(t *testing.T) {
	psc := &proxyCacheSecret{}

	// request without artifact info in context
	req, _ := http.NewRequest(http.MethodGet, "http://127.0.0.1/v2/library/hello-world/manifests/latest", nil)
	sc := psc.Generate(req)
	assert.Nil(t, sc)

	// request with invalid secret
	ctx := lib.WithArtifactInfo(req.Context(), lib.ArtifactInfo{
		Repository: "library/hello-world",
	})
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", fmt.Sprintf("Proxy-Cache-Secret %s", "invalid-secret"))
	sc = psc.Generate(req)
	assert.Nil(t, sc)

	// pass
	secret := ps.GetManager().Generate("library/hello-world")
	req.Header.Set("Authorization", fmt.Sprintf("Proxy-Cache-Secret %s", secret))
	sc = psc.Generate(req)
	assert.NotNil(t, sc)
}
