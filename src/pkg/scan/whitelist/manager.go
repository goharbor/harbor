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

package whitelist

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

// Manager defines the interface of CVE whitelist manager, it support both system level and project level whitelists
type Manager interface {
	// CreateEmpty creates empty whitelist for given project
	CreateEmpty(projectID int64) error
	// Set sets the whitelist for given project (create or update)
	Set(projectID int64, list models.CVEWhitelist) error
	// Get gets the whitelist for given project
	Get(projectID int64) (*models.CVEWhitelist, error)
	// SetSys sets system level whitelist
	SetSys(list models.CVEWhitelist) error
	// GetSys gets system level whitelist
	GetSys() (*models.CVEWhitelist, error)
}

type defaultManager struct{}

// CreateEmpty creates empty whitelist for given project
func (d *defaultManager) CreateEmpty(projectID int64) error {
	l := models.CVEWhitelist{
		ProjectID: projectID,
		Items:     []models.CVEWhitelistItem{},
	}
	_, err := dao.CreateCVEWhitelist(l)
	if err != nil {
		logger.Errorf("Failed to create empty CVE whitelist for project: %d, error: %v", projectID, err)
	}
	return err
}

// Set sets the whitelist for given project (create or update)
func (d *defaultManager) Set(projectID int64, list models.CVEWhitelist) error {
	list.ProjectID = projectID
	if err := Validate(list); err != nil {
		return err
	}
	_, err := dao.UpdateCVEWhitelist(list)
	return err
}

// Get gets the whitelist for given project
func (d *defaultManager) Get(projectID int64) (*models.CVEWhitelist, error) {
	wl, err := dao.GetCVEWhitelist(projectID)
	if wl == nil && err == nil {
		log.Debugf("No CVE whitelist found for project %d, returning empty list.", projectID)
		return &models.CVEWhitelist{ProjectID: projectID, Items: []models.CVEWhitelistItem{}}, nil
	}
	if wl.Items == nil {
		wl.Items = []models.CVEWhitelistItem{}
	}
	return wl, err
}

// SetSys sets the system level whitelist
func (d *defaultManager) SetSys(list models.CVEWhitelist) error {
	return d.Set(0, list)
}

// GetSys gets the system level whitelist
func (d *defaultManager) GetSys() (*models.CVEWhitelist, error) {
	return d.Get(0)
}

// NewDefaultManager return a new instance of defaultManager
func NewDefaultManager() Manager {
	return &defaultManager{}
}
