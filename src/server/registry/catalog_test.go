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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/testing/mock"
	repotesting "github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
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
