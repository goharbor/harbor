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

package promgr

import (
	"fmt"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/promgr/metamgr"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver"
)

// ProjectManager is the project mamager which abstracts the operations related
// to projects
type ProjectManager interface {
	Get(projectIDOrName interface{}) (*models.Project, error)
	Create(*models.Project) (int64, error)
	Delete(projectIDOrName interface{}) error
	Update(projectIDOrName interface{}, metadata map[string]string) error
	// TODO remove base
	List(query *models.ProjectQueryParam,
		base ...*models.BaseProjectCollection) (*models.ProjectQueryResult, error)
	IsPublic(projectIDOrName interface{}) (bool, error)
	Exist(projectIDOrName interface{}) (bool, error)
	// get all public project
	GetPublic() ([]*models.Project, error)
}

type defaultProjectManager struct {
	pmsDriver pmsdriver.PMSDriver
	metaMgr   metamgr.ProjectMetadataManaegr
}

// NewDefaultProjectManager returns an instance of defaultProjectManager
func NewDefaultProjectManager(driver pmsdriver.PMSDriver) ProjectManager {
	mgr := &defaultProjectManager{
		pmsDriver: driver,
	}
	if driver.EnableExternalMetaMgr() {
		mgr.metaMgr = metamgr.NewDefaultProjectMetadataManager()
	}
	return mgr
}

func (d *defaultProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	project, err := d.pmsDriver.Get(projectIDOrName)
	if err != nil {
		return nil, err
	}

	if project != nil && d.pmsDriver.EnableExternalMetaMgr() {
		meta, err := d.metaMgr.Get(project.ProjectID)
		if err != nil {
			return nil, err
		}
		project.Metadata = meta
	}
	return project, nil
}
func (d *defaultProjectManager) Create(project *models.Project) (int64, error) {
	id, err := d.pmsDriver.Create(project)
	if err != nil {
		return 0, err
	}
	if len(project.Metadata) > 0 && d.pmsDriver.EnableExternalMetaMgr() {
		if err = d.metaMgr.Add(id, project.Metadata); err != nil {
			log.Errorf("failed to add metadata for project %s: %v", project.Name, err)
		}
	}
	return id, nil
}

func (d *defaultProjectManager) Delete(projectIDOrName interface{}) error {
	project, err := d.Get(projectIDOrName)
	if err != nil {
		return err
	}
	if project == nil {
		return nil
	}
	if project.Metadata != nil && d.pmsDriver.EnableExternalMetaMgr() {
		if err = d.metaMgr.Delete(project.ProjectID); err != nil {
			return err
		}
	}
	return d.pmsDriver.Delete(project.ProjectID)
}

func (d *defaultProjectManager) Update(projectIDOrName interface{}, metadata map[string]string) error {
	if len(metadata) == 0 {
		return nil
	}

	if d.pmsDriver.EnableExternalMetaMgr() {
		project, err := d.Get(projectIDOrName)
		if err != nil {
			return err
		}
		if project == nil {
			return fmt.Errorf("project %v not found", projectIDOrName)
		}
		return d.metaMgr.Update(project.ProjectID, metadata)
	}

	return d.pmsDriver.Update(projectIDOrName, metadata)
}

// TODO remove base
func (d *defaultProjectManager) List(query *models.ProjectQueryParam,
	base ...*models.BaseProjectCollection) (*models.ProjectQueryResult, error) {
	result, err := d.pmsDriver.List(query, base...)
	if err != nil {
		return nil, err
	}
	if d.pmsDriver.EnableExternalMetaMgr() {
		for _, project := range result.Projects {
			meta, err := d.metaMgr.Get(project.ProjectID)
			if err != nil {
				return nil, err
			}
			project.Metadata = meta
		}
	}
	return result, nil
}

func (d *defaultProjectManager) IsPublic(projectIDOrName interface{}) (bool, error) {
	project, err := d.Get(projectIDOrName)
	if err != nil {
		return false, err
	}
	if project == nil {
		return false, nil
	}
	return project.Public == 1, nil
}

func (d *defaultProjectManager) Exist(projectIDOrName interface{}) (bool, error) {
	project, err := d.Get(projectIDOrName)
	return project != nil, err
}

func (d *defaultProjectManager) GetPublic() ([]*models.Project, error) {
	value := true
	result, err := d.List(&models.ProjectQueryParam{
		Public: &value,
	})
	if err != nil {
		return nil, err
	}
	return result.Projects, nil
}
