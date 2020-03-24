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
	modelsv2 "github.com/goharbor/harbor/src/controller/artifact"
)

func (c *client) ListAllArtifacts(project, repository string) ([]*modelsv2.Artifact, error) {
	url := c.buildURL(fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts", project, repository)) // should query only tag
	var arts []*modelsv2.Artifact
	if err := c.httpclient.GetAndIteratePagination(url, &arts); err != nil {
		return nil, err
	}
	return arts, nil
}

func (c *client) DeleteArtifact(project, repository, digest string) error {
	// /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}
	url := c.buildURL(fmt.Sprintf("/api/v2.0/projects/%s/repositories/%s/artifacts/%s", project, repository, digest))
	return c.httpclient.Delete(url)
}

func (c *client) DeleteArtifactRepository(project, repository string) error {
	url := c.buildURL(fmt.Sprintf("/api/repositories/%s/%s", project, repository))
	return c.httpclient.Delete(url)
}
