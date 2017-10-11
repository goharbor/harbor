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

package metamgr

import (
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

// ProjectMetadataManaegr defines the operations that a project metadata manager should
// implement
type ProjectMetadataManaegr interface {
	// Add metadatas for project specified by projectID
	Add(projectID int64, meta map[string]string) error
	// Delete metadatas whose keys are specified in parameter meta, if it
	// is absent, delete all
	Delete(projecdtID int64, meta ...string) error
	// Update metadatas
	Update(projectID int64, meta map[string]string) error
	// Get metadatas whose keys are specified in parameter meta, if it is
	// absent, get all
	Get(projectID int64, meta ...string) (map[string]string, error)
}

type defaultProjectMetadataManaegr struct{}

// NewDefaultProjectMetadataManager ...
func NewDefaultProjectMetadataManager() ProjectMetadataManaegr {
	return &defaultProjectMetadataManaegr{}
}

func (d *defaultProjectMetadataManaegr) Add(projectID int64, meta map[string]string) error {
	for k, v := range meta {
		proMeta := &models.ProjectMetadata{
			ProjectID: projectID,
			Name:      k,
			Value:     v,
		}
		if err := dao.AddProjectMetadata(proMeta); err != nil {
			return err
		}
	}
	return nil
}

func (d *defaultProjectMetadataManaegr) Delete(projectID int64, meta ...string) error {
	return dao.DeleteProjectMetadata(projectID, meta...)
}

func (d *defaultProjectMetadataManaegr) Update(projectID int64, meta map[string]string) error {
	for k, v := range meta {
		if err := dao.UpdateProjectMetadata(&models.ProjectMetadata{
			ProjectID: projectID,
			Name:      k,
			Value:     v,
		}); err != nil {
			return err
		}
	}

	// TODO remove the logic
	public, ok := meta[models.ProMetaPublic]
	if ok {
		i, err := strconv.Atoi(public)
		if err != nil {
			return err
		}
		return dao.ToggleProjectPublicity(projectID, i)
	}

	return nil
}

func (d *defaultProjectMetadataManaegr) Get(projectID int64, meta ...string) (map[string]string, error) {
	proMetas, err := dao.GetProjectMetadata(projectID, meta...)
	if err != nil {
		return nil, nil
	}

	m := map[string]string{}
	for _, proMeta := range proMetas {
		m[proMeta.Name] = proMeta.Value
	}

	return m, nil
}
