// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"fmt"
	"net/http"
	"strconv"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	uiutils "github.com/vmware/harbor/src/ui/utils"
)

// RepositoryLabelAPI handles requests for adding/removing label to/from repositories and images
type RepositoryLabelAPI struct {
	BaseController
	repository *models.RepoRecord
	tag        string
	label      *models.Label
}

// Prepare ...
func (r *RepositoryLabelAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.HandleUnauthorized()
		return
	}

	repository := r.GetString(":splat")
	project, _ := utils.ParseRepository(repository)
	if !r.SecurityCtx.HasWritePerm(project) {
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}

	repo, err := dao.GetRepositoryByName(repository)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get repository %s: %v",
			repository, err))
		return
	}
	if repo == nil {
		r.HandleNotFound(fmt.Sprintf("repository %s not found", repository))
		return
	}
	r.repository = repo

	tag := r.GetString(":tag")
	if len(tag) > 0 {
		exist, err := imageExist(r.SecurityCtx.GetUsername(), repository, tag)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to check the existence of image %s:%s: %v",
				repository, tag, err))
			return
		}
		if !exist {
			r.HandleNotFound(fmt.Sprintf("image %s:%s not found", repository, tag))
			return
		}
		r.tag = tag
	}

	if r.Ctx.Request.Method == http.MethodPost {
		l := &models.Label{}
		r.DecodeJSONReq(l)

		label, err := dao.GetLabel(l.ID)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", l.ID, err))
			return
		}

		if label == nil {
			r.HandleNotFound(fmt.Sprintf("label %d not found", l.ID))
			return
		}

		if label.Level != common.LabelLevelUser {
			r.HandleBadRequest("only user level labels can be used")
			return
		}

		if label.Scope == common.LabelScopeProject {
			p, err := r.ProjectMgr.Get(project)
			if err != nil {
				r.HandleInternalServerError(fmt.Sprintf("failed to get project %s: %v", project, err))
				return
			}

			if p.ProjectID != label.ProjectID {
				r.HandleBadRequest("can not add labels which don't belong to the project to the resources under the project")
				return
			}
		}
		r.label = label

		return
	}

	if r.Ctx.Request.Method == http.MethodDelete {
		labelID, err := r.GetInt64FromPath(":id")
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get ID parameter from path: %v", err))
			return
		}

		label, err := dao.GetLabel(labelID)
		if err != nil {
			r.HandleInternalServerError(fmt.Sprintf("failed to get label %d: %v", labelID, err))
			return
		}

		if label == nil {
			r.HandleNotFound(fmt.Sprintf("label %d not found", labelID))
			return
		}
		r.label = label
	}
}

// GetOfImage returns labels of an image
func (r *RepositoryLabelAPI) GetOfImage() {
	r.getLabels(common.ResourceTypeImage, fmt.Sprintf("%s:%s", r.repository.Name, r.tag))
}

// AddToImage adds the label to an image
func (r *RepositoryLabelAPI) AddToImage() {
	rl := &models.ResourceLabel{
		LabelID:      r.label.ID,
		ResourceType: common.ResourceTypeImage,
		ResourceName: fmt.Sprintf("%s:%s", r.repository.Name, r.tag),
	}
	r.addLabel(rl)
}

// RemoveFromImage removes the label from an image
func (r *RepositoryLabelAPI) RemoveFromImage() {
	r.removeLabel(common.ResourceTypeImage,
		fmt.Sprintf("%s:%s", r.repository.Name, r.tag), r.label.ID)
}

// GetOfRepository returns labels of a repository
func (r *RepositoryLabelAPI) GetOfRepository() {
	r.getLabels(common.ResourceTypeRepository, r.repository.RepositoryID)
}

// AddToRepository adds the label to a repository
func (r *RepositoryLabelAPI) AddToRepository() {
	rl := &models.ResourceLabel{
		LabelID:      r.label.ID,
		ResourceType: common.ResourceTypeRepository,
		ResourceID:   r.repository.RepositoryID,
	}
	r.addLabel(rl)
}

// RemoveFromRepository removes the label from a repository
func (r *RepositoryLabelAPI) RemoveFromRepository() {
	r.removeLabel(common.ResourceTypeRepository, r.repository.RepositoryID, r.label.ID)
}

func (r *RepositoryLabelAPI) getLabels(rType string, rIDOrName interface{}) {
	labels, err := dao.GetLabelsOfResource(rType, rIDOrName)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get labels of resource %s %v: %v",
			rType, rIDOrName, err))
		return
	}
	r.Data["json"] = labels
	r.ServeJSON()
}

func (r *RepositoryLabelAPI) addLabel(rl *models.ResourceLabel) {
	var rIDOrName interface{}
	if rl.ResourceID != 0 {
		rIDOrName = rl.ResourceID
	} else {
		rIDOrName = rl.ResourceName
	}
	rlabel, err := dao.GetResourceLabel(rl.ResourceType, rIDOrName, rl.LabelID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to check the existence of label %d for resource %s %v: %v",
			rl.LabelID, rl.ResourceType, rIDOrName, err))
		return
	}

	if rlabel != nil {
		r.HandleConflict()
		return
	}
	if _, err := dao.AddResourceLabel(rl); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to add label %d to resource %s %v: %v",
			rl.LabelID, rl.ResourceType, rIDOrName, err))
		return
	}

	// return the ID of label and return status code 200 rather than 201 as the label is not created
	r.Redirect(http.StatusOK, strconv.FormatInt(rl.LabelID, 10))
}

func (r *RepositoryLabelAPI) removeLabel(rType string, rIDOrName interface{}, labelID int64) {
	rl, err := dao.GetResourceLabel(rType, rIDOrName, labelID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to check the existence of label %d for resource %s %v: %v",
			labelID, rType, rIDOrName, err))
		return
	}

	if rl == nil {
		r.HandleNotFound(fmt.Sprintf("label %d of resource %s %s not found",
			labelID, rType, rIDOrName))
		return
	}
	if err = dao.DeleteResourceLabel(rl.ID); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to delete resource label record %d: %v",
			rl.ID, err))
		return
	}
}

func imageExist(username, repository, tag string) (bool, error) {
	client, err := uiutils.NewRepositoryClientForUI(username, repository)
	if err != nil {
		return false, err
	}

	_, exist, err := client.ManifestExist(tag)
	return exist, err
}
