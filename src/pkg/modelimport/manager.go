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

package modelimport

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/modelimport/dao"
	"github.com/goharbor/harbor/src/pkg/modelimport/model"
)

// Mgr is the global model import policy manager.
var Mgr = NewManager()

// Manager manages model import policies.
type Manager interface {
	Count(ctx context.Context, query *q.Query) (int64, error)
	Create(ctx context.Context, policy *model.Policy) (int64, error)
	Update(ctx context.Context, policy *model.Policy, props ...string) error
	Get(ctx context.Context, id int64) (*model.Policy, error)
	GetByName(ctx context.Context, projectID int64, name string) (*model.Policy, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, query *q.Query) ([]*model.Policy, error)
	ListByProject(ctx context.Context, projectID int64, query *q.Query) ([]*model.Policy, error)
}

type manager struct {
	dao dao.DAO
}

// NewManager returns the default manager.
func NewManager() Manager {
	return &manager{dao: dao.New()}
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) Create(ctx context.Context, policy *model.Policy) (int64, error) {
	if err := policy.Encode(); err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, policy)
}

func (m *manager) Update(ctx context.Context, policy *model.Policy, props ...string) error {
	if err := policy.Encode(); err != nil {
		return err
	}
	return m.dao.Update(ctx, policy, props...)
}

func (m *manager) Get(ctx context.Context, id int64) (*model.Policy, error) {
	policy, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return decode(policy)
}

func (m *manager) GetByName(ctx context.Context, projectID int64, name string) (*model.Policy, error) {
	policy, err := m.dao.GetByName(ctx, projectID, name)
	if err != nil {
		return nil, err
	}
	return decode(policy)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	policies, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	for i := range policies {
		if policies[i], err = decode(policies[i]); err != nil {
			return nil, err
		}
	}
	return policies, nil
}

func (m *manager) ListByProject(ctx context.Context, projectID int64, query *q.Query) ([]*model.Policy, error) {
	query = q.MustClone(query)
	if query.Keywords == nil {
		query.Keywords = map[string]any{}
	}
	query.Keywords["project_id"] = projectID
	return m.List(ctx, query)
}

func decode(policy *model.Policy) (*model.Policy, error) {
	if policy == nil {
		return nil, nil
	}
	if err := policy.Decode(); err != nil {
		return nil, err
	}
	return policy, nil
}
