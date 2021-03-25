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

package jfrog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg/reg/model"

	"github.com/stretchr/testify/suite"
)

type clientTestSuite struct {
	suite.Suite

	client     *client
	mockServer *httptest.Server
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, &clientTestSuite{})
}

func (c *clientTestSuite) SetupSuite() {
	c.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/artifactory/api/repositories?packageType=docker":
			if r.Method == http.MethodGet {
				w.Write([]byte(`[
    {
        "key": "repo1",
        "description": "",
        "type": "LOCAL",
        "url": "http://49.4.2.82:8081/artifactory/repo1",
        "packageType": "Docker"
    },
    {
        "key": "mydocker",
        "type": "LOCAL",
        "url": "http://49.4.2.82:8081/artifactory/mydocker",
        "packageType": "Docker"
    }
]`))
				return
			}
			w.WriteHeader(http.StatusNotImplemented)
		case "/artifactory/api/repositories/test":
			if r.Method == http.MethodPut {
				w.WriteHeader(200)
				return
			}
			w.WriteHeader(http.StatusNotImplemented)
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))

	c.client = newClient(&model.Registry{URL: c.mockServer.URL})
}

func (c *clientTestSuite) TearDownSuite() {
	c.mockServer.Close()
}

func (c *clientTestSuite) TestGetDockerRepositories() {
	repos, err := c.client.getDockerRepositories()
	c.NoError(err)
	c.Len(repos, 2)
	c.Equal("repo1", repos[0].Key)
}

func (c *clientTestSuite) TestCreateDockerRepository() {
	err := c.client.createDockerRepository("test")
	c.NoError(err)
}
