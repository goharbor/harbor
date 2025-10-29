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

package policy

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	dao "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/policy"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
)

// Mgr is a global instance of policy manager
var Mgr = New()

// Manager manages the policy
type Manager interface {
	// Count returns the total count of policies according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// Create the policy schema
	Create(ctx context.Context, schema *policy.Schema) (id int64, err error)
	// Update the policy schema, Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, schema *policy.Schema, props ...string) (err error)
	// Get the policy schema by id
	Get(ctx context.Context, id int64) (schema *policy.Schema, err error)
	// GetByName the policy schema by project ID and name
	GetByName(ctx context.Context, projectID int64, name string) (schema *policy.Schema, err error)
	// Delete the policy schema by id
	Delete(ctx context.Context, id int64) (err error)
	// List policy schemas by query
	ListPolicies(ctx context.Context, query *q.Query) (schemas []*policy.Schema, err error)
	// list policy schema under project
	ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) (schemas []*policy.Schema, err error)
}

type manager struct {
	dao dao.DAO
}

// New creates an instance of the default policy manager
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// Count returns the total count of policies according to the query
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Create the policy schema
func (m *manager) Create(ctx context.Context, schema *policy.Schema) (id int64, err error) {
	return m.dao.Create(ctx, schema)
}

// Update the policy schema, Only the properties specified by "props" will be updated if it is set
func (m *manager) Update(ctx context.Context, schema *policy.Schema, props ...string) (err error) {
	return m.dao.Update(ctx, schema, props...)
}

// Get the policy schema by id
func (m *manager) Get(ctx context.Context, id int64) (schema *policy.Schema, err error) {
	schema, err = m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = schema.Decode(); err != nil {
		return nil, err
	}
	return schema, nil
}

// Get the policy schema by name
func (m *manager) GetByName(ctx context.Context, projectID int64, name string) (schema *policy.Schema, err error) {
	schema, err = m.dao.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}

	if err = schema.Decode(); err != nil {
		return nil, err
	}
	return schema, nil
}

// Delete the policy schema by id
func (m *manager) Delete(ctx context.Context, id int64) (err error) {
	return m.dao.Delete(ctx, id)
}

// List policy schemas by query
func (m *manager) ListPolicies(ctx context.Context, query *q.Query) (schemas []*policy.Schema, err error) {
	schemas, err = m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}

	for i := range schemas {
		if err = schemas[i].Decode(); err != nil {
			return nil, err
		}
	}

	return schemas, nil
}

// list policy schema under project
func (m *manager) ListPoliciesByProject(ctx context.Context, project int64, query *q.Query) (schemas []*policy.Schema, err error) {
	if query == nil {
		query = &q.Query{}
	}

	if query.Keywords == nil {
		query.Keywords = make(map[string]any)
	}
	// set project filter
	query.Keywords["project_id"] = project

	return m.ListPolicies(ctx, query)
}
