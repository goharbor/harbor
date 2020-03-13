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
// limitations under the License

package artifactinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	cases := []struct {
		input  string
		expect map[string]string
		match  bool
	}{
		{
			input:  "/api/projects",
			expect: map[string]string{},
			match:  false,
		},
		{
			input:  "/v2/_catalog",
			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/no-project-repo/tags/list",
			expect: map[string]string{
				middleware.RepositorySubexp: "no-project-repo",
			},
			match: true,
		},
		{
			input: "/v2/development/golang/manifests/sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			expect: map[string]string{
				middleware.RepositorySubexp: "development/golang",
				middleware.ReferenceSubexp:  "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				middleware.DigestSubexp:     "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			},
			match: true,
		},
		{
			input: "/v2/development/golang/manifests/shaxxx:**********************************************************************************************************************************",

			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/multi/sector/repository/blobs/sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			expect: map[string]string{
				middleware.RepositorySubexp: "multi/sector/repository",
				middleware.DigestSubexp:     "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
			},
			match: true,
		},
		{
			input:  "/v2/blobs/uploads",
			expect: map[string]string{},
			match:  false,
		},
		{
			input: "/v2/library/ubuntu/blobs/uploads",
			expect: map[string]string{
				middleware.RepositorySubexp: "library/ubuntu",
			},
			match: true,
		},
		{
			input: "/v2/library/ubuntu/blobs/uploads/?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=old/ubuntu",
			expect: map[string]string{
				middleware.RepositorySubexp: "library/ubuntu",
				blobMountDigest:             "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				blobMountRepo:               "old/ubuntu",
			},
			match: true,
		},
		{
			input: "/v2/library/centos/blobs/uploads/u-12345",
			expect: map[string]string{
				middleware.RepositorySubexp: "library/centos",
			},
			match: true,
		},
	}

	for _, c := range cases {
		url, err := url.Parse(c.input)
		if err != nil {
			panic(err)
		}
		e, m := parse(url)
		assert.Equal(t, c.match, m)
		assert.Equal(t, c.expect, e)
	}
}

type handler struct {
	ctx context.Context
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	h.ctx = req.Context()
}
func TestPopulateArtifactInfo(t *testing.T) {

	none := internal.ArtifactInfo{}
	cases := []struct {
		req *http.Request
		sc  int
		art internal.ArtifactInfo
	}{
		{
			req: httptest.NewRequest(http.MethodDelete, "/v2/hello-world/manifests/latest", nil),
			sc:  http.StatusBadRequest,
			art: none,
		},
		{
			req: httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil),
			sc:  http.StatusOK,
			art: internal.ArtifactInfo{
				Repository:  "library/hello-world",
				Reference:   "latest",
				ProjectName: "library",
				Tag:         "latest",
			},
		},
		{
			req: httptest.NewRequest(http.MethodPost, "/v2/library/ubuntu/blobs/uploads/?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=no-project", nil),
			sc:  http.StatusBadRequest,
			art: none,
		},
		{
			req: httptest.NewRequest(http.MethodPost, "/v2/library/ubuntu/blobs/uploads/?from=old/ubuntu&mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f", nil),
			sc:  http.StatusOK,
			art: internal.ArtifactInfo{
				Repository:           "library/ubuntu",
				ProjectName:          "library",
				BlobMountRepository:  "old/ubuntu",
				BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				BlobMountProjectName: "old",
			},
		},
		{
			req: httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f", nil),
			sc:  http.StatusOK,
			art: internal.ArtifactInfo{
				Repository:  "library/hello-world",
				Reference:   "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				Digest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				ProjectName: "library",
			},
		},
	}
	next := &handler{}

	for _, tt := range cases {
		rec := httptest.NewRecorder()

		Middleware()(next).ServeHTTP(rec, tt.req)
		assert.Equal(t, tt.sc, rec.Code)
		if tt.art != none {
			a := internal.GetArtifactInfo(next.ctx)
			assert.NotEqual(t, none, a)
			assert.Equal(t, tt.art, a)
		}
	}
}
