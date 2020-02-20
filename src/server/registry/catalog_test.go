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
	"github.com/goharbor/harbor/src/api/repository"
	"github.com/goharbor/harbor/src/common/models"
	repotesting "github.com/goharbor/harbor/src/testing/api/repository"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type catalogTestSuite struct {
	suite.Suite
	originalRepoCtl repository.Controller
	repoCtl         *repotesting.FakeController
}

func (c *catalogTestSuite) SetupSuite() {
	c.originalRepoCtl = repository.Ctl
}

func (c *catalogTestSuite) SetupTest() {
	c.repoCtl = &repotesting.FakeController{}
	repository.Ctl = c.repoCtl
}

func (c *catalogTestSuite) TearDownTest() {
}

func (c *catalogTestSuite) TearDownSuite() {
	repository.Ctl = c.originalRepoCtl
}

func (c *catalogTestSuite) TestCatalog() {
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
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
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=1", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
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
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?n=3", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
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
	c.SetupTest()
	req := httptest.NewRequest(http.MethodGet, "/v2/_catalog?last=busybox&n=1", nil)
	var w *httptest.ResponseRecorder
	c.repoCtl.On("List").Return([]*models.RepoRecord{
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

func TestCatalogTestSuite(t *testing.T) {
	suite.Run(t, &catalogTestSuite{})
}
