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

package replication

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
)

type repository struct {
	name string
	tags []string
}

// registry wraps operations of Harbor UI and docker registry into one struct
type registry struct {
	reg.Repository                     // docker registry client
	client         *common_http.Client // Harbor client
	url            string
	insecure       bool
}

func (r *registry) GetProject(name string) (*models.Project, error) {
	url, err := url.Parse(strings.TrimRight(r.url, "/") + "/api/projects")
	if err != nil {
		return nil, err
	}
	q := url.Query()
	q.Set("name", name)
	url.RawQuery = q.Encode()

	projects := []*models.Project{}
	if err = r.client.Get(url.String(), &projects); err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.Name == name {
			return project, nil
		}
	}

	return nil, fmt.Errorf("project %s not found", name)
}

func (r *registry) CreateProject(project *models.Project) error {
	// only replicate the public property of project
	pro := struct {
		models.ProjectRequest
		Public int `json:"public"`
	}{
		ProjectRequest: models.ProjectRequest{
			Name: project.Name,
			Metadata: map[string]string{
				models.ProMetaPublic: strconv.FormatBool(project.IsPublic()),
			},
		},
	}

	// put "public" property in both metadata and public field to keep compatibility
	// with old version API(<=1.2.0)
	if project.IsPublic() {
		pro.Public = 1
	}

	return r.client.Post(strings.TrimRight(r.url, "/")+"/api/projects/", pro)
}

func (r *registry) DeleteRepository(repository string) error {
	return r.client.Delete(strings.TrimRight(r.url, "/") + "/api/repositories/" + repository)
}

func (r *registry) DeleteImage(repository, tag string) error {
	return r.client.Delete(strings.TrimRight(r.url, "/") + "/api/repositories/" + repository + "/tags/" + tag)
}
