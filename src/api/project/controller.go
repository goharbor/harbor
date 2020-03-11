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

package project

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/project"
)

var (
	// Ctl is a global project controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// Get get the project by project id
	Get(ctx context.Context, projectID int64) (*models.Project, error)
	// GetByName get the project by project name
	GetByName(ctx context.Context, projectName string) (*models.Project, error)
}

// NewController creates an instance of the default project controller
func NewController() Controller {
	return &controller{
		projectMgr: project.Mgr,
	}
}

type controller struct {
	projectMgr project.Manager
}

func (c *controller) Get(ctx context.Context, projectID int64) (*models.Project, error) {
	p, err := c.projectMgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ierror.NotFoundError(nil).WithMessage("project %d not found", projectID)
	}

	return p, nil
}

func (c *controller) GetByName(ctx context.Context, projectName string) (*models.Project, error) {
	if projectName == "" {
		return nil, ierror.BadRequestError(nil).WithMessage("project name required")
	}

	p, err := c.projectMgr.Get(projectName)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ierror.NotFoundError(nil).WithMessage("project %s not found", projectName)
	}

	return p, nil
}
