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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/goharbor/harbor/src/testing/mock"
	repotesting "github.com/goharbor/harbor/src/testing/pkg/repository"
)

type catalogTestSuite struct {
	suite.Suite
	originalRepoMgr repository.Manager
	repoMgr         *repotesting.Manager
}

func (c *catalogTestSuite) SetupSuite() {
	c.originalRepoMgr = pkg.RepositoryMgr
}

func (c *catalogTestSuite) SetupTest() {
	c.repoMgr = &repotesting.Manager{}
	pkg.RepositoryMgr = c.repoMgr
}

func (c *catalogTestSuite) TearDownTest() {
}

func (c *catalogTestSuite) TearDownSuite() {
	pkg.RepositoryMgr = c.originalRepoMgr
}

func (c *catalogTestSuite) TestCatalog() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)

	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(2, len(ctlg.Repositories))
}

func (c *catalogTestSuite) TestCatalogPaginationN1() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=1", nil)
	var w *httptest.ResponseRecorder
	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(1, len(ctlg.Repositories))
	c.Equal("busybox", ctlg.Repositories[0])
}

func (c *catalogTestSuite) TestCatalogPaginationN2() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=3", nil)
	var w *httptest.ResponseRecorder
	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(2, len(ctlg.Repositories))
	c.Equal("hello-world", ctlg.Repositories[1])
}

func (c *catalogTestSuite) TestCatalogFiltersByPermission() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	project1RepositoryResource := rbac_project.NewNamespace(1).Resource(rbac.ResourceRepository)
	sc.On("IsSysAdmin").Return(false)
	mock.OnAnything(sc, "Can").Return(func(_ context.Context, action types.Action, resource types.Resource) bool {
		return resource == project1RepositoryResource && action == rbac.ActionPull
	})
	req = req.WithContext(security.NewContext(context.Background(), sc))

	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "project_1/hello-world",
			ProjectID:    1,
		},
		{
			RepositoryID: 2,
			Name:         "project_2/busybox",
			ProjectID:    2,
		},
	}, nil)

	w := httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)

	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal([]string{"project_1/hello-world"}, ctlg.Repositories)
}

func (c *catalogTestSuite) TestCatalogFiltersEmptyWithoutRepoPermission() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(false)
	mock.OnAnything(sc, "Can").Return(false)
	req = req.WithContext(security.NewContext(context.Background(), sc))

	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "project_1/hello-world",
			ProjectID:    1,
		},
	}, nil)

	w := httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)

	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Empty(ctlg.Repositories)
}

func (c *catalogTestSuite) TestCatalogReturnsEmptyForUnauthenticatedUser() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(false)
	req = req.WithContext(security.NewContext(context.Background(), sc))

	w := httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)

	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Empty(ctlg.Repositories)
	c.repoMgr.AssertNotCalled(c.T(), "NonEmptyRepos", mock.Anything)
}

func (c *catalogTestSuite) TestCatalogReturnsAllRepositoriesForSysAdmin() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	sc := &securitytesting.Context{}
	sc.On("IsAuthenticated").Return(true)
	sc.On("IsSysAdmin").Return(true)

	req = req.WithContext(security.NewContext(context.Background(), sc))

	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "project_1/hello-world",
			ProjectID:    1,
		},
		{
			RepositoryID: 2,
			Name:         "project_2/busybox",
			ProjectID:    2,
		},
	}, nil)

	w := httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)

	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Len(ctlg.Repositories, 2)
}

func (c *catalogTestSuite) TestCatalogPaginationN3() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?last=busybox&n=1", nil)
	var w *httptest.ResponseRecorder
	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{
		{
			RepositoryID: 1,
			Name:         "hello-world",
		},
		{
			RepositoryID: 2,
			Name:         "busybox",
		},
	}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(1, len(ctlg.Repositories))
	c.Equal("hello-world", ctlg.Repositories[0])
}

func (c *catalogTestSuite) TestCatalogEmptyRepo() {
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	mock.OnAnything(c.repoMgr, "NonEmptyRepos").Return([]*model.RepoRecord{}, nil)
	w = httptest.NewRecorder()
	newRepositoryHandler().ServeHTTP(w, req)
	c.Equal(http.StatusOK, w.Code)
	var ctlg struct {
		Repositories []string `json:"repositories"`
	}
	decoder := json.NewDecoder(w.Body)
	err := decoder.Decode(&ctlg)
	c.Nil(err)
	c.Equal(0, len(ctlg.Repositories))
}

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, &catalogTestSuite{})
}
