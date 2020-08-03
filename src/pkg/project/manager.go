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
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/dao"
	"github.com/goharbor/harbor/src/pkg/project/models"
)

var (
	// Mgr is the global project manager
	Mgr = New()
)

// Manager is used for project management
type Manager interface {
	// Create create project instance
	Create(ctx context.Context, project *models.Project) (int64, error)

	// Count returns the total count of projects according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Delete delete the project instance by id
	Delete(ctx context.Context, id int64) error

	// Get the project specified by the ID or name
	Get(ctx context.Context, idOrName interface{}) (*models.Project, error)

	// List projects according to the query
	List(ctx context.Context, query ...*models.ProjectQueryParam) ([]*models.Project, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

// Create create project instance
func (m *manager) Create(ctx context.Context, project *models.Project) (int64, error) {
	if project.OwnerID <= 0 {
		return 0, errors.BadRequestError(nil).WithMessage("Owner is missing when creating project %s", project.Name)
	}
	return m.dao.Create(ctx, project)
}

// Count returns the total count of projects according to the query
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Delete delete the project instance by id
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

// Get the project specified by the ID
func (m *manager) Get(ctx context.Context, idOrName interface{}) (*models.Project, error) {
	id, ok := idOrName.(int64)
	if ok {
		return m.dao.Get(ctx, id)
	}
	name, ok := idOrName.(string)
	if ok {
		return m.dao.GetByName(ctx, name)
	}
	return nil, errors.Errorf("invalid parameter: %v, should be ID(int64) or name(string)", idOrName)
}

// List projects according to the query
func (m *manager) List(ctx context.Context, query ...*models.ProjectQueryParam) ([]*models.Project, error) {
	var param *models.ProjectQueryParam
	if len(query) > 0 {
		param = query[0]
	}
	if param == nil {
		return m.dao.List(ctx, nil)
	}

	return m.dao.List(ctx, param.ToQuery())
}
