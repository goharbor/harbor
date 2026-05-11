//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	securitySecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/controller/project"
	registryCtl "github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	libCache "github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	testingcache "github.com/goharbor/harbor/src/testing/lib/cache"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	proxytesting "github.com/goharbor/harbor/src/testing/controller/proxy"
	testingmock "github.com/goharbor/harbor/src/testing/mock"
)

func TestIsProxySession(t *testing.T) {
	sc1 := securitySecret.NewSecurityContext("123456789", nil)
	otherCtx := security.NewContext(context.Background(), sc1)

	sc2 := proxycachesecret.NewSecurityContext("library/hello-world")
	proxyCtx := security.NewContext(context.Background(), sc2)

	user := &models.User{
		Username: "robot$library+scanner-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc := local.NewSecurityContext(user)
	scannerCtx := security.NewContext(context.Background(), userSc)

	otherRobot := &models.User{
		Username: "robot$library+test-8ec3b47a-fd29-11ee-9681-0242c0a87009",
	}
	userSc2 := local.NewSecurityContext(otherRobot)
	nonScannerCtx := security.NewContext(context.Background(), userSc2)

	cases := []struct {
		name string
		in   context.Context
		want bool
	}{
		{
			name: `normal`,
			in:   otherCtx,
			want: false,
		},
		{
			name: `proxy user`,
			in:   proxyCtx,
			want: true,
		},
		{
			name: `robot account`,
			in:   scannerCtx,
			want: true,
		},
		{
			name: `non scanner robot`,
			in:   nonScannerCtx,
			want: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := isProxySession(tt.in, "library")
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

// mockRegistryController is a minimal testify mock for registry.Controller.
// Only Get is exercised by proxyReferrerGet; remaining methods panic if called.
type mockRegistryController struct {
	mock.Mock
}

func (m *mockRegistryController) Get(ctx context.Context, id int64) (*model.Registry, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Registry), args.Error(1)
}

func (m *mockRegistryController) Create(_ context.Context, _ *model.Registry) (int64, error) {
	panic("not implemented")
}
func (m *mockRegistryController) Count(_ context.Context, _ *q.Query) (int64, error) {
	panic("not implemented")
}
func (m *mockRegistryController) List(_ context.Context, _ *q.Query) ([]*model.Registry, error) {
	panic("not implemented")
}
func (m *mockRegistryController) Update(_ context.Context, _ *model.Registry, _ ...string) error {
	panic("not implemented")
}
func (m *mockRegistryController) Delete(_ context.Context, _ int64) error {
	panic("not implemented")
}
func (m *mockRegistryController) GetInfo(_ context.Context, _ int64) (*model.RegistryInfo, error) {
	panic("not implemented")
}
func (m *mockRegistryController) IsHealthy(_ context.Context, _ *model.Registry) (bool, error) {
	panic("not implemented")
}
func (m *mockRegistryController) ListRegistryProviderTypes(_ context.Context) ([]string, error) {
	panic("not implemented")
}
func (m *mockRegistryController) ListRegistryProviderInfos(_ context.Context) (map[string]*model.AdapterPattern, error) {
	panic("not implemented")
}
func (m *mockRegistryController) StartRegularHealthCheck(_ context.Context, _, _ chan struct{}) {
	panic("not implemented")
}

// --- ProxyReferrerMiddleware suite ---

type ProxyReferrerMiddlewareSuite struct {
	suite.Suite

	origProjectCtl project.Controller
	projectCtl     *projecttesting.Controller

	next        http.Handler
	nextCalled  bool
	artInfo     lib.ArtifactInfo
}

func (s *ProxyReferrerMiddlewareSuite) SetupTest() {
	s.origProjectCtl = project.Ctl
	s.projectCtl = &projecttesting.Controller{}
	project.Ctl = s.projectCtl

	s.nextCalled = false
	s.next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	s.artInfo = lib.ArtifactInfo{
		ProjectName: "testproject",
		Repository:  "testproject/image",
		Reference:   "sha256:aaaa",
		Digest:      "sha256:aaaa",
	}
}

func (s *ProxyReferrerMiddlewareSuite) TearDownTest() {
	project.Ctl = s.origProjectCtl
}

func (s *ProxyReferrerMiddlewareSuite) makeRequest() *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/v2/testproject/image/referrers/sha256:aaaa", nil)
	ctx := lib.WithArtifactInfo(req.Context(), s.artInfo)
	return req.WithContext(ctx)
}

func (s *ProxyReferrerMiddlewareSuite) TestNonProxyProject() {
	proj := &proModels.Project{ProjectID: 1, Name: "testproject", RegistryID: 0}
	testingmock.OnAnything(s.projectCtl, "GetByName").Return(proj, nil)

	w := httptest.NewRecorder()
	ProxyReferrerMiddleware()(s.next).ServeHTTP(w, s.makeRequest())

	s.True(s.nextCalled, "next handler should be called for non-proxy project")
}

func (s *ProxyReferrerMiddlewareSuite) TestProjectNotFound() {
	testingmock.OnAnything(s.projectCtl, "GetByName").Return(nil, errors.NotFoundError(nil))

	w := httptest.NewRecorder()
	ProxyReferrerMiddleware()(s.next).ServeHTTP(w, s.makeRequest())

	s.False(s.nextCalled, "next handler should not be called when project lookup fails")
	s.NotEqual(http.StatusOK, w.Code)
}

func (s *ProxyReferrerMiddlewareSuite) TestProxyReferrerAPIDisabled() {
	proj := &proModels.Project{
		ProjectID:  2,
		Name:       "testproject",
		RegistryID: 1,
		Metadata:   map[string]string{},
	}
	testingmock.OnAnything(s.projectCtl, "GetByName").Return(proj, nil)

	w := httptest.NewRecorder()
	ProxyReferrerMiddleware()(s.next).ServeHTTP(w, s.makeRequest())

	s.True(s.nextCalled, "next handler should be called when proxy referrer API is disabled")
}

func TestProxyReferrerMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(ProxyReferrerMiddlewareSuite))
}

