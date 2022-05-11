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

package registry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	beegocontext "github.com/beego/beego/context"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/server/router"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/errors"
	arttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	repotesting "github.com/goharbor/harbor/src/testing/controller/repository"
	"github.com/goharbor/harbor/src/testing/mock"
	testmanifest "github.com/goharbor/harbor/src/testing/pkg/cached/manifest/redis"
	"github.com/stretchr/testify/suite"
)

type manifestTestSuite struct {
	suite.Suite
	originalRepoCtl repository.Controller
	originalArtCtl  artifact.Controller
	originalProxy   http.Handler
	repoCtl         *repotesting.Controller
	artCtl          *arttesting.Controller
	cachedMgr       *testmanifest.CachedManager
}

func (m *manifestTestSuite) SetupSuite() {
	m.originalRepoCtl = repository.Ctl
	m.originalArtCtl = artifact.Ctl
	m.originalProxy = proxy
	m.cachedMgr = &testmanifest.CachedManager{}
}

func (m *manifestTestSuite) SetupTest() {
	m.repoCtl = &repotesting.Controller{}
	m.artCtl = &arttesting.Controller{}
	repository.Ctl = m.repoCtl
	artifact.Ctl = m.artCtl
	pkg.ManifestMgr = m.cachedMgr
}

func (m *manifestTestSuite) TearDownTest() {
	proxy = nil
}

func (m *manifestTestSuite) TearDownSuite() {
	repository.Ctl = m.originalRepoCtl
	artifact.Ctl = m.originalArtCtl
	proxy = m.originalProxy
}

func (m *manifestTestSuite) TestGetManifest() {
	// doesn't exist
	req := httptest.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/latest", nil)
	w := &httptest.ResponseRecorder{}

	mock.OnAnything(m.artCtl, "GetByReference").Return(nil, errors.New(nil).WithCode(errors.NotFoundCode))
	getManifest(w, req)
	m.Equal(http.StatusNotFound, w.Code)

	// reset the mock
	m.SetupTest()

	// exist
	proxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet && req.URL.Path == "/v2/library/hello-world/manifests/sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})

	// as we cannot set the beego input in the context, here the request doesn't carry reference part
	req = httptest.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/", nil)
	w = &httptest.ResponseRecorder{}

	art := &artifact.Artifact{}
	art.Digest = "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"
	mock.OnAnything(m.artCtl, "GetByReference").Return(art, nil)
	getManifest(w, req)
	m.Equal(http.StatusOK, w.Code)

	// if etag match, return 304
	req = httptest.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/", nil)
	w = &httptest.ResponseRecorder{}
	req.Header.Set("If-None-Match", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
	getManifest(w, req)
	m.Equal(http.StatusNotModified, w.Code)

	// should get from cache if enable cache.
	config.DefaultMgr().Set(req.Context(), "cache_enabled", true)
	defer config.DefaultMgr().Set(req.Context(), "cache_enabled", false)
	req = httptest.NewRequest(http.MethodGet, "/v2/library/hello-world/manifests/", nil)
	w = &httptest.ResponseRecorder{}
	mock.OnAnything(m.cachedMgr, "Get").Return([]byte{}, nil)
	getManifest(w, req)
	m.Equal(http.StatusOK, w.Code)
	m.Equal("sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", w.Header().Get("Docker-Content-Digest"))
	m.cachedMgr.AssertCalled(m.T(), "Get", mock.Anything, mock.Anything)
}

func (m *manifestTestSuite) TestDeleteManifest() {
	// doesn't exist
	req := httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil)
	w := &httptest.ResponseRecorder{}

	mock.OnAnything(m.artCtl, "GetByReference").Return(nil, errors.New(nil).WithCode(errors.NotFoundCode))
	deleteManifest(w, req)
	m.Equal(http.StatusBadRequest, w.Code)

	// reset the mock
	m.SetupTest()

	// reset the mock
	m.SetupTest()

	// exist
	proxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut && req.URL.Path == "/v2/library/hello-world/manifests/latest" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Docker-Content-Digest", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	req = httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", nil)
	input := &beegocontext.BeegoInput{}
	input.SetParam(":reference", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))
	w = &httptest.ResponseRecorder{}
	mock.OnAnything(m.artCtl, "GetByReference").Return(&artifact.Artifact{}, nil)
	mock.OnAnything(m.artCtl, "Delete").Return(nil)
	deleteManifest(w, req)
	m.Equal(http.StatusAccepted, w.Code)

	// should get from cache if enable cache.
	config.DefaultMgr().Set(req.Context(), "cache_enabled", true)
	defer config.DefaultMgr().Set(req.Context(), "cache_enabled", false)
	// should delete cache when manifest be deleted.
	req = httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", nil)
	input = &beegocontext.BeegoInput{}
	input.SetParam(":reference", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))
	w = &httptest.ResponseRecorder{}
	mock.OnAnything(m.cachedMgr, "Delete").Return(nil)
	deleteManifest(w, req)
	m.Equal(http.StatusAccepted, w.Code)
	m.cachedMgr.AssertCalled(m.T(), "Delete", mock.Anything, mock.Anything)
}

func (m *manifestTestSuite) TestPutManifest() {
	// the backend registry response with 500
	proxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut && req.URL.Path == "/v2/library/hello-world/manifests/latest" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	req := httptest.NewRequest(http.MethodPut, "/v2/library/hello-world/manifests/latest", nil)
	w := &httptest.ResponseRecorder{}
	mock.OnAnything(m.repoCtl, "Ensure").Return(false, int64(1), nil)
	putManifest(w, req)
	m.Equal(http.StatusInternalServerError, w.Code)

	// reset the mock
	m.SetupTest()

	// // the backend registry serves the request successfully
	proxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPut && req.URL.Path == "/v2/library/hello-world/manifests/latest" {
			w.Header().Set("Docker-Content-Digest", "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180")
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	req = httptest.NewRequest(http.MethodPut, "/v2/library/hello-world/manifests/latest", nil)
	w = &httptest.ResponseRecorder{}
	mock.OnAnything(m.repoCtl, "Ensure").Return(false, int64(1), nil)
	mock.OnAnything(m.artCtl, "Ensure").Return(true, int64(1), nil)
	putManifest(w, req)
	m.Equal(http.StatusCreated, w.Code)
}

func TestManifestTestSuite(t *testing.T) {
	suite.Run(t, &manifestTestSuite{})
}
