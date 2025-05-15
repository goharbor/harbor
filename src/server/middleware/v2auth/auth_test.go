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

package v2auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestMain(m *testing.M) {
	testutils.InitDatabaseFromEnv()
	ctl := &projecttesting.Controller{}

	mockGet := func(ctx context.Context,
		projectIDOrName any, options ...project.Option) (*proModels.Project, error) {
		name := projectIDOrName.(string)
		id, _ := strconv.Atoi(strings.TrimPrefix(name, "project_"))
		if id == 0 {
			return nil, fmt.Errorf("%s not found", name)
		}
		return &proModels.Project{
			ProjectID: int64(id),
			Name:      name,
		}, nil
	}
	mock.OnAnything(ctl, "Get").Return(
		func(ctx context.Context,
			projectIDOrName any, options ...project.Option) *proModels.Project {
			p, _ := mockGet(ctx, projectIDOrName, options...)
			return p
		},
		func(ctx context.Context,
			projectIDOrName any, options ...project.Option) error {
			_, err := mockGet(ctx, projectIDOrName, options...)
			return err
		},
	)

	checker = reqChecker{
		ctl: ctl,
	}
	conf := map[string]any{
		common.ExtEndpoint: "https://harbor.test",
		common.CoreURL:     "https://harbor.core:8443",
	}
	config.InitWithSettings(conf)
	if rc := m.Run(); rc != 0 {
		os.Exit(rc)
	}
}

func TestMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)
	mock.OnAnything(sc, "Can").Return(func(ctx context.Context, action types.Action, resource types.Resource) bool {
		perms := map[string]map[rbac.Action]struct{}{
			"/project/1/repository": {
				rbac.ActionPull: {},
				rbac.ActionPush: {},
			},
			"/project/2/repository": {
				rbac.ActionPull: {},
			},
		}
		m, ok := perms[resource.String()]
		if !ok {
			return false
		}
		_, ok = m[action]
		return ok
	})

	baseCtx := security.NewContext(context.Background(), sc)
	ar1 := lib.ArtifactInfo{
		Repository:  "project_1/hello-world",
		Reference:   "v1",
		ProjectName: "project_1",
	}
	ar2 := lib.ArtifactInfo{
		Repository:  "library/ubuntu",
		Reference:   "14.04",
		ProjectName: "library",
	}
	ar3 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_2/ubuntu",
		BlobMountProjectName: "project_2",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}
	ar4 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_3/ubuntu",
		BlobMountProjectName: "project_3",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}
	ar5 := lib.ArtifactInfo{
		Repository:           "project_1/ubuntu",
		Reference:            "14.04",
		ProjectName:          "project_1",
		BlobMountRepository:  "project_0/ubuntu",
		BlobMountProjectName: "project_0",
		BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	}

	ctx1 := lib.WithArtifactInfo(baseCtx, ar1)
	ctx2 := lib.WithArtifactInfo(baseCtx, ar2)
	ctx3 := lib.WithArtifactInfo(baseCtx, ar3)
	ctx4 := lib.WithArtifactInfo(baseCtx, ar4)
	ctx5 := lib.WithArtifactInfo(baseCtx, ar5)
	req1a, _ := http.NewRequest(http.MethodGet, "/v2/project_1/hello-world/manifest/v1", nil)
	req1b, _ := http.NewRequest(http.MethodDelete, "/v2/project_1/hello-world/manifest/v1", nil)
	req1c, _ := http.NewRequest(http.MethodHead, "/v2/project_1/hello-world/manifest/v1", nil)
	req1d, _ := http.NewRequest(http.MethodGet, "/v2/project_1/hello-world/manifest/v1", nil)
	req1d.Header.Set("Authorization", "Bearer xxx")
	req1e, _ := http.NewRequest(http.MethodHead, "/v2/project_1/hello-world/manifest/v1", nil)
	req1e.Header.Set("Authorization", "Bearer xxx")
	req2, _ := http.NewRequest(http.MethodGet, "/v2/library/ubuntu/manifest/14.04", nil)
	req3, _ := http.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	req4, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_2/ubuntu", nil)
	req5, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_3/ubuntu", nil)
	req6, _ := http.NewRequest(http.MethodPost, "/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_0/ubuntu", nil)
	req7, _ := http.NewRequest(http.MethodPost, "/v2/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_0/ubuntu", nil)

	cases := []struct {
		input  *http.Request
		status int
	}{
		{
			input:  req1a.WithContext(ctx1),
			status: http.StatusUnauthorized,
		},
		{
			input:  req1b.WithContext(ctx1),
			status: http.StatusUnauthorized,
		},
		{
			input:  req1c.WithContext(ctx1),
			status: http.StatusUnauthorized,
		},
		{
			input:  req1d.WithContext(ctx1),
			status: http.StatusOK,
		},
		{
			input:  req1e.WithContext(ctx1),
			status: http.StatusOK,
		},
		{
			input:  req2.WithContext(ctx2),
			status: http.StatusUnauthorized,
		},
		{
			input:  req3.WithContext(baseCtx),
			status: http.StatusUnauthorized,
		},
		{
			input:  req4.WithContext(ctx3),
			status: http.StatusOK,
		},
		{
			input:  req5.WithContext(ctx4),
			status: http.StatusUnauthorized,
		},
		{
			input:  req6.WithContext(ctx5),
			status: http.StatusUnauthorized,
		},
		{
			input:  req7.WithContext(ctx5),
			status: http.StatusUnauthorized,
		},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		t.Logf("req : %s, %s", c.input.Method, c.input.URL)
		Middleware()(next).ServeHTTP(rec, c.input)
		assert.Equal(t, c.status, rec.Result().StatusCode)
	}
}