// --- proxyReferrerGet suite ---

type ProxyReferrerGetSuite struct {
	suite.Suite

	origRegistryCtl registryCtl.Controller
	registryCtl     *mockRegistryController

	mockCache  *testingcache.Cache
	mockRemote *proxytesting.RemoteInterface

	art lib.ArtifactInfo
}

func (s *ProxyReferrerGetSuite) SetupSuite() {
	// Initialize memory cache as the global default for this suite.
	s.Require().NoError(libCache.Initialize(libCache.Memory, ""))
}

func (s *ProxyReferrerGetSuite) SetupTest() {
	s.origRegistryCtl = registryCtl.Ctl
	s.registryCtl = &mockRegistryController{}
	registryCtl.Ctl = s.registryCtl

	s.mockRemote = &proxytesting.RemoteInterface{}
	s.art = lib.ArtifactInfo{
		ProjectName: "proxy-cache",
		Repository:  "proxy-cache/app",
		Reference:   "sha256:bbbb",
		Digest:      "sha256:bbbb",
	}
}

func (s *ProxyReferrerGetSuite) TearDownTest() {
	registryCtl.Ctl = s.origRegistryCtl
}

func (s *ProxyReferrerGetSuite) makeRequest(rawQuery string) *http.Request {
	url := "/v2/proxy-cache/app/referrers/sha256:bbbb"
	if rawQuery != "" {
		url += "?" + rawQuery
	}
	req := httptest.NewRequest(http.MethodGet, url, nil)
	return req.WithContext(lib.WithArtifactInfo(req.Context(), s.art))
}

func (s *ProxyReferrerGetSuite) TestUpstreamNotFound() {
	healthyReg := &model.Registry{ID: 1, Status: model.Healthy}
	testingmock.OnAnything(s.registryCtl, "Get").Return(healthyReg, nil)
	testingmock.OnAnything(s.mockRemote, "ListReferrers").Return(nil, nil, errors.NotFoundError(nil))

	w := httptest.NewRecorder()
	err := proxyReferrerGet(s.makeRequest(""), w, s.art, s.mockRemote, 1)

	s.NoError(err)
	s.Equal(http.StatusNotFound, w.Code)
}

func (s *ProxyReferrerGetSuite) TestUpstreamSuccessWritesResponse() {
	healthyReg := &model.Registry{ID: 1, Status: model.Healthy}
	testingmock.OnAnything(s.registryCtl, "Get").Return(healthyReg, nil)

	upstreamIndex := &ocispec.Index{
		Manifests: []ocispec.Descriptor{
			{MediaType: "application/vnd.oci.image.manifest.v1+json", ArtifactType: "application/vnd.example.sbom"},
		},
	}
	headers := map[string][]string{
		"Content-Type":  {"application/vnd.oci.image.index.v1+json"},
		"X-Total-Count": {"1"},
	}
	testingmock.OnAnything(s.mockRemote, "ListReferrers").Return(upstreamIndex, headers, nil)

	w := httptest.NewRecorder()
	err := proxyReferrerGet(s.makeRequest(""), w, s.art, s.mockRemote, 1)

	s.NoError(err)
	s.Equal(http.StatusOK, w.Code)
	s.Equal("application/vnd.oci.image.index.v1+json", w.Header().Get("Content-Type"))

	var gotIndex ocispec.Index
	s.NoError(json.Unmarshal(w.Body.Bytes(), &gotIndex))
	s.Len(gotIndex.Manifests, 1)
}

