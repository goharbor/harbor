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

package repoproxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

func TestDefaultTagsListURL(t *testing.T) {
	cases := []struct {
		name     string
		project  string
		repo     string
		rawQuery string
		want     string
	}{
		{
			name:    "no query",
			project: "proxy",
			repo:    "busybox",
			want:    "/v2/proxy/library/busybox/tags/list",
		},
		{
			name:     "preserves pagination query",
			project:  "proxy",
			repo:     "alpine",
			rawQuery: "n=50&last=1.0",
			want:     "/v2/proxy/library/alpine/tags/list?n=50&last=1.0",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultTagsListURL(tt.project, tt.repo, tt.rawQuery); got != tt.want {
				t.Errorf("defaultTagsListURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

type TagsListMiddlewareTestSuite struct {
	suite.Suite

	originalProjectController project.Controller
	projectController         *projecttesting.Controller

	originalDefaultLibrary func(context.Context, int64, lib.ArtifactInfo) (bool, string, error)

	next http.Handler
}

func (suite *TagsListMiddlewareTestSuite) SetupTest() {
	suite.originalProjectController = project.Ctl
	suite.projectController = &projecttesting.Controller{}
	project.Ctl = suite.projectController

	suite.originalDefaultLibrary = defaultLibrary

	suite.next = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func (suite *TagsListMiddlewareTestSuite) TearDownTest() {
	project.Ctl = suite.originalProjectController
	defaultLibrary = suite.originalDefaultLibrary
}

func (suite *TagsListMiddlewareTestSuite) makeRequest(target, repository string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	info := lib.ArtifactInfo{
		ProjectName: "proxy",
		Repository:  repository,
	}
	return req.WithContext(lib.WithArtifactInfo(req.Context(), info))
}

// A single-segment Docker Hub official image must be redirected to the
// library/-prefixed tags/list path, preserving the pagination query string.
func (suite *TagsListMiddlewareTestSuite) TestRedirectsDockerHubLibraryImage() {
	mock.OnAnything(suite.projectController, "GetByName").Return(&proModels.Project{
		ProjectID:  1,
		Name:       "proxy",
		RegistryID: 1,
	}, nil)
	defaultLibrary = func(context.Context, int64, lib.ArtifactInfo) (bool, string, error) {
		return true, "busybox", nil
	}

	req := suite.makeRequest("/v2/proxy/busybox/tags/list?n=50&last=1.0", "proxy/busybox")
	rr := httptest.NewRecorder()

	TagsListMiddleware()(suite.next).ServeHTTP(rr, req)

	suite.Equal(http.StatusMovedPermanently, rr.Code)
	suite.Equal("/v2/proxy/library/busybox/tags/list?n=50&last=1.0", rr.Header().Get("Location"))
}

// A multi-segment repository is not a default-library image and must not be
// redirected; the request falls through to the next handler.
func (suite *TagsListMiddlewareTestSuite) TestNoRedirectForNamespacedImage() {
	mock.OnAnything(suite.projectController, "GetByName").Return(&proModels.Project{
		ProjectID:  1,
		Name:       "proxy",
		RegistryID: 0,
	}, nil)
	defaultLibrary = func(context.Context, int64, lib.ArtifactInfo) (bool, string, error) {
		return false, "", nil
	}

	req := suite.makeRequest("/v2/proxy/goharbor/harbor/tags/list", "proxy/goharbor/harbor")
	rr := httptest.NewRecorder()

	TagsListMiddleware()(suite.next).ServeHTTP(rr, req)

	suite.Equal(http.StatusOK, rr.Code)
	suite.Empty(rr.Header().Get("Location"))
}

func TestTagsListMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &TagsListMiddlewareTestSuite{})
}
