// Copyright 2018 Project Harbor Authors
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

package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	coreutils "github.com/goharbor/harbor/src/core/utils"
)

// RepositoryLabelAPI handles requests for adding/removing label to/from repositories and images
type RepositoryLabelAPI struct {
	LabelResourceAPI
	repository *models.RepoRecord
	tag        string
	label      *models.Label
}

// Prepare ...
func (r *RepositoryLabelAPI) Prepare() {
	// Super
	r.LabelResourceAPI.Prepare()

	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	repository := r.GetString(":splat")
	project, _ := utils.ParseRepository(repository)
	if !r.checkPermissions(project) {
		r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		return
	}

	repo, err := dao.GetRepositoryByName(repository)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get repository %s: %v", repository, err))
		return
	}

	if repo == nil {
		r.SendNotFoundError(fmt.Errorf("repository %s not found", repository))
		return
	}
	r.repository = repo

	tag := r.GetString(":tag")
	if len(tag) > 0 {
		exist, err := imageExist(r.SecurityCtx.GetUsername(), repository, tag)
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to check the existence of image %s:%s: %v", repository, tag, err))
			return
		}
		if !exist {
			r.SendNotFoundError(fmt.Errorf("image %s:%s not found", repository, tag))
			return
		}
		r.tag = tag
	}

	if r.Ctx.Request.Method == http.MethodPost {
		p, err := r.ProjectMgr.Get(project)
		if err != nil {
			r.SendInternalServerError(err)
			return
		}

		l := &models.Label{}
		r.DecodeJSONReq(l)

		label, ok := r.validate(l.ID, p.ProjectID)
		if !ok {
			return
		}
		r.label = label

		return
	}

	if r.Ctx.Request.Method == http.MethodDelete {
		labelID, err := r.GetInt64FromPath(":id")
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to get ID parameter from path: %v", err))
			return
		}

		label, ok := r.exists(labelID)
		if !ok {
			return
		}

		r.label = label
	}
}

// GetOfImage returns labels of an image
func (r *RepositoryLabelAPI) GetOfImage() {
	r.getLabelsOfResource(common.ResourceTypeImage, fmt.Sprintf("%s:%s", r.repository.Name, r.tag))
}

// AddToImage adds the label to an image
func (r *RepositoryLabelAPI) AddToImage() {
	rl := &models.ResourceLabel{
		LabelID:      r.label.ID,
		ResourceType: common.ResourceTypeImage,
		ResourceName: fmt.Sprintf("%s:%s", r.repository.Name, r.tag),
	}
	r.markLabelToResource(rl)
}

// RemoveFromImage removes the label from an image
func (r *RepositoryLabelAPI) RemoveFromImage() {
	r.removeLabelFromResource(common.ResourceTypeImage,
		fmt.Sprintf("%s:%s", r.repository.Name, r.tag), r.label.ID)
}

// GetOfRepository returns labels of a repository
func (r *RepositoryLabelAPI) GetOfRepository() {
	r.getLabelsOfResource(common.ResourceTypeRepository, r.repository.RepositoryID)
}

// AddToRepository adds the label to a repository
func (r *RepositoryLabelAPI) AddToRepository() {
	rl := &models.ResourceLabel{
		LabelID:      r.label.ID,
		ResourceType: common.ResourceTypeRepository,
		ResourceID:   r.repository.RepositoryID,
	}
	r.markLabelToResource(rl)
}

// RemoveFromRepository removes the label from a repository
func (r *RepositoryLabelAPI) RemoveFromRepository() {
	r.removeLabelFromResource(common.ResourceTypeRepository, r.repository.RepositoryID, r.label.ID)
}

func imageExist(username, repository, tag string) (bool, error) {
	client, err := coreutils.NewRepositoryClientForUI(username, repository)
	if err != nil {
		return false, err
	}

	_, exist, err := client.ManifestExist(tag)
	return exist, err
}