func (s *ProxyReferrerGetSuite) TestUnhealthyRegistryCacheHit() {
	unhealthyReg := &model.Registry{ID: 1, Status: "unhealthy"}
	testingmock.OnAnything(s.registryCtl, "Get").Return(unhealthyReg, nil)

	// Prime the cache manually.
	cachedIndex := &ocispec.Index{
		Manifests: []ocispec.Descriptor{
			{MediaType: "application/vnd.oci.image.manifest.v1+json", ArtifactType: "application/vnd.sbom"},
		},
	}
	b, err := json.Marshal(cachedIndex)
	s.Require().NoError(err)
	cacheKey := referrerCacheKey("/v2/proxy-cache/app/referrers/sha256:bbbb")
	cached := referrerCache{
		Content: b,
		Header:  map[string][]string{"Content-Type": {"application/vnd.oci.image.index.v1+json"}},
	}
	s.Require().NoError(libCache.Default().Save(context.Background(), cacheKey, cached))

	w := httptest.NewRecorder()
	err = proxyReferrerGet(s.makeRequest(""), w, s.art, s.mockRemote, 1)

	s.NoError(err)
	s.Equal(http.StatusOK, w.Code)

	var gotIndex ocispec.Index
	s.NoError(json.Unmarshal(w.Body.Bytes(), &gotIndex))
	s.Len(gotIndex.Manifests, 1)
}

func (s *ProxyReferrerGetSuite) TestUnhealthyRegistryCacheMiss() {
	unhealthyReg := &model.Registry{ID: 1, Status: "unhealthy"}
	testingmock.OnAnything(s.registryCtl, "Get").Return(unhealthyReg, nil)

	// Use a unique URI so the cache lookup finds nothing.
	req := httptest.NewRequest(http.MethodGet, "/v2/proxy-cache/app/referrers/sha256:miss", nil)
	req = req.WithContext(lib.WithArtifactInfo(req.Context(), s.art))

	w := httptest.NewRecorder()
	err := proxyReferrerGet(req, w, s.art, s.mockRemote, 1)

	s.Error(err, "should return error when registry is unhealthy and cache is empty")
}

func TestProxyReferrerGetSuite(t *testing.T) {
	suite.Run(t, new(ProxyReferrerGetSuite))
}

// TestReferrerCacheKey tests the cache key generation
func TestReferrerCacheKey(t *testing.T) {
	cases := []struct {
		name        string
		requestURI  string
		expectedKey string
	}{
		{
			name:        "simple path",
			requestURI:  "/v2/repo/image/referrers",
			expectedKey: "{referrer_cache}:/v2/repo/image/referrers",
		},
		{
			name:        "path with query",
			requestURI:  "/v2/repo/image/referrers?filter=sbom&format=json",
			expectedKey: "{referrer_cache}:/v2/repo/image/referrers?filter=sbom&format=json",
		},
		{
			name:        "path with special chars",
			requestURI:  "/v2/my-repo/my-image/referrers?annotation=key%3Dvalue",
			expectedKey: "{referrer_cache}:/v2/my-repo/my-image/referrers?annotation=key%3Dvalue",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			key := referrerCacheKey(tt.requestURI)
			require.Equal(t, tt.expectedKey, key)
		})
	}
}

// TestWriteProxyHeaders tests that only allowed headers are written
func TestWriteProxyHeaders(t *testing.T) {
	cases := []struct {
		name            string
		headerMap       map[string][]string
		expectedHeaders map[string]string
		shouldNotHave   []string
	}{
		{
			name: "allowed headers",
			headerMap: map[string][]string{
				"Content-Type":  {"application/json"},
				"Link":          {"<http://example.com>; rel=\"next\""},
				"X-Total-Count": {"100"},
			},
			expectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Link":          "<http://example.com>; rel=\"next\"",
				"X-Total-Count": "100",
			},
			shouldNotHave: []string{},
		},
		{
			name: "mixed allowed and disallowed headers",
			headerMap: map[string][]string{
				"Content-Type":    {"application/json"},
				"X-Custom-Header": {"should-not-appear"},
				"X-Total-Count":   {"50"},
				"Authorization":   {"Bearer token"},
			},
			expectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"X-Total-Count": "50",
			},
			shouldNotHave: []string{"X-Custom-Header", "Authorization"},
		},
		{
			name: "multiple values for same header",
			headerMap: map[string][]string{
				"Link": {
					"<http://example.com/page=1>; rel=\"first\"",
					"<http://example.com/page=2>; rel=\"next\"",
				},
			},
			expectedHeaders: map[string]string{
				"Link": "<http://example.com/page=1>; rel=\"first\"",
			},
			shouldNotHave: []string{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteProxyHeaders(w, tt.headerMap)

			// Check expected headers are present
			for header, _ := range tt.expectedHeaders {
				actualValue := w.Header().Get(header)
				require.NotEmpty(t, actualValue, fmt.Sprintf("header %s should be present", header))
			}

			// Check disallowed headers are not present
			for _, header := range tt.shouldNotHave {
				actualValue := w.Header().Get(header)
				require.Empty(t, actualValue, fmt.Sprintf("header %s should not be present", header))
			}
		})
	}
}
