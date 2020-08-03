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

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/allowlist"
	"github.com/goharbor/harbor/src/pkg/user"
)

var (
	// Ctl is a global project controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// Create create project instance
	Create(ctx context.Context, project *models.Project) (int64, error)
	// Count returns the total count of projects according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Delete delete the project by project id
	Delete(ctx context.Context, id int64) error
	// Get get the project by project id
	Get(ctx context.Context, projectID int64, options ...Option) (*models.Project, error)
	// GetByName get the project by project name
	GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error)
	// List list projects
	List(ctx context.Context, query *models.ProjectQueryParam, options ...Option) ([]*models.Project, error)
}

// NewController creates an instance of the default project controller
func NewController() Controller {
	return &controller{
		projectMgr:   project.Mgr,
		metaMgr:      metadata.Mgr,
		allowlistMgr: allowlist.NewDefaultManager(),
		userMgr:      user.Mgr,
	}
}

type controller struct {
	projectMgr   project.Manager
	metaMgr      metadata.Manager
	allowlistMgr allowlist.Manager
	userMgr      user.Manager
}

func (c *controller) Create(ctx context.Context, project *models.Project) (int64, error) {
	var projectID int64
	h := func(ctx context.Context) (err error) {
		projectID, err = c.projectMgr.Create(ctx, project)
		if err != nil {
			return err
		}

		if err := c.allowlistMgr.CreateEmpty(projectID); err != nil {
			log.Errorf("failed to create CVE allowlist for project %s: %v", project.Name, err)
			return err
		}

		if len(project.Metadata) > 0 {
			if err = c.metaMgr.Add(ctx, projectID, project.Metadata); err != nil {
				log.Errorf("failed to add metadata for project %s: %v", project.Name, err)
				return err
			}
		}
		return nil
	}

	if err := orm.WithTransaction(h)(ctx); err != nil {
		return 0, err
	}

	return projectID, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.projectMgr.Count(ctx, query)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.projectMgr.Delete(ctx, id)
}

func (c *controller) Get(ctx context.Context, projectID int64, options ...Option) (*models.Project, error) {
	p, err := c.projectMgr.Get(ctx, projectID)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	if opts.WithOwner {
		if err := c.loadOwners(ctx, models.Projects{p}); err != nil {
			return nil, err
		}
	}
	return c.assembleProject(ctx, p, opts)
}

func (c *controller) GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error) {
	if projectName == "" {
		return nil, errors.BadRequestError(nil).WithMessage("project name required")
	}

	p, err := c.projectMgr.Get(ctx, projectName)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	if opts.WithOwner {
		if err := c.loadOwners(ctx, models.Projects{p}); err != nil {
			return nil, err
		}
	}
	return c.assembleProject(ctx, p, newOptions(options...))
}

func (c *controller) List(ctx context.Context, query *models.ProjectQueryParam, options ...Option) ([]*models.Project, error) {
	projects, err := c.projectMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	opts := newOptions(options...)
	if opts.WithOwner {
		if err := c.loadOwners(ctx, projects); err != nil {
			return nil, err
		}
	}

	for _, p := range projects {
		if _, err := c.assembleProject(ctx, p, opts); err != nil {
			return nil, err
		}
	}

	return projects, nil
}

func (c *controller) loadOwners(ctx context.Context, projects models.Projects) error {
	owners, err := c.userMgr.List(ctx, q.New(q.KeyWords{"user_id__in": projects.OwnerIDs()}))
	if err != nil {
		return err
	}
	m := owners.MapByUserID()
	for _, p := range projects {
		owner, ok := m[p.OwnerID]
		if !ok {
			log.G(ctx).Warningf("the owner of project %s is not found, owner id is %d", p.Name, p.OwnerID)
			continue
		}

		p.OwnerName = owner.Username
	}

	return nil
}

func (c *controller) assembleProject(ctx context.Context, p *models.Project, opts *Options) (*models.Project, error) {
	if opts.Metadata {
		meta, err := c.metaMgr.Get(ctx, p.ProjectID)
		if err != nil {
			return nil, err
		}
		if len(p.Metadata) == 0 {
			p.Metadata = make(map[string]string)
		}

		for k, v := range meta {
			p.Metadata[k] = v
		}
	}

	if opts.CVEAllowlist {
		if p.ReuseSysCVEAllowlist() {
			wl, err := c.allowlistMgr.GetSys()
			if err != nil {
				log.Errorf("get system CVE allowlist failed, error: %v", err)
				return nil, err
			}

			wl.ProjectID = p.ProjectID
			p.CVEAllowlist = *wl
		} else {
			wl, err := c.allowlistMgr.Get(p.ProjectID)
			if err != nil {
				return nil, err
			}

			p.CVEAllowlist = *wl
		}

	}

	return p, nil
}
