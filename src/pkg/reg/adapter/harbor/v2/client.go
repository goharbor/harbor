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

package v2

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	ctltag "github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/encode/repository"
	labelmodel "github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/harbor/base"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

type client struct {
	*base.Client
	pageSize int64
}

func (c *client) listRepositories(project *base.Project) ([]*model.Repository, error) {
	repositories := []*repomodel.RepoRecord{}
	url := fmt.Sprintf("%s/projects/%s/repositories", c.BasePath(), project.Name)
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

func (c *client) listArtifacts(repo string) ([]*model.Artifact, error) {
	project, repo := utils.ParseRepository(repo)
	repo = repository.Encode(repo)
	// set the default value to equal the value specified when the UI submits the request
	if c.pageSize == 0 {
		c.pageSize = 15
	}
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts?page_size=%d&with_label=true&with_accessory=true",
		c.BasePath(), project, repo, c.pageSize)
	artifacts := []*artifact.Artifact{}
	if err := c.C.GetAndIteratePagination(url, &artifacts); err != nil {
		return nil, err
	}
	var arts []*model.Artifact

	for _, artItem := range artifacts {
		art := &model.Artifact{
			Type:   artItem.Type,
			Digest: artItem.Digest,
		}
		for _, label := range artItem.Labels {
			art.Labels = append(art.Labels, label.Name)
		}
		for _, tag := range artItem.Tags {
			art.Tags = append(art.Tags, tag.Name)
		}
		arts = append(arts, art)

		// append the accessory of index or individual artifact
		accArts := make([]*model.Artifact, 0)
		if err := c.getAccessoryArts(project, repo, artItem, artItem.Labels, artItem.Tags, &accArts); err != nil {
			return nil, err
		}
		arts = append(arts, accArts...)

		// append the accessory of reference if it has
		for _, ref := range artItem.References {
			url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s?with_accessory=true",
				c.BasePath(), project, repo, ref.ChildDigest)
			artRef := artifact.Artifact{}
			if err := c.C.Get(url, &artRef); err != nil {
				return nil, err
			}
			accArts := make([]*model.Artifact, 0)
			if err := c.getAccessoryArts(project, repo, &artRef, artItem.Labels, artItem.Tags, &accArts); err != nil {
				return nil, err
			}
			arts = append(arts, accArts...)
		}
	}
	return arts, nil
}

func (c *client) getAccessoryArts(project, repo string, art *artifact.Artifact, labels []*labelmodel.Label, tags []*ctltag.Tag, accArts *[]*model.Artifact) error {
	for _, acc := range art.Accessories {
		accArt := &model.Artifact{
			Type:   art.Type,
			Digest: acc.GetData().Digest,
			IsAcc:  true,
		}
		for _, tag := range tags {
			accArt.ParentTags = append(accArt.ParentTags, tag.Name)
		}
		// set the labels belonging to the subject manifest to the accessories.
		for _, label := range labels {
			accArt.Labels = append(accArt.Labels, label.Name)
		}
		// recursively get the accessories of the accessory
		art, err := c.getArtifact(project, repo, acc.GetData().Digest, true)
		if err != nil {
			return err
		}
		for _, tag := range art.Tags {
			accArt.Tags = append(accArt.Tags, tag.Name)
		}
		*accArts = append(*accArts, accArt)
		if err != c.getAccessoryArts(project, repo, art, labels, tags, accArts) {
			return err
		}
	}
	return nil
}

func (c *client) listTags(project, repo, digest string) ([]string, error) {
	tags := []*tagmodel.Tag{}
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s/tags",
		c.BasePath(), project, repo, digest)
	if err := c.C.GetAndIteratePagination(url, &tags); err != nil {
		return nil, err
	}
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}
	return tagNames, nil
}

func (c *client) getArtifact(project, repo, digest string, withAccessory bool) (*artifact.Artifact, error) {
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s?with_accessory=%t&with_tag=true",
		c.BasePath(), project, repo, digest, withAccessory)
	artifact := &artifact.Artifact{}
	if err := c.C.Get(url, &artifact); err != nil {
		return nil, err
	}
	return artifact, nil
}

func (c *client) deleteTag(repo, tag string) error {
	project, repo := utils.ParseRepository(repo)
	repo = repository.Encode(repo)
	url := fmt.Sprintf("%s/projects/%s/repositories/%s/artifacts/%s/tags/%s",
		c.BasePath(), project, repo, tag, tag)
	return c.C.Delete(url)
}

func (c *client) getRepositoryByBlobDigest(digest string) (string, error) {
	repositories := []*repomodel.RepoRecord{}
	url := fmt.Sprintf("%s/repositories?q=blob_digest=%s&page_size=1&page_number=1", c.BasePath(), digest)
	if err := c.C.Get(url, &repositories); err != nil {
		return "", err
	}
	if len(repositories) == 0 {
		return "", nil
	}
	return repositories[0].Name, nil
}
