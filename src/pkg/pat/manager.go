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

package pat

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/pat/dao"
	"github.com/goharbor/harbor/src/pkg/pat/model"
)

// Manager interface defines manager methods for personal access tokens
type Manager interface {
	Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, error)
	Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error)
	Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error)
	Count(ctx context.Context, query *q.Query) (total int64, err error)
}

// NewManager returns a new manager instance
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.DAO
}

// Create creates a new personal access token
func (m *manager) Create(ctx context.Context, pat *model.PersonalAccessToken) (int64, error) {
	return m.dao.Create(ctx, pat)
}

// Get returns a personal access token by ID
func (m *manager) Get(ctx context.Context, id int64) (*model.PersonalAccessToken, error) {
	return m.dao.Get(ctx, id)
}

// Update updates a personal access token
func (m *manager) Update(ctx context.Context, pat *model.PersonalAccessToken, props ...string) error {
	return m.dao.Update(ctx, pat, props...)
}

// Delete deletes a personal access token
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

// List lists personal access tokens
func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.PersonalAccessToken, error) {
	return m.dao.List(ctx, query)
}

// Count returns the count of personal access tokens
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}
