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
	Update(projectIDOrName interface{}, project *models.Project) error
	// TODO remove base
	List(query *models.ProjectQueryParam,
		base ...*models.BaseProjectCollection) (*models.ProjectQueryResult, error)
	IsPublic(projectIDOrName interface{}) (bool, error)
	Exists(projectIDOrName interface{}) (bool, error)
	// get all public project
	GetPublic() ([]*models.Project, error)
}

type defaultProjectManager struct {
	pmsDriver      pmsdriver.PMSDriver
	metaMgrEnabled bool // if metaMgrEnabled is enabled, metaMgr will be used to CURD metadata
	metaMgr        metamgr.ProjectMetadataManaegr
}

// NewDefaultProjectManager returns an instance of defaultProjectManager,
// if metaMgrEnabled is true, a project metadata manager will be created
// and used to CURD metadata
func NewDefaultProjectManager(driver pmsdriver.PMSDriver, metaMgrEnabled bool) ProjectManager {
	mgr := &defaultProjectManager{
		pmsDriver:      driver,
		metaMgrEnabled: metaMgrEnabled,
	}
	if metaMgrEnabled {
		mgr.metaMgr = metamgr.NewDefaultProjectMetadataManager()
	}
	return mgr
}

func (d *defaultProjectManager) Get(projectIDOrName interface{}) (*models.Project, error) {
	project, err := d.pmsDriver.Get(projectIDOrName)
	if err != nil {
		return nil, err
	}

	if project != nil && d.metaMgrEnabled {
		meta, err := d.metaMgr.Get(project.ProjectID)
		if err != nil {
			return nil, err
		}
		if len(project.Metadata) == 0 {
			project.Metadata = make(map[string]string)
		}
		for k, v := range meta {
			project.Metadata[k] = v
		}
	}
	return project, nil
}
func (d *defaultProjectManager) Create(project *models.Project) (int64, error) {
	id, err := d.pmsDriver.Create(project)
	if err != nil {
		return 0, err
	}
	if len(project.Metadata) > 0 && d.metaMgrEnabled {
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
	if project.Metadata != nil && d.metaMgrEnabled {
		if err = d.metaMgr.Delete(project.ProjectID); err != nil {
			return err
		}
	}
	return d.pmsDriver.Delete(project.ProjectID)
}

func (d *defaultProjectManager) Update(projectIDOrName interface{}, project *models.Project) error {
	if len(project.Metadata) > 0 && d.metaMgrEnabled {
		pro, err := d.Get(projectIDOrName)
		if err != nil {
			return err
		}
		if pro == nil {
			return fmt.Errorf("project %v not found", projectIDOrName)
		}
		if err = d.metaMgr.Update(pro.ProjectID, project.Metadata); err != nil {
			return err
		}
	}

	return d.pmsDriver.Update(projectIDOrName, project)
}

// TODO remove base
func (d *defaultProjectManager) List(query *models.ProjectQueryParam,
	base ...*models.BaseProjectCollection) (*models.ProjectQueryResult, error) {
	result, err := d.pmsDriver.List(query, base...)
	if err != nil {
		return nil, err
	}
	if d.metaMgrEnabled {
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

func (d *defaultProjectManager) Exists(projectIDOrName interface{}) (bool, error) {
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
