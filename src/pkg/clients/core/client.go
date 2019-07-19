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

package core

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/models"

	"github.com/goharbor/harbor/src/chartserver"
	chttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
)

// Client defines the methods that a core client should implement
// Currently, it contains only part of the whole method collection
// and we should expand it when needed
type Client interface {
	ImageClient
	ChartClient
}

// ImageClient defines the methods that an image client should implement
type ImageClient interface {
	ListAllImages(project, repository string) ([]*models.TagResp, error)
	DeleteImage(project, repository, tag string) error
}

// ChartClient defines the methods that a chart client should implement
type ChartClient interface {
	ListAllCharts(project, repository string) ([]*chartserver.ChartVersion, error)
	DeleteChart(project, repository, version string) error
}

// New returns an instance of the client which is a default implement for Client
func New(url string, httpclient *http.Client, authorizer modifier.Modifier) Client {
	return &client{
		url:        url,
		httpclient: chttp.NewClient(httpclient, authorizer),
	}
}

type client struct {
	url        string
	httpclient *chttp.Client
}

func (c *client) buildURL(path string) string {
	return fmt.Sprintf("%s/%s", c.url, path)
}
