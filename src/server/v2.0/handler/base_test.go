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

package handler

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/project"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
)

var (
	projectCtlMock *projecttesting.Controller
)

type baseHandlerTestSuite struct {
	suite.Suite
	base *BaseAPI
}

func (b *baseHandlerTestSuite) SetupSuite() {
	b.base = &BaseAPI{}
}

func (b *baseHandlerTestSuite) TestBuildQuery() {
	// nil input
	var (
		query      *string
		sort       *string
		pageNumber *int64
		pageSize   *int64
	)
	q, err := b.base.BuildQuery(nil, query, sort, pageNumber, pageSize)
	b.Require().Nil(err)
	b.Require().NotNil(q)
	b.NotNil(q.Keywords)

	// not nil input
	var (
		qs       = "q=a=b"
		st       = "a,-c"
		pn int64 = 1
		ps int64 = 10
	)
	q, err = b.base.BuildQuery(nil, &qs, &st, &pn, &ps)
	b.Require().Nil(err)
	b.Require().NotNil(q)
	b.Equal(int64(1), q.PageNumber)
	b.Equal(int64(10), q.PageSize)
	b.NotNil(q.Keywords)
	b.Require().Len(q.Sorts, 2)
	b.Equal("a", q.Sorts[0].Key)
	b.False(q.Sorts[0].DESC)
	b.Equal("c", q.Sorts[1].Key)
	b.True(q.Sorts[1].DESC)

	var (
		qs1       = "q=a%3Db"
		st1       = ""
		pn1 int64 = 1
		ps1 int64 = 10
	)
	q, err = b.base.BuildQuery(nil, &qs1, &st1, &pn1, &ps1)
	b.Require().Nil(err)
	b.Require().NotNil(q)
	b.Equal(int64(1), q.PageNumber)
	b.Equal(int64(10), q.PageSize)
	b.Equal(q.Keywords["q"], "a=b")
}

func (b *baseHandlerTestSuite) TestLinks() {
	// request first page, response contains only "next" link
	url, err := url.Parse("http://localhost/api/artifacts?page=1&page_size=1")
	b.Require().Nil(err)
	links := b.base.Links(nil, url, 3, 1, 1)
	b.Require().Len(links, 1)
	b.Equal("next", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=2&page_size=1", links[0].URL)

	// request last page, response contains only "prev" link
	url, err = url.Parse("http://localhost/api/artifacts?page=3&page_size=1")
	b.Require().Nil(err)
	links = b.base.Links(nil, url, 3, 3, 1)
	b.Require().Len(links, 1)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=2&page_size=1", links[0].URL)

	// request the second page, response contains both "prev" and "next" links
	url, err = url.Parse("http://localhost/api/artifacts?page=2&page_size=1")
	b.Require().Nil(err)
	links = b.base.Links(nil, url, 3, 2, 1)
	b.Require().Len(links, 2)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=1&page_size=1", links[0].URL)
	b.Equal("next", links[1].Rel)
	b.Equal("http://localhost/api/artifacts?page=3&page_size=1", links[1].URL)

	// path and query contain escaped characters
	url, err = url.Parse("http://localhost/api/library%252Fhello-world/artifacts?page=2&page_size=1&q=a%3D~b")
	b.Require().Nil(err)
	links = b.base.Links(nil, url, 3, 2, 1)
	b.Require().Len(links, 2)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/library%252Fhello-world/artifacts?page=1&page_size=1&q=a=~b", links[0].URL)
	b.Equal("next", links[1].Rel)
	b.Equal("http://localhost/api/library%252Fhello-world/artifacts?page=3&page_size=1&q=a=~b", links[1].URL)
}

func (b *baseHandlerTestSuite) TestPaginationLinks() {
	// page 1 with hasNext = true (only next)
	u, err := url.Parse("http://localhost/api/artifacts?page=1&page_size=10")
	b.Require().Nil(err)
	links := b.base.PaginationLinks(nil, u, 1, 10, true)
	b.Require().Len(links, 1)
	b.Equal("next", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=2&page_size=10", links[0].URL)

	// page 2 with hasNext = true (both prev and next)
	links = b.base.PaginationLinks(nil, u, 2, 10, true)
	b.Require().Len(links, 2)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=1&page_size=10", links[0].URL)
	b.Equal("next", links[1].Rel)
	b.Equal("http://localhost/api/artifacts?page=3&page_size=10", links[1].URL)

	// page 3 with hasNext = false (last page, only prev)
	links = b.base.PaginationLinks(nil, u, 3, 10, false)
	b.Require().Len(links, 1)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=2&page_size=10", links[0].URL)

	// page 5 with hasNext = false (out-of-bounds page, still has prev to navigate back)
	links = b.base.PaginationLinks(nil, u, 5, 10, false)
	b.Require().Len(links, 1)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=4&page_size=10", links[0].URL)

	// page 1 with hasNext = false (single page fit, no links)
	links = b.base.PaginationLinks(nil, u, 1, 10, false)
	b.Empty(links)

	// pageSize = 0 (no links)
	links = b.base.PaginationLinks(nil, u, 1, 0, true)
	b.Empty(links)

	// query parameters preservation (e.g. q and sort)
	uWithParams, err := url.Parse("http://localhost/api/artifacts?page=2&page_size=10&q=project_id%3D3&sort=-op_time")
	b.Require().Nil(err)
	links = b.base.PaginationLinks(nil, uWithParams, 2, 10, true)
	b.Require().Len(links, 2)
	b.Equal("prev", links[0].Rel)
	b.Equal("http://localhost/api/artifacts?page=1&page_size=10&q=project_id=3&sort=-op_time", links[0].URL)
	b.Equal("next", links[1].Rel)
	b.Equal("http://localhost/api/artifacts?page=3&page_size=10&q=project_id=3&sort=-op_time", links[1].URL)
}

func TestBaseHandler(t *testing.T) {
	suite.Run(t, &baseHandlerTestSuite{})
}

func TestMain(m *testing.M) {
	projectCtlMock = &projecttesting.Controller{}

	baseProjectCtl = projectCtlMock

	exitVal := m.Run()

	baseProjectCtl = project.Ctl

	os.Exit(exitVal)
}
