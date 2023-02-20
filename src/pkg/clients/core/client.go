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

	chttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	modelsv2 "github.com/goharbor/harbor/src/controller/artifact"
)

// Client defines the methods that a core client should implement
// Currently, it contains only part of the whole method collection
// and we should expand it when needed
type Client interface {
	ArtifactClient
}

// ArtifactClient defines the methods that an image client should implement
type ArtifactClient interface {
	ListAllArtifacts(project, repository string) ([]*modelsv2.Artifact, error)
	DeleteArtifact(project, repository, digest string) error
	DeleteArtifactRepository(project, repository string) error
}

// New returns an instance of the client which is a default implement for Client
func New(url string, httpclient *http.Client, modifiers ...modifier.Modifier) Client {
	return &client{
		url:        url,
		httpclient: chttp.NewClient(httpclient, modifiers...),
	}
}

type client struct {
	url        string
	httpclient *chttp.Client
}

func (c *client) buildURL(path string) string {
	return fmt.Sprintf("%s%s", c.url, path)
}
