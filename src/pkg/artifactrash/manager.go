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

package artifactrash

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/artifactrash/dao"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
)

var (
	// Mgr is a global artifact trash manager instance
	Mgr = NewManager()
)

// Manager is the only interface of artifact module to provide the management functions for artifacts
type Manager interface {
	// Create ...
	Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error)
	// Delete ...
	Delete(ctx context.Context, id int64) (err error)
	// Filter ...
	Filter(ctx context.Context) (arts []model.ArtifactTrash, err error)
	// Flush clean the trash table
	Flush(ctx context.Context) (err error)
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{
		dao.New(),
	}
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, artifactrsh *model.ArtifactTrash) (id int64, err error) {
	return m.dao.Create(ctx, artifactrsh)
}
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}
func (m *manager) Filter(ctx context.Context) (arts []model.ArtifactTrash, err error) {
	return m.dao.Filter(ctx)
}

func (m *manager) Flush(ctx context.Context) (err error) {
	return m.dao.Flush(ctx)
}
