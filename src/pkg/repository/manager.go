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

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/repository/dao"
	"github.com/goharbor/harbor/src/pkg/repository/model"
)

// Manager is used for repository management
type Manager interface {
	// Count returns the total count of repositories according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List repositories according to the query
	List(ctx context.Context, query *q.Query) (repositories []*model.RepoRecord, err error)
	// Get the repository specified by ID
	Get(ctx context.Context, id int64) (repository *model.RepoRecord, err error)
	// GetByName gets the repository specified by name
	GetByName(ctx context.Context, name string) (repository *model.RepoRecord, err error)
	// Create a repository
	Create(ctx context.Context, repository *model.RepoRecord) (id int64, err error)
	// Delete the repository specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update updates the repository. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, repository *model.RepoRecord, props ...string) (err error)
	// AddPullCount increase pull count for the specified repository
	AddPullCount(ctx context.Context, id int64, count uint64) error
	// NonEmptyRepos returns the repositories without any artifact or all the artifacts are untagged.
	NonEmptyRepos(ctx context.Context) ([]*model.RepoRecord, error)
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

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.RepoRecord, error) {
	repositories, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	return repositories, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*model.RepoRecord, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) GetByName(ctx context.Context, name string) (repository *model.RepoRecord, err error) {
	repositories, err := m.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"Name": name,
		},
	})
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("repository %s not found", name)
	}
	return repositories[0], nil
}

func (m *manager) Create(ctx context.Context, repository *model.RepoRecord) (int64, error) {
	return m.dao.Create(ctx, repository)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
func (m *manager) Update(ctx context.Context, repository *model.RepoRecord, props ...string) error {
	return m.dao.Update(ctx, repository, props...)
}

func (m *manager) AddPullCount(ctx context.Context, id int64, count uint64) error {
	return m.dao.AddPullCount(ctx, id, count)
}

func (m *manager) NonEmptyRepos(ctx context.Context) ([]*model.RepoRecord, error) {
	return m.dao.NonEmptyRepos(ctx)
}
