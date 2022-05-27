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

package metadata

import (
	"context"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/metadata/dao"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
)

// Manager defines the operations that a project metadata manager should implement
type Manager interface {
	// Add metadatas for project specified by projectID
	Add(ctx context.Context, projectID int64, meta map[string]string) error

	// Delete metadatas whose keys are specified in parameter meta, if it is absent, delete all
	Delete(ctx context.Context, projectID int64, meta ...string) error

	// Update metadatas
	Update(ctx context.Context, projectID int64, meta map[string]string) error

	// Get metadatas whose keys are specified in parameter meta, if it is absent, get all
	Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error)

	// List metadata according to the name and value
	List(ctx context.Context, name, value string) ([]*models.ProjectMetadata, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{dao: dao.New()}
}

type manager struct {
	dao dao.DAO
}

// Add metadatas for project specified by projectID
func (m *manager) Add(ctx context.Context, projectID int64, meta map[string]string) error {
	h := func(ctx context.Context) error {
		for name, value := range meta {
			if _, err := m.dao.Create(ctx, projectID, name, value); err != nil {
				return err
			}
		}
		return nil
	}
	return orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-add-project"))
}

// Delete metadatas whose keys are specified in parameter meta, if it is absent, delete all
func (m *manager) Delete(ctx context.Context, projectID int64, meta ...string) error {
	return m.dao.Delete(ctx, makeQuery(projectID, meta...))
}

// Update metadatas
func (m *manager) Update(ctx context.Context, projectID int64, meta map[string]string) error {
	if len(meta) == 0 {
		return nil
	}

	h := func(ctx context.Context) error {
		for name, value := range meta {
			if err := m.dao.Update(ctx, projectID, name, value); err != nil {
				return err
			}
		}

		return nil
	}

	return orm.WithTransaction(h)(orm.SetTransactionOpNameToContext(ctx, "tx-delete-project"))
}

// Get metadatas whose keys are specified in parameter meta, if it is absent, get all
func (m *manager) Get(ctx context.Context, projectID int64, meta ...string) (map[string]string, error) {
	mds, err := m.dao.List(ctx, makeQuery(projectID, meta...))
	if err != nil {
		return nil, err
	}

	data := map[string]string{}
	for _, md := range mds {
		data[md.Name] = md.Value
	}

	return data, nil
}

// List metadata according to the name and value
func (m *manager) List(ctx context.Context, name string, value string) ([]*models.ProjectMetadata, error) {
	return m.dao.List(ctx, q.New(q.KeyWords{"name": name, "value": value}))
}

func makeQuery(projectID int64, meta ...string) *q.Query {
	kw := q.KeyWords{
		"project_id": projectID,
	}
	if len(meta) > 0 {
		var names []string
		for _, name := range meta {
			names = append(names, name)
		}
		kw["name__in"] = names
	}

	return q.New(kw)
}
