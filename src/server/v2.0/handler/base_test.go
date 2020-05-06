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
	"github.com/stretchr/testify/suite"
	"net/url"
	"testing"
)

type baseHandlerTestSuite struct {
	suite.Suite
	base *BaseAPI
}

func (b *baseHandlerTestSuite) SetupSuite() {
	b.base = &BaseAPI{}
}

func (b *baseHandlerTestSuite) TestBuildQuery() {
	// nil query string and pagination pointer
	var (
		query      *string
		pageNumber *int64
		pageSize   *int64
	)
	q, err := b.base.BuildQuery(nil, query, pageNumber, pageSize)
	b.Require().Nil(err)
	b.Require().NotNil(q)
	b.NotNil(q.Keywords)

	// not nil query string and pagination pointer
	var (
		qs       = "q=a=b"
		pn int64 = 1
		ps int64 = 10
	)
	q, err = b.base.BuildQuery(nil, &qs, &pn, &ps)
	b.Require().Nil(err)
	b.Require().NotNil(q)
	b.Equal(int64(1), q.PageNumber)
	b.Equal(int64(10), q.PageSize)
	b.NotNil(q.Keywords)

	var (
		qs1       = "q=a%3Db"
		pn1 int64 = 1
		ps1 int64 = 10
	)
	q, err = b.base.BuildQuery(nil, &qs1, &pn1, &ps1)
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

func TestBaseHandler(t *testing.T) {
	suite.Run(t, &baseHandlerTestSuite{})
}
