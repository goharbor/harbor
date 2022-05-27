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

	commonmodels "github.com/goharbor/harbor/src/common/models"
	event "github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/allowlist"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/user"
)

var (
	// Ctl is a global project controller instance
	Ctl = NewController()
)

// Project alias to models.Project
type Project = models.Project

// MemberQuery alias to models.MemberQuery
type MemberQuery = models.MemberQuery

// Controller defines the operations related with blobs
type Controller interface {
	// Create create project instance
	Create(ctx context.Context, project *models.Project) (int64, error)
	// Count returns the total count of projects according to the query
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Delete delete the project by project id
	Delete(ctx context.Context, id int64) error
	// Exists returns true when the specific project exists
	Exists(ctx context.Context, projectIDOrName interface{}) (bool, error)
	// Get get the project by project id or name
	Get(ctx context.Context, projectIDOrName interface{}, options ...Option) (*models.Project, error)
	// GetByName get the project by project name
	GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error)
	// List list projects
	List(ctx context.Context, query *q.Query, options ...Option) ([]*models.Project, error)
	// Update update the project
	Update(ctx context.Context, project *models.Project) error
	// ListRoles lists the roles of user for the specific project
	ListRoles(ctx context.Context, projectID int64, u *commonmodels.User) ([]int, error)
}

// NewController creates an instance of the default project controller
func NewController() Controller {
	return &controller{
		projectMgr:   pkg.ProjectMgr,
		metaMgr:      pkg.ProjectMetaMgr,
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

		if err := c.allowlistMgr.CreateEmpty(ctx, projectID); err != nil {
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

	if err := orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-create-project")); err != nil {
		return 0, err
	}

	// fire event
	e := &event.CreateProjectEventMetadata{
		ProjectID: projectID,
		Project:   project.Name,
		Operator:  operator.FromContext(ctx),
	}
	notification.AddEvent(ctx, e)

	return projectID, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.projectMgr.Count(ctx, query)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	proj, err := c.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := c.projectMgr.Delete(ctx, id); err != nil {
		return err
	}

	e := &event.DeleteProjectEventMetadata{
		ProjectID: proj.ProjectID,
		Project:   proj.Name,
		Operator:  operator.FromContext(ctx),
	}
	notification.AddEvent(ctx, e)

	return nil
}

func (c *controller) Exists(ctx context.Context, projectIDOrName interface{}) (bool, error) {
	_, err := c.projectMgr.Get(ctx, projectIDOrName)
	if err == nil {
		return true, nil
	} else if errors.IsNotFoundErr(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (c *controller) Get(ctx context.Context, projectIDOrName interface{}, options ...Option) (*models.Project, error) {
	p, err := c.projectMgr.Get(ctx, projectIDOrName)
	if err != nil {
		return nil, err
	}

	if err := c.assembleProjects(ctx, models.Projects{p}, options...); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *controller) GetByName(ctx context.Context, projectName string, options ...Option) (*models.Project, error) {
	if projectName == "" {
		return nil, errors.BadRequestError(nil).WithMessage("project name required")
	}

	p, err := c.projectMgr.Get(ctx, projectName)
	if err != nil {
		return nil, err
	}

	if err := c.assembleProjects(ctx, models.Projects{p}, options...); err != nil {
		return nil, err
	}

	return p, nil
}

func (c *controller) List(ctx context.Context, query *q.Query, options ...Option) ([]*models.Project, error) {
	projects, err := c.projectMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(projects) == 0 {
		return projects, nil
	}

	if err := c.assembleProjects(ctx, projects, options...); err != nil {
		return nil, err
	}

	return projects, nil
}

func (c *controller) Update(ctx context.Context, p *models.Project) error {
	// currently, allowlist manager not use the ormer from the context,
	// the SQL executed in the allowlist manager will not be in the transaction with metadata manager,
	// we will update the metadata of the project first so that we can be rollback the operations for the metadata
	// when set allowlist for the project failed
	if len(p.Metadata) > 0 {
		meta, err := c.metaMgr.Get(ctx, p.ProjectID)
		if err != nil {
			return err
		}
		if meta == nil {
			meta = map[string]string{}
		}

		metaNeedUpdated := map[string]string{}
		metaNeedCreated := map[string]string{}
		for key, value := range p.Metadata {
			_, exist := meta[key]
			if exist {
				metaNeedUpdated[key] = value
			} else {
				metaNeedCreated[key] = value
			}
		}
		if err = c.metaMgr.Add(ctx, p.ProjectID, metaNeedCreated); err != nil {
			return err
		}
		if err = c.metaMgr.Update(ctx, p.ProjectID, metaNeedUpdated); err != nil {
			return err
		}
	}

	if p.CVEAllowlist.ProjectID == p.ProjectID {
		if err := c.allowlistMgr.Set(ctx, p.ProjectID, p.CVEAllowlist); err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) ListRoles(ctx context.Context, projectID int64, u *commonmodels.User) ([]int, error) {
	if u == nil {
		return nil, nil
	}

	return c.projectMgr.ListRoles(ctx, projectID, u.UserID, u.GroupIDs...)
}

func (c *controller) assembleProjects(ctx context.Context, projects models.Projects, options ...Option) error {
	opts := newOptions(options...)
	if !opts.WithDetail {
		return nil
	}
	if opts.WithMetadata {
		if err := c.loadMetadatas(ctx, projects); err != nil {
			return err
		}
	}

	if opts.WithEffectCVEAllowlist {
		if err := c.loadEffectCVEAllowlists(ctx, projects); err != nil {
			return err
		}
	} else if opts.WithCVEAllowlist {
		if err := c.loadCVEAllowlists(ctx, projects); err != nil {
			return err
		}
	}

	if opts.WithOwner {
		if err := c.loadOwners(ctx, projects); err != nil {
			return err
		}
	}

	return nil
}

func (c *controller) loadCVEAllowlists(ctx context.Context, projects models.Projects) error {
	if len(projects) == 0 {
		return nil
	}

	for _, p := range projects {
		wl, err := c.allowlistMgr.Get(ctx, p.ProjectID)
		if err != nil {
			return err
		}

		p.CVEAllowlist = *wl
	}

	return nil
}

func (c *controller) loadEffectCVEAllowlists(ctx context.Context, projects models.Projects) error {
	if len(projects) == 0 {
		return nil
	}

	for _, p := range projects {
		if p.ReuseSysCVEAllowlist() {
			wl, err := c.allowlistMgr.GetSys(ctx)
			if err != nil {
				log.Errorf("get system CVE allowlist failed, error: %v", err)
				return err
			}

			wl.ProjectID = p.ProjectID
			p.CVEAllowlist = *wl
		} else {
			wl, err := c.allowlistMgr.Get(ctx, p.ProjectID)
			if err != nil {
				return err
			}

			p.CVEAllowlist = *wl
		}
	}

	return nil
}

func (c *controller) loadMetadatas(ctx context.Context, projects models.Projects) error {
	if len(projects) == 0 {
		return nil
	}

	for _, p := range projects {
		meta, err := c.metaMgr.Get(ctx, p.ProjectID)
		if err != nil {
			return err
		}
		p.Metadata = meta
	}

	return nil
}

func (c *controller) loadOwners(ctx context.Context, projects models.Projects) error {
	if len(projects) == 0 {
		return nil
	}

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
