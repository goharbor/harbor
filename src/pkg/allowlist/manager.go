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

package allowlist

import (
	"context"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/allowlist/dao"
	"github.com/goharbor/harbor/src/pkg/allowlist/models"
)

// Manager defines the interface of CVE allowlist manager, it support both system level and project level allowlists
type Manager interface {
	// CreateEmpty creates empty allowlist for given project
	CreateEmpty(ctx context.Context, projectID int64) error
	// Set sets the allowlist for given project (create or update)
	Set(ctx context.Context, projectID int64, list models.CVEAllowlist) error
	// Get gets the allowlist for given project
	Get(ctx context.Context, projectID int64) (*models.CVEAllowlist, error)
	// SetSys sets system level allowlist
	SetSys(ctx context.Context, list models.CVEAllowlist) error
	// GetSys gets system level allowlist
	GetSys(ctx context.Context) (*models.CVEAllowlist, error)
}

type defaultManager struct {
	dao dao.DAO
}

// CreateEmpty creates empty allowlist for given project
func (d *defaultManager) CreateEmpty(ctx context.Context, projectID int64) error {
	l := models.CVEAllowlist{
		ProjectID: projectID,
		Items:     []models.CVEAllowlistItem{},
	}
	_, err := d.dao.Set(ctx, l)
	if err != nil {
		logger.Errorf("Failed to create empty CVE allowlist for project: %d, error: %v", projectID, err)
	}
	return err
}

// Set sets the allowlist for given project (create or update)
func (d *defaultManager) Set(ctx context.Context, projectID int64, list models.CVEAllowlist) error {
	list.ProjectID = projectID
	if err := Validate(list); err != nil {
		return err
	}
	_, err := d.dao.Set(ctx, list)
	return err
}

// Get gets the allowlist for given project
func (d *defaultManager) Get(ctx context.Context, projectID int64) (*models.CVEAllowlist, error) {
	wl, err := d.dao.QueryByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if wl == nil {
		log.Debugf("No CVE allowlist found for project %d, returning empty list.", projectID)
		wl = &models.CVEAllowlist{ProjectID: projectID, Items: []models.CVEAllowlistItem{}}
	} else if wl.Items == nil {
		wl.Items = []models.CVEAllowlistItem{}
	}
	return wl, nil
}

// SetSys sets the system level allowlist
func (d *defaultManager) SetSys(ctx context.Context, list models.CVEAllowlist) error {
	return d.Set(ctx, 0, list)
}

// GetSys gets the system level allowlist
func (d *defaultManager) GetSys(ctx context.Context) (*models.CVEAllowlist, error) {
	return d.Get(ctx, 0)
}

// NewDefaultManager return a new instance of defaultManager
func NewDefaultManager() Manager {
	return &defaultManager{dao: dao.New()}
}
