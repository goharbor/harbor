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

package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

// DAO is the data access object interface for project
type DAO interface {
	// Create create a project instance
	Create(ctx context.Context, project *models.Project) (int64, error)
	// Count returns the total count of projects according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// Delete delete the project instance by id
	Delete(ctx context.Context, id int64) error
	// Get get project instance by id
	Get(ctx context.Context, id int64) (*models.Project, error)
	// GetByName get project instance by name
	GetByName(ctx context.Context, name string) (*models.Project, error)
	// List list projects
	List(ctx context.Context, query *q.Query) ([]*models.Project, error)
	// Lists the roles of user for the specific project
	ListRoles(ctx context.Context, projectID int64, userID int, groupIDs ...int) ([]int, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create create a project instance
func (d *dao) Create(ctx context.Context, project *models.Project) (int64, error) {
	var projectID int64

	h := func(ctx context.Context) error {
		o, err := orm.FromContext(ctx)
		if err != nil {
			return err
		}

		now := time.Now()
		project.CreationTime = now
		project.UpdateTime = now

		projectID, err = o.Insert(project)
		if err != nil {
			return orm.WrapConflictError(err, "The project named %s already exists", project.Name)
		}

		member := &models.Member{
			ProjectID:    projectID,
			EntityID:     project.OwnerID,
			Role:         common.RoleProjectAdmin,
			EntityType:   common.UserMember,
			CreationTime: now,
			UpdateTime:   now,
		}

		if _, err := o.Insert(member); err != nil {
			return err
		}

		return nil
	}

	if err := orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-create-project")); err != nil {
		return 0, err
	}

	return projectID, nil
}

// Count returns the total count of artifacts according to the query
func (d *dao) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false
	qs, err := orm.QuerySetterForCount(ctx, &models.Project{}, query)
	if err != nil {
		return 0, err
	}

	return qs.Count()
}

// Delete delete the project instance by id
func (d *dao) Delete(ctx context.Context, id int64) error {
	project, err := d.Get(ctx, id)
	if err != nil {
		return err
	}

	project.Deleted = true
	project.Name = lib.Truncate(project.Name, fmt.Sprintf("#%d", project.ProjectID), 255)

	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = o.Update(project, "deleted", "name")
	return err
}

// Get get project instance by id
func (d *dao) Get(ctx context.Context, id int64) (*models.Project, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	project := &models.Project{ProjectID: id, Deleted: false}
	if err = o.Read(project, "project_id", "deleted"); err != nil {
		return nil, orm.WrapNotFoundError(err, "project %d not found", id)
	}
	return project, nil
}

// GetByName get project instance by name
func (d *dao) GetByName(ctx context.Context, name string) (*models.Project, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	project := &models.Project{Name: name, Deleted: false}
	if err := o.Read(project, "name", "deleted"); err != nil {
		return nil, orm.WrapNotFoundError(err, "project %s not found", name)
	}
	return project, nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.Project, error) {
	query = q.MustClone(query)
	query.Keywords["deleted"] = false

	qs, err := orm.QuerySetter(ctx, &models.Project{}, query)
	if err != nil {
		return nil, err
	}

	projects := []*models.Project{}
	if _, err := qs.All(&projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func (d *dao) ListRoles(ctx context.Context, projectID int64, userID int, groupIDs ...int) ([]int, error) {
	qs, err := orm.QuerySetter(ctx, &models.Member{}, nil)
	if err != nil {
		return nil, err
	}

	conds := []*orm.Condition{
		orm.NewCondition().And("entity_type", "u").And("project_id", projectID).And("entity_id", userID),
	}
	if len(groupIDs) > 0 {
		conds = append(conds, orm.NewCondition().And("entity_type", "g").And("project_id", projectID).And("entity_id__in", groupIDs))
	}

	cond := orm.NewCondition()
	for _, c := range conds {
		cond = cond.OrCond(c)
	}

	var values orm.ParamsList
	if _, err := qs.SetCond(cond).ValuesFlat(&values, "role"); err != nil {
		return nil, err
	}

	var roles []int
	for _, value := range values {
		roles = append(roles, int(value.(int64)))
	}

	return roles, nil
}
