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

package v1

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/base"
	"github.com/goharbor/harbor/src/replication/model"
)

type client struct {
	*base.Client
}

func (c *client) listRepositories(project *base.Project) ([]*model.Repository, error) {
	repositories := []*models.RepoRecord{}
	url := fmt.Sprintf("%s/repositories?project_id=%d", c.BasePath(), project.ID)
	if err := c.C.GetAndIteratePagination(url, &repositories); err != nil {
		return nil, err
	}
	var repos []*model.Repository
	for _, repository := range repositories {
		repos = append(repos, &model.Repository{
			Name:     repository.Name,
			Metadata: project.Metadata,
		})
	}
	return repos, nil
}

func (c *client) listArtifacts(repository string) ([]*model.Artifact, error) {
	url := fmt.Sprintf("%s/repositories/%s/tags", c.BasePath(), repository)
	tags := []*struct {
		Name   string `json:"name"`
		Labels []*struct {
			Name string `json:"name"`
		}
	}{}
	if err := c.C.Get(url, &tags); err != nil {
		return nil, err
	}
	var artifacts []*model.Artifact
	for _, tag := range tags {
		artifact := &model.Artifact{
			Type: string(model.ResourceTypeImage),
			Tags: []string{tag.Name},
		}
		for _, label := range tag.Labels {
			artifact.Labels = append(artifact.Labels, label.Name)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func (c *client) deleteManifest(repository, reference string) error {
	url := fmt.Sprintf("%s/repositories/%s/tags/%s", c.BasePath(), repository, reference)
	return c.C.Delete(url)
}
