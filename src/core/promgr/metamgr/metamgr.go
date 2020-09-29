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

package metamgr

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

// ProjectMetadataManager defines the operations that a project metadata manager should
// implement
type ProjectMetadataManager interface {
	// Add metadatas for project specified by projectID
	Add(projectID int64, meta map[string]string) error
	// Delete metadatas whose keys are specified in parameter meta, if it
	// is absent, delete all
	Delete(projectID int64, meta ...string) error
	// Update metadatas
	Update(projectID int64, meta map[string]string) error
	// Get metadatas whose keys are specified in parameter meta, if it is
	// absent, get all
	Get(projectID int64, meta ...string) (map[string]string, error)
	// List metadata according to the name and value
	List(name, value string) ([]*models.ProjectMetadata, error)
}

type defaultProjectMetadataManager struct{}

// NewDefaultProjectMetadataManager ...
func NewDefaultProjectMetadataManager() ProjectMetadataManager {
	return &defaultProjectMetadataManager{}
}

func (d *defaultProjectMetadataManager) Add(projectID int64, meta map[string]string) error {
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

func (d *defaultProjectMetadataManager) Delete(projectID int64, meta ...string) error {
	return dao.DeleteProjectMetadata(projectID, meta...)
}

func (d *defaultProjectMetadataManager) Update(projectID int64, meta map[string]string) error {
	for k, v := range meta {
		if err := dao.UpdateProjectMetadata(&models.ProjectMetadata{
			ProjectID: projectID,
			Name:      k,
			Value:     v,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (d *defaultProjectMetadataManager) Get(projectID int64, meta ...string) (map[string]string, error) {
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

func (d *defaultProjectMetadataManager) List(name, value string) ([]*models.ProjectMetadata, error) {
	metas := []*models.ProjectMetadata{}
	mds, err := dao.ListProjectMetadata(name, value)
	if err != nil {
		return nil, err
	}

	metas = append(metas, mds...)
	return metas, nil
}
