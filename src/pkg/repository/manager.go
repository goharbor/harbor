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

package repository

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository/dao"
)

// Mgr is the global repository manager instance
var Mgr = New()

// Manager is used for repository management
type Manager interface {
	// List repositories according to the query
	List(ctx context.Context, query *q.Query) (total int64, repositories []*models.RepoRecord, err error)
	// Get the repository specified by ID
	Get(ctx context.Context, id int64) (repository *models.RepoRecord, err error)
	// GetByName gets the repository specified by name
	GetByName(ctx context.Context, name string) (repository *models.RepoRecord, err error)
	// Create a repository
	Create(ctx context.Context, repository *models.RepoRecord) (id int64, err error)
	// Delete the repository specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update updates the repository. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, repository *models.RepoRecord, props ...string) (err error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) List(ctx context.Context, query *q.Query) (int64, []*models.RepoRecord, error) {
	total, err := m.dao.Count(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	repositories, err := m.dao.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	return total, repositories, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) GetByName(ctx context.Context, name string) (repository *models.RepoRecord, err error) {
	_, repositories, err := m.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Name": name,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, ierror.New(nil).WithCode(ierror.NotFoundCode).
			WithMessage("repository %s not found", name)
	}
	return repositories[0], nil
}

func (m *manager) Create(ctx context.Context, repository *models.RepoRecord) (int64, error) {
	return m.dao.Create(ctx, repository)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
func (m *manager) Update(ctx context.Context, repository *models.RepoRecord, props ...string) error {
	return m.dao.Update(ctx, repository)
}