func TestGetChallenge(t *testing.T) {
	cases := []struct {
		name      string
		request   *http.Request
		challenge string
	}{
		{
			name: "Regular login request to '/v2' should return challenge whose realm is token URL with the Host header in Request",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/", nil)
				return req
			}(),
			challenge: `Bearer realm="https://registry.test/service/token",service="harbor-registry"`,
		},
		{
			name: "Regular login request to '/v2' without 'Host', should return challenge whose realm is token URL with Ext endpoint",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/", nil)
				req.Host = ""
				return req
			}(),
			challenge: `Bearer realm="https://harbor.test/service/token",service="harbor-registry"`,
		},
		{
			name: "Request to 'v2' carrying basic auth header, the challenge should not have token service URI as realm b/c it's not from OCI client",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/", nil)
				req.SetBasicAuth("u", "p")
				return req
			}(),
			challenge: `Basic realm="harbor"`,
		},
		{
			name: "Request to '/v2/_catalog' should return the challenge should not have token service URI as realm",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/_catalog", nil)
				return req
			}(),
			challenge: `Basic realm="harbor"`,
		},
		{
			name: "Request to '/v2/_catalog' should return the challenge should not have token service URI as realm, disregarding the auth header in request",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/_catalog", nil)
				req.Header.Set("Authorization", "Bearer xx")
				return req
			}(),
			challenge: `Basic realm="harbor"`,
		},
		{
			name: "Request to mount a blob from one repo to another should return challenge with scope according to the artifact info in the context of the request",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "https://harbor.test/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_2/ubuntu", nil)
				req = req.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
					Repository:           "project_1/ubuntu",
					Reference:            "14.04",
					ProjectName:          "project_1",
					BlobMountRepository:  "project_2/ubuntu",
					BlobMountProjectName: "project_2",
					BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				}))
				return req
			}(),
			challenge: `Bearer realm="https://harbor.test/service/token",service="harbor-registry",scope="repository:project_1/ubuntu:pull,push repository:project_2/ubuntu:pull"`,
		},
		{
			name: "Request to be passed to registry, if it has basic auth header, it should return challenge without token URI as realm",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodPost, "https://harbor.test/v2/project_1/ubuntu/blobs/uploads/mount=?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=project_2/ubuntu", nil)
				req = req.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
					Repository:           "project_1/ubuntu",
					Reference:            "14.04",
					ProjectName:          "project_1",
					BlobMountRepository:  "project_2/ubuntu",
					BlobMountProjectName: "project_2",
					BlobMountDigest:      "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
				}))
				req.SetBasicAuth("user", "password")
				return req
			}(),
			challenge: `Basic realm="harbor"`,
		},
		{
			name: "Request to be passed to registry, if it is sent from internal, the token service URI in the realm of the challenge should also point to the internal URI",
			request: func() *http.Request {
				req, _ := http.NewRequest(http.MethodGet, "https://harbor.core:8443/v2/project_1/hello-world/manifests/v1", nil)
				req = req.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
					Repository:  "project_1/hello-world",
					Reference:   "v1",
					ProjectName: "project_1",
				}))
				return req
			}(),
			challenge: `Bearer realm="https://harbor.core:8443/service/token",service="harbor-registry",scope="repository:project_1/hello-world:pull"`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			acs := accessList(c.request)
			assert.Equal(t, c.challenge, getChallenge(c.request, acs))
		})
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		reqHost string
		rawURL  string
		expect  bool
	}{
		{
			"abc.com",
			"http://abc.com",
			true,
		},
		{
			"abc.com",
			"https://abc.com",
			true,
		},
		{
			"abc.com:80",
			"http://abc.com",
			true,
		},
		{
			"abc.com:80",
			"https://abc.com",
			false,
		},
		{
			"abc.com:443",
			"http://abc.com",
			false,
		},
		{
			"abc.com:443",
			"https://abc.com",
			true,
		},
		{
			"abcd.com:443",
			"https://abc.com",
			false,
		},
		{
			"abc.com:8443",
			"https://abc.com:8443",
			true,
		},
		{
			"abc.com",
			"https://abc.com:443",
			true,
		},
		{
			"abc.com",
			"http://abc.com:443",
			false,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, match(context.Background(), c.reqHost, c.rawURL))
	}
}
